package handlers

import (
	"mcs-backend/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// UserGroupHandler 用户组处理器
type UserGroupHandler struct {
	groupService *services.UserGroupService
}

// NewUserGroupHandler 创建用户组处理器实例
func NewUserGroupHandler() *UserGroupHandler {
	return &UserGroupHandler{
		groupService: services.NewUserGroupService(),
	}
}

// CreateGroup 创建用户组
// @Summary 创建用户组
// @Description 创建新的用户组
// @Tags 用户组管理
// @Accept json
// @Produce json
// @Param group body services.CreateGroupRequest true "用户组信息"
// @Success 200 {object} Response{data=models.UserGroupDTO}
// @Failure 400 {object} Response
// @Failure 500 {object} Response
// @Router /api/v1/user-groups [post]
func (h *UserGroupHandler) CreateGroup(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, Response{
			Code:    401,
			Message: "未授权访问",
		})
		return
	}

	var req services.CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	group, err := h.groupService.CreateGroup(userID.(uint), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "用户组创建成功",
		Data:    group,
	})
}

// GetGroup 获取用户组信息
// @Summary 获取用户组信息
// @Description 根据用户组ID获取详细信息
// @Tags 用户组管理
// @Accept json
// @Produce json
// @Param id path int true "用户组ID"
// @Success 200 {object} Response{data=models.UserGroupDTO}
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Router /api/v1/user-groups/{id} [get]
func (h *UserGroupHandler) GetGroup(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "无效的用户组ID",
		})
		return
	}

	group, err := h.groupService.GetGroupByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, Response{
			Code:    404,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "获取成功",
		Data:    group,
	})
}

// GetGroupList 获取用户组列表
// @Summary 获取用户组列表
// @Description 分页获取用户组列表，支持关键词搜索和筛选
// @Tags 用户组管理
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param keyword query string false "搜索关键词"
// @Param is_active query bool false "是否活跃"
// @Success 200 {object} Response{data=services.GroupListResponse}
// @Failure 400 {object} Response
// @Router /api/v1/user-groups [get]
func (h *UserGroupHandler) GetGroupList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")

	var isActive *bool
	if isActiveStr := c.Query("is_active"); isActiveStr != "" {
		if isActiveBool, err := strconv.ParseBool(isActiveStr); err == nil {
			isActive = &isActiveBool
		}
	}

	result, err := h.groupService.GetGroupList(page, pageSize, keyword, isActive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "获取用户组列表失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "获取成功",
		Data:    result,
	})
}

// UpdateGroup 更新用户组信息
// @Summary 更新用户组信息
// @Description 更新用户组信息
// @Tags 用户组管理
// @Accept json
// @Produce json
// @Param id path int true "用户组ID"
// @Param group body services.UpdateGroupRequest true "更新信息"
// @Success 200 {object} Response{data=models.UserGroupDTO}
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Router /api/v1/user-groups/{id} [put]
func (h *UserGroupHandler) UpdateGroup(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "无效的用户组ID",
		})
		return
	}

	var req services.UpdateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	group, err := h.groupService.UpdateGroup(uint(id), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "用户组信息更新成功",
		Data:    group,
	})
}

// DeleteGroup 删除用户组
// @Summary 删除用户组
// @Description 删除用户组及其所有成员关系
// @Tags 用户组管理
// @Accept json
// @Produce json
// @Param id path int true "用户组ID"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Router /api/v1/user-groups/{id} [delete]
func (h *UserGroupHandler) DeleteGroup(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "无效的用户组ID",
		})
		return
	}

	if err := h.groupService.DeleteGroup(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "用户组删除成功",
	})
}

