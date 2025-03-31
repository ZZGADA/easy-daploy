package user_manage

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/ZZGADA/easy-deploy/internal/model/conf"
	"github.com/ZZGADA/easy-deploy/internal/model/define"
	"github.com/ZZGADA/easy-deploy/internal/utils"
	"github.com/go-redis/redis/v8"

	"github.com/ZZGADA/easy-deploy/internal/model/dao"
)

var dockerLoginLock sync.Mutex

// DockerAccountService Docker账号管理服务
type DockerAccountService struct {
	dockerDao dao.UserDockerDao
}

// NewDockerAccountService 创建 DockerAccountService 实例
func NewDockerAccountService(dockerDao dao.UserDockerDao) *DockerAccountService {
	return &DockerAccountService{
		dockerDao: dockerDao,
	}
}

// SaveDockerAccount 保存Docker账号信息
func (s *DockerAccountService) SaveDockerAccount(server, username, password, comment string, userId uint) (bool, error) {
	if server == "" || username == "" || password == "" {
		return false, errors.New("必填参数不能为空")
	}

	docker := &dao.UserDocker{
		UserID:   userId,
		Server:   server,
		Username: username,
		Password: password,
		Comment:  comment,
	}

	// 如果是用户的第一个账号，设置为默认账号
	count, err := s.dockerDao.CountByUserID(userId)
	if err != nil {
		return false, err
	}
	if count == 0 {
		docker.IsDefault = true
	}

	if err := s.dockerDao.Create(docker); err != nil {
		return false, err
	}

	return true, nil
}

// UpdateDockerAccount 更新Docker账号信息
func (s *DockerAccountService) UpdateDockerAccount(id uint, server, username, password, comment string, userId uint) (bool, error) {
	if id == 0 || server == "" || username == "" || password == "" {
		return false, errors.New("必填参数不能为空")
	}

	// 检查账号是否存在且属于当前用户
	docker, err := s.dockerDao.GetByID(id)
	if err != nil {
		return false, err
	}
	if docker.UserID != userId {
		return false, errors.New("无权操作此账号")
	}

	docker.Server = server
	docker.Username = username
	docker.Password = password
	docker.Comment = comment

	if err := s.dockerDao.Update(docker); err != nil {
		return false, err
	}

	return true, nil
}

// DeleteDockerAccount 删除Docker账号
func (s *DockerAccountService) DeleteDockerAccount(id uint, userId uint) (bool, error) {
	if id == 0 {
		return false, errors.New("ID不能为空")
	}

	// 检查账号是否存在且属于当前用户
	docker, err := s.dockerDao.GetByID(id)
	if err != nil {
		return false, err
	}
	if docker.UserID != userId {
		return false, errors.New("无权操作此账号")
	}

	if err := s.dockerDao.Delete(id); err != nil {
		return false, err
	}

	return true, nil
}

// QueryDockerAccounts 查询用户的Docker账号列表
func (s *DockerAccountService) QueryDockerAccounts(userId uint) ([]*dao.UserDocker, error) {
	dockers, err := s.dockerDao.GetByUserID(userId)
	if err != nil {
		return nil, err
	}

	return dockers, nil
}

// SetDefaultAccount 设置默认Docker账号
func (s *DockerAccountService) SetDefaultAccount(dockerID uint, userId uint) (bool, error) {
	if dockerID == 0 {
		return false, errors.New("账号ID不能为空")
	}

	// 检查账号是否存在且属于当前用户
	docker, err := s.dockerDao.GetByID(dockerID)
	if err != nil {
		return false, err
	}
	if docker.UserID != userId {
		return false, errors.New("无权操作此账号")
	}

	if err := s.dockerDao.SetDefault(userId, dockerID); err != nil {
		return false, err
	}

	return true, nil
}

// LoginDockerAccount 登录 Docker 账号
func (s *DockerAccountService) LoginDockerAccount(dockerID uint, userId uint) (bool, error) {
	// 获取分布式锁
	lockKey := fmt.Sprintf(define.UserDockerLogin, userId)
	lockValue := utils.GenerateUniqueValue()
	ctx := context.Background()
	lock, err2 := utils.AcquireLock(conf.RedisClient, ctx, lockKey, lockValue, time.Minute*3)
	if err2 != nil && err2 != redis.Nil {
		return false, err2
	}

	if !lock {
		return false, nil
	}

	defer func() {
		_, err := utils.ReleaseLock(conf.RedisClient, ctx, lockKey, lockValue)
		if err != nil {
			return
		}
	}()

	// 开启事务
	tx := s.dockerDao.BeginTx()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 检查账号是否存在且属于当前用户
	docker, err := s.dockerDao.GetByID(dockerID)
	if err != nil {
		return false, err
	}
	if docker.UserID != userId {
		return false, errors.New("无权操作此账号")
	}

	// 检查 Docker 环境
	if err := checkDockerEnvironment(); err != nil {
		return false, err
	}

	// 检查是否已经登录
	if docker.IsLogin {
		return true, nil
	}

	// 获取当前登录的账号
	currentLoginAccount, err := s.dockerDao.GetLoginAccount(userId)
	if err != nil {
		return false, err
	}

	// 如果有其他账号已登录，先退出
	if currentLoginAccount != nil && currentLoginAccount.ID != dockerID {
		if err := dockerLogout(currentLoginAccount.Server); err != nil {
			return false, err
		}
		if err := s.dockerDao.UpdateLoginStatus(currentLoginAccount.ID, false); err != nil {
			tx.Rollback()
			return false, err
		}
	}

	// 执行 Docker 登录
	if err := dockerLogin(docker.Server, docker.Username, docker.Password); err != nil {
		tx.Rollback()
		return false, err
	}

	// 更新登录状态
	if err := s.dockerDao.UpdateLoginStatus(dockerID, true); err != nil {
		tx.Rollback()
		return false, err
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return false, err
	}

	return true, nil
}

// checkDockerEnvironment 检查 Docker 环境
func checkDockerEnvironment() error {
	cmd := exec.Command("docker", "version")
	if err := cmd.Run(); err != nil {
		return errors.New("Docker 环境未正确配置")
	}
	return nil
}

// dockerLogin 执行 Docker 登录
func dockerLogin(server, username, password string) error {
	cmd := exec.Command("docker", "login", server, "-u", username, "--password-stdin")
	cmd.Stdin = strings.NewReader(password)
	if err := cmd.Run(); err != nil {
		return errors.New("Docker 登录失败")
	}
	return nil
}

// dockerLogout 执行 Docker 退出
func dockerLogout(server string) error {
	cmd := exec.Command("docker", "logout", server)
	if err := cmd.Run(); err != nil {
		return errors.New("Docker 退出失败")
	}
	return nil
}

// GetLoginAccount 获取当前登录的 Docker 账号
func (s *DockerAccountService) GetLoginAccount(userId uint) (*dao.UserDocker, error) {
	if userId == 0 {
		return nil, errors.New("用户ID不能为空")
	}

	account, err := s.dockerDao.GetLoginAccount(userId)
	if err != nil {
		return nil, err
	}

	return account, nil
}
