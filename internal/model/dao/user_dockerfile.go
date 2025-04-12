package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// DockerfileItem 表示 Dockerfile 中的一个命令项
type DockerfileItem struct {
	Index         int    `json:"index"`          // 执行顺序
	DockerfileKey string `json:"dockerfile_key"` // Dockerfile 关键字
	ShellValue    string `json:"shell_value"`    // 具体的 shell 命令
}

// UserDockerfile 用户 Dockerfile 信息表
type UserDockerfile struct {
	Id             uint32     `gorm:"column:id;type:int UNSIGNED;primaryKey;not null;" json:"id"`
	UserId         uint32     `gorm:"column:user_id;type:int UNSIGNED;not null;" json:"user_id"`
	RepositoryName string     `gorm:"column:repository_name;type:varchar(255);not null;" json:"repository_name"`
	RepositoryId   string     `gorm:"column:repository_id;type:varchar(255);not null;" json:"repository_id"`
	BranchName     string     `gorm:"column:branch_name;type:varchar(255);not null;" json:"branch_name"`
	FileName       string     `gorm:"column:file_name;type:varchar(255);not null;" json:"file_name"`
	FileData       string     `gorm:"column:file_data;type:text;not null;" json:"file_data"` // JSON 格式存储
	CreatedAt      *time.Time `gorm:"column:created_at;type:datetime;not null;" json:"created_at"`
	UpdatedAt      *time.Time `gorm:"column:updated_at;type:datetime;not null;" json:"updated_at"`
	DeletedAt      *time.Time `gorm:"column:deleted_at;type:datetime;" json:"deleted_at"`
}

// TableName 指定表名
func (UserDockerfile) TableName() string {
	return "user_dockerfile"
}

// UserDockerfileDao Dockerfile 数据访问对象
type UserDockerfileDao struct {
	db *gorm.DB
}

// NewUserDockerfileDao 创建 UserDockerfileDao 实例
func NewUserDockerfileDao(db *gorm.DB) *UserDockerfileDao {
	return &UserDockerfileDao{db: db}
}

// Create 创建 Dockerfile 记录
func (d *UserDockerfileDao) Create(ctx context.Context, dockerfile *UserDockerfile) error {
	return d.db.WithContext(ctx).Create(dockerfile).Error
}

// Update 更新 Dockerfile 记录
func (d *UserDockerfileDao) Update(ctx context.Context, dockerfile *UserDockerfile) error {
	return d.db.WithContext(ctx).Save(dockerfile).Error
}

// Delete 删除 Dockerfile 记录（软删除）
func (d *UserDockerfileDao) Delete(ctx context.Context, id uint32) error {
	return d.db.WithContext(ctx).
		Model(&UserDockerfile{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Update("deleted_at", time.Now()).
		Error
}

// GetByUserIDAndRepo 根据用户 ID 和仓库id获取 Dockerfile
func (d *UserDockerfileDao) GetByUserIDAndRepo(ctx context.Context, userId uint32, repositoryId string) (*UserDockerfile, error) {
	var dockerfile UserDockerfile
	err := d.db.WithContext(ctx).
		Where("user_id = ? AND repository_id = ? AND deleted_at IS NULL", userId, repositoryId).
		First(&dockerfile).Error
	if err != nil {
		return nil, err
	}
	return &dockerfile, nil
}

// GetByID 根据fileId获取 Dockerfile
func (d *UserDockerfileDao) GetByID(ctx context.Context, fileId uint32) (*UserDockerfile, error) {
	var dockerfile UserDockerfile
	err := d.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", fileId).
		First(&dockerfile).Error
	if err != nil {
		return nil, err
	}
	return &dockerfile, nil
}

func (d *UserDockerfileDao) GetByRepoIDAndBranch(ctx context.Context, repositoryId string, branchName string) ([]*UserDockerfile, error) {
	var dockerfiles []*UserDockerfile
	result := d.db.WithContext(ctx).Where(
		"repository_id = ? AND branch_name = ? and deleted_at IS NULL",
		repositoryId, branchName,
	).Find(&dockerfiles)

	if result.Error != nil {
		return nil, result.Error
	}

	return dockerfiles, nil
}
