package main

import (
	"context"
	"fmt"
	"log"

	"fanapi/internal/billing"
	"fanapi/internal/cache"
	"fanapi/internal/config"
	"fanapi/internal/db"
	"fanapi/internal/handler"
	"fanapi/internal/middleware"
	"fanapi/internal/mq"
	"fanapi/internal/taskresult"
	"fanapi/pkg/mailer"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	if err := db.Init(&cfg.DB, true); err != nil {
		log.Fatalf("db: %v", err)
	}
	log.Println("db connected")

	if err := cache.Init(&cfg.Redis); err != nil {
		log.Fatalf("redis: %v", err)
	}
	log.Println("redis connected")

	if err := mq.Init(&cfg.NATS); err != nil {
		log.Fatalf("nats: %v", err)
	}
	log.Println("nats connected")
	if err := mq.EnsureStream(); err != nil {
		log.Fatalf("nats ensure stream: %v", err)
	}

	_ = billing.SyncBalanceToRedis // available for use

	// Start result processor: subscribes to RESULTS stream, writes DB + billing
	if err := taskresult.StartResultProcessor(cfg.Worker); err != nil {
		log.Fatalf("result processor: %v", err)
	}

	// Start async-task poller (polls DB for processing tasks with upstream_task_id)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	taskresult.StartPoller(ctx)

	m := mailer.New(&cfg.SMTP)
	authH := handler.NewAuthHandler(&cfg.Server, m)

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// 健康检查（无需认证）
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	// API 文档页面（无需认证）
	r.GET("/docs", handler.APIDocs)

	// 公开接口（无需认证）
	r.GET("/public/channels", authH.ListModels)

	// Public auth routes
	auth := r.Group("/auth")
	{
		auth.POST("/send-code", authH.SendCode)
		auth.POST("/register", authH.Register)
		auth.POST("/login", authH.Login)
		auth.POST("/forgot-password", authH.ForgotPassword)
		auth.POST("/reset-password", authH.ResetPassword)
	}

	// Authenticated user routes (JWT or API Key)
	authed := r.Group("/")
	authed.Use(middleware.Auth(&cfg.Server))
	{
		user := authed.Group("/user")
		{
			user.GET("/profile", authH.GetProfile)
			user.GET("/balance", authH.GetBalance)
			user.GET("/transactions", authH.GetTransactions)
			user.GET("/channels", authH.ListModels)
			user.GET("/apikeys", authH.ListAPIKeys)
			user.POST("/apikeys", authH.CreateAPIKey)
			user.DELETE("/apikeys/:id", authH.DeleteAPIKey)
			user.PUT("/password", authH.ChangePassword)
			user.POST("/bind-email", authH.BindEmail)
			user.POST("/cards/redeem", handler.RedeemCard)
		}

		// Admin routes (JWT or API Key + admin role)
		admin := authed.Group("/admin")
		admin.Use(middleware.Admin())
		{
			admin.POST("/channels", handler.CreateChannel)
			admin.GET("/channels", handler.ListChannels)
			admin.PUT("/channels/:id", handler.UpdateChannel)
			admin.DELETE("/channels/:id", handler.DeleteChannel)
			// 号池管理
			admin.GET("/key-pools", handler.ListKeyPools)
			admin.POST("/key-pools", handler.CreateKeyPool)
			admin.DELETE("/key-pools/:id", handler.DeleteKeyPool)
			admin.GET("/key-pools/:id/keys", handler.ListPoolKeys)
			admin.POST("/key-pools/:id/keys", handler.AddPoolKey)
			admin.DELETE("/pool-keys/:id", handler.RemovePoolKey)
			admin.GET("/users", handler.ListUsers)
			admin.POST("/users/:id/recharge", handler.Recharge)
			admin.PUT("/users/:id/password", handler.ResetUserPassword)
			admin.GET("/transactions", handler.ListAllTransactions)
			admin.GET("/tasks", handler.ListTasks)
			admin.GET("/tasks/:id", handler.GetAdminTask)
			admin.GET("/stats", handler.GetAdminStats)
			// 卡密管理
			admin.POST("/cards/generate", handler.GenerateCards)
			admin.GET("/cards", handler.ListCards)
			admin.DELETE("/cards/:id", handler.DeleteCard)
			// LLM 日志
			admin.GET("/llm-logs", handler.AdminListLLMLogs)
			admin.GET("/llm-logs/:id", handler.AdminGetLLMLog)
		}

		// 用户任务查询（支持 JWT 或 API Key）
		authed.GET("/v1/tasks", handler.ListUserTasks)
		authed.GET("/v1/tasks/:id", handler.GetTask)
		authed.GET("/v1/llm-logs", handler.UserListLLMLogs)

		// Public API (API Key required)
		v1 := authed.Group("/v1")
		v1.Use(middleware.APIKeyOnly())
		{
			v1.POST("/chat/completions", handler.LLMProxy) // OpenAI 兼容格式
			v1.POST("/messages", handler.ClaudeProxy)      // Claude 原生格式
			v1.POST("/gemini", handler.GeminiProxy)        // Gemini 原生格式
			v1.POST("/image", handler.CreateImageTask)
			v1.POST("/video", handler.CreateVideoTask)
			v1.POST("/audio", handler.CreateAudioTask)
		}
	}

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server: %v", err)
	}
}
