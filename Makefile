# Resource Share Site Makefile
# Author: Felix Wang
# Email: felixwang.biz@gmail.com

.PHONY: build test clean docker-up docker-down install dev help

# 默认目标
.DEFAULT_GOAL := help

# 变量定义
APP_NAME := resource-share-site
BINARY_NAME := main
BUILD_DIR := build
CMD_DIR := cmd

# Go 相关变量
GO_VERSION := 1.25.3
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod
GOFMT := go fmt
GOLINT := golangci-lint
GOCOVER := $(GOCMD) tool cover

# Docker 相关变量
DOCKER_IMAGE := $(APP_NAME)
DOCKER_TAG := latest
DOCKER_CONTAINER := $(APP_NAME)-container

## 格式化代码
fmt:
	@echo "=== 格式化代码 ==="
	$(GOFMT) ./...
	@echo "✅ 代码格式化完成"

## 代码检查
lint:
	@echo "=== 代码检查 ==="
	$(GOLINT) run ./...
	@echo "✅ 代码检查完成"

## 运行测试
test:
	@echo "=== 运行测试 ==="
	$(GOTEST) -v ./...
	@echo "✅ 测试完成"

## 生成测试覆盖率报告
test-coverage:
	@echo "=== 生成测试覆盖率报告 ==="
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCOVER) html coverage.out -o coverage.html
	@echo "✅ 测试覆盖率报告生成完成: coverage.html"

## 运行测试程序
test-all:
	@echo "=== 运行所有测试程序 ==="
	cd $(CMD_DIR)/testauth && $(GOBUILD) -o ../../$(BUILD_DIR)/testauth main.go
	cd $(CMD_DIR)/testuserstatus && $(GOBUILD) -o ../../$(BUILD_DIR)/testuserstatus main.go
	cd $(CMD_DIR)/testsessions && $(GOBUILD) -o ../../$(BUILD_DIR)/testsessions main.go
	cd $(CMD_DIR)/testmiddleware && $(GOBUILD) -o ../../$(BUILD_DIR)/testmiddleware main.go
	cd $(CMD_DIR)/testinvitation && $(GOBUILD) -o ../../$(BUILD_DIR)/testinvitation main.go
	cd $(CMD_DIR)/testcategory && $(GOBUILD) -o ../../$(BUILD_DIR)/testcategory main.go
	cd $(CMD_DIR)/testresource && $(GOBUILD) -o ../../$(BUILD_DIR)/testresource main.go
	cd $(CMD_DIR)/testpoints && $(GOBUILD) -o ../../$(BUILD_DIR)/testpoints main.go
	cd $(CMD_DIR)/testseo && $(GOBUILD) -o ../../$(BUILD_DIR)/testseo main.go
	@echo "✅ 测试程序构建完成"

## 构建应用
build:
	@echo "=== 构建应用 ==="
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)/server/main.go
	@echo "✅ 应用构建完成: $(BUILD_DIR)/$(BINARY_NAME)"

## 安装依赖
install:
	@echo "=== 安装依赖 ==="
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "✅ 依赖安装完成"

## 开发模式运行
dev:
	@echo "=== 开发模式运行 ==="
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)/server/main.go
	$(BUILD_DIR)/$(BINARY_NAME)
	@echo "✅ 开发服务启动完成"

## 生产模式运行
run:
	@echo "=== 生产模式运行 ==="
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)/server/main.go
	export GIN_MODE=release && $(BUILD_DIR)/$(BINARY_NAME)
	@echo "✅ 生产服务启动完成"

## Docker 构建
docker-build:
	@echo "=== Docker 构建 ==="
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "✅ Docker 镜像构建完成"

## Docker 运行
docker-up:
	@echo "=== 启动 Docker 服务 ==="
	docker-compose up -d
	@echo "✅ Docker 服务启动完成"

## Docker 停止
docker-down:
	@echo "=== 停止 Docker 服务 ==="
	docker-compose down
	@echo "✅ Docker 服务已停止"

