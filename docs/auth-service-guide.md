# 认证服务使用指南

## 概述

认证服务提供了完整的用户认证功能，支持用户名或邮箱登录、注册、密码管理等功能。

**作者**: Felix Wang  
**邮箱**: felixwang.biz@gmail.com

## 核心功能

### 1. 灵活登录
- ✅ **支持用户名登录**: 用户可以使用用户名登录
- ✅ **支持邮箱登录**: 用户也可以使用邮箱登录
- ✅ **自动识别**: 系统自动识别输入的是用户名还是邮箱

### 2. 用户注册
- ✅ **用户名唯一性检查**: 防止重复用户名
- ✅ **邮箱唯一性检查**: 防止重复邮箱
- ✅ **密码强度验证**: 至少6位字符
- ✅ **邀请码支持**: 可选邀请码，获得奖励

### 3. 密码管理
- ✅ **安全密码加密**: 使用bcrypt加密
- ✅ **密码验证**: 安全验证密码
- ✅ **修改密码**: 支持修改密码功能

### 4. 用户资料管理
- ✅ **更新用户名**: 支持修改用户名
- ✅ **更新邮箱**: 支持修改邮箱
- ✅ **唯一性检查**: 确保新用户名/邮箱未被使用

### 5. 邀请系统
- ✅ **自动奖励**: 邀请人可获得积分奖励
- ✅ **邀请记录**: 完整的邀请关系记录
- ✅ **积分自动入账**: 邀请成功后自动奖励积分

## 目录结构

```
internal/service/auth/
├── auth.go  # 认证服务实现

cmd/testauth/
└── main.go  # 使用示例

pkg/utils/
└── helpers.go  # 工具函数(JWT token等)
```

## 数据模型

### 登录请求 (LoginRequest)
```go
type LoginRequest struct {
    Identifier string `json:"identifier" binding:"required"`  // 用户名或邮箱
    Password   string `json:"password" binding:"required"`    // 密码
    Remember   bool   `json:"remember"`                       // 记住登录状态
}
```

### 注册请求 (RegisterRequest)
```go
type RegisterRequest struct {
    Username        string `json:"username"`                 // 用户名
    Email           string `json:"email"`                    // 邮箱
    Password        string `json:"password"`                 // 密码
    ConfirmPassword string `json:"confirm_password"`         // 确认密码
    InviteCode      string `json:"invite_code"`              // 邀请码(可选)
}
```

### 用户信息 (UserInfo)
```go
type UserInfo struct {
    ID              uint   `json:"id"`                       // 用户ID
    Username        string `json:"username"`                 // 用户名
    Email           string `json:"email"`                    // 邮箱
    Role            string `json:"role"`                     // 角色
    Status          string `json:"status"`                   // 状态
    CanUpload       bool   `json:"can_upload"`               // 是否有上传权限
    PointsBalance   int    `json:"points_balance"`           // 积分余额
    InviteCode      string `json:"invite_code"`              // 邀请码
    UploadedResourcesCount   uint `json:"uploaded_resources_count"`   // 上传资源数
    DownloadedResourcesCount uint `json:"downloaded_resources_count"` // 下载资源数
}
```

## 使用方法

### 1. 创建认证服务

```go
import (
    "resource-share-site/internal/config"
    "resource-share-site/internal/database"
    "resource-share-site/internal/service/auth"
)

func main() {
    // 初始化数据库
    dbConfig := &config.DatabaseConfig{
        Type:     "mysql",
        Host:     "localhost",
        Port:     "3306",
        Name:     "resource_share_site",
        User:     "root",
        Password: "123456",
    }

    db, err := database.InitDatabaseWithConfig(dbConfig)
    if err != nil {
        panic(err)
    }

    // 创建认证服务
    authService := auth.NewAuthService(db)
}
```

### 2. 用户注册

```go
req := &auth.RegisterRequest{
    Username:        "newuser",
    Email:           "newuser@example.com",
    Password:        "password123",
    ConfirmPassword: "password123",
    InviteCode:      "", // 可选，有邀请码可获得奖励
}

response, err := authService.Register(&auth.GORMContext{DB: db}, req)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("注册成功: %s\n", response.Username)
fmt.Printf("邀请码: %s\n", response.InviteCode)
fmt.Printf("积分余额: %d\n", response.PointsBalance)
```

### 3. 用户登录 (支持用户名或邮箱)

```go
// 使用用户名登录
req := &auth.LoginRequest{
    Identifier: "newuser",  // 或 "newuser@example.com"
    Password:   "password123",
    Remember:   true,
}

response, err := authService.Login(&auth.GORMContext{DB: db}, req)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("登录成功\n")
fmt.Printf("Token: %s\n", response.Token)
fmt.Printf("用户: %s\n", response.User.Username)
fmt.Printf("邮箱: %s\n", response.User.Email)
```

### 4. 修改密码

