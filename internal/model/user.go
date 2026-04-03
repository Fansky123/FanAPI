package model

import "time"

type User struct {
	ID           int64     `xorm:"pk autoincr 'id'" json:"id"`
	Email        string    `xorm:"unique notnull 'email'" json:"email"`
	PasswordHash string    `xorm:"notnull 'password_hash'" json:"-"`
	Role         string    `xorm:"notnull default('user') 'role'" json:"role"`
	IsActive     bool      `xorm:"notnull default(true) 'is_active'" json:"is_active"`
	Balance      int64     `xorm:"notnull default(0) 'balance'" json:"balance"`
	CreatedAt    time.Time `xorm:"created 'created_at'" json:"created_at"`
	UpdatedAt    time.Time `xorm:"updated 'updated_at'" json:"updated_at"`
}

func (*User) TableName() string { return "users" }
