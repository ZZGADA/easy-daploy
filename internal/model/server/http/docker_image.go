package http

import (
	"net/http"
	"strconv"

	"github.com/ZZGADA/easy-deploy/internal/model/service/docker_manage"
	"github.com/gin-gonic/gin"
)

type DockerImageHandler struct {
	dockerImageService *docker_manage.DockerImageService
}

func NewDockerImageHandler(dockerImageService *docker_manage.DockerImageService) *DockerImageHandler {
	return &DockerImageHandler{
		dockerImageService: dockerImageService,
	}
}

// QueryDockerImages 查询 Docker 镜像列表
func (h *DockerImageHandler) QueryDockerImages(c *gin.Context) {
	dockerfileIDStr := c.Query("dockerfile_id")
	repositoryID := c.Query("repository_id")

	// 检查是否提供了必要的查询参数
	if dockerfileIDStr == "" && repositoryID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "缺少必要的查询参数，请提供 dockerfile_id 或 repository_id",
		})
		return
	}

	var dockerfileID uint32
	var err error

	// 如果提供了 dockerfile_id，则解析它
	if dockerfileIDStr != "" {
		dockerfileIDUint, err := strconv.ParseUint(dockerfileIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "dockerfile_id 参数格式错误",
			})
			return
		}
		dockerfileID = uint32(dockerfileIDUint)
	}

	// 调用 service 层方法获取镜像列表
	images, err := h.dockerImageService.GetDockerImages(c.Request.Context(), dockerfileID, repositoryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": images,
	})
}
