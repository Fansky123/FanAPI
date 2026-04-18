package model

import "time"

// Vendor 号商：向平台提供 API Key 的供应商，独立于用户系统。
type Vendor struct {
	ID              int64     `xorm:"pk autoincr 'id'" json:"id"`
	Username        string    `xorm:"unique notnull 'username'" json:"username"`
	PasswordHash    string    `xorm:"notnull 'password_hash'" json:"-"`
	Email           *string   `xorm:"unique 'email' null" json:"email"`
	IsActive        bool      `xorm:"notnull default(true) 'is_active'" json:"is_active"`
	Balance         int64     `xorm:"notnull default(0) 'balance'" json:"balance"`         // 可提现余额（credits，平台扣除手续费后净额）
	CommissionRatio *float64  `xorm:"'commission_ratio' null" json:"commission_ratio,omitempty"` // 平台手续费比例（nil 使用系统默认值）
	InviteCode      string    `xorm:"unique notnull 'invite_code'" json:"invite_code"`     // 注册码（唯一，自动生成）
	CreatedAt       time.Time `xorm:"created 'created_at'" json:"created_at"`
	UpdatedAt       time.Time `xorm:"updated 'updated_at'" json:"updated_at"`
}

func (*Vendor) TableName() string { return "vendors" }
