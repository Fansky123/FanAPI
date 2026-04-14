package model

import "time"

// LLMLog 记录每次 LLM 请求的详情：向上游发送的请求体、上游返回的 HTTP 状态、
// 最终结算的 token 用量及请求状态，供管理员和用户排查计费与问题。
//
// 状态（Status）取值：
//   - pending  : 请求已发出，尚未收到完整响应（进行中）
//   - ok       : 请求成功，结算完毕
//   - error    : 上游返回错误或连接失败（已退款）
//   - refunded : 请求抵达上游但无任何输出（已全额退款）
type LLMLog struct {
	ID               int64     `xorm:"pk autoincr 'id'" json:"id"`
	UserID           int64     `xorm:"notnull index 'user_id'" json:"user_id"`
	ChannelID        int64     `xorm:"'channel_id'" json:"channel_id"`
	APIKeyID         int64     `xorm:"'api_key_id'" json:"api_key_id"`
	CorrID           string    `xorm:"index 'corr_id'" json:"corr_id"`
	Model            string    `xorm:"default('') 'model'" json:"model"`                             // 请求中的 model 字段（原始客户端值）
	IsStream         bool      `xorm:"'is_stream'" json:"is_stream"`                                 // 是否流式请求
	UpstreamURL      string    `xorm:"text default('') 'upstream_url'" json:"upstream_url"`          // 实际发往上游的完整 URL（含 {model} 替换后）
	UpstreamMethod   string    `xorm:"default('POST') 'upstream_method'" json:"upstream_method"`     // 上游请求方式
	UpstreamRequest  JSON      `xorm:"jsonb 'upstream_request'" json:"upstream_request"`             // 发往上游的完整 JSON body
	UpstreamStatus   int       `xorm:"default(0) 'upstream_status'" json:"upstream_status"`          // 上游 HTTP 状态码
	UpstreamResponse JSON      `xorm:"jsonb 'upstream_response'" json:"upstream_response,omitempty"` // 上游响应 JSON（同步模式存完整响应，流式模式存 usage chunk）
	Usage            JSON      `xorm:"jsonb 'usage'" json:"usage,omitempty"`                         // 结算用量 {prompt_tokens, completion_tokens, estimated?}
	Status           string    `xorm:"notnull default('pending') 'status'" json:"status"`
	ErrorMsg         string    `xorm:"text 'error_msg'" json:"error_msg,omitempty"`
	CreatedAt        time.Time `xorm:"created 'created_at'" json:"created_at"`
	UpdatedAt        time.Time `xorm:"updated 'updated_at'" json:"updated_at"`
}

func (*LLMLog) TableName() string { return "llm_logs" }
