package services

import (
	"errors"
	"mcs-backend/internal/database"
	"mcs-backend/internal/models"
	"time"

	"gorm.io/gorm"
)

// UserGroupService 用户组服务
type UserGroupService struct {
	db *gorm.DB
}

// NewUserGroupService 创建用户组服务实例
func NewUserGroupService() *UserGroupService {
	return &UserGroupService{
		db: database.GetDB(),
	}
}

// CreateGroupRequest 创建用户组请求
type CreateGroupRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=100"`
	Description string `json:"description" binding:"max=500"`
	Color       string `json:"color" binding:"omitempty,len=7"`
}

// UpdateGroupRequest 更新用户组请求
type UpdateGroupRequest struct {
	Name        string `json:"name,omitempty" binding:"omitempty,min=2,max=100"`
	Description string `json:"description,omitempty" binding:"omitempty,max=500"`
	Color       string `json:"color,omitempty" binding:"omitempty,len=7"`
	IsActive    *bool  `json:"is_active,omitempty"`
	SortOrder   *uint  `json:"sort_order,omitempty"`
}

// AddMemberRequest 添加成员请求
type AddMemberRequest struct {
	UserID uint   `json:"user_id" binding:"required"`
	Role   string `json:"role" binding:"required,oneof=member admin"`
}

// UpdateMemberRequest 更新成员请求
type UpdateMemberRequest struct {
	Role     string `json:"role,omitempty" binding:"omitempty,oneof=member admin"`
	IsActive *bool  `json:"is_active,omitempty"`
}

// GroupListResponse 用户组列表响应
type GroupListResponse struct {
	Groups     []models.UserGroupDTO `json:"groups"`
	Total      int64                 `json:"total"`
	Page       int                   `json:"page"`
	PageSize   int                   `json:"page_size"`
	TotalPages int                   `json:"total_pages"`
}

