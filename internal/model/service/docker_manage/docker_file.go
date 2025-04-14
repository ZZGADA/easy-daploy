package docker_manage

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ZZGADA/easy-deploy/internal/model/dao"
)

type DockerfileService struct {
	dockerfileDao *dao.UserDockerfileDao
}

func NewDockerfileService(dockerfileDao *dao.UserDockerfileDao) *DockerfileService {
	return &DockerfileService{
		dockerfileDao: dockerfileDao,
	}
}

// DockerfileRequest 表示前端传来的 Dockerfile 请求
type DockerfileRequest struct {
	Id             uint32               `json:"id"`
	RepositoryName string               `json:"repository_name"`
	RepositoryId   string               `json:"repository_id"`
	BranchName     string               `json:"branch_name"`
	FileName       string               `json:"file_name"`
	FileData       []dao.DockerfileItem `json:"file_data"`
	ShellPath      string               `json:"shell_path"`
}

type ShellPathRequest struct {
	ShellPath    string `json:"shell_path"`
	DockerFileId uint32 `json:"dockerfile_id"`
}

// UploadDockerfile 上传 Dockerfile
func (s *DockerfileService) UploadDockerfile(ctx context.Context, userId uint32, req *DockerfileRequest) error {
	// 检查是否已存在
	//existing, err := s.dockerfileDao.GetByUserIDAndRepo(ctx, userId, req.RepositoryName)
	//if err == nil && existing != nil {
	//	return errors.New("该仓库的 Dockerfile 已存在，请使用更新接口")
	//}

	// 将 FileData 转换为 JSON 字符串
	fileDataJSON, err := json.Marshal(req.FileData)
	if err != nil {
		return fmt.Errorf("序列化 Dockerfile 数据失败: %v", err)
	}

	// 创建新的 Dockerfile 记录
	dockerfile := &dao.UserDockerfile{
		UserId:         userId,
		RepositoryName: req.RepositoryName,
		RepositoryId:   req.RepositoryId,
		BranchName:     req.BranchName,
		FileName:       req.FileName,
		FileData:       string(fileDataJSON),
	}

	return s.dockerfileDao.Create(ctx, dockerfile)
}

// SaveShellPath 上传 Dockerfile
func (s *DockerfileService) SaveShellPath(ctx context.Context, userId uint32, req *ShellPathRequest) error {
	// 检查是否已存在
	existing, err := s.dockerfileDao.GetByID(ctx, req.DockerFileId)
	if err != nil {
		return nil
	}

	if existing.ShellPath != req.ShellPath {
		existing.ShellPath = req.ShellPath
	}

	return s.dockerfileDao.Update(ctx, existing)
}

// UpdateDockerfile 更新 Dockerfile
func (s *DockerfileService) UpdateDockerfile(ctx context.Context, userId uint32, req *DockerfileRequest) error {
	// 检查是否存在
	existing, err := s.dockerfileDao.GetByID(ctx, req.Id)
	if err != nil {
		return fmt.Errorf("获取 Dockerfile 失败: %v", err)
	}

	// 将 FileData 转换为 JSON 字符串
	fileDataJSON, err := json.Marshal(req.FileData)
	if err != nil {
		return fmt.Errorf("序列化 Dockerfile 数据失败: %v", err)
	}

	// 更新现有记录
	existing.FileName = req.FileName
	existing.FileData = string(fileDataJSON)

	return s.dockerfileDao.Update(ctx, existing)
}

// DeleteDockerfile 删除 Dockerfile
func (s *DockerfileService) DeleteDockerfile(ctx context.Context, userId uint32, req DockerfileRequest) error {
	return s.dockerfileDao.Delete(ctx, req.Id)
}

// QueryDockerfile 查询 Dockerfile
func (s *DockerfileService) QueryDockerfile(ctx context.Context, userId uint32, repositoryId string) (*DockerfileRequest, error) {
	// 获取 Dockerfile 记录
	dockerfile, err := s.dockerfileDao.GetByUserIDAndRepo(ctx, userId, repositoryId)
	if err != nil {
		return nil, fmt.Errorf("获取 Dockerfile 失败: %v", err)
	}

	// 解析 FileData JSON 字符串
	var fileData []dao.DockerfileItem
	if err := json.Unmarshal([]byte(dockerfile.FileData), &fileData); err != nil {
		return nil, fmt.Errorf("解析 Dockerfile 数据失败: %v", err)
	}

	// 构建响应
	return &DockerfileRequest{
		RepositoryName: dockerfile.RepositoryName,
		RepositoryId:   dockerfile.RepositoryId,
		BranchName:     dockerfile.BranchName,
		FileName:       dockerfile.FileName,
		FileData:       fileData,
	}, nil
}

// QueryDockerfilesByRepoAndBranch 查询指定仓库和分支下的所有 Dockerfile
func (s *DockerfileService) QueryDockerfilesByRepoAndBranch(ctx context.Context, userId uint32, repositoryId string, branchName string) ([]*DockerfileRequest, error) {
	// 获取所有匹配的 Dockerfile 记录
	dockerfiles, err := s.dockerfileDao.GetByRepoIDAndBranch(ctx, repositoryId, branchName)
	if err != nil {
		return nil, fmt.Errorf("获取 Dockerfile 列表失败: %v", err)
	}

	// 如果没有找到记录，返回空列表
	if len(dockerfiles) == 0 {
		return []*DockerfileRequest{}, nil
	}

	// 构建响应列表
	result := make([]*DockerfileRequest, 0, len(dockerfiles))
	for _, dockerfile := range dockerfiles {
		// 解析每个 Dockerfile 的 FileData JSON 字符串
		var fileData []dao.DockerfileItem
		if err := json.Unmarshal([]byte(dockerfile.FileData), &fileData); err != nil {
			return nil, fmt.Errorf("解析 Dockerfile 数据失败: %v", err)
		}

		// 添加到结果列表
		result = append(result, &DockerfileRequest{
			Id:             dockerfile.Id,
			RepositoryName: dockerfile.RepositoryName,
			RepositoryId:   dockerfile.RepositoryId,
			BranchName:     dockerfile.BranchName,
			FileName:       dockerfile.FileName,
			FileData:       fileData,
			ShellPath:      dockerfile.ShellPath,
		})
	}

	return result, nil
}
