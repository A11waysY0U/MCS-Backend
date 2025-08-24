package handlers

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"mcs-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// FileHandler 文件管理处理器
type FileHandler struct {
	fileService *services.FileService
}

// NewFileHandler 创建文件管理处理器
func NewFileHandler(fileService *services.FileService) *FileHandler {
	return &FileHandler{
		fileService: fileService,
	}
}

// CreateFolder 创建文件夹
// @Summary 创建文件夹
// @Description 在指定工作流下创建文件夹
// @Tags 文件管理
// @Accept json
// @Produce json
// @Param request body services.CreateFolderRequest true "创建文件夹请求"
// @Success 200 {object} Response{data=services.FolderInfo} "创建成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 500 {object} Response "服务器错误"
// @Router /api/files/folders [post]
// @Security BearerAuth
func (h *FileHandler) CreateFolder(c *gin.Context) {
	var req services.CreateFolderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(400, "请求参数错误: "+err.Error()))
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse(401, "未授权访问"))
		return
	}

	folder, err := h.fileService.CreateFolder(&req, userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(500, "创建文件夹失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse("创建文件夹成功", folder))
}

// GetFileList 获取文件列表
// @Summary 获取文件列表
// @Description 获取指定条件下的文件和文件夹列表
// @Tags 文件管理
// @Accept json
// @Produce json
// @Param folder_id query int false "文件夹ID"
// @Param workflow_id query int false "工作流ID"
// @Param task_id query int false "任务ID"
// @Param keyword query string false "搜索关键词"
// @Param mime_type query string false "文件类型"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param sort_by query string false "排序字段" default(created_at)
// @Param sort_order query string false "排序方向" default(desc)
// @Success 200 {object} Response{data=services.FileListResponse} "获取成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 500 {object} Response "服务器错误"
// @Router /api/files [get]
// @Security BearerAuth
func (h *FileHandler) GetFileList(c *gin.Context) {
	req := services.FileListRequest{
		Keyword:   c.Query("keyword"),
		MimeType:  c.Query("mime_type"),
		SortBy:    c.DefaultQuery("sort_by", "created_at"),
		SortOrder: c.DefaultQuery("sort_order", "desc"),
	}

	// 解析查询参数
	if folderIDStr := c.Query("folder_id"); folderIDStr != "" {
		if folderID, err := strconv.ParseUint(folderIDStr, 10, 32); err == nil {
			req.FolderID = uint(folderID)
		}
	}
	if workflowIDStr := c.Query("workflow_id"); workflowIDStr != "" {
		if workflowID, err := strconv.ParseUint(workflowIDStr, 10, 32); err == nil {
			req.WorkflowID = uint(workflowID)
		}
	}
	if taskIDStr := c.Query("task_id"); taskIDStr != "" {
		if taskID, err := strconv.ParseUint(taskIDStr, 10, 32); err == nil {
			req.TaskID = uint(taskID)
		}
	}
	if pageStr := c.DefaultQuery("page", "1"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil {
			req.Page = page
		}
	}
	if pageSizeStr := c.DefaultQuery("page_size", "20"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil {
			req.PageSize = pageSize
		}
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse(401, "未授权访问"))
		return
	}

	response, err := h.fileService.GetFileList(&req, userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(500, "获取文件列表失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse("获取文件列表成功", response))
}

// GetFile 获取文件信息
// @Summary 获取文件信息
// @Description 根据文件ID获取详细信息
// @Tags 文件管理
// @Accept json
// @Produce json
// @Param id path int true "文件ID"
// @Success 200 {object} Response{data=services.FileInfo} "获取成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 404 {object} Response "文件不存在"
// @Router /api/files/{id} [get]
// @Security BearerAuth
func (h *FileHandler) GetFile(c *gin.Context) {
	idStr := c.Param("id")
	fileID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(400, "文件ID格式错误"))
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse(401, "未授权访问"))
		return
	}

	file, err := h.fileService.GetFileByID(uint(fileID), userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse(404, "获取文件信息失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse("获取文件信息成功", file))
}

// UpdateFile 更新文件信息
// @Summary 更新文件信息
// @Description 更新文件的基本信息和标签
// @Tags 文件管理
// @Accept json
// @Produce json
// @Param id path int true "文件ID"
// @Param request body services.UpdateFileRequest true "更新文件请求"
// @Success 200 {object} Response{data=services.FileInfo} "更新成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 404 {object} Response "文件不存在"
// @Router /api/files/{id} [put]
// @Security BearerAuth
func (h *FileHandler) UpdateFile(c *gin.Context) {
	idStr := c.Param("id")
	fileID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(400, "文件ID格式错误"))
		return
	}

	var req services.UpdateFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(400, "请求参数错误: "+err.Error()))
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse(401, "未授权访问"))
		return
	}

	file, err := h.fileService.UpdateFile(uint(fileID), &req, userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(500, "更新文件信息失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse("更新文件信息成功", file))
}

// DeleteFile 删除文件
// @Summary 删除文件
// @Description 软删除指定的文件
// @Tags 文件管理
// @Accept json
// @Produce json
// @Param id path int true "文件ID"
// @Success 200 {object} Response "删除成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 404 {object} Response "文件不存在"
// @Router /api/files/{id} [delete]
// @Security BearerAuth
func (h *FileHandler) DeleteFile(c *gin.Context) {
	idStr := c.Param("id")
	fileID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(400, "文件ID格式错误"))
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse(401, "未授权访问"))
		return
	}

	if err := h.fileService.DeleteFile(uint(fileID), userID.(uint)); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(500, "删除文件失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse("删除文件成功", nil))
}

