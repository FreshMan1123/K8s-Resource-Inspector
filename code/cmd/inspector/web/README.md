# K8s巡检管理平台 - Web模块

## 项目概述

基于现有K8s-Resource-Inspector CLI工具，构建企业级Web管理平台。

## 目录结构

```
code/web/
├── README.md                    # 本文件
├── docs/                        # 项目文档
│   ├── project-plan.md         # 完整项目规划
│   ├── phase-breakdown.md      # 阶段拆分方案
│   └── technical-design.md     # 技术设计文档
├── backend/                     # 后端代码 (待创建)
│   ├── main.go                 # Web服务入口
│   ├── handlers/               # HTTP处理器
│   ├── services/               # 业务逻辑层
│   └── models/                 # 数据模型
├── frontend/                    # 前端代码 (待创建)
│   ├── src/                    # Vue源码
│   ├── public/                 # 静态资源
│   └── package.json            # 前端依赖
└── data/                        # 数据存储 (待创建)
    ├── inspection_history/      # 巡检历史
    ├── scheduled_tasks/         # 定时任务
    └── audit_logs/             # 审计日志
```

## 开发阶段

### 🎯 阶段1: 最小可行产品 (MVP) - 2周
- 规则展示 (只读)
- 简单巡检执行
- 基础Web框架

### 🚀 阶段2: 规则管理增强 - 2周
- 规则编辑功能
- 规则创建功能
- 规则Apply机制

### 📈 阶段3: 历史记录与结果管理 - 2周
- 巡检历史记录
- 高级巡检功能
- 结果分析功能

### ⏰ 阶段4: 定时任务与自动化 - 2-3周 (可选)
- 定时任务管理
- 通知机制
- 审计日志

## 快速开始

```bash
# 启动Web服务 (阶段1完成后)
cd code/web/backend
go run main.go --port 8080

# 访问Web界面
http://localhost:8080
```

## 当前状态

- ✅ 项目规划文档已完成
- ⏳ 阶段1开发准备中
- ⏳ 后续阶段待实施

## 开发计划

### 近期目标 (阶段1 - 2周)
1. 搭建基础Web框架
2. 实现规则展示功能
3. 集成CLI命令执行
4. 完成MVP版本

### 中期目标 (阶段2 - 2周)
1. 实现规则编辑功能
2. 开发规则创建向导
3. 完善Apply机制
4. 优化用户体验

### 长期目标 (阶段3+ - 4+周)
1. 历史记录管理
2. 定时任务调度
3. 通知机制
4. 审计日志

## 技术栈

- **前端**: Vue 3 + Element Plus + Axios
- **后端**: Go + Gin + Viper
- **存储**: 文件系统 (YAML + JSON)
- **CLI集成**: 直接调用inspector.exe

## 相关文档

- [完整项目规划](docs/project-plan.md) - 详细的功能规划和目标
- [阶段拆分方案](docs/phase-breakdown.md) - 分阶段实施计划
- [技术设计文档](docs/technical-design.md) - 架构和技术细节

## 贡献指南

1. 严格遵循TDD开发模式
2. 每个阶段完成后进行充分测试
3. 保持与现有CLI工具的兼容性
4. 遵循项目代码规范和最佳实践
