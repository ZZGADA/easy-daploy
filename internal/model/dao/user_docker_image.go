package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type UserDockerImage struct {
	Id            uint32    `gorm:"column:id;type:int(10) UNSIGNED;primaryKey;not null;" json:"id"`
	UserId        uint32    `gorm:"column:user_id;type:int(10) UNSIGNED;not null;" json:"user_id"`
	DockerfileId  uint32    `gorm:"column:dockerfile_id;type:int(10) UNSIGNED;not null;" json:"dockerfile_id"`
	FullImageName string    `gorm:"column:full_image_name;type:varchar(255);not null;" json:"full_image_name"`
	ImageName     string    `gorm:"column:image_name;type:varchar(255);not null;" json:"image_name"`
	CreatedAt     time.Time `gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP;" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at;type:timestamp;default:CURRENT_TIMESTAMP;" json:"updated_at"`
	DeletedAt     time.Time `gorm:"column:deleted_at;type:timestamp;default:NULL;" json:"deleted_at"`
}

func (UserDockerImage) TableName() string {
	return "user_docker_image"
}

type UserDockerImageDao struct {
	db *gorm.DB
}

func NewUserDockerImageDao(db *gorm.DB) *UserDockerImageDao {
	return &UserDockerImageDao{db: db}
}

// Create 创建记录
func (d *UserDockerImageDao) Create(ctx context.Context, image *UserDockerImage) error {
	return d.db.WithContext(ctx).Create(image).Error
}

// GetByDockerfileID 根据 DockerfileID 查询记录
func (d *UserDockerImageDao) GetByDockerfileID(ctx context.Context, dockerfileID uint32) ([]*UserDockerImage, error) {
	var images []*UserDockerImage
	err := d.db.WithContext(ctx).Where("dockerfile_id = ? and deleted_at IS NULL", dockerfileID).Order("id DESC").Find(&images).Error
	return images, err
}
