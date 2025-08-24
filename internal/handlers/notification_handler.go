package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"mcs-backend/internal/middleware"
	"mcs-backend/internal/services"
)

type NotificationHandler struct {
	notificationService *services.NotificationService
}

func NewNotificationHandler(notificationService *services.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
	}
}

// CreateNotification 创建通知
func (h *NotificationHandler) CreateNotification(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse(http.StatusUnauthorized, "未授权"))
		return
	}

	var req services.CreateNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(http.StatusBadRequest, "请求参数错误"))
		return
	}

	// 设置发送者ID
	req.SenderID = &userID

	notification, err := h.notificationService.CreateNotification(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(http.StatusInternalServerError, err.Error()))
		return
	}

	c.JSON(http.StatusCreated, SuccessResponse("通知创建成功", notification))
}

// GetNotificationList 获取通知列表
func (h *NotificationHandler) GetNotificationList(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse(http.StatusUnauthorized, "未授权"))
		return
	}

	// 解析查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	isRead := c.Query("is_read")
	notificationType := c.Query("type")

	req := services.NotificationListRequest{
		ReceiverID: userID,
		Page:       page,
		PageSize:   pageSize,
		Type:       notificationType,
	}

	// 处理已读状态过滤
	if isRead != "" {
		if isRead == "true" {
			isReadBool := true
			req.IsRead = &isReadBool
		} else if isRead == "false" {
			isReadBool := false
			req.IsRead = &isReadBool
		}
	}

	response, err := h.notificationService.GetNotificationList(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(http.StatusInternalServerError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse("获取通知列表成功", response))
}

// MarkAsRead 标记通知为已读
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse(http.StatusUnauthorized, "未授权"))
		return
	}

	notificationIDStr := c.Param("id")
	notificationID, err := strconv.ParseUint(notificationIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(http.StatusBadRequest, "无效的通知ID"))
		return
	}

	err = h.notificationService.MarkAsRead(uint(notificationID), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(http.StatusInternalServerError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse("标记已读成功", nil))
}

// MarkAllAsRead 标记所有通知为已读
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse(http.StatusUnauthorized, "未授权"))
		return
	}

	err := h.notificationService.MarkAllAsRead(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(http.StatusInternalServerError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse("标记全部已读成功", nil))
}

// DeleteNotification 删除通知
func (h *NotificationHandler) DeleteNotification(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse(http.StatusUnauthorized, "未授权"))
		return
	}

	notificationIDStr := c.Param("id")
	notificationID, err := strconv.ParseUint(notificationIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(http.StatusBadRequest, "无效的通知ID"))
		return
	}

	err = h.notificationService.DeleteNotification(uint(notificationID), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(http.StatusInternalServerError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse("删除通知成功", nil))
}

// GetNotificationStats 获取通知统计
func (h *NotificationHandler) GetNotificationStats(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse(http.StatusUnauthorized, "未授权"))
		return
	}

	stats, err := h.notificationService.GetNotificationStats(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(http.StatusInternalServerError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse("获取通知统计成功", stats))
}

// GetNotificationSetting 获取通知设置
func (h *NotificationHandler) GetNotificationSetting(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse(http.StatusUnauthorized, "未授权"))
		return
	}

	setting, err := h.notificationService.GetNotificationSetting(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(http.StatusInternalServerError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse("获取通知设置成功", setting))
}

// UpdateNotificationSetting 更新通知设置
func (h *NotificationHandler) UpdateNotificationSetting(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse(http.StatusUnauthorized, "未授权"))
		return
	}

	var req services.UpdateNotificationSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(http.StatusBadRequest, "请求参数错误"))
		return
	}

	// 设置用户ID
	req.UserID = userID

	err := h.notificationService.UpdateNotificationSetting(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(http.StatusInternalServerError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse("更新通知设置成功", nil))
}

// SendSystemNotification 发送系统通知（管理员功能）
func (h *NotificationHandler) SendSystemNotification(c *gin.Context) {
	_, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse(http.StatusUnauthorized, "未授权"))
		return
	}

	// TODO: 检查用户是否为管理员
	// 这里需要根据实际的权限系统来实现

	var req struct {
		Title   string `json:"title" binding:"required"`
		Content string `json:"content" binding:"required"`
		Type    string `json:"type"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(http.StatusBadRequest, "请求参数错误"))
		return
	}

	if req.Type == "" {
		req.Type = "system"
	}

	err := h.notificationService.SendSystemNotification(req.Title, req.Content, req.Type)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(http.StatusInternalServerError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse("系统通知发送成功", nil))
}