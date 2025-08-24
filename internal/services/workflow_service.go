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

// WorkflowService 工作流服务
type WorkflowService struct {
	db     *gorm.DB
	config *config.Config
}

// NewWorkflowService 创建工作流服务
func NewWorkflowService(cfg *config.Config) *WorkflowService {
	return &WorkflowService{
		db:     database.GetDB(),
		config: cfg,
	}
}

// GetConfig 获取配置
func (s *WorkflowService) GetConfig() *config.Config {
	return s.config
}

// CreateWorkflowRequest 创建工作流请求
type CreateWorkflowRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// UpdateWorkflowRequest 更新工作流请求
type UpdateWorkflowRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// WorkflowListResponse 工作流列表响应
type WorkflowListResponse struct {
	Workflows []WorkflowInfo `json:"workflows"`
	Total     int64          `json:"total"`
	Page      int            `json:"page"`
	PageSize  int            `json:"page_size"`
}

// WorkflowInfo 工作流信息
type WorkflowInfo struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	MasterID    uint      `json:"master_id"`
	MasterName  string    `json:"master_name"`
	MemberCount int64     `json:"member_count"`
	TaskCount   int64     `json:"task_count"`
	CreatedAt   time.Time `json:"created_at"`
}

// WorkflowMemberInfo 工作流成员信息
type WorkflowMemberInfo struct {
	ID         uint      `json:"id"`
	WorkflowID uint      `json:"workflow_id"`
	UserID     uint      `json:"user_id"`
	UserName   string    `json:"user_name"`
	Role       string    `json:"role"`
	JoinedAt   time.Time `json:"joined_at"`
}

// AddWorkflowMemberRequest 添加工作流成员请求
type AddWorkflowMemberRequest struct {
	UserID uint   `json:"user_id" binding:"required"`
	Role   string `json:"role" binding:"required"`
}

// UpdateWorkflowMemberRequest 更新工作流成员请求
type UpdateWorkflowMemberRequest struct {
	Role string `json:"role" binding:"required"`
}

// CreateWorkflow 创建工作流
func (s *WorkflowService) CreateWorkflow(req *CreateWorkflowRequest, userID uint) (*WorkflowInfo, error) {
	// 检查工作流名称是否已存在
	var existingWorkflow models.Workflow
	if err := s.db.Where("name = ?", req.Name).First(&existingWorkflow).Error; err == nil {
		return nil, errors.New("工作流名称已存在")
	}

	// 创建工作流
	workflow := models.Workflow{
		Name:        req.Name,
		Description: req.Description,
		MasterID:    userID,
	}

	if err := s.db.Create(&workflow).Error; err != nil {
		return nil, fmt.Errorf("创建工作流失败: %v", err)
	}

	// 添加创建者为工作流成员
	member := models.WorkflowMember{
		WorkflowID: workflow.ID,
		UserID:     userID,
		Role:       "master",
	}

	if err := s.db.Create(&member).Error; err != nil {
		return nil, fmt.Errorf("添加工作流成员失败: %v", err)
	}

	// 获取用户信息
	var user models.User
	if err := s.db.Select("username").Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %v", err)
	}

	return &WorkflowInfo{
		ID:          workflow.ID,
		Name:        workflow.Name,
		Description: workflow.Description,
		MasterID:    workflow.MasterID,
		MasterName:  user.Username,
		MemberCount: 1,
		TaskCount:   0,
		CreatedAt:   workflow.CreatedAt,
	}, nil
}

