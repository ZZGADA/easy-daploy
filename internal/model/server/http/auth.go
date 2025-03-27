package http

import (
	"net/http"

	"github.com/ZZGADA/easy-deploy/internal/model/dto"
	"github.com/ZZGADA/easy-deploy/internal/model/service/user_manage"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Register 处理用户注册请求
func Register(c *gin.Context) {
	var registerDTO dto.RegisterDTO
	if err := c.ShouldBindJSON(&registerDTO); err != nil {
		logrus.Errorf("Invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	result, err := user_manage.RegisterService(registerDTO)
	if err != nil {
		logrus.Errorf("Registration failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Login 处理用户登录请求
func Login(c *gin.Context) {
	var loginDTO dto.LoginDTO
	if err := c.ShouldBindJSON(&loginDTO); err != nil {
		logrus.Errorf("Invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	result, err := user_manage.LoginService(loginDTO)
	if err != nil {
		logrus.Errorf("Login failed: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// VerifyCode 处理验证码校验请求
func VerifyCode(c *gin.Context) {
	var verifyCodeDTO dto.VerifyCodeDTO
	if err := c.ShouldBindJSON(&verifyCodeDTO); err != nil {
		logrus.Errorf("Invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	result, err := user_manage.VerifyCodeService(verifyCodeDTO)
	if err != nil {
		logrus.Errorf("Verification code verification failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
