# 数据库架构设计文档

## 1. 概述

本文档描述了资源分享平台的完整数据库架构设计，支持用户管理、资源分享、积分系统、评论审核、邀请机制等核心业务功能。

**数据库版本**: MySQL 9.0+
**编码**: UTF8MB4
**引擎**: InnoDB
**表数量**: 14张核心业务表

## 2. 核心设计原则

### 2.1 软删除设计
- 所有业务表均使用 `DeletedAt gorm.DeletedAt` 字段实现软删除
- 避免物理删除导致的数据完整性问题
- 支持数据恢复和审计需求

### 2.2 时间戳管理
- 所有表包含 `CreatedAt` 和 `UpdatedAt` 字段
- 统一时间格式：RFC3339 (Go) / DATETIME (MySQL)
- 支持时区转换和历史追踪

### 2.3 状态机模式
- 资源状态：pending → approved/rejected
- 评论状态：pending → approved/rejected
- 邀请状态：pending → completed/expired
- 统一的状态转换逻辑确保数据一致性

### 2.4 自关联设计
- 用户邀请关系：User → Invitation → User
- 分类层级关系：Category → Category（支持多级分类）
- 评论嵌套回复：Comment → Comment（支持评论回复）

### 2.5 审计追踪
- 管理员操作日志：记录所有关键操作
- 审核流程追踪：记录审核人、时间、意见
- 积分变动记录：完整的积分流水账

## 3. 数据表详细设计

### 3.1 用户系统相关表

#### 3.1.1 users - 用户表
**核心表，用户系统的基石**

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 用户唯一标识 |
| username | VARCHAR(50) | UNIQUE, NOT NULL | 用户名，3-50字符 |
| email | VARCHAR(100) | UNIQUE, NOT NULL | 邮箱地址 |
| password_hash | VARCHAR(255) | NOT NULL | 密码哈希值 |
| role | VARCHAR(20) | DEFAULT 'user' | 角色：admin/admin, user |
| status | VARCHAR(20) | DEFAULT 'active' | 状态：active, banned |
| can_upload | BOOLEAN | DEFAULT FALSE | 上传权限控制 |
| invite_code | VARCHAR(36) | UNIQUE | 邀请码 |
| invited_by_id | BIGINT | FOREIGN KEY | 邀请人ID |
| points_balance | INT | DEFAULT 0 | 积分余额 |
| uploaded_resources_count | INT | DEFAULT 0 | 上传资源数 |
| downloaded_resources_count | INT | DEFAULT 0 | 下载资源数 |
| last_login_at | DATETIME | NULL | 最后登录时间 |
| created_at | DATETIME | NOT NULL | 创建时间 |
| updated_at | DATETIME | NOT NULL | 更新时间 |
| deleted_at | DATETIME | NULL | 软删除时间 |

**索引设计**:
- PRIMARY KEY (id)
- UNIQUE INDEX idx_users_username (username)
- UNIQUE INDEX idx_users_email (email)
- UNIQUE INDEX idx_users_invite_code (invite_code)
- INDEX idx_users_invited_by_id (invited_by_id)
- INDEX idx_users_role (role)
- INDEX idx_users_status (status)

**关联关系**:
- 一对多：User → Resource（上传播放者）
- 一对多：User → Comment（评论者）
- 一对多：User → PointRecord（积分记录）
- 一对多：User → Invitation（发送邀请）
- 一对多：User → Invitation（接收邀请）
- 自关联：User → User（邀请关系）
- 一对多：User → AdminLog（管理员操作）

#### 3.1.2 sessions - 会话表
**基于Session的认证机制**

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 会话ID |
| user_id | BIGINT | FOREIGN KEY, NOT NULL | 用户ID |
| session_id | VARCHAR(100) | UNIQUE, NOT NULL | 会话标识符 |
| data | TEXT | NULL | 会话数据（JSON格式） |
| expires_at | DATETIME | INDEX | 过期时间 |
| ip | VARCHAR(45) | NULL | IP地址（支持IPv6） |
| created_at | DATETIME | NOT NULL | 创建时间 |
| updated_at | DATETIME | NOT NULL | 更新时间 |

