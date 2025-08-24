package services

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"mcs-backend/internal/models"

	"gorm.io/gorm"
)

type DownloadService struct {
	db           *gorm.DB
	uploadPath   string
	downloadPath string
	baseURL      string
}

func NewDownloadService(db *gorm.DB, uploadPath, downloadPath, baseURL string) *DownloadService {
	return &DownloadService{
		db:           db,
		uploadPath:   uploadPath,
		downloadPath: downloadPath,
		baseURL:      baseURL,
	}
}

// CreateBatchDownloadTask 创建批量下载任务
func (s *DownloadService) CreateBatchDownloadTask(userID uint, req *models.BatchDownloadRequest) (*models.DownloadTaskInfo, error) {
	// 验证文件权限
	validFileIDs, err := s.validateFilePermissions(userID, req.FileIDs)
	if err != nil {
		return nil, fmt.Errorf("验证文件权限失败: %v", err)
	}

	if len(validFileIDs) == 0 {
		return nil, fmt.Errorf("没有可下载的文件")
	}

	// 将文件ID列表转换为JSON字符串
	fileIDsJSON, err := json.Marshal(validFileIDs)
	if err != nil {
		return nil, fmt.Errorf("序列化文件ID列表失败: %v", err)
	}

	// 创建下载任务
	task := models.DownloadTask{
		UserID:   userID,
		TaskName: req.TaskName,
		FileIDs:  string(fileIDsJSON),
		Status:   "pending",
	}

	if err := s.db.Create(&task).Error; err != nil {
		return nil, fmt.Errorf("创建下载任务失败: %v", err)
	}

	// 异步处理ZIP打包
	go s.processDownloadTask(task.ID)

	return &models.DownloadTaskInfo{
		ID:        task.ID,
		TaskName:  task.TaskName,
		Status:    task.Status,
		FileCount: len(validFileIDs),
		CreatedAt: task.CreatedAt,
		UpdatedAt: task.UpdatedAt,
	}, nil
}

// validateFilePermissions 验证文件权限
func (s *DownloadService) validateFilePermissions(userID uint, fileIDs []uint) ([]uint, error) {
	var validFileIDs []uint
	var files []models.File

	// 查询文件信息
	if err := s.db.Where("id IN ?", fileIDs).Find(&files).Error; err != nil {
		return nil, err
	}

	for _, file := range files {
		// 检查用户是否有权限访问该文件
		if s.hasFilePermission(userID, &file) {
			validFileIDs = append(validFileIDs, file.ID)
		}
	}

	return validFileIDs, nil
}

// hasFilePermission 检查用户是否有文件权限
func (s *DownloadService) hasFilePermission(userID uint, file *models.File) bool {
	// 如果是文件上传者，有权限
	if file.OwnerID == userID {
		return true
	}

	// 检查用户是否是工作流成员
	var count int64
	s.db.Table("workflow_members").Where("workflow_id = ? AND user_id = ?", file.WorkflowID, userID).Count(&count)
	return count > 0
}

