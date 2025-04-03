package http

import (
	"net/http"

	"github.com/ZZGADA/easy-deploy/internal/model/service/oss_manage"
	"github.com/gin-gonic/gin"
)

type OssHandler struct {
	ossService *oss_manage.OssService
}

func NewOssHandler(ossService *oss_manage.OssService) *OssHandler {
	return &OssHandler{
		ossService: ossService,
	}
}

type OssAccessRequest struct {
	AccessKeyID     string `json:"access_key_id" binding:"required"`
	AccessKeySecret string `json:"access_key_secret" binding:"required"`
	Bucket          string `json:"bucket" binding:"required"`
	Region          string `json:"region" binding:"required"`
}

// SaveOssAccess 保存 OSS 访问信息
func (h *OssHandler) SaveOssAccess(c *gin.Context) {
	var req OssAccessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误", "error": err.Error()})
		return
	}

	userID := c.GetUint("user_id")
	if err := h.ossService.SaveOssAccess(userID, req.AccessKeyID, req.AccessKeySecret, req.Bucket, req.Region); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "保存失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "保存成功"})
}

// UpdateOssAccess 更新 OSS 访问信息
func (h *OssHandler) UpdateOssAccess(c *gin.Context) {
	var req OssAccessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误", "error": err.Error()})
		return
	}

	userID := c.GetUint("user_id")
	if err := h.ossService.UpdateOssAccess(userID, req.AccessKeyID, req.AccessKeySecret, req.Bucket, req.Region); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "更新成功"})
}

// QueryOssAccess 查询 OSS 访问信息
func (h *OssHandler) QueryOssAccess(c *gin.Context) {
	userID := c.GetUint("user_id")
	oss, err := h.ossService.QueryOssAccess(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "查询成功",
		"data":    oss,
	})
}

// DeleteOssAccess 删除 OSS 访问信息
func (h *OssHandler) DeleteOssAccess(c *gin.Context) {
	userID := c.GetUint("user_id")
	if err := h.ossService.DeleteOssAccess(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}
