package services

import (
	"errors"
	"fmt"
	"mcs-backend/internal/config"
	"mcs-backend/internal/database"
	"mcs-backend/internal/models"
	"path/filepath"
	"time"

	"gorm.io/gorm"
)

// FileService 文件管理服务
type FileService struct {
	db     *gorm.DB
	config *config.Config
}

// NewFileService 创建文件管理服务
func NewFileService(cfg *config.Config) *FileService {
	return &FileService{
		db:     database.GetDB(),
		config: cfg,
	}
}

// GetConfig 获取配置
func (s *FileService) GetConfig() *config.Config {
	return s.config
}

// CreateFolderRequest 创建文件夹请求
type CreateFolderRequest struct {
	Name        string `json:"name" binding:"required"`
	ParentID    uint   `json:"parent_id"`
	WorkflowID  uint   `json:"workflow_id" binding:"required"`
	Description string `json:"description"`
}

// UpdateFolderRequest 更新文件夹请求
type UpdateFolderRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	SortOrder   uint   `json:"sort_order"`
}

// FileListRequest 文件列表请求
type FileListRequest struct {
	FolderID   uint   `json:"folder_id"`
	WorkflowID uint   `json:"workflow_id"`
	TaskID     uint   `json:"task_id"`
	Keyword    string `json:"keyword"`
	MimeType   string `json:"mime_type"`
	Page       int    `json:"page"`
	PageSize   int    `json:"page_size"`
	SortBy     string `json:"sort_by"`     // name, size, created_at
	SortOrder  string `json:"sort_order"`  // asc, desc
}

// FileListResponse 文件列表响应
type FileListResponse struct {
	Files   []FileInfo `json:"files"`
	Folders []FolderInfo `json:"folders"`
	Total   int64      `json:"total"`
	Page    int        `json:"page"`
	PageSize int       `json:"page_size"`
}

