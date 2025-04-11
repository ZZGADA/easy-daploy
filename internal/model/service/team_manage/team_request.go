package team_manage

import (
	"context"
	"errors"
	"fmt"
	"github.com/ZZGADA/easy-deploy/internal/define"
	"github.com/ZZGADA/easy-deploy/internal/model/conf"
	"time"

	"github.com/ZZGADA/easy-deploy/internal/model/dao"
)

// TeamRequestService 团队申请服务
type TeamRequestService struct {
	teamRequestDao *dao.TeamRequestDao
	teamDao        *dao.TeamDao
	userDao        *dao.UsersDao
}

// NewTeamRequestService 创建团队申请服务
func NewTeamRequestService(teamRequestDao *dao.TeamRequestDao, teamDao *dao.TeamDao, userDao *dao.UsersDao) *TeamRequestService {
	return &TeamRequestService{
		teamRequestDao: teamRequestDao,
		teamDao:        teamDao,
		userDao:        userDao,
	}
}

// CreateTeamRequest 创建团队申请
func (s *TeamRequestService) CreateTeamRequest(ctx context.Context, teamID uint32, userID uint, requestType int) (*dao.TeamRequest, error) {
	// 检查团队是否存在
	_, err := s.teamDao.GetByID(ctx, teamID)
	if err != nil {
		return nil, errors.New("团队不存在")
	}

	// 检查用户是否存在
	user, err := s.userDao.GetUserByID(uint32(userID))
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	// 检查是否已经存在相同的申请
	requests, err := s.teamRequestDao.GetByTeamID(ctx, teamID)
	if err != nil {
		return nil, err
	}

	if len(requests) != 0 {
		request := requests[0]
		msg := "等待中"
		if request.RequestType == define.TeamRequestTypeIn {
			msg = "申请加入"
		} else {
			msg = "申请退出"
		}
		return nil, errors.New(fmt.Sprintf("您已经有申请的请求，请勿重复提交。请求：%s", msg))
	}

	// 创建申请
	request := &dao.TeamRequest{
		TeamID:      teamID,
		UserID:      user.Id,
		RequestType: requestType,
		Status:      define.TeamRequestStatusWait,
		CreatedAt:   &time.Time{},
		UpdatedAt:   &time.Time{},
	}

	err = s.teamRequestDao.Create(ctx, request)
	if err != nil {
		return nil, err
	}

	//TODO: 发送邮件

	return request, nil
}

// CheckTeamRequest 更新团队申请状态
func (s *TeamRequestService) CheckTeamRequest(ctx context.Context, requestID uint32, status int) error {
	request, err := s.teamRequestDao.GetByID(ctx, requestID)
	if err != nil {
		return errors.New("申请不存在")
	}

	request.Status = status
	request.UpdatedAt = &time.Time{}

	tx := conf.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 如果申请被批准，更新用户的team_id
	if status == define.TeamRequestStatusApproval {
		user, err := s.userDao.GetUserByID(request.UserID)
		if err != nil {
			return err
		}
		user.TeamID = request.TeamID
		err = s.userDao.UpdateUserTx(tx, user)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err = s.teamRequestDao.UpdateTx(tx, ctx, request)
	if err != nil {
		tx.Rollback()
	}

	return nil
}

// DeleteTeamRequest 删除团队申请
func (s *TeamRequestService) DeleteTeamRequest(ctx context.Context, requestID uint32) error {
	return s.teamRequestDao.Delete(ctx, requestID)
}

// GetTeamRequestByID 根据ID获取团队申请
func (s *TeamRequestService) GetTeamRequestByID(ctx context.Context, requestID uint32) (*dao.TeamRequest, error) {
	return s.teamRequestDao.GetByID(ctx, requestID)
}

// GetTeamRequestsByTeamID 获取团队的申请列表
func (s *TeamRequestService) GetTeamRequestsByTeamID(ctx context.Context, teamID uint32) ([]*dao.TeamRequest, error) {
	return s.teamRequestDao.GetByTeamID(ctx, teamID)
}

// GetTeamRequestsByUserID 获取用户的申请列表
func (s *TeamRequestService) GetTeamRequestsByUserID(ctx context.Context, userID uint32) ([]*dao.TeamRequest, error) {
	return s.teamRequestDao.GetByUserID(ctx, userID)
}

// GetPendingTeamRequestsByTeamID 获取团队待处理的申请列表
func (s *TeamRequestService) GetPendingTeamRequestsByTeamID(ctx context.Context, teamID uint32) ([]*dao.TeamRequest, error) {
	return s.teamRequestDao.GetPendingByTeamID(ctx, teamID)
}

// QueryTeamRequests 查询团队申请列表
func (s *TeamRequestService) QueryTeamRequests(ctx context.Context, page, pageSize int) ([]*dao.TeamRequest, int64, error) {
	// 这里需要实现分页查询
	// 由于TeamRequestDao接口中没有定义分页查询方法，这里暂时返回所有申请
	requests, err := s.teamRequestDao.Query(ctx, "", "")
	if err != nil {
		return nil, 0, err
	}

	// 计算总数
	total := int64(len(requests))

	// 计算分页
	start := (page - 1) * pageSize
	end := start + pageSize
	if start >= len(requests) {
		return []*dao.TeamRequest{}, total, nil
	}
	if end > len(requests) {
		end = len(requests)
	}

	return requests[start:end], total, nil
}
