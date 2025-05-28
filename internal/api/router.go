package api

import (
	"giftredeem/internal/middleware"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// SetupRouter configures the API routes
func SetupRouter() *gin.Engine {
	r := gin.Default()

	// Set up CORS if needed
	r.Use(corsMiddleware())

	// API routes
	api := r.Group("/api")
	{
		// Auth routes
		authHandler := NewAuthHandler()
		auth := api.Group("/auth")
		{
			auth.GET("/providers", authHandler.GetProviders)
			auth.GET("/login/:provider", authHandler.Login)
			auth.GET("/callback/:provider", authHandler.Callback)
			auth.POST("/verify/:provider", authHandler.VerifyCode) // 新API：验证授权码
			auth.GET("/profile", middleware.AuthMiddleware(), authHandler.GetUserProfile)
		}

		// Benefit routes
		benefitHandler := NewBenefitHandler()
		benefits := api.Group("/benefits")
		{
			// Protected routes (require authentication)
			benefits.Use(middleware.AuthMiddleware())
			{
				benefits.POST("", benefitHandler.CreateBenefit)
				benefits.GET("/my", benefitHandler.GetUserBenefits)
				benefits.PUT("/:uuid/status", benefitHandler.UpdateBenefitStatus)
				benefits.GET("/:uuid/claims", benefitHandler.GetBenefitClaims)
			}
		}

		// Claim routes
		claims := api.Group("/claims")
		{
			claims.Use(middleware.AuthMiddleware())
			claims.GET("/my", benefitHandler.GetUserClaims)
		}

		// Public benefit routes
		claim := api.Group("/claim")
		{
			// Optional auth for viewing, required for claiming
			claim.GET("/:uuid", middleware.OptionalAuthMiddleware(), benefitHandler.GetBenefitByUUID)
			claim.POST("/:uuid", middleware.AuthMiddleware(), benefitHandler.ClaimBenefit)
		}
	}

	// 提供静态文件服务
	// 注意：将"./frontend/dist"替换为你的Vue构建文件的实际路径
	r.Static("/assets", "./frontend/dist/assets")
	r.StaticFile("/favicon.ico", "./frontend/dist/favicon.ico")

	// 处理所有前端路由 - 将非API请求重定向到index.html
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// 如果请求的不是API路径，返回前端应用的index.html
		if !strings.HasPrefix(path, "/api/") {
			c.File("./frontend/dist/index.html")
			return
		}

		// API路径的404处理
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "API endpoint not found",
		})
	})

	return r
}

// corsMiddleware configures CORS for the API
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
