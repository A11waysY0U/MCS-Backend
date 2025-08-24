package services

import (
	"time"

	"gorm.io/gorm"
	"mcs-backend/internal/models"
)

type StatisticsService struct {
	db *gorm.DB
}

func NewStatisticsService(db *gorm.DB) *StatisticsService {
	return &StatisticsService{db: db}
}

// LogOperation 记录操作日志
func (s *StatisticsService) LogOperation(userID uint, action, resource string, resourceID *uint, description, ipAddress, userAgent string) error {
	log := models.OperationLog{
		UserID:      userID,
		Action:      action,
		Resource:    resource,
		ResourceID:  resourceID,
		Description: description,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		Status:      "success",
	}

	return s.db.Create(&log).Error
}

// LogFailedOperation 记录失败的操作日志
func (s *StatisticsService) LogFailedOperation(userID uint, action, resource string, resourceID *uint, description, ipAddress, userAgent string) error {
	log := models.OperationLog{
		UserID:      userID,
		Action:      action,
		Resource:    resource,
		ResourceID:  resourceID,
		Description: description,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		Status:      "failed",
	}

	return s.db.Create(&log).Error
}

// GetOperationLogs 获取操作日志列表
type OperationLogRequest struct {
	UserID     *uint     `json:"user_id"`
	Action     string    `json:"action"`
	Resource   string    `json:"resource"`
	Status     string    `json:"status"`
	StartDate  *time.Time `json:"start_date"`
	EndDate    *time.Time `json:"end_date"`
	Page       int       `json:"page"`
	PageSize   int       `json:"page_size"`
}

type OperationLogResponse struct {
	Logs       []models.OperationLog `json:"logs"`
	Total      int64                 `json:"total"`
	Page       int                   `json:"page"`
	PageSize   int                   `json:"page_size"`
	TotalPages int                   `json:"total_pages"`
}

