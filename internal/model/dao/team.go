package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// Team 定义团队结构体
type Team struct {
	ID              uint32     `gorm:"column:id;type:int UNSIGNED;primaryKey;not null;" json:"id"`
	TeamName        string     `gorm:"column:team_name;type:varchar(255);not null;" json:"team_name"`
	TeamDescription string     `gorm:"column:team_description;type:text;" json:"team_description"`
	TeamUUID        uint32     `gorm:"column:team_uuid;type:int UNSIGNED;not null;uniqueIndex;" json:"team_uuid"`
	CreatorID       uint32     `gorm:"column:creator_id;type:int UNSIGNED;not null;" json:"creator_id"`
	CreatedAt       *time.Time `gorm:"column:created_at;type:datetime;not null;" json:"created_at"`
	UpdatedAt       *time.Time `gorm:"column:updated_at;type:datetime;not null;" json:"updated_at"`
	DeletedAt       *time.Time `gorm:"column:deleted_at;type:datetime;default:NULL;" json:"deleted_at"`
}

// TableName 指定表名
func (Team) TableName() string {
	return "team"
}

// TeamDao 团队数据访问对象
type TeamDao struct {
	db *gorm.DB
}

// NewTeamDao 创建 TeamDao 实例
func NewTeamDao(db *gorm.DB) *TeamDao {
	return &TeamDao{db: db}
}

// Create 创建团队
func (d *TeamDao) Create(ctx context.Context, team *Team) error {
	return d.db.WithContext(ctx).Create(team).Error
}

// CreateTx 创建团队
func (d *TeamDao) CreateTx(tx *gorm.DB, ctx context.Context, team *Team) error {
	return tx.WithContext(ctx).Create(team).Error
}

// Update 更新团队
func (d *TeamDao) Update(ctx context.Context, team *Team) error {
	return d.db.WithContext(ctx).Save(team).Error
}

// Delete 删除团队（软删除）
func (d *TeamDao) Delete(ctx context.Context, teamID uint32) error {
	return d.db.WithContext(ctx).Model(&Team{}).Where("id = ?", teamID).Update("deleted_at", time.Now()).Error
}

// DeleteTx 删除团队（软删除）
func (d *TeamDao) DeleteTx(tx *gorm.DB, ctx context.Context, teamID uint32) error {
	return tx.WithContext(ctx).Model(&Team{}).Where("id = ?", teamID).Update("deleted_at", time.Now()).Error
}

// GetByID 根据ID获取团队
func (d *TeamDao) GetByID(ctx context.Context, teamID uint32) (*Team, error) {
	var team Team
	err := d.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", teamID).First(&team).Error
	if err != nil {
		return nil, err
	}
	return &team, nil
}

// GetByUUID 根据UUID获取团队
func (d *TeamDao) GetByUUID(ctx context.Context, teamUUID string) (*Team, error) {
	var team Team
	err := d.db.WithContext(ctx).Where("team_uuid = ? AND deleted_at IS NULL", teamUUID).First(&team).Error
	if err != nil {
		return nil, err
	}
	return &team, nil
}

// GetByCreatorID 根据创建者ID获取团队列表
func (d *TeamDao) GetByCreatorID(ctx context.Context, creatorID uint32) ([]*Team, error) {
	var teams []*Team
	err := d.db.WithContext(ctx).Where("creator_id = ? AND deleted_at IS NULL", creatorID).Find(&teams).Error
	if err != nil {
		return nil, err
	}
	return teams, nil
}

// Query 查询团队列表
func (d *TeamDao) Query(ctx context.Context, teamName string, teamUUID int) ([]*Team, error) {
	var teams []*Team
	query := d.db.WithContext(ctx).Where("deleted_at IS NULL")

	if teamName != "" {
		query = query.Where("team_name LIKE ?", "%"+teamName+"%")
	}

	if teamUUID != 0 {
		query = query.Where("team_uuid = ?", teamUUID)
	}

	err := query.Find(&teams).Error
	if err != nil {
		return nil, err
	}

	return teams, nil
}

// UserTeamDao 用户团队数据访问对象
type UserTeamDao struct {
	db *gorm.DB
}

// NewUserTeamDao 创建 UserTeamDao 实例
func NewUserTeamDao(db *gorm.DB) *UserTeamDao {
	return &UserTeamDao{db: db}
}

// UpdateUserTeamID 更新用户团队ID
func (d *UserTeamDao) UpdateUserTeamID(ctx context.Context, userID uint32, teamID *uint32) error {
	return d.db.WithContext(ctx).Model(&Users{}).Where("id = ?", userID).Update("team_id", teamID).Error
}

// GetUsersByTeamID 根据团队ID获取用户列表
func (d *UserTeamDao) GetUsersByTeamID(ctx context.Context, teamID uint32) ([]*Users, error) {
	var users []*Users
	err := d.db.WithContext(ctx).Where("team_id = ? AND deleted_at IS NULL", teamID).Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}
