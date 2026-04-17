package model

import "time"

// OcpcRecord 存储用户广告点击追踪数据及转化上报状态。
// 每个用户一条记录（unique on user_id）。追踪参数在注册或微信扫码登录时写入；
// 订单转化在支付成功后通过 MarkOcpcOrder 更新。
type OcpcRecord struct {
	ID        int64 `xorm:"pk autoincr 'id'" json:"id"`
	UserID    int64 `xorm:"unique 'user_id' index" json:"user_id"`

	// 关联的广告平台账户（ocpc_platforms.id）；0 表示未绑定到具体账户（旧记录）
	PlatformID int64 `xorm:"'platform_id' index" json:"platform_id"`

	// 广告追踪参数
	BdVid     string `xorm:"'bd_vid'" json:"bd_vid"`           // 百度 OCPC 点击追踪 ID
	QhClickID string `xorm:"'qh_click_id'" json:"qh_click_id"` // 360 搜索推广点击 ID
	SourceID  string `xorm:"'source_id'" json:"source_id"`     // 360 展示广告 ID
	IP        string `xorm:"'ip'" json:"ip"`
	UA        string `xorm:"'ua'" json:"ua"`

	// 注册转化上报状态
	RegIsUploaded bool       `xorm:"'reg_is_uploaded'" json:"reg_is_uploaded"`
	RegUploadedAt *time.Time `xorm:"'reg_uploaded_at' null" json:"reg_uploaded_at,omitempty"`
	RegRetJSON    string     `xorm:"'reg_ret_json'" json:"reg_ret_json,omitempty"`

	// 订单转化上报状态（OrderAmount > 0 表示有待上报的付款记录）
	OrderAmount     float64    `xorm:"'order_amount'" json:"order_amount"`
	OrderIsUploaded bool       `xorm:"'order_is_uploaded'" json:"order_is_uploaded"`
	OrderUploadedAt *time.Time `xorm:"'order_uploaded_at' null" json:"order_uploaded_at,omitempty"`
	OrderRetJSON    string     `xorm:"'order_ret_json'" json:"order_ret_json,omitempty"`

	AddTime   int64     `xorm:"'add_time'" json:"add_time"`
	CreatedAt time.Time `xorm:"created 'created_at'" json:"created_at"`
	UpdatedAt time.Time `xorm:"updated 'updated_at'" json:"updated_at"`
}

func (*OcpcRecord) TableName() string { return "ocpc_records" }
