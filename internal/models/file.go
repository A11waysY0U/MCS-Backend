package models

import (
	"time"
)

// File 精简版文件模型
type File struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	FileName    string     `gorm:"not null;size:255" json:"file_name"`
	FilePath    string     `gorm:"not null;size:500" json:"file_path"`
	FileSize    int64      `gorm:"not null" json:"file_size"`
	MD5Hash     string     `gorm:"size:32;index" json:"md5_hash"`
	MimeType    string     `gorm:"size:100" json:"mime_type"`
	OwnerID     uint       `gorm:"not null;index" json:"owner_id"`
	FolderID    uint       `gorm:"index" json:"folder_id"`
	WorkflowID  uint       `gorm:"index" json:"workflow_id"`
	TaskID      uint       `gorm:"index" json:"task_id"`
	IsPrivate   bool       `gorm:"default:true" json:"is_private"`
	IsDeleted   bool       `gorm:"default:false;index" json:"is_deleted"`
	DeletedAt   *time.Time `json:"deleted_at"`
	Description string     `gorm:"size:500" json:"description"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

// FileFolder 文件夹模型
type FileFolder struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	Name        string     `gorm:"not null;size:255" json:"name"`
	Path        string     `gorm:"not null;size:1000" json:"path"` // 完整路径
	ParentID    uint       `gorm:"index" json:"parent_id"`         // 父文件夹ID
	WorkflowID  uint       `gorm:"not null;index" json:"workflow_id"`
	CreatorID   uint       `gorm:"not null" json:"creator_id"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	IsDeleted   bool       `gorm:"default:false;index" json:"is_deleted"`
	DeletedAt   *time.Time `json:"deleted_at"`
	Description string     `gorm:"size:500" json:"description"`
	SortOrder   uint       `gorm:"default:0" json:"sort_order"`
}

// FileVersion 文件版本记录
type FileVersion struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	FileID    uint      `gorm:"not null;index" json:"file_id"`
	Version   uint      `gorm:"not null" json:"version"`
	FilePath  string    `gorm:"not null;size:500" json:"file_path"`
	FileSize  int64     `gorm:"not null" json:"file_size"`
	MD5Hash   string    `gorm:"size:32" json:"md5_hash"`
	CreatedBy uint      `gorm:"not null" json:"created_by"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	ChangeLog string    `gorm:"size:1000" json:"change_log"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
}

// FileShare 文件分享
type FileShare struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	FileID    uint       `gorm:"not null;index" json:"file_id"`
	ShareCode string     `gorm:"unique;not null;size:32" json:"share_code"`
	CreatorID uint       `gorm:"not null" json:"creator_id"`
	Password  string     `gorm:"size:32" json:"password"` // 分享密码
	ExpiresAt *time.Time `json:"expires_at"`
	MaxViews  uint       `gorm:"default:0" json:"max_views"` // 0表示无限制
	ViewCount uint       `gorm:"default:0" json:"view_count"`
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"created_at"`
	IsActive  bool       `gorm:"default:true" json:"is_active"`
}
