package model

import "time"

// PaymentOrder tracks pending and completed payment orders.
// Supports both Epay and 中台 (pay-apply) payment channels.
type PaymentOrder struct {
	ID         int64   `xorm:"pk autoincr 'id'" json:"id"`
	UserID     int64   `xorm:"notnull 'user_id' index" json:"user_id"`
	OutTradeNo string  `xorm:"unique notnull 'out_trade_no'" json:"out_trade_no"` // 本系统订单号
	Amount     float64 `xorm:"notnull 'amount'" json:"amount"`                    // 充值金额（元）
	Credits    int64   `xorm:"notnull 'credits'" json:"credits"`                  // 等值积分（1元=1000000）
	Status     string  `xorm:"notnull default('pending') 'status'" json:"status"` // pending/paid/failed
	TradeNo    string  `xorm:"notnull default('') 'trade_no'" json:"trade_no"`    // 三方平台交易号
	// 中台支付扩展字段
	PayFlat   int        `xorm:"notnull default(0) 'pay_flat'" json:"pay_flat"`  // 0=Epay 1=微信 2=支付宝
	PayFrom   string     `xorm:"notnull default('') 'pay_from'" json:"pay_from"` // 支付终端来源
	ProName   string     `xorm:"notnull default('') 'pro_name'" json:"pro_name"` // 商品名称
	CreatedAt time.Time  `xorm:"created 'created_at'" json:"created_at"`
	PaidAt    *time.Time `xorm:"null 'paid_at'" json:"paid_at"`
}

func (*PaymentOrder) TableName() string { return "payment_orders" }
