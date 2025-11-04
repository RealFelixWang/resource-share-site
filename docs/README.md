# 数据库使用说明

## 概述

资源分享平台使用 MySQL 9.0+ 作为主数据库，SQLite 用于开发测试。本文档说明数据库的配置、迁移和维护方法。

## 目录结构

```
docs/
├── database-architecture.md      # 完整数据库架构设计文档
├── database-quick-reference.md   # 快速参考指南
├── mysql-schema.sql             # MySQL 9 建表脚本
└── README.md                    # 本文档

internal/
├── config/
│   └── database.go              # 数据库配置管理
├── database/
│   └── migration.go             # 数据库迁移脚本
└── model/                       # 数据模型定义
    ├── user.go
    ├── category.go
    ├── resource.go
    ├── comment.go
    ├── invitation.go
    ├── point_record.go
    ├── points_rule.go
    ├── visit_log.go
    ├── ip_blacklist.go
    ├── ad.go
    └── other.go

cmd/testdb/
└── main.go                      # 数据库连接测试程序
```

## 快速开始

### 1. 使用 SQLite（开发环境）

```go
package main

import (
    "resource-share-site/internal/config"
    "resource-share-site/internal/database"
)

func main() {
    // 配置 SQLite 数据库
    dbConfig := &config.DatabaseConfig{
        Type: "sqlite",
        Name: "resource_share_site",
    }

    // 初始化数据库
    db, err := database.InitDatabaseWithConfig(dbConfig)
    if err != nil {
        panic(err)
    }

    // 创建默认数据
    err = database.CreateDefaultData(db)
    if err != nil {
        panic(err)
    }

    println("数据库初始化完成!")
}
```

### 2. 使用 MySQL（生产环境）

```go
package main

import (
    "resource-share-site/internal/config"
    "resource-share-site/internal/database"
)

func main() {
    // 配置 MySQL 数据库
    dbConfig := &config.DatabaseConfig{
        Type:     "mysql",
        Host:     "localhost",
        Port:     "3306",
        Name:     "resource_share_site",
        User:     "root",
        Password: "123456",
        Charset:  "utf8mb4",
    }

    // 初始化数据库
    db, err := database.InitDatabaseWithConfig(dbConfig)
    if err != nil {
        panic(err)
    }

    // 创建默认数据
    err = database.CreateDefaultData(db)
    if err != nil {
        panic(err)
    }

    println("数据库初始化完成!")
}
```

## 数据库配置

### 支持的数据库类型

- **MySQL 9.0+**: 生产环境推荐
- **SQLite**: 开发测试环境

### 配置参数

```go
type DatabaseConfig struct {
    Type     string `mapstructure:"type"`     // 数据库类型: mysql/sqlite
    Host     string `mapstructure:"host"`     // MySQL 主机地址
    Port     string `mapstructure:"port"`     // MySQL 端口
    Name     string `mapstructure:"name"`     // 数据库名
    User     string `mapstructure:"user"`     // 用户名
    Password string `mapstructure:"password"` // 密码
    Charset  string `mapstructure:"charset"`  // 字符集
}
```

## 数据模型

### 核心表 (14张)

1. **用户系统**
   - `users` - 用户信息
   - `sessions` - 会话管理

2. **分类系统**
   - `categories` - 资源分类

3. **资源系统**
   - `resources` - 资源信息
   - `comments` - 用户评论

4. **邀请系统**
   - `invitations` - 邀请关系

5. **积分系统**
   - `points_rules` - 积分规则
   - `point_records` - 积分流水

6. **监控审计**
   - `visit_logs` - 访问日志
   - `ip_blacklists` - IP黑名单
   - `admin_logs` - 管理员日志

7. **系统管理**
   - `ads` - 广告管理
   - `permissions` - 权限配置
   - `import_tasks` - 导入任务

### 软删除

所有业务表都实现了软删除机制，使用 `DeletedAt` 字段标记删除时间。查询时会自动过滤已删除的记录。

```go
// GORM 会自动处理软删除
var users []model.User
db.Find(&users) // 只返回未删除的用户

// 查询包含软删除的记录
var allUsers []model.User
db.Unscoped().Find(&allUsers)
```

## 数据库迁移

### 自动迁移

使用 GORM 的 `AutoMigrate` 功能自动创建表结构：

```go
db.AutoMigrate(
    &model.User{},
    &model.Category{},
    &model.Resource{},
    // ... 其他模型
)
```

### 手动执行迁移

```bash
# 运行测试程序进行迁移
go run cmd/testdb/main.go
```

### 回滚迁移（谨慎使用）

```go
database.RollbackMigrations(db) // 会删除所有数据!
```

## 初始化数据

### 默认管理员账户

- **用户名**: admin
- **邮箱**: admin@example.com
- **密码**: admin123
- **角色**: admin

**⚠️ 生产环境请立即修改默认密码！**

### 默认积分规则

