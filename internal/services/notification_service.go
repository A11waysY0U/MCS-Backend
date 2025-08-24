package services

import (
	"errors"
	"fmt"
	"time"

	"mcs-backend/internal/models"

	"gorm.io/gorm"
)

// NotificationService 通知服务
type NotificationService struct {
	db *gorm.DB
}

// NewNotificationService 创建通知服务实例
func NewNotificationService(cfg interface{}) *NotificationService {
	// TODO: 从配置中获取数据库连接
	// 这里暂时返回一个空的服务实例，需要在实际使用时传入正确的数据库连接
	return &NotificationService{db: nil}
}

// 请求和响应结构体
type CreateNotificationRequest struct {
	ReceiverID uint                   `json:"receiver_id" binding:"required"`
	SenderID   *uint                  `json:"sender_id"`
	Type       string                 `json:"type" binding:"required"`
	Title      string                 `json:"title" binding:"required"`
	Content    string                 `json:"content"`
	TargetType string                 `json:"target_type"`
	TargetID   *uint                  `json:"target_id"`
	Data       map[string]interface{} `json:"data"`
	Priority   string                 `json:"priority"`
	ExpiresAt  *time.Time             `json:"expires_at"`
}

type NotificationInfo struct {
	ID           uint                   `json:"id"`
	ReceiverID   uint                   `json:"receiver_id"`
	SenderID     *uint                  `json:"sender_id"`
	SenderName   string                 `json:"sender_name"`
	Type         string                 `json:"type"`
	Title        string                 `json:"title"`
	Content      string                 `json:"content"`
	TargetType   string                 `json:"target_type"`
	TargetID     *uint                  `json:"target_id"`
	Data         map[string]interface{} `json:"data"`
	IsRead       bool                   `json:"is_read"`
	ReadAt       *time.Time             `json:"read_at"`
	Priority     string                 `json:"priority"`
	CreatedAt    time.Time              `json:"created_at"`
	ExpiresAt    *time.Time             `json:"expires_at"`
}

type NotificationListRequest struct {
	ReceiverID uint   `json:"receiver_id"`
	Page       int    `json:"page"`
	PageSize   int    `json:"page_size"`
	Type       string `json:"type"`
	IsRead     *bool  `json:"is_read"`
	Priority   string `json:"priority"`
	TargetType string `json:"target_type"`
}

type NotificationListResponse struct {
	Notifications []NotificationInfo `json:"notifications"`
	Total         int64              `json:"total"`
	UnreadCount   int64              `json:"unread_count"`
	Page          int                `json:"page"`
	PageSize      int                `json:"page_size"`
}

type UpdateNotificationSettingRequest struct {
	UserID             uint    `json:"user_id"`
	TaskAssigned       *bool   `json:"task_assigned"`
	TaskCompleted      *bool   `json:"task_completed"`
	TaskOverdue        *bool   `json:"task_overdue"`
	FileUploaded       *bool   `json:"file_uploaded"`
	FileShared         *bool   `json:"file_shared"`
	WorkflowInvited    *bool   `json:"workflow_invited"`
	GroupInvited       *bool   `json:"group_invited"`
	SystemAnnouncement *bool   `json:"system_announcement"`
	EmailNotification  *bool   `json:"email_notification"`
	SMSNotification    *bool   `json:"sms_notification"`
	PushNotification   *bool   `json:"push_notification"`
	QuietHoursStart    *string `json:"quiet_hours_start"`
	QuietHoursEnd      *string `json:"quiet_hours_end"`
}

type NotificationStats struct {
	TotalCount    int64 `json:"total_count"`
	UnreadCount   int64 `json:"unread_count"`
	TodayCount    int64 `json:"today_count"`
	HighPriority  int64 `json:"high_priority"`
	UrgentCount   int64 `json:"urgent_count"`
}

// CreateNotification 创建通知
func (s *NotificationService) CreateNotification(req *CreateNotificationRequest) (*NotificationInfo, error) {
	// 检查接收者是否存在
	var receiver models.User
	if err := s.db.Where("id = ?", req.ReceiverID).First(&receiver).Error; err != nil {
		return nil, errors.New("接收者不存在")
	}

	// 检查发送者是否存在（如果指定）
	var senderName string
	if req.SenderID != nil {
		var sender models.User
		if err := s.db.Select("username").Where("id = ?", *req.SenderID).First(&sender).Error; err != nil {
			return nil, errors.New("发送者不存在")
		}
		senderName = sender.Username
	}

	// 设置默认值
	if req.Priority == "" {
		req.Priority = "normal"
	}

	// 创建通知
	notification := models.Notification{
		ReceiverID: req.ReceiverID,
		SenderID:   req.SenderID,
		Type:       req.Type,
		Title:      req.Title,
		Content:    req.Content,
		TargetType: req.TargetType,
		TargetID:   req.TargetID,
		Data:       models.JSONField(req.Data),
		Priority:   req.Priority,
		ExpiresAt:  req.ExpiresAt,
	}

	if err := s.db.Create(&notification).Error; err != nil {
		return nil, fmt.Errorf("创建通知失败: %v", err)
	}

	return &NotificationInfo{
		ID:           notification.ID,
		ReceiverID:   notification.ReceiverID,
		SenderID:     notification.SenderID,
		SenderName:   senderName,
		Type:         notification.Type,
		Title:        notification.Title,
		Content:      notification.Content,
		TargetType:   notification.TargetType,
		TargetID:     notification.TargetID,
		Data:         map[string]interface{}(notification.Data),
		IsRead:       notification.IsRead,
		ReadAt:       notification.ReadAt,
		Priority:     notification.Priority,
		CreatedAt:    notification.CreatedAt,
		ExpiresAt:    notification.ExpiresAt,
	}, nil
}

