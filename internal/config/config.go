package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config 应用配置结构
type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	JWT      JWTConfig      `json:"jwt"`
	File     FileConfig     `json:"file"`
	Redis    RedisConfig    `json:"redis"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port    string `json:"port"`
	Mode    string `json:"mode"` // debug, release, test
	BaseURL string `json:"base_url"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"db_name"`
	SSLMode  string `json:"ssl_mode"`
	TimeZone string `json:"time_zone"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	SecretKey       string `json:"secret_key"`
	ExpirationHours int    `json:"expiration_hours"`
}

// FileConfig 文件存储配置
type FileConfig struct {
	UploadPath    string `json:"upload_path"`
	ThumbnailPath string `json:"thumbnail_path"`
	TempPath      string `json:"temp_path"`
	DownloadPath  string `json:"download_path"`
	MaxFileSize   int64  `json:"max_file_size"` // MB
	AllowedTypes  string `json:"allowed_types"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

// LoadConfig 加载配置
func LoadConfig() *Config {
	// 加载.env文件
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	config := &Config{
		Server: ServerConfig{
			Port:    getEnv("SERVER_PORT", "8080"),
			Mode:    getEnv("GIN_MODE", "debug"),
			BaseURL: getEnv("BASE_URL", "http://localhost:8080"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "mcs_db"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
			TimeZone: getEnv("DB_TIMEZONE", "Asia/Shanghai"),
		},
		JWT: JWTConfig{
			SecretKey:       getEnv("JWT_SECRET", "your-secret-key-here"),
			ExpirationHours: getEnvAsInt("JWT_EXPIRE_TIME", 24),
		},
		File: FileConfig{
			UploadPath:    getEnv("UPLOAD_PATH", "./uploads"),
			ThumbnailPath: getEnv("THUMBNAIL_PATH", "./thumbnails"),
			TempPath:      getEnv("TEMP_PATH", "./temp"),
			DownloadPath:  getEnv("DOWNLOAD_PATH", "./downloads"),
			MaxFileSize:   getEnvAsInt64("MAX_FILE_SIZE", 100), // 100MB
			AllowedTypes:  getEnv("ALLOWED_TYPES", "jpg,jpeg,png,gif,mp4,mov,avi,raw,cr2,nef"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
	}

	return config
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt 获取环境变量并转换为int
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsInt64 获取环境变量并转换为int64
func getEnvAsInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// GetDSN 获取数据库连接字符串
func (c *Config) GetDSN() string {
	return "host=" + c.Database.Host +
		" user=" + c.Database.User +
		" password=" + c.Database.Password +
		" dbname=" + c.Database.DBName +
		" port=" + c.Database.Port +
		" sslmode=" + c.Database.SSLMode +
		" TimeZone=" + c.Database.TimeZone
}