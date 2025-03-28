package dao

import (
	"github.com/ZZGADA/easy-deploy/internal/model/conf"
	"time"
)

// User 定义用户结构体
type Users struct {
	Id        uint32     `gorm:"column:id;type:bigint;primaryKey;" json:"id"`
	Email     string     `gorm:"column:email;type:varchar(255);not null;" json:"email"`
	Password  string     `gorm:"column:password;type:varchar(255);not null;" json:"password"`
	CreatedAt *time.Time `gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP;" json:"created_at"`
	UpdatedAt *time.Time `gorm:"column:updated_at;type:timestamp;default:CURRENT_TIMESTAMP;" json:"updated_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at;type:timestamp;" json:"deleted_at"`
}

func (Users) TableName() string {
	return "users"
}

// CreateUser 创建用户
func CreateUser(user *Users) error {
	return conf.DB.Create(user).Error
}

// GetUserByEmail 根据邮箱获取用户
func GetUserByEmail(email string) (*Users, error) {
	var user Users
	err := conf.DB.Where("email =?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
