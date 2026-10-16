package handler

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fanapi/internal/cache"
	"fanapi/internal/config"
	"fanapi/internal/db"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/redis/go-redis/v9"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"xorm.io/xorm"
	"xorm.io/xorm/core"

	"github.com/gin-gonic/gin"
)

type studioRedisHook struct {
	value   string
	expires time.Time
}

func (h *studioRedisHook) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) { return next(ctx, network, addr) }
}
func (h *studioRedisHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return next
}
func (h *studioRedisHook) ProcessHook(_ redis.ProcessHook) redis.ProcessHook {
	return func(_ context.Context, cmd redis.Cmder) error {
		if cmd.Name() != "getdel" {
			return fmt.Errorf("unexpected Redis command %s", cmd.Name())
		}
		if h.value == "" || time.Now().After(h.expires) {
			cmd.SetErr(redis.Nil)
			return redis.Nil
		}
		cmd.(*redis.StringCmd).SetVal(h.value)
		h.value = ""
		return nil
	}
}

func TestImageStudioRedeemRechecksOwnershipAndCannotReplay(t *testing.T) {
	secret := strings.Repeat("s", 32)
	t.Setenv("AIPIC_HANDOFF_SECRET", secret)
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	engine, err := xorm.NewEngineWithDB("postgres", "host=localhost dbname=unit_test sslmode=disable", core.FromDB(sqlDB))
	if err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal(imageStudioTicket{UserID: 7, KeyID: 9, Model: "gpt-image-test"})
	hook := &studioRedisHook{value: string(data), expires: time.Now().Add(time.Minute)}
	client := redis.NewClient(&redis.Options{Addr: "unused:6379"})
	client.AddHook(hook)
	oldDB, oldCache := db.Engine, cache.Client
	db.Engine, cache.Client = engine, client
	t.Cleanup(func() {
		db.Engine, cache.Client = oldDB, oldCache
		client.Close()
		engine.Close()
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
	})
	key := sha256.Sum256([]byte(secret + ":fanapi:apikey"))
	block, _ := aes.NewCipher(key[:])
	gcm, _ := cipher.NewGCM(block)
	nonce := make([]byte, gcm.NonceSize()) // Test fixture only.
	enc := base64.StdEncoding.EncodeToString(append(nonce, gcm.Seal(nil, nonce, []byte("test-only-key"), nil)...))
	mock.ExpectQuery(`SELECT .* FROM "api_keys" WHERE .*`).WithArgs(int64(9), int64(7)).WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "raw_key_enc", "is_active"}).AddRow(9, 7, enc, true))
	mock.ExpectQuery(`SELECT .* FROM "users" WHERE .*`).WithArgs(int64(7)).WillReturnRows(sqlmock.NewRows([]string{"id", "is_active"}).AddRow(7, true))
	mock.ExpectQuery(`SELECT .* FROM "api_key_model_groups" WHERE .*`).WithArgs(int64(9)).WillReturnRows(sqlmock.NewRows([]string{"id", "api_key_id"}).AddRow(1, 9))
	h := NewAuthHandler(&config.ServerConfig{APIKeySecret: secret}, nil)
	r := gin.New()
	r.POST("/redeem", h.RedeemImageStudioHandoff)
	redeem := func() *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/redeem", strings.NewReader(`{"code":"`+strings.Repeat("a", 64)+`"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-AIPic-Handoff-Secret", secret)
		r.ServeHTTP(w, req)
		return w
	}
	first := redeem()
	if first.Code != 200 || !strings.Contains(first.Body.String(), "test-only-key") {
		t.Fatalf("first redemption failed: %d %s", first.Code, first.Body.String())
	}
	if w := redeem(); w.Code != 410 {
		t.Fatalf("replay accepted: %d", w.Code)
	}
	hook.value, hook.expires = string(data), time.Now().Add(-time.Second)
	if w := redeem(); w.Code != 410 {
		t.Fatalf("expired ticket accepted: %d", w.Code)
	}
	// Revoked/other-user key: no credential is returned, even with a valid ticket.
	hook.value, hook.expires = string(data), time.Now().Add(time.Minute)
	mock.ExpectQuery(`SELECT .* FROM "api_keys" WHERE .*`).WithArgs(int64(9), int64(7)).WillReturnRows(sqlmock.NewRows([]string{"id"}))
	if w := redeem(); w.Code != 403 || strings.Contains(w.Body.String(), "test-only-key") {
		t.Fatal("unavailable key accepted")
	}
}

