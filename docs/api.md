# MCS Backend API 文档

## 概述

MCS (Media Cooperation System) 后端API提供了完整的媒体协作系统功能，包括用户管理、文件管理、工作流管理、通知系统、统计报表和下载服务等。

## 基础信息

- **Base URL**: `http://localhost:8080/api/v1`
- **认证方式**: JWT Bearer Token
- **数据格式**: JSON

## 认证

### 用户注册
```http
POST /auth/register
Content-Type: application/json

{
  "username": "string",
  "email": "string",
  "password": "string",
  "invitation_code": "string"
}
```

### 用户登录
```http
POST /auth/login
Content-Type: application/json

{
  "username": "string",
  "password": "string"
}
```

### 刷新Token
```http
POST /auth/refresh
Authorization: Bearer <token>
```

## 用户管理

### 获取用户列表
```http
GET /users?page=1&limit=10&search=keyword
Authorization: Bearer <token>
```

### 获取用户详情
```http
GET /users/{id}
Authorization: Bearer <token>
```

### 更新用户信息
```http
PUT /users/{id}
Authorization: Bearer <token>
Content-Type: application/json

{
  "username": "string",
  "email": "string",
  "role": "string"
}
```

### 删除用户
```http
DELETE /users/{id}
Authorization: Bearer <token>
```

## 用户组管理

### 获取用户组列表
```http
GET /groups
Authorization: Bearer <token>
```

### 创建用户组
```http
POST /groups
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "string",
  "description": "string",
  "permissions": ["string"]
}
```

### 更新用户组
```http
PUT /groups/{id}
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "string",
  "description": "string",
  "permissions": ["string"]
}
```

### 删除用户组
```http
DELETE /groups/{id}
Authorization: Bearer <token>
```

## 文件管理

### 文件上传
```http
POST /upload/chunk
Authorization: Bearer <token>
Content-Type: multipart/form-data

chunk: file
chunk_number: integer
total_chunks: integer
file_name: string
file_size: integer
md5_hash: string
```

### 完成上传
```http
POST /upload/complete
Authorization: Bearer <token>
Content-Type: application/json

{
  "file_name": "string",
  "total_chunks": 10,
  "md5_hash": "string",
  "folder_id": 1,
  "workflow_id": 1,
  "task_id": 1
}
```

### 获取文件列表
```http
GET /files?page=1&limit=10&folder_id=1&workflow_id=1
Authorization: Bearer <token>
```

### 获取文件详情
```http
GET /files/{id}
Authorization: Bearer <token>
```

### 下载文件
```http
GET /files/{id}/download
Authorization: Bearer <token>
```

### 更新文件信息
```http
PUT /files/{id}
Authorization: Bearer <token>
Content-Type: application/json

{
  "file_name": "string",
  "description": "string",
  "tags": ["string"]
}
```

### 删除文件
```http
DELETE /files/{id}
Authorization: Bearer <token>
```

## 文件夹管理

### 获取文件夹列表
```http
GET /folders?workflow_id=1&parent_id=0
Authorization: Bearer <token>
```

### 创建文件夹
```http
POST /folders
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "string",
  "parent_id": 0,
  "workflow_id": 1,
  "description": "string"
}
```

### 更新文件夹
```http
PUT /folders/{id}
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "string",
  "description": "string"
}
```

### 删除文件夹
```http
DELETE /folders/{id}
Authorization: Bearer <token>
```

## 工作流管理

### 获取工作流列表
```http
GET /workflows?page=1&limit=10&status=active
Authorization: Bearer <token>
```

### 创建工作流
```http
POST /workflows
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "string",
  "description": "string",
  "deadline": "2024-12-31T23:59:59Z",
  "priority": "high"
}
```

### 更新工作流
```http
PUT /workflows/{id}
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "string",
  "description": "string",
  "status": "active",
  "deadline": "2024-12-31T23:59:59Z"
}
```

### 删除工作流
```http
DELETE /workflows/{id}
Authorization: Bearer <token>
```

## 任务管理

### 获取任务列表
```http
GET /tasks?workflow_id=1&status=pending&assignee_id=1
Authorization: Bearer <token>
```

### 创建任务
```http
POST /tasks
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "string",
  "description": "string",
  "workflow_id": 1,
  "assignee_id": 1,
  "deadline": "2024-12-31T23:59:59Z",
  "priority": "high"
}
```

