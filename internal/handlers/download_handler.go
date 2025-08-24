package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"mcs-backend/internal/middleware"
	"mcs-backend/internal/models"
	"mcs-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// DownloadHandler 下载处理器
type DownloadHandler struct {
	downloadService *services.DownloadService
}

// NewDownloadHandler 创建下载处理器
func NewDownloadHandler(downloadService *services.DownloadService) *DownloadHandler {
	return &DownloadHandler{
		downloadService: downloadService,
	}
}

// CreateBatchDownload 创建批量下载任务
// @Summary 创建批量下载任务
// @Description 创建批量下载任务，将多个文件打包成ZIP
// @Tags 下载管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param request body models.BatchDownloadRequest true "批量下载请求"
// @Success 200 {object} Response{data=models.DownloadTaskInfo} "创建成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 401 {object} Response "未授权"
// @Failure 500 {object} Response "服务器错误"
// @Router /api/v1/download/batch [post]
func (h *DownloadHandler) CreateBatchDownload(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse(http.StatusUnauthorized, "未授权"))
		return
	}

	var req models.BatchDownloadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(http.StatusBadRequest, "请求参数错误: "+err.Error()))
		return
	}

	taskInfo, err := h.downloadService.CreateBatchDownloadTask(userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(http.StatusBadRequest, err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse("批量下载任务创建成功", taskInfo))
}

// GetDownloadTask 获取下载任务信息
// @Summary 获取下载任务信息
// @Description 获取指定下载任务的详细信息
// @Tags 下载管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path int true "任务ID"
// @Success 200 {object} Response{data=models.DownloadTaskInfo} "获取成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 401 {object} Response "未授权"
// @Failure 404 {object} Response "任务不存在"
// @Router /api/v1/download/tasks/{id} [get]
func (h *DownloadHandler) GetDownloadTask(c *gin.Context) {
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

	taskInfo, err := h.downloadService.GetDownloadTask(userID, uint(taskID))
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse(http.StatusNotFound, err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse("获取下载任务成功", taskInfo))
}

// GetDownloadTasks 获取用户下载任务列表
// @Summary 获取用户下载任务列表
// @Description 获取当前用户的下载任务列表，支持分页
// @Tags 下载管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} Response{data=object} "获取成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 401 {object} Response "未授权"
// @Router /api/v1/download/tasks [get]
func (h *DownloadHandler) GetDownloadTasks(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse(http.StatusUnauthorized, "未授权"))
		return
	}

	// 解析分页参数
	page := 1
	pageSize := 10

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	tasks, total, err := h.downloadService.GetUserDownloadTasks(userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(http.StatusInternalServerError, "获取下载任务列表失败"))
		return
	}

	response := gin.H{
		"tasks": tasks,
		"pagination": gin.H{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	}

	c.JSON(http.StatusOK, SuccessResponse("获取下载任务列表成功", response))
}

// DownloadZipFile 下载ZIP文件
// @Summary 下载ZIP文件
// @Description 下载打包好的ZIP文件
// @Tags 下载管理
// @Accept json
// @Produce application/octet-stream
// @Param Authorization header string true "Bearer token"
// @Param filename path string true "文件名"
// @Success 200 {file} binary "文件内容"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 401 {object} Response "未授权"
// @Failure 404 {object} Response "文件不存在"
// @Router /api/v1/download/zip/{filename} [get]
func (h *DownloadHandler) DownloadZipFile(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse(http.StatusUnauthorized, "未授权"))
		return
	}

	fileName := c.Param("filename")
	if fileName == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse(http.StatusBadRequest, "文件名不能为空"))
		return
	}

	filePath, err := h.downloadService.DownloadZipFile(userID, fileName)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse(http.StatusNotFound, err.Error()))
		return
	}

	// 设置响应头
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Header("Content-Type", "application/zip")

	// 发送文件
	c.File(filePath)
}

// DownloadSingleFile 下载单个文件
// @Summary 下载单个文件
// @Description 下载指定的单个文件
// @Tags 下载管理
// @Accept json
// @Produce application/octet-stream
// @Param Authorization header string true "Bearer token"
// @Param id path int true "文件ID"
// @Success 200 {file} binary "文件内容"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 401 {object} Response "未授权"
// @Failure 404 {object} Response "文件不存在"
// @Router /api/v1/download/file/{id} [get]
func (h *DownloadHandler) DownloadSingleFile(c *gin.Context) {
	_, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse(http.StatusUnauthorized, "未授权"))
		return
	}

	fileIDStr := c.Param("id")
	fileID, err := strconv.ParseUint(fileIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(http.StatusBadRequest, "无效的文件ID"))
		return
	}

	// 返回提示信息
	c.JSON(http.StatusOK, gin.H{
		"message": "请使用 `/api/v1/files/{id}/download` 接口下载单个文件",
		"redirect_url": fmt.Sprintf("/api/v1/files/%d/download", fileID),
	})
}

// GetDownloadStats 获取下载统计
// @Summary 获取下载统计
// @Description 获取下载统计信息，包括下载次数、大小等
// @Tags 下载管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} Response{data=models.DownloadStatsResponse} "获取成功"
// @Failure 401 {object} Response "未授权"
// @Failure 500 {object} Response "服务器错误"
// @Router /api/v1/download/stats [get]
func (h *DownloadHandler) GetDownloadStats(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse(http.StatusUnauthorized, "未授权"))
		return
	}

	stats, err := h.downloadService.GetDownloadStats(&userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(http.StatusInternalServerError, "获取下载统计失败"))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse("获取下载统计成功", stats))
}

// GetGlobalDownloadStats 获取全局下载统计（管理员）
// @Summary 获取全局下载统计
// @Description 获取全局下载统计信息，仅管理员可访问
// @Tags 下载管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} Response{data=models.DownloadStatsResponse} "获取成功"
// @Failure 401 {object} Response "未授权"
// @Failure 403 {object} Response "权限不足"
// @Failure 500 {object} Response "服务器错误"
// @Router /api/v1/download/stats/global [get]
func (h *DownloadHandler) GetGlobalDownloadStats(c *gin.Context) {
	stats, err := h.downloadService.GetDownloadStats(nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(http.StatusInternalServerError, "获取全局下载统计失败"))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse("获取全局下载统计成功", stats))
}