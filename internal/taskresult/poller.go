package taskresult

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"fanapi/internal/cache"
	"fanapi/internal/db"
	"fanapi/internal/model"
	"fanapi/internal/script"
	"fanapi/internal/service"
)

const (
	pollInterval = 5 * time.Second
	maxAge       = 2 * time.Hour

	// pollLockTTL 必须大于最大可能的 query_timeout_ms，保证分布式锁在上游 HTTP 调用期间不过期。
	pollLockTTL           = 120 * time.Second
	defaultQueryTimeoutMs = 30_000 // channel.QueryTimeoutMs 为 0 时的默认分钟
)

// StartPoller 启动一个 goroutine 定期轮询上游 API 的异步任务
// （即含 upstream_task_id 的 processing 状态任务）。
// 只应在 API 服务器进程中调用。
func StartPoller(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(pollInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				pollPendingTasks(ctx)
				cleanStaleTasks(ctx)
			}
		}
	}()
	log.Println("[poller] started, interval =", pollInterval)
}

// cleanStaleTasks 清理僵尸任务：
//   - pending 状态：NATS 消息丢失（Worker 重启等），任务从未被执行
//   - processing 状态且无 upstream_task_id：Worker 中途崩溃，结果未发布
//
// 超过 5 分钟（默认超时兜底）未完成的此类任务一律标为失败并退款。
func cleanStaleTasks(ctx context.Context) {
	const staleTimeout = 5 * time.Minute
	cutoff := time.Now().Add(-staleTimeout)

	var tasks []model.Task
	err := db.Engine.
		Where("(status = ? OR (status = ? AND upstream_task_id = '')) AND created_at < ?",
			"pending", "processing", cutoff).
		Find(&tasks)
	if err != nil || len(tasks) == 0 {
		return
	}
	for i := range tasks {
		t := &tasks[i]
		log.Printf("[poller] stale task %d (status=%s, age=%s), marking failed", t.ID, t.Status, time.Since(t.CreatedAt).Round(time.Second))
		failTaskDB(ctx, t.ID, t.UserID, t.ChannelID, t.APIKeyID, t.CorrID, t.CreditsCharged,
			"task timed out: worker did not complete within "+staleTimeout.String())
	}
}

func pollPendingTasks(ctx context.Context) {
	var tasks []model.Task
	err := db.Engine.
		Where("status = ? AND upstream_task_id != ''", "processing").
		Find(&tasks)
	if err != nil || len(tasks) == 0 {
		return
	}

	for i := range tasks {
		task := &tasks[i]

		lockKey := fmt.Sprintf("poll:lock:%d", task.ID)
		acquired, lockErr := cache.Client.SetNX(ctx, lockKey, "1", pollLockTTL).Result()
		if lockErr != nil || !acquired {
			continue
		}

		if time.Since(task.CreatedAt) > maxAge {
			cache.Client.Del(ctx, lockKey)
			failTaskDB(ctx, task.ID, task.UserID, task.ChannelID, task.APIKeyID, task.CorrID, task.CreditsCharged,
				"task timed out after "+maxAge.String())
			continue
		}

		ch, err := service.GetChannel(ctx, task.ChannelID)
		if err != nil {
			cache.Client.Del(ctx, lockKey)
			log.Printf("[poller] task %d: channel not found: %v", task.ID, err)
			continue
		}
		if ch.QueryURL == "" {
			cache.Client.Del(ctx, lockKey)
			continue
		}

		go func(t *model.Task, c *model.Channel, lk string) {
			defer cache.Client.Del(ctx, lk)
			pollOneTask(ctx, t, c)
		}(task, ch, lockKey)

	}
}

