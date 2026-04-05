package model

// TaskJob is the fat message published by the API server to NATS for a worker to execute.
// It embeds everything the worker needs, so the worker only requires a NATS connection —
// no direct access to PostgreSQL or Redis.
type TaskJob struct {
	// Task identity & billing (needed by result handler for refund on failure)
	TaskID         int64  `json:"task_id"`
	TaskType       string `json:"task_type"` // image / video / audio
	UserID         int64  `json:"user_id"`
	APIKeyID       int64  `json:"api_key_id"`
	CorrID         string `json:"corr_id"`
	CreditsCharged int64  `json:"credits_charged"`
	ChannelID      int64  `json:"channel_id"`

	// Channel execution config (embedded so worker doesn't need DB/Redis)
	BaseURL        string                 `json:"base_url"`
	Method         string                 `json:"method"`
	Headers        map[string]interface{} `json:"headers"`
	TimeoutMs      int64                  `json:"timeout_ms"`
	RequestScript  string                 `json:"request_script,omitempty"`
	ResponseScript string                 `json:"response_script,omitempty"`
	ErrorScript    string                 `json:"error_script,omitempty"`
	QueryURL       string                 `json:"query_url,omitempty"`
	QueryMethod    string                 `json:"query_method,omitempty"`
	QueryScript    string                 `json:"query_script,omitempty"`

	// Pre-resolved pool key (if channel uses a key pool)
	PoolKeyID    int64  `json:"pool_key_id,omitempty"`
	PoolKeyValue string `json:"pool_key_value,omitempty"`

	// Request payload (platform standard format — request_script not yet applied)
	Payload map[string]interface{} `json:"payload"`

	// Retry counter — incremented by the server on 429 key-rotation retries
	RetryCount int `json:"retry_count,omitempty"`
}

// WorkerResult outcome constants.
const (
	OutcomeDone        = "done"
	OutcomeFailed      = "failed"
	OutcomeAsync       = "async"        // upstream returned an async task ID; poller will finish it
	OutcomeRateLimited = "rate_limited" // HTTP 429; server should rotate pool key and retry
)

// WorkerResult is published by the worker to NATS after executing a task.
// The API server subscribes and handles DB writes + billing.
type WorkerResult struct {
	TaskID         int64  `json:"task_id"`
	TaskType       string `json:"task_type"`
	UserID         int64  `json:"user_id"`
	APIKeyID       int64  `json:"api_key_id"`
	CorrID         string `json:"corr_id"`
	CreditsCharged int64  `json:"credits_charged"`
	ChannelID      int64  `json:"channel_id"`
	PoolKeyID      int64  `json:"pool_key_id,omitempty"`

	Outcome string `json:"outcome"` // one of the Outcome* constants

	// OutcomeDone
	Result map[string]interface{} `json:"result,omitempty"`

	// OutcomeAsync
	UpstreamTaskID string `json:"upstream_task_id,omitempty"`

	// OutcomeFailed / OutcomeRateLimited
	ErrorMsg string `json:"error_msg,omitempty"`

	// Debug info
	UpstreamRequest  map[string]interface{} `json:"upstream_request,omitempty"`
	UpstreamResponse map[string]interface{} `json:"upstream_response,omitempty"`

	// Passed back so server can re-publish on OutcomeRateLimited
	RetryCount int                    `json:"retry_count,omitempty"`
	Payload    map[string]interface{} `json:"payload,omitempty"`
}
