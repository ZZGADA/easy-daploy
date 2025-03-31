package websocket

import (
	"fmt"
	"github.com/ZZGADA/easy-deploy/internal/model/dao"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gorilla/websocket"
)

type SocketDockerService struct {
	dockerfileDao *dao.UserDockerfileDao
	dockerDao     dao.UserDockerDao
}

func NewSocketDockerService(dockerfileDao *dao.UserDockerfileDao, dockerDao dao.UserDockerDao) *SocketDockerService {
	return &SocketDockerService{
		dockerfileDao: dockerfileDao,
		dockerDao:     dockerDao,
	}
}

// DockerfileInstruction Dockerfile 指令结构
type DockerfileInstruction struct {
	Command string
	Args    string
}

// HandleGenerateDockerfile 处理生成 Dockerfile 的请求
func (s *SocketDockerService) HandleGenerateDockerfile(conn *websocket.Conn, data map[string]interface{}, userID uint) {
	// 验证必要参数
	instructions, ok := data["instructions"].([]interface{})
	if !ok {
		SendError(conn, "缺少 Dockerfile 指令")
		return
	}

	// 创建临时工作目录
	workDir := filepath.Join("/tmp", "easy-deploy", "dockerfiles")
	if err := os.MkdirAll(workDir, 0755); err != nil {
		SendError(conn, fmt.Sprintf("创建工作目录失败: %v", err))
		return
	}

	// 生成 Dockerfile 路径
	dockerfilePath := filepath.Join(workDir, "Dockerfile")

	// 解析并验证指令
	var dockerfileContent strings.Builder
	for _, inst := range instructions {
		instruction, ok := inst.(map[string]interface{})
		if !ok {
			SendError(conn, "无效的指令格式")
			return
		}

		command, ok := instruction["command"].(string)
		if !ok {
			SendError(conn, "指令缺少命令")
			return
		}

		args, ok := instruction["args"].(string)
		if !ok {
			SendError(conn, "指令缺少参数")
			return
		}

		// 验证指令
		if !isValidDockerCommand(command) {
			SendError(conn, fmt.Sprintf("无效的 Dockerfile 指令: %s", command))
			return
		}

		// 添加指令到 Dockerfile 内容
		dockerfileContent.WriteString(fmt.Sprintf("%s %s\n", strings.ToUpper(command), args))
	}

	// 写入 Dockerfile
	if err := os.WriteFile(dockerfilePath, []byte(dockerfileContent.String()), 0644); err != nil {
		SendError(conn, fmt.Sprintf("写入 Dockerfile 失败: %v", err))
		return
	}

	// 验证 Dockerfile 语法
	if err := validateDockerfile(dockerfilePath); err != nil {
		SendError(conn, fmt.Sprintf("Dockerfile 语法验证失败: %v", err))
		// 清理无效的 Dockerfile
		os.Remove(dockerfilePath)
		return
	}

	SendSuccess(conn, "Dockerfile 生成成功", map[string]string{
		"path":    dockerfilePath,
		"content": dockerfileContent.String(),
	})
}

// isValidDockerCommand 验证 Docker 指令是否有效
func isValidDockerCommand(command string) bool {
	validCommands := map[string]bool{
		"from":        true,
		"run":         true,
		"cmd":         true,
		"label":       true,
		"expose":      true,
		"env":         true,
		"add":         true,
		"copy":        true,
		"entrypoint":  true,
		"volume":      true,
		"user":        true,
		"workdir":     true,
		"arg":         true,
		"onbuild":     true,
		"stopsignal":  true,
		"healthcheck": true,
		"shell":       true,
	}

	return validCommands[strings.ToLower(command)]
}

// validateDockerfile 验证 Dockerfile 语法
func validateDockerfile(dockerfilePath string) error {
	cmd := exec.Command("docker", "build", "-f", dockerfilePath, "-t", "dockerfile-validation-test", ".")
	cmd.Dir = filepath.Dir(dockerfilePath)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("Dockerfile 验证失败: %s", string(output))
	}

	// 删除测试镜像
	exec.Command("docker", "rmi", "dockerfile-validation-test").Run()
	return nil
}

// HandleCloneRepository 处理克隆仓库的请求
func (s *SocketDockerService) HandleCloneRepository(conn *websocket.Conn, data map[string]interface{}, userID uint) {
	// 验证必要参数
	repoURL, ok := data["repo_url"].(string)
	if !ok {
		SendError(conn, "缺少仓库 URL")
		return
	}

	branch, _ := data["branch"].(string)
	if branch == "" {
		branch = "main"
	}

	// 创建临时目录
	workDir := filepath.Join("/tmp", "easy-deploy", "repos")
	if err := exec.Command("mkdir", "-p", workDir).Run(); err != nil {
		SendError(conn, fmt.Sprintf("创建工作目录失败: %v", err))
		return
	}

	// 克隆仓库
	cmd := exec.Command("git", "clone", "-b", branch, repoURL, workDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		SendError(conn, fmt.Sprintf("克隆仓库失败: %s", string(output)))
		return
	}

	SendSuccess(conn, "仓库克隆成功", map[string]string{
		"work_dir": workDir,
	})
}

// HandleBuildImage 处理构建镜像的请求
func (s *SocketDockerService) HandleBuildImage(conn *websocket.Conn, data map[string]interface{}, userID uint) {
	// 验证必要参数
	workDir, ok := data["work_dir"].(string)
	if !ok {
		SendError(conn, "缺少工作目录")
		return
	}

	imageName, ok := data["image_name"].(string)
	if !ok {
		SendError(conn, "缺少镜像名称")
		return
	}

	// 检查 Dockerfile 是否存在
	dockerfilePath := filepath.Join(workDir, "Dockerfile")
	if _, err := exec.Command("test", "-f", dockerfilePath).Output(); err != nil {
		SendError(conn, "Dockerfile 不存在")
		return
	}

	// 构建镜像
	cmd := exec.Command("docker", "build", "-t", imageName, ".")
	cmd.Dir = workDir

	// 创建管道读取命令输出
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		SendError(conn, fmt.Sprintf("创建输出管道失败: %v", err))
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		SendError(conn, fmt.Sprintf("创建错误管道失败: %v", err))
		return
	}

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

	go func() {
		buffer := make([]byte, 1024)
		for {
			n, err := stderr.Read(buffer)
			if n > 0 {
				SendSuccess(conn, "build_error", string(buffer[:n]))
			}
			if err != nil {
				break
			}
		}
	}()

	// 等待命令完成
	if err := cmd.Wait(); err != nil {
		log.Printf("构建失败: %v\n", err)
		SendError(conn, "镜像构建失败")
		return
	}

	SendSuccess(conn, "镜像构建成功", map[string]string{
		"image_name": imageName,
	})
}

// WSResponse WebSocket 响应结构
type WSResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
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