func (s *StatisticsService) GetOperationLogs(req *OperationLogRequest) (*OperationLogResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	query := s.db.Model(&models.OperationLog{}).Preload("User")

	// 添加过滤条件
	if req.UserID != nil {
		query = query.Where("user_id = ?", *req.UserID)
	}
	if req.Action != "" {
		query = query.Where("action = ?", req.Action)
	}
	if req.Resource != "" {
		query = query.Where("resource = ?", req.Resource)
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.StartDate != nil {
		query = query.Where("created_at >= ?", *req.StartDate)
	}
	if req.EndDate != nil {
		query = query.Where("created_at <= ?", *req.EndDate)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// 获取分页数据
	var logs []models.OperationLog
	offset := (req.Page - 1) * req.PageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(req.PageSize).Find(&logs).Error; err != nil {
		return nil, err
	}

	totalPages := int((total + int64(req.PageSize) - 1) / int64(req.PageSize))

	return &OperationLogResponse{
		Logs:       logs,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetStorageStats 获取存储统计
type StorageStatsRequest struct {
	StartDate *time.Time `json:"start_date"`
	EndDate   *time.Time `json:"end_date"`
	GroupBy   string     `json:"group_by"` // day, week, month
}

func (s *StatisticsService) GetStorageStats(req *StorageStatsRequest) ([]models.StorageStats, error) {
	query := s.db.Model(&models.StorageStats{})

	if req.StartDate != nil {
		query = query.Where("date >= ?", *req.StartDate)
	}
	if req.EndDate != nil {
		query = query.Where("date <= ?", *req.EndDate)
	}

	var stats []models.StorageStats
	err := query.Order("date DESC").Find(&stats).Error
	return stats, err
}

// UpdateStorageStats 更新存储统计
func (s *StatisticsService) UpdateStorageStats() error {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// 计算总文件数和总大小
	var totalFiles int64
	var totalSize int64
	s.db.Model(&models.File{}).Count(&totalFiles)
	s.db.Model(&models.File{}).Select("COALESCE(SUM(size), 0)").Scan(&totalSize)

	// 计算当日上传文件数和大小
	var uploadedFiles int64
	var uploadedSize int64
	s.db.Model(&models.File{}).Where("created_at >= ? AND created_at < ?", today, today.AddDate(0, 0, 1)).Count(&uploadedFiles)
	s.db.Model(&models.File{}).Where("created_at >= ? AND created_at < ?", today, today.AddDate(0, 0, 1)).Select("COALESCE(SUM(size), 0)").Scan(&uploadedSize)

	// 计算当日删除文件数（从操作日志中统计）
	var deletedFiles int64
	s.db.Model(&models.OperationLog{}).Where("action = ? AND created_at >= ? AND created_at < ?", "delete", today, today.AddDate(0, 0, 1)).Count(&deletedFiles)

	// 计算活跃用户数
	var activeUsers int64
	s.db.Model(&models.OperationLog{}).Where("created_at >= ? AND created_at < ?", today, today.AddDate(0, 0, 1)).Distinct("user_id").Count(&activeUsers)

	// 更新或创建统计记录
	stats := models.StorageStats{
		Date:          today,
		TotalFiles:    totalFiles,
		TotalSize:     totalSize,
		UploadedFiles: uploadedFiles,
		UploadedSize:  uploadedSize,
		DeletedFiles:  deletedFiles,
		ActiveUsers:   activeUsers,
		StorageUsage:  0, // 需要根据实际存储配置计算
	}

	return s.db.Where("date = ?", today).Assign(stats).FirstOrCreate(&stats).Error
}

// GetUserActivityStats 获取用户活跃度统计
type UserActivityRequest struct {
	UserID    *uint     `json:"user_id"`
	StartDate *time.Time `json:"start_date"`
	EndDate   *time.Time `json:"end_date"`
	Page      int       `json:"page"`
	PageSize  int       `json:"page_size"`
}

type UserActivityResponse struct {
	Stats      []models.UserActivityStats `json:"stats"`
	Total      int64                      `json:"total"`
	Page       int                        `json:"page"`
	PageSize   int                        `json:"page_size"`
	TotalPages int                        `json:"total_pages"`
}

func (s *StatisticsService) GetUserActivityStats(req *UserActivityRequest) (*UserActivityResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	query := s.db.Model(&models.UserActivityStats{}).Preload("User")

	if req.UserID != nil {
		query = query.Where("user_id = ?", *req.UserID)
	}
	if req.StartDate != nil {
		query = query.Where("date >= ?", *req.StartDate)
	}
	if req.EndDate != nil {
		query = query.Where("date <= ?", *req.EndDate)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var stats []models.UserActivityStats
	offset := (req.Page - 1) * req.PageSize
	if err := query.Order("date DESC").Offset(offset).Limit(req.PageSize).Find(&stats).Error; err != nil {
		return nil, err
	}

	totalPages := int((total + int64(req.PageSize) - 1) / int64(req.PageSize))

	return &UserActivityResponse{
		Stats:      stats,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

// UpdateUserActivityStats 更新用户活跃度统计
func (s *StatisticsService) UpdateUserActivityStats() error {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// 获取所有活跃用户
	var userIDs []uint
	s.db.Model(&models.OperationLog{}).Where("created_at >= ? AND created_at < ?", today, today.AddDate(0, 0, 1)).Distinct("user_id").Pluck("user_id", &userIDs)

	for _, userID := range userIDs {
		// 统计各种操作次数
		var loginCount int64
		s.db.Model(&models.OperationLog{}).Where("user_id = ? AND action = ? AND created_at >= ? AND created_at < ?", userID, "login", today, today.AddDate(0, 0, 1)).Count(&loginCount)

		var uploadCount int64
		s.db.Model(&models.OperationLog{}).Where("user_id = ? AND action = ? AND created_at >= ? AND created_at < ?", userID, "upload", today, today.AddDate(0, 0, 1)).Count(&uploadCount)

		var downloadCount int64
		s.db.Model(&models.OperationLog{}).Where("user_id = ? AND action = ? AND created_at >= ? AND created_at < ?", userID, "download", today, today.AddDate(0, 0, 1)).Count(&downloadCount)

		var fileCount int64
		s.db.Model(&models.OperationLog{}).Where("user_id = ? AND resource = ? AND created_at >= ? AND created_at < ?", userID, "file", today, today.AddDate(0, 0, 1)).Count(&fileCount)

		var workflowCount int64
		s.db.Model(&models.OperationLog{}).Where("user_id = ? AND resource = ? AND created_at >= ? AND created_at < ?", userID, "workflow", today, today.AddDate(0, 0, 1)).Count(&workflowCount)

		// 获取最后活跃时间
		var lastActiveAt time.Time
		s.db.Model(&models.OperationLog{}).Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, today, today.AddDate(0, 0, 1)).Order("created_at DESC").Limit(1).Pluck("created_at", &lastActiveAt)

		// 更新或创建统计记录
		stats := models.UserActivityStats{
			UserID:        userID,
			Date:          today,
			LoginCount:    int(loginCount),
			UploadCount:   int(uploadCount),
			DownloadCount: int(downloadCount),
			FileCount:     int(fileCount),
			WorkflowCount: int(workflowCount),
			LastActiveAt:  lastActiveAt,
		}

		if err := s.db.Where("user_id = ? AND date = ?", userID, today).Assign(stats).FirstOrCreate(&stats).Error; err != nil {
			return err
		}
	}

	return nil
}

// GetSystemOverview 获取系统概览统计
type SystemOverview struct {
	TotalUsers        int64   `json:"total_users"`
	ActiveUsers       int64   `json:"active_users"`
	TotalFiles        int64   `json:"total_files"`
	TotalSize         int64   `json:"total_size"`
	TotalWorkflows    int64   `json:"total_workflows"`
	TotalNotifications int64  `json:"total_notifications"`
	StorageUsage      float64 `json:"storage_usage"`
	TodayUploads      int64   `json:"today_uploads"`
	TodayDownloads    int64   `json:"today_downloads"`
}

func (s *StatisticsService) GetSystemOverview() (*SystemOverview, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	overview := &SystemOverview{}

	// 总用户数
	s.db.Model(&models.User{}).Count(&overview.TotalUsers)

	// 活跃用户数（今日有操作的用户）
	s.db.Model(&models.OperationLog{}).Where("created_at >= ?", today).Distinct("user_id").Count(&overview.ActiveUsers)

	// 总文件数和大小
	s.db.Model(&models.File{}).Count(&overview.TotalFiles)
	s.db.Model(&models.File{}).Select("COALESCE(SUM(size), 0)").Scan(&overview.TotalSize)

	// 总工作流数
	s.db.Model(&models.Workflow{}).Count(&overview.TotalWorkflows)

	// 总通知数
	s.db.Model(&models.Notification{}).Count(&overview.TotalNotifications)

	// 今日上传和下载次数
	s.db.Model(&models.OperationLog{}).Where("action = ? AND created_at >= ?", "upload", today).Count(&overview.TodayUploads)
	s.db.Model(&models.OperationLog{}).Where("action = ? AND created_at >= ?", "download", today).Count(&overview.TodayDownloads)

	// 存储使用率（这里需要根据实际配置计算）
	overview.StorageUsage = 0.0

	return overview, nil
}

// GetOperationStats 获取操作统计
type OperationStats struct {
	Action string `json:"action"`
	Count  int64  `json:"count"`
}

func (s *StatisticsService) GetOperationStats(startDate, endDate *time.Time) ([]OperationStats, error) {
	query := s.db.Model(&models.OperationLog{}).Select("action, COUNT(*) as count").Group("action")

	if startDate != nil {
		query = query.Where("created_at >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("created_at <= ?", *endDate)
	}

	var stats []OperationStats
	err := query.Order("count DESC").Find(&stats).Error
	return stats, err
}

// GetTopActiveUsers 获取最活跃用户
type TopActiveUser struct {
	UserID      uint   `json:"user_id"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	ActionCount int64  `json:"action_count"`
}

func (s *StatisticsService) GetTopActiveUsers(startDate, endDate *time.Time, limit int) ([]TopActiveUser, error) {
	if limit <= 0 {
		limit = 10
	}

	query := `
		SELECT 
			ol.user_id,
			u.username,
			u.email,
			COUNT(*) as action_count
		FROM operation_logs ol
		JOIN users u ON ol.user_id = u.id
		WHERE ol.deleted_at IS NULL
	`

	args := []interface{}{}
	if startDate != nil {
		query += " AND ol.created_at >= ?"
		args = append(args, *startDate)
	}
	if endDate != nil {
		query += " AND ol.created_at <= ?"
		args = append(args, *endDate)
	}

	query += " GROUP BY ol.user_id, u.username, u.email ORDER BY action_count DESC LIMIT ?"
	args = append(args, limit)

	var users []TopActiveUser
	err := s.db.Raw(query, args...).Scan(&users).Error
	return users, err
}