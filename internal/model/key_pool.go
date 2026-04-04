package model

import "time"

// KeyPool 是一组共享相同接口配置的三方 API Key 集合，关联到一个渠道。
// 当渠道配置了 KeyPoolID（非 0），则请求时使用号池进行 Sticky 轮转，
// 而不使用渠道 Headers 中的静态 Authorization Key。
type KeyPool struct {
	ID        int64     `xorm:"pk autoincr 'id'" json:"id"`
	ChannelID int64     `xorm:"notnull index 'channel_id'" json:"channel_id"`
	Name      string    `xorm:"notnull 'name'" json:"name"`
	IsActive  bool      `xorm:"notnull default(true) 'is_active'" json:"is_active"`
	CreatedAt time.Time `xorm:"created 'created_at'" json:"created_at"`
	UpdatedAt time.Time `xorm:"updated 'updated_at'" json:"updated_at"`
}

func (*KeyPool) TableName() string { return "key_pools" }

// PoolKey 号池中的单个三方 API Key。
// Value 存储原始 Key 字符串（如 sk-xxxxx），
// 发请求时注入为 Authorization: Bearer {Value}，覆盖渠道静态 Headers。
type PoolKey struct {
	ID        int64     `xorm:"pk autoincr 'id'" json:"id"`
	PoolID    int64     `xorm:"notnull index 'pool_id'" json:"pool_id"`
	Value     string    `xorm:"notnull text 'value'" json:"value"`
	Priority  int       `xorm:"notnull default(0) 'priority'" json:"priority"` // 越小越优先
	IsActive  bool      `xorm:"notnull default(true) 'is_active'" json:"is_active"`
	CreatedAt time.Time `xorm:"created 'created_at'" json:"created_at"`
}

func (*PoolKey) TableName() string { return "pool_keys" }
