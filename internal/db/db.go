package db

import (
	"fmt"
	"log"

	"fanapi/internal/config"
	"fanapi/internal/model"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"xorm.io/xorm"
)

var Engine *xorm.Engine

// Init connects to the database. Pass migrate=true only in the server process
// to run schema migrations (Sync2). Worker processes pass migrate=false.
func Init(cfg *config.DBConfig, migrate bool) error {
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

	if !migrate {
		return nil
	}

	if err := Engine.Sync2(
		new(model.User),
		new(model.EmailVerification),
		new(model.APIKey),
		new(model.Channel),
		new(model.KeyPool),
		new(model.PoolKey),
		new(model.Task),
		new(model.BillingTransaction),
	); err != nil {
		return err
	}

	return seedAdmin()
}

const (
	defaultAdminEmail    = "admin@fanapi.dev"
	defaultAdminPassword = "Admin@2026!"
	defaultTestEmail     = "test@fanapi.dev"
	defaultTestPassword  = "Test@2026!"
)

// seedAdmin creates default admin and test accounts on first startup.
// Safe to call on every startup — uses INSERT ... WHERE NOT EXISTS.
func seedAdmin() error {
	accounts := []struct {
		email    string
		password string
		role     string
	}{
		{defaultAdminEmail, defaultAdminPassword, "admin"},
		{defaultTestEmail, defaultTestPassword, "user"},
	}
	for _, a := range accounts {
		exists, err := Engine.Where("email = ?", a.email).Exist(&model.User{})
		if err != nil {
			return fmt.Errorf("seed check %s: %w", a.email, err)
		}
		if exists {
			continue
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(a.password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("seed hash %s: %w", a.email, err)
		}
		if _, err := Engine.Insert(&model.User{
			Email:        a.email,
			PasswordHash: string(hash),
			Role:         a.role,
			IsActive:     true,
		}); err != nil {
			return fmt.Errorf("seed insert %s: %w", a.email, err)
		}
		log.Printf("[db] seeded account: %s (role=%s)", a.email, a.role)
	}
	return nil
}
