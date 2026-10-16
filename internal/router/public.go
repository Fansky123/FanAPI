package router

import (
	"net/http"

	"fanapi/internal/handler"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func registerPublicRoutes(r *gin.Engine, deps Dependencies) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/openapi.json", handler.SwaggerJSON)
	r.GET("/openapi-user.json", handler.UserSwaggerJSON)
	r.GET("/docs", handler.APIDocs)

	r.GET("/public/channels", deps.Auth.ListModels)
	r.GET("/public/settings", handler.GetPublicSettings)
	r.POST("/public/image-studio/redeem", deps.Auth.RedeemImageStudioHandoff)

	r.GET("/pay/epay/callback", handler.EpayCallback)
	r.POST("/pay/epay/callback", handler.EpayCallback)
	r.POST("/pay/apply/notify", handler.PayApplyNotify)
	r.POST("/pay/shouqianba/notify", handler.ShouqianbaNotify)
}
