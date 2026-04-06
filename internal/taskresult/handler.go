package taskresult

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"fanapi/internal/billing"
	"fanapi/internal/config"
	"fanapi/internal/db"
	"fanapi/internal/model"
	"fanapi/internal/mq"
	"fanapi/internal/service"

	"github.com/nats-io/nats.go"
)

// StartResultProcessor 订阅 RESULTS JetStream 流。
// 只应在 API 服务器进程中调用。
func StartResultProcessor(_ config.WorkerConfig) error {
	if _, err := mq.QueueSubscribe("result.>", "result-proc", handleResult); err != nil {
		return fmt.Errorf("subscribe results: %w", err)
	}
	log.Println("[result-proc] subscribed to result.>")
	return nil
}

func handleResult(msg *nats.Msg) {
	var res model.WorkerResult
	if err := json.Unmarshal(msg.Data, &res); err != nil {
		log.Printf("[result-proc] bad message: %v", err)
		_ = msg.Term()
		return
	}

	ctx := context.Background()

	upstreamReq := toJSON(res.UpstreamRequest)
	upstreamResp := toJSON(res.UpstreamResponse)

	switch res.Outcome {

	case model.OutcomeDone:
		enqueueDoneUpdate(doneItem{
			msg:          msg,
			taskID:       res.TaskID,
			status:       "done",
			progress:     100,
			result:       toJSON(res.Result),
			upstreamReq:  upstreamReq,
			upstreamResp: upstreamResp,
		})
		return // ACK 由批量写入器处理

	case model.OutcomeAsync:
		enqueueDoneUpdate(doneItem{
			msg:            msg,
			taskID:         res.TaskID,
			status:         "processing",
			upstreamTaskID: res.UpstreamTaskID,
			upstreamReq:    upstreamReq,
		})
		log.Printf("[result-proc] task %d async, upstream_task_id=%s", res.TaskID, res.UpstreamTaskID)
		return // ACK 由批量写入器处理

	case model.OutcomeRateLimited:
		if res.RetryCount >= 1 {
			saveAndFail(ctx, res, upstreamReq, upstreamResp, "upstream rate limited after retry")
			_ = msg.Ack()
			return
		}
		ch, err := service.GetChannel(ctx, res.ChannelID)
		if err != nil {
			saveAndFail(ctx, res, upstreamReq, upstreamResp, "rate limited + channel load failed: "+err.Error())
			_ = msg.Ack()
			return
		}
		newKey, err := service.MarkExhaustedAndRotate(ctx, ch.KeyPoolID, res.PoolKeyID, res.UserID)
		if err != nil || newKey == nil {
			saveAndFail(ctx, res, upstreamReq, upstreamResp, "rate limited, key rotation failed: "+fmt.Sprint(err))
			_ = msg.Ack()
			return
		}
		job := &model.TaskJob{
			TaskID:         res.TaskID,
			TaskType:       res.TaskType,
			UserID:         res.UserID,
			APIKeyID:       res.APIKeyID,
			CorrID:         res.CorrID,
			CreditsCharged: res.CreditsCharged,
			ChannelID:      res.ChannelID,
			BaseURL:        ch.BaseURL,
			Method:         ch.Method,
			Headers:        ch.Headers,
			TimeoutMs:      ch.TimeoutMs,
			QueryTimeoutMs: ch.QueryTimeoutMs,
			RequestScript:  ch.RequestScript,
			ResponseScript: ch.ResponseScript,
			ErrorScript:    ch.ErrorScript,
			QueryURL:       ch.QueryURL,
			QueryMethod:    ch.QueryMethod,
			QueryScript:    ch.QueryScript,
			PoolKeyID:      newKey.ID,
			PoolKeyValue:   newKey.Value,
			Payload:        res.Payload,
			RetryCount:     res.RetryCount + 1,
		}
		data, _ := json.Marshal(job)
		subject := fmt.Sprintf("task.%s.%d", res.TaskType, res.ChannelID)
		if pubErr := mq.Publish(subject, data); pubErr != nil {
			saveAndFail(ctx, res, upstreamReq, upstreamResp, "rate limited, retry publish failed")
		}
		_ = msg.Ack()
		return

	case model.OutcomeFailed:
		if len(upstreamReq) > 0 || len(upstreamResp) > 0 {
			db.Engine.Where("id = ?", res.TaskID).
				Cols("upstream_request", "upstream_response").
				Update(&model.Task{UpstreamRequest: upstreamReq, UpstreamResponse: upstreamResp})
		}
		failTaskDB(ctx, res.TaskID, res.UserID, res.ChannelID, res.APIKeyID, res.CorrID, res.CreditsCharged, res.ErrorMsg)

	default:
		log.Printf("[result-proc] unknown outcome %q for task %d", res.Outcome, res.TaskID)
	}

	_ = msg.Ack()
}

// saveAndFail 一次性写入上游字段并将任务标记为失败。
func saveAndFail(ctx context.Context, res model.WorkerResult, req, resp model.JSON, msg string) {
	if len(req) > 0 || len(resp) > 0 {
		db.Engine.Where("id = ?", res.TaskID).
			Cols("upstream_request", "upstream_response").
			Update(&model.Task{UpstreamRequest: req, UpstreamResponse: resp})
	}
	failTaskDB(ctx, res.TaskID, res.UserID, res.ChannelID, res.APIKeyID, res.CorrID, res.CreditsCharged, msg)
}

// failTaskDB 将任务标记为失败并退还 credits。
// 幂等操作：通过条件更新 (status != 'failed') 保持幂等。
func failTaskDB(ctx context.Context, taskID, userID, channelID, apiKeyID int64, corrID string, credits int64, errMsg string) {
	log.Printf("[result-proc] task %d failed: %s", taskID, errMsg)
	n, _ := db.Engine.
		Where("id = ? AND status != ?", taskID, "failed").
		Cols("status", "error_msg").
		Update(&model.Task{Status: "failed", ErrorMsg: errMsg})
	if n == 0 {
		return
	}
	if credits <= 0 {
		return
	}

	// 从原收费流水中查询上游成本，使退款流水与收费流水担载相同成本，
	// 从而保持利润分析中（收费 + 退款）净额为零。
	var chargeTx model.BillingTransaction
	upstreamCost := int64(0)
	if found, _ := db.Engine.Where("corr_id = ? AND type = ?", corrID, "charge").Get(&chargeTx); found {
		upstreamCost = chargeTx.Cost
	}

	if err := billing.Refund(ctx, userID, credits); err != nil {
		log.Printf("[result-proc] task %d: refund (Redis) failed: %v — proceeding to update DB", taskID, err)
	}
	_ = service.WriteTx(ctx, userID, channelID, apiKeyID, corrID, "refund", credits, upstreamCost, model.JSON{
		"task_id": taskID,
		"reason":  errMsg,
	})
	log.Printf("[result-proc] task %d: refunded %d credits to user %d", taskID, credits, userID)
}

func toJSON(m map[string]interface{}) model.JSON {
	j := model.JSON{}
	for k, v := range m {
		j[k] = v
	}
	return j
}
