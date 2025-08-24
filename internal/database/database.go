package database

import (
	"fmt"
	"log"
	"time"

	"mcs-backend/internal/config"
	"mcs-backend/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDatabase 初始化数据库连接
func InitDatabase(cfg *config.Config) error {
	// 配置GORM日志
	logLevel := logger.Info
	if cfg.Server.Mode == "release" {
		logLevel = logger.Error
	}

	// 连接数据库
	db, err := gorm.Open(postgres.Open(cfg.GetDSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
	})

	if err != nil {
		return err
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	DB = db
	log.Println("Database connected successfully")
	return nil
}

// AutoMigrate 自动迁移数据库表
func AutoMigrate() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	// 迁移所有模型
	err := DB.AutoMigrate(
		// 用户相关
		&models.User{},
		&models.UserGroup{},
		&models.UserGroupMember{},
		&models.InviteCode{},

		// 文件相关
		&models.File{},
		&models.FileVersion{},
		&models.FileShare{},
		&models.Tag{},
		&models.FileTag{},

		// 工作流相关
		&models.Workflow{},
		&models.WorkflowMember{},
		&models.Task{},
		&models.TaskMember{},
		&models.TaskStatusLog{},
		&models.TaskStagingArea{},
		&models.TaskSubmission{},
		&models.TaskTemplate{},
		&models.TaskSubmissionAnnotation{},

		// 通知相关
		&models.Notification{},
		&models.NotificationSetting{},

		// 统计和日志
		&models.Statistics{},
		&models.ActivityLog{},
		&models.UserFavorite{},
		&models.PopularityScore{},
		&models.ReportData{},
	)

	if err != nil {
		return err
	}

	log.Println("Database migration completed successfully")
	return nil
}

// CreateIndexes 创建数据库索引
func CreateIndexes() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	// 创建复合索引
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_users_email_deleted ON users(email) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_files_workflow_folder ON files(workflow_id, folder_id) WHERE is_deleted = false",
		"CREATE INDEX IF NOT EXISTS idx_tasks_workflow_status ON tasks(workflow_id, status) WHERE is_deleted = false",
		"CREATE INDEX IF NOT EXISTS idx_activity_logs_user_action ON activity_logs(user_id, action, created_at)",
		"CREATE INDEX IF NOT EXISTS idx_statistics_target ON statistics(target_type, target_id, date)",
		"CREATE INDEX IF NOT EXISTS idx_notifications_user_read ON notifications(user_id, is_read, created_at)",
	}

	for _, index := range indexes {
		if err := DB.Exec(index).Error; err != nil {
			log.Printf("Warning: Failed to create index: %v", err)
		}
	}

	log.Println("Database indexes created successfully")
	return nil
}

// SeedData 初始化种子数据
func SeedData() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	// 检查是否已有管理员用户
	var adminCount int64
	DB.Model(&models.User{}).Where("role = ?", "super_admin").Count(&adminCount)

	if adminCount == 0 {
		// 创建默认超级管理员
		admin := &models.User{
			Username:     "admin",
			Email:        "admin@mcs.local",
			PasswordHash: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // password
			RealName:     "系统管理员",
			Role:         "super_admin",
			Status:       "active",
		}

		if err := DB.Create(admin).Error; err != nil {
			return err
		}

		// 创建默认邀请码
		inviteCode := &models.InviteCode{
			Code:        "WELCOME2024",
			CreatedBy:   admin.ID,
			MaxUses:     100,
			UsedCount:   0,
			Description: "默认邀请码",
			Status:      "active",
		}

		if err := DB.Create(inviteCode).Error; err != nil {
			return err
		}

		// 创建默认用户组
		groups := []*models.UserGroup{
			{
				Name:        "摄影组",
				Description: "负责拍摄工作的团队",
				CreaterID:   admin.ID,
				IsActive:    true,
			},
			{
				Name:        "设计组",
				Description: "负责设计工作的团队",
				CreaterID:   admin.ID,
				IsActive:    true,
			},
			{
				Name:        "后期组",
				Description: "负责后期制作的团队",
				CreaterID:   admin.ID,
				IsActive:    true,
			},
		}

		for _, group := range groups {
			if err := DB.Create(group).Error; err != nil {
				return err
			}
		}

		// 创建默认标签
		tags := []*models.Tag{
			{TagName: "人像", Color: "#FF6B6B", CreaterID: admin.ID},
			{TagName: "风景", Color: "#4ECDC4", CreaterID: admin.ID},
			{TagName: "产品", Color: "#45B7D1", CreaterID: admin.ID},
			{TagName: "活动", Color: "#96CEB4", CreaterID: admin.ID},
			{TagName: "建筑", Color: "#FFEAA7", CreaterID: admin.ID},
			{TagName: "街拍", Color: "#DDA0DD", CreaterID: admin.ID},
		}

		for _, tag := range tags {
			if err := DB.Create(tag).Error; err != nil {
				return err
			}
		}

		log.Println("Seed data created successfully")
		log.Println("Default admin credentials: admin@mcs.local / password")
		log.Println("Default invite code: WELCOME2024")
	} else {
		log.Println("Seed data already exists, skipping...")
	}

	return nil
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}

// CloseDB 关闭数据库连接
func CloseDB() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}