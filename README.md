# K8s-Resource-Inspector

## 项目简介

K8s-Resource-Inspector 是一个专注于 Kubernetes 资源配置审计、合规检查和最佳实践验证的多集群资源巡检工具。它能够帮助 DevOps 团队和平台工程师快速识别集群中的配置问题、安全风险和潜在的性能瓶颈，确保集群资源符合企业标准和最佳实践。

## 核心功能

- **多集群管理**：统一管理和巡检多个 Kubernetes 集群
- **配置审计**：检查资源配置是否符合企业标准和最佳实践
- **合规性检查**：验证集群是否符合行业标准（如 CIS 基准）
- **资源优化**：识别资源浪费和提供优化建议
- **灵活报告**：支持多种报告格式和输出方式
- **自动化集成**：支持与 CI/CD 流程集成，实现自动化巡检

## 安装与使用

### 安装方法

#### 方法1: 使用预编译二进制文件

1. 从 GitHub Releases 页面下载最新版本的二进制文件
2. 解压并移动到系统路径中:

```bash
# Linux/MacOS
chmod +x inspector
sudo mv inspector /usr/local/bin/

# Windows
# 将inspector.exe移动到你的PATH路径中
```

#### 方法2: 从源码构建

需要Go 1.17+版本:

```bash
git clone https://github.com/FreshMan1123/k8s-resource-inspector.git
cd k8s-resource-inspector/code
go build -o inspector cmd/inspector/main.go
```

#### 方法3: 使用Docker镜像

```bash
docker pull freshman1123/k8s-resource-inspector:latest
docker run --rm -v ~/.kube:/root/.kube freshman1123/k8s-resource-inspector:latest cluster info
```

### 使用指南

#### 集群管理

K8s-Resource-Inspector支持管理多个Kubernetes集群的配置:

```bash
# 列出所有可用的集群上下文
inspector cluster list

# 切换当前集群上下文
inspector cluster use my-cluster

# 显示当前集群信息
inspector cluster info

# 添加新的集群配置
inspector cluster add --name production --file /path/to/kubeconfig
```

#### 资源管理

查看和管理集群中的各种资源:

```bash
# 获取所有命名空间中的资源
inspector resource get pods --all-namespaces

# 获取特定命名空间的资源
inspector resource get deployments -n kube-system

# 列出命名空间
inspector resource namespace list
```

#### 资源巡检

进行资源配置审计和健康检查:

```bash
# 检查节点状态
inspector inspect node

# 使用自定义规则文件进行检查
inspector inspect node --rules-file /path/to/rules.yaml

# 生成JSON格式报告并保存到文件
inspector inspect node --output json --output-file node-report.json

# 只显示有问题的资源
inspector inspect node --only-issues
```

### 常用示例

#### 示例1: 生成节点健康报告

```bash
# 生成详细的节点健康报告
inspector inspect node --output text
```

输出示例:
```
========================================
    K8S RESOURCE INSPECTOR REPORT
========================================

Generated: Wed, 23 Jun 2023 14:30:45 CST
Cluster:   production

SUMMARY
----------------------------------------
Total resources analyzed:    5
Resources with issues:       2

Issue severity breakdown:
  CRITICAL 1
  ERROR    2
  WARNING  3
  INFO     1

FINDINGS
----------------------------------------

Resource: Node/worker-1
-----------------------
[WARNING] Rule: node-cpu-usage
Message: 节点CPU使用率过高
Recommendation: 考虑添加更多节点或优化工作负载
CPU Utilization: 85.3%
Memory Utilization: 67.2%
```

#### 示例2: 检查特定命名空间的资源

```bash
# 检查生产命名空间中的所有Pods
inspector resource get pods -n production
```

## 配置与自定义

### 规则配置

K8s-Resource-Inspector使用YAML格式的规则文件定义检查标准。规则文件示例:

```yaml
apiVersion: v1
kind: RulesConfig
config:
  environment: prod
clusterEnvironments:
  production-cluster: prod
  staging-cluster: dev
  default: dev
rules:
  - id: node-cpu-usage
    name: "节点CPU使用率检查"
    description: "检查节点CPU使用率是否超过阈值"
    category: node
    severity: warning
    condition:
      metric: cpu.utilization
      operator: "<"
      thresholds:
        prod: 80
        dev: 90
    remediation: "考虑添加更多节点或优化工作负载"
    enabled: true
```

### 输出格式

支持多种输出格式:

- `text`: 人类可读的格式化文本 (默认)
- `json`: JSON格式，便于程序处理
- `yaml`: YAML格式输出

## 项目结构

项目采用模块化设计，清晰划分功能边界：

