# 认证服务快速参考

## 核心功能

✅ **支持用户名或邮箱登录** - 用户可以使用用户名或邮箱登录  
✅ **完整注册流程** - 包含邀请码和积分奖励  
✅ **安全密码管理** - bcrypt加密和验证  
✅ **用户资料管理** - 支持更新用户名和邮箱  
✅ **邀请奖励系统** - 自动奖励积分  

## 快速使用

### 1. 创建服务

```go
authService := auth.NewAuthService(db)
```

### 2. 注册用户

```go
req := &auth.RegisterRequest{
    Username:        "username",
    Email:           "email@example.com",
    Password:        "password123",
    ConfirmPassword: "password123",
    InviteCode:      "", // 可选
}

response, err := authService.Register(&auth.GORMContext{DB: db}, req)
```

### 3. 登录 (用户名或邮箱)

```go
req := &auth.LoginRequest{
    Identifier: "username",       // 或 "email@example.com"
    Password:   "password123",
    Remember:   true,              // 记住登录
}

response, err := authService.Login(&auth.GORMContext{DB: db}, req)
```

### 4. 修改密码

```go
req := &auth.ChangePasswordRequest{
    OldPassword:     "oldpassword",
    NewPassword:     "newpassword",
    ConfirmPassword: "newpassword",
}

err := authService.ChangePassword(&auth.GORMContext{DB: db}, userID, req)
```

### 5. 更新资料

```go
req := &auth.UpdateProfileRequest{
    Username: "newusername",      // 可选
    Email:    "new@email.com",    // 可选
}

err := authService.UpdateProfile(&auth.GORMContext{DB: db}, userID, req)
```

## 主要结构

### 请求结构
- `LoginRequest` - 登录请求
- `RegisterRequest` - 注册请求
- `ChangePasswordRequest` - 修改密码请求
- `UpdateProfileRequest` - 更新资料请求

### 响应结构
- `LoginResponse` - 登录响应
- `RegisterResponse` - 注册响应
- `UserInfo` - 用户信息

## 关键特性

| 特性 | 说明 |
|------|------|
| **灵活登录** | 支持用户名或邮箱自动识别 |
| **密码安全** | bcrypt加密，强度验证 |
| **邀请奖励** | 自动奖励积分和记录 |
| **数据验证** | 完整输入验证和唯一性检查 |
| **错误处理** | 详细错误信息和状态码 |

## 错误处理

```go
response, err := authService.Login(&auth.GORMContext{DB: db}, req)
if err != nil {
    switch {
    case strings.Contains(err.Error(), "用户不存在"):
        // 用户不存在或密码错误
    case strings.Contains(err.Error(), "已被使用"):
        // 用户名或邮箱已被使用
    case strings.Contains(err.Error(), "禁用"):
        // 账户被禁用
    }
}
```

## 测试

```bash
go run cmd/testauth/main.go
```

## 依赖

- `internal/model/user.go` - 用户模型
- `internal/model/invitation.go` - 邀请模型
- `internal/model/point_record.go` - 积分记录模型
- `pkg/utils/helpers.go` - JWT工具函数

---

**参考**: [完整认证服务指南](auth-service-guide.md)
