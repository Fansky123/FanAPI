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

// StartResultProcessor subscribes to the RESULTS JetStream stream.
// Call this from the API server process only.
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
		result := toJSON(res.Result)
		db.Engine.Where("id = ?", res.TaskID).
			Cols("status", "progress", "result", "upstream_request", "upstream_response").
			Update(&model.Task{
				Status:           "done",
				Progress:         100,
				Result:           result,
				UpstreamRequest:  upstreamReq,
				UpstreamResponse: upstreamResp,
			})

	case model.OutcomeAsync:
		db.Engine.Where("id = ?", res.TaskID).
			Cols("status", "upstream_task_id", "upstream_request", "upstream_response").
			Update(&model.Task{
				Status:           "processing",
				UpstreamTaskID:   res.UpstreamTaskID,
				UpstreamRequest:  upstreamReq,
				UpstreamResponse: upstreamResp,
			})
		log.Printf("[result-proc] task %d async, upstream_task_id=%s", res.TaskID, res.UpstreamTaskID)

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

// saveAndFail writes upstream fields and fails the task in one shot.
func saveAndFail(ctx context.Context, res model.WorkerResult, req, resp model.JSON, msg string) {
	if len(req) > 0 || len(resp) > 0 {
		db.Engine.Where("id = ?", res.TaskID).
			Cols("upstream_request", "upstream_response").
			Update(&model.Task{UpstreamRequest: req, UpstreamResponse: resp})
	}
	failTaskDB(ctx, res.TaskID, res.UserID, res.ChannelID, res.APIKeyID, res.CorrID, res.CreditsCharged, msg)
}

// failTaskDB marks a task as failed and refunds credits.
// Idempotent: guarded by conditional UPDATE (status != 'failed').
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
	if err := billing.Refund(ctx, userID, credits); err != nil {
		log.Printf("[result-proc] task %d: refund failed: %v", taskID, err)
		return
	}
	_ = service.WriteTx(ctx, userID, channelID, apiKeyID, corrID, "refund", credits, 0, model.JSON{
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
