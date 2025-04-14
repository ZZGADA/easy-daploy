package team_manage

import (
	"context"
	"errors"
	"fmt"
	"github.com/ZZGADA/easy-deploy/internal/utils"
	"github.com/sirupsen/logrus"
	"time"

	"github.com/ZZGADA/easy-deploy/internal/define"
	"github.com/ZZGADA/easy-deploy/internal/model/conf"

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
	team, err := s.teamDao.GetByID(ctx, teamID)
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
		for _, request := range requests {
			if request.UserID == uint32(userID) {
				msg := "等待中"
				if request.RequestType == define.TeamRequestTypeIn {
					msg = "申请加入"
				} else {
					msg = "申请退出"
				}
				return nil, errors.New(fmt.Sprintf("您已经有申请的请求，请勿重复提交。请求：%s", msg))
			}
		}
	}

	// 创建申请
	request := &dao.TeamRequest{
		TeamID:      teamID,
		UserID:      user.Id,
		RequestType: requestType,
		Status:      define.TeamRequestStatusWait,
	}

	err = s.teamRequestDao.Create(ctx, request)
	if err != nil {
		return nil, err
	}

	userRequest, err := s.teamRequestDao.GetByUserIDWaitingJoin(ctx, uint32(userID), requestType)
	if err != nil {
		return nil, err
	}
	if len(userRequest) != 1 {
		return nil, errors.New("已经审批通过无需重复申请")
	}

	// 获取申请者的完整信息
	applicant, err := s.userDao.GetUserWithGithubInfo(ctx, userID)
	creatorInfo, err := s.userDao.GetUserWithGithubInfo(ctx, uint(team.CreatorID))
	if err != nil {
		logrus.Errorf("获取申请者信息失败: %v", err)
	} else {
		// 发送邮件
		if err := utils.SendJoinTeamEmail(applicant, creatorInfo.Email, userRequest[0].ID, requestType); err != nil {
			logrus.Errorf("发送加入团队申请邮件失败: %v", err)
		}
	}

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

	user, err := s.userDao.GetUserByID(request.UserID)
	if err != nil {
		return err
	}

	// 如果申请被批准，更新用户的team_id
	// 如果拒绝就保持原样
	if status == define.TeamRequestStatusApproval {
		if request.RequestType == define.TeamRequestTypeOut {
			user.TeamID = 0
		} else {
			user.TeamID = request.TeamID
		}
	}

	err = s.userDao.UpdateUserTx(tx, user)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = s.teamRequestDao.UpdateTx(tx, ctx, request)
	if err != nil {
		tx.Rollback()
	}
	tx.Commit()

	team, err := s.teamDao.GetByID(ctx, request.TeamID)

	if err = utils.SendTeamRequestEmail(team, user.Email, request.RequestType, status); err != nil {
		return err
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
func (s *TeamRequestService) GetTeamRequestsByTeamID(ctx context.Context, teamID uint32) ([]map[string]interface{}, error) {
	teamRequest, err := s.teamRequestDao.GetByTeamID(ctx, teamID)
	if err != nil {
		return nil, err
	}

	var usersIds []uint32
	userIdMapTeam := make(map[uint32]int)
	userIdMapRequest := make(map[uint32]uint32)
	for _, request := range teamRequest {
		usersIds = append(usersIds, request.UserID)
		userIdMapTeam[request.UserID] = request.RequestType
		userIdMapRequest[request.UserID] = request.ID
	}

	infos, err := s.userDao.GetUserListWithGithubInfo(ctx, usersIds)
	if err != nil {
		return nil, err
	}

	var result []map[string]interface{}
	for _, info := range infos {
		result = append(result, map[string]interface{}{
			"user_id":      info.ID,
			"request_type": userIdMapTeam[uint32(info.ID)],
			"github_name":  info.Name,
			"github_id":    info.GithubID,
			"email":        info.Email,
			"request_id":   userIdMapRequest[uint32(info.ID)],
		})
	}

	return result, nil
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
