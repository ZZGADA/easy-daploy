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
	if dockerfileIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "缺少 dockerfile_id 参数",
		})
		return
	}

	dockerfileID, err := strconv.ParseUint(dockerfileIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "dockerfile_id 参数格式错误",
		})
		return
	}

	images, err := h.dockerImageService.GetDockerImagesByDockerfileID(c.Request.Context(), uint32(dockerfileID))
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