func pollOneTask(ctx context.Context, task *model.Task, ch *model.Channel) {
	queryURL := strings.ReplaceAll(ch.QueryURL, "{id}", task.UpstreamTaskID)
	// 号池 Key 注入 URL
	var poolKeyValue string
	if ch.KeyPoolID > 0 {
		if pk, pkErr := service.GetOrAssignPoolKey(ctx, ch.KeyPoolID, task.UserID); pkErr == nil && pk != nil {
			poolKeyValue = pk.Value
		}
	}
	if poolKeyValue != "" {
		queryURL = strings.ReplaceAll(queryURL, "{{pool_key}}", poolKeyValue)
	}

	method := ch.QueryMethod
	if method == "" {
		method = "GET"
	}

	qtMs := ch.QueryTimeoutMs
	if qtMs <= 0 {
		qtMs = defaultQueryTimeoutMs
	}
	reqCtx, cancel := context.WithTimeout(ctx, time.Duration(qtMs)*time.Millisecond)
	defer cancel()

	var bodyReader io.Reader
	if method == "POST" {
		b, _ := json.Marshal(map[string]string{"id": task.UpstreamTaskID})
		bodyReader = bytes.NewReader(b)
	}

	httpReq, err := http.NewRequestWithContext(reqCtx, method, queryURL, bodyReader)
	if err != nil {
		log.Printf("[poller] task %d: build request error: %v", task.ID, err)
		return
	}
	if method == "POST" {
		httpReq.Header.Set("Content-Type", "application/json")
	}
	// Header 注入（含 {{pool_key}} 占位符替换）
	poolKeyUsedInHeaders := false
	for k, v := range ch.Headers {
		if sv, ok := v.(string); ok {
			if strings.Contains(sv, "{{pool_key}}") {
				poolKeyUsedInHeaders = true
			}
			httpReq.Header.Set(k, script.ResolveHeaderValue(sv, poolKeyValue))
		}
	}
	// Fallback：Header 里没有 {{pool_key}} 占位符，自动注入 Authorization
	if poolKeyValue != "" && !poolKeyUsedInHeaders && !strings.Contains(ch.QueryURL, "{{pool_key}}") {
		httpReq.Header.Set("Authorization", "Bearer "+poolKeyValue)
	}

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		log.Printf("[poller] task %d: upstream query error: %v", task.ID, err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// 记录本次轮询请求信息（使用与 worker 一致的 _url/_headers 键，方便前端统一展示）
	pollHeaders := make(map[string]interface{})
	for k, v := range ch.Headers {
		pollHeaders[k] = v
	}
	pollHeaders["Content-Type"] = "application/json"
	upstreamReqInfo := model.JSON{"_url": queryURL, "_headers": pollHeaders, "method": method}
	db.Engine.Where("id = ?", task.ID).Cols("upstream_request").
		Update(&model.Task{UpstreamRequest: upstreamReqInfo})

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		errBody := model.JSON{"http_status": resp.StatusCode, "body": string(body)}
		db.Engine.Where("id = ?", task.ID).Cols("upstream_response").
			Update(&model.Task{UpstreamResponse: errBody})
		log.Printf("[poller] task %d: upstream returned %d: %s", task.ID, resp.StatusCode, string(body))
		return
	}

	var rawResp map[string]interface{}
	if err := json.Unmarshal(body, &rawResp); err != nil {
		errBody := model.JSON{"parse_error": err.Error(), "body": string(body)}
		db.Engine.Where("id = ?", task.ID).Cols("upstream_response").
			Update(&model.Task{UpstreamResponse: errBody})
		log.Printf("[poller] task %d: invalid JSON from upstream: %v", task.ID, err)
		return
	}

	// 解析成功后立即写入原始响应，确保脚本报错时管理端也能看到上游返回了什么
	db.Engine.Where("id = ?", task.ID).Cols("upstream_response").
		Update(&model.Task{UpstreamResponse: toJSON(rawResp)})

	mappedResp := rawResp
	if ch.QueryScript != "" {
		mapped, scriptErr := script.RunMapResponse(ch.QueryScript, rawResp)
		if scriptErr != nil {
			log.Printf("[poller] task %d: query_script error: %v", task.ID, scriptErr)
			return // upstream_response 已写入，管理端可看到原始响应排查脚本问题
		}
		mappedResp = mapped
	}

	statusVal := toIntField(mappedResp, "status")
	upstreamResp := toJSON(rawResp)

	// 错误检测
	{
		var detectedErr string
		var isErr bool
		if ch.ErrorScript != "" {
			var scriptErr error
			detectedErr, scriptErr = script.RunCheckError(ch.ErrorScript, mappedResp)
			if scriptErr != nil {
				log.Printf("[poller] task %d: error_script failed: %v", task.ID, scriptErr)
			}
			isErr = detectedErr != ""
		} else {
			detectedErr, isErr = script.DetectUpstreamError(mappedResp)
		}
		if isErr {
			db.Engine.Where("id = ?", task.ID).Cols("upstream_response").
				Update(&model.Task{UpstreamResponse: upstreamResp})
			failTaskDB(ctx, task.ID, task.UserID, task.ChannelID, task.APIKeyID, task.CorrID, task.CreditsCharged, detectedErr)
			return
		}
	}

	switch statusVal {
	case 2: // 成功
		result := toJSON(mappedResp)
		if task.Type == "image" {
			result = convertResultURLs(result, ch.BaseURL)
		}
		db.Engine.Where("id = ?", task.ID).
			Cols("status", "progress", "result", "upstream_response").
			Update(&model.Task{
				Status:           "done",
				Progress:         100,
				Result:           result,
				UpstreamResponse: upstreamResp,
			})
		log.Printf("[poller] task %d done", task.ID)

	case 3: // 失败
		db.Engine.Where("id = ?", task.ID).Cols("upstream_response").
			Update(&model.Task{UpstreamResponse: upstreamResp})
		failMsg := fmt.Sprintf("%v", mappedResp["msg"])
		failTaskDB(ctx, task.ID, task.UserID, task.ChannelID, task.APIKeyID, task.CorrID, task.CreditsCharged,
			"upstream failed: "+failMsg)

	default: // 仍在处理中
		prog := toIntField(mappedResp, "progress")
		db.Engine.Where("id = ?", task.ID).Cols("upstream_response", "progress").
			Update(&model.Task{UpstreamResponse: upstreamResp, Progress: prog})
		log.Printf("[poller] task %d still processing (status=%d, progress=%d)", task.ID, statusVal, prog)
	}
}

// toIntField 从 map 里安全取整数值，兼容 goja 导出的 int64 / float64 / int。
func toIntField(m map[string]interface{}, key string) int {
	v, ok := m[key]
	if !ok {
		return 0
	}
	switch n := v.(type) {
	case int64:
		return int(n)
	case float64:
		return int(n)
	case int:
		return n
	case int32:
		return int(n)
	}
	return 0
}
