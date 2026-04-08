package handlers

import (
	"xykitchen/backend/internal/config"
	"xykitchen/backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册 /api 下路由（xyKitchen 精简版）
func RegisterRoutes(engine *gin.Engine, cfg *config.Config) {
	api := engine.Group("/api")

	api.GET("/health", Health)

	api.GET("/media/cos", MediaCosStream)
	api.POST("/analytics/page-view", AnalyticsPageView)
	api.GET("/admin/page-visit-stats", middleware.Auth(cfg), middleware.Admin(), AnalyticsStats)

	auth := api.Group("/auth")
	{
		auth.POST("/send-code", func(c *gin.Context) { authSendCode(c, cfg) })
		auth.POST("/send-sms-code", func(c *gin.Context) { authSendSmsCode(c, cfg) })
		auth.POST("/register", func(c *gin.Context) { authRegister(c, cfg) })
		auth.POST("/login", func(c *gin.Context) { authLogin(c, cfg) })
		auth.POST("/bind-phone", middleware.Auth(cfg), func(c *gin.Context) { authBindPhone(c, cfg) })
		auth.GET("/profile", middleware.Auth(cfg), authGetProfile)
		auth.PUT("/profile", middleware.Auth(cfg), authUpdateProfile)
		auth.POST("/upload-avatar", middleware.Auth(cfg), func(c *gin.Context) { authUploadAvatar(c, cfg) })
		auth.GET("/admin/users", middleware.Auth(cfg), middleware.Admin(), authAdminGetUsers)
		auth.DELETE("/admin/users/:userId", middleware.Auth(cfg), middleware.Admin(), authAdminDeleteUser)
	}

	orders := api.Group("/orders")
	{
		orders.POST("/", middleware.Auth(cfg), orderCreate)
		orders.GET("/mine", middleware.Auth(cfg), orderMyOrders)
		orders.GET("/admin/list", middleware.Auth(cfg), middleware.Admin(), orderAdminList)
		orders.GET("/admin/stats", middleware.Auth(cfg), middleware.Admin(), orderAdminStats)
		orders.GET("/:id", middleware.Auth(cfg), orderDetail)
		orders.PUT("/:id/cancel", middleware.Auth(cfg), orderCancel)
		orders.PUT("/admin/:id/status", middleware.Auth(cfg), middleware.Admin(), orderAdminUpdateStatus)
		orders.PUT("/admin/:id/price", middleware.Auth(cfg), middleware.Admin(), orderAdminUpdatePrice)
		orders.POST("/admin/:id/remark", middleware.Auth(cfg), middleware.Admin(), orderAdminAddRemark)
		orders.GET("/admin/:id/logs", middleware.Auth(cfg), middleware.Admin(), orderAdminLogs)
	}

	hc := api.Group("/home-config")
	{
		hc.GET("/", hcList)
		hc.POST("/", middleware.Auth(cfg), middleware.Admin(), hcCreate)
		hc.POST("/upload", middleware.Auth(cfg), middleware.Admin(), func(c *gin.Context) { hcUploadImage(c, cfg) })
		hc.PUT("/:id", middleware.Auth(cfg), middleware.Admin(), hcUpdate)
		hc.DELETE("/:id", middleware.Auth(cfg), middleware.Admin(), hcRemove)
	}

	api.GET("/i18n", I18nList)
	api.POST("/i18n/bulk", middleware.Auth(cfg), middleware.Admin(), I18nBulkUpsert)
	api.PUT("/i18n/:id", middleware.Auth(cfg), middleware.Admin(), I18nUpdate)
	api.DELETE("/i18n/:id", middleware.Auth(cfg), middleware.Admin(), I18nRemove)
}