```go
req := &auth.ChangePasswordRequest{
    OldPassword:     "password123",
    NewPassword:     "newpassword456",
    ConfirmPassword: "newpassword456",
}

if err := authService.ChangePassword(&auth.GORMContext{DB: db}, userID, req); err != nil {
    log.Fatal(err)
}

fmt.Println("密码修改成功")
```

### 5. 更新用户资料

```go
req := &auth.UpdateProfileRequest{
    Username: "updateduser",      // 修改用户名
    Email:    "updated@email.com", // 修改邮箱
}

if err := authService.UpdateProfile(&auth.GORMContext{DB: db}, userID, req); err != nil {
    log.Fatal(err)
}

fmt.Println("资料更新成功")
```

## 完整示例

运行以下命令查看完整示例：

```bash
go run cmd/testauth/main.go
```

示例包含：
1. 用户注册
2. 使用用户名登录
3. 使用邮箱登录
4. 修改密码
5. 更新用户资料

## API集成示例

### 使用Gin框架

```go
package main

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "resource-share-site/internal/service/auth"
)

func main() {
    r := gin.Default()

    // 登录路由
    r.POST("/login", func(c *gin.Context) {
        var req auth.LoginRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        response, err := authService.Login(&auth.GORMContext{DB: db}, &req)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusOK, response)
    })

    // 注册路由
    r.POST("/register", func(c *gin.Context) {
        var req auth.RegisterRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        response, err := authService.Register(&auth.GORMContext{DB: db}, &req)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusOK, response)
    })

    r.Run(":8080")
}
```

## 错误处理

### 常见错误类型

| 错误信息 | 原因 |
|----------|------|
| "用户不存在或密码错误" | 用户名/邮箱不存在或密码错误 |
| "用户名已被使用" | 注册时用户名已存在 |
| "邮箱已被使用" | 注册时邮箱已存在 |
| "账户已被禁用或未激活" | 用户状态不是active |
| "两次输入的密码不一致" | 密码确认不匹配 |
| "旧密码错误" | 修改密码时旧密码错误 |

### 错误示例

```go
response, err := authService.Login(&auth.GORMContext{DB: db}, &req)
if err != nil {
    switch {
    case strings.Contains(err.Error(), "用户不存在"):
        // 处理用户不存在
    case strings.Contains(err.Error(), "密码错误"):
        // 处理密码错误
    case strings.Contains(err.Error(), "禁用"):
        // 处理账户禁用
    default:
        // 处理其他错误
    }
}
```

## 安全特性

### 1. 密码安全
- ✅ **bcrypt加密**: 使用bcrypt进行密码哈希
- ✅ **盐值随机**: bcrypt自动生成随机盐值
- ✅ **强度控制**: 最少6位字符

### 2. 令牌安全
- ✅ **JWT token**: 使用JWT生成安全的访问令牌
- ✅ **过期时间**: 支持24小时或30天过期
- ✅ **唯一标识**: token包含用户ID和用户名

### 3. 输入验证
- ✅ **用户名检查**: 3-50字符，仅允许字母数字下划线
- ✅ **邮箱验证**: 标准邮箱格式验证
- ✅ **SQL注入防护**: 使用GORM参数化查询

### 4. 状态管理
- ✅ **账户状态**: 支持active/banned状态
- ✅ **登录记录**: 记录最后登录时间
- ✅ **会话管理**: 支持记住登录状态

## 数据库集成

### 依赖的表
- `users` - 用户基本信息
- `point_records` - 积分记录(邀请奖励)
- `invitations` - 邀请记录

### 自动操作
- 注册时: 创建用户记录
- 邀请时: 创建积分记录和邀请记录
- 登录时: 更新最后登录时间

## 测试

### 运行测试

```bash
# 运行认证服务示例
go run cmd/testauth/main.go

# 运行所有测试
go test ./internal/service/auth/... -v
```

### 测试覆盖

- ✅ 用户注册
- ✅ 用户名登录
- ✅ 邮箱登录
- ✅ 修改密码
- ✅ 更新资料
- ✅ 错误处理

## 最佳实践

### 1. 密码策略
- 使用强密码(8位以上，包含大小写字母、数字和特殊字符)
- 定期更换密码
- 不要使用常见密码

### 2. 令牌管理
- 在客户端安全存储token
- 设置合理的过期时间
- 实现token刷新机制

### 3. 错误处理
- 不暴露敏感信息
- 统一错误格式
- 记录错误日志

### 4. 性能优化
- 数据库索引优化
- 合理使用事务
- 缓存用户信息

## 扩展功能

### 未来计划
- 🔄 **双因子认证**: SMS/邮箱验证码
- 🔄 **OAuth登录**: Google/GitHub登录
- 🔄 **密码重置**: 邮箱重置密码
- 🔄 **账户锁定**: 多次失败自动锁定
- 🔄 **设备管理**: 查看和管理登录设备

## 相关文档

- [数据库设计](database-architecture.md)
- [API设计](../api/)
- [JWT工具函数](../pkg/utils/helpers.go)

---

**维护者**: Felix Wang  
**邮箱**: felixwang.biz@gmail.com  
**最后更新**: 2025-10-31
