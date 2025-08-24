package handlers

import (
	"net/http"
	"strconv"

	"mcs-backend/internal/services"
	"mcs-backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

// WorkflowHandler 工作流处理器
type WorkflowHandler struct {
	workflowService *services.WorkflowService
}

// NewWorkflowHandler 创建工作流处理器
func NewWorkflowHandler(workflowService *services.WorkflowService) *WorkflowHandler {
	return &WorkflowHandler{
		workflowService: workflowService,
	}
}

// CreateWorkflow 创建工作流
// @Summary 创建工作流
// @Description 创建新的工作流
// @Tags 工作流管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param request body services.CreateWorkflowRequest true "创建工作流请求"
// @Success 200 {object} Response{data=services.WorkflowInfo} "创建成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 401 {object} Response "未授权"
// @Failure 500 {object} Response "服务器错误"
// @Router /api/workflows [post]
func (h *WorkflowHandler) CreateWorkflow(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse(http.StatusUnauthorized, "未授权"))
		return
	}

	var req services.CreateWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(http.StatusBadRequest, "请求参数错误: "+err.Error()))
		return
	}

	workflow, err := h.workflowService.CreateWorkflow(&req, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(http.StatusBadRequest, err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse("创建工作流成功", workflow))
}

// GetWorkflow 获取工作流详情
// @Summary 获取工作流详情
// @Description 根据ID获取工作流详细信息
// @Tags 工作流管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path int true "工作流ID"
// @Success 200 {object} Response{data=services.WorkflowInfo} "获取成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 401 {object} Response "未授权"
// @Failure 404 {object} Response "工作流不存在"
// @Router /api/workflows/{id} [get]
func (h *WorkflowHandler) GetWorkflow(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse(http.StatusUnauthorized, "未授权"))
		return
	}

	workflowIDStr := c.Param("id")
	workflowID, err := strconv.ParseUint(workflowIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(http.StatusBadRequest, "无效的工作流ID"))
		return
	}

	workflow, err := h.workflowService.GetWorkflowByID(uint(workflowID), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse(http.StatusNotFound, err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse("获取工作流成功", workflow))
}

// GetWorkflowList 获取工作流列表
// @Summary 获取工作流列表
// @Description 获取用户参与的工作流列表，支持分页和筛选
// @Tags 工作流管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param status query string false "工作流状态"
// @Param keyword query string false "关键词搜索"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param sort_by query string false "排序字段" default(created_at)
// @Param sort_order query string false "排序方向" default(desc)
// @Success 200 {object} Response{data=services.WorkflowListResponse} "获取成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 401 {object} Response "未授权"
// @Router /api/workflows [get]
func (h *WorkflowHandler) GetWorkflowList(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse(http.StatusUnauthorized, "未授权"))
		return
	}

	// 解析查询参数
	page := 1
	pageSize := 10

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			pageSize = ps
		}
	}

	workflows, err := h.workflowService.GetWorkflowList(userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(http.StatusBadRequest, err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse("获取工作流列表成功", workflows))
}

// UpdateWorkflow 更新工作流
// @Summary 更新工作流
// @Description 更新工作流信息
// @Tags 工作流管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path int true "工作流ID"
// @Param request body services.UpdateWorkflowRequest true "更新工作流请求"
// @Success 200 {object} Response{data=services.WorkflowInfo} "更新成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 401 {object} Response "未授权"
// @Failure 404 {object} Response "工作流不存在"
// @Router /api/workflows/{id} [put]
func (h *WorkflowHandler) UpdateWorkflow(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse(http.StatusUnauthorized, "未授权"))
		return
	}

	workflowIDStr := c.Param("id")
	workflowID, err := strconv.ParseUint(workflowIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(http.StatusBadRequest, "无效的工作流ID"))
		return
	}

	var req services.UpdateWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(http.StatusBadRequest, "请求参数错误: "+err.Error()))
		return
	}

	workflow, err := h.workflowService.UpdateWorkflow(uint(workflowID), &req, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(http.StatusBadRequest, err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse("更新工作流成功", workflow))
}

// DeleteWorkflow 删除工作流
// @Summary 删除工作流
// @Description 删除工作流（软删除）
// @Tags 工作流管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path int true "工作流ID"
// @Success 200 {object} Response "删除成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 401 {object} Response "未授权"
// @Failure 404 {object} Response "工作流不存在"
// @Router /api/workflows/{id} [delete]
func (h *WorkflowHandler) DeleteWorkflow(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse(http.StatusUnauthorized, "未授权"))
		return
	}

	workflowIDStr := c.Param("id")
	workflowID, err := strconv.ParseUint(workflowIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(http.StatusBadRequest, "无效的工作流ID"))
		return
	}

	if err := h.workflowService.DeleteWorkflow(uint(workflowID), userID); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(http.StatusBadRequest, err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse("删除工作流成功", nil))
}

