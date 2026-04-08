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
		new(model.SystemSetting),
		new(model.PaymentOrder),
	); err != nil {
		return err
	}

	if err := seedAdmin(); err != nil {
		return err
	}
	if err := seedChannels(); err != nil {
		return err
	}
	return ensureIndexes()
}

// ensureIndexes creates performance indexes if they don't already exist.
// Uses CONCURRENTLY so it won't lock tables on a live database.
func ensureIndexes() error {
	indexes := []struct {
		name string
		ddl  string
	}{
		{
			"idx_tasks_processing_upstream",
			`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_tasks_processing_upstream
			ON tasks (status, upstream_task_id)
			WHERE status = 'processing' AND upstream_task_id != ''`,
		},
		{
			"idx_tasks_user_id_created",
			`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_tasks_user_id_created
			ON tasks (user_id, id DESC)`,
		},
		{
			"idx_tasks_type_status_created",
			`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_tasks_type_status_created
			ON tasks (type, status, created_at DESC)`,
		},
		{
			"idx_billing_tx_user_created",
			`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_billing_tx_user_created
			ON billing_transactions (user_id, created_at DESC)`,
		},
		{
			"idx_billing_tx_corr_id",
			`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_billing_tx_corr_id
			ON billing_transactions (corr_id)
			WHERE corr_id != ''`,
		},
	}
	for _, idx := range indexes {
		if _, err := Engine.Exec(idx.ddl); err != nil {
			return fmt.Errorf("create index %s: %w", idx.name, err)
		}
		log.Printf("[db] index ensured: %s", idx.name)
	}
	return nil
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
		queryURL       string
		queryTimeoutMs int64
		queryScript    string
		billingType    string
		billingConfig  string
		headers        model.JSON // nil 则使用默认 ChatFire Key
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
		{
			name:      "Suno V5 音乐生成",
			modelName: "suno-music",
			chType:    "music",
			baseURL:   "YOUR_SUNO_BASE_URL/_open/suno/music/generate",
			timeoutMs: 120000,
			requestScript: `function mapRequest(input) {
    var b = {
        mvVersion:        input.mv_version || 'chirp-v5',
        inputType:        input.input_type  || '10',
        makeInstrumental: input.make_instrumental !== undefined ? input.make_instrumental : false,
        callbackUrl:      input.callback_url || ''
    };
    if (b.inputType === '10') {
        b.gptDescriptionPrompt = input.gpt_description_prompt || '';
    } else {
        b.prompt = input.prompt || '';
        b.tags   = input.tags   || '';
        b.title  = input.title  || '';
        if (input.continue_clip_id) { b.continueClipId = input.continue_clip_id; }
        if (input.continue_at)      { b.continueAt     = input.continue_at; }
        if (input.cover_clip_id)    { b.coverClipId    = input.cover_clip_id; }
        if (input.task) { b.task = input.task; b.metadataParams = input.metadata_params || {}; }
    }
    return b;
}`,
			responseScript: `function mapResponse(output) {
    if (!output || output.code !== 200) {
        return { status: 3, msg: (output && output.msg) ? output.msg : '提交任务失败' };
    }
    var taskBatchId = output.data && output.data.taskBatchId;
    if (!taskBatchId) { return { status: 3, msg: '上游未返回 taskBatchId' }; }
    return { status: 1, upstream_task_id: String(taskBatchId), msg: '生成中' };
}`,
			billingType:    "count",
			billingConfig:  `{"base_price": 5000000}`,
			queryURL:       "YOUR_SUNO_BASE_URL/_open/suno/music/getState?taskBatchId={id}",
			queryTimeoutMs: 30000,
			queryScript: `function mapResponse(output) {
    if (!output || output.code !== 200) {
        return { status: 3, msg: (output && output.msg) ? output.msg : '查询失败' };
    }
    var data = output.data || {};
    var taskStatus = data.taskStatus || '';
    var items = data.items || [];
    if (taskStatus !== 'finished') {
        var tot = 0;
        for (var i = 0; i < items.length; i++) { tot += (items[i].progress || 0); }
        return { status: 1, msg: '生成中', progress: items.length > 0 ? Math.round(tot / items.length) : 0 };
    }
    var ok = [];
    for (var j = 0; j < items.length; j++) {
        var it = items[j];
        if (it.status === 30) {
            ok.push({ id: it.id||'', clip_id: it.clipId||'', title: it.title||'', tags: it.tags||'',
                      prompt: it.prompt||'', duration: it.duration||0,
                      audio_url: it.cld2AudioUrl||'', image_url: it.cld2ImageUrl||'', progress_msg: it.progressMsg||'' });
        }
    }
    if (ok.length === 0) {
        return { status: 3, code: 500, msg: (items[0] && items[0].progressMsg) ? items[0].progressMsg : '创作失败' };
    }
    return { status: 2, code: 200, msg: '创作完成', items: ok };
}`,
			headers: model.JSON{"Authorization": "Bearer YOUR_SUNO_KEY", "Content-Type": "application/json"},
		},
	}

	for _, s := range seeds {
		var bc model.JSON
		_ = json.Unmarshal([]byte(s.billingConfig), &bc)
		headers := s.headers
		if headers == nil {
			headers = model.JSON{"Authorization": "Bearer YOUR_CHATFIRE_KEY", "Content-Type": "application/json"}
		}
		ch := &model.Channel{
			Name:           s.name,
			Model:          s.modelName,
			Type:           s.chType,
			BaseURL:        s.baseURL,
			Method:         "POST",
			Headers:        headers,
			TimeoutMs:      s.timeoutMs,
			RequestScript:  s.requestScript,
			ResponseScript: s.responseScript,
			QueryURL:       s.queryURL,
			QueryMethod:    "GET",
			QueryTimeoutMs: s.queryTimeoutMs,
			QueryScript:    s.queryScript,
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