| 规则键名 | 规则名称 | 积分 | 说明 |
|----------|----------|------|------|
| invite_reward | 邀请奖励 | +50 | 成功邀请一个用户注册 |
| resource_download | 资源下载 | -10 | 下载需要积分的资源 |
| daily_checkin | 每日签到 | +5 | 每日登录奖励 |
| upload_reward | 上传奖励 | +10 | 审核通过一个资源 |

### 默认权限

| 权限键名 | 权限名称 | 说明 |
|----------|----------|------|
| user.upload | 用户上传 | 允许用户上传资源 |
| user.comment | 用户评论 | 允许用户评论资源 |
| admin.review | 资源审核 | 允许审核资源 |
| admin.ban_user | 封禁用户 | 允许封禁/解封用户 |
| admin.ip_ban | IP封禁 | 允许封禁IP地址 |
| admin.manage_ads | 广告管理 | 允许管理广告 |
| admin.view_logs | 查看日志 | 允许查看系统日志 |
| admin.import | 导入数据 | 允许导入资源数据 |

### 默认分类

- 软件工具
- 电子资料
- 多媒体
- 游戏
- 其他

## 使用示例

### 1. 创建用户

```go
user := model.User{
    Username:     "newuser",
    Email:        "newuser@example.com",
    PasswordHash: utils.HashPassword("password123"),
    Role:         "user",
    Status:       "active",
    CanUpload:    false,
    InviteCode:   utils.GenerateInviteCode(),
    PointsBalance: 0,
}

if err := db.Create(&user).Error; err != nil {
    log.Fatal(err)
}
```

### 2. 创建资源

```go
resource := model.Resource{
    Title:        "测试资源",
    Description:  "这是一个测试资源",
    CategoryID:   1,
    NetdiskURL:   "https://example.com/resource",
    PointsPrice:  10,
    Source:       model.ResourceSourceUser,
    UploadedByID: user.ID,
    Status:       model.ResourceStatusPending,
}

if err := db.Create(&resource).Error; err != nil {
    log.Fatal(err)
}
```

### 3. 积分变动

```go
// 在事务中处理积分变动
err := db.Transaction(func(tx *gorm.DB) error {
    // 更新用户积分
    if err := tx.Model(&user).Update("points_balance", gorm.Expr("points_balance + ?", 50)).Error; err != nil {
        return err
    }

    // 记录积分流水
    record := model.PointRecord{
        UserID:       user.ID,
        Type:         model.PointTypeIncome,
        Points:       50,
        BalanceAfter: user.PointsBalance + 50,
        Source:       model.PointSourceInviteReward,
        Description:  "邀请奖励",
    }
    return tx.Create(&record).Error
})

if err != nil {
    log.Fatal(err)
}
```

### 4. 查询资源列表

```go
var resources []model.Resource
if err := db.Preload("Category").Preload("UploadedBy").
    Where("status = ?", model.ResourceStatusApproved).
    Order("created_at DESC").
    Find(&resources).Error; err != nil {
    log.Fatal(err)
}
```

### 5. 审核评论

```go
comment := model.Comment{}
if err := db.First(&comment, commentID).Error; err != nil {
    log.Fatal(err)
}

// 审核通过
if err := db.Model(&comment).Updates(map[string]interface{}{
    "status":        model.CommentStatusApproved,
    "reviewed_by_id": adminID,
    "reviewed_at":   time.Now(),
    "review_notes":  "审核通过",
}).Error; err != nil {
    log.Fatal(err)
}
```

## 最佳实践

### 1. 使用预加载

```go
// 预加载关联数据
db.Preload("Category").Preload("UploadedBy").Find(&resources)

// 使用 Joins 代替预加载（性能更好）
db.Joins("Category").Joins("UploadedBy").Find(&resources)
```

### 2. 索引优化

确保查询频繁的字段已建立索引：

```sql
-- 用户表
CREATE INDEX idx_users_invited_by_id ON users(invited_by_id);

-- 资源表
CREATE INDEX idx_resources_status ON resources(status);

-- 积分记录表
CREATE INDEX idx_point_records_user_id ON point_records(user_id);
```

### 3. 分页查询

```go
var resources []model.Resource
var total int64

// 查询总数
db.Model(&model.Resource{}).Where("status = ?", "approved").Count(&total)

// 分页查询
offset := (page - 1) * pageSize
err := db.Offset(offset).Limit(pageSize).
    Where("status = ?", model.ResourceStatusApproved).
    Find(&resources).Error
```

### 4. 软删除查询

```go
// 只查询未删除的记录
db.Find(&users)

// 查询包含软删除的记录
db.Unscoped().Find(&users)

// 查询已删除的记录
db.Unscoped().Where("deleted_at IS NOT NULL").Find(&users)
```

## 测试

### 运行数据库测试

```bash
# 运行数据库连接测试
go run cmd/testdb/main.go
```

测试程序会执行以下操作：
1. 连接数据库
2. 执行迁移
3. 创建默认数据
4. 测试CRUD操作
5. 验证数据完整性

### 预期输出

