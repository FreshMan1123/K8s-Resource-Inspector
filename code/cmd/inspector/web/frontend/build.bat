@echo off
chcp 65001 >nul

echo 🚀 开始构建 K8s Resource Inspector 前端...

REM 检查 Node.js 和 npm
where node >nul 2>nul
if %errorlevel% neq 0 (
    echo ❌ 错误: 未找到 Node.js，请先安装 Node.js 16+
    exit /b 1
)

where npm >nul 2>nul
if %errorlevel% neq 0 (
    echo ❌ 错误: 未找到 npm，请先安装 npm
    exit /b 1
)

REM 进入前端目录
cd /d "%~dp0"

echo 📦 安装依赖...
npm install

if %errorlevel% neq 0 (
    echo ❌ 依赖安装失败
    exit /b 1
)

echo 🔨 构建前端资源...
npm run build

if %errorlevel% neq 0 (
    echo ❌ 前端构建失败
    exit /b 1
)

echo ✅ 前端构建完成！
echo 📁 构建产物位置: ..\cmd\inspector\web\frontend\dist
echo.
echo 🎯 下一步：
echo    1. 编译 Go 程序: cd ..\..\code ^&^& go build -o inspector.exe .\cmd\inspector\
echo    2. 启动 Web 服务: .\inspector.exe web
echo    3. 访问: http://localhost:8080

pause
