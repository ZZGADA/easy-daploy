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
		c.JSON(400, gin.H{"error": "缺少 k8s_resource_id 参数"})
		return
	}

	// 将 k8s_resource_id 转换为 uint
	k8sResourceID, err := strconv.ParseUint(k8sResourceIDStr, 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"error": "k8s_resource_id 参数格式不正确"})
		return
	}

	// 查询操作日志
	logs, err := h.k8sResourceOperationLogService.QueryByK8sResourceID(uint(k8sResourceID))
	if err != nil {
		c.JSON(500, gin.H{"error": "查询操作日志失败: " + err.Error()})
		return
	}

	// 返回操作日志
	c.JSON(200, gin.H{
		"code":    http.StatusOK,
		"message": "success",
		"logs":    logs,
	})
}
