# K8s Resource Inspector Web Frontend

这是 K8s Resource Inspector 的 Vue.js 前端界面，提供了直观的 Web 界面来管理和执行 Kubernetes 资源巡检。

## 功能特性

- 🎯 **仪表板** - 系统概览和快速操作
- 📏 **规则管理** - 查看、编辑和管理巡检规则
- 🔍 **快速巡检** - 预设模板快速执行巡检
- ⚙️ **自定义巡检** - 灵活配置巡检参数
- 📊 **历史记录** - 查看和分析历史巡检结果
- 🔧 **系统设置** - 个性化配置和系统管理

## 技术栈

- **Vue 3** - 现代化前端框架
- **TypeScript** - 类型安全的 JavaScript
- **Element Plus** - 企业级 UI 组件库
- **Pinia** - 轻量级状态管理
- **Vue Router 4** - 单页面应用路由
- **Vite** - 快速构建工具
- **Axios** - HTTP 客户端

## 快速开始

### 环境要求

- Node.js 16+
- npm 或 yarn
- Go 1.17+ (用于编译后端)

### 🚀 正确的启动方式

**重要**: 这个前端界面需要与Go后端集成使用，通过 `./inspector.exe web` 启动，而不是独立运行。

#### 1. 构建前端资源

```bash
cd code/web/frontend

# Windows
build.bat

# Linux/macOS
chmod +x build.sh
./build.sh
```

#### 2. 编译Go程序

```bash
cd code
go build -o inspector.exe ./cmd/inspector/
```

#### 3. 启动Web服务

```bash
./inspector.exe web
```

启动后访问: http://localhost:8080

### 开发模式 (仅用于前端开发)

如果只是开发前端界面，可以使用开发模式：

```bash
npm run dev
```

启动后访问: http://localhost:3000 (需要后端API服务同时运行)

### 类型检查

```bash
npm run type-check
```

## 项目结构

```
src/
├── main.ts                 # 应用入口
├── App.vue                 # 根组件
├── router/                 # 路由配置
│   └── index.ts
├── stores/                 # Pinia 状态管理
│   ├── rules.ts           # 规则状态
│   └── inspection.ts      # 巡检状态
├── api/                    # API 接口
│   ├── index.ts           # HTTP 客户端
│   ├── rules.ts           # 规则 API
│   └── inspection.ts      # 巡检 API
├── components/             # 组件
│   ├── Layout/            # 布局组件
│   │   ├── AppLayout.vue
│   │   ├── AppHeader.vue
│   │   └── AppSidebar.vue
│   └── Business/          # 业务组件
│       └── RuleCard.vue
├── views/                  # 页面组件
│   ├── Dashboard.vue      # 仪表板
│   ├── Rules/             # 规则管理
│   │   ├── RuleList.vue
│   │   ├── RuleDetail.vue
│   │   └── RuleEditor.vue
│   ├── Inspection/        # 巡检管理
│   │   ├── QuickInspection.vue
│   │   ├── CustomInspection.vue
│   │   └── ExecutionStatus.vue
│   ├── History/           # 历史记录
│   │   ├── HistoryList.vue
│   │   └── HistoryDetail.vue
│   └── Settings/          # 系统设置
│       └── SystemSettings.vue
├── utils/                  # 工具函数
└── styles/                 # 样式文件
    └── main.css
```

## API 接口

前端通过以下 API 与后端通信：

### 规则管理
- `GET /api/rules` - 获取规则分类
- `GET /api/rules/:category` - 获取分类规则
- `PUT /api/rules/:category/:id` - 更新规则

### 巡检执行
- `POST /api/inspection/execute` - 执行巡检
- `GET /api/inspection/status/:id` - 获取执行状态
- `GET /api/inspection/history` - 获取历史记录

## 开发指南

### 添加新页面

1. 在 `src/views/` 下创建 Vue 组件
2. 在 `src/router/index.ts` 中添加路由配置
3. 在侧边栏菜单中添加导航项

### 添加新 API

1. 在 `src/api/` 下定义接口类型和方法
2. 在对应的 store 中调用 API
3. 在组件中使用 store 的状态和方法

### 状态管理

使用 Pinia 进行状态管理：

```typescript
// 在组件中使用
import { useRulesStore } from '@/stores/rules'

const rulesStore = useRulesStore()
await rulesStore.fetchRules()
```

### 样式规范

- 使用 CSS 变量定义主题色彩
- 遵循 Element Plus 的设计规范
- 响应式设计，支持不同屏幕尺寸

## 配置说明

### 代理配置

开发环境下，API 请求会代理到后端服务：

```typescript
// vite.config.ts
server: {
  proxy: {
    '/api': {
      target: 'http://localhost:8080',
      changeOrigin: true
    }
  }
}
```

### 环境变量

可以通过环境变量配置：

- `VITE_API_BASE_URL` - API 基础地址
- `VITE_APP_TITLE` - 应用标题

## 部署

### 静态部署

构建后的文件可以部署到任何静态文件服务器：

```bash
npm run build
# 将 ../dist 目录部署到 Web 服务器
```

### 与后端集成

前端已经完全集成到 Go 后端中：

1. 构建前端: `npm run build` (输出到 `../cmd/inspector/web/frontend/dist`)
2. Go 程序通过 embed 嵌入静态文件
3. 通过 `./inspector.exe web` 提供统一的服务端点

### 集成架构

```
./inspector.exe web
├── 静态文件服务 (前端界面)
├── API 服务 (/api/*)
│   ├── /api/rules - 规则管理
│   ├── /api/inspection - 巡检执行
│   └── /api/history - 历史记录
└── SPA 路由支持 (Vue Router)
```

## 故障排除

### 常见问题

1. **依赖安装失败**
   ```bash
   rm -rf node_modules package-lock.json
   npm install
   ```

2. **类型错误**
   ```bash
   npm run type-check
   ```

3. **API 请求失败**
   - 检查后端服务是否启动
   - 确认代理配置是否正确

### 调试技巧

- 使用 Vue DevTools 调试组件状态
- 在浏览器开发者工具中查看网络请求
- 检查控制台错误信息

## 贡献指南

1. Fork 项目
2. 创建功能分支
3. 提交更改
4. 发起 Pull Request

## 许可证

本项目采用 MIT 许可证。