```
code/
├── cmd/                      # 命令行入口点
│   └── inspector/            # 主程序入口
│       ├── main.go           # 程序主入口
│       ├── resource.go       # 资源管理命令
│       ├── cluster.go        # 集群管理命令
│       └── resource/         # 资源相关子命令
│           ├── get.go        # 获取资源命令
│           ├── namespace.go  # 命名空间命令
│           └── node.go       # 节点检查命令
├── internal/                 # 内部实现代码(不对外暴露)
│   ├── cluster/              # 集群管理相关
│   │   └── client.go         # 集群客户端
│   ├── kubeconfig/           # kubeconfig管理
│   │   └── manager.go        # 配置文件管理器
│   ├── collector/            # 数据收集层
│   │   ├── node.go           # 节点数据收集
│   │   └── pod.go            # Pod数据收集(计划)
│   ├── rules/                # 规则引擎
│   │   ├── engine.go         # 规则执行引擎
│   │   ├── loader.go         # 规则加载器
│   │   └── types.go          # 规则类型定义
│   ├── analyzer/             # 分析器层
│   │   ├── report.go         # 报告生成器
│   │   ├── node/             # 节点资源分析
│   │   │   └── analyzer.go   # 节点分析器
│   │   └── pod/              # Pod资源分析(计划)
│   └── models/               # 数据模型定义
│       └── node.go           # 节点数据模型
├── configs/                  # 配置文件
│   └── rules/                # 规则定义文件
│       └── node.yaml         # 节点规则
└── pkg/                      # 可供外部使用的包(预留)
```

## 常见问题

### Q: 为什么需要设置自定义规则文件?

A: 默认规则适用于一般场景，但不同组织有不同的标准和需求。通过自定义规则文件，可以根据组织特定要求调整检查标准。

### Q: 如何在CI/CD流程中集成K8s-Resource-Inspector?

A: 可以在部署前的验证阶段运行检查命令，例如:

```bash
inspector inspect node --output json --output-file report.json --only-issues
# 根据report.json的内容决定是否允许部署
```

### Q: 工具报告某些问题，但实际上这是预期行为，如何忽略?

A: 可以通过以下方式解决:
1. 修改规则文件，关闭或调整特定规则
2. 使用`--rules-file`参数指定自定义规则文件

## 贡献指南

我们欢迎社区贡献，请参考[贡献指南](CONTRIBUTING.md)了解如何参与项目开发。

## 迭代规划

### 迭代 v0.1：基础框架（2周）

**主要目标**：建立基础框架，实现单集群连接和基本资源检查

**具体任务**：
1. 项目结构搭建
2. 单集群连接实现
3. 获取基本资源（Pod、Deployment等）
4. 实现简单检查逻辑

**交付物**：可运行的CLI工具，能连接集群并列出资源状态

### 迭代 v0.2：规则引擎（2周）

**主要目标**：构建基于YAML的灵活规则引擎，实现基础报告生成

**具体任务**：
1. 基于YAML的规则引擎设计与实现
2. 可配置规则集的加载与评估
3. 多格式报告生成（表格、JSON、YAML）
4. 分层架构（收集器-规则引擎-分析器）实现

**交付物**：支持外部可配置规则的巡检工具，能生成多格式报告

### 迭代 v0.3：多集群支持（2周）

**主要目标**：扩展为多集群巡检工具，增强报告功能

**具体任务**：
1. 多集群管理实现
2. 并行巡检能力
3. JSON/Markdown报告格式支持
4. 资源过滤功能

**交付物**：支持多集群的巡检工具，具备多格式报告能力

### 迭代 v0.4：容器化与自动化（2周）

**主要目标**：实现容器化部署和自动化流程

**具体任务**：
1. Dockerfile编写（多阶段构建）
2. Kubernetes部署清单创建
3. GitHub Actions CI流程设置
4. 定时巡检任务实现

**交付物**：容器化的巡检工具，完整的CI流程，支持定时巡检

### 迭代 v0.5：集成与扩展（3周）

**主要目标**：增加企业级功能，提升用户体验

**具体任务**：
1. 告警集成（Slack、邮件等）
2. 历史报告存储与对比
3. 更多资源类型支持
4. 性能优化与扩展性提升

**交付物**：企业级巡检系统，支持告警和历史对比

## 技术栈

- **语言**：Go
- **K8s交互**：client-go、metrics-client
- **CLI框架**：cobra、viper
- **规则配置**：YAML
- **容器化**：Docker
- **CI/CD**：GitHub Actions
- **部署**：Kubernetes

## 详细文档

- [架构设计文档](docs/architecture.md) - 详细的架构说明和规则配置指南

## 使用场景

- **定期安全审计**：检查所有集群是否存在安全风险
- **上线前检查**：确保新应用配置符合最佳实践
- **合规性报告**：为安全团队生成定期合规报告
- **配置标准化**：确保多集群环境中资源配置的一致性
- **问题根因分析**：快速检查资源配置是否符合预期

## 后续规划

- **Web界面**：提供图形化操作界面
- **插件系统**：支持自定义规则和报告格式
- **高级分析**：基于历史数据的趋势分析
- **自动修复**：对于常见问题提供自动修复能力 