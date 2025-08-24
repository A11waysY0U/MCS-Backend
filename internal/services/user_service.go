package services

import (
	"errors"
	"mcs-backend/internal/database"
	"mcs-backend/internal/models"
	"mcs-backend/internal/utils"
	"time"

	"gorm.io/gorm"
)

// UserService 用户服务
type UserService struct {
	db *gorm.DB
}

// NewUserService 创建用户服务实例
func NewUserService() *UserService {
	return &UserService{
		db: database.GetDB(),
	}
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username    string `json:"username" binding:"required,min=3,max=50"`
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=6"`
	RealName    string `json:"real_name" binding:"required,min=2,max=50"`
	Role        string `json:"role" binding:"required,oneof=user admin super_admin"`
	InviteCode  string `json:"invite_code,omitempty"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Email    string `json:"email,omitempty" binding:"omitempty,email"`
	RealName string `json:"real_name,omitempty" binding:"omitempty,min=2,max=50"`
	Role     string `json:"role,omitempty" binding:"omitempty,oneof=user admin super_admin"`
	Status   string `json:"status,omitempty" binding:"omitempty,oneof=active inactive banned"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// UserListResponse 用户列表响应
type UserListResponse struct {
	Users      []models.UserDTO `json:"users"`
	Total      int64            `json:"total"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	TotalPages int              `json:"total_pages"`
}

// CreateUser 创建用户
func (s *UserService) CreateUser(req *CreateUserRequest) (*models.UserDTO, error) {
	// 检查用户名是否已存在
	var existingUser models.User
	if err := s.db.Where("username = ? OR email = ?", req.Username, req.Email).First(&existingUser).Error; err == nil {
		return nil, errors.New("用户名或邮箱已存在")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 验证邀请码（如果提供）
	var inviteCodeID *uint
	if req.InviteCode != "" {
		var inviteCode models.InviteCode
		if err := s.db.Where("code = ? AND status = 'active'", req.InviteCode).First(&inviteCode).Error; err != nil {
			return nil, errors.New("无效的邀请码")
		}
		
		// 检查邀请码使用次数
		if inviteCode.MaxUses > 0 && inviteCode.UsedCount >= inviteCode.MaxUses {
			return nil, errors.New("邀请码已达到最大使用次数")
		}
		
		inviteCodeID = &inviteCode.ID
	}

	// 哈希密码
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// 创建用户
	user := &models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		RealName:     req.RealName,
		Role:         req.Role,
		Status:       "active",
		InviteCodeID: inviteCodeID,
	}

	if err := s.db.Create(user).Error; err != nil {
		return nil, err
	}

	// 更新邀请码使用次数
	if inviteCodeID != nil {
		s.db.Model(&models.InviteCode{}).Where("id = ?", *inviteCodeID).UpdateColumn("used_count", gorm.Expr("used_count + 1"))
	}

	return user.ToUserDTO(), nil
}

// GetUserByID 根据ID获取用户
func (s *UserService) GetUserByID(id uint) (*models.UserDTO, error) {
	var user models.User
	if err := s.db.Where("id = ? AND deleted_at IS NULL", id).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}

	return user.ToUserDTO(), nil
}

// GetUserList 获取用户列表
func (s *UserService) GetUserList(page, pageSize int, keyword, role, status string) (*UserListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	query := s.db.Model(&models.User{}).Where("deleted_at IS NULL")

	// 关键词搜索
	if keyword != "" {
		query = query.Where("username ILIKE ? OR email ILIKE ? OR real_name ILIKE ?", 
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 角色筛选
	if role != "" {
		query = query.Where("role = ?", role)
	}

	// 状态筛选
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// 获取用户列表
	var users []models.User
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, err
	}

	// 转换为DTO
	userDTOs := make([]models.UserDTO, len(users))
	for i, user := range users {
		userDTOs[i] = *user.ToUserDTO()
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	return &UserListResponse{
		Users:      userDTOs,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// UpdateUser 更新用户信息
func (s *UserService) UpdateUser(id uint, req *UpdateUserRequest) (*models.UserDTO, error) {
	var user models.User
	if err := s.db.Where("id = ? AND deleted_at IS NULL", id).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}

	// 检查邮箱是否已被其他用户使用
	if req.Email != "" && req.Email != user.Email {
		var existingUser models.User
		if err := s.db.Where("email = ? AND id != ? AND deleted_at IS NULL", req.Email, id).First(&existingUser).Error; err == nil {
			return nil, errors.New("邮箱已被其他用户使用")
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}

	// 更新字段
	updates := make(map[string]interface{})
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.RealName != "" {
		updates["real_name"] = req.RealName
	}
	if req.Role != "" {
		updates["role"] = req.Role
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	updates["updated_at"] = time.Now()

	if err := s.db.Model(&user).Updates(updates).Error; err != nil {
		return nil, err
	}

	// 重新获取更新后的用户信息
	if err := s.db.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}

	return user.ToUserDTO(), nil
}

// ChangePassword 修改密码
func (s *UserService) ChangePassword(userID uint, req *ChangePasswordRequest) error {
	var user models.User
	if err := s.db.Where("id = ? AND deleted_at IS NULL", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户不存在")
		}
		return err
	}

	// 验证旧密码
	if !utils.CheckPassword(req.OldPassword, user.PasswordHash) {
		return errors.New("旧密码不正确")
	}

	// 哈希新密码
	newHashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	// 更新密码
	return s.db.Model(&user).Updates(map[string]interface{}{
		"password_hash": newHashedPassword,
		"updated_at":    time.Now(),
	}).Error
}

// DeleteUser 删除用户（软删除）
func (s *UserService) DeleteUser(id uint) error {
	var user models.User
	if err := s.db.Where("id = ? AND deleted_at IS NULL", id).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户不存在")
		}
		return err
	}

	// 不能删除超级管理员
	if user.Role == "super_admin" {
		return errors.New("不能删除超级管理员")
	}

	// 软删除
	now := time.Now()
	return s.db.Model(&user).Updates(map[string]interface{}{
		"deleted_at": &now,
		"updated_at": now,
	}).Error
}

// UpdateLastLogin 更新最后登录时间
func (s *UserService) UpdateLastLogin(userID uint) error {
	now := time.Now()
	return s.db.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"last_login_at": &now,
		"updated_at":    now,
	}).Error
}

// GetUserStats 获取用户统计信息
func (s *UserService) GetUserStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 总用户数
	var totalUsers int64
	if err := s.db.Model(&models.User{}).Where("deleted_at IS NULL").Count(&totalUsers).Error; err != nil {
		return nil, err
	}
	stats["total_users"] = totalUsers

	// 活跃用户数
	var activeUsers int64
	if err := s.db.Model(&models.User{}).Where("deleted_at IS NULL AND status = 'active'").Count(&activeUsers).Error; err != nil {
		return nil, err
	}
	stats["active_users"] = activeUsers

	// 今日新增用户
	today := time.Now().Truncate(24 * time.Hour)
	var todayNewUsers int64
	if err := s.db.Model(&models.User{}).Where("deleted_at IS NULL AND created_at >= ?", today).Count(&todayNewUsers).Error; err != nil {
		return nil, err
	}
	stats["today_new_users"] = todayNewUsers

	// 按角色统计
	roleStats := make(map[string]int64)
	rows, err := s.db.Model(&models.User{}).Select("role, count(*) as count").Where("deleted_at IS NULL").Group("role").Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var role string
		var count int64
		if err := rows.Scan(&role, &count); err != nil {
			return nil, err
		}
		roleStats[role] = count
	}
	stats["role_stats"] = roleStats

	return stats, nil
}