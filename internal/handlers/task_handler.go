package handlers

import (
	"net/http"
	"strconv"

	"mcs-backend/internal/middleware"
	"mcs-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// TaskHandler 任务处理器
type TaskHandler struct {
	taskService *services.TaskService
}

// NewTaskHandler 创建任务处理器
func NewTaskHandler(taskService *services.TaskService) *TaskHandler {
	return &TaskHandler{
		taskService: taskService,
	}
}

// CreateTask 创建任务
// @Summary 创建任务
// @Description 创建新的任务
// @Tags 任务管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param request body services.CreateTaskRequest true "创建任务请求"
// @Success 200 {object} Response{data=services.TaskInfo} "创建成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 401 {object} Response "未授权"
// @Failure 500 {object} Response "服务器错误"
// @Router /api/tasks [post]
func (h *TaskHandler) CreateTask(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse(http.StatusUnauthorized, "未授权"))
		return
	}

	var req services.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(http.StatusBadRequest, "请求参数错误: "+err.Error()))
		return
	}

	task, err := h.taskService.CreateTask(&req, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(http.StatusBadRequest, err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse("创建任务成功", task))
}

// GetTask 获取任务详情
// @Summary 获取任务详情
// @Description 根据ID获取任务详细信息
// @Tags 任务管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path int true "任务ID"
// @Success 200 {object} Response{data=services.TaskInfo} "获取成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 401 {object} Response "未授权"
// @Failure 404 {object} Response "任务不存在"
// @Router /api/tasks/{id} [get]
func (h *TaskHandler) GetTask(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse(http.StatusUnauthorized, "未授权"))
		return
	}

	taskIDStr := c.Param("id")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(http.StatusBadRequest, "无效的任务ID"))
		return
	}

	task, err := h.taskService.GetTaskByID(uint(taskID), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse(http.StatusNotFound, err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse("获取任务成功", task))
}

// GetTaskList 获取任务列表
// @Summary 获取任务列表
// @Description 获取任务列表，支持分页和筛选
// @Tags 任务管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param workflow_id query int false "工作流ID"
// @Param status query string false "任务状态"
// @Param priority query string false "优先级"
// @Param responsible_id query int false "负责人ID"
// @Param creator_id query int false "创建者ID"
// @Param keyword query string false "关键词搜索"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param sort_by query string false "排序字段" default(created_at)
// @Param sort_order query string false "排序方向" default(desc)
// @Success 200 {object} Response{data=services.TaskListResponse} "获取成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 401 {object} Response "未授权"
// @Router /api/tasks [get]
func (h *TaskHandler) GetTaskList(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse(http.StatusUnauthorized, "未授权"))
		return
	}

	// 解析查询参数
	req := services.TaskListRequest{
		Page:     1,
		PageSize: 10,
		SortBy:   "created_at",
		SortOrder: "desc",
	}

	if workflowIDStr := c.Query("workflow_id"); workflowIDStr != "" {
		if workflowID, err := strconv.ParseUint(workflowIDStr, 10, 32); err == nil {
			req.WorkflowID = uint(workflowID)
		}
	}

	if status := c.Query("status"); status != "" {
		req.Status = status
	}

	if priority := c.Query("priority"); priority != "" {
		req.Priority = priority
	}

	if responsibleIDStr := c.Query("responsible_id"); responsibleIDStr != "" {
		if responsibleID, err := strconv.ParseUint(responsibleIDStr, 10, 32); err == nil {
			req.ResponsibleID = uint(responsibleID)
		}
	}

	if creatorIDStr := c.Query("creator_id"); creatorIDStr != "" {
		if creatorID, err := strconv.ParseUint(creatorIDStr, 10, 32); err == nil {
			req.CreatorID = uint(creatorID)
		}
	}

	if keyword := c.Query("keyword"); keyword != "" {
		req.Keyword = keyword
	}

	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			req.Page = page
		}
	}

	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 {
			req.PageSize = pageSize
		}
	}

	if sortBy := c.Query("sort_by"); sortBy != "" {
		req.SortBy = sortBy
	}

	if sortOrder := c.Query("sort_order"); sortOrder != "" {
		req.SortOrder = sortOrder
	}

	tasks, err := h.taskService.GetTaskList(&req, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(http.StatusBadRequest, err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse("获取任务列表成功", tasks))
}

// UpdateTask 更新任务
// @Summary 更新任务
// @Description 更新任务信息
// @Tags 任务管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path int true "任务ID"
// @Param request body services.UpdateTaskRequest true "更新任务请求"
// @Success 200 {object} Response{data=services.TaskInfo} "更新成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 401 {object} Response "未授权"
// @Failure 404 {object} Response "任务不存在"
// @Router /api/tasks/{id} [put]
func (h *TaskHandler) UpdateTask(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse(http.StatusUnauthorized, "未授权"))
		return
	}

	taskIDStr := c.Param("id")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(http.StatusBadRequest, "无效的任务ID"))
		return
	}

	var req services.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(http.StatusBadRequest, "请求参数错误: "+err.Error()))
		return
	}

	task, err := h.taskService.UpdateTask(uint(taskID), &req, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(http.StatusBadRequest, err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse("更新任务成功", task))
}