### 更新任务
```http
PUT /tasks/{id}
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "string",
  "description": "string",
  "status": "in_progress",
  "assignee_id": 1
}
```

### 删除任务
```http
DELETE /tasks/{id}
Authorization: Bearer <token>
```

## 通知管理

### 获取通知列表
```http
GET /notifications?page=1&limit=10&is_read=false
Authorization: Bearer <token>
```

### 标记通知为已读
```http
PUT /notifications/{id}/read
Authorization: Bearer <token>
```

### 批量标记通知为已读
```http
PUT /notifications/batch-read
Authorization: Bearer <token>
Content-Type: application/json

{
  "notification_ids": [1, 2, 3]
}
```

### 删除通知
```http
DELETE /notifications/{id}
Authorization: Bearer <token>
```

### 获取系统公告
```http
GET /notifications/announcements
Authorization: Bearer <token>
```

### 创建系统公告（管理员）
```http
POST /notifications/announcements
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "string",
  "content": "string",
  "priority": "high",
  "target_users": [1, 2, 3]
}
```

## 统计报表

### 获取操作日志
```http
GET /statistics/operation-logs?page=1&limit=10&start_date=2024-01-01&end_date=2024-12-31
Authorization: Bearer <token>
```

### 获取存储统计
```http
GET /statistics/storage
Authorization: Bearer <token>
```

### 获取用户活跃度统计
```http
GET /statistics/user-activity?user_id=1&start_date=2024-01-01&end_date=2024-12-31
Authorization: Bearer <token>
```

### 获取系统概览
```http
GET /statistics/system-overview
Authorization: Bearer <token>
```

### 获取操作统计
```http
GET /statistics/operations?start_date=2024-01-01&end_date=2024-12-31
Authorization: Bearer <token>
```

### 获取最活跃用户
```http
GET /statistics/top-users?limit=10&start_date=2024-01-01&end_date=2024-12-31
Authorization: Bearer <token>
```

## 下载服务

### 创建批量下载任务
```http
POST /download/batch
Authorization: Bearer <token>
Content-Type: application/json

{
  "file_ids": [1, 2, 3],
  "archive_name": "my_files.zip"
}
```

### 获取下载任务信息
```http
GET /download/tasks/{task_id}
Authorization: Bearer <token>
```

### 获取用户下载任务列表
```http
GET /download/tasks?page=1&limit=10&status=completed
Authorization: Bearer <token>
```

### 下载ZIP文件
```http
GET /download/zip/{task_id}
Authorization: Bearer <token>
```

### 获取下载统计
```http
GET /download/stats?start_date=2024-01-01&end_date=2024-12-31
Authorization: Bearer <token>
```

### 获取全局下载统计（管理员）
```http
GET /download/stats/global?start_date=2024-01-01&end_date=2024-12-31
Authorization: Bearer <token>
```

## 错误响应格式

所有API错误响应都遵循以下格式：

```json
{
  "error": {
    "code": 400,
    "message": "错误描述",
    "details": "详细错误信息（可选）"
  }
}
```

## 状态码说明

- `200 OK`: 请求成功
- `201 Created`: 资源创建成功
- `400 Bad Request`: 请求参数错误
- `401 Unauthorized`: 未授权或Token无效
- `403 Forbidden`: 权限不足
- `404 Not Found`: 资源不存在
- `409 Conflict`: 资源冲突
- `422 Unprocessable Entity`: 请求格式正确但语义错误
- `500 Internal Server Error`: 服务器内部错误

## 分页响应格式

所有分页接口都遵循以下响应格式：

```json
{
  "data": [],
  "pagination": {
    "page": 1,
    "limit": 10,
    "total": 100,
    "total_pages": 10
  }
}
```

## 权限说明

- **admin**: 系统管理员，拥有所有权限
- **user**: 普通用户，只能访问自己的资源
- **guest**: 访客用户，只读权限

## 注意事项

1. 所有需要认证的接口都必须在请求头中包含有效的JWT Token
2. 文件上传支持分片上传和断点续传
3. 批量下载会异步处理，需要轮询任务状态
4. 系统会自动清理过期的下载任务和临时文件
5. 所有时间格式都使用ISO 8601标准（RFC3339）