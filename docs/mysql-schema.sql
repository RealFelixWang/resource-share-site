-- =====================================================
-- 资源分享平台数据库架构
-- 数据库版本: MySQL 9.0+
-- 字符集: UTF8MB4
-- 引擎: InnoDB
-- 创建时间: 2025-10-31
-- =====================================================

-- 创建数据库
CREATE DATABASE IF NOT EXISTS resource_share_site
DEFAULT CHARACTER SET utf8mb4
COLLATE utf8mb4_unicode_ci;

USE resource_share_site;

-- =====================================================
-- 1. 用户相关表
-- =====================================================

-- 1.1 users - 用户表
CREATE TABLE users (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    deleted_at DATETIME NULL,

    -- 基本信息
    username VARCHAR(50) NOT NULL,
    email VARCHAR(100) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,

    -- 用户状态
    role VARCHAR(20) NOT NULL DEFAULT 'user' COMMENT '角色: admin, user',
    status VARCHAR(20) NOT NULL DEFAULT 'active' COMMENT '状态: active, banned',

    -- 上传权限
    can_upload BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否有上传权限',

    -- 邀请相关
    invite_code VARCHAR(36) NOT NULL,
    invited_by_id BIGINT UNSIGNED NULL,
    points_balance INT NOT NULL DEFAULT 0 COMMENT '积分余额',

    -- 统计信息
    uploaded_resources_count INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '上传资源数',
    downloaded_resources_count INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '下载资源数',

    -- 最后登录时间
    last_login_at DATETIME NULL,

    PRIMARY KEY (id),
    UNIQUE KEY idx_users_username (username),
    UNIQUE KEY idx_users_email (email),
    UNIQUE KEY idx_users_invite_code (invite_code),
    KEY idx_users_invited_by_id (invited_by_id),
    KEY idx_users_role (role),
    KEY idx_users_status (status),
    KEY idx_users_deleted_at (deleted_at),

    CONSTRAINT fk_users_invited_by_id
        FOREIGN KEY (invited_by_id)
        REFERENCES users (id)
        ON DELETE SET NULL
        ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- 1.2 sessions - 会话表
CREATE TABLE sessions (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,

    user_id BIGINT UNSIGNED NOT NULL,
    session_id VARCHAR(100) NOT NULL,
    data TEXT NULL COMMENT '会话数据(JSON格式)',
    expires_at DATETIME NOT NULL,
    ip VARCHAR(45) NULL COMMENT 'IP地址(支持IPv6)',

    PRIMARY KEY (id),
    UNIQUE KEY idx_sessions_session_id (session_id),
    KEY idx_sessions_user_id (user_id),
    KEY idx_sessions_expires_at (expires_at),

    CONSTRAINT fk_sessions_user_id
        FOREIGN KEY (user_id)
        REFERENCES users (id)
        ON DELETE CASCADE
        ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='会话表';

-- =====================================================
-- 2. 分类系统
-- =====================================================

-- 2.1 categories - 分类表
CREATE TABLE categories (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    deleted_at DATETIME NULL,

    name VARCHAR(50) NOT NULL COMMENT '分类名称',
    description VARCHAR(255) NULL COMMENT '分类描述',
    icon VARCHAR(100) NULL COMMENT '图标URL',
    color VARCHAR(20) NULL COMMENT '主题色',

    -- 层级关系
    parent_id BIGINT UNSIGNED NULL,
    sort_order INT NOT NULL DEFAULT 0 COMMENT '排序权重',
    resources_count INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '资源数量',

    PRIMARY KEY (id),
    UNIQUE KEY idx_categories_name (name),
    KEY idx_categories_parent_id (parent_id),
    KEY idx_categories_sort_order (sort_order),
    KEY idx_categories_deleted_at (deleted_at),

    CONSTRAINT fk_categories_parent_id
        FOREIGN KEY (parent_id)
        REFERENCES categories (id)
        ON DELETE SET NULL
        ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='分类表';

-- =====================================================
-- 3. 资源系统
-- =====================================================

-- 3.1 resources - 资源表
CREATE TABLE resources (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    deleted_at DATETIME NULL,

    -- 基本信息
    title VARCHAR(200) NOT NULL COMMENT '资源标题',
    description TEXT NULL COMMENT '资源描述',
    category_id BIGINT UNSIGNED NOT NULL,

    -- 资源信息
    netdisk_url VARCHAR(500) NOT NULL COMMENT '网盘链接',
    points_price INT NOT NULL DEFAULT 0 COMMENT '积分价格(0=免费)',

    -- 来源信息
    source VARCHAR(20) NOT NULL COMMENT '来源: manual, user, crawler, excel',

    -- 上传者
    uploaded_by_id BIGINT UNSIGNED NOT NULL,

    -- 状态
    status VARCHAR(20) NOT NULL DEFAULT 'pending' COMMENT '状态: pending, approved, rejected',

    -- 审核信息
    reviewed_by_id BIGINT UNSIGNED NULL,
    reviewed_at DATETIME NULL,
    review_notes VARCHAR(500) NULL COMMENT '审核意见',

    -- 统计信息
    downloads_count INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '下载次数',
    views_count INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '查看次数',

    -- 标签(JSON格式)
    tags TEXT NULL,

    -- 导入任务ID
    import_task_id BIGINT UNSIGNED NULL,

    PRIMARY KEY (id),
    KEY idx_resources_category_id (category_id),
    KEY idx_resources_uploaded_by_id (uploaded_by_id),
    KEY idx_resources_status (status),
    KEY idx_resources_source (source),
    KEY idx_resources_points_price (points_price),
    KEY idx_resources_import_task_id (import_task_id),
    KEY idx_resources_deleted_at (deleted_at),
    FULLTEXT KEY idx_resources_title_description (title, description),

    CONSTRAINT fk_resources_category_id
        FOREIGN KEY (category_id)
        REFERENCES categories (id)
        ON DELETE RESTRICT
        ON UPDATE CASCADE,

    CONSTRAINT fk_resources_uploaded_by_id
        FOREIGN KEY (uploaded_by_id)
        REFERENCES users (id)
        ON DELETE RESTRICT
        ON UPDATE CASCADE,

    CONSTRAINT fk_resources_reviewed_by_id
        FOREIGN KEY (reviewed_by_id)
        REFERENCES users (id)
        ON DELETE SET NULL
        ON UPDATE CASCADE,

    CONSTRAINT fk_resources_import_task_id
        FOREIGN KEY (import_task_id)
        REFERENCES import_tasks (id)
        ON DELETE SET NULL
        ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='资源表';

-- =====================================================
-- 4. 评论系统
-- =====================================================

-- 4.1 comments - 评论表
CREATE TABLE comments (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    deleted_at DATETIME NULL,

    -- 基本信息
    content TEXT NOT NULL COMMENT '评论内容(最大1000字符)',

    -- 关联信息
    user_id BIGINT UNSIGNED NOT NULL,
    resource_id BIGINT UNSIGNED NOT NULL,

    -- 状态
    status VARCHAR(20) NOT NULL DEFAULT 'pending' COMMENT '状态: pending, approved, rejected',

    -- 审核信息
    reviewed_by_id BIGINT UNSIGNED NULL,
    reviewed_at DATETIME NULL,
    review_notes VARCHAR(500) NULL COMMENT '审核意见',

    -- 父评论(支持嵌套回复)
    parent_id BIGINT UNSIGNED NULL,

    PRIMARY KEY (id),
    KEY idx_comments_user_id (user_id),
    KEY idx_comments_resource_id (resource_id),
    KEY idx_comments_status (status),
    KEY idx_comments_parent_id (parent_id),
    KEY idx_comments_reviewed_by_id (reviewed_by_id),
    KEY idx_comments_deleted_at (deleted_at),

    CONSTRAINT fk_comments_user_id
        FOREIGN KEY (user_id)
        REFERENCES users (id)
        ON DELETE RESTRICT
        ON UPDATE CASCADE,

    CONSTRAINT fk_comments_resource_id
        FOREIGN KEY (resource_id)
        REFERENCES resources (id)
        ON DELETE CASCADE
        ON UPDATE CASCADE,

    CONSTRAINT fk_comments_reviewed_by_id
        FOREIGN KEY (reviewed_by_id)
        REFERENCES users (id)
        ON DELETE SET NULL
        ON UPDATE CASCADE,

    CONSTRAINT fk_comments_parent_id
        FOREIGN KEY (parent_id)
        REFERENCES comments (id)
        ON DELETE CASCADE
        ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评论表';

-- =====================================================
-- 5. 邀请系统
-- =====================================================

-- 5.1 invitations - 邀请表
CREATE TABLE invitations (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,

    -- 邀请关系
    inviter_id BIGINT UNSIGNED NOT NULL,
    invitee_id BIGINT UNSIGNED NULL,

    -- 邀请码
    invite_code VARCHAR(36) NOT NULL,

    -- 状态和积分奖励
    status VARCHAR(20) NOT NULL DEFAULT 'pending' COMMENT '状态: pending, completed, expired',
    points_awarded INT NOT NULL DEFAULT 0 COMMENT '奖励积分数',
    awarded_at DATETIME NULL COMMENT '奖励发放时间',

    -- 过期时间
    expires_at DATETIME NOT NULL COMMENT '过期时间',

    PRIMARY KEY (id),
    UNIQUE KEY idx_invitations_invite_code (invite_code),
    KEY idx_invitations_inviter_id (inviter_id),
    KEY idx_invitations_invitee_id (invitee_id),
    KEY idx_invitations_status (status),
    KEY idx_invitations_expires_at (expires_at),

    CONSTRAINT fk_invitations_inviter_id
        FOREIGN KEY (inviter_id)
        REFERENCES users (id)
        ON DELETE RESTRICT
        ON UPDATE CASCADE,

    CONSTRAINT fk_invitations_invitee_id
        FOREIGN KEY (invitee_id)
        REFERENCES users (id)
        ON DELETE SET NULL
        ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='邀请表';

-- =====================================================
-- 6. 积分系统
-- =====================================================

-- 6.1 points_rules - 积分规则表
CREATE TABLE points_rules (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,

    rule_key VARCHAR(50) NOT NULL COMMENT '规则键名',
    rule_name VARCHAR(100) NOT NULL COMMENT '规则名称',
    description VARCHAR(255) NULL COMMENT '规则描述',
    points INT NOT NULL COMMENT '积分数量(可为负)',
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE COMMENT '是否启用',

    PRIMARY KEY (id),
    UNIQUE KEY idx_points_rules_rule_key (rule_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='积分规则表';

-- 6.2 point_records - 积分记录表
CREATE TABLE point_records (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    created_at DATETIME NOT NULL,

    -- 用户信息
    user_id BIGINT UNSIGNED NOT NULL,

    -- 积分变动
    type VARCHAR(10) NOT NULL COMMENT '类型: income, expense',
    points INT NOT NULL COMMENT '积分变动(正数=收入,负数=支出)',
    balance_after INT NOT NULL COMMENT '变动后余额',

    -- 来源信息
    source VARCHAR(30) NOT NULL COMMENT '来源',

    -- 关联信息(可选)
    resource_id BIGINT UNSIGNED NULL,
    invitation_id BIGINT UNSIGNED NULL,

    -- 描述信息
    description VARCHAR(255) NULL COMMENT '描述信息',

    -- 操作人(管理员操作时)
    operated_by_id BIGINT UNSIGNED NULL,

    PRIMARY KEY (id),
    KEY idx_point_records_user_id (user_id),
    KEY idx_point_records_type (type),
    KEY idx_point_records_source (source),
    KEY idx_point_records_resource_id (resource_id),
    KEY idx_point_records_invitation_id (invitation_id),
    KEY idx_point_records_operated_by_id (operated_by_id),
    KEY idx_point_records_created_at (created_at),

    CONSTRAINT fk_point_records_user_id
        FOREIGN KEY (user_id)
        REFERENCES users (id)
        ON DELETE RESTRICT
        ON UPDATE CASCADE,

    CONSTRAINT fk_point_records_resource_id
        FOREIGN KEY (resource_id)
        REFERENCES resources (id)
        ON DELETE SET NULL
        ON UPDATE CASCADE,

    CONSTRAINT fk_point_records_invitation_id
        FOREIGN KEY (invitation_id)
        REFERENCES invitations (id)
        ON DELETE SET NULL
        ON UPDATE CASCADE,

    CONSTRAINT fk_point_records_operated_by_id
        FOREIGN KEY (operated_by_id)
        REFERENCES users (id)
        ON DELETE SET NULL
        ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='积分记录表';

-- =====================================================
-- 7. 监控审计
-- =====================================================

-- 7.1 visit_logs - 访问日志表
CREATE TABLE visit_logs (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    created_at DATETIME NOT NULL,

    -- 访问者信息
    user_id BIGINT UNSIGNED NULL,

    -- 访问信息
    ip VARCHAR(45) NOT NULL COMMENT 'IP地址(支持IPv6)',
    path VARCHAR(500) NOT NULL COMMENT '访问路径',
    method VARCHAR(10) NOT NULL COMMENT 'HTTP方法',
    user_agent VARCHAR(500) NULL COMMENT 'User-Agent',
    referer VARCHAR(500) NULL COMMENT '来源页面',

    -- 设备信息(可选)
    device_type VARCHAR(20) NULL COMMENT '设备类型',
    os VARCHAR(50) NULL COMMENT '操作系统',
    browser VARCHAR(50) NULL COMMENT '浏览器',

    -- 地理位置信息(可选)
    country VARCHAR(50) NULL COMMENT '国家',
    city VARCHAR(50) NULL COMMENT '城市',

    -- 响应信息
    status_code INT NOT NULL COMMENT 'HTTP状态码',
    response_time BIGINT NOT NULL COMMENT '响应时间(毫秒)',

    -- 会话信息
    session_id VARCHAR(100) NULL COMMENT '会话ID',

    PRIMARY KEY (id),
    KEY idx_visit_logs_user_id (user_id),
    KEY idx_visit_logs_ip (ip),
    KEY idx_visit_logs_path (path),
    KEY idx_visit_logs_method (method),
    KEY idx_visit_logs_created_at (created_at),
    KEY idx_visit_logs_session_id (session_id),

    CONSTRAINT fk_visit_logs_user_id
        FOREIGN KEY (user_id)
        REFERENCES users (id)
        ON DELETE SET NULL
        ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='访问日志表'
PARTITION BY RANGE (MONTH(created_at)) (
    PARTITION p1 VALUES LESS THAN (2),
    PARTITION p2 VALUES LESS THAN (3),
    PARTITION p3 VALUES LESS THAN (4),
    PARTITION p4 VALUES LESS THAN (5),
    PARTITION p5 VALUES LESS THAN (6),
    PARTITION p6 VALUES LESS THAN (7),
    PARTITION p7 VALUES LESS THAN (8),
    PARTITION p8 VALUES LESS THAN (9),
    PARTITION p9 VALUES LESS THAN (10),
    PARTITION p10 VALUES LESS THAN (11),
    PARTITION p11 VALUES LESS THAN (12),
    PARTITION p12 VALUES LESS THAN (13)
);

-- 7.2 ip_blacklists - IP黑名单表
CREATE TABLE ip_blacklists (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,

    -- IP地址
    ip VARCHAR(45) NOT NULL COMMENT 'IP地址(IPv4或IPv6)',

    -- 禁止信息
    reason VARCHAR(255) NOT NULL COMMENT '禁止原因',
    banned_by_id BIGINT UNSIGNED NOT NULL COMMENT '禁止者(管理员)',

    -- 时间信息
    banned_at DATETIME NOT NULL COMMENT '禁止时间',
    expires_at DATETIME NULL COMMENT '过期时间(NULL表示永久)',

    -- 统计信息
    access_count INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '访问次数',
    last_access_at DATETIME NULL COMMENT '最后访问时间',

    PRIMARY KEY (id),
    UNIQUE KEY idx_ip_blacklists_ip (ip),
    KEY idx_ip_blacklists_banned_by_id (banned_by_id),
    KEY idx_ip_blacklists_expires_at (expires_at),

    CONSTRAINT fk_ip_blacklists_banned_by_id
        FOREIGN KEY (banned_by_id)
        REFERENCES users (id)
        ON DELETE RESTRICT
        ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='IP黑名单表';

-- 7.3 admin_logs - 管理员操作日志表
CREATE TABLE admin_logs (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    created_at DATETIME NOT NULL,

    admin_id BIGINT UNSIGNED NOT NULL COMMENT '管理员ID',
    action VARCHAR(50) NOT NULL COMMENT '操作类型',
    target_type VARCHAR(50) NOT NULL COMMENT '目标类型',
    target_id BIGINT UNSIGNED NOT NULL COMMENT '目标ID',

    -- 详细信息
    before_data TEXT NULL COMMENT '操作前数据(JSON格式)',
    after_data TEXT NULL COMMENT '操作后数据(JSON格式)',
    ip VARCHAR(45) NULL COMMENT 'IP地址',
    user_agent VARCHAR(500) NULL COMMENT 'User-Agent',

    PRIMARY KEY (id),
    KEY idx_admin_logs_admin_id (admin_id),
    KEY idx_admin_logs_action (action),
    KEY idx_admin_logs_target_type_id (target_type, target_id),
    KEY idx_admin_logs_created_at (created_at),

    CONSTRAINT fk_admin_logs_admin_id
        FOREIGN KEY (admin_id)
        REFERENCES users (id)
        ON DELETE RESTRICT
        ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理员操作日志表';

-- =====================================================
-- 8. 系统管理
-- =====================================================

-- 8.1 ads - 广告表
CREATE TABLE ads (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    deleted_at DATETIME NULL,

    title VARCHAR(100) NOT NULL COMMENT '广告标题',
    image_url VARCHAR(500) NOT NULL COMMENT '图片链接',
    link_url VARCHAR(500) NOT NULL COMMENT '跳转链接',
    ad_position VARCHAR(50) NOT NULL COMMENT '广告位置',

    -- 显示设置
    is_active BOOLEAN NOT NULL DEFAULT TRUE COMMENT '是否启用',
    sort_order INT NOT NULL DEFAULT 0 COMMENT '排序',
    click_count INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '点击次数',

    -- 时间设置
    start_date DATETIME NULL COMMENT '开始日期',
    end_date DATETIME NULL COMMENT '结束日期',

    PRIMARY KEY (id),
    KEY idx_ads_position_sort (ad_position, sort_order),
    KEY idx_ads_active_date (is_active, start_date, end_date),
    KEY idx_ads_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='广告表';

-- 8.2 permissions - 权限配置表
CREATE TABLE permissions (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,

    key VARCHAR(50) NOT NULL COMMENT '权限键名',
    name VARCHAR(100) NOT NULL COMMENT '权限名称',
    description VARCHAR(255) NULL COMMENT '权限描述',
    is_enabled BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否启用',

    PRIMARY KEY (id),
    UNIQUE KEY idx_permissions_key (key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='权限配置表';

-- 8.3 import_tasks - 导入任务表
CREATE TABLE import_tasks (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,

    task_type VARCHAR(30) NOT NULL COMMENT '任务类型: crawler, excel',
    status VARCHAR(20) NOT NULL COMMENT '状态: pending, running, completed, failed',

    -- 统计信息
    total_count INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '总数',
    success_count INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '成功数',
    fail_count INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '失败数',

    -- 详细信息
    config_data TEXT NULL COMMENT '配置信息(JSON格式)',
    error_log TEXT NULL COMMENT '错误日志',

    -- 执行信息
    started_at DATETIME NULL COMMENT '开始时间',
    completed_at DATETIME NULL COMMENT '完成时间',

    PRIMARY KEY (id),
    KEY idx_import_tasks_task_type (task_type),
    KEY idx_import_tasks_status (status),
    KEY idx_import_tasks_started_at (started_at),
    KEY idx_import_tasks_completed_at (completed_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='导入任务表';

-- =====================================================
-- 9. 初始化数据
-- =====================================================

-- 9.1 插入默认积分规则
INSERT INTO points_rules (created_at, updated_at, rule_key, rule_name, description, points, is_enabled) VALUES
(NOW(), NOW(), 'invite_reward', '邀请奖励', '成功邀请一个用户注册', 50, TRUE),
(NOW(), NOW(), 'resource_download', '资源下载', '下载需要积分的资源', -10, TRUE),
(NOW(), NOW(), 'daily_checkin', '每日签到', '每日登录奖励', 5, TRUE),
(NOW(), NOW(), 'upload_reward', '上传奖励', '审核通过一个资源', 10, TRUE);

-- 9.2 插入默认权限
INSERT INTO permissions (created_at, updated_at, key, name, description, is_enabled) VALUES
(NOW(), NOW(), 'user.upload', '用户上传', '允许用户上传资源', TRUE),
(NOW(), NOW(), 'user.comment', '用户评论', '允许用户评论资源', TRUE),
(NOW(), NOW(), 'admin.review', '资源审核', '允许审核资源', TRUE),
(NOW(), NOW(), 'admin.ban_user', '封禁用户', '允许封禁/解封用户', TRUE),
(NOW(), NOW(), 'admin.ip_ban', 'IP封禁', '允许封禁IP地址', TRUE),
(NOW(), NOW(), 'admin.manage_ads', '广告管理', '允许管理广告', TRUE),
(NOW(), NOW(), 'admin.view_logs', '查看日志', '允许查看系统日志', TRUE),
(NOW(), NOW(), 'admin.import', '导入数据', '允许导入资源数据', TRUE);

-- 9.3 创建默认管理员账户
-- 密码: admin123 (请在生产环境中立即修改)
INSERT INTO users (created_at, updated_at, username, email, password_hash, role, status, can_upload, invite_code, points_balance) VALUES
(NOW(), NOW(), 'admin', 'felixwang.biz@gmail.com', '$2a$10$N.zmdr9k7uOCQb376NoUnuTJ8iK.vpDh6Y5wR.MZCP.mC.bC5dv', 'admin', 'active', TRUE, 'ADMIN0000000000000000000000000000000000', 0);

-- 9.4 创建默认分类
INSERT INTO categories (created_at, updated_at, name, description, icon, color, sort_order) VALUES
(NOW(), NOW(), '软件工具', '各类实用软件和工具', 'software.png', '#3498db', 1),
(NOW(), NOW(), '电子资料', '电子书、文档、教程等学习资料', 'document.png', '#2ecc71', 2),
(NOW(), NOW(), '多媒体', '电影、音乐、图片等娱乐资源', 'multimedia.png', '#e74c3c', 3),
(NOW(), NOW(), '游戏', '游戏资源、游戏补丁等', 'game.png', '#9b59b6', 4),
(NOW(), NOW(), '其他', '其他类型的资源', 'other.png', '#95a5a6', 5);

-- =====================================================
-- 10. 创建视图
-- =====================================================

-- 10.1 资源统计视图
CREATE VIEW v_resource_stats AS
SELECT
    r.id,
    r.title,
    r.status,
    r.downloads_count,
    r.views_count,
    c.name AS category_name,
    u.username AS uploader_name,
    r.created_at,
    r.reviewed_at
FROM resources r
LEFT JOIN categories c ON r.category_id = c.id
LEFT JOIN users u ON r.uploaded_by_id = u.id;

-- 10.2 用户积分统计视图
CREATE VIEW v_user_points AS
SELECT
    u.id,
    u.username,
    u.email,
    u.points_balance,
    COALESCE(SUM(CASE WHEN pr.type = 'income' THEN pr.points ELSE 0 END), 0) AS total_income,
    COALESCE(SUM(CASE WHEN pr.type = 'expense' THEN ABS(pr.points) ELSE 0 END), 0) AS total_expense,
    COUNT(DISTINCT pr.id) AS transaction_count
FROM users u
LEFT JOIN point_records pr ON u.id = pr.user_id
WHERE u.deleted_at IS NULL
GROUP BY u.id, u.username, u.email, u.points_balance;

-- =====================================================
-- 11. 创建存储过程
-- =====================================================

-- 11.1 获取用户积分统计
DELIMITER //
CREATE PROCEDURE GetUserPointsStats(IN user_id BIGINT)
BEGIN
    DECLARE total_income INT DEFAULT 0;
    DECLARE total_expense INT DEFAULT 0;
    DECLARE transaction_count INT DEFAULT 0;

    SELECT
        COALESCE(SUM(CASE WHEN type = 'income' THEN points ELSE 0 END), 0) INTO total_income,
        COALESCE(SUM(CASE WHEN type = 'expense' THEN ABS(points) ELSE 0 END), 0) INTO total_expense,
        COUNT(*) INTO transaction_count
    FROM point_records
    WHERE user_id = user_id;

    SELECT
        user_id,
        total_income,
        total_expense,
        transaction_count,
        (total_income - total_expense) AS net_points;
END//
DELIMITER ;

-- 11.2 清理过期邀请
DELIMITER //
CREATE PROCEDURE CleanupExpiredInvitations()
BEGIN
    UPDATE invitations
    SET status = 'expired'
    WHERE status = 'pending'
      AND expires_at < NOW();
END//
DELIMITER ;

-- =====================================================
-- 12. 触发器
-- =====================================================

-- 12.1 更新资源下载计数
DELIMITER //
CREATE TRIGGER tr_resource_download_count
AFTER UPDATE ON point_records
FOR EACH ROW
BEGIN
    IF NEW.source = 'resource_download' AND NEW.type = 'expense' AND NEW.resource_id IS NOT NULL THEN
        UPDATE resources
        SET downloads_count = downloads_count + 1
        WHERE id = NEW.resource_id;
    END IF;
END//
DELIMITER ;

-- 12.2 更新分类资源计数
DELIMITER //
CREATE TRIGGER tr_category_resource_count_insert
AFTER INSERT ON resources
FOR EACH ROW
BEGIN
    UPDATE categories
    SET resources_count = resources_count + 1
    WHERE id = NEW.category_id;
END//
DELIMITER ;

DELIMITER //
CREATE TRIGGER tr_category_resource_count_delete
AFTER DELETE ON resources
FOR EACH ROW
BEGIN
    UPDATE categories
    SET resources_count = resources_count - 1
    WHERE id = OLD.category_id AND resources_count > 0;
END//
DELIMITER ;

-- =====================================================
-- 13. 创建数据库用户和授权
-- =====================================================

-- 13.1 创建应用用户
CREATE USER IF NOT EXISTS 'resource_share'@'%' IDENTIFIED BY 'StrongPassword123!';
GRANT SELECT, INSERT, UPDATE, DELETE ON resource_share_site.* TO 'resource_share'@'%';
GRANT INDEX, CREATE, ALTER ON resource_share_site.* TO 'resource_share'@'%';

-- 13.2 创建只读用户
CREATE USER IF NOT EXISTS 'resource_share_readonly'@'%' IDENTIFIED BY 'ReadOnlyPassword123!';
GRANT SELECT ON resource_share_site.* TO 'resource_share_readonly'@'%';

-- 13.3 创建管理员用户（仅DBA使用）
CREATE USER IF NOT EXISTS 'resource_share_admin'@'%' IDENTIFIED BY 'AdminPassword123!';
GRANT ALL PRIVILEGES ON resource_share_site.* TO 'resource_share_admin'@'%';

-- =====================================================
-- 14. 验证脚本
-- =====================================================

-- 验证表创建
SELECT
    table_name,
    engine,
    table_collation,
    table_comment
FROM information_schema.tables
WHERE table_schema = 'resource_share_site'
ORDER BY table_name;

-- 验证外键约束
SELECT
    table_name,
    column_name,
    constraint_name,
    referenced_table_name,
    referenced_column_name
FROM information_schema.key_column_usage
WHERE table_schema = 'resource_share_site'
    AND referenced_table_name IS NOT NULL
ORDER BY table_name;

-- 验证索引
SELECT
    table_name,
    index_name,
    column_name,
    non_unique,
    index_type
FROM information_schema.statistics
WHERE table_schema = 'resource_share_site'
ORDER BY table_name, index_name, seq_in_index;

-- =====================================================
-- 完成
-- =====================================================

SELECT 'Database schema created successfully!' AS status;
