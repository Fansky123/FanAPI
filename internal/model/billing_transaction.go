package model

import "time"

type BillingTransaction struct {
	ID        int64     `xorm:"pk autoincr 'id'" json:"id"`
	UserID    int64     `xorm:"notnull index 'user_id'" json:"user_id"`
	ChannelID int64     `xorm:"'channel_id'" json:"channel_id"`
	APIKeyID  int64     `xorm:"'api_key_id'" json:"api_key_id"`
	CorrID    string    `xorm:"'corr_id'" json:"corr_id"`   // links hold+settle pairs
	Type      string    `xorm:"notnull 'type'" json:"type"` // charge,hold,settle,refund,recharge
	Credits   int64     `xorm:"notnull 'credits'" json:"credits"`
	Metrics   JSON      `xorm:"jsonb 'metrics'" json:"metrics"`
	CreatedAt time.Time `xorm:"created 'created_at'" json:"created_at"`
}

func (*BillingTransaction) TableName() string { return "billing_transactions" }