// AddMember 添加成员到用户组
// @Summary 添加成员到用户组
// @Description 将用户添加到指定用户组
// @Tags 用户组管理
// @Accept json
// @Produce json
// @Param id path int true "用户组ID"
// @Param member body services.AddMemberRequest true "成员信息"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Router /api/v1/user-groups/{id}/members [post]
func (h *UserGroupHandler) AddMember(c *gin.Context) {
	idStr := c.Param("id")
	groupID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "无效的用户组ID",
		})
		return
	}

	inviterID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, Response{
			Code:    401,
			Message: "未授权访问",
		})
		return
	}

	var req services.AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	if err := h.groupService.AddMember(uint(groupID), inviterID.(uint), &req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "成员添加成功",
	})
}

// RemoveMember 从用户组移除成员
// @Summary 从用户组移除成员
// @Description 将用户从指定用户组中移除
// @Tags 用户组管理
// @Accept json
// @Produce json
// @Param id path int true "用户组ID"
// @Param user_id path int true "用户ID"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Router /api/v1/user-groups/{id}/members/{user_id} [delete]
func (h *UserGroupHandler) RemoveMember(c *gin.Context) {
	idStr := c.Param("id")
	groupID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "无效的用户组ID",
		})
		return
	}

	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "无效的用户ID",
		})
		return
	}

	if err := h.groupService.RemoveMember(uint(groupID), uint(userID)); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "成员移除成功",
	})
}

// UpdateMember 更新用户组成员信息
// @Summary 更新用户组成员信息
// @Description 更新用户在用户组中的角色等信息
// @Tags 用户组管理
// @Accept json
// @Produce json
// @Param id path int true "用户组ID"
// @Param user_id path int true "用户ID"
// @Param member body services.UpdateMemberRequest true "成员信息"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Router /api/v1/user-groups/{id}/members/{user_id} [put]
func (h *UserGroupHandler) UpdateMember(c *gin.Context) {
	idStr := c.Param("id")
	groupID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "无效的用户组ID",
		})
		return
	}

	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "无效的用户ID",
		})
		return
	}

	var req services.UpdateMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	if err := h.groupService.UpdateMember(uint(groupID), uint(userID), &req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "成员信息更新成功",
	})
}

// GetGroupMembers 获取用户组成员列表
// @Summary 获取用户组成员列表
// @Description 分页获取指定用户组的成员列表
// @Tags 用户组管理
// @Accept json
// @Produce json
// @Param id path int true "用户组ID"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} Response{data=map[string]interface{}}
// @Failure 400 {object} Response
// @Router /api/v1/user-groups/{id}/members [get]
func (h *UserGroupHandler) GetGroupMembers(c *gin.Context) {
	idStr := c.Param("id")
	groupID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "无效的用户组ID",
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	members, total, err := h.groupService.GetGroupMembers(uint(groupID), page, pageSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	result := map[string]interface{}{
		"members":     members,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "获取成功",
		Data:    result,
	})
}

// GetUserGroups 获取用户所属的用户组列表
// @Summary 获取用户所属的用户组列表
// @Description 获取当前用户或指定用户所属的用户组列表
// @Tags 用户组管理
// @Accept json
// @Produce json
// @Param user_id query int false "用户ID（不传则获取当前用户的）"
// @Success 200 {object} Response{data=[]models.UserGroupDTO}
// @Failure 400 {object} Response
// @Router /api/v1/user-groups/my-groups [get]
func (h *UserGroupHandler) GetUserGroups(c *gin.Context) {
	var targetUserID uint

	// 如果提供了user_id参数，使用该参数（需要管理员权限）
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		userRole, _ := c.Get("user_role")
		if userRole != "admin" && userRole != "super_admin" {
			c.JSON(http.StatusForbidden, Response{
				Code:    403,
				Message: "权限不足，无法查看其他用户的用户组",
			})
			return
		}

		id, err := strconv.ParseUint(userIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, Response{
				Code:    400,
				Message: "无效的用户ID",
			})
			return
		}
		targetUserID = uint(id)
	} else {
		// 获取当前用户的用户组
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, Response{
				Code:    401,
				Message: "未授权访问",
			})
			return
		}
		targetUserID = userID.(uint)
	}

	groups, err := h.groupService.GetUserGroups(targetUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "获取用户组列表失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "获取成功",
		Data:    groups,
	})
}