## Docker 重启
docker-restart:
	@echo "=== 重启 Docker 服务 ==="
	docker-compose restart
	@echo "✅ Docker 服务已重启"

## 查看 Docker 日志
docker-logs:
	docker-compose logs -f

## 清理构建文件
clean:
	@echo "=== 清理构建文件 ==="
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	@echo "✅ 清理完成"

## 数据库迁移
migrate:
	@echo "=== 数据库迁移 ==="
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)/server/main.go
	$(BUILD_DIR)/$(BINARY_NAME) migrate
	@echo "✅ 数据库迁移完成"

## 重置数据库
migrate-reset:
	@echo "=== 重置数据库 ==="
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)/server/main.go
	$(BUILD_DIR)/$(BINARY_NAME) migrate-reset
	@echo "✅ 数据库已重置"

## 查看服务状态
status:
	@echo "=== 服务状态 ==="
	docker-compose ps
	@echo ""

## 生成 API 文档
gen-docs:
	@echo "=== 生成 API 文档 ==="
	swagger init spec
	@echo "✅ API 文档生成完成"

## 代码安全扫描
security:
	@echo "=== 代码安全扫描 ==="
	gosec ./...
	@echo "✅ 安全扫描完成"

## 性能分析
profile:
	@echo "=== 性能分析 ==="
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)/server/main.go
	go tool pprof $(BUILD_DIR)/$(BINARY_NAME)
	@echo "✅ 性能分析完成"

## 检查依赖漏洞
deps-check:
	@echo "=== 检查依赖漏洞 ==="
	go list -json -m all | nancy sleuth
	@echo "✅ 依赖漏洞检查完成"

## 更新依赖
deps-update:
	@echo "=== 更新依赖 ==="
	$(GOMOD) get -u
	$(GOMOD) tidy
	@echo "✅ 依赖更新完成"

## 生成 Graph
graph:
	@echo "=== 生成依赖图 ==="
	$(GOCMD) mod graph > $(BUILD_DIR)/deps.txt
	@echo "✅ 依赖图生成完成: $(BUILD_DIR)/deps.txt"

## 显示帮助
help:
	@echo "Resource Share Site Makefile"
	@echo ""
	@echo "可用命令:"
	@echo ""
	@echo "  代码质量:"
	@echo "    fmt        - 格式化代码"
	@echo "    lint       - 代码检查"
	@echo "    test       - 运行测试"
	@echo "    test-coverage - 生成测试覆盖率报告"
	@echo "    security   - 代码安全扫描"
	@echo ""
	@echo "  构建和运行:"
	@echo "    build      - 构建应用"
	@echo "    install    - 安装依赖"
	@echo "    dev        - 开发模式运行"
	@echo "    run        - 生产模式运行"
	@echo ""
	@echo "  Docker:"
	@echo "    docker-build - 构建 Docker 镜像"
	@echo "    docker-up  - 启动 Docker 服务"
	@echo "    docker-down - 停止 Docker 服务"
	@echo "    docker-restart - 重启 Docker 服务"
	@echo "    docker-logs - 查看 Docker 日志"
	@echo ""
	@echo "  数据库:"
	@echo "    migrate    - 数据库迁移"
	@echo "    migrate-reset - 重置数据库"
	@echo ""
	@echo "  维护:"
	@echo "    clean      - 清理构建文件"
	@echo "    deps-check - 检查依赖漏洞"
	@echo "    deps-update - 更新依赖"
	@echo "    status     - 查看服务状态"
	@echo "    profile    - 性能分析"
	@echo "    graph      - 生成依赖图"
	@echo ""
	@echo "  帮助:"
	@echo "    help       - 显示此帮助信息"
	@echo ""
	@echo "示例:"
	@echo "  make test-coverage   # 生成测试覆盖率报告"
	@echo "  make docker-up      # 启动 Docker 服务"
	@echo "  make build && make dev  # 构建并运行开发模式"
	@echo ""
