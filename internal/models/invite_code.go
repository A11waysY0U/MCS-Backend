package models

import "time"

// InviteCode 邀请码模型
type InviteCode struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	Code        string     `gorm:"unique;not null;size:32" json:"code"`
	CreatedBy   uint       `gorm:"not null;index" json:"created_by"`
	MaxUses     int        `gorm:"default:0" json:"max_uses"`     // 0表示无限制
	UsedCount   int        `gorm:"default:0" json:"used_count"`   // 已使用次数
	ExpiresAt   *time.Time `json:"expires_at"`                   // 过期时间，可为空表示永不过期
	Description string     `gorm:"size:255" json:"description"`   // 邀请码描述
	Status      string     `gorm:"type:varchar(20);default:'active'" json:"status"` // active, inactive, expired
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   *time.Time `gorm:"index" json:"deleted_at"`
}

// InviteCodeEnhanced 增强版邀请码模型（保持向后兼容）
type InviteCodeEnhanced struct {
	InviteCode
}

// InviteCodeUsage 邀请码使用记录
type InviteCodeUsage struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	InviteCodeID uint      `gorm:"not null;index" json:"invite_code_id"`
	UserID       uint      `gorm:"not null;index" json:"user_id"`
	UsedAt       time.Time `gorm:"autoCreateTime" json:"used_at"`
	IPAddress    string    `gorm:"size:45" json:"ip_address"` // 支持IPv6
	UserAgent    string    `gorm:"size:500" json:"user_agent"`
}
