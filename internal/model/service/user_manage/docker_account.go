package user_manage

import (
	"errors"

	"github.com/ZZGADA/easy-deploy/internal/model/dao"
)

// DockerAccountService Docker账号管理服务
type DockerAccountService struct {
	UserID    uint
	dockerDao dao.UserDockerDao
}

// NewDockerAccountService 创建 DockerAccountService 实例
func NewDockerAccountService(dockerDao dao.UserDockerDao) *DockerAccountService {
	return &DockerAccountService{
		dockerDao: dockerDao,
	}
}

// SaveDockerAccount 保存Docker账号信息
func (s *DockerAccountService) SaveDockerAccount(server, username, password, comment string) (bool, error) {
	if server == "" || username == "" || password == "" {
		return false, errors.New("必填参数不能为空")
	}

	docker := &dao.UserDocker{
		UserID:   s.UserID,
		Server:   server,
		Username: username,
		Password: password,
		Comment:  comment,
	}

	// 如果是用户的第一个账号，设置为默认账号
	count, err := s.dockerDao.CountByUserID(s.UserID)
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
func (s *DockerAccountService) UpdateDockerAccount(id uint, server, username, password, comment string) (bool, error) {
	if id == 0 || server == "" || username == "" || password == "" {
		return false, errors.New("必填参数不能为空")
	}

	// 检查账号是否存在且属于当前用户
	docker, err := s.dockerDao.GetByID(id)
	if err != nil {
		return false, err
	}
	if docker.UserID != s.UserID {
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
func (s *DockerAccountService) DeleteDockerAccount(id uint) (bool, error) {
	if id == 0 {
		return false, errors.New("ID不能为空")
	}

	// 检查账号是否存在且属于当前用户
	docker, err := s.dockerDao.GetByID(id)
	if err != nil {
		return false, err
	}
	if docker.UserID != s.UserID {
		return false, errors.New("无权操作此账号")
	}

	if err := s.dockerDao.Delete(id); err != nil {
		return false, err
	}

	return true, nil
}

// QueryDockerAccounts 查询用户的Docker账号列表
func (s *DockerAccountService) QueryDockerAccounts() ([]*dao.UserDocker, error) {
	dockers, err := s.dockerDao.GetByUserID(s.UserID)
	if err != nil {
		return nil, err
	}

	return dockers, nil
}

// SetDefaultAccount 设置默认Docker账号
func (s *DockerAccountService) SetDefaultAccount(dockerID uint) (bool, error) {
	if dockerID == 0 {
		return false, errors.New("账号ID不能为空")
	}

	// 检查账号是否存在且属于当前用户
	docker, err := s.dockerDao.GetByID(dockerID)
	if err != nil {
		return false, err
	}
	if docker.UserID != s.UserID {
		return false, errors.New("无权操作此账号")
	}

	if err := s.dockerDao.SetDefault(s.UserID, dockerID); err != nil {
		return false, err
	}

	return true, nil
}
