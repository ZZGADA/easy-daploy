package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/ZZGADA/easy-deploy/internal/model/dao"

	"github.com/gorilla/websocket"
)

// HandleGenerateDockerfile 处理生成 Dockerfile 的请求
func (s *SocketService) HandleGenerateDockerfile(conn *websocket.Conn, data map[string]interface{}, userID uint) {
	log.Info("=== HandleGenerateDockerfile 开始 ===")
	log.Infof("接收到的数据: %+v", data)

	ctx := context.Background()
	// 验证必要参数
	dockerfileID, ok := data["id"].(float64)
	if !ok {
		log.Error("缺少 Dockerfile ID")
		SendError(conn, "缺少 Dockerfile ID")
		return
	}

	log.Infof("处理参数 - DockerfileID: %v", dockerfileID)

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

	log.Info("=== HandleGenerateDockerfile 结束 ===")
}

// HandleCloneRepository 处理克隆仓库的请求
func (s *SocketService) HandleCloneRepository(conn *websocket.Conn, data map[string]interface{}, userID uint) {
	log.Info("=== HandleCloneRepository 开始 ===")
	log.Infof("接收到的数据: %+v", data)

	ctx := context.Background()
	// 验证必要参数
	dockerfileID, ok := data["id"].(float64)
	if !ok {
		log.Error("缺少 Dockerfile ID")
		SendError(conn, "缺少 Dockerfile ID")
		return
	}

	log.Infof("处理参数 - DockerfileID: %v", dockerfileID)

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

	log.Info("=== HandleCloneRepository 结束 ===")
}

// HandleBuildImage 处理构建镜像的请求
func (s *SocketService) HandleBuildImage(conn *websocket.Conn, data map[string]interface{}, userID uint) string {
	log.Info("=== HandleBuildImage 开始 ===")
	log.Infof("接收到的数据: %+v", data)

	ctx := context.Background()
	// 验证必要参数
	dockerfileID, ok := data["id"].(float64)
	if !ok {
		log.Error("缺少 Dockerfile ID")
		SendError(conn, "缺少 Dockerfile ID")
		return ""
	}

	imageName, ok := data["docker_image_name"].(string)
	if !ok {
		log.Error("缺少镜像名称")
		SendError(conn, "缺少镜像名称")
		return ""
	}

	// 获取 Dockerfile 信息
	dockerfile, err := s.userDockerfileDao.GetByID(ctx, uint32(dockerfileID))
	if err != nil {
		log.Errorf("获取 Dockerfile 失败: %v", err)
		SendError(conn, fmt.Sprintf("获取 Dockerfile 失败: %v", err))
		return ""
	}

	// 获取 Docker 账号信息
	dockerAccount, err := s.userDockerDao.GetLoginAccount(userID)
	if err != nil {
		log.Errorf("获取 Docker 账号失败: %v", err)
		SendError(conn, fmt.Sprintf("获取 Docker 账号失败: %v", err))
		return ""
	}

	if dockerAccount == nil {
		log.Error("未找到已登录的 Docker 账号")
		SendError(conn, "未找到已登录的 Docker 账号")
		return ""
	}

	// 构建完整的镜像名称
	fullImageName := fmt.Sprintf("%s/%s/%s", dockerAccount.Server, dockerAccount.Namespace, imageName)
	log.Infof("构建完整镜像名称: %s", fullImageName)

	// 构建镜像
	repoDir := filepath.Join("docker", dockerfile.RepositoryName)
	dockerfilePath := filepath.Join(repoDir, fmt.Sprintf("dockerfile_%d_*", dockerfile.Id))

	// 查找最新的 Dockerfile
	matches, err := filepath.Glob(dockerfilePath)
	if err != nil || len(matches) == 0 {
		SendError(conn, "未找到 Dockerfile")
		return ""
	}
	latestDockerfile := matches[len(matches)-1]

	exePath, err := os.Executable()
	if err != nil {
		SendError(conn, "系统 错误")
		return ""
	}
	// 获取执行文件所在的目录，即项目目录（如果执行文件在项目根目录下）
	cur := filepath.Dir(exePath)
	parent := filepath.Dir(cur)
	log.Info("当前项目地址：", parent)

	// shell 脚本执行
	if dockerfile.ShellPath != "" {
		cmdBuild := exec.Command("/bin/bash", filepath.Join(parent, "docker", dockerfile.ShellPath))
		log.Infof("project build and test path :%s", cmdBuild.Path)

		err = runMonitoring(cmdBuild, conn)
		if err != nil {
			return ""
		}
	}

	// 构建镜像
	cmd := exec.Command("docker", "build", "-f", filepath.Join(parent, latestDockerfile), "-t", fullImageName, ".")
	cmd.Dir = repoDir

	err = runMonitoring(cmd, conn)
	if err != nil {
		return ""
	}

	// 推送镜像到仓库
	pushCmd := exec.Command("docker", "push", fullImageName)
	output, err := pushCmd.CombinedOutput()
	if err != nil {
		SendError(conn, fmt.Sprintf("推送镜像失败: %s", string(output)))
		return ""
	}

	SendSuccess(conn, "docker build & push success", map[string]string{
		"image_name": fullImageName,
	})

	log.Info("=== HandleBuildImage 结束 ===")

	return fullImageName
}

func runMonitoring(cmd *exec.Cmd, conn *websocket.Conn) error {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		SendError(conn, fmt.Sprintf("创建输出管道失败: %v", err))
		return err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		SendError(conn, fmt.Sprintf("创建错误管道失败: %v", err))
		return err
	}

	// 启动命令
	if err := cmd.Start(); err != nil {
		SendError(conn, fmt.Sprintf("启动构建失败: %v", err))
		return err
	}

	// 读取并发送 stdout 输出
	go func() {
		buffer := make([]byte, 1024)
		for {
			n, err := stdout.Read(buffer)
			if n > 0 {
				log.Infof("stdout: %s", string(buffer[:n]))
				SendSuccess(conn, "build_output", string(buffer[:n]))
			}
			if err != nil {
				log.Info("stdout 读取结束")
				break
			}
		}
	}()

	// 读取并发送 stderr 输出
	// Docker 构建过程中的输出默认是写入到 stderr 而不是 stdout。
	go func() {
		buffer := make([]byte, 1024)
		for {
			n, err := stderr.Read(buffer)
			if n > 0 {
				log.Infof("stderr: %s", string(buffer[:n]))
				SendSuccess(conn, "build_output", string(buffer[:n]))
			}
			if err != nil {
				log.Info("stderr 读取结束")
				break
			}
		}
	}()

	// 等待命令完成
	if err := cmd.Wait(); err != nil {
		SendError(conn, "镜像构建失败")
		return err
	}
	return nil
}
