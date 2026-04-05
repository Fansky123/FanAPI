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

	// pollLockTTL must be greater than the maximum possible query_timeout_ms so the
	// distributed lock outlives the upstream HTTP call.
	pollLockTTL           = 120 * time.Second
	defaultQueryTimeoutMs = 30_000 // fallback when channel.QueryTimeoutMs == 0
)

// StartPoller starts a goroutine that periodically polls upstream APIs for
// async tasks (status=processing with an upstream_task_id).
// Call this from the API server process only.
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

	mappedResp := rawResp
	if ch.QueryScript != "" {
		mapped, scriptErr := script.RunMapResponse(ch.QueryScript, rawResp)
		if scriptErr != nil {
			log.Printf("[poller] task %d: query_script error: %v", task.ID, scriptErr)
			return
		}
		mappedResp = mapped
	}

	statusVal, _ := mappedResp["status"].(float64)
	upstreamResp := toJSON(rawResp)

	// Error detection
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

	switch int(statusVal) {
	case 2: // success
		result := toJSON(mappedResp)
		db.Engine.Where("id = ?", task.ID).
			Cols("status", "progress", "result", "upstream_response").
			Update(&model.Task{
				Status:           "done",
				Progress:         100,
				Result:           result,
				UpstreamResponse: upstreamResp,
			})
		log.Printf("[poller] task %d done", task.ID)

	case 3: // failed
		db.Engine.Where("id = ?", task.ID).Cols("upstream_response").
			Update(&model.Task{UpstreamResponse: upstreamResp})
		failMsg := fmt.Sprintf("%v", mappedResp["msg"])
		failTaskDB(ctx, task.ID, task.UserID, task.ChannelID, task.APIKeyID, task.CorrID, task.CreditsCharged,
			"upstream failed: "+failMsg)

	default: // still in progress
		log.Printf("[poller] task %d still processing (status=%d)", task.ID, int(statusVal))
	}
}
