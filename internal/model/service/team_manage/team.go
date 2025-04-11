package team_manage

import (
	"context"
	"errors"
	"time"

	"github.com/ZZGADA/easy-deploy/internal/model/conf"

	"github.com/ZZGADA/easy-deploy/internal/model/dao"
	"github.com/ZZGADA/easy-deploy/internal/utils"
)

// TeamService 团队服务
type TeamService struct {
	teamDao *dao.TeamDao
	userDao *dao.UsersDao
}

// NewTeamService 创建团队服务
func NewTeamService(teamDao *dao.TeamDao, userDao *dao.UsersDao) *TeamService {
	return &TeamService{
		teamDao: teamDao,
		userDao: userDao,
	}
}

// CreateTeam 创建团队
func (s *TeamService) CreateTeam(ctx context.Context, teamName, teamDescription string, userId uint) (*dao.Team, error) {
	creatorId := uint32(userId)
	// 检查用户是否存在
	user, err := s.userDao.GetUserByID(creatorId)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	// 生成团队UUID
	teamUUID := utils.GenerateUUID()

	tx := conf.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	// 创建团队
	team := &dao.Team{
		TeamName:        teamName,
		TeamDescription: teamDescription,
		TeamUUID:        teamUUID,
		CreatorID:       creatorId,
	}

	err = s.teamDao.CreateTx(tx, ctx, team)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// 更新用户的team_id
	user.TeamID = team.ID
	err = s.userDao.UpdateUserTx(tx, user)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()

	return team, nil
}

// UpdateTeam 更新团队信息
func (s *TeamService) UpdateTeam(ctx context.Context, teamID uint32, teamName, teamDescription string) error {
	team, err := s.teamDao.GetByID(ctx, teamID)
	if err != nil {
		return errors.New("团队不存在")
	}

	team.TeamName = teamName
	team.TeamDescription = teamDescription
	team.UpdatedAt = &time.Time{}

	return s.teamDao.Update(ctx, team)
}

// DeleteTeam 删除团队
func (s *TeamService) DeleteTeam(ctx context.Context, teamID uint32) error {
	team, err := s.teamDao.GetByID(ctx, teamID)
	if err != nil {
		return errors.New("团队不存在")
	}

	tx := conf.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	// 批量清除所有成员的team_id
	err = s.userDao.BatchUpdateTeamIDTx(tx, ctx, team.ID, nil)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = s.teamDao.DeleteTx(tx, ctx, teamID)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

// GetTeamByID 根据ID获取团队
func (s *TeamService) GetTeamByID(ctx context.Context, teamID uint32) (*dao.Team, error) {
	return s.teamDao.GetByID(ctx, teamID)
}

// QueryTeams 查询团队列表
func (s *TeamService) QueryTeams(ctx context.Context, page, pageSize int, teamName string, teamUuid int) ([]*dao.Team, int64, error) {
	// 这里需要实现分页查询
	// 由于TeamDao接口中没有定义分页查询方法，这里暂时返回所有团队
	teams, err := s.teamDao.Query(ctx, teamName, teamUuid)
	if err != nil {
		return nil, 0, err
	}

	// 计算总数
	total := int64(len(teams))

	// 计算分页
	start := (page - 1) * pageSize
	end := start + pageSize
	if start >= len(teams) {
		return []*dao.Team{}, total, nil
	}
	if end > len(teams) {
		end = len(teams)
	}

	return teams[start:end], total, nil
}

// GetUserTeam 获取用户自己的团队信息
func (s *TeamService) GetUserTeam(ctx context.Context, userID uint32) (*dao.Team, error) {
	// 获取用户信息
	user, err := s.userDao.GetUserByID(userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	// 检查用户是否属于团队
	if user.TeamID == 0 {
		return nil, errors.New("用户未加入任何团队")
	}

	// 获取团队信息
	team, err := s.teamDao.GetByID(ctx, user.TeamID)
	if err != nil {
		return nil, errors.New("团队不存在")
	}

	return team, nil
}

func (s *TeamService) GetTeamMemberByID(ctx context.Context, teamId uint32) ([]*dao.UsersGithub, error) {
	users, err := s.userDao.GetUsersByTeamIDGithub(teamId)
	if err != nil {
		return nil, err
	}

	return users, nil
}
