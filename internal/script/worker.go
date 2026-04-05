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

	"fanapi/internal/billing"
	"fanapi/internal/config"
	"fanapi/internal/db"
	"fanapi/internal/model"
	"fanapi/internal/mq"
	"fanapi/internal/service"

	"github.com/nats-io/nats.go"
)

type taskRequest struct {
	TaskID    int64                  `json:"task_id"`
	ChannelID int64                  `json:"channel_id"`
	UserID    int64                  `json:"user_id"`
	Payload   map[string]interface{} `json:"payload"`
}

// StartWorkers subscribes to NATS task subjects based on WorkerConfig.
//
// Default (no config): subscribes to "task.>" with consumer "workers-all".
// Example specialised worker (add to config.yaml):
//
//	worker:
//	  subjects:
//	    - "task.video.*"   # consumer: workers-video
//	    - "task.audio.*"   # consumer: workers-audio
//
// Each subject gets its own durable consumer so multiple specialised workers
// can coexist on the NATS stream without conflict.
func StartWorkers(cfg config.WorkerConfig) error {
	// Remove stale consumers from previous runs before subscribing.
	// Must only run in the worker process (not in the server),
	// otherwise the server restart would kill the worker's active consumer.
	mq.PurgeConsumers()

	subjects := cfg.Subjects
	if len(subjects) == 0 {
		subjects = []string{"task.>"}
	}
	for _, subj := range subjects {
		consumer := subjectToConsumer(subj)
		if _, err := mq.QueueSubscribe(subj, consumer, handleTask); err != nil {
			return fmt.Errorf("subscribe %s: %w", subj, err)
		}
		log.Printf("[script worker] subscribed to %s (consumer: %s)", subj, consumer)
	}
	return nil
}

// subjectToConsumer derives a stable, JetStream-safe consumer name from a subject.
// Examples:
//
//	"task.>"       → "workers-all"
//	"task.image.*" → "workers-image"
//	"task.video.*" → "workers-video"
func subjectToConsumer(subject string) string {
	s := strings.TrimPrefix(subject, "task.")
	s = strings.TrimSuffix(s, ".*")
	s = strings.ReplaceAll(s, ".", "-")
	s = strings.ReplaceAll(s, ">", "all")
	s = strings.ReplaceAll(s, "*", "any")
	return "workers-" + s
}

