package models

import (
	"time"

	"gorm.io/gorm"
)

// TaskSubmissionAnnotation 任务提交批注
type TaskSubmissionAnnotation struct {
	ID               uint           `json:"id" gorm:"primaryKey"`
	TaskSubmissionID uint           `json:"task_submission_id" gorm:"index;not null"`
	UserID           uint           `json:"user_id" gorm:"index;not null"`
	Content          string         `json:"content" gorm:"type:text;not null"`
	CreatedAt        time.Time      `json:"created_at"`
	DeletedAt        gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}
