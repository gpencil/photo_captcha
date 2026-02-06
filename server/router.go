package server

import (
	"github.com/gin-gonic/gin"
)

// SetupRouter 配置路由
func SetupRouter() *gin.Engine {
	router := gin.Default()

	// CORS中间件
	router.Use(CORSMiddleware())

	// API路由
	api := router.Group("/api")
	{
		captchaGroup := api.Group("/captcha")
		{
			captchaGroup.GET("/generate", GenerateCaptchaHandler)
			captchaGroup.POST("/verify", VerifyCaptchaHandler)
		}
	}

	// 首页
	router.GET("/", IndexHandler)
	router.GET("/index.html", IndexHandler)

	return router
}

// CORSMiddleware CORS中间件
func CORSMiddleware() gin.HandlerFunc {
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