func handleTask(msg *nats.Msg) {
	var req taskRequest
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		log.Printf("[worker] bad message: %v", err)
		// Malformed message: terminate immediately, no retry.
		_ = msg.Term()
		return
	}

	ctx := context.Background()

	// Load task
	task := &model.Task{}
	found, err := db.Engine.Where("id = ?", req.TaskID).Get(task)
	if err != nil || !found {
		log.Printf("[worker] task %d not found: %v", req.TaskID, err)
		_ = msg.Term()
		return
	}

	// Idempotency: if the task was already processed (e.g. message redelivered after a crash),
	// just ACK and exit without re-processing.
	if task.Status == "done" || task.Status == "failed" {
		_ = msg.Ack()
		return
	}

	// Mark processing
	db.Engine.Where("id = ?", task.ID).Update(&model.Task{Status: "processing", Progress: 0})

	// Load channel (uses Redis cache via service)
	ch, err := service.GetChannel(ctx, req.ChannelID)
	if err != nil {
		failTask(task.ID, "channel not found: "+err.Error())
		_ = msg.Term()
		return
	}

	// Map request
	payload := req.Payload
	if ch.RequestScript != "" {
		mapped, scriptErr := RunMapRequest(ch.RequestScript, payload)
		if scriptErr != nil {
			failTask(task.ID, "request mapping error: "+scriptErr.Error())
			_ = msg.Term()
			return
		}
		payload = mapped
	}

	// 号池：为该用户选取 Sticky 分配的三方 Key
	var poolKey *model.PoolKey
	if ch.KeyPoolID > 0 {
		pk, pkErr := service.GetOrAssignPoolKey(ctx, ch.KeyPoolID, req.UserID)
		if pkErr != nil {
			failTask(task.ID, "key pool error: "+pkErr.Error())
			_ = msg.Term()
			return
		}
		poolKey = pk
	}

	// upstream_request 在调用上游前就落库，确保任意失败路径下日志可查
	upstreamReq := model.JSON{}
	for k, v := range payload {
		upstreamReq[k] = v
	}
	db.Engine.Where("id = ?", task.ID).Cols("upstream_request").Update(&model.Task{
		UpstreamRequest: upstreamReq,
	})

	// Call third-party API
	respData, statusCode, err := callUpstream(ch, payload, poolKey)
	if err != nil {
		failTask(task.ID, "upstream error: "+err.Error())
		_ = msg.Term()
		return
	}

	// 若 429，号池轮转并重试一次
	if statusCode == http.StatusTooManyRequests && ch.KeyPoolID > 0 && poolKey != nil {
		newKey, rotErr := service.MarkExhaustedAndRotate(ctx, ch.KeyPoolID, poolKey.ID, req.UserID)
		if rotErr == nil {
			poolKey = newKey
			respData, _, err = callUpstream(ch, payload, poolKey)
			if err != nil {
				failTask(task.ID, "upstream error (retry): "+err.Error())
				_ = msg.Term()
				return
			}
		}
	}

	upstreamResp := model.JSON{}
	for k, v := range respData {
		upstreamResp[k] = v
	}
	db.Engine.Where("id = ?", task.ID).Cols("upstream_response").Update(&model.Task{
		UpstreamResponse: upstreamResp,
	})

	// Map response：执行 response_script 将 vendor 原始响应转换为平台标准格式
	// 标准格式（同步）：{"code":200, "url":"...", "status":2, "msg":""}
	// 标准格式（异步）：{"upstream_task_id":"xxx"} —— 表示第三方为异步接口，
	//   worker 只保存 upstream_task_id，由 poller 定期轮询最终状态
	if ch.ResponseScript != "" {
		mapped, scriptErr := RunMapResponse(ch.ResponseScript, respData)
		if scriptErr != nil {
			failTask(task.ID, "response mapping error: "+scriptErr.Error())
			_ = msg.Term()
			return
		}
		respData = mapped
	}

	// 判断是否为异步渠道：渠道配置了 QueryURL 或响应中含 upstream_task_id
	upstreamTaskID, _ := respData["upstream_task_id"].(string)
	if upstreamTaskID == "" {
		if v, ok := respData["id"].(string); ok && ch.QueryURL != "" {
			upstreamTaskID = v
		}
	}

	if upstreamTaskID != "" {
		// 异步模式：保存第三方任务 ID，等待 poller 轮询
		db.Engine.Where("id = ?", task.ID).Cols("status", "upstream_task_id", "upstream_request", "upstream_response").Update(&model.Task{
			Status:           "processing",
			UpstreamTaskID:   upstreamTaskID,
			UpstreamRequest:  upstreamReq,
			UpstreamResponse: upstreamResp,
		})
		log.Printf("[worker] task %d is async, upstream_task_id=%s", task.ID, upstreamTaskID)
		_ = msg.Ack()
		return
	}

	// 同步模式：将标准格式结果写入 task.result
	// 错误检测顺序：先跐 ch.ErrorScript（肌道可配置），无配置时降级到内置通用格式检测。
	// 这两个检测都在 response_script 映射之后的 respData 上运行。
	errMsg, isErr := "", false
	if ch.ErrorScript != "" {
		var scriptErr error
		errMsg, scriptErr = RunCheckError(ch.ErrorScript, respData)
		if scriptErr != nil {
			log.Printf("[worker] task %d: error_script failed: %v", task.ID, scriptErr)
		}
		isErr = errMsg != ""
	} else {
		errMsg, isErr = detectUpstreamError(respData)
	}
	if isErr {
		db.Engine.Where("id = ?", task.ID).Cols("upstream_request", "upstream_response").Update(&model.Task{
			UpstreamRequest:  upstreamReq,
			UpstreamResponse: upstreamResp,
		})
		failTask(task.ID, errMsg)
		_ = msg.Ack()
		return
	}
	// response_script 标准出参 status=3 表示业务失败
	if statusVal, _ := respData["status"].(float64); int(statusVal) == 3 {
		failMsg := fmt.Sprintf("%v", respData["msg"])
		// 先保存 upstream_response，再调 failTask（failTask 只改 status/error_msg）
		db.Engine.Where("id = ?", task.ID).Cols("upstream_request", "upstream_response").Update(&model.Task{
			UpstreamRequest:  upstreamReq,
			UpstreamResponse: upstreamResp,
		})
		failTask(task.ID, "upstream failed: "+failMsg)
		_ = msg.Ack()
		return
	}

	result := model.JSON{}
	for k, v := range respData {
		result[k] = v
	}
	db.Engine.Where("id = ?", task.ID).Cols("status", "progress", "result", "upstream_request", "upstream_response").Update(&model.Task{
		Status:           "done",
		Progress:         100,
		Result:           result,
		UpstreamRequest:  upstreamReq,
		UpstreamResponse: upstreamResp,
	})
	_ = msg.Ack()
}

