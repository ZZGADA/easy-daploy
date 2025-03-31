package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ZZGADA/easy-deploy/internal/model/dao"

	"github.com/gorilla/websocket"
)

type SocketDockerService struct {
	userDockerfileDao *dao.UserDockerfileDao
	userDockerDao     dao.UserDockerDao
	userGithubDao     *dao.UserGithubDao
}

func NewSocketDockerService(dockerfileDao *dao.UserDockerfileDao, dockerDao dao.UserDockerDao, GithubDao *dao.UserGithubDao) *SocketDockerService {
	return &SocketDockerService{
		userDockerfileDao: dockerfileDao,
		userDockerDao:     dockerDao,
		userGithubDao:     GithubDao,
	}
}

// WSRequest WebSocket 请求结构
type WSRequest struct {
	DockerBuildStep string                 `json:"docker_build_step"`
	Data            map[string]interface{} `json:"data"`
}

// WSResponse WebSocket 响应结构
type WSResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// HandleGenerateDockerfile 处理生成 Dockerfile 的请求
func (s *SocketDockerService) HandleGenerateDockerfile(conn *websocket.Conn, data map[string]interface{}, userID uint) {
	ctx := context.Background()
	// 验证必要参数
	dockerfileID, ok := data["id"].(float64)
	if !ok {
		SendError(conn, "缺少 Dockerfile ID")
		return
	}

	// 从数据库获取 Dockerfile 信息
	dockerfile, err := s.userDockerfileDao.GetByID(ctx, uint32(dockerfileID))
	if err != nil {
		SendError(conn, fmt.Sprintf("获取 Dockerfile 失败: %v", err))
		return
	}

	// 创建 dockerfile 目录
	dockerfileDir := filepath.Join("docker", dockerfile.RepositoryName)
	if err := os.MkdirAll(dockerfileDir, 0755); err != nil {
		SendError(conn, fmt.Sprintf("创建目录失败: %v", err))
		return
	}

	// 生成文件名
	timestamp := time.Now().Format("20060102150405")
	filename := fmt.Sprintf("dockerfile_%d_%s", dockerfile.Id, timestamp)
	dockerfilePath := filepath.Join(dockerfileDir, filename)

	// 生成 Dockerfile 内容
	var content strings.Builder
	var fileData []dao.DockerfileItem
	if err := json.Unmarshal([]byte(dockerfile.FileData), &fileData); err != nil {
		SendError(conn, fmt.Sprintf("解析 Dockerfile 数据失败: %v", err))
		return
	}

	// 按顺序写入指令
	for _, item := range fileData {
		content.WriteString(fmt.Sprintf("%s %s\n", item.DockerfileKey, item.ShellValue))
	}

	// 写入文件
	if err := os.WriteFile(dockerfilePath, []byte(content.String()), 0644); err != nil {
		SendError(conn, fmt.Sprintf("写入 Dockerfile 失败: %v", err))
		return
	}

	SendSuccess(conn, "Dockerfile build success", map[string]string{
		"filename": filename,
	})
}

// HandleCloneRepository 处理克隆仓库的请求
func (s *SocketDockerService) HandleCloneRepository(conn *websocket.Conn, data map[string]interface{}, userID uint) {
	ctx := context.Background()
	// 验证必要参数
	dockerfileID, ok := data["id"].(float64)
	if !ok {
		SendError(conn, "缺少 Dockerfile ID")
		return
	}

	// 从数据库获取 Dockerfile 信息
	dockerfile, err := s.userDockerfileDao.GetByID(ctx, uint32(dockerfileID))
	if err != nil {
		SendError(conn, fmt.Sprintf("获取 Dockerfile 失败: %v", err))
		return
	}

	// 获取用户的 GitHub 信息
	githubInfo, err := s.userGithubDao.GetByUserID(ctx, userID)
	if err != nil {
		SendError(conn, fmt.Sprintf("获取 GitHub 信息失败: %v", err))
		return
	}

	// 创建仓库目录
	repoDir := filepath.Join("docker", dockerfile.RepositoryName)
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		SendError(conn, fmt.Sprintf("创建目录失败: %v", err))
		return
	}

	// 如果目录已存在，先删除
	if _, err := os.Stat(repoDir); err == nil {
		if err := os.RemoveAll(repoDir); err != nil {
			SendError(conn, fmt.Sprintf("删除已存在的仓库目录失败: %v", err))
			return
		}
	}

	// 克隆仓库
	repoURL := fmt.Sprintf("https://github.com/%s/%s.git", githubInfo.Login, dockerfile.RepositoryName)
	cmd := exec.Command("git", "clone", "-b", dockerfile.BranchName, repoURL, repoDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		SendError(conn, fmt.Sprintf("克隆仓库失败: %s", string(output)))
		return
	}

	SendSuccess(conn, "git clone success", nil)
}