func TestImageStudioRedirectValidation(t *testing.T) {
	for _, raw := range []string{"https://images.example.com/", "http://localhost:8788/", "http://127.0.0.1:8788/"} {
		if _, err := imageStudioURL(raw); err != nil {
			t.Errorf("rejected %s: %v", raw, err)
		}
	}
	for _, raw := range []string{"", "//example.com", "http://example.com", "https://user:pass@example.com", "https://example.com/?redirect=evil", "https://example.com/#key", "javascript:alert(1)"} {
		if _, err := imageStudioURL(raw); err == nil {
			t.Errorf("accepted %s", raw)
		}
	}
}

func TestImageStudioExchangeRequiresServerSecret(t *testing.T) {
	secret := strings.Repeat("s", 32)
	t.Setenv("AIPIC_HANDOFF_SECRET", secret)
	h := &AuthHandler{}
	r := gin.New()
	r.POST("/redeem", h.RedeemImageStudioHandoff)
	for _, value := range []string{"", "incorrect", secret + "extra"} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/redeem", strings.NewReader(`{"code":"`+strings.Repeat("a", 64)+`"}`))
		req.Header.Set("X-AIPic-Handoff-Secret", value)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", w.Code)
		}
		if w.Header().Get("Cache-Control") != "no-store" {
			t.Fatal("credential exchange must not be cached")
		}
	}
	if imageStudioSecretMatches("", "") || imageStudioSecretMatches("short", "short") {
		t.Fatal("weak or missing secret accepted")
	}
}

func TestImageStudioRequiresUserLogin(t *testing.T) {
	for _, auth := range []string{"", "apikey", "jwt_with_key"} {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("auth_type", auth)
		c.Set("role", "user")
		(&AuthHandler{}).CreateImageStudioHandoff(c)
		if w.Code != http.StatusForbidden {
			t.Fatalf("auth %q accepted", auth)
		}
	}
}
func TestImageStudioAllKeysScopedAndKeepsUnboundKeys(t *testing.T) {
	secret := strings.Repeat("s", 32)
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	engine, err := xorm.NewEngineWithDB("postgres", "host=localhost dbname=unit_test sslmode=disable", core.FromDB(sqlDB))
	if err != nil {
		t.Fatal(err)
	}
	old := db.Engine
	db.Engine = engine
	t.Cleanup(func() {
		db.Engine = old
		engine.Close()
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
	})
	key := sha256.Sum256([]byte(secret + ":fanapi:apikey"))
	block, _ := aes.NewCipher(key[:])
	gcm, _ := cipher.NewGCM(block)
	nonce := make([]byte, gcm.NonceSize())
	enc := func(raw string) string {
		return base64.StdEncoding.EncodeToString(append(nonce, gcm.Seal(nil, nonce, []byte(raw), nil)...))
	}
	mock.ExpectQuery(`SELECT .* FROM "users" WHERE .*`).WithArgs(int64(7)).WillReturnRows(sqlmock.NewRows([]string{"id", "is_active"}).AddRow(7, true))
	mock.ExpectQuery(`SELECT .* FROM "api_keys" WHERE .*user_id.*is_active.*"id" IN.*`).WithArgs(int64(7), int64(9), int64(10), int64(11)).WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name", "raw_key_enc", "is_active"}).AddRow(9, 7, "text", enc("text-key"), true).AddRow(10, 7, "unbound", enc("unbound-key"), true).AddRow(11, 7, "legacy", "", true))
	mock.ExpectQuery(`SELECT .* FROM "model_groups" WHERE .*`).WillReturnRows(sqlmock.NewRows([]string{"id"}))
	for _, id := range []int64{9, 10} {
		mock.ExpectQuery(`SELECT .* FROM "api_key_model_groups" WHERE .*`).WithArgs(id).WillReturnRows(sqlmock.NewRows([]string{"id", "api_key_id"}))
	}
	h := NewAuthHandler(&config.ServerConfig{APIKeySecret: secret}, nil)
	profiles, err := h.imageStudioProfiles(7, []int64{9, 10, 11})
	if err != nil {
		t.Fatal(err)
	}
	if len(profiles) != 2 || profiles[0].APIKey != "text-key" || profiles[1].APIKey != "unbound-key" || profiles[0].ID == profiles[1].ID || len(profiles[1].Models) != 0 {
		t.Fatalf("incorrect profiles: count=%d", len(profiles))
	}
}
