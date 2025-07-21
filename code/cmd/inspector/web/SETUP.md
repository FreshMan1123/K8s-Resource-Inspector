# K8s Resource Inspector Web界面 - 完整设置指南

## 🎯 正确的启动方式

**重要说明**: Web界面通过 `./inspector.exe web` 启动，前端资源已嵌入到Go二进制文件中。

## 📋 环境要求

- **Node.js 16+** (用于构建前端)
- **npm** (包管理器)
- **Go 1.17+** (用于编译后端)

## 🚀 完整设置流程

### 步骤1: 构建前端资源

```bash
# 进入前端目录
cd code/web/frontend

# Windows 用户
build.bat

# Linux/macOS 用户
chmod +x build.sh
./build.sh
```

这将会：
- 安装前端依赖
- 构建Vue应用
- 输出到 `code/cmd/inspector/web/frontend/dist`

### 步骤2: 编译Go程序

```bash
# 回到项目根目录
cd code

# 编译二进制文件
go build -o inspector.exe ./cmd/inspector/
```

### 步骤3: 启动Web服务

```bash
# 启动Web服务 (默认端口8080)
./inspector.exe web

# 或指定端口
./inspector.exe web --port 9090
```

### 步骤4: 访问Web界面

打开浏览器访问: http://localhost:8080

## 🎨 界面功能

启动后您将看到完整的Web管理界面，包括：

### 🏠 仪表板
- 系统概览统计
- 快速巡检按钮
- 最近巡检结果

### 📏 规则管理
- 浏览所有规则分类 (Pod、Node、Deployment、Service、Security)
- 查看规则详情
- 编辑规则配置
- 启用/禁用规则

### 🔍 巡检执行
- **快速巡检**: 预设模板一键执行
- **自定义巡检**: 灵活配置参数
- 实时执行状态监控
- 结果摘要展示

### 📊 历史记录
- 查看所有巡检历史
- 详细结果分析
- 导出报告功能
- 历史对比分析

### ⚙️ 系统设置
- 个性化配置
- 系统信息查看
- 配置导入导出

## 🔧 开发模式 (可选)

如果需要开发前端界面，可以使用开发模式：

```bash
# 启动后端API服务
./inspector.exe web

# 新开终端，启动前端开发服务器
cd code/web/frontend
npm run dev
```

- 后端API: http://localhost:8080
- 前端开发: http://localhost:3000

## 📁 项目结构

```
code/
├── cmd/inspector/
│   ├── web/
│   │   ├── server.go              # Web服务器
│   │   ├── handlers/              # API处理器
│   │   └── frontend/dist/         # 前端构建产物 (embed)
│   ├── web.go                     # Web命令
│   └── main.go                    # 程序入口
├── web/frontend/                  # Vue前端源码
│   ├── src/                       # 源代码
│   ├── package.json               # 依赖配置
│   ├── vite.config.ts             # 构建配置
│   ├── build.sh                   # Linux/macOS构建脚本
│   └── build.bat                  # Windows构建脚本
└── inspector.exe                  # 编译后的二进制文件
```

## 🔍 API接口

Web界面通过以下API与后端通信：

### 规则管理
- `GET /api/rules` - 获取规则分类
- `GET /api/rules/:category` - 获取分类规则
- `PUT /api/rules/:category/:id` - 更新规则

### 巡检执行
- `POST /api/inspection/execute` - 执行巡检
- `GET /api/inspection/status/:id` - 获取执行状态
- `GET /api/inspection/result/:id` - 获取执行结果

## 🐛 故障排除

### 1. 前端构建失败
```bash
# 清理依赖重新安装
cd code/web/frontend
rm -rf node_modules package-lock.json
npm install
npm run build
```

### 2. Go编译失败
```bash
# 检查Go版本
go version

# 清理模块缓存
go clean -modcache
go mod tidy
go build -o inspector.exe ./cmd/inspector/
```

### 3. Web服务启动失败
```bash
# 检查端口是否被占用
netstat -an | grep 8080

# 使用其他端口
./inspector.exe web --port 9090
```

### 4. 前端页面空白
- 确保前端已正确构建到 `code/cmd/inspector/web/frontend/dist`
- 检查浏览器控制台错误信息
- 确认API接口正常响应

## 🎉 使用体验

启动成功后，您可以：

1. **查看系统概览** - 在仪表板了解集群状态
2. **管理巡检规则** - 浏览和编辑各类检查规则
3. **执行巡检任务** - 选择模板或自定义配置进行巡检
4. **分析历史结果** - 查看详细报告和趋势分析
5. **个性化设置** - 根据需求调整系统配置

这个Web界面将命令行工具的强大功能转化为直观的可视化操作，大大提升了K8s Resource Inspector的易用性！

## 📞 技术支持

如果遇到问题，请检查：
1. 环境要求是否满足
2. 构建步骤是否正确执行
3. 端口是否被占用
4. 浏览器控制台是否有错误信息
