package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// UserGithub GitHub 用户信息表
type UserGithub struct {
	ID          uint           `gorm:"primarykey"`
	UserID      uint           `gorm:"not null;index"`       // 关联的用户 ID
	GithubID    uint           `gorm:"not null;uniqueIndex"` // GitHub 用户 ID
	Login       string         `gorm:"size:255;not null"`    // GitHub 登录名
	Name        string         `gorm:"size:255"`             // GitHub 用户名
	Email       string         `gorm:"size:255"`             // GitHub 邮箱
	AvatarURL   string         `gorm:"size:255"`             // GitHub 头像 URL
	AccessToken string         `gorm:"size:255;not null"`    // GitHub 访问令牌
	CreatedAt   *time.Time     `gorm:"not null"`             // 创建时间
	UpdatedAt   *time.Time     `gorm:"not null"`             // 更新时间
	DeletedAt   gorm.DeletedAt // 软删除时间
}

// TableName 指定表名
func (UserGithub) TableName() string {
	return "user_github"
}

// UserGithubDao GitHub 用户信息数据访问对象
type UserGithubDao struct {
	db *gorm.DB
}

// NewUserGithubDao 创建 UserGithubDao 实例
func NewUserGithubDao(db *gorm.DB) *UserGithubDao {
	return &UserGithubDao{db: db}
}

// Create 创建 GitHub 用户信息记录
func (d *UserGithubDao) Create(ctx context.Context, userGithub *UserGithub) error {
	return d.db.WithContext(ctx).Create(userGithub).Error
}

// GetByUserID 根据用户 ID 获取 GitHub 用户信息
func (d *UserGithubDao) GetByUserID(ctx context.Context, userID uint) (*UserGithub, error) {
	var userGithub UserGithub
	err := d.db.WithContext(ctx).Where("user_id = ?", userID).First(&userGithub).Error
	if err != nil {
		return nil, err
	}
	return &userGithub, nil
}

// GetByGithubID 根据 GitHub ID 获取用户信息
func (d *UserGithubDao) GetByGithubID(ctx context.Context, githubID uint) (*UserGithub, error) {
	var userGithub UserGithub
	err := d.db.WithContext(ctx).Where("github_id = ?", githubID).First(&userGithub).Error
	if err != nil {
		return nil, err
	}
	return &userGithub, nil
}

// Update 更新 GitHub 用户信息
func (d *UserGithubDao) Update(ctx context.Context, userGithub *UserGithub) error {
	return d.db.WithContext(ctx).Save(userGithub).Error
}

// Delete 删除 GitHub 用户信息（软删除）
func (d *UserGithubDao) Delete(ctx context.Context, userID uint) error {
	return d.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&UserGithub{}).Error
}
