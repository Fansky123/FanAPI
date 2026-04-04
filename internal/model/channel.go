package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// JSON 是 jsonb 列的辅助类型，实现 xorm 的 driver.Valuer 和 Scanner 接口。
type JSON map[string]interface{}

func (j JSON) Value() (driver.Value, error) {
	if j == nil {
		return "{}", nil
	}
	b, err := json.Marshal(j)
	return string(b), err
}

func (j *JSON) Scan(src interface{}) error {
	var data []byte
	switch v := src.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	case nil:
		*j = JSON{}
		return nil
	default:
		return fmt.Errorf("unsupported type: %T", src)
	}
	return json.Unmarshal(data, j)
}

// Channel 渠道表：每条记录代表一个可调用的第三方 API 渠道。
// 同一个模型（Model 字段相同）可以有多个渠道，各自有不同的计费方式和脚本。
type Channel struct {
	ID             int64  `xorm:"pk autoincr 'id'" json:"id"`
	Name           string `xorm:"notnull 'name'" json:"name"`               // 渠道显示名称，如"ChatFire - Claude 3.5 Sonnet"
	Model          string `xorm:"notnull default('') 'model'" json:"model"` // 标准模型名称，如"claude-3-5-sonnet-20241022"，用于分组查询
	Type           string `xorm:"notnull 'type'" json:"type"`               // 接口类型：llm / image / video / audio
	BaseURL        string `xorm:"notnull 'base_url'" json:"base_url"`
	Method         string `xorm:"notnull default('POST') 'method'" json:"method"`
	Headers        JSON   `xorm:"jsonb 'headers'" json:"headers"` // 固定请求头，如 Authorization
	TimeoutMs      int64  `xorm:"notnull default(30000) 'timeout_ms'" json:"timeout_ms"`
	RequestScript  string `xorm:"text 'request_script'" json:"request_script"`   // JS 脚本：mapRequest(input) → 将平台请求映射为上游格式
	ResponseScript string `xorm:"text 'response_script'" json:"response_script"` // JS 脚本：mapResponse(input) → 映射上游响应（同步）或提取 upstream_task_id（异步）
	// 异步轮询配置（video/audio 等异步接口使用）：
	// 提交请求后从 response_script 映射结果中取 upstream_task_id，
	// 然后定期请求 QueryURL（支持 {id} 占位符）获取最终状态。
	QueryURL      string    `xorm:"text default('') 'query_url'" json:"query_url"`             // 轮询 URL，如 https://api.example.com/v1/tasks/{id}
	QueryMethod   string    `xorm:"notnull default('GET') 'query_method'" json:"query_method"` // 轮询 HTTP 方法，默认 GET
	QueryScript   string    `xorm:"text 'query_script'" json:"query_script"`                   // JS 脚本：mapResponse(input) → 将轮询响应映射为标准格式
	BillingType   string    `xorm:"notnull 'billing_type'" json:"billing_type"`                // 计费类型：token / image / video / audio / count / custom
	BillingConfig JSON      `xorm:"jsonb 'billing_config'" json:"billing_config"`
	BillingScript string    `xorm:"text 'billing_script'" json:"billing_script"`          // billing_type=custom 时的计费脚本
	KeyPoolID     int64     `xorm:"default(0) 'key_pool_id'" json:"key_pool_id"`          // 号池 ID（0=不启用），启用后用号池 Key 覆盖 Headers 中的静态 Authorization
	Protocol      string    `xorm:"notnull default('openai') 'protocol'" json:"protocol"` // API 协议格式：openai（默认）/ claude / gemini
	IsActive      bool      `xorm:"notnull default(true) 'is_active'" json:"is_active"`
	CreatedAt     time.Time `xorm:"created 'created_at'" json:"created_at"`
	UpdatedAt     time.Time `xorm:"updated 'updated_at'" json:"updated_at"`
}

func (*Channel) TableName() string { return "channels" }
