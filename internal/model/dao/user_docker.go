package dao

import (
	"gorm.io/gorm"
	"time"
)

// UserDocker 结构体
type UserDocker struct {
	ID        uint       `gorm:"column:id;type:int UNSIGNED;primaryKey;not null;" json:"id"`
	UserID    uint       `gorm:"column:user_id;type:int UNSIGNED;not null;index:idx_user_id" json:"user_id"` // 用户ID
	Server    string     `gorm:"column:server;type:varchar(255);not null" json:"server"`                     // Docker仓库地址
	Username  string     `gorm:"column:user_name;type:varchar(255);not null" json:"username"`                // Docker仓库用户名
	Password  string     `gorm:"column:password;type:varchar(255);not null" json:"password"`                 // Docker仓库密码
	Comment   string     `gorm:"column:comment;type:text; null" json:"comment"`                              // 用户备注
	IsDefault bool       `gorm:"column:is_default;type:bool;default:false" json:"is_default"`                // 是否为默认账号
	CreatedAt *time.Time `gorm:"column:created_at;type:datetime;not null;" json:"created_at"`
	UpdatedAt *time.Time `gorm:"column:updated_at;type:datetime;not null;" json:"updated_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at;type:datetime;" json:"deleted_at"`
}

// UserDockerDao Docker账号管理的DAO接口
type UserDockerDao interface {
	Create(docker *UserDocker) error
	Update(docker *UserDocker) error
	Delete(id uint) error
	GetByID(id uint) (*UserDocker, error)
	GetByUserID(userID uint) ([]*UserDocker, error)
	SetDefault(userID, dockerID uint) error
	CountByUserID(userID uint) (int64, error)
}

func (UserDocker) TableName() string {
	return "user_docker"
}

// UserDockerDaoImpl Docker账号管理的DAO实现
type UserDockerDaoImpl struct {
	db *gorm.DB
}

// NewUserDockerDao 创建 UserDockerDao 实例
func NewUserDockerDao(db *gorm.DB) UserDockerDao {
	return &UserDockerDaoImpl{
		db: db,
	}
}

// Create 创建Docker账号
func (d *UserDockerDaoImpl) Create(docker *UserDocker) error {
	return d.db.Create(docker).Error
}

// Update 更新Docker账号
func (d *UserDockerDaoImpl) Update(docker *UserDocker) error {
	return d.db.Model(docker).Updates(docker).Error
}

// Delete 删除Docker账号
func (d *UserDockerDaoImpl) Delete(id uint) error {

	return d.db.Model(&UserGithub{}).Where("id = ? and deleted_at IS NULL", id).Update("deleted_at", time.Now()).Error
}

// GetByID 根据ID获取Docker账号
func (d *UserDockerDaoImpl) GetByID(id uint) (*UserDocker, error) {
	var docker UserDocker
	err := d.db.First(&docker, id).Error
	if err != nil {
		return nil, err
	}
	return &docker, nil
}

// GetByUserID 获取用户的所有Docker账号
func (d *UserDockerDaoImpl) GetByUserID(userID uint) ([]*UserDocker, error) {
	var dockers []*UserDocker
	err := d.db.Where("user_id = ? and deleted_at IS NULL", userID).
		Order("is_default DESC, id ASC").
		Find(&dockers).Error
	if err != nil {
		return nil, err
	}
	return dockers, nil
}

// SetDefault 设置默认Docker账号
func (d *UserDockerDaoImpl) SetDefault(userID, dockerID uint) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		// 先将所有账号设置为非默认
		if err := tx.Model(&UserDocker{}).
			Where("user_id = ? and deleted_at IS NULL", userID).
			Update("is_default", false).Error; err != nil {
			return err
		}

		// 将指定账号设置为默认
		return tx.Model(&UserDocker{}).
			Where("id = ? AND user_id = ? and deleted_at IS NULL", dockerID, userID).
			Update("is_default", true).Error
	})
}

// CountByUserID 统计用户的Docker账号数量
func (d *UserDockerDaoImpl) CountByUserID(userID uint) (int64, error) {
	var count int64
	err := d.db.Model(&UserDocker{}).Where("user_id = ? and deleted_at IS NULL", userID).Count(&count).Error
	return count, err
}