// processDownloadTask 处理下载任务（异步）
func (s *DownloadService) processDownloadTask(taskID uint) {
	// 更新任务状态为处理中
	s.db.Model(&models.DownloadTask{}).Where("id = ?", taskID).Update("status", "processing")

	// 获取任务信息
	var task models.DownloadTask
	if err := s.db.First(&task, taskID).Error; err != nil {
		s.updateTaskError(taskID, fmt.Sprintf("获取任务信息失败: %v", err))
		return
	}

	// 解析文件ID列表
	var fileIDs []uint
	if err := json.Unmarshal([]byte(task.FileIDs), &fileIDs); err != nil {
		s.updateTaskError(taskID, fmt.Sprintf("解析文件ID列表失败: %v", err))
		return
	}

	// 获取文件信息
	var files []models.File
	if err := s.db.Where("id IN ?", fileIDs).Find(&files).Error; err != nil {
		s.updateTaskError(taskID, fmt.Sprintf("获取文件信息失败: %v", err))
		return
	}

	// 创建ZIP文件
	zipFileName := fmt.Sprintf("%s_%d_%d.zip", task.TaskName, task.UserID, time.Now().Unix())
	zipFilePath := filepath.Join(s.downloadPath, zipFileName)

	// 确保下载目录存在
	if err := os.MkdirAll(s.downloadPath, 0755); err != nil {
		s.updateTaskError(taskID, fmt.Sprintf("创建下载目录失败: %v", err))
		return
	}

	// 创建ZIP文件
	zipFile, err := os.Create(zipFilePath)
	if err != nil {
		s.updateTaskError(taskID, fmt.Sprintf("创建ZIP文件失败: %v", err))
		return
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	var totalSize int64

	// 添加文件到ZIP
	for _, file := range files {
		filePath := filepath.Join(s.uploadPath, file.FilePath)
		if err := s.addFileToZip(zipWriter, filePath, file.FileName); err != nil {
			s.updateTaskError(taskID, fmt.Sprintf("添加文件到ZIP失败: %v", err))
			return
		}
		totalSize += file.FileSize
	}

	// 获取ZIP文件大小
	zipInfo, err := zipFile.Stat()
	if err != nil {
		s.updateTaskError(taskID, fmt.Sprintf("获取ZIP文件信息失败: %v", err))
		return
	}

	// 生成下载URL
	downloadURL := fmt.Sprintf("%s/api/v1/download/zip/%s", s.baseURL, zipFileName)

	// 设置过期时间（24小时后）
	expiresAt := time.Now().Add(24 * time.Hour)

	// 更新任务状态为完成
	updates := map[string]interface{}{
		"status":        "completed",
		"zip_file_path": zipFilePath,
		"file_size":     zipInfo.Size(),
		"download_url":  downloadURL,
		"expires_at":    expiresAt,
	}

	s.db.Model(&models.DownloadTask{}).Where("id = ?", taskID).Updates(updates)
}

// addFileToZip 添加文件到ZIP
func (s *DownloadService) addFileToZip(zipWriter *zip.Writer, filePath, fileName string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 创建ZIP文件条目
	zipEntry, err := zipWriter.Create(fileName)
	if err != nil {
		return err
	}

	// 复制文件内容
	_, err = io.Copy(zipEntry, file)
	return err
}

// updateTaskError 更新任务错误状态
func (s *DownloadService) updateTaskError(taskID uint, errorMsg string) {
	updates := map[string]interface{}{
		"status":    "failed",
		"error_msg": errorMsg,
	}
	s.db.Model(&models.DownloadTask{}).Where("id = ?", taskID).Updates(updates)
}

// GetDownloadTask 获取下载任务信息
func (s *DownloadService) GetDownloadTask(userID, taskID uint) (*models.DownloadTaskInfo, error) {
	var task models.DownloadTask
	if err := s.db.Where("id = ? AND user_id = ?", taskID, userID).First(&task).Error; err != nil {
		return nil, fmt.Errorf("下载任务不存在")
	}

	// 解析文件ID列表获取文件数量
	var fileIDs []uint
	fileCount := 0
	if err := json.Unmarshal([]byte(task.FileIDs), &fileIDs); err == nil {
		fileCount = len(fileIDs)
	}

	return &models.DownloadTaskInfo{
		ID:          task.ID,
		TaskName:    task.TaskName,
		Status:      task.Status,
		FileCount:   fileCount,
		FileSize:    task.FileSize,
		DownloadURL: task.DownloadURL,
		ExpiresAt:   task.ExpiresAt,
		ErrorMsg:    task.ErrorMsg,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}, nil
}

// GetUserDownloadTasks 获取用户下载任务列表
func (s *DownloadService) GetUserDownloadTasks(userID uint, page, pageSize int) ([]models.DownloadTaskInfo, int64, error) {
	var tasks []models.DownloadTask
	var total int64

	// 获取总数
	s.db.Model(&models.DownloadTask{}).Where("user_id = ?", userID).Count(&total)

	// 分页查询
	offset := (page - 1) * pageSize
	if err := s.db.Where("user_id = ?", userID).Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&tasks).Error; err != nil {
		return nil, 0, err
	}

	var taskInfos []models.DownloadTaskInfo
	for _, task := range tasks {
		// 解析文件ID列表获取文件数量
		var fileIDs []uint
		fileCount := 0
		if err := json.Unmarshal([]byte(task.FileIDs), &fileIDs); err == nil {
			fileCount = len(fileIDs)
		}

		taskInfos = append(taskInfos, models.DownloadTaskInfo{
			ID:          task.ID,
			TaskName:    task.TaskName,
			Status:      task.Status,
			FileCount:   fileCount,
			FileSize:    task.FileSize,
			DownloadURL: task.DownloadURL,
			ExpiresAt:   task.ExpiresAt,
			ErrorMsg:    task.ErrorMsg,
			CreatedAt:   task.CreatedAt,
			UpdatedAt:   task.UpdatedAt,
		})
	}

	return taskInfos, total, nil
}

