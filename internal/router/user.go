package router

import (
	"fanapi/internal/handler"
	"fanapi/internal/middleware"

	"github.com/gin-gonic/gin"
)

func registerUserRoutes(authed *gin.RouterGroup, deps Dependencies) {
	authed.POST("/upload/image", handler.UploadImage)
	authed.POST("/upload/video", handler.UploadVideo)

	resellerPlatform := authed.Group("/reseller/platform")
	resellerPlatform.Use(middleware.APIKeyOnly())
	{
		resellerPlatform.GET("/channels", handler.ResellerPlatformChannels)
	}

	user := authed.Group("/user")
	{
		user.GET("/profile", deps.Auth.GetProfile)
		user.POST("/image-studio/handoff", deps.Auth.CreateImageStudioHandoff)
		user.GET("/balance", deps.Auth.GetBalance)
		user.GET("/transactions", deps.Auth.GetTransactions)
		user.GET("/stats", deps.Auth.GetUserStats)
		user.GET("/stats/tokens", handler.GetUserTokenStats)
		user.GET("/model-credits", deps.Auth.GetModelCredits)
		user.GET("/channels", deps.Auth.ListModels)
		user.GET("/model-availability", handler.GetModelAvailability)
		user.GET("/model-groups", deps.Auth.ListAvailableModelGroups)
		user.GET("/apikeys", deps.Auth.ListAPIKeys)
		user.POST("/apikeys", deps.Auth.CreateAPIKey)
		user.GET("/apikeys/:id/model-groups", deps.Auth.ListAPIKeyModelGroups)
		user.PUT("/apikeys/:id/model-groups", deps.Auth.ReplaceAPIKeyModelGroups)
		user.DELETE("/apikeys/:id", deps.Auth.DeleteAPIKey)
		user.PUT("/password", deps.Auth.ChangePassword)
		user.POST("/bind-email", deps.Auth.BindEmail)
		user.POST("/reference-images", handler.UploadReferenceImage)
		user.POST("/cards/redeem", handler.RedeemCard)
		user.GET("/cards/redeem-history", handler.GetRedeemHistory)
		user.GET("/payment-orders", handler.GetUserPaymentOrders)
		user.GET("/invite", handler.GetInviteInfo)
		user.GET("/invite/list", handler.GetInviteeList)
		user.POST("/invite/bind", handler.BindInviteCode)
		user.POST("/invite/convert", handler.ConvertFrozenBalance)
		user.GET("/payment-qr", handler.GetPaymentQR)
		user.PUT("/payment-qr", handler.SavePaymentQR)
		user.POST("/withdraw", handler.SubmitWithdraw)
		user.GET("/withdraw/history", handler.ListWithdrawHistory)
		user.GET("/coupons/validate", handler.ValidateCoupon)
	}

	authed.POST("/pay/epay/create", handler.CreateEpayOrder)
	authed.POST("/pay/apply/create", handler.CreatePayApplyOrder)
	authed.POST("/pay/shouqianba/create", handler.CreateShouqianbaOrder)
	authed.GET("/pay/order/status", handler.GetPaymentOrderStatus)
}
