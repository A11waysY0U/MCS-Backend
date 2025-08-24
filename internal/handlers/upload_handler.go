package handlers

import (
	"io"
	"mcs-backend/internal/config"
	"mcs-backend/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// UploadHandler 文件上传处理器
type UploadHandler struct {
	uploadService *services.UploadService
}

// NewUploadHandler 创建文件上传处理器
func NewUploadHandler(cfg *config.Config) *UploadHandler {
	return &UploadHandler{
		uploadService: services.NewUploadService(cfg),
	}
}

// InitUpload 初始化上传
// @Summary 初始化文件上传
// @Description 初始化文件上传，支持秒传检测
// @Tags 文件上传
// @Accept json
// @Produce json
// @Param request body services.InitUploadRequest true "初始化上传请求"
// @Success 200 {object} Response{data=services.InitUploadResponse} "初始化成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 500 {object} Response "服务器错误"
// @Router /api/upload/init [post]
// @Security BearerAuth
func (h *UploadHandler) InitUpload(c *gin.Context) {
	var req services.InitUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(400, "请求参数错误: "+err.Error()))
		return
	}

	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse(401, "未授权访问"))
		return
	}

	response, err := h.uploadService.InitUpload(&req, userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(500, "初始化上传失败: "+err.Error()))
		return
	}

	if response.IsSecUpload {
		c.JSON(http.StatusOK, SuccessResponse("文件秒传成功", response))
	} else {
		c.JSON(http.StatusOK, SuccessResponse("初始化上传成功", response))
	}
}

// UploadChunk 上传分片
// @Summary 上传文件分片
// @Description 上传文件分片数据
// @Tags 文件上传
// @Accept multipart/form-data
// @Produce json
// @Param upload_id formData string true "上传ID"
// @Param chunk_index formData int true "分片索引"
// @Param chunk_md5 formData string true "分片MD5"
// @Param chunk formData file true "分片文件"
// @Success 200 {object} Response "上传成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 500 {object} Response "服务器错误"
// @Router /api/upload/chunk [post]
// @Security BearerAuth
func (h *UploadHandler) UploadChunk(c *gin.Context) {
	uploadID := c.PostForm("upload_id")
	if uploadID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse(400, "上传ID不能为空"))
		return
	}

	chunkIndexStr := c.PostForm("chunk_index")
	chunkIndex, err := strconv.Atoi(chunkIndexStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(400, "分片索引格式错误"))
		return
	}

	chunkMD5 := c.PostForm("chunk_md5")
	if chunkMD5 == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse(400, "分片MD5不能为空"))
		return
	}

	// 获取上传的文件
	file, err := c.FormFile("chunk")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(400, "获取分片文件失败: "+err.Error()))
		return
	}

	// 读取文件内容
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(500, "打开分片文件失败: "+err.Error()))
		return
	}
	defer src.Close()

	chunkData, err := io.ReadAll(src)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(500, "读取分片文件失败: "+err.Error()))
		return
	}

	// 上传分片
	if err := h.uploadService.UploadChunk(uploadID, chunkIndex, chunkData, chunkMD5); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(500, "上传分片失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse("分片上传成功", nil))
}

// CompleteUpload 完成上传
// @Summary 完成文件上传
// @Description 合并所有分片，完成文件上传
// @Tags 文件上传
// @Accept json
// @Produce json
// @Param upload_id path string true "上传ID"
// @Success 200 {object} Response{data=services.File} "上传完成"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 500 {object} Response "服务器错误"
// @Router /api/upload/complete/{upload_id} [post]
// @Security BearerAuth
func (h *UploadHandler) CompleteUpload(c *gin.Context) {
	uploadID := c.Param("upload_id")
	if uploadID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse(400, "上传ID不能为空"))
		return
	}

	file, err := h.uploadService.CompleteUpload(uploadID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(500, "完成上传失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse("文件上传完成", file))
}

// GetUploadProgress 获取上传进度
// @Summary 获取上传进度
// @Description 获取指定上传任务的进度信息
// @Tags 文件上传
// @Accept json
// @Produce json
// @Param upload_id path string true "上传ID"
// @Success 200 {object} Response{data=services.UploadProgress} "获取成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 404 {object} Response "上传任务不存在"
// @Router /api/upload/progress/{upload_id} [get]
// @Security BearerAuth
func (h *UploadHandler) GetUploadProgress(c *gin.Context) {
	uploadID := c.Param("upload_id")
	if uploadID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse(400, "上传ID不能为空"))
		return
	}

	progress, err := h.uploadService.GetUploadProgress(uploadID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse(404, "获取上传进度失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse("获取上传进度成功", progress))
}

// CancelUpload 取消上传
// @Summary 取消文件上传
// @Description 取消指定的上传任务并清理临时文件
// @Tags 文件上传
// @Accept json
// @Produce json
// @Param upload_id path string true "上传ID"
// @Success 200 {object} Response "取消成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 404 {object} Response "上传任务不存在"
// @Router /api/upload/cancel/{upload_id} [delete]
// @Security BearerAuth
func (h *UploadHandler) CancelUpload(c *gin.Context) {
	uploadID := c.Param("upload_id")
	if uploadID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse(400, "上传ID不能为空"))
		return
	}

	if err := h.uploadService.CancelUpload(uploadID); err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse(404, "取消上传失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse("取消上传成功", nil))
}