// GetWorkflowByID 根据ID获取工作流
func (s *WorkflowService) GetWorkflowByID(workflowID uint, userID uint) (*WorkflowInfo, error) {
	// 检查用户是否有权限访问该工作流
	var member models.WorkflowMember
	if err := s.db.Where("workflow_id = ? AND user_id = ?", workflowID, userID).First(&member).Error; err != nil {
		return nil, errors.New("无权限访问该工作流")
	}

	// 获取工作流信息
	var workflow models.Workflow
	if err := s.db.Where("id = ?", workflowID).First(&workflow).Error; err != nil {
		return nil, errors.New("工作流不存在")
	}

	// 获取主管信息
	var master models.User
	if err := s.db.Select("username").Where("id = ?", workflow.MasterID).First(&master).Error; err != nil {
		return nil, fmt.Errorf("获取主管信息失败: %v", err)
	}

	// 统计成员数量
	var memberCount int64
	s.db.Model(&models.WorkflowMember{}).Where("workflow_id = ?", workflowID).Count(&memberCount)

	// 统计任务数量
	var taskCount int64
	s.db.Model(&models.Task{}).Where("workflow_id = ? AND deleted_at IS NULL", workflowID).Count(&taskCount)

	return &WorkflowInfo{
		ID:          workflow.ID,
		Name:        workflow.Name,
		Description: workflow.Description,
		MasterID:    workflow.MasterID,
		MasterName:  master.Username,
		MemberCount: memberCount,
		TaskCount:   taskCount,
		CreatedAt:   workflow.CreatedAt,
	}, nil
}

// GetWorkflowList 获取工作流列表
func (s *WorkflowService) GetWorkflowList(userID uint, page, pageSize int) (*WorkflowListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	// 获取用户参与的工作流ID列表
	var memberWorkflowIDs []uint
	if err := s.db.Model(&models.WorkflowMember{}).Where("user_id = ?", userID).Pluck("workflow_id", &memberWorkflowIDs).Error; err != nil {
		return nil, fmt.Errorf("获取用户工作流失败: %v", err)
	}

	if len(memberWorkflowIDs) == 0 {
		return &WorkflowListResponse{
			Workflows: []WorkflowInfo{},
			Total:     0,
			Page:      page,
			PageSize:  pageSize,
		}, nil
	}

	// 获取工作流列表
	var workflows []models.Workflow
	if err := s.db.Where("id IN ?", memberWorkflowIDs).Offset(offset).Limit(pageSize).Find(&workflows).Error; err != nil {
		return nil, fmt.Errorf("获取工作流列表失败: %v", err)
	}

	// 统计总数
	var total int64
	s.db.Model(&models.Workflow{}).Where("id IN ?", memberWorkflowIDs).Count(&total)

	// 构建响应数据
	workflowInfos := make([]WorkflowInfo, 0, len(workflows))
	for _, workflow := range workflows {
		// 获取主管信息
		var master models.User
		if err := s.db.Select("username").Where("id = ?", workflow.MasterID).First(&master).Error; err != nil {
			continue
		}

		// 统计成员数量
		var memberCount int64
		s.db.Model(&models.WorkflowMember{}).Where("workflow_id = ?", workflow.ID).Count(&memberCount)

		// 统计任务数量
		var taskCount int64
		s.db.Model(&models.Task{}).Where("workflow_id = ? AND deleted_at IS NULL", workflow.ID).Count(&taskCount)

		workflowInfos = append(workflowInfos, WorkflowInfo{
			ID:          workflow.ID,
			Name:        workflow.Name,
			Description: workflow.Description,
			MasterID:    workflow.MasterID,
			MasterName:  master.Username,
			MemberCount: memberCount,
			TaskCount:   taskCount,
			CreatedAt:   workflow.CreatedAt,
		})
	}

	return &WorkflowListResponse{
		Workflows: workflowInfos,
		Total:     total,
		Page:      page,
		PageSize:  pageSize,
	}, nil
}