**索引设计**:
- PRIMARY KEY (id)
- UNIQUE INDEX idx_sessions_session_id (session_id)
- INDEX idx_sessions_user_id (user_id)
- INDEX idx_sessions_expires_at (expires_at)

### 3.2 资源系统相关表

#### 3.2.1 categories - 分类表
**层级分类系统**

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 分类ID |
| name | VARCHAR(50) | UNIQUE, NOT NULL | 分类名称 |
| description | VARCHAR(255) | NULL | 分类描述 |
| icon | VARCHAR(100) | NULL | 图标URL |
| color | VARCHAR(20) | NULL | 主题色 |
| parent_id | BIGINT | FOREIGN KEY | 父分类ID |
| sort_order | INT | DEFAULT 0 | 排序权重 |
| resources_count | INT | DEFAULT 0 | 资源数量 |
| created_at | DATETIME | NOT NULL | 创建时间 |
| updated_at | DATETIME | NOT NULL | 更新时间 |
| deleted_at | DATETIME | NULL | 软删除时间 |

**索引设计**:
- PRIMARY KEY (id)
- UNIQUE INDEX idx_categories_name (name)
- INDEX idx_categories_parent_id (parent_id)
- INDEX idx_categories_sort_order (sort_order)

**层级关系**:
- 自关联：Category → Category（多级分类）
- 一对多：Category → Resource

#### 3.2.2 resources - 资源表
**核心业务表，存储所有共享资源**

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 资源ID |
| title | VARCHAR(200) | NOT NULL | 资源标题 |
| description | TEXT | NULL | 资源描述 |
| category_id | BIGINT | FOREIGN KEY, NOT NULL | 分类ID |
| netdisk_url | VARCHAR(500) | NOT NULL | 网盘链接 |
| points_price | INT | DEFAULT 0 | 积分价格（0=免费） |
| source | VARCHAR(20) | NOT NULL | 来源：manual/user/crawler/excel |
| uploaded_by_id | BIGINT | FOREIGN KEY, NOT NULL | 上传者ID |
| status | VARCHAR(20) | DEFAULT 'pending' | 状态：pending/approved/rejected |
| reviewed_by_id | BIGINT | FOREIGN KEY | 审核人ID |
| reviewed_at | DATETIME | NULL | 审核时间 |
| review_notes | VARCHAR(500) | NULL | 审核意见 |
| downloads_count | INT | DEFAULT 0 | 下载次数 |
| views_count | INT | DEFAULT 0 | 查看次数 |
| tags | TEXT | NULL | 标签（JSON格式） |
| import_task_id | BIGINT | FOREIGN KEY | 导入任务ID |
| created_at | DATETIME | NOT NULL | 创建时间 |
| updated_at | DATETIME | NOT NULL | 更新时间 |
| deleted_at | DATETIME | NULL | 软删除时间 |

**索引设计**:
- PRIMARY KEY (id)
- INDEX idx_resources_category_id (category_id)
- INDEX idx_resources_uploaded_by_id (uploaded_by_id)
- INDEX idx_resources_status (status)
- INDEX idx_resources_source (source)
- INDEX idx_resources_points_price (points_price)
- INDEX idx_resources_import_task_id (import_task_id)
- FULLTEXT INDEX idx_resources_title_description (title, description)

**状态流转**:
- pending → approved（审核通过）
- pending → rejected（审核拒绝）
- approved → rejected（撤销发布）

**关联关系**:
- 多对一：Resource → Category（分类）
- 多对一：Resource → User（上传播放者）
- 多对一：Resource → User（审核人）
- 一对多：Resource → Comment
- 一对多：Resource → PointRecord（积分消费记录）
- 多对一：Resource → ImportTask（导入来源）

### 3.3 社交系统相关表

