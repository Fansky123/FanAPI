package model

import "time"

type User struct {
	ID           int64     `xorm:"pk autoincr 'id'" json:"id"`
	Username     string    `xorm:"unique 'username'" json:"username"` // 注册用户名（唯一，可空留给老数据）
	Email        string    `xorm:"unique 'email' null" json:"email"`  // 绑定邮箱（可空，用于找回密码）
	PasswordHash string    `xorm:"notnull 'password_hash'" json:"-"`
	Role         string    `xorm:"notnull default('user') 'role'" json:"role"`
	Group        string    `xorm:"notnull default('') 'group'" json:"group"` // 用户分组，用于差异化定价（空=默认定价）
	IsActive     bool      `xorm:"notnull default(true) 'is_active'" json:"is_active"`
	Balance      int64     `xorm:"notnull default(0) 'balance'" json:"balance"`
	CreatedAt    time.Time `xorm:"created 'created_at'" json:"created_at"`
	UpdatedAt    time.Time `xorm:"updated 'updated_at'" json:"updated_at"`
}

func (*User) TableName() string { return "users" }
