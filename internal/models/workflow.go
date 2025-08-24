package models

import "time"

type Workflow struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"unique" json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	MasterID    uint      `gorm:"not null" json:"master_id"`
}

type WorkflowMember struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	WorkflowID uint      `gorm:"index;not null" json:"workflow_id"`
	UserID     uint      `gorm:"index;not null" json:"user_id"`
	Role       string    `gorm:"default:'normal'" json:"role"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
