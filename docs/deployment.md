# 资源分享网站 - 部署指南

## 目录
- [系统要求](#系统要求)
- [Docker 部署（推荐）](#docker-部署推荐)
- [手动部署](#手动部署)
- [配置说明](#配置说明)
- [数据库迁移](#数据库迁移)
- [常见问题](#常见问题)

## 系统要求

### 最低要求
- **操作系统**: Linux/Windows/macOS
- **CPU**: 1核心
- **内存**: 2GB RAM
- **存储**: 10GB 可用空间
- **网络**: 1Mbps 带宽

### 推荐配置
- **操作系统**: Ubuntu 20.04 LTS / CentOS 8
- **CPU**: 2核心或更多
- **内存**: 4GB RAM 或更多
- **存储**: 50GB SSD
- **网络**: 10Mbps 带宽

### 软件依赖
- **Go**: 1.25.3 或更高版本
- **MySQL**: 8.0 或更高版本（可选：SQLite3）
- **Redis**: 7.0 或更高版本（可选）
- **Docker**: 20.10 或更高版本
- **Docker Compose**: 2.0 或更高版本

## Docker 部署（推荐）

### 1. 克隆项目
```bash
git clone <your-repository-url>
cd resource-share-site
```

### 2. 准备配置文件
```bash
# 复制配置文件模板
cp config/config.yaml.example config/config.yaml

# 编辑配置文件
vim config/config.yaml
```

### 3. 启动服务
```bash
# 使用启动脚本（推荐）
chmod +x start.sh
./start.sh

# 或手动启动
docker-compose up -d
```

### 4. 检查服务状态
```bash
docker-compose ps
```

### 5. 查看日志
```bash
# 查看所有服务日志
docker-compose logs

# 查看特定服务日志
docker-compose logs -f app
docker-compose logs -f mysql
docker-compose logs -f redis
```

### 6. 停止服务
```bash
docker-compose down

# 停止并删除数据卷
docker-compose down -v
```

## 手动部署

### 1. 安装 Go 环境

#### Ubuntu/Debian
```bash
sudo apt update
sudo apt install golang-go
```

#### CentOS/RHEL
```bash
sudo yum install golang
```

#### macOS
```bash
brew install go
```

### 2. 安装 MySQL

#### Ubuntu/Debian
```bash
sudo apt install mysql-server
sudo mysql_secure_installation
```

#### CentOS/RHEL
```bash
sudo yum install mysql-server
sudo systemctl start mysqld
sudo systemctl enable mysqld
```

### 3. 安装 Redis（可选）
```bash
# Ubuntu/Debian
sudo apt install redis-server

# CentOS/RHEL
sudo yum install redis
```

### 4. 编译项目
```bash
# 进入项目目录
cd resource-share-site

# 下载依赖
go mod download

# 编译
go build -o main cmd/server/main.go
```

### 5. 配置数据库

#### 创建数据库
```sql
CREATE DATABASE resource_share CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'app'@'localhost' IDENTIFIED BY 'your_password';
GRANT ALL PRIVILEGES ON resource_share.* TO 'app'@'localhost';
FLUSH PRIVILEGES;
```

#### 执行迁移
```bash
# 如果有迁移工具
./main migrate

# 或手动导入 SQL 文件
mysql -u app -p resource_share < config/mysql/schema.sql
```

### 6. 创建系统服务

#### 创建服务文件
```bash
sudo vim /etc/systemd/system/resource-share.service
```

#### 服务配置内容
```ini
[Unit]
Description=Resource Share Site
After=network.target mysql.service

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/resource-share-site
ExecStart=/opt/resource-share-site/main
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

#### 启动服务
```bash
sudo systemctl daemon-reload
sudo systemctl enable resource-share
sudo systemctl start resource-share
```

### 7. 配置 Nginx（可选但推荐）

#### 安装 Nginx
```bash
# Ubuntu/Debian
sudo apt install nginx

# CentOS/RHEL
sudo yum install nginx
```

#### 创建站点配置
```bash
sudo vim /etc/nginx/sites-available/resource-share
```

#### Nginx 配置示例
```nginx
server {
    listen 80;
    server_name your-domain.com;

    client_max_body_size 100M;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /static/ {
        alias /opt/resource-share-site/web/static/;
        expires 30d;
    }
}
```

#### 启用站点
```bash
sudo ln -s /etc/nginx/sites-available/resource-share /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl restart nginx
```

## 配置说明

### 主配置文件 (config/config.yaml)

```yaml
# 服务配置
server:
  host: "0.0.0.0"
  port: 8080
  mode: "debug"  # debug, release, test

# 数据库配置
database:
  type: "mysql"  # mysql, sqlite
  host: "localhost"
  port: 3306
  user: "root"
  password: "secret"
  name: "resource_share"
  charset: "utf8mb4"

# Redis 配置（可选）
redis:
  enabled: false
  host: "localhost"
  port: 6379
  password: ""
  db: 0

# JWT 配置
jwt:
  secret: "your-secret-key-here"
  expires_hours: 72

# 密码加密
password:
  bcrypt_cost: 12

# 日志配置
log:
  level: "info"
  output: "stdout"  # stdout, file
  file: "logs/app.log"
```

### 环境变量配置

| 变量名 | 描述 | 默认值 |
|--------|------|--------|
| SERVER_HOST | 服务器监听地址 | 0.0.0.0 |
| SERVER_PORT | 服务器端口 | 8080 |
| DB_TYPE | 数据库类型 | mysql |
| DB_HOST | 数据库主机 | localhost |
| DB_PORT | 数据库端口 | 3306 |
| DB_USER | 数据库用户 | root |
| DB_PASSWORD | 数据库密码 | secret |
| DB_NAME | 数据库名称 | resource_share |
| JWT_SECRET | JWT 密钥 | - |
| REDIS_ENABLED | 是否启用Redis | false |

## 数据库迁移

### 自动迁移
应用启动时会自动执行数据库迁移。

### 手动迁移
```bash
# 查看迁移状态
./main migrate status

# 执行迁移
./main migrate up

# 回滚迁移
./main migrate down
```

### 数据备份

#### MySQL 备份
```bash
mysqldump -u root -p resource_share > backup_$(date +%Y%m%d_%H%M%S).sql
```

#### MySQL 恢复
```bash
mysql -u root -p resource_share < backup_20251104_120000.sql
```

## 常见问题

### Q1: 端口被占用
```bash
# 查看端口占用
sudo netstat -tlnp | grep 8080

# 杀死占用进程
sudo kill -9 <PID>
```

### Q2: 数据库连接失败
1. 检查数据库服务状态
2. 验证连接参数
3. 检查防火墙设置
4. 查看应用日志

### Q3: 静态资源加载失败
1. 检查 web 目录权限
2. 验证文件路径
3. 查看 Nginx 配置

### Q4: 权限不足
```bash
# 设置正确的文件权限
sudo chown -R www-data:www-data /opt/resource-share-site
sudo chmod -R 755 /opt/resource-share-site
```

### Q5: 内存不足
1. 调整数据库连接池大小
2. 启用 Redis 缓存
3. 优化查询语句

### Q6: 高并发问题
1. 启用 Redis 缓存
2. 调整数据库连接池
3. 使用负载均衡
4. 配置 CDN

## 性能优化

### 1. 数据库优化
- 启用查询缓存
- 优化索引
- 调整连接池参数
- 使用读写分离

### 2. 缓存策略
- Redis 缓存热点数据
- 静态资源使用 CDN
- 页面级缓存

### 3. 监控
- 应用性能监控 (APM)
- 数据库性能监控
- 服务器资源监控

## 安全建议

### 1. 基础安全
- 修改默认密码
- 禁用 root 远程登录
- 配置防火墙
- 启用 SSL/TLS

### 2. 应用安全
- 定期更新依赖
- 启用安全头
- 配置限流
- 审计日志

### 3. 数据安全
- 定期备份
- 加密敏感数据
- 访问控制
- 审计跟踪

## 升级指南

### 1. 备份数据
```bash
# 备份数据库
mysqldump -u root -p resource_share > backup.sql

# 备份应用文件
tar -czf app_backup.tar.gz /opt/resource-share-site
```

### 2. 更新代码
```bash
git pull origin master
go mod download
go build -o main cmd/server/main.go
```

### 3. 执行迁移
```bash
./main migrate up
```

### 4. 重启服务
```bash
sudo systemctl restart resource-share
```

---

**作者**: Felix Wang  
**邮箱**: felixwang.biz@gmail.com  
**更新时间**: 2025-11-04
