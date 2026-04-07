package model

import "time"

// SystemSetting stores platform-wide configuration as key-value pairs.
type SystemSetting struct {
	ID        int64     `xorm:"pk autoincr 'id'" json:"id"`
	Key       string    `xorm:"unique notnull 'key'" json:"key"`
	Value     string    `xorm:"notnull 'value' text" json:"value"`
	CreatedAt time.Time `xorm:"created 'created_at'" json:"created_at"`
	UpdatedAt time.Time `xorm:"updated 'updated_at'" json:"updated_at"`
}

func (*SystemSetting) TableName() string { return "system_settings" }
