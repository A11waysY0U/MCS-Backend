package services

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"mcs-backend/internal/config"
	"mcs-backend/internal/database"
	"mcs-backend/internal/models"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"
)

// UploadService 文件上传服务
type UploadService struct {
	db     *gorm.DB
	config *config.Config
}

// NewUploadService 创建文件上传服务
func NewUploadService(cfg *config.Config) *UploadService {
	return &UploadService{
		db:     database.GetDB(),
		config: cfg,
	}
}

// ChunkUploadRequest 分片上传请求
type ChunkUploadRequest struct {
	FileName    string `json:"file_name" binding:"required"`
	FileSize    int64  `json:"file_size" binding:"required"`
	MD5Hash     string `json:"md5_hash" binding:"required"`
	ChunkIndex  int    `json:"chunk_index" binding:"required"`
	ChunkSize   int64  `json:"chunk_size" binding:"required"`
	TotalChunks int    `json:"total_chunks" binding:"required"`
	ChunkMD5    string `json:"chunk_md5" binding:"required"`
	FolderID    uint   `json:"folder_id"`
	WorkflowID  uint   `json:"workflow_id"`
	TaskID      uint   `json:"task_id"`
	Description string `json:"description"`
	IsPrivate   bool   `json:"is_private"`
}

// InitUploadRequest 初始化上传请求
type InitUploadRequest struct {
	FileName    string `json:"file_name" binding:"required"`
	FileSize    int64  `json:"file_size" binding:"required"`
	MD5Hash     string `json:"md5_hash" binding:"required"`
	ChunkSize   int64  `json:"chunk_size" binding:"required"`
	FolderID    uint   `json:"folder_id"`
	WorkflowID  uint   `json:"workflow_id"`
	TaskID      uint   `json:"task_id"`
	Description string `json:"description"`
	IsPrivate   bool   `json:"is_private"`
}

// InitUploadResponse 初始化上传响应
type InitUploadResponse struct {
	UploadID     string `json:"upload_id"`
	TotalChunks  int    `json:"total_chunks"`
	ChunkSize    int64  `json:"chunk_size"`
	ExistingFile *File  `json:"existing_file,omitempty"` // 秒传时返回已存在的文件
	IsSecUpload  bool   `json:"is_sec_upload"`           // 是否为秒传
}

