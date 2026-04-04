package model

import "time"

// Card 卡密 — 管理员生成，用户通过兑换码充值余额
type Card struct {
	ID        int64      `xorm:"pk autoincr 'id'" json:"id"`
	Code      string     `xorm:"notnull unique 'code'" json:"code"`                // 兑换码，唯一
	Credits   int64      `xorm:"notnull 'credits'" json:"credits"`                 // 面值（微元）
	Status    string     `xorm:"notnull default('unused') 'status'" json:"status"` // unused / used
	Note      string     `xorm:"default('') 'note'" json:"note"`                   // 备注
	CreatedBy int64      `xorm:"default(0) 'created_by'" json:"created_by"`
	UsedBy    int64      `xorm:"default(0) 'used_by'" json:"used_by"`
	UsedAt    *time.Time `xorm:"'used_at'" json:"used_at"`
	CreatedAt time.Time  `xorm:"created 'created_at'" json:"created_at"`
}

func (*Card) TableName() string { return "cards" }
