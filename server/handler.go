package server

import (
	"net/http"
	"strconv"

	"github.com/gpencil/photo_captcha/captcha"

	"github.com/gin-gonic/gin"
)

// GenerateCaptchaHandler 生成验证码处理器
func GenerateCaptchaHandler(c *gin.Context) {
	sliderCaptcha, err := captcha.Generate()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to generate captcha: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"id":         sliderCaptcha.ID,
			"background": sliderCaptcha.Background,
			"slider":     sliderCaptcha.Slider,
			"positionY":  sliderCaptcha.PositionY,
		},
	})
}

// VerifyCaptchaRequest 验证请求结构
type VerifyCaptchaRequest struct {
	ID string `json:"id" binding:"required"`
	X  string `json:"x" binding:"required"`
}

// VerifyCaptchaHandler 验证滑块位置处理器
func VerifyCaptchaHandler(c *gin.Context) {
	var req VerifyCaptchaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	// 将X坐标字符串转换为整数
	userX, err := strconv.Atoi(req.X)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid x coordinate",
		})
		return
	}

	// 验证
	success, err := captcha.VerifyWithTolerance(req.ID, userX)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": err.Error(),
			"data": gin.H{
				"success": false,
			},
		})
		return
	}

	if success {
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "Verification successful",
			"data": gin.H{
				"success": true,
			},
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "Verification failed",
			"data": gin.H{
				"success": false,
			},
		})
	}
}

// IndexHandler 首页处理器
func IndexHandler(c *gin.Context) {
	c.File("./web/index.html")
}