#### 3.3.1 comments - 评论表
**支持审核的评论系统**

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 评论ID |
| content | TEXT | NOT NULL | 评论内容（最大1000字符） |
| user_id | BIGINT | FOREIGN KEY, NOT NULL | 评论者ID |
| resource_id | BIGINT | FOREIGN KEY, NOT NULL | 资源ID |
| status | VARCHAR(20) | DEFAULT 'pending' | 状态：pending/approved/rejected |
| reviewed_by_id | BIGINT | FOREIGN KEY | 审核人ID |
| reviewed_at | DATETIME | NULL | 审核时间 |
| review_notes | VARCHAR(500) | NULL | 审核意见 |
| parent_id | BIGINT | FOREIGN KEY | 父评论ID（支持回复） |
| created_at | DATETIME | NOT NULL | 创建时间 |
| updated_at | DATETIME | NOT NULL | 更新时间 |
| deleted_at | DATETIME | NULL | 软删除时间 |

**索引设计**:
- PRIMARY KEY (id)
- INDEX idx_comments_user_id (user_id)
- INDEX idx_comments_resource_id (resource_id)
- INDEX idx_comments_status (status)
- INDEX idx_comments_parent_id (parent_id)
- INDEX idx_comments_reviewed_by_id (reviewed_by_id)

**嵌套评论**:
- parent_id = NULL：顶级评论
- parent_id ≠ NULL：回复评论
- 支持多级回复（实际使用时建议限制深度）

#### 3.3.2 invitations - 邀请表
**邀请注册机制**

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 邀请ID |
| inviter_id | BIGINT | FOREIGN KEY, NOT NULL | 邀请人ID |
| invitee_id | BIGINT | FOREIGN KEY | 被邀请人ID |
| invite_code | VARCHAR(36) | UNIQUE, NOT NULL | 邀请码 |
| status | VARCHAR(20) | DEFAULT 'pending' | 状态：pending/completed/expired |
| points_awarded | INT | DEFAULT 0 | 奖励积分数 |
| awarded_at | DATETIME | NULL | 奖励发放时间 |
| expires_at | DATETIME | INDEX | 过期时间 |
| created_at | DATETIME | NOT NULL | 创建时间 |
| updated_at | DATETIME | NOT NULL | 更新时间 |

**索引设计**:
- PRIMARY KEY (id)
- UNIQUE INDEX idx_invitations_invite_code (invite_code)
- INDEX idx_invitations_inviter_id (inviter_id)
- INDEX idx_invitations_invitee_id (invitee_id)
- INDEX idx_invitations_status (status)
- INDEX idx_invitations_expires_at (expires_at)

**生命周期**:
1. 邀请创建：pending，30天有效期
2. 被邀请人注册：invitee_id填充，状态可能变更
3. 奖励发放：points_awarded > 0，awarded_at记录时间
4. 过期处理：expires_at < NOW() 且状态为pending → expired

### 3.4 积分系统相关表

#### 3.4.1 points_rules - 积分规则表
**可配置的积分规则**

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 规则ID |
| rule_key | VARCHAR(50) | UNIQUE, NOT NULL | 规则键名（如：invite_reward） |
| rule_name | VARCHAR(100) | NOT NULL | 规则名称 |
| description | VARCHAR(255) | NULL | 规则描述 |
| points | INT | NOT NULL | 积分数量（可为负） |
| is_enabled | BOOLEAN | DEFAULT TRUE | 是否启用 |
| created_at | DATETIME | NOT NULL | 创建时间 |
| updated_at | DATETIME | NOT NULL | 更新时间 |

**预置规则**:
- invite_reward：邀请奖励（+50分）
- resource_download：资源下载消费（-X分）
- daily_checkin：每日签到（+5分）
- upload_reward：上传奖励（+10分）

#### 3.4.2 point_records - 积分记录表
**完整的积分流水账**

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 记录ID |
| user_id | BIGINT | FOREIGN KEY, NOT NULL | 用户ID |
| type | VARCHAR(10) | NOT NULL | 类型：income/expense |
| points | INT | NOT NULL | 积分变动（正数=收入，负数=支出） |
| balance_after | INT | NOT NULL | 变动后余额 |
| source | VARCHAR(30) | NOT NULL | 来源 |
| resource_id | BIGINT | FOREIGN KEY | 关联资源ID |
| invitation_id | BIGINT | FOREIGN KEY | 关联邀请ID |
| description | VARCHAR(255) | NULL | 描述信息 |
| operated_by_id | BIGINT | FOREIGN KEY | 操作人ID（管理员） |
| created_at | DATETIME | NOT NULL | 创建时间 |

