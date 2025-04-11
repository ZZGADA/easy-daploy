package http

import (
	"fmt"
	"github.com/ZZGADA/easy-deploy/internal/model/dao"
	"net/http"
	"strconv"

	"github.com/ZZGADA/easy-deploy/internal/model/service/team_manage"
	"github.com/gin-gonic/gin"
)

// TeamRequest 团队请求结构体
type TeamRequest struct {
	ID              uint   `json:"id,omitempty"`
	TeamName        string `json:"team_name" binding:"required"`
	TeamDescription string `json:"team_description"`
}

type TeamRequestDelete struct {
	TeamId uint `json:"team_id" binding:"required"`
}

// TeamHandler 团队管理处理器
type TeamHandler struct {
	teamService *team_manage.TeamService
}

// NewTeamHandler 创建 TeamHandler 实例
func NewTeamHandler(teamService *team_manage.TeamService) *TeamHandler {
	return &TeamHandler{
		teamService: teamService,
	}
}

// CreateTeam 创建团队
func (h *TeamHandler) CreateTeam(c *gin.Context) {
	userID := c.GetUint("user_id")
	var req TeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "无效的请求参数",
		})
		return
	}

	team, err := h.teamService.CreateTeam(c, req.TeamName, req.TeamDescription, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusInternalServerError,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "创建成功",
		"data":    team,
	})
}

// UpdateTeam 更新团队信息
func (h *TeamHandler) UpdateTeam(c *gin.Context) {
	userID := c.GetUint("user_id")
	var req TeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "无效的请求参数",
		})
		return
	}

	if req.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "团队ID不能为空",
		})
		return
	}

	// 检查用户是否有权限更新团队
	team, err := h.teamService.GetTeamByID(c, uint32(req.ID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "团队不存在",
		})
		return
	}

	if team.CreatorID != uint32(userID) {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    http.StatusForbidden,
			"message": "没有权限更新团队",
		})
		return
	}

	err = h.teamService.UpdateTeam(c, uint32(req.ID), req.TeamName, req.TeamDescription)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "更新成功",
	})
}

// DeleteTeam 删除团队
func (h *TeamHandler) DeleteTeam(c *gin.Context) {
	userID := c.GetUint("user_id")
	var req TeamRequestDelete
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "无效的请求参数",
		})
		return
	}

	// 检查用户是否有权限删除团队
	team, err := h.teamService.GetTeamByID(c, uint32(req.TeamId))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "团队不存在",
		})
		return
	}

	if team.CreatorID != uint32(userID) {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    http.StatusForbidden,
			"message": "没有权限删除团队",
		})
		return
	}

	err = h.teamService.DeleteTeam(c, uint32(req.TeamId))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "删除成功",
	})
}

// GetTeamMemberByID 根据ID获取团队
func (h *TeamHandler) GetTeamMemberByID(c *gin.Context) {
	teamID := c.Query("team_id")
	parseUint, err2 := strconv.ParseUint(teamID, 10, 32)
	if err2 != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "参数错误",
		})
		return
	}

	users, err := h.teamService.GetTeamMemberByID(c, uint32(parseUint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "团队不存在",
		})
		return
	}

	team, err2 := h.teamService.GetTeamByID(c, uint32(parseUint))
	if err2 != nil || team == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "团队不存在",
		})
		return
	}

	var res []map[string]interface{}
	for _, user := range users {
		ifCreator := false
		if user.Id == team.CreatorID {
			ifCreator = true
		}

		res = append(res, map[string]interface{}{
			"id":         team.ID,
			"user_id":    user.Id,
			"user_email": user.Email,
			"if_creator": ifCreator,
			"user_name":  user.GithubName,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "查询成功",
		"data":    res,
	})
}

// QueryTeams 查询团队列表
func (h *TeamHandler) QueryTeams(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("page_size", "10")
	teamName := c.Query("team_name")
	teamUuidS := c.Query("team_uuid")
	teamUuid := 0

	if teamUuidS != "" {
		teamUuidT, err := strconv.ParseInt(teamUuidS, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": "参数格式错误",
			})
		}
		teamUuid = int(teamUuidT)

	}

	pageInt := 1
	pageSizeInt := 10
	_, err := fmt.Sscanf(page, "%d", &pageInt)
	if err != nil {
		pageInt = 1
	}
	_, err = fmt.Sscanf(pageSize, "%d", &pageSizeInt)
	if err != nil {
		pageSizeInt = 10
	}

	teams, total, err := h.teamService.QueryTeams(c, pageInt, pageSizeInt, teamName, teamUuid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "查询成功",
		"data": gin.H{
			"teams": teams,
			"total": total,
		},
	})
}

// GetUserTeam 获取用户自己的团队信息
func (h *TeamHandler) GetUserTeam(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "参数错误",
		})
		return
	}

	team, err := h.teamService.GetUserTeam(c, uint32(userID))
	if err != nil {
		if err.Error() == "用户未加入任何团队" {
			c.JSON(http.StatusOK, gin.H{
				"code":    http.StatusOK,
				"message": err.Error(),
				"data":    dao.Team{},
			})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "查询成功",
		"data":    team,
	})
}