// FileInfo 文件信息
type FileInfo struct {
	ID          uint      `json:"id"`
	FileName    string    `json:"file_name"`
	FileSize    int64     `json:"file_size"`
	MimeType    string    `json:"mime_type"`
	OwnerID     uint      `json:"owner_id"`
	OwnerName   string    `json:"owner_name"`
	FolderID    uint      `json:"folder_id"`
	WorkflowID  uint      `json:"workflow_id"`
	TaskID      uint      `json:"task_id"`
	IsPrivate   bool      `json:"is_private"`
	Description string    `json:"description"`
	Tags        []TagInfo `json:"tags"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// FolderInfo 文件夹信息
type FolderInfo struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	ParentID    uint      `json:"parent_id"`
	WorkflowID  uint      `json:"workflow_id"`
	CreatorID   uint      `json:"creator_id"`
	CreatorName string    `json:"creator_name"`
	Description string    `json:"description"`
	SortOrder   uint      `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TagInfo 标签信息
type TagInfo struct {
	ID      uint   `json:"id"`
	TagName string `json:"tag_name"`
	Color   string `json:"color"`
}

// UpdateFileRequest 更新文件请求
type UpdateFileRequest struct {
	FileName    string `json:"file_name"`
	FolderID    uint   `json:"folder_id"`
	Description string `json:"description"`
	IsPrivate   bool   `json:"is_private"`
	TagIDs      []uint `json:"tag_ids"`
}

// CreateFolder 创建文件夹
func (s *FileService) CreateFolder(req *CreateFolderRequest, userID uint) (*FolderInfo, error) {
	// 构建文件夹路径
	var path string
	if req.ParentID > 0 {
		var parentFolder models.FileFolder
		if err := s.db.Where("id = ? AND is_deleted = false", req.ParentID).First(&parentFolder).Error; err != nil {
			return nil, fmt.Errorf("父文件夹不存在")
		}
		path = filepath.Join(parentFolder.Path, req.Name)
	} else {
		path = req.Name
	}

	// 检查同级目录下是否存在同名文件夹
	var count int64
	s.db.Model(&models.FileFolder{}).Where("name = ? AND parent_id = ? AND workflow_id = ? AND is_deleted = false", req.Name, req.ParentID, req.WorkflowID).Count(&count)
	if count > 0 {
		return nil, errors.New("同级目录下已存在同名文件夹")
	}

	folder := models.FileFolder{
		Name:        req.Name,
		Path:        path,
		ParentID:    req.ParentID,
		WorkflowID:  req.WorkflowID,
		CreatorID:   userID,
		Description: req.Description,
	}

	if err := s.db.Create(&folder).Error; err != nil {
		return nil, fmt.Errorf("创建文件夹失败: %v", err)
	}

	return &FolderInfo{
		ID:          folder.ID,
		Name:        folder.Name,
		Path:        folder.Path,
		ParentID:    folder.ParentID,
		WorkflowID:  folder.WorkflowID,
		CreatorID:   folder.CreatorID,
		Description: folder.Description,
		SortOrder:   folder.SortOrder,
		CreatedAt:   folder.CreatedAt,
		UpdatedAt:   folder.UpdatedAt,
	}, nil
}

// GetFileList 获取文件列表
func (s *FileService) GetFileList(req *FileListRequest, userID uint) (*FileListResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.SortBy == "" {
		req.SortBy = "created_at"
	}
	if req.SortOrder == "" {
		req.SortOrder = "desc"
	}

	offset := (req.Page - 1) * req.PageSize

	// 构建查询条件
	fileQuery := s.db.Model(&models.File{}).Where("is_deleted = false")
	folderQuery := s.db.Model(&models.FileFolder{}).Where("is_deleted = false")

	if req.FolderID > 0 {
		fileQuery = fileQuery.Where("folder_id = ?", req.FolderID)
		folderQuery = folderQuery.Where("parent_id = ?", req.FolderID)
	}
	if req.WorkflowID > 0 {
		fileQuery = fileQuery.Where("workflow_id = ?", req.WorkflowID)
		folderQuery = folderQuery.Where("workflow_id = ?", req.WorkflowID)
	}
	if req.TaskID > 0 {
		fileQuery = fileQuery.Where("task_id = ?", req.TaskID)
	}
	if req.Keyword != "" {
		keyword := "%" + req.Keyword + "%"
		fileQuery = fileQuery.Where("file_name ILIKE ?", keyword)
		folderQuery = folderQuery.Where("name ILIKE ?", keyword)
	}
	if req.MimeType != "" {
		fileQuery = fileQuery.Where("mime_type LIKE ?", req.MimeType+"%")
	}

	// 权限过滤：只能看到自己的私有文件或公开文件
	fileQuery = fileQuery.Where("(is_private = false OR owner_id = ?)", userID)

	// 获取文件总数
	var fileTotal int64
	fileQuery.Count(&fileTotal)

	// 获取文件夹总数
	var folderTotal int64
	folderQuery.Count(&folderTotal)

	// 获取文件列表
	var files []models.File
	fileQuery = fileQuery.Preload("Owner").Order(fmt.Sprintf("%s %s", req.SortBy, req.SortOrder))
	if err := fileQuery.Offset(offset).Limit(req.PageSize).Find(&files).Error; err != nil {
		return nil, fmt.Errorf("获取文件列表失败: %v", err)
	}

	// 获取文件夹列表
	var folders []models.FileFolder
	folderQuery = folderQuery.Preload("Creator").Order("sort_order ASC, created_at DESC")
	if err := folderQuery.Find(&folders).Error; err != nil {
		return nil, fmt.Errorf("获取文件夹列表失败: %v", err)
	}

	// 转换为响应格式
	fileInfos := make([]FileInfo, len(files))
	for i, file := range files {
		// 获取文件标签
		tags := s.getFileTags(file.ID)
		
		fileInfos[i] = FileInfo{
			ID:          file.ID,
			FileName:    file.FileName,
			FileSize:    file.FileSize,
			MimeType:    file.MimeType,
			OwnerID:     file.OwnerID,
			FolderID:    file.FolderID,
			WorkflowID:  file.WorkflowID,
			TaskID:      file.TaskID,
			IsPrivate:   file.IsPrivate,
			Description: file.Description,
			Tags:        tags,
			CreatedAt:   file.CreatedAt,
			UpdatedAt:   file.UpdatedAt,
		}
	}

	folderInfos := make([]FolderInfo, len(folders))
	for i, folder := range folders {
		folderInfos[i] = FolderInfo{
			ID:          folder.ID,
			Name:        folder.Name,
			Path:        folder.Path,
			ParentID:    folder.ParentID,
			WorkflowID:  folder.WorkflowID,
			CreatorID:   folder.CreatorID,
			Description: folder.Description,
			SortOrder:   folder.SortOrder,
			CreatedAt:   folder.CreatedAt,
			UpdatedAt:   folder.UpdatedAt,
		}
	}

	return &FileListResponse{
		Files:    fileInfos,
		Folders:  folderInfos,
		Total:    fileTotal + folderTotal,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// GetFileByID 根据ID获取文件信息
func (s *FileService) GetFileByID(fileID uint, userID uint) (*FileInfo, error) {
	var file models.File
	query := s.db.Where("id = ? AND is_deleted = false", fileID)
	
	// 权限检查：只能访问自己的私有文件或公开文件
	query = query.Where("(is_private = false OR owner_id = ?)", userID)
	
	if err := query.Preload("Owner").First(&file).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("文件不存在或无权限访问")
		}
		return nil, fmt.Errorf("获取文件信息失败: %v", err)
	}

	// 获取文件标签
	tags := s.getFileTags(file.ID)

	return &FileInfo{
		ID:          file.ID,
		FileName:    file.FileName,
		FileSize:    file.FileSize,
		MimeType:    file.MimeType,
		OwnerID:     file.OwnerID,
		FolderID:    file.FolderID,
		WorkflowID:  file.WorkflowID,
		TaskID:      file.TaskID,
		IsPrivate:   file.IsPrivate,
		Description: file.Description,
		Tags:        tags,
		CreatedAt:   file.CreatedAt,
		UpdatedAt:   file.UpdatedAt,
	}, nil
}

// UpdateFile 更新文件信息
func (s *FileService) UpdateFile(fileID uint, req *UpdateFileRequest, userID uint) (*FileInfo, error) {
	// 检查文件是否存在且用户有权限
	var file models.File
	if err := s.db.Where("id = ? AND owner_id = ? AND is_deleted = false", fileID, userID).First(&file).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("文件不存在或无权限修改")
		}
		return nil, fmt.Errorf("获取文件信息失败: %v", err)
	}

	// 更新文件信息
	updates := make(map[string]interface{})
	if req.FileName != "" {
		updates["file_name"] = req.FileName
	}
	if req.FolderID > 0 {
		updates["folder_id"] = req.FolderID
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	updates["is_private"] = req.IsPrivate

	if err := s.db.Model(&file).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("更新文件信息失败: %v", err)
	}

	// 更新文件标签
	if len(req.TagIDs) > 0 {
		if err := s.updateFileTags(fileID, req.TagIDs); err != nil {
			return nil, fmt.Errorf("更新文件标签失败: %v", err)
		}
	}

	// 重新获取更新后的文件信息
	return s.GetFileByID(fileID, userID)
}