func callUpstream(ch *model.Channel, payload map[string]interface{}, poolKey *model.PoolKey) (map[string]interface{}, int, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, err
	}

	timeout := time.Duration(ch.TimeoutMs) * time.Millisecond
	client := &http.Client{Timeout: timeout}

	req, err := http.NewRequest(ch.Method, ch.BaseURL, bytes.NewReader(body))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range ch.Headers {
		if sv, ok := v.(string); ok {
			req.Header.Set(k, sv)
		}
	}
	// 号池 Key 覆盖渠道静态 Authorization
	if poolKey != nil {
		req.Header.Set("Authorization", "Bearer "+poolKey.Value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, resp.StatusCode, nil
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, resp.StatusCode, fmt.Errorf("upstream returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, resp.StatusCode, fmt.Errorf("upstream response not JSON: %w", err)
	}
	return result, resp.StatusCode, nil
}

// failTask 将任务标记为失败并退还已扣的 credits。
// 使用条件 UPDATE（status != 'failed'）保证幂等：即使消息重投也不会双重退款。
func failTask(taskID int64, msg string) {
	log.Printf("[worker] task %d failed: %s", taskID, msg)

	// 原子地将状态从非-failed 改为 failed，并返回受影响行数。
	// n==0 表示任务不存在或已经是 failed 状态，直接跳出防止重复退款。
	n, _ := db.Engine.
		Where("id = ? AND status != ?", taskID, "failed").
		Cols("status", "error_msg").
		Update(&model.Task{Status: "failed", ErrorMsg: msg})
	if n == 0 {
		return
	}

	// 加载退款所需字段
	ctx := context.Background()
	task := &model.Task{}
	found, err := db.Engine.
		Where("id = ?", taskID).
		Cols("user_id", "channel_id", "api_key_id", "credits_charged", "corr_id").
		Get(task)
	if err != nil || !found || task.CreditsCharged == 0 {
		return // 无需退款（未扣费或加载失败）
	}

	if refundErr := billing.Refund(ctx, task.UserID, task.CreditsCharged); refundErr != nil {
		log.Printf("[worker] task %d: refund %d credits failed: %v", taskID, task.CreditsCharged, refundErr)
		return
	}
	_ = service.WriteTx(ctx, task.UserID, task.ChannelID, task.APIKeyID, task.CorrID, "refund", task.CreditsCharged, 0, model.JSON{
		"task_id": taskID,
		"reason":  msg,
	})
	log.Printf("[worker] task %d: refunded %d credits to user %d", taskID, task.CreditsCharged, task.UserID)
}

// detectUpstreamError 检测常见的厂商错误响应格式，返回错误描述和 true。
// 若响应看起来正常则返回 ("", false)。
//
// 覆盖的格式：
//   - OpenAI / 通用：{"error": {"message": "...", "code": "..."}}
//   - 字符串错误：{"error": "some message"}
//   - 自定义 code+message：{"code": "InvalidParameter", "message": "..."}
//     （仅当 code 不是 2xx 数字且不为空时才视为错误，避免误判正常 status code）
func detectUpstreamError(resp map[string]interface{}) (string, bool) {
	// --- 模式1：顶层 "error" 字段 ---
	if errVal, ok := resp["error"]; ok && errVal != nil {
		switch e := errVal.(type) {
		case map[string]interface{}:
			msg, _ := e["message"].(string)
			code, _ := e["code"].(string)
			switch {
			case code != "" && msg != "":
				return code + ": " + msg, true
			case msg != "":
				return msg, true
			case code != "":
				return code, true
			}
		case string:
			if e != "" {
				return e, true
			}
		}
		return "upstream returned error", true
	}

	// --- 模式2：顶层字符串 "code" + "message"（code 不是 2xx 且不为整数成功码）---
	// 只有当 code 为字符串（非数字 2xx）时才判定为错误，避免误判 {"code":200, "url":"..."}
	codeVal, hasCode := resp["code"]
	msgStr, _ := resp["message"].(string)
	if hasCode && msgStr != "" {
		switch c := codeVal.(type) {
		case string:
			// 字符串 code 一定不是 HTTP 成功码，视为错误
			if c != "" {
				return c + ": " + msgStr, true
			}
		case float64:
			// 数字 code：仅非2xx才报错（200/201等跳过）
			if c < 200 || c >= 300 {
				return fmt.Sprintf("code %d: %s", int(c), msgStr), true
			}
		}
	}

	return "", false
}
