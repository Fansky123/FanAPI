package model

import "time"

// WithdrawRequest 积分提现申请。
type WithdrawRequest struct {
	ID          int64     `xorm:"pk autoincr 'id'" json:"id"`
	UserID      int64     `xorm:"notnull 'user_id'" json:"user_id"`
	Amount      int64     `xorm:"notnull 'amount'" json:"amount"`                         // 微单位积分
	Status      string    `xorm:"notnull default('pending') 'status'" json:"status"`      // pending/approved/rejected
	PaymentType string    `xorm:"notnull default('') 'payment_type'" json:"payment_type"` // wechat/alipay
	PaymentQR   string    `xorm:"notnull default('') 'payment_qr'" json:"payment_qr"`     // 收款码快照
	AdminRemark string    `xorm:"notnull default('') 'admin_remark'" json:"admin_remark,omitempty"`
	CreatedAt   time.Time `xorm:"created 'created_at'" json:"created_at"`
	UpdatedAt   time.Time `xorm:"updated 'updated_at'" json:"updated_at"`

	// 关联字段（查询时 JOIN 填充，不入库）
	Username string `xorm:"-" json:"username,omitempty"`
}

func (*WithdrawRequest) TableName() string { return "withdraw_requests" }
