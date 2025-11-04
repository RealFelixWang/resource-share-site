#!/bin/bash

# Resource Share Site 启动脚本
# Author: Felix Wang
# Email: felixwang.biz@gmail.com

set -e

echo "======================================"
echo "  Resource Share Site 启动脚本"
echo "======================================"
echo ""

# 检查是否安装了 Docker
if ! command -v docker &> /dev/null; then
    echo "❌ Docker 未安装，请先安装 Docker"
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose 未安装，请先安装 Docker Compose"
    exit 1
fi

# 检查配置文件
if [ ! -f "config/config.yaml" ]; then
    echo "⚠️  配置文件不存在，使用默认配置"
fi

# 启动服务
echo "🚀 启动服务..."
docker-compose up -d

echo ""
echo "⏳ 等待服务启动..."
sleep 10

# 检查服务状态
echo ""
echo "📊 服务状态:"
docker-compose ps

echo ""
echo "✅ 启动完成！"
echo ""
echo "🌐 访问地址:"
echo "  - 主站: http://localhost:8080"
echo "  - 数据库管理: http://localhost:8081"
echo ""
echo "📝 管理员账户:"
echo "  - 用户名: admin"
echo "  - 密码: admin123"
echo ""
echo "🛑 停止服务: docker-compose down"
echo "📋 查看日志: docker-compose logs -f app"
echo ""
