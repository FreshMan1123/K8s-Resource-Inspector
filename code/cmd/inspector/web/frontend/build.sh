#!/bin/bash

# K8s Resource Inspector 前端构建脚本

echo "🚀 开始构建 K8s Resource Inspector 前端..."

# 检查 Node.js 和 npm
if ! command -v node &> /dev/null; then
    echo "❌ 错误: 未找到 Node.js，请先安装 Node.js 16+"
    exit 1
fi

if ! command -v npm &> /dev/null; then
    echo "❌ 错误: 未找到 npm，请先安装 npm"
    exit 1
fi

# 进入前端目录
cd "$(dirname "$0")"

echo "📦 安装依赖..."
npm install

if [ $? -ne 0 ]; then
    echo "❌ 依赖安装失败"
    exit 1
fi

echo "🔨 构建前端资源..."
npm run build

if [ $? -ne 0 ]; then
    echo "❌ 前端构建失败"
    exit 1
fi

echo "✅ 前端构建完成！"
echo "📁 构建产物位置: ../cmd/inspector/web/frontend/dist"
echo ""
echo "🎯 下一步："
echo "   1. 编译 Go 程序: cd ../../ && go build -o inspector.exe ./cmd/inspector/"
echo "   2. 启动 Web 服务: ./inspector.exe web"
echo "   3. 访问: http://localhost:8080"