// DeleteFile 删除文件
func (s *FileService) DeleteFile(fileID uint, userID uint) error {
	// 检查文件是否存在且用户有权限
	var file models.File
	if err := s.db.Where("id = ? AND owner_id = ? AND is_deleted = false", fileID, userID).First(&file).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("文件不存在或无权限删除")
		}
		return fmt.Errorf("获取文件信息失败: %v", err)
	}

	// 软删除文件记录
	now := time.Now()
	if err := s.db.Model(&file).Updates(map[string]interface{}{
		"is_deleted": true,
		"deleted_at": &now,
	}).Error; err != nil {
		return fmt.Errorf("删除文件记录失败: %v", err)
	}

	return nil
}

// DeleteFolder 删除文件夹
func (s *FileService) DeleteFolder(folderID uint, userID uint) error {
	// 检查文件夹是否存在且用户有权限
	var folder models.FileFolder
	if err := s.db.Where("id = ? AND creator_id = ? AND is_deleted = false", folderID, userID).First(&folder).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("文件夹不存在或无权限删除")
		}
		return fmt.Errorf("获取文件夹信息失败: %v", err)
	}

	// 检查文件夹是否为空
	var fileCount, subFolderCount int64
	s.db.Model(&models.File{}).Where("folder_id = ? AND is_deleted = false", folderID).Count(&fileCount)
	s.db.Model(&models.FileFolder{}).Where("parent_id = ? AND is_deleted = false", folderID).Count(&subFolderCount)

	if fileCount > 0 || subFolderCount > 0 {
		return errors.New("文件夹不为空，无法删除")
	}

	// 软删除文件夹
	now := time.Now()
	if err := s.db.Model(&folder).Updates(map[string]interface{}{
		"is_deleted": true,
		"deleted_at": &now,
	}).Error; err != nil {
		return fmt.Errorf("删除文件夹失败: %v", err)
	}

	return nil
}

