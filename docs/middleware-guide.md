# 中间件使用指南

## 概述

中间件是Web应用中用于处理请求和响应的组件，它们在路由处理程序之前或之后执行。资源分享平台提供了一套完整的认证和授权中间件。

**作者**: Felix Wang  
**邮箱**: felixwang.biz@gmail.com

## 目录

- [认证中间件](#认证中间件)
- [权限中间件](#权限中间件)
- [其他中间件](#其他中间件)
- [使用示例](#使用示例)
- [最佳实践](#最佳实践)

## 认证中间件

### 1. AuthMiddleware - JWT认证中间件

**功能**: 验证JWT Token并设置用户信息到上下文

```go
r := gin.Default()
r.Use(middleware.AuthMiddleware())

// 需要认证的路由
r.GET("/profile", func(c *gin.Context) {
    user, _ := middleware.AuthUserFromContext(c)
    c.JSON(200, gin.H{
        "user": user,
    })
})
```

**特性**:
- ✅ 从Header中获取Bearer Token
- ✅ 自动解析JWT Token
- ✅ 将用户信息存储到上下文
- ✅ Token无效或过期时返回401错误

### 2. OptionalAuthMiddleware - 可选认证中间件

**功能**: 尝试验证JWT Token，但不强制要求登录

```go
r := gin.Default()
r.Use(middleware.OptionalAuthMiddleware())

// 公开路由，但已登录用户可获得更多信息
r.GET("/dashboard", func(c *gin.Context) {
    user, exists := c.Get(string(middleware.AuthUserKey))
    if exists {
        // 已登录用户
        c.JSON(200, gin.H{"user": user})
    } else {
        // 未登录用户
        c.JSON(200, gin.H{"message": "请登录"})
    }
})
```

**特性**:
- ✅ 尝试解析Token，但不会阻断请求
- ✅ 无Token时正常执行
- ✅ Token无效时正常执行

## 权限中间件

### 1. AdminMiddleware - 管理员权限中间件

**功能**: 检查用户是否为管理员

```go
r := gin.Default()
r.Use(middleware.AuthMiddleware())

// 管理员路由
adminGroup := r.Group("/admin")
adminGroup.Use(middleware.AdminMiddleware())
{
    adminGroup.GET("/users", func(c *gin.Context) {
        // 只有管理员可以访问
        c.JSON(200, gin.H{"message": "用户列表"})
    })
}
```

**特性**:
- ✅ 先执行JWT认证
- ✅ 检查用户角色是否为admin
- ✅ 不是管理员时返回403错误

### 2. RequireUploadMiddleware - 上传权限中间件

**功能**: 检查用户是否有上传权限

```go
r := gin.Default()
r.Use(middleware.AuthMiddleware())

// 上传路由
r.POST("/resources", middleware.RequireUploadMiddleware(), func(c *gin.Context) {
    // 只有有上传权限的用户可以访问
    c.JSON(200, gin.H{"message": "创建资源"})
})
```

**特性**:
- ✅ 先执行JWT认证
- ✅ 检查用户CanUpload字段
- ✅ 无上传权限时返回403错误

### 3. ActiveUserMiddleware - 活跃用户验证中间件

**功能**: 检查用户账户是否处于活跃状态

```go
userStatusService := user.NewUserStatusService(db)

r := gin.Default()
r.Use(middleware.ActiveUserMiddleware(userStatusService))

// 只有活跃用户可以访问
r.GET("/profile", func(c *gin.Context) {
    c.JSON(200, gin.H{"message": "用户资料"})
})
```

**特性**:
- ✅ 先执行JWT认证
- ✅ 检查用户状态是否为active
- ✅ 账户被禁用时返回403错误

### 4. OwnerOrAdminMiddleware - 资源所有者或管理员权限中间件

**功能**: 检查是否为资源所有者或管理员

```go
// 获取资源所有者ID的函数
func getResourceOwnerID(c *gin.Context) (uint, error) {
    resourceID := c.Param("id")
    // 从数据库查询资源所有者
    // return ownerID, nil
    return 1, nil // 模拟
}

r := gin.Default()
r.Use(middleware.AuthMiddleware())

// 资源路由
r.PUT("/resources/:id", 
    middleware.OwnerOrAdminMiddleware(getResourceOwnerID),
    func(c *gin.Context) {
        // 只有资源所有者或管理员可以修改
        c.JSON(200, gin.H{"message": "更新成功"})
    })
```

**特性**:
- ✅ 检查用户是否为资源所有者
- ✅ 检查用户是否为管理员
- ✅ 其他用户无法访问

## 其他中间件

### 1. RateLimitMiddleware - 速率限制中间件

**功能**: 限制客户端的请求频率

```go
r := gin.Default()

// 每分钟最多5次请求
r.Use(middleware.RateLimitMiddleware(5, time.Minute))

r.GET("/api", func(c *gin.Context) {
    c.JSON(200, gin.H{"message": "API响应"})
})
```

**特性**:
- ✅ 简单内存存储（生产环境应使用Redis）
- ✅ 可配置限制次数和时间窗口
- ✅ 超出限制时返回429错误

### 2. CORSMiddleware - CORS中间件

**功能**: 处理跨域资源共享

```go
r := gin.Default()
r.Use(middleware.CORSMiddleware())

r.GET("/", func(c *gin.Context) {
    c.JSON(200, gin.H{"message": "CORS已启用"})
})
```

**特性**:
- ✅ 支持所有常用CORS头
- ✅ 支持预检请求(OPTIONS)
- ✅ 可配置允许的源

### 3. RecoverMiddleware - 恢复中间件

**功能**: 处理Panic异常

```go
r := gin.Default()
r.Use(middleware.RecoverMiddleware())

// 如果此路由发生Panic，中间件会捕获并返回错误
r.GET("/panic", func(c *gin.Context) {
    panic("测试Panic")
})
```

**特性**:
- ✅ 自动捕获Panic
- ✅ 返回500错误
- ✅ 记录错误信息

### 4. RequestIDMiddleware - 请求ID中间件

**功能**: 为每个请求生成唯一ID

```go
r := gin.Default()
r.Use(middleware.RequestIDMiddleware())

r.GET("/", func(c *gin.Context) {
    requestID, _ := c.Get("request_id")
    c.JSON(200, gin.H{
        "message": "请求ID: " + requestID.(string),
    })
})
```

**特性**:
- ✅ 生成或传递请求ID
- ✅ 设置响应头X-Request-ID
- ✅ 便于请求追踪

### 5. RequireUserIDMiddleware - 用户ID参数中间件

**功能**: 强制要求用户ID参数

```go
r := gin.Default()
r.Use(middleware.AuthMiddleware())

r.DELETE("/users/:user_id", 
    middleware.RequireUserIDMiddleware(),
    func(c *gin.Context) {
        userID := c.Param("user_id")
        // 删除用户逻辑
        c.JSON(200, gin.H{"message": "删除成功"})
    })
```

**特性**:
- ✅ 检查user_id参数是否存在
- ✅ 不存在时返回400错误

## 使用示例

### 基础使用

```go
package main

import (
    "github.com/gin-gonic/gin"
    "resource-share-site/internal/middleware"
)

func main() {
    r := gin.Default()

    // 添加全局中间件
    r.Use(middleware.LoggingMiddleware())
    r.Use(middleware.RecoverMiddleware())
    r.Use(middleware.CORSMiddleware())
    r.Use(middleware.RequestIDMiddleware())

    // 公开路由
    r.GET("/public", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "公开数据"})
    })

    // 需要认证的路由
    r.Use(middleware.AuthMiddleware())
    r.GET("/profile", func(c *gin.Context) {
        user, _ := middleware.AuthUserFromContext(c)
        c.JSON(200, gin.H{"user": user})
    })

    r.Run(":8080")
}
```

### 分层使用

```go
// 1. 公开路由
public := r.Group("/api/public")
{
    public.GET("/info", func(c *gin.Context) {
        c.JSON(200, gin.H{"info": "公开信息"})
    })
}

// 2. 用户路由（需要登录）
user := r.Group("/api/user")
user.Use(middleware.AuthMiddleware())
{
    user.GET("/profile", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "用户资料"})
    })
    
    user.POST("/resources", 
        middleware.RequireUploadMiddleware(),
        func(c *gin.Context) {
            c.JSON(200, gin.H{"message": "创建资源"})
        })
}

// 3. 管理员路由（需要管理员权限）
admin := r.Group("/api/admin")
admin.Use(middleware.AdminMiddleware())
{
    admin.GET("/users", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "用户列表"})
    })
    
    admin.POST("/ban/:user_id", 
        middleware.RequireUserIDMiddleware(),
        func(c *gin.Context) {
            c.JSON(200, gin.H{"message": "用户已封禁"})
        })
}
```

### 自定义认证逻辑

```go
// 自定义获取资源所有者ID的函数
func getResourceOwnerID(c *gin.Context) (uint, error) {
    resourceID := c.Param("id")
    
    // 实际项目中从数据库查询
    // resource, err := db.GetResource(resourceID)
    // if err != nil {
    //     return 0, err
    // }
    // return resource.OwnerID, nil
    
    // 示例：返回固定值
    return 1, nil
}

// 使用自定义权限检查
r.PUT("/resources/:id", 
    middleware.AuthMiddleware(),
    middleware.OwnerOrAdminMiddleware(getResourceOwnerID),
    func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "资源已更新"})
    })
```

## 最佳实践

### 1. 中间件顺序

正确的中间件顺序很重要：

```go
// ✅ 正确的顺序
r.Use(middleware.LoggingMiddleware())        // 日志（最先）
r.Use(middleware.RecoverMiddleware())        // 恢复
r.Use(middleware.CORSMiddleware())           // CORS
r.Use(middleware.RequestIDMiddleware())      // 请求ID
r.Use(middleware.AuthMiddleware())           // 认证
r.Use(middleware.AdminMiddleware())          // 管理员权限
r.Use(middleware.RequireUploadMiddleware())  // 上传权限
```

### 2. 错误处理

始终处理中间件返回的错误：

```go
r.Use(middleware.AuthMiddleware())

r.GET("/secure", func(c *gin.Context) {
    user, err := middleware.AuthUserFromContext(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{
            "error": err.Error(),
        })
        return
    }
    // 继续处理
})
```

### 3. 上下文安全

从中间件获取用户信息时检查类型：

```go
func GetUser(c *gin.Context) (*middleware.AuthUser, error) {
    user, exists := c.Get(string(middleware.AuthUserKey))
    if !exists {
        return nil, errors.New("用户未认证")
    }
    
    authUser, ok := user.(*middleware.AuthUser)
    if !ok {
        return nil, errors.New("用户信息类型错误")
    }
    
    return authUser, nil
}
```

### 4. 性能优化

生产环境建议：

1. **速率限制**: 使用Redis而非内存存储
2. **日志**: 配置日志轮转
3. **缓存**: 缓存认证结果
4. **监控**: 监控中间件性能

### 5. 安全考虑

1. **Token验证**: 验证Token的签名和过期时间
2. **权限检查**: 最小权限原则
3. **参数验证**: 验证所有参数
4. **错误信息**: 不暴露敏感信息

## 常见问题

### Q1: 中间件执行顺序混乱

**A**: 确保按照逻辑顺序添加中间件

```go
r.Use(middleware.AuthMiddleware())    // 先认证
r.Use(middleware.AdminMiddleware())   // 后权限检查
```

### Q2: 无法获取用户信息

**A**: 检查是否正确使用了AuthMiddleware

```go
r := gin.Default()
r.Use(middleware.AuthMiddleware()) // 必须添加认证中间件

r.GET("/profile", func(c *gin.Context) {
    user, err := middleware.AuthUserFromContext(c) // 然后获取用户
})
```

### Q3: 速率限制不生效

**A**: 生产环境应使用Redis

```go
// 开发环境（内存存储）
r.Use(middleware.RateLimitMiddleware(10, time.Minute))

// 生产环境（Redis存储）
// TODO: 实现Redis版本的速率限制中间件
```

### Q4: CORS不生效

**A**: 确保CORS中间件在所有路由之前

```go
r := gin.Default()
r.Use(middleware.CORSMiddleware()) // 最先添加

// 其他路由...
```

## 测试

### 单元测试

```go
func TestAuthMiddleware(t *testing.T) {
    r := gin.Default()
    r.Use(middleware.AuthMiddleware())
    
    r.GET("/test", func(c *gin.Context) {
        user, err := middleware.AuthUserFromContext(c)
        assert.NotNil(t, err)
    })
    
    req, _ := http.NewRequest("GET", "/test", nil)
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)
    
    assert.Equal(t, 401, w.Code)
}
```

### 集成测试

```bash
# 启动测试服务器
go run cmd/testmiddleware/main.go

# 测试公开端点
curl http://localhost:8080/public

# 测试需要认证的端点
curl -H "Authorization: Bearer <token>" http://localhost:8080/profile
```

## 相关文档

- [认证服务指南](auth-service-guide.md)
- [用户状态管理](user-status-guide.md)
- [Session管理](session-guide.md)

---

**维护者**: Felix Wang  
**邮箱**: felixwang.biz@gmail.com  
**最后更新**: 2025-10-31
