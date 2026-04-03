package script

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

	"fanapi/internal/db"
	"fanapi/internal/model"
	"fanapi/internal/service"
)

const (
	pollInterval = 5 * time.Second  // 每次轮询间隔
	pollTimeout  = 30 * time.Second // 单次查询上游超时
	maxAge       = 2 * time.Hour    // 超过此时间仍未完成则标记失败（防止僵死任务）
)

// StartPoller 启动轮询协程，定期检查所有 status=processing 且有 upstream_task_id 的任务。
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
			}
		}
	}()
	log.Println("[poller] started, interval =", pollInterval)
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

		// 超时保护：任务创建超过 maxAge 仍未完成，标记为失败
		if time.Since(task.CreatedAt) > maxAge {
			failTask(task.ID, "task timed out after "+maxAge.String())
			continue
		}

		ch, err := service.GetChannel(ctx, task.ChannelID)
		if err != nil {
			log.Printf("[poller] task %d: channel not found: %v", task.ID, err)
			continue
		}
		if ch.QueryURL == "" {
			// 渠道未配置查询地址，跳过（不应出现，防御性处理）
			continue
		}

		go pollOneTask(ctx, task, ch)
	}
}

func pollOneTask(ctx context.Context, task *model.Task, ch *model.Channel) {
	queryURL := strings.ReplaceAll(ch.QueryURL, "{id}", task.UpstreamTaskID)

	method := ch.QueryMethod
	if method == "" {
		method = "GET"
	}

	reqCtx, cancel := context.WithTimeout(ctx, pollTimeout)
	defer cancel()

	var bodyReader io.Reader
	if method == "POST" {
		// POST 查询时把 upstream_task_id 作为 body 传递
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
	for k, v := range ch.Headers {
		if sv, ok := v.(string); ok {
			httpReq.Header.Set(k, sv)
		}
	}

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		log.Printf("[poller] task %d: upstream query error: %v", task.ID, err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("[poller] task %d: upstream returned %d: %s", task.ID, resp.StatusCode, string(body))
		return
	}

	var rawResp map[string]interface{}
	if err := json.Unmarshal(body, &rawResp); err != nil {
		log.Printf("[poller] task %d: invalid JSON from upstream: %v", task.ID, err)
		return
	}

	// 用 query_script（MapQueryResponse）映射为标准格式；未配置时直接使用原始响应
	mappedResp := rawResp
	if ch.QueryScript != "" {
		mapped, scriptErr := RunMapResponse(ch.QueryScript, rawResp)
		if scriptErr != nil {
			log.Printf("[poller] task %d: query_script error: %v", task.ID, scriptErr)
			return
		}
		mappedResp = mapped
	}

	// 标准格式中 status 字段：2=成功 3=失败，其他值=仍在进行中
	statusVal, _ := mappedResp["status"].(float64)
	upstreamResp := model.JSON{}
	for k, v := range rawResp {
		upstreamResp[k] = v
	}
	switch int(statusVal) {
	case 2: // 成功
		result := model.JSON{}
		for k, v := range mappedResp {
			result[k] = v
		}
		db.Engine.Where("id = ?", task.ID).Cols("status", "progress", "result", "upstream_response").Update(&model.Task{
			Status:           "done",
			Progress:         100,
			Result:           result,
			UpstreamResponse: upstreamResp,
		})
		log.Printf("[poller] task %d done", task.ID)

	case 3: // 失败
		msg := fmt.Sprintf("%v", mappedResp["msg"])
		failTask(task.ID, "upstream failed: "+msg)

	default:
		db.Engine.Where("id = ?", task.ID).Cols("upstream_response").Update(&model.Task{UpstreamResponse: upstreamResp})
		// 仍在处理中，更新进度（如果上游返回了 progress 字段）
		if progress, ok := mappedResp["progress"].(float64); ok {
			db.Engine.Exec(
				"UPDATE tasks SET progress = $1 WHERE id = $2",
				int(progress), task.ID,
			)
		}
	}
}
