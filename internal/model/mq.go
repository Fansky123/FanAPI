package model

// TaskJob 是 API 服务器发布到 NATS 中供 Worker 执行的完整任务消息。
// 嵌入了 Worker 所需的全部信息，Worker 只需 NATS 连接，
// 无需直接访问 PostgreSQL 或 Redis。
type TaskJob struct {
	// 任务身丽与计费信息（结果处理器在失败退款时需要）
	TaskID         int64  `json:"task_id"`
	TaskType       string `json:"task_type"` // image / video / audio
	UserID         int64  `json:"user_id"`
	APIKeyID       int64  `json:"api_key_id"`
	CorrID         string `json:"corr_id"`
	CreditsCharged int64  `json:"credits_charged"`
	ChannelID      int64  `json:"channel_id"`

	// 渠道执行配置（嵌入以免 Worker 访问 DB/Redis）
	BaseURL        string                 `json:"base_url"`
	Method         string                 `json:"method"`
	Headers        map[string]interface{} `json:"headers"`
	TimeoutMs      int64                  `json:"timeout_ms"`
	QueryTimeoutMs int64                  `json:"query_timeout_ms,omitempty"`
	RequestScript  string                 `json:"request_script,omitempty"`
	ResponseScript string                 `json:"response_script,omitempty"`
	ErrorScript    string                 `json:"error_script,omitempty"`
	QueryURL       string                 `json:"query_url,omitempty"`
	QueryMethod    string                 `json:"query_method,omitempty"`
	QueryScript    string                 `json:"query_script,omitempty"`

	// 预分配的号池 Key（渠道已设置号池时嵌入）
	PoolKeyID    int64  `json:"pool_key_id,omitempty"`
	PoolKeyValue string `json:"pool_key_value,omitempty"`

	// 请求载荷（平台标准格式，尚未应用 request_script）
	Payload map[string]interface{} `json:"payload"`

	// 重试计数器——服务器在 429 轮转 Key 重试时递增
	RetryCount int `json:"retry_count,omitempty"`

	// 稳定密钥失败重试：按售价升序排列的剩余待试渠道 ID 列表（不含当前渠道）。
	// 结果处理器在 OutcomeFailed 时若此列表非空，则换下一个渠道重新发布任务。
	RetryChannelIDs []int64 `json:"retry_channel_ids,omitempty"`
}

// WorkerResult 结果常量。
const (
	OutcomeDone        = "done"
	OutcomeFailed      = "failed"
	OutcomeAsync       = "async"        // 上游返回了异步任务 ID，由轮询器完成后续处理
	OutcomeRateLimited = "rate_limited" // HTTP 429；服务器应轮转号池 Key 并重试
)

// WorkerResult 是 Worker 执行完成后发布到 NATS 的结果消息。
// API 服务器订阅并处理 DB 写入与计费结算。
type WorkerResult struct {
	TaskID         int64  `json:"task_id"`
	TaskType       string `json:"task_type"`
	UserID         int64  `json:"user_id"`
	APIKeyID       int64  `json:"api_key_id"`
	CorrID         string `json:"corr_id"`
	CreditsCharged int64  `json:"credits_charged"`
	ChannelID      int64  `json:"channel_id"`
	PoolKeyID      int64  `json:"pool_key_id,omitempty"`

	Outcome string `json:"outcome"` // 取値为 Outcome* 常量之一

	// OutcomeDone 时填充
	Result map[string]interface{} `json:"result,omitempty"`

	// OutcomeAsync 时填充
	UpstreamTaskID string `json:"upstream_task_id,omitempty"`

	// OutcomeFailed / OutcomeRateLimited 时填充
	ErrorMsg string `json:"error_msg,omitempty"`

	// 调试信息
	UpstreamRequest  map[string]interface{} `json:"upstream_request,omitempty"`
	UpstreamResponse map[string]interface{} `json:"upstream_response,omitempty"`

	// 传回给服务器以便在 OutcomeRateLimited 时重新发布
	RetryCount      int                    `json:"retry_count,omitempty"`
	Payload         map[string]interface{} `json:"payload,omitempty"`
	RetryChannelIDs []int64                `json:"retry_channel_ids,omitempty"` // 稳定密钥：剩余待试渠道 ID
}
