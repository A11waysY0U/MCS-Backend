package services

import (
	"errors"
	"fmt"
	"time"

	"mcs-backend/internal/config"
	"mcs-backend/internal/database"
	"mcs-backend/internal/models"

	"gorm.io/gorm"
)

// TaskService 任务服务
type TaskService struct {
	db     *gorm.DB
	config *config.Config
}

// NewTaskService 创建任务服务
func NewTaskService(cfg *config.Config) *TaskService {
	return &TaskService{
		db:     database.GetDB(),
		config: cfg,
	}
}

// GetConfig 获取配置
func (s *TaskService) GetConfig() *config.Config {
	return s.config
}

// CreateTaskRequest 创建任务请求
type CreateTaskRequest struct {
	Name           string     `json:"name" binding:"required"`
	Description    string     `json:"description"`
	WorkflowID     uint       `json:"workflow_id" binding:"required"`
	ResponsibleID  uint       `json:"responsible_id" binding:"required"`
	Priority       string     `json:"priority"`
	RequireReview  bool       `json:"require_review"`
	ReviewerID     uint       `json:"reviewer_id"`
	StartDate      *time.Time `json:"start_date"`
	DueDate        *time.Time `json:"due_date"`
	EstimatedHours float64    `json:"estimated_hours"`
	Tags           []string   `json:"tags"`
}

// UpdateTaskRequest 更新任务请求
type UpdateTaskRequest struct {
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	ResponsibleID  uint       `json:"responsible_id"`
	Priority       string     `json:"priority"`
	RequireReview  bool       `json:"require_review"`
	ReviewerID     uint       `json:"reviewer_id"`
	StartDate      *time.Time `json:"start_date"`
	DueDate        *time.Time `json:"due_date"`
	EstimatedHours float64    `json:"estimated_hours"`
	ActualHours    float64    `json:"actual_hours"`
	Progress       uint       `json:"progress"`
	Tags           []string   `json:"tags"`
}

// TaskListRequest 任务列表请求
type TaskListRequest struct {
	WorkflowID    uint     `json:"workflow_id"`
	Status        string   `json:"status"`
	Priority      string   `json:"priority"`
	ResponsibleID uint     `json:"responsible_id"`
	CreatorID     uint     `json:"creator_id"`
	Keyword       string   `json:"keyword"`
	Tags          []string `json:"tags"`
	Page          int      `json:"page"`
	PageSize      int      `json:"page_size"`
	SortBy        string   `json:"sort_by"`    // created_at, due_date, priority
	SortOrder     string   `json:"sort_order"` // asc, desc
}

// TaskListResponse 任务列表响应
type TaskListResponse struct {
	Tasks    []TaskInfo `json:"tasks"`
	Total    int64      `json:"total"`
	Page     int        `json:"page"`
	PageSize int        `json:"page_size"`
}

