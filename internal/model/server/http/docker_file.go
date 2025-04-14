package http

import (
	"github.com/ZZGADA/easy-deploy/internal/model/service/docker_manage"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DockerfileHandler struct {
	dockerfileService *docker_manage.DockerfileService
}

func NewDockerfileHandler(dockerfileService *docker_manage.DockerfileService) *DockerfileHandler {
	return &DockerfileHandler{
		dockerfileService: dockerfileService,
	}
}

// UploadDockerfile 处理 Dockerfile 上传请求
func (h *DockerfileHandler) UploadDockerfile(c *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    http.StatusUnauthorized,
			"message": "未授权的访问",
		})
		return
	}

	var req docker_manage.DockerfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "无效的请求参数",
		})
		return
	}

	// 上传 Dockerfile
	err := h.dockerfileService.UploadDockerfile(c.Request.Context(), uint32(userID.(uint)), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "上传 Dockerfile 失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "上传 Dockerfile 成功",
	})
}

// UpdateDockerfile 处理 Dockerfile 更新请求
func (h *DockerfileHandler) UpdateDockerfile(c *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    http.StatusUnauthorized,
			"message": "未授权的访问",
		})
		return
	}

	var req docker_manage.DockerfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "无效的请求参数",
		})
		return
	}

	// 更新 Dockerfile
	err := h.dockerfileService.UpdateDockerfile(c.Request.Context(), uint32(userID.(uint)), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "更新 Dockerfile 失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "更新 Dockerfile 成功",
	})
}

// DeleteDockerfile 处理 Dockerfile 删除请求
func (h *DockerfileHandler) DeleteDockerfile(c *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    http.StatusUnauthorized,
			"message": "未授权的访问",
		})
		return
	}

	var req docker_manage.DockerfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "无效的请求参数",
		})
		return
	}

	// 删除 Dockerfile
	err := h.dockerfileService.DeleteDockerfile(c.Request.Context(), uint32(userID.(uint)), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "删除 Dockerfile 失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "删除 Dockerfile 成功",
	})
}

// QueryDockerfile 处理 Dockerfile 查询请求
func (h *DockerfileHandler) QueryDockerfile(c *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    http.StatusUnauthorized,
			"message": "未授权的访问",
		})
		return
	}

	// 获取查询参数
	repositoryId := c.Query("repository_id")
	branchName := c.Query("branch_name")

	// 验证必要参数
	if repositoryId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "仓库ID不能为空",
		})
		return
	}

	if branchName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "分支名称不能为空",
		})
		return
	}

	// 查询 Dockerfile 列表
	dockerfiles, err := h.dockerfileService.QueryDockerfilesByRepoAndBranch(
		c.Request.Context(),
		uint32(userID.(uint)),
		repositoryId,
		branchName,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "查询 Dockerfile 列表失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"data": dockerfiles,
	})
}

// SaveShellPath 保存shell
func (h *DockerfileHandler) SaveShellPath(c *gin.Context) {
	userID := c.GetUint("user_id")

	var res docker_manage.ShellPathRequest
	if err := c.ShouldBindJSON(&res); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "参数错误",
		})
		return
	}

	err := h.dockerfileService.SaveShellPath(c, uint32(userID), &res)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "error",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "success",
	})

}
