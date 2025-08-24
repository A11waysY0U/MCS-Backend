package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"mcs-backend/internal/middleware"
	"mcs-backend/internal/services"
)

type StatisticsHandler struct {
	statisticsService *services.StatisticsService
}

func NewStatisticsHandler(statisticsService *services.StatisticsService) *StatisticsHandler {
	return &StatisticsHandler{
		statisticsService: statisticsService,
	}
}

// GetOperationLogs 获取操作日志
func (h *StatisticsHandler) GetOperationLogs(c *gin.Context) {
	_, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse(http.StatusUnauthorized, "未授权"))
		return
	}

	// TODO: 检查用户是否有查看日志的权限

	// 解析查询参数
	var req services.OperationLogRequest
	req.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	req.PageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "20"))
	req.Action = c.Query("action")
	req.Resource = c.Query("resource")
	req.Status = c.Query("status")

	// 解析用户ID
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if userID, err := strconv.ParseUint(userIDStr, 10, 32); err == nil {
			userIDUint := uint(userID)
			req.UserID = &userIDUint
		}
	}

	// 解析日期范围
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			req.StartDate = &startDate
		}
	}
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			// 设置为当天结束时间
			endOfDay := endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			req.EndDate = &endOfDay
		}
	}

	response, err := h.statisticsService.GetOperationLogs(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(http.StatusInternalServerError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse("获取操作日志成功", response))
}

// GetStorageStats 获取存储统计
func (h *StatisticsHandler) GetStorageStats(c *gin.Context) {
	_, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse(http.StatusUnauthorized, "未授权"))
		return
	}

	// TODO: 检查用户是否有查看统计的权限

	var req services.StorageStatsRequest
	req.GroupBy = c.DefaultQuery("group_by", "day")

	// 解析日期范围
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			req.StartDate = &startDate
		}
	}
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			req.EndDate = &endDate
		}
	}

	// 如果没有指定日期范围，默认获取最近30天的数据
	if req.StartDate == nil && req.EndDate == nil {
		now := time.Now()
		startDate := now.AddDate(0, 0, -30)
		req.StartDate = &startDate
		req.EndDate = &now
	}

	stats, err := h.statisticsService.GetStorageStats(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(http.StatusInternalServerError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse("获取存储统计成功", stats))
}

// GetUserActivityStats 获取用户活跃度统计
func (h *StatisticsHandler) GetUserActivityStats(c *gin.Context) {
	_, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse(http.StatusUnauthorized, "未授权"))
		return
	}

	// TODO: 检查用户是否有查看统计的权限

	var req services.UserActivityRequest
	req.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	req.PageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// 解析用户ID
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if userID, err := strconv.ParseUint(userIDStr, 10, 32); err == nil {
			userIDUint := uint(userID)
			req.UserID = &userIDUint
		}
	}

	// 解析日期范围
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			req.StartDate = &startDate
		}
	}
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			req.EndDate = &endDate
		}
	}

	// 如果没有指定日期范围，默认获取最近7天的数据
	if req.StartDate == nil && req.EndDate == nil {
		now := time.Now()
		startDate := now.AddDate(0, 0, -7)
		req.StartDate = &startDate
		req.EndDate = &now
	}

	response, err := h.statisticsService.GetUserActivityStats(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(http.StatusInternalServerError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse("获取用户活跃度统计成功", response))
}

// GetSystemOverview 获取系统概览
func (h *StatisticsHandler) GetSystemOverview(c *gin.Context) {
	_, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse(http.StatusUnauthorized, "未授权"))
		return
	}

	// TODO: 检查用户是否有查看统计的权限

	overview, err := h.statisticsService.GetSystemOverview()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(http.StatusInternalServerError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse("获取系统概览成功", overview))
}

// GetOperationStats 获取操作统计
func (h *StatisticsHandler) GetOperationStats(c *gin.Context) {
	_, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse(http.StatusUnauthorized, "未授权"))
		return
	}

	// TODO: 检查用户是否有查看统计的权限

	// 解析日期范围
	var startDate, endDate *time.Time
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = &parsed
		}
	}
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", endDateStr); err == nil {
			// 设置为当天结束时间
			endOfDay := parsed.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			endDate = &endOfDay
		}
	}

	// 如果没有指定日期范围，默认获取最近30天的数据
	if startDate == nil && endDate == nil {
		now := time.Now()
		start := now.AddDate(0, 0, -30)
		startDate = &start
		endDate = &now
	}

	stats, err := h.statisticsService.GetOperationStats(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(http.StatusInternalServerError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse("获取操作统计成功", stats))
}

// GetTopActiveUsers 获取最活跃用户
func (h *StatisticsHandler) GetTopActiveUsers(c *gin.Context) {
	_, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse(http.StatusUnauthorized, "未授权"))
		return
	}

	// TODO: 检查用户是否有查看统计的权限

	// 解析限制数量
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	// 解析日期范围
	var startDate, endDate *time.Time
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = &parsed
		}
	}
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", endDateStr); err == nil {
			// 设置为当天结束时间
			endOfDay := parsed.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			endDate = &endOfDay
		}
	}

	// 如果没有指定日期范围，默认获取最近30天的数据
	if startDate == nil && endDate == nil {
		now := time.Now()
		start := now.AddDate(0, 0, -30)
		startDate = &start
		endDate = &now
	}

	users, err := h.statisticsService.GetTopActiveUsers(startDate, endDate, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(http.StatusInternalServerError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse("获取最活跃用户成功", users))
}

// UpdateStorageStats 手动更新存储统计（管理员功能）
func (h *StatisticsHandler) UpdateStorageStats(c *gin.Context) {
	_, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse(http.StatusUnauthorized, "未授权"))
		return
	}

	// TODO: 检查用户是否为管理员

	err := h.statisticsService.UpdateStorageStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(http.StatusInternalServerError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse("更新存储统计成功", nil))
}

// UpdateUserActivityStats 手动更新用户活跃度统计（管理员功能）
func (h *StatisticsHandler) UpdateUserActivityStats(c *gin.Context) {
	_, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse(http.StatusUnauthorized, "未授权"))
		return
	}

	// TODO: 检查用户是否为管理员

	err := h.statisticsService.UpdateUserActivityStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(http.StatusInternalServerError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse("更新用户活跃度统计成功", nil))
}