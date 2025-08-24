package models

import "time"

// UserGroup 用户组模型
type UserGroup struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"unique;not null;size:100" json:"name"`
	Description string    `gorm:"size:500" json:"description"`
	Color       string    `gorm:"size:7;default:'#409EFF'" json:"color"` // 用户组颜色标识
	CreaterID   uint      `gorm:"not null" json:"creater_id"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	SortOrder   uint      `gorm:"default:0" json:"sort_order"` // 排序字段
}

// UserGroupMember 用户-用户组关联表
type UserGroupMember struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	GroupID   uint      `gorm:"not null;index" json:"group_id"`
	Role      string    `gorm:"size:20;default:'member'" json:"role"` // member, admin
	JoinedAt  time.Time `gorm:"autoCreateTime" json:"joined_at"`
	InviterID uint      `json:"inviter_id"` // 邀请人ID
	IsActive  bool      `gorm:"default:true" json:"is_active"`
}

// UserGroupInvitation 用户组邀请
type UserGroupInvitation struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	GroupID   uint      `gorm:"not null;index" json:"group_id"`
	InviterID uint      `gorm:"not null" json:"inviter_id"`
	InviteeID uint      `gorm:"not null;index" json:"invitee_id"`
	Status    uint      `gorm:"not null;check:status IN (0,1,2);default:0" json:"status"` // 0:待处理 1:已接受 2:已拒绝
	Message   string    `gorm:"size:255" json:"message"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	ExpiresAt time.Time `json:"expires_at"` // 邀请过期时间
}

// UserGroupDTO 用户组DTO
type UserGroupDTO struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
	MemberCount int64  `json:"member_count"`
	IsActive    bool   `json:"is_active"`
}

// ToUserGroupDTO 转换为DTO
func (ug *UserGroup) ToUserGroupDTO() *UserGroupDTO {
	return &UserGroupDTO{
		ID:          ug.ID,
		Name:        ug.Name,
		Description: ug.Description,
		Color:       ug.Color,
		IsActive:    ug.IsActive,
	}
}