// UpdateWorkflow 更新工作流
func (s *WorkflowService) UpdateWorkflow(workflowID uint, req *UpdateWorkflowRequest, userID uint) (*WorkflowInfo, error) {
	// 检查用户是否为工作流主管
	var workflow models.Workflow
	if err := s.db.Where("id = ? AND master_id = ?", workflowID, userID).First(&workflow).Error; err != nil {
		return nil, errors.New("无权限更新该工作流")
	}

	// 检查工作流名称是否已存在（排除当前工作流）
	if req.Name != "" && req.Name != workflow.Name {
		var existingWorkflow models.Workflow
		if err := s.db.Where("name = ? AND id != ?", req.Name, workflowID).First(&existingWorkflow).Error; err == nil {
			return nil, errors.New("工作流名称已存在")
		}
	}

	// 更新工作流信息
	updateData := make(map[string]interface{})
	if req.Name != "" {
		updateData["name"] = req.Name
	}
	if req.Description != "" {
		updateData["description"] = req.Description
	}

	if len(updateData) > 0 {
		if err := s.db.Model(&workflow).Updates(updateData).Error; err != nil {
			return nil, fmt.Errorf("更新工作流失败: %v", err)
		}
	}

	// 重新获取更新后的工作流信息
	return s.GetWorkflowByID(workflowID, userID)
}

// DeleteWorkflow 删除工作流
func (s *WorkflowService) DeleteWorkflow(workflowID uint, userID uint) error {
	// 检查用户是否为工作流主管
	var workflow models.Workflow
	if err := s.db.Where("id = ? AND master_id = ?", workflowID, userID).First(&workflow).Error; err != nil {
		return errors.New("无权限删除该工作流")
	}

	// 检查是否还有未完成的任务
	var taskCount int64
	s.db.Model(&models.Task{}).Where("workflow_id = ? AND status NOT IN ? AND deleted_at IS NULL", workflowID, []string{"completed", "cancelled"}).Count(&taskCount)
	if taskCount > 0 {
		return errors.New("工作流中还有未完成的任务，无法删除")
	}

	// 开始事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 删除工作流成员
	if err := tx.Where("workflow_id = ?", workflowID).Delete(&models.WorkflowMember{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("删除工作流成员失败: %v", err)
	}

	// 软删除所有相关任务
	if err := tx.Model(&models.Task{}).Where("workflow_id = ?", workflowID).Update("deleted_at", time.Now()).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("删除相关任务失败: %v", err)
	}

	// 删除工作流
	if err := tx.Delete(&workflow).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("删除工作流失败: %v", err)
	}

	return tx.Commit().Error
}

// AddMember 添加工作流成员
func (s *WorkflowService) AddMember(workflowID uint, req *AddWorkflowMemberRequest, operatorID uint) (*WorkflowMemberInfo, error) {
	// 检查操作者是否为工作流主管
	var workflow models.Workflow
	if err := s.db.Where("id = ? AND master_id = ?", workflowID, operatorID).First(&workflow).Error; err != nil {
		return nil, errors.New("无权限添加成员")
	}

	// 检查用户是否已经是成员
	var existingMember models.WorkflowMember
	if err := s.db.Where("workflow_id = ? AND user_id = ?", workflowID, req.UserID).First(&existingMember).Error; err == nil {
		return nil, errors.New("用户已经是工作流成员")
	}

	// 检查用户是否存在
	var user models.User
	if err := s.db.Where("id = ?", req.UserID).First(&user).Error; err != nil {
		return nil, errors.New("用户不存在")
	}

	// 添加成员
	member := models.WorkflowMember{
		WorkflowID: workflowID,
		UserID:     req.UserID,
		Role:       req.Role,
	}

	if err := s.db.Create(&member).Error; err != nil {
		return nil, fmt.Errorf("添加成员失败: %v", err)
	}

	return &WorkflowMemberInfo{
		ID:         member.ID,
		WorkflowID: member.WorkflowID,
		UserID:     member.UserID,
		UserName:   user.Username,
		Role:       member.Role,
		JoinedAt:   time.Now(),
	}, nil
}

