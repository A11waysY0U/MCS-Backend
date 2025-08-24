# MCS Backend - 媒体协作系统后端

## 项目简介

MCS (Media Cooperation System) 是一个专为媒体团队设计的协作系统后端服务，提供完整的文件管理、工作流管理、用户管理和协作功能。

## 主要功能

### 🔐 用户认证与权限管理
- JWT认证机制
- 邀请码注册系统
- 基于角色的权限控制（RBAC）
- 用户组管理

### 📁 文件管理系统
- 分片上传与断点续传
- MD5校验与秒传功能
- 文件版本控制
- 文件夹层级管理
- 文件标签与搜索
- 批量下载与ZIP打包

### 🔄 工作流管理
- 项目工作流创建与管理
- 任务分配与状态跟踪
- 暂存区文件管理
- 截止日期与优先级管理

### 🔔 通知系统
- 实时通知推送
- 系统公告发布
- 消息已读状态管理
- 批量操作支持

### 📊 统计报表
- 操作日志记录与分析
- 存储空间统计
- 用户活跃度分析
- 系统概览仪表板

### 📥 下载服务
- 单文件下载
- 批量文件打包下载
- 下载任务管理
- 下载统计分析

## 技术栈

- **语言**: Go 1.21+
- **框架**: Gin Web Framework
- **数据库**: PostgreSQL
- **缓存**: Redis
- **认证**: JWT
- **文件存储**: 本地文件系统
- **配置管理**: 环境变量 + .env文件

## 项目结构

```
MCS-Backend/
├── cmd/
│   └── main.go                 # 应用入口
├── internal/
│   ├── api/
│   │   └── routes.go           # 路由配置
│   ├── config/
│   │   └── config.go           # 配置管理
│   ├── database/
│   │   └── database.go         # 数据库连接
│   ├── handlers/
│   │   ├── auth_handler.go     # 认证处理器
│   │   ├── user_handler.go     # 用户管理处理器
│   │   ├── file_handler.go     # 文件管理处理器
│   │   ├── workflow_handler.go # 工作流处理器
│   │   ├── notification_handler.go # 通知处理器
│   │   ├── statistics_handler.go   # 统计处理器
│   │   └── download_handler.go     # 下载处理器
│   ├── middleware/
│   │   ├── auth.go             # 认证中间件
│   │   ├── cors.go             # CORS中间件
│   │   └── logger.go           # 日志中间件
│   ├── models/
│   │   ├── user.go             # 用户模型
│   │   ├── file.go             # 文件模型
│   │   ├── workflow.go         # 工作流模型
│   │   ├── notification.go     # 通知模型
│   │   ├── statistics.go       # 统计模型
│   │   └── download.go         # 下载模型
│   └── services/
│       ├── auth_service.go     # 认证服务
│       ├── user_service.go     # 用户服务
│       ├── file_service.go     # 文件服务
│       ├── workflow_service.go # 工作流服务
│       ├── notification_service.go # 通知服务
│       ├── statistics_service.go   # 统计服务
│       └── download_service.go     # 下载服务
├── migrations/
│   ├── 001_create_users_table.sql
│   ├── 002_create_files_table.sql
│   ├── 003_create_workflows_table.sql
│   ├── 004_create_tasks_table.sql
│   ├── 005_create_notifications_table.sql
│   ├── 006_create_user_groups_table.sql
│   ├── 007_create_file_tags_table.sql
│   ├── 008_create_statistics_tables.sql
│   └── 009_create_download_tables.sql
├── docs/
│   └── api.md                  # API文档
├── .env.example                # 环境变量示例
├── go.mod                      # Go模块文件
├── go.sum                      # Go依赖校验
└── README.md                   # 项目说明
```

## 快速开始

### 环境要求

- Go 1.21 或更高版本
- PostgreSQL 12 或更高版本
- Redis 6.0 或更高版本

### 安装步骤

1. **克隆项目**
```bash
git clone <repository-url>
cd MCS-Backend
```

2. **安装依赖**
```bash
go mod download
```

