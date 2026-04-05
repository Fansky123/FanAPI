package db

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

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

	// 连接池调优（生产环境建议显式配置）
	if cfg.MaxOpenConns > 0 {
		Engine.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		Engine.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxIdleSec > 0 {
		Engine.SetConnMaxIdleTime(time.Duration(cfg.ConnMaxIdleSec) * time.Second)
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
		new(model.Card),
		new(model.LLMLog),
	); err != nil {
		return err
	}

	if err := seedAdmin(); err != nil {
		return err
	}
	return seedChannels()
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
		username string
		email    string
		password string
		role     string
	}{
		{"admin", defaultAdminEmail, defaultAdminPassword, "admin"},
		{"test", defaultTestEmail, defaultTestPassword, "user"},
	}
	for _, a := range accounts {
		exists, err := Engine.Where("email = ?", a.email).Exist(&model.User{})
		if err != nil {
			return fmt.Errorf("seed check %s: %w", a.email, err)
		}
		if exists {
			// Ensure correct role and backfill username (for accounts seeded before username field was added).
			patch := &model.User{}
			cols := []string{}
			if a.role == "admin" {
				patch.Role = "admin"
				patch.IsActive = true
				cols = append(cols, "role", "is_active")
			}
			patch.Username = a.username
			cols = append(cols, "username")
			if len(cols) > 0 {
				Engine.Where("email = ? AND (username IS NULL OR username = '')", a.email).
					Cols(cols...).Update(patch)
			}
			continue
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(a.password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("seed hash %s: %w", a.email, err)
		}
		if _, err := Engine.Insert(&model.User{
			Username:     a.username,
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

// seedChannels inserts default test channels on first startup (only when the
// channels table is empty). The API keys are left as placeholder strings so
// the operator can update them via the admin UI or a direct SQL UPDATE.
func seedChannels() error {
	count, err := Engine.Count(&model.Channel{})
	if err != nil {
		return fmt.Errorf("seed channels count: %w", err)
	}
	if count > 0 {
		return nil // already seeded
	}

	type channelSeed struct {
		name           string
		modelName      string
		chType         string
		baseURL        string
		timeoutMs      int64
		requestScript  string
		responseScript string
		billingType    string
		billingConfig  string
	}

	seeds := []channelSeed{
		{
			name:          "ChatFire - GPT-4o",
			modelName:     "gpt-4o",
			chType:        "llm",
			baseURL:       "https://api.chatfire.cn/v1/chat/completions",
			timeoutMs:     60000,
			billingType:   "token",
			billingConfig: `{"input_price_per_1m_tokens":18000000,"output_price_per_1m_tokens":72000000,"input_from_response":true,"metric_paths":{"input_tokens":"response.usage.prompt_tokens","output_tokens":"response.usage.completion_tokens"}}`,
		},
		{
			name:          "ChatFire - GPT-4o-mini",
			modelName:     "gpt-4o-mini",
			chType:        "llm",
			baseURL:       "https://api.chatfire.cn/v1/chat/completions",
			timeoutMs:     60000,
			billingType:   "token",
			billingConfig: `{"input_price_per_1m_tokens":1100000,"output_price_per_1m_tokens":4400000,"input_from_response":true,"metric_paths":{"input_tokens":"response.usage.prompt_tokens","output_tokens":"response.usage.completion_tokens"}}`,
		},
		{
			name:      "ChatFire - Nano Banana Pro",
			modelName: "nano-banana-pro",
			chType:    "image",
			baseURL:   "https://api.chatfire.cn/v1/images/generations",
			timeoutMs: 120000,
			requestScript: `function mapRequest(input) {
    var out = {};
    out.prompt = input.prompt;
    // size 未填时默认 1k
    var size = input.size && input.size !== '' ? input.size : '1k';
    out.model = (input.model || 'nano-banana-pro') + '_' + size;
    // aspect_ratio "16:9" → chatfire size "16x9"
    var ar = input.aspect_ratio;
    if (ar && ar !== '') { out.size = ar.replace(':', 'x'); }
    // refer_images[0] → image（chatfire 接受单个 URL 字符串）
    var refs = input.refer_images;
    if (refs && refs.length > 0) { out.image = refs[0]; }
    return out;
}`,
			responseScript: `function mapResponse(input) {
    var out = { code: 200, status: 2, msg: '' };
    if (input.data && input.data.length > 0) { out.url = input.data[0].url; }
    return out;
}`,
			billingType: "image",
			// size_prices：按档位直接定价（credits），不同档位成本差异大
			// 1k ≈ 0.005 CNY / 2k ≈ 0.015 CNY / 3k ≈ 0.030 CNY / 4k ≈ 0.050 CNY
			// （1 CNY = 1,000,000 credits，以下数值可在管理后台按实际成本调整）
			billingConfig: `{
				"size_prices": {
					"1k": 5000,
					"2k": 15000,
					"3k": 30000,
					"4k": 50000
				},
				"default_size_price": 50000,
				"metric_paths": {
					"size":  "request.size",
					"count": "request.n"
				}
			}`,
		},
	}

	for _, s := range seeds {
		var bc model.JSON
		_ = json.Unmarshal([]byte(s.billingConfig), &bc)
		ch := &model.Channel{
			Name:           s.name,
			Model:          s.modelName,
			Type:           s.chType,
			BaseURL:        s.baseURL,
			Method:         "POST",
			Headers:        model.JSON{"Authorization": "Bearer YOUR_CHATFIRE_KEY", "Content-Type": "application/json"},
			TimeoutMs:      s.timeoutMs,
			RequestScript:  s.requestScript,
			ResponseScript: s.responseScript,
			QueryMethod:    "GET",
			BillingType:    s.billingType,
			BillingConfig:  bc,
			Protocol:       "openai",
			IsActive:       true,
		}
		if _, err := Engine.Insert(ch); err != nil {
			return fmt.Errorf("seed channel %s: %w", s.name, err)
		}
		log.Printf("[db] seeded channel: %s (model=%s)", s.name, s.modelName)
	}
	return nil
}