// AddMember 添加工作流成员
// @Summary 添加工作流成员
// @Description 向工作流添加成员
// @Tags 工作流管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path int true "工作流ID"
// @Param request body services.AddWorkflowMemberRequest true "添加成员请求"
// @Success 200 {object} Response{data=services.WorkflowMemberInfo} "添加成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 401 {object} Response "未授权"
// @Failure 404 {object} Response "工作流不存在"
// @Router /api/workflows/{id}/members [post]
func (h *WorkflowHandler) AddMember(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse(http.StatusUnauthorized, "未授权"))
		return
	}

	workflowIDStr := c.Param("id")
	workflowID, err := strconv.ParseUint(workflowIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(http.StatusBadRequest, "无效的工作流ID"))
		return
	}

	var req services.AddWorkflowMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(http.StatusBadRequest, "请求参数错误: "+err.Error()))
		return
	}

	member, err := h.workflowService.AddMember(uint(workflowID), &req, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(http.StatusBadRequest, err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse("添加成员成功", member))
}

// RemoveMember 移除工作流成员
// @Summary 移除工作流成员
// @Description 从工作流中移除成员
// @Tags 工作流管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path int true "工作流ID"
// @Param member_id path int true "成员ID"
// @Success 200 {object} Response "移除成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 401 {object} Response "未授权"
// @Failure 404 {object} Response "工作流或成员不存在"
// @Router /api/workflows/{id}/members/{member_id} [delete]
func (h *WorkflowHandler) RemoveMember(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse(http.StatusUnauthorized, "未授权"))
		return
	}

	workflowIDStr := c.Param("id")
	workflowID, err := strconv.ParseUint(workflowIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(http.StatusBadRequest, "无效的工作流ID"))
		return
	}

	memberIDStr := c.Param("member_id")
	memberID, err := strconv.ParseUint(memberIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(http.StatusBadRequest, "无效的成员ID"))
		return
	}

	if err := h.workflowService.RemoveMember(uint(workflowID), uint(memberID), userID); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(http.StatusBadRequest, err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse("移除成员成功", nil))
}

// UpdateMemberRole 更新成员角色
// @Summary 更新成员角色
// @Description 更新工作流成员的角色
// @Tags 工作流管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path int true "工作流ID"
// @Param member_id path int true "成员ID"
// @Param request body services.UpdateWorkflowMemberRequest true "更新角色请求"
// @Success 200 {object} Response{data=services.WorkflowMemberInfo} "更新成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 401 {object} Response "未授权"
// @Failure 404 {object} Response "工作流或成员不存在"
// @Router /api/workflows/{id}/members/{member_id}/role [put]
func (h *WorkflowHandler) UpdateMemberRole(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse(http.StatusUnauthorized, "未授权"))
		return
	}

	workflowIDStr := c.Param("id")
	workflowID, err := strconv.ParseUint(workflowIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(http.StatusBadRequest, "无效的工作流ID"))
		return
	}

	memberIDStr := c.Param("member_id")
	memberID, err := strconv.ParseUint(memberIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(http.StatusBadRequest, "无效的成员ID"))
		return
	}

	var req services.UpdateWorkflowMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(http.StatusBadRequest, "请求参数错误: "+err.Error()))
		return
	}

	member, err := h.workflowService.UpdateMemberRole(uint(workflowID), uint(memberID), &req, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(http.StatusBadRequest, err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse("更新成员角色成功", member))
}

// GetMembers 获取工作流成员列表
// @Summary 获取工作流成员列表
// @Description 获取工作流的成员列表
// @Tags 工作流管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path int true "工作流ID"
// @Param role query string false "角色筛选"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} Response{data=services.WorkflowMemberListResponse} "获取成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 401 {object} Response "未授权"
// @Failure 404 {object} Response "工作流不存在"
// @Router /api/workflows/{id}/members [get]
func (h *WorkflowHandler) GetMembers(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse(http.StatusUnauthorized, "未授权"))
		return
	}

	workflowIDStr := c.Param("id")
	workflowID, err := strconv.ParseUint(workflowIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(http.StatusBadRequest, "无效的工作流ID"))
		return
	}

	members, err := h.workflowService.GetWorkflowMembers(uint(workflowID), userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(http.StatusBadRequest, err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse("获取成员列表成功", members))
}