// RemoveMember 移除工作流成员
func (s *WorkflowService) RemoveMember(workflowID uint, userID uint, operatorID uint) error {
	// 检查操作者是否为工作流主管
	var workflow models.Workflow
	if err := s.db.Where("id = ? AND master_id = ?", workflowID, operatorID).First(&workflow).Error; err != nil {
		return errors.New("无权限移除成员")
	}

	// 不能移除主管
	if userID == workflow.MasterID {
		return errors.New("不能移除工作流主管")
	}

	// 检查成员是否存在
	var member models.WorkflowMember
	if err := s.db.Where("workflow_id = ? AND user_id = ?", workflowID, userID).First(&member).Error; err != nil {
		return errors.New("成员不存在")
	}

	// 检查该成员是否有未完成的任务
	var taskCount int64
	s.db.Model(&models.Task{}).Where("workflow_id = ? AND responsible_id = ? AND status NOT IN ? AND deleted_at IS NULL", workflowID, userID, []string{"completed", "cancelled"}).Count(&taskCount)
	if taskCount > 0 {
		return errors.New("该成员还有未完成的任务，无法移除")
	}

	// 移除成员
	if err := s.db.Delete(&member).Error; err != nil {
		return fmt.Errorf("移除成员失败: %v", err)
	}

	return nil
}

// UpdateMemberRole 更新工作流成员角色
func (s *WorkflowService) UpdateMemberRole(workflowID uint, userID uint, req *UpdateWorkflowMemberRequest, operatorID uint) (*WorkflowMemberInfo, error) {
	// 检查操作者是否为工作流主管
	var workflow models.Workflow
	if err := s.db.Where("id = ? AND master_id = ?", workflowID, operatorID).First(&workflow).Error; err != nil {
		return nil, errors.New("无权限更新成员角色")
	}

	// 不能更新主管角色
	if userID == workflow.MasterID {
		return nil, errors.New("不能更新工作流主管角色")
	}

	// 检查成员是否存在
	var member models.WorkflowMember
	if err := s.db.Where("workflow_id = ? AND user_id = ?", workflowID, userID).First(&member).Error; err != nil {
		return nil, errors.New("成员不存在")
	}

	// 更新角色
	if err := s.db.Model(&member).Update("role", req.Role).Error; err != nil {
		return nil, fmt.Errorf("更新成员角色失败: %v", err)
	}

	// 获取用户信息
	var user models.User
	if err := s.db.Select("username").Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %v", err)
	}

	return &WorkflowMemberInfo{
		ID:         member.ID,
		WorkflowID: member.WorkflowID,
		UserID:     member.UserID,
		UserName:   user.Username,
		Role:       req.Role,
		JoinedAt:   member.CreatedAt,
	}, nil
}

// GetWorkflowMembers 获取工作流成员列表
func (s *WorkflowService) GetWorkflowMembers(workflowID uint, userID uint) ([]WorkflowMemberInfo, error) {
	// 检查用户是否有权限访问该工作流
	var member models.WorkflowMember
	if err := s.db.Where("workflow_id = ? AND user_id = ?", workflowID, userID).First(&member).Error; err != nil {
		return nil, errors.New("无权限访问该工作流")
	}

	// 获取成员列表
	var members []models.WorkflowMember
	if err := s.db.Where("workflow_id = ?", workflowID).Find(&members).Error; err != nil {
		return nil, fmt.Errorf("获取成员列表失败: %v", err)
	}

	// 构建响应数据
	memberInfos := make([]WorkflowMemberInfo, 0, len(members))
	for _, m := range members {
		// 获取用户信息
		var user models.User
		if err := s.db.Select("username").Where("id = ?", m.UserID).First(&user).Error; err != nil {
			continue
		}

		memberInfos = append(memberInfos, WorkflowMemberInfo{
			ID:         m.ID,
			WorkflowID: m.WorkflowID,
			UserID:     m.UserID,
			UserName:   user.Username,
			Role:       m.Role,
			JoinedAt:   m.CreatedAt,
		})
	}

	return memberInfos, nil
}