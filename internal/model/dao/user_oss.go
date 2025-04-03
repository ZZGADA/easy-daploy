package dao

import (
	"time"

	"gorm.io/gorm"
)

type UserOss struct {
	ID              uint32         `gorm:"column:id;primaryKey" json:"id"`
	UserID          uint32         `gorm:"column:user_id;not null" json:"user_id"`
	AccessKeyID     string         `gorm:"column:access_key_id;type:varchar(255);not null" json:"access_key_id"`
	AccessKeySecret string         `gorm:"column:access_key_secret;type:varchar(255);not null" json:"access_key_secret"`
	Bucket          string         `gorm:"column:bucket;type:varchar(255);not null" json:"bucket"`
	Region          string         `gorm:"column:region;type:varchar(255);not null" json:"region"`
	CreatedAt       *time.Time     `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt       *time.Time     `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at"`
}

func NewUserOssDao(db *gorm.DB) *UserOssDao {
	return &UserOssDao{db: db}
}

type UserOssDao struct {
	db *gorm.DB
}

// Create 创建 OSS 访问信息
func (d *UserOssDao) Create(oss *UserOss) error {
	return d.db.Create(oss).Error
}

// Update 更新 OSS 访问信息
func (d *UserOssDao) Update(userID uint, accessKeyID, accessKeySecret, bucket, region string) error {
	return d.db.Model(&UserOss{}).Where("user_id = ? and deleted_at IS NULL", userID).Updates(map[string]interface{}{
		"access_key_id":     accessKeyID,
		"access_key_secret": accessKeySecret,
		"bucket":            bucket,
		"region":            region,
	}).Error
}

// QueryByUserID 根据用户 ID 查询 OSS 访问信息
func (d *UserOssDao) QueryByUserID(userID uint) (*UserOss, error) {
	var oss UserOss
	err := d.db.Where("user_id = ? and deleted_at IS NULL", userID).First(&oss).Error
	if err != nil {
		return nil, err
	}
	return &oss, nil
}

// DeleteByUserID 根据用户 ID 删除 OSS 访问信息
func (d *UserOssDao) DeleteByUserID(userID uint) error {
	return d.db.Where("user_id = ? and deleted_at IS NULL", userID).Delete(&UserOss{}).Error
}
