package db

import (
	"fmt"

	"fanapi/internal/config"
	"fanapi/internal/model"

	_ "github.com/lib/pq"
	"xorm.io/xorm"
)

var Engine *xorm.Engine

func Init(cfg *config.DBConfig) error {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	var err error
	Engine, err = xorm.NewEngine("postgres", dsn)
	if err != nil {
		return err
	}

	if err = Engine.Ping(); err != nil {
		return err
	}

	return Engine.Sync2(
		new(model.User),
		new(model.EmailVerification),
		new(model.APIKey),
		new(model.Channel),
		new(model.Task),
		new(model.BillingTransaction),
	)
}
