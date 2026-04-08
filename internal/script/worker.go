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

	"fanapi/internal/config"
	"fanapi/internal/model"
	"fanapi/internal/mq"

	"github.com/nats-io/nats.go"
)

// StartWorkers 根据 WorkerConfig 订阅 NATS 任务主题。
//
// 默认（未配置）：订阅 "task.>"  ，consumer 名为 "workers-all"。
// 专用 Worker 示例（在 config.yaml 中添加）：
//
//	worker:
//	  subjects:
//	    - "task.video.*"
//	    - "task.audio.*"
func StartWorkers(cfg config.WorkerConfig) error {
	// 清理上次运行遗留的失效 Consumer，再进行订阅。
	// 只应在 Worker 进程中运行——如在服务器进程中运行会杀死服务器的 result-proc Consumer。
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

func subjectToConsumer(subject string) string {
	s := strings.TrimPrefix(subject, "task.")
	s = strings.TrimSuffix(s, ".*")
	s = strings.ReplaceAll(s, ".", "-")
	s = strings.ReplaceAll(s, ">", "all")
	s = strings.ReplaceAll(s, "*", "any")
	return "workers-" + s
}

func handleTask(msg *nats.Msg) {
	var job model.TaskJob
	if err := json.Unmarshal(msg.Data, &job); err != nil {
		log.Printf("[worker] bad message: %v", err)
		_ = msg.Term()
		return
	}

	result := execJob(context.Background(), &job)

	// 先发布结果再 ACK——若发布失败则消息会被重新投递，Worker 将重试（奖励幂等）。
	subject := fmt.Sprintf("result.%d", job.TaskID)
	data, _ := json.Marshal(result)
	if err := mq.PublishResult(subject, data); err != nil {
		log.Printf("[worker] task %d: failed to publish result: %v", job.TaskID, err)
		// Do NOT ack — let the message be redelivered.
		return
	}
	_ = msg.Ack()
}

// execJob 执行一个 TaskJob 并返回 WorkerResult，永不返回 nil。
func execJob(ctx context.Context, job *model.TaskJob) *model.WorkerResult {
	base := &model.WorkerResult{
		TaskID:         job.TaskID,
		TaskType:       job.TaskType,
		UserID:         job.UserID,
		APIKeyID:       job.APIKeyID,
		CorrID:         job.CorrID,
		CreditsCharged: job.CreditsCharged,
		ChannelID:      job.ChannelID,
		PoolKeyID:      job.PoolKeyID,
		RetryCount:     job.RetryCount,
		Payload:        job.Payload, // 保留下来以便服务器在 OutcomeRateLimited 时重新发布
	}

	fail := func(msg string) *model.WorkerResult {
		base.Outcome = model.OutcomeFailed
		base.ErrorMsg = msg
		return base
	}

	// 应用 request_script
	payload := job.Payload
	if job.RequestScript != "" {
		mapped, err := RunMapRequest(job.RequestScript, payload)
		if err != nil {
			return fail("request mapping error: " + err.Error())
		}
		payload = mapped
	}

	// 记录上游请求（调试用）
	upstreamReq := make(map[string]interface{})
	for k, v := range payload {
		upstreamReq[k] = v
	}
	base.UpstreamRequest = upstreamReq

	// 调用上游 HTTP
	respData, statusCode, err := callUpstream(job, payload)
	if err != nil {
		return fail("upstream error: " + err.Error())
	}

	// 429: report rate_limited so server can rotate key and retry (once)
	if statusCode == http.StatusTooManyRequests {
		if job.PoolKeyID > 0 && job.RetryCount < 1 {
			base.Outcome = model.OutcomeRateLimited
			return base
		}
		return fail("upstream rate limited")
	}

	upstreamResp := make(map[string]interface{})
	for k, v := range respData {
		upstreamResp[k] = v
	}
	base.UpstreamResponse = upstreamResp

	// 应用 response_script
	if job.ResponseScript != "" {
		mapped, err := RunMapResponse(job.ResponseScript, respData)
		if err != nil {
			return fail("response mapping error: " + err.Error())
		}
		respData = mapped
	}

	// 检查是否有异步上游任务 ID
	upstreamTaskID, _ := respData["upstream_task_id"].(string)
	if upstreamTaskID == "" {
		if v, ok := respData["id"].(string); ok && job.QueryURL != "" {
			upstreamTaskID = v
		}
	}
	if upstreamTaskID != "" {
		base.Outcome = model.OutcomeAsync
		base.UpstreamTaskID = upstreamTaskID
		return base
	}

	// 错误检测（error_script 或内置识别逻辑）
	errMsg, isErr := "", false
	if job.ErrorScript != "" {
		var scriptErr error
		errMsg, scriptErr = RunCheckError(job.ErrorScript, respData)
		if scriptErr != nil {
			log.Printf("[worker] task %d: error_script failed: %v", job.TaskID, scriptErr)
		}
		isErr = errMsg != ""
	} else {
		errMsg, isErr = DetectUpstreamError(respData)
	}
	if isErr {
		return fail(errMsg)
	}

	// response_script 返回 status=3 表示业务失败
	if statusVal, _ := respData["status"].(float64); int(statusVal) == 3 {
		return fail("upstream failed: " + fmt.Sprintf("%v", respData["msg"]))
	}

	result := make(map[string]interface{})
	for k, v := range respData {
		result[k] = v
	}
	base.Outcome = model.OutcomeDone
	base.Result = result
	return base
}

func callUpstream(job *model.TaskJob, payload map[string]interface{}) (map[string]interface{}, int, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, err
	}

	timeout := time.Duration(job.TimeoutMs) * time.Millisecond
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	client := &http.Client{Timeout: timeout}

	req, err := http.NewRequest(job.Method, job.BaseURL, bytes.NewReader(body))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range job.Headers {
		if sv, ok := v.(string); ok {
			req.Header.Set(k, ResolveHeaderValue(sv))
		}
	}
	if job.PoolKeyValue != "" {
		req.Header.Set("Authorization", "Bearer "+job.PoolKeyValue)
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

// DetectUpstreamError 检测常见厂商错误响应格式。
// 如识别到错误，返回错误消息和 true。
//
// 支持的格式：
//   - OpenAI / 通用：{"error": {"message": "...", "code": "..."}}
//   - 字符串错误：{"error": "some message"}
//   - 自定义 code+msg：{"code": "InvalidParameter", "message": "..."}
func DetectUpstreamError(resp map[string]interface{}) (string, bool) {
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

	codeVal, hasCode := resp["code"]
	msgStr, _ := resp["message"].(string)
	if hasCode && msgStr != "" {
		switch c := codeVal.(type) {
		case string:
			if c != "" {
				return c + ": " + msgStr, true
			}
		case float64:
			if c < 200 || c >= 300 {
				return fmt.Sprintf("code %d: %s", int(c), msgStr), true
			}
		}
	}

	return "", false
}