**索引设计**:
- PRIMARY KEY (id)
- INDEX idx_point_records_user_id (user_id)
- INDEX idx_point_records_type (type)
- INDEX idx_point_records_source (source)
- INDEX idx_point_records_resource_id (resource_id)
- INDEX idx_point_records_invitation_id (invitation_id)
- INDEX idx_point_records_operated_by_id (operated_by_id)

**数据一致性**:
- balance_after = 上一次记录 + points
- 通过事务确保积分变动原子性

### 3.5 监控审计相关表

#### 3.5.1 visit_logs - 访问日志表
**全量访问记录**

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 日志ID |
| user_id | BIGINT | FOREIGN KEY | 用户ID（可为空） |
| ip | VARCHAR(45) | NOT NULL | IP地址（支持IPv6） |
| path | VARCHAR(500) | NOT NULL | 访问路径 |
| method | VARCHAR(10) | NOT NULL | HTTP方法 |
| user_agent | VARCHAR(500) | NULL | User-Agent |
| referer | VARCHAR(500) | NULL | 来源页面 |
| device_type | VARCHAR(20) | NULL | 设备类型 |
| os | VARCHAR(50) | NULL | 操作系统 |
| browser | VARCHAR(50) | NULL | 浏览器 |
| country | VARCHAR(50) | NULL | 国家 |
| city | VARCHAR(50) | NULL | 城市 |
| status_code | INT | NOT NULL | HTTP状态码 |
| response_time | BIGINT | NOT NULL | 响应时间（毫秒） |
| session_id | VARCHAR(100) | NULL | 会话ID |
| created_at | DATETIME | NOT NULL | 访问时间 |

**索引设计**:
- PRIMARY KEY (id)
- INDEX idx_visit_logs_user_id (user_id)
- INDEX idx_visit_logs_ip (ip)
- INDEX idx_visit_logs_path (path)
- INDEX idx_visit_logs_created_at (created_at)
- INDEX idx_visit_logs_session_id (session_id)

**应用场景**:
- Dashboard访问统计
- 用户行为分析
- 安全审计
- 性能监控

#### 3.5.2 ip_blacklists - IP黑名单表
**IP访问控制**

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 记录ID |
| ip | VARCHAR(45) | UNIQUE, NOT NULL | IP地址 |
| reason | VARCHAR(255) | NOT NULL | 禁止原因 |
| banned_by_id | BIGINT | FOREIGN KEY, NOT NULL | 禁止者ID |
| banned_at | DATETIME | NOT NULL | 禁止时间 |
| expires_at | DATETIME | NULL | 过期时间（NULL=永久） |
| access_count | INT | DEFAULT 0 | 访问次数 |
| last_access_at | DATETIME | NULL | 最后访问时间 |
| created_at | DATETIME | NOT NULL | 创建时间 |
| updated_at | DATETIME | NOT NULL | 更新时间 |

**索引设计**:
- PRIMARY KEY (id)
- UNIQUE INDEX idx_ip_blacklists_ip (ip)
- INDEX idx_ip_blacklists_banned_by_id (banned_by_id)
- INDEX idx_ip_blacklists_expires_at (expires_at)

#### 3.5.3 admin_logs - 管理员操作日志表
**完整的审计追踪**

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 日志ID |
| admin_id | BIGINT | FOREIGN KEY, NOT NULL | 管理员ID |
| action | VARCHAR(50) | NOT NULL | 操作类型 |
| target_type | VARCHAR(50) | NOT NULL | 目标类型 |
| target_id | BIGINT | NOT NULL | 目标ID |
| before_data | TEXT | NULL | 操作前数据（JSON） |
| after_data | TEXT | NULL | 操作后数据（JSON） |
| ip | VARCHAR(45) | NULL | 操作IP |
| user_agent | VARCHAR(500) | NULL | User-Agent |
| created_at | DATETIME | NOT NULL | 操作时间 |

