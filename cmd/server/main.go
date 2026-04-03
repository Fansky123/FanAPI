package main

import (
	"fmt"
	"log"

	"fanapi/internal/billing"
	"fanapi/internal/cache"
	"fanapi/internal/config"
	"fanapi/internal/db"
	"fanapi/internal/handler"
	"fanapi/internal/middleware"
	"fanapi/internal/mq"
	"fanapi/pkg/mailer"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	if err := db.Init(&cfg.DB); err != nil {
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

	_ = billing.SyncBalanceToRedis // available for use

	m := mailer.New(&cfg.SMTP)
	authH := handler.NewAuthHandler(&cfg.Server, m)

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// 健康检查（无需认证）
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	// API 文档页面（无需认证）
	r.GET("/docs", handler.APIDocs)

	// Public auth routes
	auth := r.Group("/auth")
	{
		auth.POST("/send-code", authH.SendCode)
		auth.POST("/register", authH.Register)
		auth.POST("/login", authH.Login)
	}

	// Authenticated user routes (JWT or API Key)
	authed := r.Group("/")
	authed.Use(middleware.Auth(&cfg.Server))
	{
		user := authed.Group("/user")
		{
			user.GET("/balance", authH.GetBalance)
			user.GET("/transactions", authH.GetTransactions)
			user.GET("/channels", authH.ListModels) // 对用户暴露渠道列表（含价格展示）
			user.GET("/apikeys", authH.ListAPIKeys)
			user.POST("/apikeys", authH.CreateAPIKey)
			user.DELETE("/apikeys/:id", authH.DeleteAPIKey)
		}

		// Admin routes (JWT or API Key + admin role)
		admin := authed.Group("/admin")
		admin.Use(middleware.Admin())
		{
			admin.POST("/channels", handler.CreateChannel)
			admin.GET("/channels", handler.ListChannels)
			admin.PUT("/channels/:id", handler.UpdateChannel)
			admin.DELETE("/channels/:id", handler.DeleteChannel)
			admin.GET("/users", handler.ListUsers)
			admin.POST("/users/:id/recharge", handler.Recharge)
			admin.GET("/transactions", handler.ListAllTransactions)
			admin.GET("/tasks", handler.ListTasks)
			admin.GET("/tasks/:id", handler.GetAdminTask)
		}

		// Public API (API Key required)
		v1 := authed.Group("/v1")
		v1.Use(middleware.APIKeyOnly())
		{
			v1.POST("/llm", handler.LLMProxy)
			v1.POST("/image", handler.CreateImageTask)
			v1.POST("/video", handler.CreateVideoTask)
			v1.POST("/audio", handler.CreateAudioTask)
			v1.GET("/tasks", handler.ListUserTasks)
			v1.GET("/tasks/:id", handler.GetTask)
		}
	}

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server: %v", err)
	}
}