// GetNotificationList 获取通知列表
func (s *NotificationService) GetNotificationList(req *NotificationListRequest) (*NotificationListResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	offset := (req.Page - 1) * req.PageSize

	// 构建查询条件
	query := s.db.Model(&models.Notification{}).Where("receiver_id = ?", req.ReceiverID)

	// 过滤已过期的通知
	query = query.Where("expires_at IS NULL OR expires_at > ?", time.Now())

	// 添加过滤条件
	if req.Type != "" {
		query = query.Where("type = ?", req.Type)
	}
	if req.IsRead != nil {
		query = query.Where("is_read = ?", *req.IsRead)
	}
	if req.Priority != "" {
		query = query.Where("priority = ?", req.Priority)
	}
	if req.TargetType != "" {
		query = query.Where("target_type = ?", req.TargetType)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// 获取未读数量
	var unreadCount int64
	s.db.Model(&models.Notification{}).Where("receiver_id = ? AND is_read = false AND (expires_at IS NULL OR expires_at > ?)", req.ReceiverID, time.Now()).Count(&unreadCount)

	// 获取通知列表
	var notifications []models.Notification
	if err := query.Order("priority DESC, created_at DESC").Offset(offset).Limit(req.PageSize).Find(&notifications).Error; err != nil {
		return nil, err
	}

	// 构建响应数据
	notificationInfos := make([]NotificationInfo, 0, len(notifications))
	for _, notification := range notifications {
		var senderName string
		if notification.SenderID != nil && *notification.SenderID != 0 {
			var sender models.User
			if err := s.db.Select("username").Where("id = ?", *notification.SenderID).First(&sender).Error; err == nil {
				senderName = sender.Username
			}
		}

		notificationInfos = append(notificationInfos, NotificationInfo{
			ID:           notification.ID,
			ReceiverID:   notification.ReceiverID,
			SenderID:     notification.SenderID,
			SenderName:   senderName,
			Type:         notification.Type,
			Title:        notification.Title,
			Content:      notification.Content,
			TargetType:   notification.TargetType,
			TargetID:     notification.TargetID,
			Data:         map[string]interface{}(notification.Data),
			IsRead:       notification.IsRead,
			ReadAt:       notification.ReadAt,
			Priority:     notification.Priority,
			CreatedAt:    notification.CreatedAt,
			ExpiresAt:    notification.ExpiresAt,
		})
	}

	return &NotificationListResponse{
		Notifications: notificationInfos,
		Total:         total,
		UnreadCount:   unreadCount,
		Page:          req.Page,
		PageSize:      req.PageSize,
	}, nil
}

// MarkAsRead 标记通知为已读
func (s *NotificationService) MarkAsRead(notificationID uint, userID uint) error {
	now := time.Now()
	result := s.db.Model(&models.Notification{}).Where("id = ? AND receiver_id = ?", notificationID, userID).Updates(map[string]interface{}{
		"is_read": true,
		"read_at": &now,
	})

	if result.Error != nil {
		return fmt.Errorf("标记通知已读失败: %v", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("通知不存在或无权限")
	}

	return nil
}

// MarkAllAsRead 标记所有通知为已读
func (s *NotificationService) MarkAllAsRead(userID uint) error {
	now := time.Now()
	if err := s.db.Model(&models.Notification{}).Where("receiver_id = ? AND is_read = false", userID).Updates(map[string]interface{}{
		"is_read": true,
		"read_at": &now,
	}).Error; err != nil {
		return fmt.Errorf("标记所有通知已读失败: %v", err)
	}

	return nil
}

// DeleteNotification 删除通知
func (s *NotificationService) DeleteNotification(notificationID uint, userID uint) error {
	result := s.db.Where("id = ? AND receiver_id = ?", notificationID, userID).Delete(&models.Notification{})

	if result.Error != nil {
		return fmt.Errorf("删除通知失败: %v", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("通知不存在或无权限")
	}

	return nil
}

// GetNotificationStats 获取通知统计信息
func (s *NotificationService) GetNotificationStats(userID uint) (*NotificationStats, error) {
	stats := &NotificationStats{}

	// 总通知数
	s.db.Model(&models.Notification{}).Where("receiver_id = ? AND (expires_at IS NULL OR expires_at > ?)", userID, time.Now()).Count(&stats.TotalCount)

	// 未读通知数
	s.db.Model(&models.Notification{}).Where("receiver_id = ? AND is_read = false AND (expires_at IS NULL OR expires_at > ?)", userID, time.Now()).Count(&stats.UnreadCount)

	// 今日通知数
	today := time.Now().Truncate(24 * time.Hour)
	s.db.Model(&models.Notification{}).Where("receiver_id = ? AND created_at >= ? AND (expires_at IS NULL OR expires_at > ?)", userID, today, time.Now()).Count(&stats.TodayCount)

	// 高优先级通知数
	s.db.Model(&models.Notification{}).Where("receiver_id = ? AND priority = 'high' AND is_read = false AND (expires_at IS NULL OR expires_at > ?)", userID, time.Now()).Count(&stats.HighPriority)

	// 紧急通知数
	s.db.Model(&models.Notification{}).Where("receiver_id = ? AND priority = 'urgent' AND is_read = false AND (expires_at IS NULL OR expires_at > ?)", userID, time.Now()).Count(&stats.UrgentCount)

	return stats, nil
}

// GetNotificationSetting 获取通知设置
func (s *NotificationService) GetNotificationSetting(userID uint) (*models.NotificationSetting, error) {
	var setting models.NotificationSetting
	err := s.db.Where("user_id = ?", userID).First(&setting).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 创建默认设置
		setting = models.NotificationSetting{
			UserID:             userID,
			TaskAssigned:       true,
			TaskCompleted:      true,
			TaskOverdue:        true,
			FileUploaded:       true,
			FileShared:         true,
			WorkflowInvited:    true,
			GroupInvited:       true,
			SystemAnnouncement: true,
			EmailNotification:  false,
			SMSNotification:    false,
			PushNotification:   true,
			QuietHoursStart:    "22:00",
			QuietHoursEnd:      "08:00",
		}

		if err := s.db.Create(&setting).Error; err != nil {
			return nil, fmt.Errorf("创建默认通知设置失败: %v", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("获取通知设置失败: %v", err)
	}

	return &setting, nil
}

// UpdateNotificationSetting 更新通知设置
func (s *NotificationService) UpdateNotificationSetting(req *UpdateNotificationSettingRequest) error {
	// 获取现有设置
	setting, err := s.GetNotificationSetting(req.UserID)
	if err != nil {
		return err
	}

	// 更新字段
	updates := make(map[string]interface{})
	if req.TaskAssigned != nil {
		updates["task_assigned"] = *req.TaskAssigned
	}
	if req.TaskCompleted != nil {
		updates["task_completed"] = *req.TaskCompleted
	}
	if req.TaskOverdue != nil {
		updates["task_overdue"] = *req.TaskOverdue
	}
	if req.FileUploaded != nil {
		updates["file_uploaded"] = *req.FileUploaded
	}
	if req.FileShared != nil {
		updates["file_shared"] = *req.FileShared
	}
	if req.WorkflowInvited != nil {
		updates["workflow_invited"] = *req.WorkflowInvited
	}
	if req.GroupInvited != nil {
		updates["group_invited"] = *req.GroupInvited
	}
	if req.SystemAnnouncement != nil {
		updates["system_announcement"] = *req.SystemAnnouncement
	}
	if req.EmailNotification != nil {
		updates["email_notification"] = *req.EmailNotification
	}
	if req.SMSNotification != nil {
		updates["sms_notification"] = *req.SMSNotification
	}
	if req.PushNotification != nil {
		updates["push_notification"] = *req.PushNotification
	}
	if req.QuietHoursStart != nil {
		updates["quiet_hours_start"] = *req.QuietHoursStart
	}
	if req.QuietHoursEnd != nil {
		updates["quiet_hours_end"] = *req.QuietHoursEnd
	}
	updates["updated_at"] = time.Now()

	if err := s.db.Model(setting).Updates(updates).Error; err != nil {
		return fmt.Errorf("更新通知设置失败: %v", err)
	}

	return nil
}

// CleanExpiredNotifications 清理过期通知
func (s *NotificationService) CleanExpiredNotifications() error {
	if err := s.db.Where("expires_at IS NOT NULL AND expires_at <= ?", time.Now()).Delete(&models.Notification{}).Error; err != nil {
		return fmt.Errorf("清理过期通知失败: %v", err)
	}
	return nil
}

// SendSystemNotification 发送系统通知
func (s *NotificationService) SendSystemNotification(title, content, notificationType string) error {
	// 获取所有用户ID
	var userIDs []uint
	if err := s.db.Model(&models.User{}).Pluck("id", &userIDs).Error; err != nil {
		return fmt.Errorf("获取用户列表失败: %v", err)
	}

	// 为每个用户创建通知
	for _, userID := range userIDs {
		notification := models.Notification{
			ReceiverID: userID,
			Type:       notificationType,
			Title:      title,
			Content:    content,
			Priority:   "high",
		}

		if err := s.db.Create(&notification).Error; err != nil {
			return fmt.Errorf("发送系统通知失败: %v", err)
		}
	}

	return nil
}