// UpdateFolder 更新文件夹
func (s *FileService) UpdateFolder(folderID uint, req *UpdateFolderRequest, userID uint) (*FolderInfo, error) {
	// 检查文件夹是否存在且用户有权限
	var folder models.FileFolder
	if err := s.db.Where("id = ? AND creator_id = ? AND is_deleted = false", folderID, userID).First(&folder).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("文件夹不存在或无权限修改")
		}
		return nil, fmt.Errorf("获取文件夹信息失败: %v", err)
	}

	// 更新文件夹信息
	updates := make(map[string]interface{})
	if req.Name != "" {
		// 检查同级目录下是否存在同名文件夹
		var count int64
		s.db.Model(&models.FileFolder{}).Where("name = ? AND parent_id = ? AND workflow_id = ? AND id != ? AND is_deleted = false", req.Name, folder.ParentID, folder.WorkflowID, folderID).Count(&count)
		if count > 0 {
			return nil, errors.New("同级目录下已存在同名文件夹")
		}
		updates["name"] = req.Name
		
		// 更新路径
		var newPath string
		if folder.ParentID > 0 {
			var parentFolder models.FileFolder
			s.db.Where("id = ?", folder.ParentID).First(&parentFolder)
			newPath = filepath.Join(parentFolder.Path, req.Name)
		} else {
			newPath = req.Name
		}
		updates["path"] = newPath
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.SortOrder > 0 {
		updates["sort_order"] = req.SortOrder
	}

	if err := s.db.Model(&folder).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("更新文件夹失败: %v", err)
	}

	// 重新获取更新后的文件夹信息
	s.db.Where("id = ?", folderID).First(&folder)

	return &FolderInfo{
		ID:          folder.ID,
		Name:        folder.Name,
		Path:        folder.Path,
		ParentID:    folder.ParentID,
		WorkflowID:  folder.WorkflowID,
		CreatorID:   folder.CreatorID,
		Description: folder.Description,
		SortOrder:   folder.SortOrder,
		CreatedAt:   folder.CreatedAt,
		UpdatedAt:   folder.UpdatedAt,
	}, nil
}