// DeleteTask 删除任务
// @Summary 删除任务
// @Description 删除任务（软删除）
// @Tags 任务管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path int true "任务ID"
// @Success 200 {object} Response "删除成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 401 {object} Response "未授权"
// @Failure 404 {object} Response "任务不存在"
// @Router /api/tasks/{id} [delete]
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse(http.StatusUnauthorized, "未授权"))
		return
	}

	taskIDStr := c.Param("id")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(http.StatusBadRequest, "无效的任务ID"))
		return
	}

	if err := h.taskService.DeleteTask(uint(taskID), userID); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(http.StatusBadRequest, err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse("删除任务成功", nil))
}

// ChangeTaskStatus 更改任务状态
// @Summary 更改任务状态
// @Description 更改任务状态（开始、提交审核、完成、取消等）
// @Tags 任务管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path int true "任务ID"
// @Param request body services.ChangeStatusRequest true "更改状态请求"
// @Success 200 {object} Response{data=services.TaskInfo} "状态更改成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 401 {object} Response "未授权"
// @Failure 404 {object} Response "任务不存在"
// @Router /api/tasks/{id}/status [put]
func (h *TaskHandler) ChangeTaskStatus(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse(http.StatusUnauthorized, "未授权"))
		return
	}

	taskIDStr := c.Param("id")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(http.StatusBadRequest, "无效的任务ID"))
		return
	}

	var req services.ChangeStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(http.StatusBadRequest, "请求参数错误: "+err.Error()))
		return
	}

	task, err := h.taskService.ChangeTaskStatus(uint(taskID), &req, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(http.StatusBadRequest, err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse("状态更改成功", task))
}

// AddToStagingArea 添加文件到暂存区
// @Summary 添加文件到暂存区
// @Description 将文件添加到任务的暂存区
// @Tags 任务管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path int true "任务ID"
// @Param request body services.StagingAreaRequest true "暂存区请求"
// @Success 200 {object} Response{data=services.StagingAreaInfo} "添加成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 401 {object} Response "未授权"
// @Failure 404 {object} Response "任务不存在"
// @Router /api/tasks/{id}/staging [post]
func (h *TaskHandler) AddToStagingArea(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse(http.StatusUnauthorized, "未授权"))
		return
	}

	taskIDStr := c.Param("id")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(http.StatusBadRequest, "无效的任务ID"))
		return
	}

	var req services.StagingAreaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(http.StatusBadRequest, "请求参数错误: "+err.Error()))
		return
	}

	staging, err := h.taskService.AddToStagingArea(uint(taskID), &req, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(http.StatusBadRequest, err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse("添加到暂存区成功", staging))
}

// GetStagingArea 获取任务暂存区
// @Summary 获取任务暂存区
// @Description 获取任务的暂存区文件列表
// @Tags 任务管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path int true "任务ID"
// @Success 200 {object} Response{data=[]services.StagingAreaInfo} "获取成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 401 {object} Response "未授权"
// @Failure 404 {object} Response "任务不存在"
// @Router /api/tasks/{id}/staging [get]
func (h *TaskHandler) GetStagingArea(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse(http.StatusUnauthorized, "未授权"))
		return
	}

	taskIDStr := c.Param("id")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(http.StatusBadRequest, "无效的任务ID"))
		return
	}

	staging, err := h.taskService.GetStagingArea(uint(taskID), userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(http.StatusBadRequest, err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse("获取暂存区成功", staging))
}

// SubmitStagingArea 提交暂存区
// @Summary 提交暂存区
// @Description 提交任务的暂存区文件
// @Tags 任务管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path int true "任务ID"
// @Success 200 {object} Response "提交成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 401 {object} Response "未授权"
// @Failure 404 {object} Response "任务不存在"
// @Router /api/tasks/{id}/staging/submit [post]
func (h *TaskHandler) SubmitStagingArea(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse(http.StatusUnauthorized, "未授权"))
		return
	}

	taskIDStr := c.Param("id")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(http.StatusBadRequest, "无效的任务ID"))
		return
	}

	if err := h.taskService.SubmitStagingArea(uint(taskID), userID); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(http.StatusBadRequest, err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse("提交暂存区成功", nil))
}

// ClearStagingArea 清空暂存区
// @Summary 清空暂存区
// @Description 清空任务的暂存区文件
// @Tags 任务管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path int true "任务ID"
// @Success 200 {object} Response "清空成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 401 {object} Response "未授权"
// @Failure 404 {object} Response "任务不存在"
// @Router /api/tasks/{id}/staging/clear [delete]
func (h *TaskHandler) ClearStagingArea(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse(http.StatusUnauthorized, "未授权"))
		return
	}

	taskIDStr := c.Param("id")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(http.StatusBadRequest, "无效的任务ID"))
		return
	}

	if err := h.taskService.ClearStagingArea(uint(taskID), userID); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(http.StatusBadRequest, err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse("清空暂存区成功", nil))
}