// UpdateFolder 更新文件夹
// @Summary 更新文件夹
// @Description 更新文件夹的基本信息
// @Tags 文件管理
// @Accept json
// @Produce json
// @Param id path int true "文件夹ID"
// @Param request body services.UpdateFolderRequest true "更新文件夹请求"
// @Success 200 {object} Response{data=services.FolderInfo} "更新成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 404 {object} Response "文件夹不存在"
// @Router /api/files/folders/{id} [put]
// @Security BearerAuth
func (h *FileHandler) UpdateFolder(c *gin.Context) {
	idStr := c.Param("id")
	folderID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(400, "文件夹ID格式错误"))
		return
	}

	var req services.UpdateFolderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(400, "请求参数错误: "+err.Error()))
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse(401, "未授权访问"))
		return
	}

	folder, err := h.fileService.UpdateFolder(uint(folderID), &req, userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(500, "更新文件夹失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse("更新文件夹成功", folder))
}

// DeleteFolder 删除文件夹
// @Summary 删除文件夹
// @Description 删除空的文件夹
// @Tags 文件管理
// @Accept json
// @Produce json
// @Param id path int true "文件夹ID"
// @Success 200 {object} Response "删除成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 404 {object} Response "文件夹不存在"
// @Router /api/files/folders/{id} [delete]
// @Security BearerAuth
func (h *FileHandler) DeleteFolder(c *gin.Context) {
	idStr := c.Param("id")
	folderID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(400, "文件夹ID格式错误"))
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse(401, "未授权访问"))
		return
	}

	if err := h.fileService.DeleteFolder(uint(folderID), userID.(uint)); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(500, "删除文件夹失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse("删除文件夹成功", nil))
}

// SearchFiles 搜索文件
// @Summary 搜索文件
// @Description 根据关键词搜索文件
// @Tags 文件管理
// @Accept json
// @Produce json
// @Param keyword query string true "搜索关键词"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} Response{data=services.FileListResponse} "搜索成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 500 {object} Response "服务器错误"
// @Router /api/files/search [get]
// @Security BearerAuth
func (h *FileHandler) SearchFiles(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse(400, "搜索关键词不能为空"))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse(401, "未授权访问"))
		return
	}

	response, err := h.fileService.SearchFiles(keyword, userID.(uint), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(500, "搜索文件失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse("搜索文件成功", response))
}

// GetFileVersions 获取文件版本列表
// @Summary 获取文件版本列表
// @Description 获取指定文件的所有版本
// @Tags 文件管理
// @Accept json
// @Produce json
// @Param id path int true "文件ID"
// @Success 200 {object} Response{data=[]models.FileVersion} "获取成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 404 {object} Response "文件不存在"
// @Router /api/files/{id}/versions [get]
// @Security BearerAuth
func (h *FileHandler) GetFileVersions(c *gin.Context) {
	idStr := c.Param("id")
	fileID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(400, "文件ID格式错误"))
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse(401, "未授权访问"))
		return
	}

	versions, err := h.fileService.GetFileVersions(uint(fileID), userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(500, "获取文件版本失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse("获取文件版本成功", versions))
}

// DownloadFile 下载文件
// @Summary 下载文件
// @Description 下载指定文件
// @Tags 文件管理
// @Accept json
// @Produce application/octet-stream
// @Param id path string true "文件ID"
// @Success 200 {file} binary "文件内容"
// @Failure 400 {object} handlers.Response "请求参数错误"
// @Failure 404 {object} handlers.Response "文件不存在"
// @Failure 500 {object} handlers.Response "服务器内部错误"
// @Router /api/v1/files/{id}/download [get]
// @Security ApiKeyAuth
func (h *FileHandler) DownloadFile(c *gin.Context) {
	fileIDStr := c.Param("id")
	if fileIDStr == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse(400, "文件ID不能为空"))
		return
	}

	// 转换文件ID
	fileID, err := strconv.ParseUint(fileIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(400, "无效的文件ID"))
		return
	}

	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse(401, "未授权访问"))
		return
	}

	// 获取文件信息
	fileInfo, err := h.fileService.GetFileByID(uint(fileID), userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse(404, "文件不存在"))
		return
	}

	// 构建文件路径（需要从数据库获取实际的文件路径）
	filePath := fmt.Sprintf("%s/%s", h.fileService.GetConfig().File.UploadPath, fileInfo.FileName)

	// 检查文件是否存在于磁盘
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, ErrorResponse(404, "文件不存在于服务器"))
		return
	}

	// 设置响应头
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileInfo.FileName))
	c.Header("Content-Type", "application/octet-stream")

	// 发送文件
	c.File(filePath)
}