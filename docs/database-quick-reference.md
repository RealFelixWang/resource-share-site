# 数据库快速参考指南

## 📋 概述

- **数据库版本**: MySQL 9.0+
- **表总数**: 14张
- **字符集**: UTF8MB4
- **引擎**: InnoDB

## 📊 核心表概览

| 表名 | 说明 | 记录示例 |
|------|------|----------|
| **用户系统** |
| users | 用户基本信息 | admin, 普通用户 |
| sessions | 会话管理 | 登录状态 |
| **资源系统** |
| categories | 资源分类 | 软件工具、电子资料 |
| resources | 资源信息 | 网盘链接、积分价格 |
| comments | 用户评论 | 待审核、已通过 |
| **邀请系统** |
| invitations | 邀请关系 | 邀请码、奖励积分 |
| **积分系统** |
| points_rules | 积分规则 | 邀请奖励、下载消费 |
| point_records | 积分流水 | 每笔积分变动 |
| **监控审计** |
| visit_logs | 访问日志 | IP、路径、响应时间 |
| ip_blacklists | IP黑名单 | 禁止IP及原因 |
| admin_logs | 管理员日志 | 操作类型、变更前后数据 |
| **系统管理** |
| ads | 广告管理 | 首页banner、侧边栏 |
| permissions | 权限配置 | 用户上传、审核资源 |
| import_tasks | 导入任务 | 爬虫抓取、Excel导入 |

## 🔑 关键字段说明

### 用户状态 (status)
- `active`: 正常用户
- `banned`: 已封禁

### 用户角色 (role)
- `user`: 普通用户
- `admin`: 管理员

### 资源状态 (status)
- `pending`: 待审核
- `approved`: 已发布
- `rejected`: 已拒绝

### 资源来源 (source)
- `manual`: 管理员手动添加
- `user`: 用户上传
- `crawler`: 爬虫抓取
- `excel`: Excel导入

### 评论状态 (status)
- `pending`: 待审核
- `approved`: 已通过
- `rejected`: 已拒绝

### 邀请状态 (status)
- `pending`: 待注册
- `completed`: 已完成
- `expired`: 已过期

### 积分类型 (type)
- `income`: 收入
- `expense`: 支出

### 积分来源 (source)
- `invite_reward`: 邀请奖励
- `resource_download`: 资源下载消费
- `daily_checkin`: 每日签到
- `upload_reward`: 上传奖励

## 🗝️ 主键和外键关系

```
users (id)
  ├── sessions (user_id)
  ├── resources (uploaded_by_id, reviewed_by_id)
  ├── comments (user_id, reviewed_by_id)
  ├── invitations (inviter_id, invitee_id)
  ├── point_records (user_id, operated_by_id)
  ├── visit_logs (user_id)
  ├── ip_blacklists (banned_by_id)
  ├── admin_logs (admin_id)
  └── self-reference (invited_by_id)

categories (id)
  └── resources (category_id)
  └── self-reference (parent_id)

resources (id)
  ├── comments (resource_id)
  └── point_records (resource_id)

invitations (id)
  └── point_records (invitation_id)

import_tasks (id)
  └── resources (import_task_id)

comments (id)
  └── self-reference (parent_id)
```

## 📈 常用查询示例

### 1. 获取用户积分统计

```sql
-- 累计收入
SELECT SUM(points) 
FROM point_records 
WHERE user_id = ? AND type = 'income';

-- 累计支出
SELECT SUM(ABS(points)) 
FROM point_records 
WHERE user_id = ? AND type = 'expense';

-- 当前余额
SELECT points_balance 
FROM users 
WHERE id = ?;
```

### 2. 资源审核统计

```sql
-- 按状态分组
SELECT 
    status,
    COUNT(*) as count
FROM resources
GROUP BY status;

-- 每日新增资源
SELECT 
    DATE(created_at) as date,
    COUNT(*) as count
FROM resources
WHERE created_at >= DATE_SUB(NOW(), INTERVAL 30 DAY)
GROUP BY DATE(created_at)
ORDER BY date DESC;
```

### 3. 用户活跃度统计

