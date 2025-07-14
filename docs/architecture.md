# K8s-Resource-Inspector 架构设计

## 概述

K8s-Resource-Inspector 采用模块化的分层架构设计，以实现关注点分离和功能扩展性。核心架构分为三个主要层级：

1. **数据收集层**：负责从Kubernetes集群收集原始指标和状态数据
2. **规则引擎层**：负责定义、加载和评估检查规则
3. **分析器层**：将收集的数据与规则结合，生成分析结果和改进建议

这种设计允许系统各部分独立演化，同时保持整体功能的一致性。

## 架构图

```
┌─────────────────────┐     ┌───────────────────┐     ┌─────────────────┐
│                     │     │                   │     │                 │
│   数据收集层        │────▶│   规则引擎层      │────▶│   分析器层      │
│   (Collector)       │     │   (Rules)         │     │   (Analyzer)    │
│                     │     │                   │     │                 │
└─────────────────────┘     └───────────────────┘     └─────────────────┘
         │                          ▲                         │
         │                          │                         │
         │                  ┌───────────────┐                 │
         │                  │               │                 │
         └─────────────────▶│  数据模型     │◀────────────────┘
                            │  (Models)     │
                            │               │
                            └───────────────┘
                                    ▲
                                    │
                            ┌───────────────┐
                            │               │
                            │  配置规则     │
                            │  (Configs)    │
                            │               │
                            └───────────────┘
```

## 模块说明

### 数据收集层 (Collector)

数据收集层负责从Kubernetes API获取原始数据，并将其转换为标准化的内部数据结构。

主要组件：
- `collector/node.go`: 节点数据收集，包括资源使用情况、状态、条件等
- `collector/pod.go`: Pod数据收集（计划中）

这一层的关键功能：
- 连接Kubernetes集群
- 获取资源数据
- 规范化数据格式
- 缓存数据（减少API调用）

### 规则引擎层 (Rules)

规则引擎层定义了检查规则的结构，以及如何加载和评估这些规则。

主要组件：
- `rules/types.go`: 定义规则结构和接口
- `rules/loader.go`: 从YAML文件加载规则
- `rules/engine.go`: 评估规则逻辑

这一层的关键功能：
- 定义规则数据结构
- 加载外部规则定义
- 提供规则评估引擎
- 生成规则评估结果

### 分析器层 (Analyzer)

分析器层结合收集的数据和规则，生成分析报告和改进建议。

主要组件：
- `analyzer/node/analyzer.go`: 节点资源分析器
- `analyzer/report.go`: 报告生成器

这一层的关键功能：
- 调用收集器获取数据
- 应用规则引擎评估数据
- 生成问题报告
- 提供改进建议

### 数据模型 (Models)

定义系统中使用的数据结构，连接各个层。

主要组件：
- `models/node.go`: 节点相关数据结构
- 其他资源类型的模型（计划中）

### 命令行接口 (CLI)

提供用户交互的命令行界面。

主要组件：
- `cmd/inspector/resource/node.go`: 节点检查命令
- 其他资源类型的命令（计划中）

## 基于YAML的规则配置

K8s-Resource-Inspector 使用YAML格式定义检查规则，使非开发人员也能轻松配置和维护规则。

### 规则结构

```yaml
rules:
  - name: "high_cpu_utilization"     # 规则名称
    description: "节点CPU使用率过高"   # 规则描述
    category: "node"                 # 规则类别
    severity: "warning"              # 严重程度：critical, warning, info
    condition:                       # 触发条件
      metric: "cpu_utilization"      # 要检查的指标
      operator: ">"                  # 比较操作符
      threshold: 80                  # 阈值
    remediation: "考虑扩展集群或限制节点上的工作负载"  # 修复建议
    enabled: true                    # 规则是否启用
```

### 支持的操作符

- `>` : 大于
- `>=`: 大于等于
- `<` : 小于
- `<=`: 小于等于
- `==`: 等于
- `!=`: 不等于
- `contains`: 包含（字符串）
- `not_contains`: 不包含（字符串）
- `exists`: 存在
- `not_exists`: 不存在

### 节点指标

以下是可以在节点规则中使用的一些关键指标：

- `status`: 节点状态（Ready, NotReady等）
- `cpu_utilization`: CPU使用率百分比
- `memory_utilization`: 内存使用率百分比
- `pods_utilization`: Pod使用率百分比
- `ephemeral_storage_utilization`: 临时存储使用率百分比
- `unschedulable`: 节点是否可调度
- `memory_pressure`: 节点是否处于内存压力状态
- `disk_pressure`: 节点是否处于磁盘压力状态
- `pid_pressure`: 节点是否处于PID压力状态
- `network_unavailable`: 节点网络是否不可用
- `container_runtime`: 容器运行时
- `kubelet_version`: Kubelet版本

## 使用自定义规则

用户可以通过以下方式使用自定义规则：

1. 在 `configs/rules/` 目录中创建或编辑规则文件
2. 使用命令行参数 `--rules-dir` 指定自定义规则目录
3. 规则文件应按资源类型命名（如 `node.yaml`, `pod.yaml`）

例如：

```bash
# 使用自定义规则目录检查节点
k8s-resource-inspector resource node --rules-dir=/path/to/custom/rules
```

## 扩展性

该架构设计允许系统轻松扩展以支持新的资源类型和规则：

1. 添加新的收集器以支持新的资源类型
2. 定义新的数据模型
3. 创建相应的规则评估逻辑
4. 实现分析器
5. 添加命令行接口

没有需要修改现有代码，只需添加新的模块。 