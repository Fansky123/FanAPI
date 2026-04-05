package taskresult

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"fanapi/internal/db"
	"fanapi/internal/model"

	"github.com/nats-io/nats.go"
)

const (
	writerBatchSize     = 500
	writerFlushInterval = 50 * time.Millisecond
	writerChanCap       = 20000
)

type doneItem struct {
	msg            *nats.Msg
	taskID         int64
	status         string // "done" | "processing" (async)
	progress       int
	result         model.JSON
	upstreamReq    model.JSON
	upstreamResp   model.JSON
	upstreamTaskID string // non-empty for async
}

var writeCh chan doneItem

// StartBatchWriter starts the background goroutine that batches task status
// updates into PostgreSQL. Call once from the server process.
func StartBatchWriter(ctx context.Context) {
	writeCh = make(chan doneItem, writerChanCap)
	go runWriter(ctx)
	log.Printf("[result-writer] started (batch=%d flush=%s)", writerBatchSize, writerFlushInterval)
}

func enqueueDoneUpdate(item doneItem) {
	select {
	case writeCh <- item:
	default:
		// Channel full — write immediately to avoid blocking the NATS handler goroutine.
		log.Printf("[result-writer] channel full, flushing task %d immediately", item.taskID)
		flushBatch([]doneItem{item})
	}
}

func runWriter(ctx context.Context) {
	ticker := time.NewTicker(writerFlushInterval)
	defer ticker.Stop()
	batch := make([]doneItem, 0, writerBatchSize)

	for {
		select {
		case item := <-writeCh:
			batch = append(batch, item)
			if len(batch) >= writerBatchSize {
				flushBatch(batch)
				batch = batch[:0]
			}

		case <-ticker.C:
			if len(batch) > 0 {
				flushBatch(batch)
				batch = batch[:0]
			}

		case <-ctx.Done():
			// Drain whatever is already in the channel before exiting.
		drain:
			for {
				select {
				case item := <-writeCh:
					batch = append(batch, item)
				default:
					break drain
				}
			}
			if len(batch) > 0 {
				flushBatch(batch)
			}
			return
		}
	}
}

func flushBatch(items []doneItem) {
	if len(items) == 0 {
		return
	}
	var doneItems, asyncItems []doneItem
	for _, item := range items {
		if item.status == "done" {
			doneItems = append(doneItems, item)
		} else {
			asyncItems = append(asyncItems, item)
		}
	}
	if len(doneItems) > 0 {
		batchUpdateDone(doneItems)
	}
	if len(asyncItems) > 0 {
		batchUpdateAsync(asyncItems)
	}
	// ACK all NATS messages after DB writes complete.
	for _, item := range items {
		_ = item.msg.Ack()
	}
}

// batchUpdateDone issues a single UPDATE … FROM (VALUES …) for all done tasks.
func batchUpdateDone(items []doneItem) {
	args := make([]interface{}, 0, len(items)*6)
	placeholders := make([]string, 0, len(items))
	for i, item := range items {
		n := i * 6
		placeholders = append(placeholders, fmt.Sprintf(
			"($%d::bigint,$%d,$%d::int,$%d::jsonb,$%d::jsonb,$%d::jsonb)",
			n+1, n+2, n+3, n+4, n+5, n+6,
		))
		args = append(args,
			item.taskID,
			item.status,
			item.progress,
			marshalJSON(item.result),
			marshalJSON(item.upstreamReq),
			marshalJSON(item.upstreamResp),
		)
	}
	query := fmt.Sprintf(`
UPDATE tasks AS t SET
    status            = v.status,
    progress          = v.progress,
    result            = v.result,
    upstream_request  = v.upstream_request,
    upstream_response = v.upstream_response
FROM (VALUES %s) AS v(id, status, progress, result, upstream_request, upstream_response)
WHERE t.id = v.id`, strings.Join(placeholders, ","))

	execArgs := make([]interface{}, 0, 1+len(args))
	execArgs = append(execArgs, query)
	execArgs = append(execArgs, args...)
	if _, err := db.Engine.Exec(execArgs...); err != nil {
		log.Printf("[result-writer] batch done update (%d rows): %v", len(items), err)
	}
}

// batchUpdateAsync issues a single UPDATE … FROM (VALUES …) for all async tasks.
func batchUpdateAsync(items []doneItem) {
	args := make([]interface{}, 0, len(items)*4)
	placeholders := make([]string, 0, len(items))
	for i, item := range items {
		n := i * 4
		placeholders = append(placeholders, fmt.Sprintf(
			"($%d::bigint,$%d,$%d,$%d::jsonb)",
			n+1, n+2, n+3, n+4,
		))
		args = append(args,
			item.taskID,
			"processing",
			item.upstreamTaskID,
			marshalJSON(item.upstreamReq),
		)
	}
	query := fmt.Sprintf(`
UPDATE tasks AS t SET
    status           = v.status,
    upstream_task_id = v.upstream_task_id,
    upstream_request = v.upstream_request
FROM (VALUES %s) AS v(id, status, upstream_task_id, upstream_request)
WHERE t.id = v.id`, strings.Join(placeholders, ","))

	execArgs := make([]interface{}, 0, 1+len(args))
	execArgs = append(execArgs, query)
	execArgs = append(execArgs, args...)
	if _, err := db.Engine.Exec(execArgs...); err != nil {
		log.Printf("[result-writer] batch async update (%d rows): %v", len(items), err)
	}
}

func marshalJSON(j model.JSON) string {
	if len(j) == 0 {
		return "{}"
	}
	b, _ := json.Marshal(j)
	return string(b)
}