```sql
-- 活跃用户（30天内登录）
SELECT 
    u.id,
    u.username,
    u.last_login_at
FROM users u
WHERE u.last_login_at >= DATE_SUB(NOW(), INTERVAL 30 DAY)
    AND u.status = 'active'
ORDER BY u.last_login_at DESC;

-- 新用户注册（按月统计）
SELECT 
    DATE_FORMAT(created_at, '%Y-%m') as month,
    COUNT(*) as count
FROM users
WHERE created_at >= DATE_SUB(NOW(), INTERVAL 12 MONTH)
GROUP BY DATE_FORMAT(created_at, '%Y-%m')
ORDER BY month;
```

### 4. 评论审核统计

```sql
-- 待审核评论数
SELECT COUNT(*) 
FROM comments 
WHERE status = 'pending';

-- 审核通过率
SELECT 
    SUM(CASE WHEN status = 'approved' THEN 1 ELSE 0 END) / COUNT(*) * 100 as approval_rate
FROM comments;
```

### 5. 访问日志分析

```sql
-- 热门页面（Top 10）
SELECT 
    path,
    COUNT(*) as visits,
    COUNT(DISTINCT ip) as unique_visitors,
    AVG(response_time) as avg_response_time
FROM visit_logs
WHERE created_at >= DATE_SUB(NOW(), INTERVAL 7 DAY)
GROUP BY path
ORDER BY visits DESC
LIMIT 10;

-- 流量趋势（每日）
SELECT 
    DATE(created_at) as date,
    COUNT(DISTINCT ip) as unique_visitors,
    COUNT(*) as total_visits
FROM visit_logs
WHERE created_at >= DATE_SUB(NOW(), INTERVAL 30 DAY)
GROUP BY DATE(created_at)
ORDER BY date;
```

### 6. IP黑名单管理

```sql
-- 活跃黑名单（未过期）
SELECT 
    ip,
    reason,
    banned_at,
    access_count
FROM ip_blacklists
WHERE expires_at IS NULL 
    OR expires_at > NOW()
ORDER BY banned_at DESC;
```

## 📝 索引使用建议

### 必须创建的索引

```sql
-- 用户表
CREATE INDEX idx_users_invited_by_id ON users(invited_by_id);

-- 资源表
CREATE INDEX idx_resources_status ON resources(status);
CREATE INDEX idx_resources_category_id ON resources(category_id);

-- 评论表
CREATE INDEX idx_comments_resource_id ON comments(resource_id);
CREATE INDEX idx_comments_status ON comments(status);

-- 积分记录表
CREATE INDEX idx_point_records_user_id ON point_records(user_id);
CREATE INDEX idx_point_records_created_at ON point_records(created_at);

-- 访问日志表
CREATE INDEX idx_visit_logs_created_at ON visit_logs(created_at);
CREATE INDEX idx_visit_logs_ip ON visit_logs(ip);
```

### 复合索引

```sql
-- 资源筛选（状态+分类）
CREATE INDEX idx_resources_status_category ON resources(status, category_id);

-- 积分记录查询（用户+时间）
CREATE INDEX idx_point_records_user_time ON point_records(user_id, created_at);

-- 管理员日志查询（操作+时间）
CREATE INDEX idx_admin_logs_action_time ON admin_logs(action, created_at);
```

## 🔄 数据一致性规则

### 1. 积分变动原子性

```go
// 伪代码示例
func DeductPoints(db *gorm.DB, userID uint, points int, resourceID uint) error {
    return db.Transaction(func(tx *gorm.DB) error {
        // 1. 检查余额
        var user model.User
        if err := tx.First(&user, userID).Error; err != nil {
            return err
        }
        if user.PointsBalance < points {
            return errors.New("积分余额不足")
        }

        // 2. 更新用户积分
        if err := tx.Model(&user).Update("points_balance", gorm.Expr("points_balance - ?", points)).Error; err != nil {
            return err
        }

        // 3. 记录积分流水
        record := model.PointRecord{
            UserID:       userID,
            Type:         model.PointTypeExpense,
            Points:       -points,
            BalanceAfter: user.PointsBalance - points,
            Source:       model.PointSourceResourceDownload,
            ResourceID:   &resourceID,
            Description:  "下载资源消费",
        }
        return tx.Create(&record).Error
    })
}
```

