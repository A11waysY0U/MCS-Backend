package models

import (
	"time"
)

type User struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	Username     string     `gorm:"unique;not null;size:50" json:"username"`
	Email        string     `gorm:"unique;not null;size:100" json:"email"`
	PasswordHash string     `gorm:"not null;column:password_hash" json:"-"`
	RealName     string     `gorm:"size:100" json:"real_name"`
	Role         string     `gorm:"type:varchar(20);default:'user'" json:"role"`
	Status       string     `gorm:"type:varchar(20);default:'active'" json:"status"`
	InviteCodeID *uint      `gorm:"index" json:"invite_code_id"`
	LastLoginAt  *time.Time `json:"last_login_at"`
	CreatedAt    time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt    *time.Time `gorm:"index" json:"deleted_at"`
}

type UserDTO struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	RealName  string    `json:"real_name"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// UserInfo 用户信息结构（用于认证响应）
type UserInfo struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	RealName string `json:"real_name"`
	Role     string `json:"role"`
	Status   string `json:"status"`
}

func (u *User) ToUserDTO() *UserDTO {
	return &UserDTO{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		RealName:  u.RealName,
		Role:      u.Role,
		Status:    u.Status,
		CreatedAt: u.CreatedAt,
	}
}

func (u *User) ToUserInfo() *UserInfo {
	return &UserInfo{
		ID:       u.ID,
		Username: u.Username,
		Email:    u.Email,
		RealName: u.RealName,
		Role:     u.Role,
		Status:   u.Status,
	}
}
