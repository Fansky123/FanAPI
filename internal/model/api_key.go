package model

import "time"

type APIKey struct {
	ID         int64      `xorm:"pk autoincr 'id'" json:"id"`
	UserID     int64      `xorm:"notnull index 'user_id'" json:"user_id"`
	KeyHash    string     `xorm:"notnull unique 'key_hash'" json:"-"`
	RawKeyEnc  string     `xorm:"text 'raw_key_enc'" json:"-"`
	Name       string     `xorm:"notnull 'name'" json:"name"`
	KeyType    string     `xorm:"notnull default('low_price') 'key_type'" json:"key_type"` // low_price | stable
	IsActive   bool       `xorm:"notnull default(true) 'is_active'" json:"is_active"`
	LastUsedAt *time.Time `xorm:"'last_used_at'" json:"last_used_at"`
	CreatedAt  time.Time  `xorm:"created 'created_at'" json:"created_at"`
}

func (*APIKey) TableName() string { return "api_keys" }