// File 文件信息
type File struct {
	ID          uint      `json:"id"`
	FileName    string    `json:"file_name"`
	FilePath    string    `json:"file_path"`
	FileSize    int64     `json:"file_size"`
	MD5Hash     string    `json:"md5_hash"`
	MimeType    string    `json:"mime_type"`
	OwnerID     uint      `json:"owner_id"`
	FolderID    uint      `json:"folder_id"`
	WorkflowID  uint      `json:"workflow_id"`
	TaskID      uint      `json:"task_id"`
	IsPrivate   bool      `json:"is_private"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UploadProgress 上传进度
type UploadProgress struct {
	UploadID        string `json:"upload_id"`
	FileName        string `json:"file_name"`
	FileSize        int64  `json:"file_size"`
	UploadedSize    int64  `json:"uploaded_size"`
	TotalChunks     int    `json:"total_chunks"`
	UploadedChunks  int    `json:"uploaded_chunks"`
	Progress        float64 `json:"progress"`
	Status          string `json:"status"` // uploading, completed, failed
	CreatedAt       time.Time `json:"created_at"`
}

// UploadSession 上传会话（存储在内存或Redis中）
type UploadSession struct {
	UploadID     string
	FileName     string
	FileSize     int64
	MD5Hash      string
	ChunkSize    int64
	TotalChunks  int
	UploadedChunks map[int]bool
	UserID       uint
	FolderID     uint
	WorkflowID   uint
	TaskID       uint
	Description  string
	IsPrivate    bool
	CreatedAt    time.Time
	TempDir      string
}

// 内存中的上传会话存储
var uploadSessions = make(map[string]*UploadSession)

// InitUpload 初始化上传
func (s *UploadService) InitUpload(req *InitUploadRequest, userID uint) (*InitUploadResponse, error) {
	// 检查是否可以秒传
	var existingFile models.File
	if err := s.db.Where("md5_hash = ? AND file_size = ? AND is_deleted = false", req.MD5Hash, req.FileSize).First(&existingFile).Error; err == nil {
		// 秒传：复制文件记录
		newFile := models.File{
			FileName:    req.FileName,
			FilePath:    existingFile.FilePath,
			FileSize:    existingFile.FileSize,
			MD5Hash:     existingFile.MD5Hash,
			MimeType:    existingFile.MimeType,
			OwnerID:     userID,
			FolderID:    req.FolderID,
			WorkflowID:  req.WorkflowID,
			TaskID:      req.TaskID,
			IsPrivate:   req.IsPrivate,
			Description: req.Description,
		}

		if err := s.db.Create(&newFile).Error; err != nil {
			return nil, fmt.Errorf("创建文件记录失败: %v", err)
		}

		return &InitUploadResponse{
			ExistingFile: &File{
				ID:          newFile.ID,
				FileName:    newFile.FileName,
				FilePath:    newFile.FilePath,
				FileSize:    newFile.FileSize,
				MD5Hash:     newFile.MD5Hash,
				MimeType:    newFile.MimeType,
				OwnerID:     newFile.OwnerID,
				FolderID:    newFile.FolderID,
				WorkflowID:  newFile.WorkflowID,
				TaskID:      newFile.TaskID,
				IsPrivate:   newFile.IsPrivate,
				Description: newFile.Description,
				CreatedAt:   newFile.CreatedAt,
				UpdatedAt:   newFile.UpdatedAt,
			},
			IsSecUpload: true,
		}, nil
	}

	// 计算分片数量
	totalChunks := int((req.FileSize + req.ChunkSize - 1) / req.ChunkSize)

	// 生成上传ID
	uploadID := fmt.Sprintf("%d_%s_%d", userID, req.MD5Hash, time.Now().Unix())

	// 创建临时目录
	tempDir := filepath.Join(s.config.File.TempPath, uploadID)
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, fmt.Errorf("创建临时目录失败: %v", err)
	}

	// 创建上传会话
	session := &UploadSession{
		UploadID:       uploadID,
		FileName:       req.FileName,
		FileSize:       req.FileSize,
		MD5Hash:        req.MD5Hash,
		ChunkSize:      req.ChunkSize,
		TotalChunks:    totalChunks,
		UploadedChunks: make(map[int]bool),
		UserID:         userID,
		FolderID:       req.FolderID,
		WorkflowID:     req.WorkflowID,
		TaskID:         req.TaskID,
		Description:    req.Description,
		IsPrivate:      req.IsPrivate,
		CreatedAt:      time.Now(),
		TempDir:        tempDir,
	}

	uploadSessions[uploadID] = session

	return &InitUploadResponse{
		UploadID:    uploadID,
		TotalChunks: totalChunks,
		ChunkSize:   req.ChunkSize,
		IsSecUpload: false,
	}, nil
}

// UploadChunk 上传分片
func (s *UploadService) UploadChunk(uploadID string, chunkIndex int, chunkData []byte, chunkMD5 string) error {
	session, exists := uploadSessions[uploadID]
	if !exists {
		return errors.New("上传会话不存在")
	}

	// 验证分片MD5
	hash := md5.Sum(chunkData)
	if fmt.Sprintf("%x", hash) != chunkMD5 {
		return errors.New("分片MD5校验失败")
	}

	// 保存分片文件
	chunkPath := filepath.Join(session.TempDir, fmt.Sprintf("chunk_%d", chunkIndex))
	file, err := os.Create(chunkPath)
	if err != nil {
		return fmt.Errorf("创建分片文件失败: %v", err)
	}
	defer file.Close()

	if _, err := file.Write(chunkData); err != nil {
		return fmt.Errorf("写入分片文件失败: %v", err)
	}

	// 标记分片已上传
	session.UploadedChunks[chunkIndex] = true

	return nil
}

// CompleteUpload 完成上传
func (s *UploadService) CompleteUpload(uploadID string) (*File, error) {
	session, exists := uploadSessions[uploadID]
	if !exists {
		return nil, errors.New("上传会话不存在")
	}

	// 检查所有分片是否都已上传
	for i := 0; i < session.TotalChunks; i++ {
		if !session.UploadedChunks[i] {
			return nil, fmt.Errorf("分片 %d 未上传", i)
		}
	}

	// 合并分片
	finalPath := filepath.Join(s.config.File.UploadPath, fmt.Sprintf("%s_%s", session.MD5Hash, session.FileName))
	if err := os.MkdirAll(filepath.Dir(finalPath), 0755); err != nil {
		return nil, fmt.Errorf("创建目标目录失败: %v", err)
	}

	finalFile, err := os.Create(finalPath)
	if err != nil {
		return nil, fmt.Errorf("创建最终文件失败: %v", err)
	}
	defer finalFile.Close()

	// 按顺序合并分片
	for i := 0; i < session.TotalChunks; i++ {
		chunkPath := filepath.Join(session.TempDir, fmt.Sprintf("chunk_%d", i))
		chunkFile, err := os.Open(chunkPath)
		if err != nil {
			return nil, fmt.Errorf("打开分片文件失败: %v", err)
		}

		if _, err := io.Copy(finalFile, chunkFile); err != nil {
			chunkFile.Close()
			return nil, fmt.Errorf("合并分片失败: %v", err)
		}
		chunkFile.Close()
	}

	// 验证最终文件MD5
	finalFile.Seek(0, 0)
	hash := md5.New()
	if _, err := io.Copy(hash, finalFile); err != nil {
		return nil, fmt.Errorf("计算文件MD5失败: %v", err)
	}

	if fmt.Sprintf("%x", hash.Sum(nil)) != session.MD5Hash {
		return nil, errors.New("文件MD5校验失败")
	}

	// 获取文件MIME类型
	mimeType := getMimeType(session.FileName)

	// 创建文件记录
	file := models.File{
		FileName:    session.FileName,
		FilePath:    finalPath,
		FileSize:    session.FileSize,
		MD5Hash:     session.MD5Hash,
		MimeType:    mimeType,
		OwnerID:     session.UserID,
		FolderID:    session.FolderID,
		WorkflowID:  session.WorkflowID,
		TaskID:      session.TaskID,
		IsPrivate:   session.IsPrivate,
		Description: session.Description,
	}

	if err := s.db.Create(&file).Error; err != nil {
		return nil, fmt.Errorf("创建文件记录失败: %v", err)
	}

	// 清理临时文件和会话
	os.RemoveAll(session.TempDir)
	delete(uploadSessions, uploadID)

	return &File{
		ID:          file.ID,
		FileName:    file.FileName,
		FilePath:    file.FilePath,
		FileSize:    file.FileSize,
		MD5Hash:     file.MD5Hash,
		MimeType:    file.MimeType,
		OwnerID:     file.OwnerID,
		FolderID:    file.FolderID,
		WorkflowID:  file.WorkflowID,
		TaskID:      file.TaskID,
		IsPrivate:   file.IsPrivate,
		Description: file.Description,
		CreatedAt:   file.CreatedAt,
		UpdatedAt:   file.UpdatedAt,
	}, nil
}

// GetUploadProgress 获取上传进度
func (s *UploadService) GetUploadProgress(uploadID string) (*UploadProgress, error) {
	session, exists := uploadSessions[uploadID]
	if !exists {
		return nil, errors.New("上传会话不存在")
	}

	uploadedChunks := 0
	uploadedSize := int64(0)
	for i := 0; i < session.TotalChunks; i++ {
		if session.UploadedChunks[i] {
			uploadedChunks++
			if i == session.TotalChunks-1 {
				// 最后一个分片可能不是完整大小
				uploadedSize += session.FileSize - int64(i)*session.ChunkSize
			} else {
				uploadedSize += session.ChunkSize
			}
		}
	}

	progress := float64(uploadedChunks) / float64(session.TotalChunks) * 100
	status := "uploading"
	if uploadedChunks == session.TotalChunks {
		status = "completed"
	}

	return &UploadProgress{
		UploadID:       uploadID,
		FileName:       session.FileName,
		FileSize:       session.FileSize,
		UploadedSize:   uploadedSize,
		TotalChunks:    session.TotalChunks,
		UploadedChunks: uploadedChunks,
		Progress:       progress,
		Status:         status,
		CreatedAt:      session.CreatedAt,
	}, nil
}

// CancelUpload 取消上传
func (s *UploadService) CancelUpload(uploadID string) error {
	session, exists := uploadSessions[uploadID]
	if !exists {
		return errors.New("上传会话不存在")
	}

	// 清理临时文件和会话
	os.RemoveAll(session.TempDir)
	delete(uploadSessions, uploadID)

	return nil
}

// getMimeType 根据文件扩展名获取MIME类型
func getMimeType(fileName string) string {
	ext := strings.ToLower(filepath.Ext(fileName))
	mimeTypes := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".pdf":  "application/pdf",
		".doc":  "application/msword",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".xls":  "application/vnd.ms-excel",
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".ppt":  "application/vnd.ms-powerpoint",
		".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
		".txt":  "text/plain",
		".zip":  "application/zip",
		".rar":  "application/x-rar-compressed",
		".mp4":  "video/mp4",
		".avi":  "video/x-msvideo",
		".mp3":  "audio/mpeg",
		".wav":  "audio/wav",
	}

	if mimeType, exists := mimeTypes[ext]; exists {
		return mimeType
	}
	return "application/octet-stream"
}