```
=== 数据库连接测试 ===

1. 初始化数据库配置...
数据库类型: sqlite
数据库名称: resource_share_site

2. 连接数据库...
✅ 数据库连接成功

3. 测试数据库连接...
✅ 数据库Ping成功

4. 测试数据库迁移...
开始执行数据库迁移...
数据库迁移完成!
✅ 数据库迁移成功

5. 创建默认数据...
开始创建默认数据...
默认数据创建完成!
✅ 默认数据创建成功

6. 测试CRUD操作...
  ✅ 创建用户: testuser (ID: 1)
  ✅ 读取用户: testuser
  ✅ 更新用户积分: 150
  ✅ 软删除用户: testuser
✅ CRUD测试通过

7. 查询数据验证...
  📊 用户总数: 2
  📊 分类总数: 5
  📊 积分规则数: 4
    - 邀请奖励: 50积分 (true)
    - 资源下载: -10积分 (true)
    - 每日签到: 5积分 (true)
    - 上传奖励: 10积分 (true)
  📊 权限配置数: 8
    - 用户上传: 允许用户上传资源
    - 用户评论: 允许用户评论资源
    - 资源审核: 允许审核资源
    - 封禁用户: 允许封禁/解封用户
    - IP封禁: 允许封禁IP地址
    - 广告管理: 允许管理广告
    - 查看日志: 允许查看系统日志
    - 导入数据: 允许导入资源数据
  📊 管理员数量: 1

8. 数据库状态...
=== 数据库表列表 ===
1. ads
2. admin_logs
3. categories
4. comments
5. import_tasks
6. invitations
7. ip_blacklists
8. permissions
9. point_records
10. points_rules
11. resources
12. sessions
13. users
14. visit_logs

用户总数: 2
资源总数: 0
分类总数: 5

=== 所有测试通过! ===
```

## 生产环境部署

### 1. 创建数据库

```sql
CREATE DATABASE resource_share_site
DEFAULT CHARACTER SET utf8mb4
COLLATE utf8mb4_unicode_ci;
```

### 2. 创建用户

```sql
CREATE USER 'root'@'%' IDENTIFIED BY '123456';
GRANT SELECT, INSERT, UPDATE, DELETE ON resource_share_site.* TO 'root'@'%';
GRANT INDEX, CREATE, ALTER ON resource_share_site.* TO 'root'@'%';
FLUSH PRIVILEGES;
```

### 3. 执行建表脚本

```bash
# 使用 MySQL 客户端执行
mysql -u root -p resource_share_site < docs/mysql-schema.sql
```

### 4. 配置环境变量

```bash
export DB_TYPE=mysql
export DB_HOST=localhost
export DB_PORT=3306
export DB_NAME=resource_share_site
export DB_USER=root
export DB_PASSWORD=123456
export DB_CHARSET=utf8mb4
```

### 5. 修改默认密码

```sql
UPDATE users SET password_hash = '$2a$10$...' WHERE username = 'admin';
```

## 监控和维护

### 1. 查看数据库状态

```go
database.GetMigrationStatus(db)
```

### 2. 定期优化

```sql
-- 分析表统计信息
ANALYZE TABLE users, resources, comments;

-- 优化表
OPTIMIZE TABLE visit_logs;
```

### 3. 备份

```bash
# 完整备份
mysqldump -u root -p --single-transaction resource_share_site > backup_$(date +%Y%m%d_%H%M%S).sql

# 恢复
mysql -u root -p resource_share_site < backup_20251031_120000.sql
```

## 常见问题

### Q1: 连接数据库失败

**A**: 检查数据库配置和网络连接

```go
// 验证配置
dbConfig := &config.DatabaseConfig{
    Type:     "mysql",
    Host:     "localhost",
    Port:     "3306",
    Name:     "resource_share_site",
    User:     "root",
    Password: "123456",
    Charset:  "utf8mb4",
}
```

### Q2: 迁移失败

**A**: 检查模型定义和外键约束

```go
// 确保所有模型都已导入
db.AutoMigrate(
    &model.User{},
    &model.Category{},
    // ... 所有模型
)
```

### Q3: 积分不一致

**A**: 使用事务确保数据一致性

```go
err := db.Transaction(func(tx *gorm.DB) error {
    // 更新用户积分
    // 记录积分流水
    // 同时完成或同时失败
})
```

### Q4: 查询性能慢

**A**: 检查索引和查询语句

```sql
-- 查看慢查询
SHOW VARIABLES LIKE 'slow_query_log';

-- 查看索引使用情况
SHOW INDEX FROM users;
```

## 相关文档

- [数据库架构设计](database-architecture.md) - 完整的数据库架构设计文档
- [快速参考指南](database-quick-reference.md) - 查询示例和索引建议
- [MySQL建表脚本](mysql-schema.sql) - 完整的MySQL建表脚本
- [OpenSpec变更提案](../openspec/changes/build-resource-sharing-platform/) - 原始需求和设计

## 贡献指南

修改数据库结构时：
1. 更新对应的模型定义
2. 创建迁移脚本
3. 更新文档
4. 运行测试验证
5. 更新建表脚本

## 许可证

本项目采用 MIT 许可证。

---

**维护者**: Felix Wang  
**邮箱**: felixwang.biz@gmail.com
**最后更新**: 2025-10-31
