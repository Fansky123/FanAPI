package script

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

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

// StartWorkers 订阅 NATS 的 task.image.*、task.video.*、task.audio.* 三类主题，
// 使用 queue group "script-workers" 保证同一任务只被一个 worker 处理。
func StartWorkers() error {
	subjects := []string{"task.image.*", "task.video.*", "task.audio.*"}
	for _, subj := range subjects {
		if _, err := mq.QueueSubscribe(subj, "script-workers", handleTask); err != nil {
			return fmt.Errorf("subscribe %s: %w", subj, err)
		}
	}
	log.Println("[script worker] subscribed to task.image.*, task.video.*, task.audio.*")
	return nil
}

func handleTask(msg *nats.Msg) {
	var req taskRequest
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		log.Printf("[worker] bad message: %v", err)
		return
	}

	ctx := context.Background()

	// Load task
	task := &model.Task{}
	found, err := db.Engine.Where("id = ?", req.TaskID).Get(task)
	if err != nil || !found {
		log.Printf("[worker] task %d not found: %v", req.TaskID, err)
		return
	}

	// Mark processing
	db.Engine.Where("id = ?", task.ID).Update(&model.Task{Status: "processing", Progress: 0})

	// Load channel (uses Redis cache via service)
	ch, err := service.GetChannel(ctx, req.ChannelID)
	if err != nil {
		failTask(task.ID, "channel not found: "+err.Error())
		return
	}

	// Map request
	payload := req.Payload
	if ch.RequestScript != "" {
		mapped, scriptErr := RunMapRequest(ch.RequestScript, payload)
		if scriptErr != nil {
			failTask(task.ID, "request mapping error: "+scriptErr.Error())
			return
		}
		payload = mapped
	}

	// Call third-party API
	respData, err := callUpstream(ch, payload)
	if err != nil {
		failTask(task.ID, "upstream error: "+err.Error())
		return
	}

	upstreamReq := model.JSON{}
	for k, v := range payload {
		upstreamReq[k] = v
	}
	upstreamResp := model.JSON{}
	for k, v := range respData {
		upstreamResp[k] = v
	}
	db.Engine.Where("id = ?", task.ID).Cols("upstream_request", "upstream_response").Update(&model.Task{
		UpstreamRequest:  upstreamReq,
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
		return
	}

	// 同步模式：将标准格式结果写入 task.result
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
}

func callUpstream(ch *model.Channel, payload map[string]interface{}) (map[string]interface{}, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	timeout := time.Duration(ch.TimeoutMs) * time.Millisecond
	client := &http.Client{Timeout: timeout}

	req, err := http.NewRequest(ch.Method, ch.BaseURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range ch.Headers {
		if sv, ok := v.(string); ok {
			req.Header.Set(k, sv)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("upstream returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("upstream response not JSON: %w", err)
	}
	return result, nil
}

// failTask 将任务标记为失败，同时在 result 中写入标准错误格式，
// 方便 handler 统一处理（也可以直接读 error_msg，但保持 result 有值更一致）。
func failTask(taskID int64, msg string) {
	log.Printf("[worker] task %d failed: %s", taskID, msg)
	db.Engine.Where("id = ?", taskID).Update(&model.Task{
		Status:   "failed",
		ErrorMsg: msg,
	})
}
