package model

import "time"

// PaymentOrder tracks pending and completed Epay payment orders.
type PaymentOrder struct {
	ID         int64      `xorm:"pk autoincr 'id'" json:"id"`
	UserID     int64      `xorm:"notnull 'user_id' index" json:"user_id"`
	OutTradeNo string     `xorm:"unique notnull 'out_trade_no'" json:"out_trade_no"`
	Amount     float64    `xorm:"notnull 'amount'" json:"amount"`    // 充值金额（元）
	Credits    int64      `xorm:"notnull 'credits'" json:"credits"` // 等值积分（1元=1000000）
	Status     string     `xorm:"notnull default('pending') 'status'" json:"status"` // pending/paid/failed
	TradeNo    string     `xorm:"notnull default('') 'trade_no'" json:"trade_no"`    // Epay 交易号
	CreatedAt  time.Time  `xorm:"created 'created_at'" json:"created_at"`
	PaidAt     *time.Time `xorm:"null 'paid_at'" json:"paid_at"`
}

func (*PaymentOrder) TableName() string { return "payment_orders" }
