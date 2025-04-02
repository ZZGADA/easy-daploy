package docker_manage

import (
	"context"
	"errors"

	log "github.com/sirupsen/logrus"

	"github.com/ZZGADA/easy-deploy/internal/model/dao"
)

type DockerImageService struct {
	userDockerImageDao *dao.UserDockerImageDao
}

func NewDockerImageService(userDockerImageDao *dao.UserDockerImageDao) *DockerImageService {
	return &DockerImageService{
		userDockerImageDao: userDockerImageDao,
	}
}

// SaveDockerImage 保存 Docker 镜像信息
func (s *DockerImageService) SaveDockerImage(ctx context.Context, userID uint32, dockerfileID uint32, fullImageName, imageName string) error {
	image := &dao.UserDockerImage{
		UserId:        userID,
		DockerfileId:  dockerfileID,
		FullImageName: fullImageName,
		ImageName:     imageName,
	}
	return s.userDockerImageDao.Create(ctx, image)
}

// SaveDockerImageWS 保存 Docker 镜像信息-WS
func (s *DockerImageService) SaveDockerImageWS(data map[string]interface{}, userID uint, fullImageName string) error {
	log.Info("=== HandleBuildImage 开始 ===")
	log.Infof("接收到的数据: %+v", data)

	ctx := context.Background()
	// 验证必要参数
	dockerfileID, ok := data["id"].(float64)
	if !ok {
		log.Error("缺少 Dockerfile ID")
		return errors.New("缺少 Dockerfile ID")
	}

	imageName, ok := data["docker_image_name"].(string)
	if !ok {
		log.Error("缺少镜像名称")
		return errors.New("缺少镜像名称")
	}
	log.Infof("处理参数 - DockerfileID: %v, ImageName: %s", dockerfileID, imageName)

	image := &dao.UserDockerImage{
		UserId:        uint32(userID),
		DockerfileId:  uint32(dockerfileID),
		FullImageName: fullImageName,
		ImageName:     imageName,
	}
	return s.userDockerImageDao.Create(ctx, image)
}

// GetDockerImagesByDockerfileID 根据 DockerfileID 查询 Docker 镜像列表
func (s *DockerImageService) GetDockerImagesByDockerfileID(ctx context.Context, dockerfileID uint32) ([]*dao.UserDockerImage, error) {
	return s.userDockerImageDao.GetByDockerfileID(ctx, dockerfileID)
}

// GetDockerImagesByRepositoryID 根据仓库ID获取镜像列表
func (s *DockerImageService) GetDockerImagesByRepositoryID(ctx context.Context, repositoryID string) ([]*dao.UserDockerImage, error) {
	return s.userDockerImageDao.GetByRepositoryID(ctx, repositoryID)
}

// GetDockerImages 根据条件获取镜像列表
func (s *DockerImageService) GetDockerImages(ctx context.Context, dockerfileID uint32, repositoryID string) ([]*dao.UserDockerImage, error) {
	if dockerfileID > 0 {
		return s.GetDockerImagesByDockerfileID(ctx, dockerfileID)
	}
	if repositoryID != "" {
		return s.GetDockerImagesByRepositoryID(ctx, repositoryID)
	}
	return nil, nil
}
