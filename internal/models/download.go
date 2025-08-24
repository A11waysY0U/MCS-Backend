package models

import (
	"time"

	"gorm.io/gorm"
)

// DownloadTask 下载任务模型
type DownloadTask struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	UserID      uint           `json:"user_id" gorm:"not null;index"`
	TaskName    string         `json:"task_name" gorm:"size:255;not null"`
	FileIDs     string         `json:"file_ids" gorm:"type:text;not null"` // JSON数组存储文件ID列表
	Status      string         `json:"status" gorm:"size:50;not null;default:'pending'"` // pending, processing, completed, failed
	ZipFilePath string         `json:"zip_file_path" gorm:"size:500"`
	FileSize    int64          `json:"file_size" gorm:"default:0"`
	DownloadURL string         `json:"download_url" gorm:"size:500"`
	ExpiresAt   *time.Time     `json:"expires_at"`
	ErrorMsg    string         `json:"error_msg" gorm:"type:text"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	User User `json:"user" gorm:"foreignKey:UserID"`
}

// DownloadLog 下载日志模型
type DownloadLog struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	UserID     uint      `json:"user_id" gorm:"not null;index"`
	FileID     uint      `json:"file_id" gorm:"not null;index"`
	FileName   string    `json:"file_name" gorm:"size:255;not null"`
	FileSize   int64     `json:"file_size" gorm:"not null"`
	IPAddress  string    `json:"ip_address" gorm:"size:45"`
	UserAgent  string    `json:"user_agent" gorm:"type:text"`
	DownloadAt time.Time `json:"download_at"`

	// 关联
	User User `json:"user" gorm:"foreignKey:UserID"`
	File File `json:"file" gorm:"foreignKey:FileID"`
}

// BatchDownloadRequest 批量下载请求
type BatchDownloadRequest struct {
	FileIDs  []uint `json:"file_ids" binding:"required,min=1"`
	TaskName string `json:"task_name" binding:"required,max=255"`
}

// DownloadTaskInfo 下载任务信息
type DownloadTaskInfo struct {
	ID          uint       `json:"id"`
	TaskName    string     `json:"task_name"`
	Status      string     `json:"status"`
	FileCount   int        `json:"file_count"`
	FileSize    int64      `json:"file_size"`
	DownloadURL string     `json:"download_url,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	ErrorMsg    string     `json:"error_msg,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// DownloadStatsResponse 下载统计响应
type DownloadStatsResponse struct {
	TotalDownloads     int64 `json:"total_downloads"`
	TodayDownloads     int64 `json:"today_downloads"`
	WeekDownloads      int64 `json:"week_downloads"`
	MonthDownloads     int64 `json:"month_downloads"`
	TotalDownloadSize  int64 `json:"total_download_size"`
	TodayDownloadSize  int64 `json:"today_download_size"`
	WeekDownloadSize   int64 `json:"week_download_size"`
	MonthDownloadSize  int64 `json:"month_download_size"`
	PopularFiles       []PopularFileInfo `json:"popular_files"`
}

// PopularFileInfo 热门文件信息
type PopularFileInfo struct {
	FileID       uint   `json:"file_id"`
	FileName     string `json:"file_name"`
	DownloadCount int64  `json:"download_count"`
	FileSize     int64  `json:"file_size"`
}