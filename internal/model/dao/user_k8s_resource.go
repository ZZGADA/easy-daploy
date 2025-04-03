package dao

import (
	"time"

	"gorm.io/gorm"
)

type UserK8sResource struct {
	Id               uint32         `gorm:"column:id;type:int UNSIGNED;primaryKey;not null;" json:"id"`
	UserID           uint32         `gorm:"column:user_id;not null" json:"user_id"`
	RepositoryID     string         `gorm:"column:repository_id;not null" json:"repository_id"`
	ResourceType     string         `gorm:"column:resource_type;not null" json:"resource_type"`
	OssURL           string         `gorm:"column:oss_url;not null" json:"oss_url"`
	FileName         string         `gorm:"column:file_name;not null" json:"file_name"`
	IsUpdate         bool           `gorm:"column:is_update;not null" json:"is_update"`
	FatherResourceId uint32         `gorm:"column:father_resource_id" json:"father_resource_id"`
	CreatedAt        *time.Time     `gorm:"column:created_at;type:datetime;not null;" json:"created_at"`
	UpdatedAt        *time.Time     `gorm:"column:updated_at;type:datetime;not null;" json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"column:deleted_at;type:datetime;default:NULL;" json:"deleted_at"`
}

func (UserK8sResource) TableName() string {
	return "user_k8s_resource"
}

func NewUserK8sResourceDao(db *gorm.DB) *UserK8sResourceDao {
	return &UserK8sResourceDao{db: db}
}

type UserK8sResourceDao struct {
	db *gorm.DB
}

// BeginTx 开启事务
func (d *UserK8sResourceDao) BeginTx() *gorm.DB {
	return d.db.Begin()
}

// Create 创建 K8s 资源配置
func (d *UserK8sResourceDao) Create(resource *UserK8sResource) error {
	return d.db.Create(resource).Error
}

// Delete 删除 K8s 资源配置（软删除）
func (d *UserK8sResourceDao) Delete(id uint) error {
	return d.db.Delete(&UserK8sResource{}, id).Error
}

func (d *UserK8sResourceDao) UpdateIsUpdate(resource *UserK8sResource) error {
	return d.db.Model(&UserK8sResource{}).Where("id = ?", resource.Id).Updates(resource).Error
}

func (d *UserK8sResourceDao) QueryById(id uint32) (UserK8sResource, error) {
	var resources UserK8sResource
	err := d.db.Where("id = ? and  deleted_at IS NULL", id).First(&resources).Error
	return resources, err
}

// QueryByRepositoryAndType 根据仓库ID和资源类型查询配置列表
func (d *UserK8sResourceDao) QueryByRepositoryAndType(repositoryID string, resourceType string) ([]UserK8sResource, error) {
	var resources []UserK8sResource
	err := d.db.Where("repository_id = ? AND resource_type = ? and is_update = false and deleted_at IS NULL", repositoryID, resourceType).Find(&resources).Error
	return resources, err
}

// QueryByRepositoryALL 根据仓库ID查询配置列表
func (d *UserK8sResourceDao) QueryByRepositoryALL(repositoryID string) ([]UserK8sResource, error) {
	var resources []UserK8sResource
	err := d.db.Where("repository_id = ? and is_update = false and deleted_at IS NULL", repositoryID).Find(&resources).Error
	return resources, err
}
