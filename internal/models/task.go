package models

import (
	"time"
)

// Task 任务模型
type Task struct {
	ID             uint        `gorm:"primaryKey" json:"id"`
	Name           string      `gorm:"not null;size:255" json:"name"`
	Description    string      `gorm:"size:1000" json:"description"`
	WorkflowID     uint        `gorm:"not null;index" json:"workflow_id"`
	CreatorID      uint        `gorm:"not null" json:"creator_id"`
	ResponsibleID  uint        `gorm:"not null;index" json:"responsible_id"`
	Status         string      `gorm:"size:20;default:'pending';index" json:"status"`
	Priority       string      `gorm:"size:10;default:'medium'" json:"priority"`
	StartDate      *time.Time  `json:"start_date"`
	DueDate        *time.Time  `json:"due_date"`
	CompletedAt    *time.Time  `json:"completed_at"`
	CreatedAt      time.Time   `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time   `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt      *time.Time  `gorm:"index" json:"deleted_at"`
	Tags           StringArray `gorm:"type:jsonb" json:"tags"`
}

// TaskEnhanced 增强版任务模型（保持向后兼容）
type TaskEnhanced struct {
	Task
	ID             uint        `gorm:"primaryKey" json:"id"`
	Name           string      `gorm:"not null;size:255" json:"name"`
	Description    string      `gorm:"size:1000" json:"description"`
	WorkflowID     uint        `gorm:"not null;index" json:"workflow_id"`
	CreatorID      uint        `gorm:"not null" json:"creator_id"`
	ResponsibleID  uint        `gorm:"not null;index" json:"responsible_id"`          // 负责人ID
	Status         string      `gorm:"size:20;default:'pending';index" json:"status"` // pending, in_progress, review, completed
	Priority       string      `gorm:"size:10;default:'medium'" json:"priority"`      // low, medium, high, urgent
	RequireReview  bool        `gorm:"default:false" json:"require_review"`
	ReviewerID     uint        `json:"reviewer_id"` // 审核人ID
	StartDate      *time.Time  `json:"start_date"`
	DueDate        *time.Time  `json:"due_date"`
	CompletedAt    *time.Time  `json:"completed_at"`
	CreatedAt      time.Time   `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time   `gorm:"autoUpdateTime" json:"updated_at"`
	EstimatedHours float64     `json:"estimated_hours"`
	ActualHours    float64     `json:"actual_hours"`
	Progress       uint        `gorm:"default:0;check:progress <= 100" json:"progress"` // 0-100
	IsDeleted      bool        `gorm:"default:false;index" json:"is_deleted"`
	DeletedAt      *time.Time  `json:"deleted_at"`
	Tags           StringArray `gorm:"type:jsonb" json:"tags"` // 任务标签
}

// TaskMember 任务成员模型
type TaskMember struct {
	ID       uint      `gorm:"primaryKey" json:"id"`
	TaskID   uint      `gorm:"not null;index" json:"task_id"`
	UserID   uint      `gorm:"not null;index" json:"user_id"`
	Role     string    `gorm:"size:20;default:'member'" json:"role"`
	JoinedAt time.Time `gorm:"autoCreateTime" json:"joined_at"`
}

// TaskMemberEnhanced 任务成员增强版模型（保持向后兼容）
type TaskMemberEnhanced struct {
	TaskMember
	ID       uint       `gorm:"primaryKey" json:"id"`
	TaskID   uint       `gorm:"not null;index" json:"task_id"`
	UserID   uint       `gorm:"not null;index" json:"user_id"`
	Role     string     `gorm:"not null;size:20" json:"role"` // responsible, collaborator, reviewer
	JoinedAt time.Time  `gorm:"autoCreateTime" json:"joined_at"`
	LeftAt   *time.Time `json:"left_at"`
	IsActive bool       `gorm:"default:true" json:"is_active"`
}

// TaskStatusLog 任务状态变更日志
type TaskStatusLog struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	TaskID     uint      `gorm:"not null;index" json:"task_id"`
	FromStatus string    `gorm:"size:20" json:"from_status"`
	ToStatus   string    `gorm:"size:20;not null" json:"to_status"`
	OperatorID uint      `gorm:"not null" json:"operator_id"`
	Remark     string    `gorm:"size:500" json:"remark"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TaskStagingArea 任务暂存区
type TaskStagingArea struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	TaskID      uint       `gorm:"not null;index" json:"task_id"`
	UserID      uint       `gorm:"not null;index" json:"user_id"`
	FileID      uint       `gorm:"not null;index" json:"file_id"`
	Operation   string     `gorm:"size:20;not null" json:"operation"` // add, update, delete
	Version     uint       `gorm:"default:1" json:"version"`
	IsSubmitted bool       `gorm:"default:false" json:"is_submitted"`
	SubmittedAt *time.Time `json:"submitted_at"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	Remark      string     `gorm:"size:500" json:"remark"`
}

// TaskSubmission 任务提交记录
type TaskSubmission struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	TaskID      uint       `gorm:"not null;index" json:"task_id"`
	SubmitterID uint       `gorm:"not null" json:"submitter_id"`
	Version     uint       `gorm:"not null" json:"version"`
	Description string     `gorm:"size:1000" json:"description"`
	FileCount   uint       `gorm:"default:0" json:"file_count"`
	Status      string     `gorm:"size:20;default:'pending'" json:"status"` // pending, approved, rejected
	ReviewerID  uint       `json:"reviewer_id"`
	ReviewAt    *time.Time `json:"review_at"`
	ReviewNote  string     `gorm:"size:1000" json:"review_note"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

// TaskTemplate 任务模板
type TaskTemplate struct {
	ID             uint        `gorm:"primaryKey" json:"id"`
	Name           string      `gorm:"not null;size:255" json:"name"`
	Description    string      `gorm:"size:1000" json:"description"`
	CreatorID      uint        `gorm:"not null" json:"creator_id"`
	RequireReview  bool        `gorm:"default:false" json:"require_review"`
	EstimatedHours float64     `json:"estimated_hours"`
	Tags           StringArray `gorm:"type:jsonb" json:"tags"`
	IsPublic       bool        `gorm:"default:false" json:"is_public"`
	UsageCount     uint        `gorm:"default:0" json:"usage_count"`
	CreatedAt      time.Time   `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time   `gorm:"autoUpdateTime" json:"updated_at"`
}
