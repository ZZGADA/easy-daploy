package dao

import (
	"time"

	"gorm.io/gorm"
)

// UserK8sResourceOperationLog 用户 K8s 资源操作日志
type UserK8sResourceOperationLog struct {
	ID             uint           `gorm:"primaryKey;column:id" json:"id"`
	K8sResourceID  uint           `gorm:"not null;column:k8s_resource_id" json:"k8s_resource_id"`
	UserID         uint           `gorm:"not null;column:user_id" json:"user_id"`
	Namespace      string         `gorm:"size:255;not null;column:namespace" json:"namespace"`
	MetadataName   string         `gorm:"size:255;not null;column:metadata_name" json:"metadata_name"`
	MetadataLabels string         `gorm:"type:text;column:metadata_labels" json:"metadata_labels"`
	OperationType  string         `gorm:"size:50;not null;column:operation_type" json:"operation_type"`
	Command        string         `gorm:"size:500;not null;column:command" json:"command"`
	CreatedAt      *time.Time     `gorm:"column:created_at" json:"created_at"`
	UpdatedAt      *time.Time     `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index;column:deleted_at" json:"deleted_at"`
}

func (UserK8sResourceOperationLog) TableName() string {
	return "user_k8s_resource_operation_logs"
}

// UserK8sResourceOperationLogDao 用户 K8s 资源操作日志 DAO
type UserK8sResourceOperationLogDao struct {
	db *gorm.DB
}

// NewUserK8sResourceOperationLogDao 创建用户 K8s 资源操作日志 DAO
func NewUserK8sResourceOperationLogDao(db *gorm.DB) *UserK8sResourceOperationLogDao {
	return &UserK8sResourceOperationLogDao{db: db}
}

// Create 创建用户 K8s 资源操作日志
func (d *UserK8sResourceOperationLogDao) Create(log *UserK8sResourceOperationLog) error {
	return d.db.Create(log).Error
}

// QueryByK8sResourceID 根据 K8s 资源 ID 查询操作日志
func (d *UserK8sResourceOperationLogDao) QueryByK8sResourceID(k8sResourceID uint) ([]*UserK8sResourceOperationLog, error) {
	var logs []*UserK8sResourceOperationLog
	err := d.db.Where("k8s_resource_id = ?", k8sResourceID).Order("id desc").Find(&logs).Error
	return logs, err
}

// QueryByUserID 根据用户 ID 查询操作日志
func (d *UserK8sResourceOperationLogDao) QueryByUserID(userID uint) ([]*UserK8sResourceOperationLog, error) {
	var logs []*UserK8sResourceOperationLog
	err := d.db.Where("user_id = ?", userID).Find(&logs).Error
	return logs, err
}
