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

// StartWorkers subscribes to NATS task subjects based on WorkerConfig.
//
// Default (no config): subscribes to "task.>" with consumer "workers-all".
// Example specialised workers (add to config.yaml):
//
//worker:
//  subjects:
//    - "task.video.*"
//    - "task.audio.*"
func StartWorkers(cfg config.WorkerConfig) error {
// Purge stale consumers left from previous runs before subscribing.
// Must only run in the worker process — running in the server would kill
// the server's result-proc consumer.
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

// Publish result BEFORE acking — if publish fails the message will be
// redelivered and the worker will retry (idempotent by task_id).
subject := fmt.Sprintf("result.%d", job.TaskID)
data, _ := json.Marshal(result)
if err := mq.PublishResult(subject, data); err != nil {
log.Printf("[worker] task %d: failed to publish result: %v", job.TaskID, err)
// Do NOT ack — let the message be redelivered.
return
}
_ = msg.Ack()
}

// execJob executes a TaskJob and returns a WorkerResult. Never returns nil.
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
Payload:        job.Payload, // preserved for server-side retry on rate_limited
}

fail := func(msg string) *model.WorkerResult {
base.Outcome = model.OutcomeFailed
base.ErrorMsg = msg
return base
}

// Apply request_script
payload := job.Payload
if job.RequestScript != "" {
mapped, err := RunMapRequest(job.RequestScript, payload)
if err != nil {
return fail("request mapping error: " + err.Error())
}
payload = mapped
}

// Record upstream request (for debugging)
upstreamReq := make(map[string]interface{})
for k, v := range payload {
upstreamReq[k] = v
}
base.UpstreamRequest = upstreamReq

// Call upstream HTTP
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

// Apply response_script
if job.ResponseScript != "" {
mapped, err := RunMapResponse(job.ResponseScript, respData)
if err != nil {
return fail("response mapping error: " + err.Error())
}
respData = mapped
}

// Check for async upstream task ID
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

// Error detection (error_script or built-in heuristic)
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

// response_script status=3 means business failure
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
req.Header.Set(k, sv)
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

// DetectUpstreamError detects common vendor error response formats.
// Returns the error message and true if an error is detected.
//
// Covered formats:
//   - OpenAI / generic: {"error": {"message": "...", "code": "..."}}
//   - String error:     {"error": "some message"}
//   - Custom code+msg:  {"code": "InvalidParameter", "message": "..."}
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
