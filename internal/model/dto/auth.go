package dto

// RegisterDTO 注册请求的 DTO 对象
type RegisterDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginDTO 登录请求的 DTO 对象
type LoginDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// VerifyCodeDTO 验证码校验的 DTO 对象
type VerifyCodeDTO struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}
