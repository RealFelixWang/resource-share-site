# Resource Share Site API 文档

## 概述
本文档描述了资源分享网站的 RESTful API 接口。

**基础 URL**: `http://localhost:8080/api/v1`

## 认证方式
API 使用 JWT Token 认证，在请求头中携带：
```
Authorization: Bearer <token>
```

## 通用响应格式
```json
{
  "code": 200,
  "message": "success",
  "data": {}
}
```

## 错误码
- 200: 成功
- 400: 请求参数错误
- 401: 未授权
- 403: 禁止访问
- 404: 资源不存在
- 500: 服务器内部错误

## API 端点

### 1. 用户认证

#### 1.1 用户注册
```
POST /auth/register
```

**请求体**:
```json
{
  "username": "user123",
  "email": "user@example.com",
  "password": "password123",
  "confirm_password": "password123",
  "invite_code": "optional_invite_code"
}
```

**响应**:
```json
{
  "code": 200,
  "message": "注册成功",
  "data": {
    "id": 1,
    "username": "user123",
    "email": "user@example.com",
    "invite_code": "abc123",
    "points_balance": 0
  }
}
```

#### 1.2 用户登录
```
POST /auth/login
```

**请求体**:
```json
{
  "identifier": "user123",
  "password": "password123",
  "remember": false
}
```

**响应**:
```json
{
  "code": 200,
  "message": "登录成功",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": "2025-11-04T12:00:00Z",
    "user": {
      "id": 1,
      "username": "user123",
      "email": "user@example.com",
      "role": "user",
      "status": "active",
      "can_upload": false,
      "points_balance": 0
    }
  }
}
```

#### 1.3 用户登出
```
POST /auth/logout
```
需要认证

**响应**:
```json
{
  "code": 200,
  "message": "登出成功"
}
```

### 2. 资源管理

#### 2.1 获取资源列表
```
GET /resources?page=1&limit=10&category_id=1
```

**查询参数**:
- `page`: 页码 (默认1)
- `limit`: 每页数量 (默认10)
- `category_id`: 分类ID
- `keyword`: 搜索关键词

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "total": 100,
    "resources": [
      {
        "id": 1,
        "title": "示例资源",
        "description": "资源描述",
        "category": {
          "id": 1,
          "name": "软件"
        },
        "uploaded_by": {
          "id": 1,
          "username": "admin"
        },
        "points_price": 0,
        "download_count": 10,
        "created_at": "2025-11-04T00:00:00Z"
      }
    ]
  }
}
```

#### 2.2 获取资源详情
```
GET /resources/{id}
```

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "title": "示例资源",
    "description": "资源描述",
    "content": "资源内容",
    "category": {
      "id": 1,
      "name": "软件"
    },
    "uploaded_by": {
      "id": 1,
      "username": "admin"
    },
    "points_price": 10,
    "download_count": 10,
    "created_at": "2025-11-04T00:00:00Z",
    "comments": []
  }
}
```

### 3. 分类管理

#### 3.1 获取分类列表
```
GET /categories
```

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "categories": [
      {
        "id": 1,
        "name": "软件",
        "parent_id": null,
        "children": [
          {
            "id": 2,
            "name": "系统工具",
            "parent_id": 1
          }
        ]
      }
    ]
  }
}
```

### 4. 积分系统

#### 4.1 获取积分余额
```
GET /points/balance
```
需要认证

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "balance": 100
  }
}
```

#### 4.2 获取积分记录
```
GET /points/records?page=1&limit=10
```
需要认证

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "total": 50,
    "records": [
      {
        "id": 1,
        "type": "earn",
        "amount": 10,
        "reason": "邀请奖励",
        "created_at": "2025-11-04T00:00:00Z"
      }
    ]
  }
}
```

## 管理员 API

### 5. 用户管理

#### 5.1 获取用户列表
```
GET /admin/users?page=1&limit=10
```
需要管理员权限

#### 5.2 禁用用户
```
POST /admin/users/{id}/ban
```
需要管理员权限

**请求体**:
```json
{
  "reason": "违规行为"
}
```

#### 5.3 启用用户
```
POST /admin/users/{id}/unban
```
需要管理员权限

### 6. 资源审核

#### 6.1 审核资源
```
POST /admin/resources/{id}/review
```
需要管理员权限

**请求体**:
```json
{
  "status": "approved",
  "reason": "审核通过"
}
```

### 7. 统计分析

#### 7.1 获取统计概览
```
GET /admin/statistics/overview
```
需要管理员权限

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "total_users": 100,
    "total_resources": 500,
    "total_downloads": 1000,
    "today_visits": 50
  }
}
```

## 状态码说明

### HTTP 状态码
- 200 OK: 请求成功
- 201 Created: 资源创建成功
- 400 Bad Request: 请求参数错误
- 401 Unauthorized: 未授权访问
- 403 Forbidden: 禁止访问
- 404 Not Found: 资源不存在
- 422 Unprocessable Entity: 请求格式正确但语义错误
- 500 Internal Server Error: 服务器内部错误

### 业务错误码
- 1001: 用户名已存在
- 1002: 邮箱已存在
- 1003: 邀请码无效
- 1004: 密码错误
- 2001: 资源不存在
- 2002: 积分不足
- 3001: 权限不足

## 注意事项

1. 所有需要认证的接口必须在请求头中携带有效的 JWT Token
2. 管理员接口需要用户角色为 "admin"
3. 积分价格为 0 表示免费资源
4. 所有时间格式使用 ISO 8601 标准 (RFC3339)
5. API 使用 GIN 框架，启用 RESTful 风格路由

## 错误处理

示例错误响应:
```json
{
  "code": 400,
  "message": "请求参数错误",
  "error": {
    "field": "email",
    "message": "邮箱格式不正确"
  }
}
```

---

**作者**: Felix Wang  
**邮箱**: felixwang.biz@gmail.com  
**更新时间**: 2025-11-04