// DownloadZipFile 下载ZIP文件
func (s *DownloadService) DownloadZipFile(userID uint, fileName string) (string, error) {
	// 从文件名解析用户ID和任务信息
	parts := strings.Split(fileName, "_")
	if len(parts) < 3 {
		return "", fmt.Errorf("无效的文件名格式")
	}

	// 验证用户ID
	fileUserID, err := strconv.ParseUint(parts[len(parts)-2], 10, 32)
	if err != nil || uint(fileUserID) != userID {
		return "", fmt.Errorf("无权限下载此文件")
	}

	// 构建文件路径
	filePath := filepath.Join(s.downloadPath, fileName)

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", fmt.Errorf("文件不存在或已过期")
	}

	return filePath, nil
}

// LogDownload 记录下载日志
func (s *DownloadService) LogDownload(userID, fileID uint, fileName string, fileSize int64, ipAddress, userAgent string) error {
	log := models.DownloadLog{
		UserID:     userID,
		FileID:     fileID,
		FileName:   fileName,
		FileSize:   fileSize,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		DownloadAt: time.Now(),
	}

	return s.db.Create(&log).Error
}

// GetDownloadStats 获取下载统计
func (s *DownloadService) GetDownloadStats(userID *uint) (*models.DownloadStatsResponse, error) {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	weekStart := todayStart.AddDate(0, 0, -int(now.Weekday()))
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	query := s.db.Model(&models.DownloadLog{})
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}

	stats := &models.DownloadStatsResponse{}

	// 总下载次数和大小
	query.Select("COUNT(*) as count, COALESCE(SUM(file_size), 0) as size").Row().Scan(&stats.TotalDownloads, &stats.TotalDownloadSize)

	// 今日下载次数和大小
	query.Where("download_at >= ?", todayStart).Select("COUNT(*) as count, COALESCE(SUM(file_size), 0) as size").Row().Scan(&stats.TodayDownloads, &stats.TodayDownloadSize)

	// 本周下载次数和大小
	query.Where("download_at >= ?", weekStart).Select("COUNT(*) as count, COALESCE(SUM(file_size), 0) as size").Row().Scan(&stats.WeekDownloads, &stats.WeekDownloadSize)

	// 本月下载次数和大小
	query.Where("download_at >= ?", monthStart).Select("COUNT(*) as count, COALESCE(SUM(file_size), 0) as size").Row().Scan(&stats.MonthDownloads, &stats.MonthDownloadSize)

	// 热门文件（下载次数最多的前10个文件）
	var popularFiles []models.PopularFileInfo
	s.db.Model(&models.DownloadLog{}).
		Select("file_id, file_name, COUNT(*) as download_count, MAX(file_size) as file_size").
		Group("file_id, file_name").
		Order("download_count DESC").
		Limit(10).
		Scan(&popularFiles)

	stats.PopularFiles = popularFiles

	return stats, nil
}

// CleanExpiredTasks 清理过期任务
func (s *DownloadService) CleanExpiredTasks() error {
	now := time.Now()

	// 查找过期任务
	var expiredTasks []models.DownloadTask
	if err := s.db.Where("expires_at < ? AND status = 'completed'", now).Find(&expiredTasks).Error; err != nil {
		return err
	}

	// 删除过期的ZIP文件
	for _, task := range expiredTasks {
		if task.ZipFilePath != "" {
			os.Remove(task.ZipFilePath)
		}
	}

	// 删除过期任务记录
	return s.db.Where("expires_at < ? AND status = 'completed'", now).Delete(&models.DownloadTask{}).Error
}