**索引设计**:
- PRIMARY KEY (id)
- INDEX idx_admin_logs_admin_id (admin_id)
- INDEX idx_admin_logs_action (action)
- INDEX idx_admin_logs_target_type_id (target_type, target_id)
- INDEX idx_admin_logs_created_at (created_at)

### 3.6 系统管理相关表

#### 3.6.1 ads - 广告表
**广告位管理**

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 广告ID |
| title | VARCHAR(100) | NOT NULL | 广告标题 |
| image_url | VARCHAR(500) | NOT NULL | 图片链接 |
| link_url | VARCHAR(500) | NOT NULL | 跳转链接 |
| ad_position | VARCHAR(50) | NOT NULL | 广告位置 |
| is_active | BOOLEAN | DEFAULT TRUE | 是否启用 |
| sort_order | INT | DEFAULT 0 | 排序 |
| click_count | INT | DEFAULT 0 | 点击次数 |
| start_date | DATETIME | NULL | 开始日期 |
| end_date | DATETIME | NULL | 结束日期 |
| created_at | DATETIME | NOT NULL | 创建时间 |
| updated_at | DATETIME | NOT NULL | 更新时间 |
| deleted_at | DATETIME | NULL | 软删除时间 |

**索引设计**:
- PRIMARY KEY (id)
- INDEX idx_ads_position_sort (ad_position, sort_order)
- INDEX idx_ads_active_date (is_active, start_date, end_date)

#### 3.6.2 permissions - 权限配置表
**细粒度权限控制**

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 权限ID |
| key | VARCHAR(50) | UNIQUE, NOT NULL | 权限键名 |
| name | VARCHAR(100) | NOT NULL | 权限名称 |
| description | VARCHAR(255) | NULL | 权限描述 |
| is_enabled | BOOLEAN | DEFAULT FALSE | 是否启用 |
| created_at | DATETIME | NOT NULL | 创建时间 |
| updated_at | DATETIME | NOT NULL | 更新时间 |

#### 3.6.3 import_tasks - 导入任务表
**资源批量导入追踪**

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 任务ID |
| task_type | VARCHAR(30) | NOT NULL | 任务类型：crawler/excel |
| status | VARCHAR(20) | NOT NULL | 状态：pending/running/completed/failed |
| total_count | INT | DEFAULT 0 | 总数 |
| success_count | INT | DEFAULT 0 | 成功数 |
| fail_count | INT | DEFAULT 0 | 失败数 |
| config_data | TEXT | NULL | 配置信息（JSON） |
| error_log | TEXT | NULL | 错误日志 |
| started_at | DATETIME | NULL | 开始时间 |
| completed_at | DATETIME | NULL | 完成时间 |
| created_at | DATETIME | NOT NULL | 创建时间 |
| updated_at | DATETIME | NOT NULL | 更新时间 |

## 4. 实体关系图（ERD）

```
┌─────────────┐
│    users    │◄──────────┐
└─────────────┘           │
      │                  │
      │                  │
      │                  │
      ▼                  ▼
┌─────────────┐    ┌─────────────┐
│  sessions   │    │ invitations │
└─────────────┘    └─────────────┘
                           │
                           ▼
┌─────────────┐    ┌─────────────┐
│ resources   │    │    users    │
│             │    │  (invited)  │
└─────────────┘    └─────────────┘
      │
      │
      ▼
┌─────────────┐
│  comments   │
└─────────────┘

┌─────────────┐
│ categories  │
└─────────────┘
      │
      ▼
┌─────────────┐
│  resources  │
└─────────────┘

┌─────────────┐
│point_records│
└─────────────┘
      │
      ▼
┌─────────────┐
│    users    │
└─────────────┘

┌─────────────┐
│ point_rules │
└─────────────┘

┌─────────────┐
│  visit_logs │
└─────────────┘

┌─────────────┐
│ip_blacklists│
└─────────────┘

┌─────────────┐
│  admin_logs │
└─────────────┘

┌─────────────┐
│     ads     │
└─────────────┘

┌─────────────┐
│permissions  │
└─────────────┘

┌─────────────┐
│import_tasks │
└─────────────┘
```

