package services

import (
	"errors"
	"time"

	"mcs-backend/internal/config"
	"mcs-backend/internal/database"
	"mcs-backend/internal/models"
	"mcs-backend/internal/utils"

	"gorm.io/gorm"
)

// AuthService 认证服务
type AuthService struct {
	db         *gorm.DB
	jwtManager *utils.JWTManager
}

// NewAuthService 创建认证服务
func NewAuthService(cfg *config.Config) *AuthService {
	return &AuthService{
		db:         database.GetDB(),
		jwtManager: utils.NewJWTManager(cfg),
	}
}

// RegisterRequest 注册请求结构
type RegisterRequest struct {
	Username   string `json:"username" binding:"required,min=3,max=50"`
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required,min=6"`
	InviteCode string `json:"invite_code" binding:"required"`
	RealName   string `json:"real_name" binding:"required,max=100"`
}

// LoginRequest 登录请求结构
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse 认证响应结构
type AuthResponse struct {
	Token     string           `json:"token"`
	ExpiresAt time.Time        `json:"expires_at"`
	User      *models.UserInfo `json:"user"`
}

// RefreshRequest 刷新令牌请求结构
type RefreshRequest struct {
	Token string `json:"token" binding:"required"`
}

// Register 用户注册
func (s *AuthService) Register(req *RegisterRequest) (*AuthResponse, error) {
	// 验证邀请码
	var inviteCode models.InviteCode
	err := s.db.Where("code = ? AND status = 'active'", req.InviteCode).First(&inviteCode).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid or expired invite code")
		}
		return nil, err
	}

	// 检查邀请码是否已达到使用限制
	if inviteCode.MaxUses > 0 && inviteCode.UsedCount >= inviteCode.MaxUses {
		return nil, errors.New("invite code has reached maximum usage limit")
	}

	// 检查邀请码是否过期
	if inviteCode.ExpiresAt != nil && inviteCode.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("invite code has expired")
	}

	// 检查用户名是否已存在
	var existingUser models.User
	err = s.db.Where("username = ? OR email = ?", req.Username, req.Email).First(&existingUser).Error
	if err == nil {
		return nil, errors.New("username or email already exists")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 哈希密码
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// 创建用户
	user := models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		RealName:     req.RealName,
		Role:         "user", // 默认角色
		Status:       "active",
		InviteCodeID: &inviteCode.ID,
	}

	// 开始事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 创建用户
	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// 更新邀请码使用次数
	if err := tx.Model(&inviteCode).Update("used_count", inviteCode.UsedCount+1).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	// 生成JWT令牌
	token, err := s.jwtManager.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, err
	}

	// 构造用户信息
	userInfo := &models.UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		RealName: user.RealName,
		Role:     user.Role,
		Status:   user.Status,
	}

	return &AuthResponse{
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour), // 假设24小时过期
		User:      userInfo,
	}, nil
}

// Login 用户登录
func (s *AuthService) Login(req *LoginRequest) (*AuthResponse, error) {
	// 查找用户
	var user models.User
	err := s.db.Where("username = ? OR email = ?", req.Username, req.Username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid username or password")
		}
		return nil, err
	}

	// 检查用户状态
	if user.Status != "active" {
		return nil, errors.New("user account is not active")
	}

	// 验证密码
	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		return nil, errors.New("invalid username or password")
	}

	// 更新最后登录时间
	now := time.Now()
	s.db.Model(&user).Update("last_login_at", &now)

	// 生成JWT令牌
	token, err := s.jwtManager.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, err
	}

	// 构造用户信息
	userInfo := &models.UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		RealName: user.RealName,
		Role:     user.Role,
		Status:   user.Status,
	}

	return &AuthResponse{
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour), // 假设24小时过期
		User:      userInfo,
	}, nil
}

// RefreshToken 刷新令牌
func (s *AuthService) RefreshToken(req *RefreshRequest) (*AuthResponse, error) {
	// 验证并刷新令牌
	newToken, err := s.jwtManager.RefreshToken(req.Token)
	if err != nil {
		return nil, err
	}

	// 从旧令牌中提取用户信息
	claims, err := s.jwtManager.ValidateToken(req.Token)
	if err != nil {
		return nil, err
	}

	// 获取最新的用户信息
	var user models.User
	err = s.db.First(&user, claims.UserID).Error
	if err != nil {
		return nil, err
	}

	// 构造用户信息
	userInfo := &models.UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		RealName: user.RealName,
		Role:     user.Role,
		Status:   user.Status,
	}

	return &AuthResponse{
		Token:     newToken,
		ExpiresAt: time.Now().Add(24 * time.Hour), // 假设24小时过期
		User:      userInfo,
	}, nil
}

// ValidateInviteCode 验证邀请码
func (s *AuthService) ValidateInviteCode(code string) error {
	var inviteCode models.InviteCode
	err := s.db.Where("code = ? AND status = 'active'", code).First(&inviteCode).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("invalid or expired invite code")
		}
		return err
	}

	// 检查邀请码是否已达到使用限制
	if inviteCode.MaxUses > 0 && inviteCode.UsedCount >= inviteCode.MaxUses {
		return errors.New("invite code has reached maximum usage limit")
	}

	// 检查邀请码是否过期
	if inviteCode.ExpiresAt != nil && inviteCode.ExpiresAt.Before(time.Now()) {
		return errors.New("invite code has expired")
	}

	return nil
}