// HandleBuildImage 处理构建镜像的请求
func (s *SocketDockerService) HandleBuildImage(conn *websocket.Conn, data map[string]interface{}, userID uint) {
	ctx := context.Background()
	// 验证必要参数
	dockerfileID, ok := data["id"].(float64)
	if !ok {
		SendError(conn, "缺少 Dockerfile ID")
		return
	}

	imageName, ok := data["docker_image_name"].(string)
	if !ok {
		SendError(conn, "缺少镜像名称")
		return
	}

	// 获取 Dockerfile 信息
	dockerfile, err := s.userDockerfileDao.GetByID(ctx, uint32(dockerfileID))
	if err != nil {
		SendError(conn, fmt.Sprintf("获取 Dockerfile 失败: %v", err))
		return
	}

	// 获取 Docker 账号信息
	dockerAccount, err := s.userDockerDao.GetLoginAccount(userID)
	if err != nil {
		SendError(conn, fmt.Sprintf("获取 Docker 账号失败: %v", err))
		return
	}

	if dockerAccount == nil {
		SendError(conn, "未找到已登录的 Docker 账号")
		return
	}

	// 构建完整的镜像名称
	fullImageName := fmt.Sprintf("%s/%s/%s", dockerAccount.Server, dockerAccount.Namespace, imageName)

	// 构建镜像
	repoDir := filepath.Join("docker", dockerfile.RepositoryName)
	dockerfilePath := filepath.Join(repoDir, fmt.Sprintf("dockerfile_%d_*", dockerfile.Id))

	// 查找最新的 Dockerfile
	matches, err := filepath.Glob(dockerfilePath)
	if err != nil || len(matches) == 0 {
		SendError(conn, "未找到 Dockerfile")
		return
	}
	latestDockerfile := matches[len(matches)-1]

	exePath, err := os.Executable()
	if err != nil {
		SendError(conn, "系统 错误")
		return
	}
	// 获取执行文件所在的目录，即项目目录（如果执行文件在项目根目录下）
	cur := filepath.Dir(exePath)
	parent := filepath.Dir(cur)
	log.Info("当前项目地址：", parent)

	// 构建镜像
	cmd := exec.Command("docker", "build", "-f", filepath.Join(parent, latestDockerfile), "-t", fullImageName, ".")
	cmd.Dir = repoDir

	// 创建管道读取命令输出
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		SendError(conn, fmt.Sprintf("创建输出管道失败: %v", err))
		return
	}

	//stderr, err := cmd.StderrPipe()
	//if err != nil {
	//	SendError(conn, fmt.Sprintf("创建错误管道失败: %v", err))
	//	return
	//}

	// 启动命令
	if err := cmd.Start(); err != nil {
		SendError(conn, fmt.Sprintf("启动构建失败: %v", err))
		return
	}

	// 读取并发送输出
	go func() {
		buffer := make([]byte, 1024)
		for {
			n, err := stdout.Read(buffer)
			if n > 0 {
				SendSuccess(conn, "build_output", string(buffer[:n]))
			}
			if err != nil {
				break
			}
		}
	}()

	//go func() {
	//	buffer := make([]byte, 1024)
	//	for {
	//		n, err := stderr.Read(buffer)
	//		if n > 0 {
	//			SendSuccess(conn, "build_error", string(buffer[:n]))
	//		}
	//		if err != nil {
	//			break
	//		}
	//	}
	//}()

	// 等待命令完成
	if err := cmd.Wait(); err != nil {
		SendError(conn, "镜像构建失败")
		return
	}

	// 推送镜像到仓库
	pushCmd := exec.Command("docker", "push", fullImageName)
	output, err := pushCmd.CombinedOutput()
	if err != nil {
		SendError(conn, fmt.Sprintf("推送镜像失败: %s", string(output)))
		return
	}

	SendSuccess(conn, "build & push success", map[string]string{
		"image_name": fullImageName,
	})
}

// SendError 发送错误消息
func SendError(conn *websocket.Conn, message string) {
	response := WSResponse{
		Success: false,
		Message: message,
	}
	conn.WriteJSON(response)
}

// SendSuccess 发送成功消息
func SendSuccess(conn *websocket.Conn, message string, data interface{}) {
	response := WSResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
	conn.WriteJSON(response)
}