// GroupMemberResponse 用户组成员响应
type GroupMemberResponse struct {
	ID       uint   `json:"id"`
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	RealName string `json:"real_name"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	JoinedAt string `json:"joined_at"`
	IsActive bool   `json:"is_active"`
}

// CreateGroup 创建用户组
func (s *UserGroupService) CreateGroup(creatorID uint, req *CreateGroupRequest) (*models.UserGroupDTO, error) {
	// 检查用户组名称是否已存在
	var existingGroup models.UserGroup
	if err := s.db.Where("name = ?", req.Name).First(&existingGroup).Error; err == nil {
		return nil, errors.New("用户组名称已存在")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 设置默认颜色
	color := req.Color
	if color == "" {
		color = "#409EFF"
	}

	// 创建用户组
	group := &models.UserGroup{
		Name:        req.Name,
		Description: req.Description,
		Color:       color,
		CreaterID:   creatorID,
		IsActive:    true,
		SortOrder:   0,
	}

	if err := s.db.Create(group).Error; err != nil {
		return nil, err
	}

	// 将创建者添加为管理员
	member := &models.UserGroupMember{
		UserID:    creatorID,
		GroupID:   group.ID,
		Role:      "admin",
		InviterID: creatorID,
		IsActive:  true,
	}

	if err := s.db.Create(member).Error; err != nil {
		// 如果添加成员失败，删除已创建的用户组
		s.db.Delete(group)
		return nil, err
	}

	return group.ToUserGroupDTO(), nil
}

// GetGroupByID 根据ID获取用户组
func (s *UserGroupService) GetGroupByID(id uint) (*models.UserGroupDTO, error) {
	var group models.UserGroup
	if err := s.db.Where("id = ?", id).First(&group).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户组不存在")
		}
		return nil, err
	}

	return group.ToUserGroupDTO(), nil
}

// GetGroupList 获取用户组列表
func (s *UserGroupService) GetGroupList(page, pageSize int, keyword string, isActive *bool) (*GroupListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	query := s.db.Model(&models.UserGroup{})

	// 关键词搜索
	if keyword != "" {
		query = query.Where("name ILIKE ? OR description ILIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 状态筛选
	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// 获取用户组列表
	var groups []models.UserGroup
	offset := (page - 1) * pageSize
	if err := query.Order("sort_order ASC, created_at DESC").Offset(offset).Limit(pageSize).Find(&groups).Error; err != nil {
		return nil, err
	}

	// 转换为DTO并获取成员数量
	groupDTOs := make([]models.UserGroupDTO, len(groups))
	for i, group := range groups {
		groupDTO := *group.ToUserGroupDTO()
		
		// 获取成员数量
		var memberCount int64
		s.db.Model(&models.UserGroupMember{}).Where("group_id = ? AND is_active = true", group.ID).Count(&memberCount)
		groupDTO.MemberCount = memberCount
		
		groupDTOs[i] = groupDTO
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	return &GroupListResponse{
		Groups:     groupDTOs,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// UpdateGroup 更新用户组信息
func (s *UserGroupService) UpdateGroup(id uint, req *UpdateGroupRequest) (*models.UserGroupDTO, error) {
	var group models.UserGroup
	if err := s.db.Where("id = ?", id).First(&group).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户组不存在")
		}
		return nil, err
	}

	// 检查名称是否已被其他用户组使用
	if req.Name != "" && req.Name != group.Name {
		var existingGroup models.UserGroup
		if err := s.db.Where("name = ? AND id != ?", req.Name, id).First(&existingGroup).Error; err == nil {
			return nil, errors.New("用户组名称已被使用")
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}

	// 更新字段
	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Color != "" {
		updates["color"] = req.Color
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}
	updates["updated_at"] = time.Now()

	if err := s.db.Model(&group).Updates(updates).Error; err != nil {
		return nil, err
	}

	// 重新获取更新后的用户组信息
	if err := s.db.Where("id = ?", id).First(&group).Error; err != nil {
		return nil, err
	}

	return group.ToUserGroupDTO(), nil
}

// DeleteGroup 删除用户组
func (s *UserGroupService) DeleteGroup(id uint) error {
	var group models.UserGroup
	if err := s.db.Where("id = ?", id).First(&group).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户组不存在")
		}
		return err
	}

	// 开始事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 删除所有成员关系
	if err := tx.Where("group_id = ?", id).Delete(&models.UserGroupMember{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 删除用户组
	if err := tx.Delete(&group).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// AddMember 添加成员到用户组
func (s *UserGroupService) AddMember(groupID uint, inviterID uint, req *AddMemberRequest) error {
	// 检查用户组是否存在
	var group models.UserGroup
	if err := s.db.Where("id = ? AND is_active = true", groupID).First(&group).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户组不存在或已禁用")
		}
		return err
	}

	// 检查用户是否存在
	var user models.User
	if err := s.db.Where("id = ? AND deleted_at IS NULL", req.UserID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户不存在")
		}
		return err
	}

	// 检查用户是否已在用户组中
	var existingMember models.UserGroupMember
	if err := s.db.Where("user_id = ? AND group_id = ?", req.UserID, groupID).First(&existingMember).Error; err == nil {
		if existingMember.IsActive {
			return errors.New("用户已在该用户组中")
		}
		// 如果用户之前在组中但已被移除，重新激活
		return s.db.Model(&existingMember).Updates(map[string]interface{}{
			"role":       req.Role,
			"inviter_id": inviterID,
			"is_active":  true,
			"joined_at":  time.Now(),
		}).Error
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// 添加新成员
	member := &models.UserGroupMember{
		UserID:    req.UserID,
		GroupID:   groupID,
		Role:      req.Role,
		InviterID: inviterID,
		IsActive:  true,
	}

	return s.db.Create(member).Error
}

// RemoveMember 从用户组移除成员
func (s *UserGroupService) RemoveMember(groupID, userID uint) error {
	var member models.UserGroupMember
	if err := s.db.Where("user_id = ? AND group_id = ? AND is_active = true", userID, groupID).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户不在该用户组中")
		}
		return err
	}

	// 软删除（设置为非活跃状态）
	return s.db.Model(&member).Update("is_active", false).Error
}

// UpdateMember 更新用户组成员信息
func (s *UserGroupService) UpdateMember(groupID, userID uint, req *UpdateMemberRequest) error {
	var member models.UserGroupMember
	if err := s.db.Where("user_id = ? AND group_id = ?", userID, groupID).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户不在该用户组中")
		}
		return err
	}

	// 更新字段
	updates := make(map[string]interface{})
	if req.Role != "" {
		updates["role"] = req.Role
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if len(updates) > 0 {
		return s.db.Model(&member).Updates(updates).Error
	}

	return nil
}

// GetGroupMembers 获取用户组成员列表
func (s *UserGroupService) GetGroupMembers(groupID uint, page, pageSize int) ([]GroupMemberResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 检查用户组是否存在
	var group models.UserGroup
	if err := s.db.Where("id = ?", groupID).First(&group).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, errors.New("用户组不存在")
		}
		return nil, 0, err
	}

	// 获取总数
	var total int64
	if err := s.db.Model(&models.UserGroupMember{}).Where("group_id = ? AND is_active = true", groupID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取成员列表
	var members []GroupMemberResponse
	offset := (page - 1) * pageSize

	query := `
		SELECT 
			ugm.id, ugm.user_id, ugm.role, ugm.joined_at, ugm.is_active,
			u.username, u.real_name, u.email
		FROM user_group_members ugm
		JOIN users u ON ugm.user_id = u.id
		WHERE ugm.group_id = ? AND ugm.is_active = true AND u.deleted_at IS NULL
		ORDER BY ugm.joined_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := s.db.Raw(query, groupID, pageSize, offset).Rows()
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var member GroupMemberResponse
		var joinedAt time.Time
		if err := rows.Scan(&member.ID, &member.UserID, &member.Role, &joinedAt, &member.IsActive,
			&member.Username, &member.RealName, &member.Email); err != nil {
			return nil, 0, err
		}
		member.JoinedAt = joinedAt.Format("2006-01-02 15:04:05")
		members = append(members, member)
	}

	return members, total, nil
}

// GetUserGroups 获取用户所属的用户组列表
func (s *UserGroupService) GetUserGroups(userID uint) ([]models.UserGroupDTO, error) {
	var groups []models.UserGroup

	query := `
		SELECT ug.*
		FROM user_groups ug
		JOIN user_group_members ugm ON ug.id = ugm.group_id
		WHERE ugm.user_id = ? AND ugm.is_active = true AND ug.is_active = true
		ORDER BY ug.sort_order ASC, ug.created_at DESC
	`

	if err := s.db.Raw(query, userID).Scan(&groups).Error; err != nil {
		return nil, err
	}

	// 转换为DTO
	groupDTOs := make([]models.UserGroupDTO, len(groups))
	for i, group := range groups {
		groupDTOs[i] = *group.ToUserGroupDTO()
	}

	return groupDTOs, nil
}