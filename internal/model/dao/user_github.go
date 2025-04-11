package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// UserGithub GitHub 用户信息表
type UserGithub struct {
	Id                       uint32     `gorm:"column:id;type:int UNSIGNED;primaryKey;not null;" json:"id"`
	UserId                   uint32     `gorm:"column:user_id;type:int UNSIGNED;not null;" json:"user_id"`
	GithubId                 uint32     `gorm:"column:github_id;type:int UNSIGNED;not null;" json:"github_id"`
	Login                    string     `gorm:"column:login;type:varchar(255);not null;" json:"login"`
	Name                     string     `gorm:"column:name;type:varchar(255);" json:"name"`
	Email                    string     `gorm:"column:email;type:varchar(255);" json:"email"`
	AvatarUrl                string     `gorm:"column:avatar_url;type:varchar(255);" json:"avatar_url"`
	AccessToken              string     `gorm:"column:access_token;type:varchar(255);not null;" json:"access_token"`
	DeveloperToken           string     `gorm:"column:developer_token;type:varchar(255);" json:"developer_token"`
	DeveloperTokenComment    string     `gorm:"column:developer_token_comment;type:varchar(255);" json:"developer_token_comment"`
	DeveloperTokenExpireTime *time.Time `gorm:"column:developer_token_expire_time;" json:"developer_token_expire_time"`
	DeveloperRepositoryName  string     `gorm:"column:developer_repository_name;type:varchar(255);" json:"developer_repository_name"`
	CreatedAt                *time.Time `gorm:"column:created_at;type:datetime;not null;" json:"created_at"`
	UpdatedAt                *time.Time `gorm:"column:updated_at;type:datetime;not null;" json:"updated_at"`
	DeletedAt                *time.Time `gorm:"column:deleted_at;type:datetime;default:NULL;" json:"deleted_at"`
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
	userGithub := UserGithub{}
	err := d.db.WithContext(ctx).Where("user_id = ? and deleted_at IS NULL", userID).First(&userGithub).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return &userGithub, err
	}
	return &userGithub, nil
}

// GetByGithubID 根据 GitHub ID 获取用户信息
func (d *UserGithubDao) GetByGithubID(ctx context.Context, githubID uint) (*UserGithub, error) {
	var userGithub UserGithub
	err := d.db.WithContext(ctx).Where("github_id = ? and deleted_at IS NULL", githubID).First(&userGithub).Error
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
	return d.db.WithContext(ctx).Model(&UserGithub{}).Where("user_id = ? and deleted_at IS NULL", userID).Update("deleted_at", time.Now()).Error
}

// SaveDeveloperToken 保存开发者令牌信息
func (d *UserGithubDao) SaveDeveloperToken(ctx context.Context, userID uint, token string, expireTime time.Time, comment string) error {
	return d.db.WithContext(ctx).Model(&UserGithub{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"developer_token":             token,
			"developer_token_comment":     comment,
			"developer_token_expire_time": expireTime,
		}).Error
}

// UpdateDeveloperToken 更新开发者令牌信息
func (d *UserGithubDao) UpdateDeveloperToken(ctx context.Context, userID uint, updates map[string]interface{}) error {
	return d.db.WithContext(ctx).Model(&UserGithub{}).
		Where("user_id = ?", userID).
		Updates(updates).Error
}

// GetUserWithGithubInfo 根据用户 ID 获取用户及其 GitHub 信息
func (d *UserGithubDao) GetUserWithGithubInfo(ctx context.Context, userID uint) (*UserWithGithubInfo, error) {
	var userWithGithubInfo UserWithGithubInfo
	err := d.db.WithContext(ctx).
		Table("users").
		Select("users.id, users.email, user_github.github_id, user_github.name").
		Joins("JOIN user_github ON users.id = user_github.user_id").
		Where("users.id = ? AND users.deleted_at IS NULL AND user_github.deleted_at IS NULL", userID).
		First(&userWithGithubInfo).Error
	if err != nil {
		return nil, err
	}
	return &userWithGithubInfo, nil
}

// UserWithGithubInfo 用户及其 GitHub 信息结构体
type UserWithGithubInfo struct {
	ID       uint   `json:"id"`
	Email    string `json:"email"`
	GithubID uint32 `json:"github_id"`
	Name     string `json:"name"`
}
