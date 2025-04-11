package user_manage

import (
	"fmt"
	"github.com/ZZGADA/easy-deploy/internal/config"
	"github.com/ZZGADA/easy-deploy/internal/define"
	"github.com/go-redis/redis/v8"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"github.com/ZZGADA/easy-deploy/internal/model/conf"
	"github.com/ZZGADA/easy-deploy/internal/model/dao"
	"github.com/ZZGADA/easy-deploy/internal/model/dto"
	"github.com/sirupsen/logrus"
	"gopkg.in/gomail.v2"
)

// RegisterService 处理用户注册的业务逻辑
func RegisterService(dto dto.RegisterDTO) (interface{}, error) {
	// 验证邮箱格式
	if !strings.Contains(dto.Email, "@") {
		logrus.Warnf("Invalid email format: %s", dto.Email)
		return nil, fmt.Errorf("Invalid email format")
	}

	// 验证密码格式
	if !isValidPassword(dto.Password) {
		logrus.Warnf("Invalid password format for email: %s", dto.Email)
		return nil, fmt.Errorf("Password must contain uppercase, lowercase and special characters")
	}

	// 检查邮箱是否已注册
	existingUser, err := dao.GetUserByEmail(dto.Email)
	if err == nil && existingUser != nil {
		logrus.Warnf("Email already registered: %s", dto.Email)
		return nil, fmt.Errorf("Email already registered")
	}

	// 生成验证码
	code := generateCode()
	// 将验证码存入Redis，过期时间5分钟
	err = conf.RedisClient.Set(conf.RedisClient.Context(), fmt.Sprintf(define.VerifyCodeEmail, dto.Email), code, 5*time.Minute).Err()
	if err != nil {
		logrus.Errorf("Failed to save verification code for email %s: %v", dto.Email, err)
		return nil, fmt.Errorf("Failed to save verification code")
	}

	// 将密码存入Redis，过期时间5分钟
	err = conf.RedisClient.Set(conf.RedisClient.Context(), fmt.Sprintf(define.RegisterPassword, dto.Email), dto.Password, 5*time.Minute).Err()
	if err != nil {
		logrus.Errorf("Failed to save password for email %s: %v", dto.Email, err)
		return nil, fmt.Errorf("Failed to save password")
	}

	// 发送验证码邮件
	if err := sendVerificationEmail(dto.Email, code); err != nil {
		logrus.Errorf("Failed to send verification email to %s: %v", dto.Email, err)
		return nil, fmt.Errorf("Failed to send verification email")
	}

	logrus.Infof("Verification code sent to email: %s", dto.Email)
	return gin.H{"message": "Verification code sent to your email. Please complete registration with the code."}, nil
}

// LoginService 处理用户登录的业务逻辑
func LoginService(dto dto.LoginDTO) (interface{}, error) {
	user, err := dao.GetUserByEmail(dto.Email)
	if err != nil {
		logrus.Warnf("User not found with email: %s", dto.Email)
		return nil, fmt.Errorf("Invalid email or password")
	}

	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(dto.Password))
	if err != nil {
		logrus.Warnf("Invalid password for email: %s", dto.Email)
		return nil, fmt.Errorf("Invalid email or password")
	}

	// 幂等校验
	tokenHasDone, err := conf.RedisClient.Get(conf.RedisClient.Context(), fmt.Sprintf(define.UserEmailToken, user.Email)).Result()
	if err != nil && err != redis.Nil {
		logrus.Warnf("redis broken token for email %s: %v", dto.Email, err)
		return nil, fmt.Errorf("system goes wrong")
	}

	// token 不为空 表示用户已经登陆
	if tokenHasDone != "" {
		return gin.H{"message": "Login successful", "token": tokenHasDone}, nil
	}

	// 将用户ID存入Redis
	token := generateToken()
	err = conf.RedisClient.SetNX(conf.RedisClient.Context(), fmt.Sprintf(define.UserToken, token), user.Id, 24*time.Hour).Err()
	err = conf.RedisClient.SetNX(conf.RedisClient.Context(), fmt.Sprintf(define.UserEmailToken, user.Email), token, 24*time.Hour).Err()
	if err != nil {
		logrus.Errorf("Failed to save user session for email %s: %v", user.Email, err)
		return nil, fmt.Errorf("Failed to save user session")
	}

	logrus.Infof("User logged in successfully with email: %s", user.Email)
	return gin.H{"message": "Login successful", "token": token}, nil
}

// VerifyCodeService 处理验证码校验的业务逻辑
func VerifyCodeService(dto dto.VerifyCodeDTO) (interface{}, error) {
	verifyKey := fmt.Sprintf(define.VerifyCodeEmail, dto.Email)
	storedCode, err := conf.RedisClient.Get(conf.RedisClient.Context(), verifyKey).Result()
	if err != nil {
		logrus.Warnf("Verification code has expired or is incorrect for email: %s", dto.Email)
		return nil, fmt.Errorf("Verification code has expired or is incorrect")
	}
	if storedCode != dto.Code {
		logrus.Warnf("Verification code is incorrect for email: %s", dto.Email)
		return nil, fmt.Errorf("Verification code is incorrect")
	}

	// 从Redis获取密码
	passwordKey := fmt.Sprintf(define.RegisterPassword, dto.Email)
	password, err := conf.RedisClient.Get(conf.RedisClient.Context(), passwordKey).Result()
	if err != nil {
		logrus.Warnf("Password has expired for email: %s", dto.Email)
		return nil, fmt.Errorf("Registration session has expired")
	}

	// 对密码进行加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logrus.Errorf("Failed to hash password for email %s: %v", dto.Email, err)
		return nil, fmt.Errorf("Failed to process password")
	}

	// 创建用户
	user := &dao.Users{
		Email:    dto.Email,
		Password: string(hashedPassword),
	}
	if err := dao.CreateUser(user); err != nil {
		logrus.Errorf("Failed to create user for email %s: %v", dto.Email, err)
		return nil, fmt.Errorf("Failed to create user")
	}

	// 删除Redis中的验证码和密码
	conf.RedisClient.Del(conf.RedisClient.Context(), verifyKey, passwordKey)

	logrus.Infof("User registered successfully with email: %s", dto.Email)
	return gin.H{"message": "Registration completed successfully"}, nil
}

// isValidPassword 验证密码格式
func isValidPassword(password string) bool {
	hasUpper := false
	hasLower := false
	hasSpecial := false
	for _, char := range password {
		if char >= 'A' && char <= 'Z' {
			hasUpper = true
		} else if char >= 'a' && char <= 'z' {
			hasLower = true
		} else if (char >= '!' && char <= '/') || (char >= ':' && char <= '@') || (char >= '[' && char <= '`') || (char >= '{' && char <= '~') {
			hasSpecial = true
		}
	}
	return hasUpper && hasLower && hasSpecial
}

// generateCode 生成6位数字验证码
func generateCode() string {
	return fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)
}

// generateToken 生成简单的token
func generateToken() string {
	return fmt.Sprintf("%x", time.Now().UnixNano())
}

// sendVerificationEmail 发送验证码邮件
func sendVerificationEmail(email, code string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", config.GlobalConfig.Smtp.From)
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Verification Code for ZZGEDA Registration")
	m.SetBody("text/plain", fmt.Sprintf("Your verification code is: %s", code))

	d := gomail.NewDialer(
		config.GlobalConfig.Smtp.Host,
		config.GlobalConfig.Smtp.Port,
		config.GlobalConfig.Smtp.User,
		config.GlobalConfig.Smtp.Password)

	return d.DialAndSend(m)
}
