package dao

import (
	"context"
	"github.com/ZZGADA/easy-deploy/internal/define"
	"time"

	"gorm.io/gorm"
)

// TeamRequest 定义团队申请结构体
type TeamRequest struct {
	ID          uint32     `gorm:"column:id;type:int UNSIGNED;primaryKey;not null;" json:"id"`
	TeamID      uint32     `gorm:"column:team_id;type:int UNSIGNED;not null;" json:"team_id"`
	UserID      uint32     `gorm:"column:user_id;type:int UNSIGNED;not null;" json:"user_id"`
	RequestType int        `gorm:"column:request_type;type:int;not null;default:0;" json:"request_type"` // 0: 加入团队, 1: 退出团队
	Status      int        `gorm:"column:status;type:int;not null;default:0;" json:"status"`             // 0: 待处理, 1: 已同意, 2: 已拒绝
	CreatedAt   *time.Time `gorm:"column:created_at;type:datetime;not null;" json:"created_at"`
	UpdatedAt   *time.Time `gorm:"column:updated_at;type:datetime;not null;" json:"updated_at"`
	DeletedAt   *time.Time `gorm:"column:deleted_at;type:datetime;default:NULL;" json:"deleted_at"`
}

// TableName 指定表名
func (TeamRequest) TableName() string {
	return "team_request"
}

// TeamRequestDao 团队申请数据访问对象
type TeamRequestDao struct {
	db *gorm.DB
}

// NewTeamRequestDao 创建 TeamRequestDao 实例
func NewTeamRequestDao(db *gorm.DB) *TeamRequestDao {
	return &TeamRequestDao{db: db}
}

// Create 创建团队申请
func (d *TeamRequestDao) Create(ctx context.Context, request *TeamRequest) error {
	return d.db.WithContext(ctx).Create(request).Error
}

// Update 更新团队申请
func (d *TeamRequestDao) Update(ctx context.Context, request *TeamRequest) error {
	return d.db.WithContext(ctx).Save(request).Error
}

// UpdateTx 更新团队申请
func (d *TeamRequestDao) UpdateTx(tx *gorm.DB, ctx context.Context, request *TeamRequest) error {
	return d.db.WithContext(ctx).Save(request).Error
}

// Delete 删除团队申请（软删除）
func (d *TeamRequestDao) Delete(ctx context.Context, requestID uint32) error {
	return d.db.WithContext(ctx).Model(&TeamRequest{}).Where("id = ?", requestID).Update("deleted_at", time.Now()).Error
}

// GetByID 根据ID获取团队申请
func (d *TeamRequestDao) GetByID(ctx context.Context, requestID uint32) (*TeamRequest, error) {
	var request TeamRequest
	err := d.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", requestID).First(&request).Error
	if err != nil {
		return nil, err
	}
	return &request, nil
}

// GetByTeamID 根据团队ID获取申请列表
func (d *TeamRequestDao) GetByTeamID(ctx context.Context, teamID uint32) ([]*TeamRequest, error) {
	var requests []*TeamRequest
	err := d.db.WithContext(ctx).Where("team_id = ? AND status = ? AND deleted_at IS NULL", teamID, define.TeamRequestStatusWait).Order("id desc").Limit(1).Find(&requests).Error
	if err != nil {
		return nil, err
	}
	return requests, nil
}

// GetByUserID 根据用户ID获取申请列表
func (d *TeamRequestDao) GetByUserID(ctx context.Context, userID uint32) ([]*TeamRequest, error) {
	var requests []*TeamRequest
	err := d.db.WithContext(ctx).Where("user_id = ? AND deleted_at IS NULL", userID).Find(&requests).Error
	if err != nil {
		return nil, err
	}
	return requests, nil
}

// GetPendingByTeamID 获取团队待处理的申请列表
func (d *TeamRequestDao) GetPendingByTeamID(ctx context.Context, teamID uint32) ([]*TeamRequest, error) {
	var requests []*TeamRequest
	err := d.db.WithContext(ctx).Where("team_id = ? AND status = 0 AND deleted_at IS NULL", teamID).Find(&requests).Error
	if err != nil {
		return nil, err
	}
	return requests, nil
}

// Query 查询团队申请列表
func (d *TeamRequestDao) Query(ctx context.Context, teamID string, userID string) ([]*TeamRequest, error) {
	var requests []*TeamRequest
	query := d.db.WithContext(ctx).Where("deleted_at IS NULL")

	if teamID != "" {
		query = query.Where("team_id = ?", teamID)
	}

	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	err := query.Find(&requests).Error
	if err != nil {
		return nil, err
	}

	return requests, nil
}
