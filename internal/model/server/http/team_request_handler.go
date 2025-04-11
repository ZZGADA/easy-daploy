package http

import (
	"net/http"
	"strconv"

	"github.com/ZZGADA/easy-deploy/internal/model/service/team_manage"
	"github.com/gin-gonic/gin"
)

// TeamRequestCreateRequest 创建团队申请请求结构体
type TeamRequestCreateRequest struct {
	TeamID      uint32 `json:"team_id" binding:"required"`
	RequestType int    `json:"request_type" binding:"required"`
}

// TeamRequestHandleRequest 处理团队申请请求结构体
type TeamRequestHandleRequest struct {
	RequestID uint32 `json:"request_id" binding:"required"`
	Status    int    `json:"status" binding:"required"` // 1: 同意, 2: 拒绝
}

// TeamRequestHandler 团队申请处理器
type TeamRequestHandler struct {
	teamRequestService *team_manage.TeamRequestService
	teamService        *team_manage.TeamService
}

// NewTeamRequestHandler 创建 TeamRequestHandler 实例
func NewTeamRequestHandler(teamRequestService *team_manage.TeamRequestService, teamService *team_manage.TeamService) *TeamRequestHandler {
	return &TeamRequestHandler{
		teamRequestService: teamRequestService,
		teamService:        teamService,
	}
}

// CreateTeamRequest 创建团队申请
func (h *TeamRequestHandler) CreateTeamRequest(c *gin.Context) {
	userID := c.GetUint("user_id")
	var req TeamRequestCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "无效的请求参数",
		})
		return
	}

	// 检查团队是否存在
	team, err := h.teamService.GetTeamByID(c, req.TeamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "团队不存在",
		})
		return
	}

	// 检查用户是否已经在团队中
	if team.CreatorID == uint32(userID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "您已经是团队成员",
		})
		return
	}

	_, err = h.teamRequestService.CreateTeamRequest(c, req.TeamID, userID, req.RequestType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "申请已提交",
	})
}

// CheckTeamRequest 请求校验接口
func (h *TeamRequestHandler) CheckTeamRequest(c *gin.Context) {
	requestId := c.Query("request_id")
	status := c.Query("status")
	if status == "" || requestId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "参数信息不完善",
		})
		return
	}

	requestIdI, err2 := strconv.ParseUint(requestId, 10, 32)
	if err2 != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "系统错误",
		})
		return
	}
	statusI, err2 := strconv.ParseInt(status, 10, 32)
	if err2 != nil {

		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "系统错误",
		})
		return

	}

	err := h.teamRequestService.CheckTeamRequest(c, uint32(requestIdI), int(statusI))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusBadRequest,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "OK",
	})
	return
}

//// GetTeamRequestByID 根据ID获取申请
//func (h *TeamRequestHandler) GetTeamRequestByID(c *gin.Context) {
//	requestID := c.GetUint32("request_id")
//
//	request, err := h.teamRequestService.GetTeamRequestByID(c, requestID)
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{
//			"code":    http.StatusInternalServerError,
//			"message": "申请不存在",
//		})
//		return
//	}
//
//	c.JSON(http.StatusOK, gin.H{
//		"code":    http.StatusOK,
//		"message": "查询成功",
//		"data":    request,
//	})
//}

//// GetTeamRequestsByTeamID 获取团队的申请列表
//func (h *TeamRequestHandler) GetTeamRequestsByTeamID(c *gin.Context) {
//	userID := c.GetUint32("user_id")
//	teamID := c.GetUint32("team_id")
//
//	// 检查用户是否有权限查看申请
//	team, err := h.teamService.GetTeamByID(c, teamID)
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{
//			"code":    http.StatusInternalServerError,
//			"message": "团队不存在",
//		})
//		return
//	}
//
//	if team.CreatorID != userID {
//		c.JSON(http.StatusForbidden, gin.H{
//			"code":    http.StatusForbidden,
//			"message": "没有权限查看申请",
//		})
//		return
//	}
//
//	requests, err := h.teamRequestService.GetTeamRequestsByTeamID(c, teamID)
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{
//			"code":    http.StatusInternalServerError,
//			"message": err.Error(),
//		})
//		return
//	}
//
//	c.JSON(http.StatusOK, gin.H{
//		"code":    http.StatusOK,
//		"message": "查询成功",
//		"data":    requests,
//	})
//}

//// GetTeamRequestsByUserID 获取用户的申请列表
//func (h *TeamRequestHandler) GetTeamRequestsByUserID(c *gin.Context) {
//	userID := c.GetUint32("user_id")
//
//	requests, err := h.teamRequestService.GetTeamRequestsByUserID(c, userID)
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{
//			"code":    http.StatusInternalServerError,
//			"message": err.Error(),
//		})
//		return
//	}
//
//	c.JSON(http.StatusOK, gin.H{
//		"code":    http.StatusOK,
//		"message": "查询成功",
//		"data":    requests,
//	})
//}

//// GetPendingTeamRequestsByTeamID 获取团队待处理的申请列表
//func (h *TeamRequestHandler) GetPendingTeamRequestsByTeamID(c *gin.Context) {
//	userID := c.GetUint32("user_id")
//	teamID := c.GetUint32("team_id")
//
//	// 检查用户是否有权限查看申请
//	team, err := h.teamService.GetTeamByID(c, teamID)
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{
//			"code":    http.StatusInternalServerError,
//			"message": "团队不存在",
//		})
//		return
//	}
//
//	if team.CreatorID != userID {
//		c.JSON(http.StatusForbidden, gin.H{
//			"code":    http.StatusForbidden,
//			"message": "没有权限查看申请",
//		})
//		return
//	}
//
//	requests, err := h.teamRequestService.GetPendingTeamRequestsByTeamID(c, teamID)
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{
//			"code":    http.StatusInternalServerError,
//			"message": err.Error(),
//		})
//		return
//	}
//
//	c.JSON(http.StatusOK, gin.H{
//		"code":    http.StatusOK,
//		"message": "查询成功",
//		"data":    requests,
//	})
//}