3. **配置环境变量**
```bash
cp .env.example .env
# 编辑 .env 文件，配置数据库连接等信息
```

4. **创建数据库**
```sql
CREATE DATABASE mcs_db;
```

5. **运行数据库迁移**
```bash
# 按顺序执行 migrations/ 目录下的SQL文件
psql -U postgres -d mcs_db -f migrations/001_create_users_table.sql
psql -U postgres -d mcs_db -f migrations/002_create_files_table.sql
# ... 执行所有迁移文件
```

6. **创建必要目录**
```bash
mkdir -p uploads thumbnails temp downloads
```

7. **编译并运行**
```bash
go build -o mcs-backend ./cmd/main.go
./mcs-backend
```

### 环境变量配置

创建 `.env` 文件并配置以下变量：

```env
# 服务器配置
SERVER_PORT=8080
GIN_MODE=debug
BASE_URL=http://localhost:8080

# 数据库配置
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=mcs_db
DB_SSLMODE=disable
DB_TIMEZONE=Asia/Shanghai

# JWT配置
JWT_SECRET=your-secret-key-here
JWT_EXPIRE_TIME=24

# 文件存储配置
UPLOAD_PATH=./uploads
THUMBNAIL_PATH=./thumbnails
TEMP_PATH=./temp
DOWNLOAD_PATH=./downloads
MAX_FILE_SIZE=100
ALLOWED_TYPES=jpg,jpeg,png,gif,mp4,mov,avi,raw,cr2,nef

# Redis配置
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
```

## API文档

详细的API文档请参考 [docs/api.md](docs/api.md)

## 开发指南

### 代码结构说明

- `cmd/`: 应用程序入口点
- `internal/api/`: API路由配置
- `internal/config/`: 配置管理
- `internal/database/`: 数据库连接和初始化
- `internal/handlers/`: HTTP请求处理器
- `internal/middleware/`: 中间件（认证、CORS、日志等）
- `internal/models/`: 数据模型定义
- `internal/services/`: 业务逻辑服务层
- `migrations/`: 数据库迁移文件

### 添加新功能

1. 在 `models/` 中定义数据模型
2. 在 `services/` 中实现业务逻辑
3. 在 `handlers/` 中实现HTTP处理器
4. 在 `routes.go` 中注册路由
5. 创建相应的数据库迁移文件

### 代码规范

- 使用 `gofmt` 格式化代码
- 遵循 Go 官方代码规范
- 为公共函数和结构体添加注释
- 使用有意义的变量和函数名

## 部署

### Docker部署（推荐）

```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o mcs-backend ./cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/mcs-backend .
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080
CMD ["./mcs-backend"]
```

### 生产环境配置

1. 设置 `GIN_MODE=release`
2. 使用强密码和安全的JWT密钥
3. 配置HTTPS
4. 设置适当的文件权限
5. 配置日志轮转
6. 设置监控和告警

## 性能优化

### 数据库优化
- 为常用查询字段添加索引
- 使用连接池管理数据库连接
- 定期清理过期数据

### 文件存储优化
- 使用CDN加速文件访问
- 实现文件压缩和缩略图生成
- 定期清理临时文件

### 缓存策略
- 使用Redis缓存热点数据
- 实现查询结果缓存
- 缓存用户会话信息

## 安全考虑

- JWT Token定期刷新
- 文件上传类型和大小限制
- SQL注入防护
- XSS攻击防护
- CSRF攻击防护
- 敏感信息加密存储

## 监控和日志

- 使用结构化日志记录
- 监控API响应时间
- 监控数据库连接状态
- 监控文件存储使用情况
- 设置错误告警

## 贡献指南

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

## 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情

## 联系方式

如有问题或建议，请通过以下方式联系：

- 项目Issues: [GitHub Issues](https://github.com/your-repo/issues)
- 邮箱: your-email@example.com

## 更新日志

### v1.0.0 (2024-01-15)
- 初始版本发布
- 实现基础用户认证功能
- 实现文件上传下载功能
- 实现工作流管理功能
- 实现通知系统
- 实现统计报表功能
- 实现批量下载功能