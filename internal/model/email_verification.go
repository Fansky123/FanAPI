package model

import "time"

type EmailVerification struct {
	ID        int64     `xorm:"pk autoincr 'id'" json:"id"`
	Email     string    `xorm:"notnull 'email'" json:"email"`
	Code      string    `xorm:"notnull 'code'" json:"code"`
	ExpiresAt time.Time `xorm:"notnull 'expires_at'" json:"expires_at"`
	Used      bool      `xorm:"notnull default(false) 'used'" json:"used"`
	CreatedAt time.Time `xorm:"created 'created_at'" json:"created_at"`
}

func (*EmailVerification) TableName() string { return "email_verifications" }
