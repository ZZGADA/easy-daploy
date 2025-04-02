package http

import (
	"net/http"

	"github.com/ZZGADA/easy-deploy/internal/model/service/k8s_manage"
	"github.com/gin-gonic/gin"
)

type K8sResourceHandler struct {
	k8sResourceService *k8s_manage.K8sResourceService
}

func NewK8sResourceHandler(k8sResourceService *k8s_manage.K8sResourceService) *K8sResourceHandler {
	return &K8sResourceHandler{
		k8sResourceService: k8sResourceService,
	}
}

type SaveResourceRequest struct {
	RepositoryID string `json:"repository_id" binding:"required"`
	ResourceType string `json:"resource_type" binding:"required"`
	OssURL       string `json:"oss_url" binding:"required"`
}

type DeleteResourceRequest struct {
	ID uint `json:"id" binding:"required"`
}

// SaveResource 保存 K8s 资源配置
func (h *K8sResourceHandler) SaveResource(c *gin.Context) {
	var req SaveResourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证资源类型
	if !h.k8sResourceService.ValidateResourceType(req.ResourceType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid resource type"})
		return
	}

	userID := c.GetUint("user_id")
	if err := h.k8sResourceService.SaveResource(userID, req.RepositoryID, req.ResourceType, req.OssURL); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "success"})
}

// DeleteResource 删除 K8s 资源配置
func (h *K8sResourceHandler) DeleteResource(c *gin.Context) {
	var req DeleteResourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.k8sResourceService.DeleteResource(req.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "success"})
}

// QueryResources 查询 K8s 资源配置列表
func (h *K8sResourceHandler) QueryResources(c *gin.Context) {
	repositoryID := c.Query("repository_id")
	resourceType := c.Query("resource_type")

	if repositoryID == "" || resourceType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "repository_id and resource_type are required"})
		return
	}

	resources, err := h.k8sResourceService.QueryResources(repositoryID, resourceType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "success",
		"data":    resources})
}