// TaskInfo 任务信息
type TaskInfo struct {
	ID               uint       `json:"id"`
	Name             string     `json:"name"`
	Description      string     `json:"description"`
	WorkflowID       uint       `json:"workflow_id"`
	WorkflowName     string     `json:"workflow_name"`
	CreatorID        uint       `json:"creator_id"`
	CreatorName      string     `json:"creator_name"`
	ResponsibleID    uint       `json:"responsible_id"`
	ResponsibleName  string     `json:"responsible_name"`
	Status           string     `json:"status"`
	Priority         string     `json:"priority"`
	RequireReview    bool       `json:"require_review"`
	ReviewerID       uint       `json:"reviewer_id"`
	ReviewerName     string     `json:"reviewer_name"`
	StartDate        *time.Time `json:"start_date"`
	DueDate          *time.Time `json:"due_date"`
	CompletedAt      *time.Time `json:"completed_at"`
	EstimatedHours   float64    `json:"estimated_hours"`
	ActualHours      float64    `json:"actual_hours"`
	Progress         uint       `json:"progress"`
	Tags             []string   `json:"tags"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// ChangeStatusRequest 更改状态请求
type ChangeStatusRequest struct {
	Status string `json:"status" binding:"required"`
	Remark string `json:"remark"`
}

// StagingAreaRequest 暂存区请求
type StagingAreaRequest struct {
	FileID    uint   `json:"file_id" binding:"required"`
	Operation string `json:"operation" binding:"required"` // add, update, delete
	Remark    string `json:"remark"`
}

// StagingAreaInfo 暂存区信息
type StagingAreaInfo struct {
	ID          uint      `json:"id"`
	TaskID      uint      `json:"task_id"`
	UserID      uint      `json:"user_id"`
	UserName    string    `json:"user_name"`
	FileID      uint      `json:"file_id"`
	FileName    string    `json:"file_name"`
	Operation   string    `json:"operation"`
	Version     uint      `json:"version"`
	IsSubmitted bool      `json:"is_submitted"`
	SubmittedAt *time.Time `json:"submitted_at"`
	Remark      string    `json:"remark"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateTask 创建任务
func (s *TaskService) CreateTask(req *CreateTaskRequest, userID uint) (*TaskInfo, error) {
	// 检查用户是否有权限在该工作流中创建任务
	var member models.WorkflowMember
	if err := s.db.Where("workflow_id = ? AND user_id = ?", req.WorkflowID, userID).First(&member).Error; err != nil {
		return nil, errors.New("无权限在该工作流中创建任务")
	}

	// 检查负责人是否为工作流成员
	var responsibleMember models.WorkflowMember
	if err := s.db.Where("workflow_id = ? AND user_id = ?", req.WorkflowID, req.ResponsibleID).First(&responsibleMember).Error; err != nil {
		return nil, errors.New("负责人不是工作流成员")
	}

	// 如果需要审核，检查审核人是否为工作流成员
	if req.RequireReview && req.ReviewerID > 0 {
		var reviewerMember models.WorkflowMember
		if err := s.db.Where("workflow_id = ? AND user_id = ?", req.WorkflowID, req.ReviewerID).First(&reviewerMember).Error; err != nil {
			return nil, errors.New("审核人不是工作流成员")
		}
	}

	// 设置默认值
	if req.Priority == "" {
		req.Priority = "medium"
	}

	// 创建任务
	task := models.TaskEnhanced{
		Task: models.Task{
			Name:          req.Name,
			Description:   req.Description,
			WorkflowID:    req.WorkflowID,
			CreatorID:     userID,
			ResponsibleID: req.ResponsibleID,
			Status:        "pending",
			Priority:      req.Priority,
			StartDate:     req.StartDate,
			DueDate:       req.DueDate,
			Tags:          models.StringArray(req.Tags),
		},
		RequireReview:  req.RequireReview,
		ReviewerID:     req.ReviewerID,
		EstimatedHours: req.EstimatedHours,
		Progress:       0,
	}

	if err := s.db.Create(&task).Error; err != nil {
		return nil, fmt.Errorf("创建任务失败: %v", err)
	}

	// 记录状态变更日志
	statusLog := models.TaskStatusLog{
		TaskID:     task.ID,
		FromStatus: "",
		ToStatus:   "pending",
		OperatorID: userID,
		Remark:     "任务创建",
	}
	s.db.Create(&statusLog)

	// 获取任务详细信息
	return s.GetTaskByID(task.ID, userID)
}

// GetTaskByID 根据ID获取任务
func (s *TaskService) GetTaskByID(taskID uint, userID uint) (*TaskInfo, error) {
	// 获取任务信息
	var task models.TaskEnhanced
	if err := s.db.Where("id = ? AND is_deleted = false", taskID).First(&task).Error; err != nil {
		return nil, errors.New("任务不存在")
	}

	// 检查用户是否有权限访问该任务
	var member models.WorkflowMember
	if err := s.db.Where("workflow_id = ? AND user_id = ?", task.WorkflowID, userID).First(&member).Error; err != nil {
		return nil, errors.New("无权限访问该任务")
	}

	// 获取相关用户信息
	var creator, responsible, reviewer models.User
	var workflow models.Workflow

	s.db.Select("username").Where("id = ?", task.CreatorID).First(&creator)
	s.db.Select("username").Where("id = ?", task.ResponsibleID).First(&responsible)
	s.db.Select("name").Where("id = ?", task.WorkflowID).First(&workflow)

	taskInfo := &TaskInfo{
		ID:               task.ID,
		Name:             task.Name,
		Description:      task.Description,
		WorkflowID:       task.WorkflowID,
		WorkflowName:     workflow.Name,
		CreatorID:        task.CreatorID,
		CreatorName:      creator.Username,
		ResponsibleID:    task.ResponsibleID,
		ResponsibleName:  responsible.Username,
		Status:           task.Status,
		Priority:         task.Priority,
		RequireReview:    task.RequireReview,
		ReviewerID:       task.ReviewerID,
		StartDate:        task.StartDate,
		DueDate:          task.DueDate,
		CompletedAt:      task.CompletedAt,
		EstimatedHours:   task.EstimatedHours,
		ActualHours:      task.ActualHours,
		Progress:         task.Progress,
		Tags:             []string(task.Tags),
		CreatedAt:        task.CreatedAt,
		UpdatedAt:        task.UpdatedAt,
	}

	if task.ReviewerID > 0 {
		s.db.Select("username").Where("id = ?", task.ReviewerID).First(&reviewer)
		taskInfo.ReviewerName = reviewer.Username
	}

	return taskInfo, nil
}

// GetTaskList 获取任务列表
func (s *TaskService) GetTaskList(req *TaskListRequest, userID uint) (*TaskListResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	offset := (req.Page - 1) * req.PageSize

	// 构建查询条件
	query := s.db.Model(&models.TaskEnhanced{}).Where("is_deleted = false")

	// 如果指定了工作流ID，检查用户权限
	if req.WorkflowID > 0 {
		var member models.WorkflowMember
		if err := s.db.Where("workflow_id = ? AND user_id = ?", req.WorkflowID, userID).First(&member).Error; err != nil {
			return nil, errors.New("无权限访问该工作流的任务")
		}
		query = query.Where("workflow_id = ?", req.WorkflowID)
	} else {
		// 获取用户参与的工作流ID列表
		var memberWorkflowIDs []uint
		if err := s.db.Model(&models.WorkflowMember{}).Where("user_id = ?", userID).Pluck("workflow_id", &memberWorkflowIDs).Error; err != nil {
			return nil, fmt.Errorf("获取用户工作流失败: %v", err)
		}
		if len(memberWorkflowIDs) == 0 {
			return &TaskListResponse{
				Tasks:    []TaskInfo{},
				Total:    0,
				Page:     req.Page,
				PageSize: req.PageSize,
			}, nil
		}
		query = query.Where("workflow_id IN ?", memberWorkflowIDs)
	}

	// 添加其他过滤条件
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.Priority != "" {
		query = query.Where("priority = ?", req.Priority)
	}
	if req.ResponsibleID > 0 {
		query = query.Where("responsible_id = ?", req.ResponsibleID)
	}
	if req.CreatorID > 0 {
		query = query.Where("creator_id = ?", req.CreatorID)
	}
	if req.Keyword != "" {
		query = query.Where("name ILIKE ? OR description ILIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}

	// 排序
	sortBy := "created_at"
	if req.SortBy != "" {
		sortBy = req.SortBy
	}
	sortOrder := "desc"
	if req.SortOrder != "" {
		sortOrder = req.SortOrder
	}
	query = query.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))

	// 统计总数
	var total int64
	query.Count(&total)

	// 获取任务列表
	var tasks []models.TaskEnhanced
	if err := query.Offset(offset).Limit(req.PageSize).Find(&tasks).Error; err != nil {
		return nil, fmt.Errorf("获取任务列表失败: %v", err)
	}

	// 构建响应数据
	taskInfos := make([]TaskInfo, 0, len(tasks))
	for _, task := range tasks {
		// 获取相关用户信息
		var creator, responsible, reviewer models.User
		var workflow models.Workflow

		s.db.Select("username").Where("id = ?", task.CreatorID).First(&creator)
		s.db.Select("username").Where("id = ?", task.ResponsibleID).First(&responsible)
		s.db.Select("name").Where("id = ?", task.WorkflowID).First(&workflow)

		taskInfo := TaskInfo{
			ID:               task.ID,
			Name:             task.Name,
			Description:      task.Description,
			WorkflowID:       task.WorkflowID,
			WorkflowName:     workflow.Name,
			CreatorID:        task.CreatorID,
			CreatorName:      creator.Username,
			ResponsibleID:    task.ResponsibleID,
			ResponsibleName:  responsible.Username,
			Status:           task.Status,
			Priority:         task.Priority,
			RequireReview:    task.RequireReview,
			ReviewerID:       task.ReviewerID,
			StartDate:        task.StartDate,
			DueDate:          task.DueDate,
			CompletedAt:      task.CompletedAt,
			EstimatedHours:   task.EstimatedHours,
			ActualHours:      task.ActualHours,
			Progress:         task.Progress,
			Tags:             []string(task.Tags),
			CreatedAt:        task.CreatedAt,
			UpdatedAt:        task.UpdatedAt,
		}

		if task.ReviewerID > 0 {
			s.db.Select("username").Where("id = ?", task.ReviewerID).First(&reviewer)
			taskInfo.ReviewerName = reviewer.Username
		}

		taskInfos = append(taskInfos, taskInfo)
	}

	return &TaskListResponse{
		Tasks:    taskInfos,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// UpdateTask 更新任务
func (s *TaskService) UpdateTask(taskID uint, req *UpdateTaskRequest, userID uint) (*TaskInfo, error) {
	// 获取任务信息
	var task models.TaskEnhanced
	if err := s.db.Where("id = ? AND is_deleted = false", taskID).First(&task).Error; err != nil {
		return nil, errors.New("任务不存在")
	}

	// 检查用户是否有权限更新任务（创建者或负责人或工作流主管）
	var workflow models.Workflow
	s.db.Where("id = ?", task.WorkflowID).First(&workflow)

	if task.CreatorID != userID && task.ResponsibleID != userID && workflow.MasterID != userID {
		return nil, errors.New("无权限更新该任务")
	}

	// 如果任务已完成，不允许更新
	if task.Status == "completed" {
		return nil, errors.New("已完成的任务不能更新")
	}

	// 构建更新数据
	updateData := make(map[string]interface{})
	if req.Name != "" {
		updateData["name"] = req.Name
	}
	if req.Description != "" {
		updateData["description"] = req.Description
	}
	if req.ResponsibleID > 0 {
		// 检查新负责人是否为工作流成员
		var member models.WorkflowMember
		if err := s.db.Where("workflow_id = ? AND user_id = ?", task.WorkflowID, req.ResponsibleID).First(&member).Error; err != nil {
			return nil, errors.New("新负责人不是工作流成员")
		}
		updateData["responsible_id"] = req.ResponsibleID
	}
	if req.Priority != "" {
		updateData["priority"] = req.Priority
	}
	if req.RequireReview {
		updateData["require_review"] = req.RequireReview
	}
	if req.ReviewerID > 0 {
		// 检查审核人是否为工作流成员
		var member models.WorkflowMember
		if err := s.db.Where("workflow_id = ? AND user_id = ?", task.WorkflowID, req.ReviewerID).First(&member).Error; err != nil {
			return nil, errors.New("审核人不是工作流成员")
		}
		updateData["reviewer_id"] = req.ReviewerID
	}
	if req.StartDate != nil {
		updateData["start_date"] = req.StartDate
	}
	if req.DueDate != nil {
		updateData["due_date"] = req.DueDate
	}
	if req.EstimatedHours > 0 {
		updateData["estimated_hours"] = req.EstimatedHours
	}
	if req.ActualHours > 0 {
		updateData["actual_hours"] = req.ActualHours
	}
	if req.Progress <= 100 {
		updateData["progress"] = req.Progress
	}
	if len(req.Tags) > 0 {
		updateData["tags"] = models.StringArray(req.Tags)
	}

	// 更新任务
	if len(updateData) > 0 {
		if err := s.db.Model(&task).Updates(updateData).Error; err != nil {
			return nil, fmt.Errorf("更新任务失败: %v", err)
		}
	}

	// 重新获取更新后的任务信息
	return s.GetTaskByID(taskID, userID)
}

// DeleteTask 删除任务
func (s *TaskService) DeleteTask(taskID uint, userID uint) error {
	// 获取任务信息
	var task models.TaskEnhanced
	if err := s.db.Where("id = ? AND is_deleted = false", taskID).First(&task).Error; err != nil {
		return errors.New("任务不存在")
	}

	// 检查用户是否有权限删除任务（创建者或工作流主管）
	var workflow models.Workflow
	s.db.Where("id = ?", task.WorkflowID).First(&workflow)

	if task.CreatorID != userID && workflow.MasterID != userID {
		return errors.New("无权限删除该任务")
	}

	// 如果任务正在进行中，不允许删除
	if task.Status == "in_progress" {
		return errors.New("正在进行中的任务不能删除")
	}

	// 软删除任务
	now := time.Now()
	if err := s.db.Model(&task).Updates(map[string]interface{}{
		"is_deleted": true,
		"deleted_at": &now,
	}).Error; err != nil {
		return fmt.Errorf("删除任务失败: %v", err)
	}

	return nil
}

// ChangeTaskStatus 更改任务状态
func (s *TaskService) ChangeTaskStatus(taskID uint, req *ChangeStatusRequest, userID uint) (*TaskInfo, error) {
	// 获取任务信息
	var task models.TaskEnhanced
	if err := s.db.Where("id = ? AND is_deleted = false", taskID).First(&task).Error; err != nil {
		return nil, errors.New("任务不存在")
	}

	// 检查用户是否有权限更改状态
	var workflow models.Workflow
	s.db.Where("id = ?", task.WorkflowID).First(&workflow)

	canChange := false
	switch req.Status {
	case "in_progress":
		// 负责人可以开始任务
		canChange = task.ResponsibleID == userID
	case "review":
		// 负责人可以提交审核
		canChange = task.ResponsibleID == userID && task.RequireReview
	case "completed":
		if task.RequireReview {
			// 需要审核的任务，审核人可以完成
			canChange = task.ReviewerID == userID
		} else {
			// 不需要审核的任务，负责人可以完成
			canChange = task.ResponsibleID == userID
		}
	case "cancelled":
		// 创建者或工作流主管可以取消任务
		canChange = task.CreatorID == userID || workflow.MasterID == userID
	default:
		return nil, errors.New("无效的状态")
	}

	if !canChange {
		return nil, errors.New("无权限更改任务状态")
	}

	// 检查状态流转是否合法
	validTransitions := map[string][]string{
		"pending":     {"in_progress", "cancelled"},
		"in_progress": {"review", "completed", "cancelled"},
		"review":      {"completed", "in_progress"},
		"completed":   {},
		"cancelled":   {},
	}

	validNext, exists := validTransitions[task.Status]
	if !exists {
		return nil, errors.New("当前状态无效")
	}

	isValidTransition := false
	for _, validStatus := range validNext {
		if validStatus == req.Status {
			isValidTransition = true
			break
		}
	}

	if !isValidTransition {
		return nil, fmt.Errorf("不能从状态 %s 转换到 %s", task.Status, req.Status)
	}

	// 更新任务状态
	updateData := map[string]interface{}{
		"status": req.Status,
	}

	if req.Status == "completed" {
		now := time.Now()
		updateData["completed_at"] = &now
		updateData["progress"] = 100
	}

	if err := s.db.Model(&task).Updates(updateData).Error; err != nil {
		return nil, fmt.Errorf("更新任务状态失败: %v", err)
	}

	// 记录状态变更日志
	statusLog := models.TaskStatusLog{
		TaskID:     taskID,
		FromStatus: task.Status,
		ToStatus:   req.Status,
		OperatorID: userID,
		Remark:     req.Remark,
	}
	s.db.Create(&statusLog)

	// 重新获取更新后的任务信息
	return s.GetTaskByID(taskID, userID)
}

// AddToStagingArea 添加文件到暂存区
func (s *TaskService) AddToStagingArea(taskID uint, req *StagingAreaRequest, userID uint) (*StagingAreaInfo, error) {
	// 检查任务是否存在且用户有权限
	var task models.TaskEnhanced
	if err := s.db.Where("id = ? AND is_deleted = false", taskID).First(&task).Error; err != nil {
		return nil, errors.New("任务不存在")
	}

	// 检查用户是否为任务负责人或工作流成员
	var member models.WorkflowMember
	if err := s.db.Where("workflow_id = ? AND user_id = ?", task.WorkflowID, userID).First(&member).Error; err != nil {
		return nil, errors.New("无权限操作该任务")
	}

	// 检查文件是否存在
	var file models.File
	if err := s.db.Where("id = ? AND is_deleted = false", req.FileID).First(&file).Error; err != nil {
		return nil, errors.New("文件不存在")
	}

	// 检查是否已经在暂存区中
	var existingStaging models.TaskStagingArea
	if err := s.db.Where("task_id = ? AND user_id = ? AND file_id = ? AND is_submitted = false", taskID, userID, req.FileID).First(&existingStaging).Error; err == nil {
		return nil, errors.New("文件已在暂存区中")
	}

	// 添加到暂存区
	staging := models.TaskStagingArea{
		TaskID:    taskID,
		UserID:    userID,
		FileID:    req.FileID,
		Operation: req.Operation,
		Version:   1,
		Remark:    req.Remark,
	}

	if err := s.db.Create(&staging).Error; err != nil {
		return nil, fmt.Errorf("添加到暂存区失败: %v", err)
	}

	// 获取用户信息
	var user models.User
	s.db.Select("username").Where("id = ?", userID).First(&user)

	return &StagingAreaInfo{
		ID:          staging.ID,
		TaskID:      staging.TaskID,
		UserID:      staging.UserID,
		UserName:    user.Username,
		FileID:      staging.FileID,
		FileName:    file.FileName,
		Operation:   staging.Operation,
		Version:     staging.Version,
		IsSubmitted: staging.IsSubmitted,
		SubmittedAt: staging.SubmittedAt,
		Remark:      staging.Remark,
		CreatedAt:   staging.CreatedAt,
		UpdatedAt:   staging.UpdatedAt,
	}, nil
}

// GetStagingArea 获取任务暂存区
func (s *TaskService) GetStagingArea(taskID uint, userID uint) ([]StagingAreaInfo, error) {
	// 检查任务是否存在且用户有权限
	var task models.TaskEnhanced
	if err := s.db.Where("id = ? AND is_deleted = false", taskID).First(&task).Error; err != nil {
		return nil, errors.New("任务不存在")
	}

	// 检查用户是否为工作流成员
	var member models.WorkflowMember
	if err := s.db.Where("workflow_id = ? AND user_id = ?", task.WorkflowID, userID).First(&member).Error; err != nil {
		return nil, errors.New("无权限访问该任务")
	}

	// 获取暂存区列表
	var stagingList []models.TaskStagingArea
	if err := s.db.Where("task_id = ?", taskID).Find(&stagingList).Error; err != nil {
		return nil, fmt.Errorf("获取暂存区失败: %v", err)
	}

	// 构建响应数据
	stagingInfos := make([]StagingAreaInfo, 0, len(stagingList))
	for _, staging := range stagingList {
		// 获取用户和文件信息
		var user models.User
		var file models.File
		s.db.Select("username").Where("id = ?", staging.UserID).First(&user)
		s.db.Select("file_name").Where("id = ?", staging.FileID).First(&file)

		stagingInfos = append(stagingInfos, StagingAreaInfo{
			ID:          staging.ID,
			TaskID:      staging.TaskID,
			UserID:      staging.UserID,
			UserName:    user.Username,
			FileID:      staging.FileID,
			FileName:    file.FileName,
			Operation:   staging.Operation,
			Version:     staging.Version,
			IsSubmitted: staging.IsSubmitted,
			SubmittedAt: staging.SubmittedAt,
			Remark:      staging.Remark,
			CreatedAt:   staging.CreatedAt,
			UpdatedAt:   staging.UpdatedAt,
		})
	}

	return stagingInfos, nil
}

// SubmitStagingArea 提交暂存区
func (s *TaskService) SubmitStagingArea(taskID uint, userID uint) error {
	// 检查任务是否存在且用户有权限
	var task models.TaskEnhanced
	if err := s.db.Where("id = ? AND is_deleted = false", taskID).First(&task).Error; err != nil {
		return errors.New("任务不存在")
	}

	// 检查用户是否为任务负责人
	if task.ResponsibleID != userID {
		return errors.New("只有任务负责人可以提交暂存区")
	}

	// 获取用户的未提交暂存区项目
	var stagingList []models.TaskStagingArea
	if err := s.db.Where("task_id = ? AND user_id = ? AND is_submitted = false", taskID, userID).Find(&stagingList).Error; err != nil {
		return fmt.Errorf("获取暂存区失败: %v", err)
	}

	if len(stagingList) == 0 {
		return errors.New("暂存区为空")
	}

	// 开始事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 标记暂存区项目为已提交
	now := time.Now()
	if err := tx.Model(&models.TaskStagingArea{}).Where("task_id = ? AND user_id = ? AND is_submitted = false", taskID, userID).Updates(map[string]interface{}{
		"is_submitted": true,
		"submitted_at": &now,
	}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("提交暂存区失败: %v", err)
	}

	// 创建任务提交记录
	submission := models.TaskSubmission{
		TaskID:      taskID,
		SubmitterID: userID,
		Version:     1, // 这里应该根据实际情况计算版本号
		Description: "暂存区提交",
		FileCount:   uint(len(stagingList)),
		Status:      "pending",
	}

	if err := tx.Create(&submission).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("创建提交记录失败: %v", err)
	}

	return tx.Commit().Error
}

// ClearStagingArea 清空暂存区
func (s *TaskService) ClearStagingArea(taskID uint, userID uint) error {
	// 检查任务是否存在且用户有权限
	var task models.TaskEnhanced
	if err := s.db.Where("id = ? AND is_deleted = false", taskID).First(&task).Error; err != nil {
		return errors.New("任务不存在")
	}

	// 检查用户是否为任务负责人
	if task.ResponsibleID != userID {
		return errors.New("只有任务负责人可以清空暂存区")
	}

	// 删除用户的未提交暂存区项目
	if err := s.db.Where("task_id = ? AND user_id = ? AND is_submitted = false", taskID, userID).Delete(&models.TaskStagingArea{}).Error; err != nil {
		return fmt.Errorf("清空暂存区失败: %v", err)
	}

	return nil
}