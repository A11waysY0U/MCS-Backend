package models

import (
	"time"

	"gorm.io/gorm"
)

// OperationLog 操作日志模型
type OperationLog struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	UserID      uint           `json:"user_id" gorm:"not null;index"`
	User        User           `json:"user" gorm:"foreignKey:UserID"`
	Action      string         `json:"action" gorm:"not null;size:100"` // 操作类型：login, upload, download, create, update, delete
	Resource    string         `json:"resource" gorm:"size:100"`        // 操作资源：file, workflow, user, etc.
	ResourceID  *uint          `json:"resource_id" gorm:"index"`        // 资源ID
	Description string         `json:"description" gorm:"size:500"`     // 操作描述
	IPAddress   string         `json:"ip_address" gorm:"size:45"`       // IP地址
	UserAgent   string         `json:"user_agent" gorm:"size:500"`      // 用户代理
	Status      string         `json:"status" gorm:"size:20;default:success"` // success, failed
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// StorageStats 存储统计模型
type StorageStats struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	Date            time.Time `json:"date" gorm:"not null;uniqueIndex:idx_date"`
	TotalFiles      int64     `json:"total_files"`      // 总文件数
	TotalSize       int64     `json:"total_size"`       // 总存储大小（字节）
	UploadedFiles   int64     `json:"uploaded_files"`   // 当日上传文件数
	UploadedSize    int64     `json:"uploaded_size"`    // 当日上传大小
	DeletedFiles    int64     `json:"deleted_files"`    // 当日删除文件数
	DeletedSize     int64     `json:"deleted_size"`     // 当日删除大小
	ActiveUsers     int64     `json:"active_users"`     // 活跃用户数
	StorageUsage    float64   `json:"storage_usage"`    // 存储使用率
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// UserActivityStats 用户活跃度统计模型
type UserActivityStats struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	UserID        uint      `json:"user_id" gorm:"not null;index"`
	User          User      `json:"user" gorm:"foreignKey:UserID"`
	Date          time.Time `json:"date" gorm:"not null;index"`
	LoginCount    int       `json:"login_count"`    // 登录次数
	UploadCount   int       `json:"upload_count"`   // 上传次数
	DownloadCount int       `json:"download_count"` // 下载次数
	FileCount     int       `json:"file_count"`     // 文件操作次数
	WorkflowCount int       `json:"workflow_count"` // 工作流操作次数
	LastActiveAt  time.Time `json:"last_active_at"` // 最后活跃时间
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// SystemStats 系统统计模型
type SystemStats struct {
	ID                uint      `json:"id" gorm:"primaryKey"`
	Date              time.Time `json:"date" gorm:"not null;uniqueIndex:idx_system_date"`
	TotalUsers        int64     `json:"total_users"`        // 总用户数
	ActiveUsers       int64     `json:"active_users"`       // 活跃用户数
	NewUsers          int64     `json:"new_users"`          // 新增用户数
	TotalFiles        int64     `json:"total_files"`        // 总文件数
	TotalWorkflows    int64     `json:"total_workflows"`    // 总工作流数
	TotalNotifications int64    `json:"total_notifications"` // 总通知数
	SystemLoad        float64   `json:"system_load"`        // 系统负载
	MemoryUsage       float64   `json:"memory_usage"`       // 内存使用率
	DiskUsage         float64   `json:"disk_usage"`         // 磁盘使用率
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// TableName 设置表名
func (OperationLog) TableName() string {
	return "operation_logs"
}

func (StorageStats) TableName() string {
	return "storage_stats"
}

func (UserActivityStats) TableName() string {
	return "user_activity_stats"
}

func (SystemStats) TableName() string {
	return "system_stats"
}