// SearchFiles 搜索文件
func (s *FileService) SearchFiles(keyword string, userID uint, page, pageSize int) (*FileListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	keywordPattern := "%" + keyword + "%"

	// 搜索文件
	var files []models.File
	fileQuery := s.db.Where("is_deleted = false AND (is_private = false OR owner_id = ?) AND (file_name ILIKE ? OR description ILIKE ?)", userID, keywordPattern, keywordPattern)
	
	var total int64
	fileQuery.Count(&total)
	
	if err := fileQuery.Preload("Owner").Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&files).Error; err != nil {
		return nil, fmt.Errorf("搜索文件失败: %v", err)
	}

	// 转换为响应格式
	fileInfos := make([]FileInfo, len(files))
	for i, file := range files {
		tags := s.getFileTags(file.ID)
		
		fileInfos[i] = FileInfo{
			ID:          file.ID,
			FileName:    file.FileName,
			FileSize:    file.FileSize,
			MimeType:    file.MimeType,
			OwnerID:     file.OwnerID,
			FolderID:    file.FolderID,
			WorkflowID:  file.WorkflowID,
			TaskID:      file.TaskID,
			IsPrivate:   file.IsPrivate,
			Description: file.Description,
			Tags:        tags,
			CreatedAt:   file.CreatedAt,
			UpdatedAt:   file.UpdatedAt,
		}
	}

	return &FileListResponse{
		Files:    fileInfos,
		Folders:  []FolderInfo{},
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// getFileTags 获取文件标签
func (s *FileService) getFileTags(fileID uint) []TagInfo {
	var fileTags []models.FileTag
	s.db.Where("file_id = ?", fileID).Find(&fileTags)

	if len(fileTags) == 0 {
		return []TagInfo{}
	}

	tagIDs := make([]uint, len(fileTags))
	for i, ft := range fileTags {
		tagIDs[i] = ft.TagID
	}

	var tags []models.Tag
	s.db.Where("id IN ?", tagIDs).Find(&tags)

	tagInfos := make([]TagInfo, len(tags))
	for i, tag := range tags {
		tagInfos[i] = TagInfo{
			ID:      tag.ID,
			TagName: tag.TagName,
			Color:   tag.Color,
		}
	}

	return tagInfos
}

// updateFileTags 更新文件标签
func (s *FileService) updateFileTags(fileID uint, tagIDs []uint) error {
	// 删除现有标签关联
	if err := s.db.Where("file_id = ?", fileID).Delete(&models.FileTag{}).Error; err != nil {
		return err
	}

	// 添加新的标签关联
	for _, tagID := range tagIDs {
		fileTag := models.FileTag{
			FileID: fileID,
			TagID:  tagID,
		}
		if err := s.db.Create(&fileTag).Error; err != nil {
			return err
		}
	}

	return nil
}

// CreateFileVersion 创建文件版本
func (s *FileService) CreateFileVersion(fileID uint, filePath string, fileSize int64, md5Hash string, changeLog string, userID uint) error {
	// 检查文件是否存在
	var file models.File
	if err := s.db.Where("id = ? AND is_deleted = false", fileID).First(&file).Error; err != nil {
		return fmt.Errorf("文件不存在: %v", err)
	}

	// 获取当前最大版本号
	var maxVersion uint
	s.db.Model(&models.FileVersion{}).Where("file_id = ?", fileID).Select("COALESCE(MAX(version), 0)").Scan(&maxVersion)

	// 创建新版本记录
	version := models.FileVersion{
		FileID:    fileID,
		Version:   maxVersion + 1,
		FilePath:  filePath,
		FileSize:  fileSize,
		MD5Hash:   md5Hash,
		CreatedBy: userID,
		ChangeLog: changeLog,
		IsActive:  true,
	}

	// 将之前的版本设为非活跃
	if err := s.db.Model(&models.FileVersion{}).Where("file_id = ? AND is_active = true", fileID).Update("is_active", false).Error; err != nil {
		return fmt.Errorf("更新版本状态失败: %v", err)
	}

	if err := s.db.Create(&version).Error; err != nil {
		return fmt.Errorf("创建文件版本失败: %v", err)
	}

	// 更新文件主记录
	if err := s.db.Model(&file).Updates(map[string]interface{}{
		"file_path": filePath,
		"file_size": fileSize,
		"md5_hash":  md5Hash,
	}).Error; err != nil {
		return fmt.Errorf("更新文件记录失败: %v", err)
	}

	return nil
}

// GetFileVersions 获取文件版本列表
func (s *FileService) GetFileVersions(fileID uint, userID uint) ([]models.FileVersion, error) {
	// 检查用户是否有权限访问该文件
	var file models.File
	if err := s.db.Where("id = ? AND (is_private = false OR owner_id = ?) AND is_deleted = false", fileID, userID).First(&file).Error; err != nil {
		return nil, fmt.Errorf("文件不存在或无权限访问: %v", err)
	}

	var versions []models.FileVersion
	if err := s.db.Where("file_id = ?", fileID).Order("version DESC").Find(&versions).Error; err != nil {
		return nil, fmt.Errorf("获取文件版本失败: %v", err)
	}

	return versions, nil
}