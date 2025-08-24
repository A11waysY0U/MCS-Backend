package models

import (
	"time"
)

// UserFavorite 统一用户收藏表
type UserFavorite struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"not null;index" json:"user_id"`
	TargetType string    `gorm:"not null;size:20;index" json:"target_type"` // file, folder, workflow, task
	TargetID   uint      `gorm:"not null;index" json:"target_id"`
	FolderID   uint      `gorm:"index" json:"folder_id"` // 收藏夹ID，0表示默认收藏夹
	Note       string    `gorm:"size:255" json:"note"`
	SortOrder  uint      `gorm:"default:0" json:"sort_order"`
	CreatedAt  time.Time `gorm:"autoCreateTime;index" json:"created_at"`
}

// FavoriteFolder 收藏夹
type FavoriteFolder struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"not null;index" json:"user_id"`
	Name        string    `gorm:"not null;size:100" json:"name"`
	Description string    `gorm:"size:500" json:"description"`
	Color       string    `gorm:"size:7;default:'#409EFF'" json:"color"`
	SortOrder   uint      `gorm:"default:0" json:"sort_order"`
	IsDefault   bool      `gorm:"default:false" json:"is_default"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// Statistics 通用统计表
type Statistics struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	TargetType string     `gorm:"not null;size:20;index" json:"target_type"` // system, user, file, workflow
	TargetID   uint       `gorm:"index" json:"target_id"`                    // 0表示系统级统计
	Date       *time.Time `gorm:"index" json:"date"`                         // 统计日期，可为空表示累计统计
	Metrics    JSONField  `gorm:"type:jsonb" json:"metrics"`                 // 统计指标JSON
	CreatedAt  time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

// ActivityLog 统一活动日志表（合并AccessLog和OperationLog）
type ActivityLog struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"not null;index" json:"user_id"`
	Action     string    `gorm:"not null;size:50;index" json:"action"` // view, download, upload, edit, create_task, delete, etc.
	TargetType string    `gorm:"size:20;index" json:"target_type"`     // file, folder, workflow, task, user
	TargetID   uint      `gorm:"index" json:"target_id"`
	Details    JSONField `gorm:"type:jsonb" json:"details"` // 详细信息（包含duration, referer等）
	IPAddress  string    `gorm:"size:45" json:"ip_address"`
	UserAgent  string    `gorm:"size:500" json:"user_agent"`
	Success    bool      `gorm:"default:true;index" json:"success"`
	ErrorMsg   string    `gorm:"size:500" json:"error_msg"`
	CreatedAt  time.Time `gorm:"autoCreateTime;index" json:"created_at"`
}

// PopularityScore 热度评分
type PopularityScore struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	TargetType     string    `gorm:"size:20;index" json:"target_type"`
	TargetID       uint      `gorm:"index" json:"target_id"`
	Score          float64   `gorm:"default:0;index" json:"score"`
	ViewWeight     float64   `gorm:"default:1" json:"view_weight"`
	DownloadWeight float64   `gorm:"default:2" json:"download_weight"`
	ShareWeight    float64   `gorm:"default:3" json:"share_weight"`
	FavoriteWeight float64   `gorm:"default:5" json:"favorite_weight"`
	CalculatedAt   time.Time `gorm:"autoUpdateTime" json:"calculated_at"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// ReportData 报表数据
type ReportData struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Type        string    `gorm:"not null;size:50;index" json:"type"` // daily, weekly, monthly, yearly
	Category    string    `gorm:"size:50;index" json:"category"`      // user, file, task, storage
	StartDate   time.Time `gorm:"not null;index" json:"start_date"`
	EndDate     time.Time `gorm:"not null;index" json:"end_date"`
	Data        JSONField `gorm:"type:jsonb" json:"data"`
	GeneratedBy uint      `gorm:"not null" json:"generated_by"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// GetStorageUsagePercent 获取存储使用百分比
