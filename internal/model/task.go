package model

import "time"

// Task 异步任务记录（图片/视频/音频生成任务）。
type Task struct {
	ID             int64     `xorm:"pk autoincr 'id'" json:"id"`
	UserID         int64     `xorm:"notnull index 'user_id'" json:"user_id"`
	ChannelID      int64     `xorm:"notnull 'channel_id'" json:"channel_id"`
	APIKeyID       int64     `xorm:"notnull 'api_key_id'" json:"api_key_id"`
	Type           string    `xorm:"notnull 'type'" json:"type"` // image / video / audio
	Status         string    `xorm:"notnull default('pending') 'status'" json:"status"`
	Progress       int       `xorm:"notnull default(0) 'progress'" json:"progress"`
	Request        JSON      `xorm:"jsonb 'request'" json:"request"`
	Result         JSON      `xorm:"jsonb 'result'" json:"result"`                                     // 经 response_script / query_script 映射后的标准格式
	UpstreamTaskID string    `xorm:"default('') 'upstream_task_id'" json:"upstream_task_id,omitempty"` // 异步渠道：第三方返回的任务 ID，用于轮询
	ErrorMsg       string    `xorm:"text 'error_msg'" json:"error_msg,omitempty"`
	CreditsCharged int64     `xorm:"notnull default(0) 'credits_charged'" json:"credits_charged"`
	CreatedAt      time.Time `xorm:"created 'created_at'" json:"created_at"`
	UpdatedAt      time.Time `xorm:"updated 'updated_at'" json:"updated_at"`
}

func (*Task) TableName() string { return "tasks" }

// TaskResult 是 GET /v1/tasks/:id 返回给用户的统一响应格式。
//
// Code 取值：
//   - 150：任务进行中（排队 / 生成中）
//   - 200：任务成功
//   - 500：任务失败（通用错误）
//   - 其他 >200 值：由 response_script 自定义的精细错误码
//
// Status 取值：
//   - 0：排队中（pending）
//   - 1：生成中（processing）
//   - 2：成功（done）
//   - 3：失败（failed）
type TaskResult struct {
	Code   int    `json:"code"`
	URL    string `json:"url,omitempty"` // 生成结果 URL（成功时）
	Status int    `json:"status"`
	Msg    string `json:"msg,omitempty"` // 状态描述或错误信息
}
