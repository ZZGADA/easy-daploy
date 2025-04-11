package dao

import (
	"context"
	"time"

	"github.com/ZZGADA/easy-deploy/internal/model/conf"
	"gorm.io/gorm"
)

// Users 定义用户结构体
type Users struct {
	Id        uint32     `gorm:"column:id;type:bigint;primaryKey;" json:"id"`
	Email     string     `gorm:"column:email;type:varchar(255);not null;" json:"email"`
	Password  string     `gorm:"column:password;type:varchar(255);not null;" json:"password"`
	TeamID    uint32     `gorm:"column:team_id;type:bigint;" json:"team_id"`
	CreatedAt *time.Time `gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP;" json:"created_at"`
	UpdatedAt *time.Time `gorm:"column:updated_at;type:timestamp;default:CURRENT_TIMESTAMP;" json:"updated_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at;type:timestamp;" json:"deleted_at"`
}

type UsersGithub struct {
	Id         uint32 `gorm:"column:id;type:bigint;primaryKey;" json:"id"`
	Email      string `gorm:"column:email;type:varchar(255);not null;" json:"email"`
	Password   string `gorm:"column:password;type:varchar(255);not null;" json:"password"`
	TeamID     uint32 `gorm:"column:team_id;type:bigint;" json:"team_id"`
	GithubName string `gorm:"column:name;type:varchar(255);not null;" json:"name"`
}

type UserGithubFull struct {
}

func (Users) TableName() string {
	return "users"
}

// NewUsersDao 创建用户DAO
func NewUsersDao(db *gorm.DB) *UsersDao {
	return &UsersDao{db: db}
}

type UsersDao struct {
	db *gorm.DB
}

// CreateUser 创建用户
func CreateUser(user *Users) error {
	return conf.DB.Create(user).Error
}

// UpdateUser 更新用户
func (d *UsersDao) UpdateUser(user *Users) error {
	return d.db.Model(user).Updates(map[string]interface{}{
		"team_id":    user.TeamID,
		"updated_at": time.Now(),
	}).Error
}

// UpdateUserTx 更新用户
func (d *UsersDao) UpdateUserTx(tx *gorm.DB, user *Users) error {
	return tx.Model(user).Updates(map[string]interface{}{
		"team_id":    user.TeamID,
		"updated_at": time.Now(),
	}).Error
}

// DeleteUser 删除用户（软删除）
func (d *UsersDao) DeleteUser(userID uint32) error {
	return d.db.Model(&Users{}).Where("id = ?", userID).Update("deleted_at", time.Now()).Error
}

// GetUserByID 根据ID获取用户
func (d *UsersDao) GetUserByID(userID uint32) (*Users, error) {
	var user Users
	err := d.db.Where("id = ? AND deleted_at IS NULL", userID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail 根据邮箱获取用户
func (d *UsersDao) GetUserByEmail(email string) (*Users, error) {
	var user Users
	err := d.db.Where("email = ? AND deleted_at IS NULL", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail 根据邮箱获取用户
func GetUserByEmail(email string) (*Users, error) {
	var user Users
	err := conf.DB.Where("email =? and deleted_at IS NULL", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func GetUserByID(id uint) (Users, error) {
	var user Users
	err := conf.DB.Where("id=? and deleted_at IS NULL", id).First(&user).Error
	if err != nil {
		return user, err
	}
	return user, nil
}

// GetUsersByTeamID 根据团队ID获取用户列表
func (d *UsersDao) GetUsersByTeamID(teamID uint32) ([]*Users, error) {
	var users []*Users
	err := d.db.Where("team_id = ? AND deleted_at IS NULL", teamID).Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

// GetUsersByTeamIDGithub 根据团队ID获取用户列表
func (d *UsersDao) GetUsersByTeamIDGithub(teamID uint32) ([]*UsersGithub, error) {
	var users []*UsersGithub
	err := d.db.Select(
		"users.email,"+
			"users.id,"+
			"users.team_id,"+
			"user_github.name").Table("users").Joins("join user_github on user_github.user_id = users.id ").
		Where("users.team_id = ? AND users.deleted_at IS NULL AND user_github.deleted_at IS NULL", teamID).
		Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

// QueryUsers 查询用户列表
func (d *UsersDao) QueryUsers(email string, teamID string) ([]*Users, error) {
	var users []*Users
	query := d.db.Where("deleted_at IS NULL")

	if email != "" {
		query = query.Where("email LIKE ?", "%"+email+"%")
	}

	if teamID != "" {
		query = query.Where("team_id = ?", teamID)
	}

	err := query.Find(&users).Error
	if err != nil {
		return nil, err
	}

	return users, nil
}

// UpdateUserTeamID 更新用户团队ID
func (d *UsersDao) UpdateUserTeamID(ctx context.Context, userID uint32, teamID *uint32) error {
	return d.db.WithContext(ctx).Model(&Users{}).Where("id = ?", userID).Update("team_id", teamID).Error
}

// BatchUpdateTeamID 批量更新用户团队ID
func (d *UsersDao) BatchUpdateTeamID(ctx context.Context, teamID uint32, newTeamID *uint32) error {
	return d.db.WithContext(ctx).Model(&Users{}).Where("team_id = ?", teamID).Update("team_id", newTeamID).Error
}

// BatchUpdateTeamIDTx 批量更新用户团队ID
func (d *UsersDao) BatchUpdateTeamIDTx(tx *gorm.DB, ctx context.Context, teamID uint32, newTeamID *uint32) error {
	return tx.WithContext(ctx).Model(&Users{}).Where("team_id = ?", teamID).Update("team_id", newTeamID).Error
}

// GetUserWithGithubInfo 根据用户 ID 获取用户及其 GitHub 信息
func (d *UsersDao) GetUserWithGithubInfo(ctx context.Context, userID uint) (*UserWithGithubInfo, error) {
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

// GetUserListWithGithubInfo 根据用户 ID 获取用户及其 GitHub 信息
func (d *UsersDao) GetUserListWithGithubInfo(ctx context.Context, userIDs []uint32) ([]*UserWithGithubInfo, error) {
	var userWithGithubInfoList []*UserWithGithubInfo
	err := d.db.WithContext(ctx).
		Table("users").
		Select("users.id, users.email, user_github.github_id, user_github.name").
		Joins("JOIN user_github ON users.id = user_github.user_id").
		Where("users.id in ? AND users.deleted_at IS NULL AND user_github.deleted_at IS NULL", userIDs).
		Find(&userWithGithubInfoList).Error
	if err != nil {
		return nil, err
	}
	return userWithGithubInfoList, nil
}
