package models

import (
	"time"
)

// Notification 通知模型
type Notification struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	ReceiverID uint       `gorm:"not null;index" json:"receiver_id"`
	SenderID   *uint      `json:"sender_id"`                          // 可为空，系统通知时为空
	Type       string     `gorm:"not null;size:50;index" json:"type"` // task_assigned, task_completed, file_uploaded, etc.
	Title      string     `gorm:"not null;size:255" json:"title"`
	Content    string     `gorm:"size:1000" json:"content"`
	TargetType string     `gorm:"size:20" json:"target_type"` // task, file, workflow, user
	TargetID   *uint      `json:"target_id"`
	Data       JSONField  `gorm:"type:jsonb" json:"data"` // 额外数据
	IsRead     bool       `gorm:"default:false;index" json:"is_read"`
	ReadAt     *time.Time `json:"read_at"`
	Priority   string     `gorm:"size:10;default:'normal'" json:"priority"` // low, normal, high, urgent
	CreatedAt  time.Time  `gorm:"autoCreateTime;index" json:"created_at"`
	ExpiresAt  *time.Time `json:"expires_at"` // 通知过期时间
}

// NotificationSetting 通知设置
type NotificationSetting struct {
	ID                 uint      `gorm:"primaryKey" json:"id"`
	UserID             uint      `gorm:"not null;unique" json:"user_id"`
	TaskAssigned       bool      `gorm:"default:true" json:"task_assigned"`
	TaskCompleted      bool      `gorm:"default:true" json:"task_completed"`
	TaskOverdue        bool      `gorm:"default:true" json:"task_overdue"`
	FileUploaded       bool      `gorm:"default:true" json:"file_uploaded"`
	FileShared         bool      `gorm:"default:true" json:"file_shared"`
	WorkflowInvited    bool      `gorm:"default:true" json:"workflow_invited"`
	GroupInvited       bool      `gorm:"default:true" json:"group_invited"`
	SystemAnnouncement bool      `gorm:"default:true" json:"system_announcement"`
	EmailNotification  bool      `gorm:"default:false" json:"email_notification"`
	SMSNotification    bool      `gorm:"default:false" json:"sms_notification"`
	PushNotification   bool      `gorm:"default:true" json:"push_notification"`
	QuietHoursStart    string    `gorm:"size:5;default:'22:00'" json:"quiet_hours_start"` // 免打扰开始时间
	QuietHoursEnd      string    `gorm:"size:5;default:'08:00'" json:"quiet_hours_end"`   // 免打扰结束时间
	CreatedAt          time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
