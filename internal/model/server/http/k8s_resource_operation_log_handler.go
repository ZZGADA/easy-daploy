package http

import (
	"net/http"
	"strconv"

	"github.com/ZZGADA/easy-deploy/internal/model/service/k8s_manage"
	"github.com/gin-gonic/gin"
)

// K8sResourceOperationLogHandler K8s 资源操作日志处理程序
type K8sResourceOperationLogHandler struct {
	k8sResourceOperationLogService *k8s_manage.K8sResourceOperationLogService
}

// NewK8sResourceOperationLogHandler 创建 K8s 资源操作日志处理程序
func NewK8sResourceOperationLogHandler(k8sResourceOperationLogService *k8s_manage.K8sResourceOperationLogService) *K8sResourceOperationLogHandler {
	return &K8sResourceOperationLogHandler{
		k8sResourceOperationLogService: k8sResourceOperationLogService,
	}
}

// QueryOperationLogs 查询 K8s 资源操作日志
func (h *K8sResourceOperationLogHandler) QueryOperationLogs(c *gin.Context) {
	// 从请求参数中获取 k8s_resource_id
	k8sResourceIDStr := c.Query("k8s_resource_id")
	if k8sResourceIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "k8s_resource_id is required"})
		return
	}

	// 将 k8s_resource_id 转换为 uint
	k8sResourceID, err := strconv.ParseUint(k8sResourceIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid k8s_resource_id"})
		return
	}

	// 获取分页参数
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 10
	}

	// 调用服务层查询操作日志
	logs, total, err := h.k8sResourceOperationLogService.QueryByK8sResourceID(uint(k8sResourceID), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	// 返回操作日志
	c.JSON(http.StatusOK, gin.H{
		"code":      200,
		"message":   "success",
		"logs":      logs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}