### 2. 邀请奖励原子性

```go
// 伪代码示例
func AwardInvitePoints(db *gorm.DB, invitationID uint, adminID uint) error {
    return db.Transaction(func(tx *gorm.DB) error {
        // 获取邀请信息
        var invitation model.Invitation
        if err := tx.First(&invitation, invitationID).Error; err != nil {
            return err
        }

        if invitation.Status != model.InvitationStatusCompleted {
            return errors.New("邀请未完成")
        }

        // 获取积分规则
        var rule model.PointsRule
        if err := tx.Where("rule_key = ?", "invite_reward").First(&rule).Error; err != nil {
            return err
        }

        // 奖励邀请人积分
        inviter := invitation.Inviter
        newBalance := inviter.PointsBalance + rule.Points
        if err := tx.Model(&inviter).Update("points_balance", newBalance).Error; err != nil {
            return err
        }

        // 记录积分流水
        record := model.PointRecord{
            UserID:       inviter.ID,
            Type:         model.PointTypeIncome,
            Points:       rule.Points,
            BalanceAfter: newBalance,
            Source:       model.PointSourceInviteReward,
            InvitationID: &invitationID,
            Description:  "邀请奖励",
        }
        if err := tx.Create(&record).Error; err != nil {
            return err
        }

        // 更新邀请记录
        return tx.Model(&invitation).Updates(map[string]interface{}{
            "status":         model.InvitationStatusCompleted,
            "points_awarded": rule.Points,
            "awarded_at":     time.Now(),
        }).Error
    })
}
```

## ⚠️ 注意事项

### 1. 软删除
- 所有业务表都使用 `DeletedAt` 字段进行软删除
- 查询时需要使用 `Where("deleted_at IS NULL")` 或使用 GORM 的 `Unscoped()`
- 物理删除需谨慎，建议定期归档

### 2. 外键约束
- 资源删除会级联删除评论（`ON DELETE CASCADE`）
- 用户删除不会级联删除资源（`ON DELETE RESTRICT`）
- 分类删除不会级联删除资源（`ON DELETE RESTRICT`）

### 3. 分区表
- `visit_logs` 表按月分区，提高查询性能
- 定期归档历史数据，避免单分区过大

### 4. 全文索引
- `resources` 表的 `title` 和 `description` 字段建立了全文索引
- 支持中文搜索需配置中文分词器

## 🛠️ 维护命令

### 1. 检查表状态

```sql
-- 查看表大小
SELECT 
    table_name,
    ROUND((data_length + index_length) / 1024 / 1024, 2) AS 'DB Size (MB)'
FROM information_schema.TABLES
WHERE table_schema = 'resource_share_site'
ORDER BY (data_length + index_length) DESC;

-- 查看索引使用情况
SELECT 
    object_name,
    index_name,
    count_read,
    count_write,
    count_fetch,
    count_insert,
    count_update,
    count_delete
FROM performance_schema.table_io_waits_summary_by_index_usage
WHERE object_schema = 'resource_share_site'
ORDER BY count_read DESC;
```

### 2. 优化表

```sql
-- 分析表
ANALYZE TABLE users, resources, comments;

-- 优化表
OPTIMIZE TABLE visit_logs;

-- 检查表完整性
CHECK TABLE users, resources, comments;
```

### 3. 备份命令

```bash
# 完整备份
mysqldump -u root -p --single-transaction --routines --triggers resource_share_site > backup_$(date +%Y%m%d).sql

# 仅结构备份
mysqldump -u root -p --no-data resource_share_site > schema_$(date +%Y%m%d).sql

# 仅数据备份
mysqldump -u root -p --no-create-info resource_share_site > data_$(date +%Y%m%d).sql
```

## 📚 相关文档

- [完整数据库架构设计](database-architecture.md)
- [MySQL建表脚本](mysql-schema.sql)
- [OpenSpec变更提案](../openspec/changes/build-resource-sharing-platform/)
- [API设计文档](../api/)
- [数据模型说明](../model/)

---

**最后更新**: 2025-10-31