## 5. 关键索引策略

### 5.1 主键索引
- 所有表使用 BIGINT AUTO_INCREMENT 作为主键
- 聚簇索引，提高查询性能

### 5.2 唯一索引
- username, email, invite_code：确保唯一性
- session_id, invite_code：防止重复

### 5.3 复合索引
- (ad_position, sort_order)：广告排序查询
- (target_type, target_id)：管理员日志查询
- (status, created_at)：状态筛选+时间排序

### 5.4 时间索引
- CreatedAt/UpdatedAt：时间范围查询
- ExpiresAt：过期任务清理
- LastLoginAt：活跃用户统计

### 5.5 外键索引
- 所有外键字段均建立索引
- 提高JOIN查询性能

### 5.6 全文索引
- resources表：title + description
- 支持中文全文搜索（需配置分词器）

## 6. 数据一致性保证

### 6.1 外键约束
- 所有关联关系通过外键维护
- 级联删除策略：
  - 用户删除：软删除，不级联
  - 分类删除：软删除，不级联
  - 资源删除：软删除，关联评论一并软删除

### 6.2 事务边界
- 积分变动：PointRecord + User.points_balance 原子更新
- 邀请奖励：Invitation.status + User.points_balance 原子更新
- 审核操作：Resource/Comment.status + 审核记录 原子更新

### 6.3 业务规则校验
- 积分余额不能为负数
- 邀请码唯一且有效
- 资源状态流转受控
- 评论审核后才能显示

## 7. 性能优化建议

### 7.1 分区策略
**visit_logs表按月分区**
- 减少单表数据量
- 提高查询效率
- 便于历史数据归档

### 7.2 读写分离
- 写操作：主库（Insert, Update）
- 读操作：从库（Select）
- 统计数据：定期同步到缓存

### 7.3 缓存策略
- 用户信息：Redis缓存
- 分类列表：本地缓存
- 热门资源：LRU缓存

### 7.4 数据归档
- visit_logs：保留6个月
- admin_logs：保留1年
- 软删除数据：定期物理清理

## 8. 安全考虑

### 8.1 数据加密
- 密码：bcrypt哈希存储
- 敏感字段：AES加密（如手机号）

### 8.2 SQL注入防护
- 使用GORM参数化查询
- 禁止字符串拼接SQL

### 8.3 访问控制
- IP黑名单实时生效
- Session超时自动登出
- 敏感操作二次验证

## 9. 扩展性设计

### 9.1 分库分表准备
- 用户表：按ID哈希分表
- 资源表：按分类分表
- 日志表：按时间分区

### 9.2 垂直拆分
- 用户核心信息：users表
- 用户扩展信息：user_profiles表
- 历史数据：archive_xxx表

### 9.3 水平扩展
- 读写分离
- 多级缓存
- CDN加速静态资源

## 10. 维护和监控

### 10.1 定期维护
- 每周：ANALYZE TABLE更新统计信息
- 每月：OPTIMIZE TABLE整理碎片
- 每季度：归档历史数据

### 10.2 监控指标
- 连接数：max_connections使用率
- 慢查询：slow_query_log分析
- 磁盘空间：数据文件大小
- 索引使用：unused indexes清理

### 10.3 备份恢复
- 全量备份：每天凌晨执行
- 增量备份：每小时执行
- 恢复测试：每月验证

## 11. 迁移脚本

详见：`docs/mysql-schema.sql`

执行顺序：
1. 创建数据库
2. 创建用户和授权
3. 执行建表脚本
4. 插入初始数据
5. 创建索引
6. 验证迁移结果

## 12. 版本历史

| 版本 | 日期 | 变更内容 |
|------|------|----------|
| v1.0 | 2025-10-31 | 初始版本，14张表架构 |

---

**文档维护**: Felix Wang  
**邮箱**: felixwang.biz@gmail.com
**最后更新**: 2025-10-31
