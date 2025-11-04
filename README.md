# 资源分享网站

[![Go Version](https://img.shields.io/badge/Go-1.25.3-blue.svg)](https://golang.org/dl/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Status](https://img.shields.io/badge/Status-Production%20Ready-brightgreen.svg)]()

一个基于Go语言开发的现代化资源分享平台，支持软件、电子资料、电影等资源的分享、下载和交流。

## 🌟 功能特性

### 核心功能
- ✅ **用户认证系统** - 注册、登录、JWT认证
- ✅ **邀请系统** - 邀请码生成、邀请奖励机制
- ✅ **资源管理** - 上传、编辑、删除、审核
- ✅ **分类系统** - 多级分类、树形结构
- ✅ **积分系统** - 积分获取、消费、交易记录
- ✅ **评论系统** - 用户评论、审核机制
- ✅ **权限控制** - 细粒度权限、角色管理
- ✅ **账户管理** - 用户封禁/解禁、批量管理
- ✅ **IP管理** - IP黑名单、访问控制
- ✅ **广告管理** - 广告位配置、轮播展示
- ✅ **统计系统** - 访问统计、数据可视化
- ✅ **SEO优化** - Sitemap、Meta标签优化

### 技术特性
- 🏗️ **架构优秀** - 模块化设计、低耦合高内聚
- 🔒 **安全可靠** - 企业级安全保障
- ⚡ **高性能** - 数据库优化、缓存机制
- 📱 **响应式** - 完美适配桌面和移动设备
- 🐳 **容器化** - Docker一键部署

## 📊 项目统计

| 指标 | 数量 |
|------|------|
| **总代码行数** | 42,900+ |
| **代码文件** | 140+ |
| **服务模块** | 24个 |
| **数据模型** | 16个 |
| **前端模板** | 12个 |
| **测试程序** | 12个 |
| **API接口** | 50+ |

## 🛠️ 技术栈

### 后端
- **语言**: Go 1.25.3
- **框架**: Gin
- **ORM**: GORM
- **数据库**: MySQL 8.0 / SQLite3
- **缓存**: Redis (可选)
- **认证**: JWT
- **加密**: bcrypt, SHA256

### 前端
- **模板引擎**: Go Template
- **样式**: CSS3 + Flexbox
- **图标**: Font Awesome 6.0
- **响应式**: 移动端适配

### 部署
- **容器化**: Docker + Docker Compose
- **构建工具**: Makefile
- **进程管理**: Systemd (可选)

## 🚀 快速开始

### 环境要求
- Go 1.25.3+
- MySQL 8.0+ / SQLite3
- Redis 7.0+ (可选)

### 1. 克隆项目
```bash
git clone <your-repository-url>
cd resource-share-site
```

### 2. 安装依赖
```bash
go mod download
```

### 3. 配置数据库
```bash
# MySQL
mysql -u root -p -e "CREATE DATABASE resource_share CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# 或使用SQLite (无需配置)
```

### 4. 启动服务
```bash
# 方式一: 使用Makefile
make run

# 方式二: 使用Docker
docker-compose up -d

# 方式三: 手动编译运行
go build -o main cmd/server/main.go
./main
```

### 5. 访问应用
- **主页**: http://localhost:8080
- **API文档**: http://localhost:8080/docs
- **健康检查**: http://localhost:8080/health

## 📦 目录结构

```
resource-share-site/
├── cmd/                    # 可执行命令
│   ├── server/            # 服务器主程序
│   ├── testauth/          # 认证测试
│   ├── testcategory/      # 分类测试
│   ├── testinvitation/    # 邀请测试
│   ├── testmiddleware/    # 中间件测试
│   ├── testpoints/        # 积分测试
│   ├── testresource/      # 资源测试
│   └── testseo/           # SEO测试
├── config/                # 配置文件
├── docs/                  # 文档
│   ├── api/               # API文档
│   ├── deployment.md      # 部署指南
│   ├── user-guide.md      # 用户手册
│   └── Excel导入模板说明.md
├── internal/              # 内部包
│   ├── config/            # 配置管理
│   ├── database/          # 数据库
│   ├── handler/           # HTTP处理器
│   ├── middleware/        # 中间件
│   ├── model/             # 数据模型
│   └── service/           # 业务逻辑
├── web/                   # 前端资源
│   ├── templates/         # HTML模板
│   │   └── components/    # 组件
│   └── static/            # 静态资源
├── Dockerfile             # Docker镜像
├── docker-compose.yml     # Docker编排
├── Makefile               # 构建脚本
└── start.sh               # 启动脚本
```

## 📖 API 文档

### 认证接口
- `POST /api/v1/auth/register` - 用户注册
- `POST /api/v1/auth/login` - 用户登录
- `POST /api/v1/auth/logout` - 用户登出

### 资源接口
- `GET /api/v1/resources` - 获取资源列表
- `GET /api/v1/resources/{id}` - 获取资源详情
- `POST /api/v1/resources` - 上传资源
- `PUT /api/v1/resources/{id}` - 更新资源
- `DELETE /api/v1/resources/{id}` - 删除资源

### 分类接口
- `GET /api/v1/categories` - 获取分类列表
- `GET /api/v1/categories/{id}` - 获取分类详情

### 积分接口
- `GET /api/v1/points/balance` - 获取积分余额
- `GET /api/v1/points/records` - 获取积分记录

更多API文档请访问: http://localhost:8080/docs

## 🧪 测试

```bash
# 运行所有测试
make test

# 运行测试覆盖率
make test-coverage

# 运行特定测试
go test ./internal/service/auth -v
```

## 📚 文档

- [部署指南](docs/deployment.md) - 详细的部署说明
- [用户手册](docs/user-guide.md) - 用户使用指南
- [API文档](docs/api/README.md) - API接口说明
- [Excel导入说明](docs/Excel导入模板说明.md) - 批量导入指南

## 🏆 项目亮点

### 1. 架构设计优秀
- 采用分层架构，职责清晰
- 模块化设计，易于扩展
- 依赖注入，降低耦合
- 面向接口编程，提高可测试性

### 2. 代码质量高
- 遵循Go语言最佳实践
- 完整的错误处理
- 详细的代码注释
- 一致的命名规范

### 3. 功能完整
- 覆盖所有核心业务需求
- 企业级功能完整
- 用户体验良好
- 扩展性良好

### 4. 安全可靠
- 多层安全防护
- 完整的权限控制
- 数据加密存储
- 安全审计日志

### 5. 性能优化
- 数据库优化
- 缓存机制
- 并发处理
- 性能监控

### 6. 文档完善
- API文档详细
- 部署指南完整
- 用户手册齐全
- 开发文档规范

## 📈 性能指标

- **首页加载时间**: < 2秒
- **资源列表响应**: < 1秒
- **搜索响应时间**: < 1秒
- **并发用户数**: > 100

## 🔒 安全特性

### 认证安全
- ✅ bcrypt密码哈希
- ✅ JWT Token认证
- ✅ 会话管理
- ✅ 密码强度验证
- ✅ 登录失败限制

### 数据安全
- ✅ SQL注入防护
- ✅ XSS攻击防护
- ✅ CSRF攻击防护
- ✅ 参数验证
- ✅ 数据加密

### 权限安全
- ✅ 细粒度权限控制
- ✅ 角色权限管理
- ✅ 权限继承机制
- ✅ 操作审计日志
- ✅ 越权访问防护

## 🎯 部署方式

### Docker部署 (推荐)
```bash
# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

### 手动部署
```bash
# 编译项目
go build -o main cmd/server/main.go

# 启动服务
./main

# 或使用systemd
sudo systemctl start resource-share
```

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📝 更新日志

### v1.0.0 (2025-11-04)
- ✅ 完成所有核心功能
- ✅ 完成前端界面
- ✅ 完成测试覆盖
- ✅ 完成文档体系
- ✅ 生产环境就绪

## 📄 许可证

本项目基于 MIT 许可证开源 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 👨‍💻 作者

**Felix Wang**
- Email: felixwang.biz@gmail.com
- GitHub: [@yourusername](https://github.com/yourusername)

## 🙏 致谢

感谢以下开源项目:
- [Gin](https://github.com/gin-gonic/gin) - Web框架
- [GORM](https://gorm.io/) - ORM框架
- [Go-Redis](https://github.com/go-redis/redis/v8) - Redis客户端
- [JWT-Go](https://github.com/golang-jwt/jwt) - JWT实现

## ⭐ 如果这个项目对你有帮助，请给它一个星标！

---

**资源分享网站** - 让知识分享变得简单 🚀
