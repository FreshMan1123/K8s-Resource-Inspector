# Cobra 使用文档

## 基本概念

Cobra 是一个用于创建强大的现代化 CLI 应用程序的 Go 库，被广泛应用于 Kubernetes、Hugo、GitHub CLI 等项目中。

## 命令结构

### 命令定义

```go
var rootCmd = &cobra.Command{
    Use:   "inspector",
    Short: "K8s-Resource-Inspector是一个Kubernetes资源配置审计和合规检查工具",
    Long: `K8s-Resource-Inspector是一个专注于Kubernetes资源配置审计、合规检查和最佳实践验证的多集群资源巡检工具。
它能够帮助DevOps团队和平台工程师快速识别集群中的配置问题、安全风险和潜在的性能瓶颈，
确保集群资源符合企业标准和最佳实践。`,
}
```

### 命令执行函数

命令执行函数接收两个参数：
- `cmd *cobra.Command`：当前执行的命令对象
- `args []string`：命令行中传入的位置参数的切片

```go
Run: func(cmd *cobra.Command, args []string) {
    // 命令执行逻辑
}
```

## 标志（Flags）

### 持久性标志（PersistentFlags）

持久性标志会应用于当前命令及其所有子命令：

```go
rootCmd.PersistentFlags().StringP("kubeconfig", "k", "", "kubeconfig文件路径 (默认为$HOME/.kube/config)")
```

### 本地标志（Flags）

本地标志只应用于当前命令：

```go
clusterAddCmd.Flags().StringVarP(&clusterName, "name", "n", "", "集群的名称")
```

## 参数处理

### 位置参数

位置参数通过 `args []string` 获取：

```go
var clusterUseCmd = &cobra.Command{
    Use:   "use [context-name]",
    Short: "切换到指定的集群上下文",
    Args:  cobra.ExactArgs(1), // 指定必须有一个参数
    Run: func(cmd *cobra.Command, args []string) {
        contextName := args[0] // 获取第一个位置参数
        // ...
    },
}
```

### 标志参数

标志参数通过 `cmd.Flags()` 获取：

```go
configPath, _ := cmd.Flags().GetString("kubeconfig")
contextName, _ := cmd.Flags().GetString("contextName")
```

## 位置参数与标志参数的区别

位置参数和标志参数是CLI应用程序中两种不同的参数传递方式，各有优缺点：

### 位置参数 (args[0])
- **定义方式**：基于参数的位置识别参数用途
- **使用示例**：`resource get pods`，其中`pods`是位置参数
- **特点**：
  - 使用简洁，减少输入
  - 适用于必选且概念清晰的参数
  - 顺序固定且重要
  - 不支持默认值（除非通过代码逻辑处理）
  - 无自描述性（没有参数名称）
  - 需要记住参数顺序
- **适用场景**：简单、常用且顺序明确的操作

### 标志参数 (-f, --file)
- **定义方式**：通过名称标识符识别参数用途
- **使用示例**：`resource apply -f deployment.yaml`
- **特点**：
  - 使用明确，参数含义清晰
  - 支持可选参数
  - 顺序不重要
  - 可设置默认值
  - 自描述性强（通过参数名）
  - 支持短名称(-f)和长名称(--file)
  - 可以设置是否必需
- **适用场景**：可选配置、复杂选项或需要说明的参数

### 选择建议
- 对于必需的、概念清晰的核心操作参数，使用位置参数
- 对于可选的、配置性的、含义需要说明的参数，使用标志参数
- 两种方式可以混合使用，如`kubectl get pods -n default`
- 保持与行业标准工具（如kubectl）的一致性

### 实际例子
```go
// 位置参数示例 (resource get)
cmd := &cobra.Command{
    Use:   "get [resource-type] [name]",
    Args:  cobra.RangeArgs(1, 2),
    Run: func(cmd *cobra.Command, args []string) {
        resourceType := args[0]  // 使用位置参数
        // ...
    },
}

// 标志参数示例 (resource apply)
cmd := &cobra.Command{
    Use:   "apply -f [file]",
    Run: func(cmd *cobra.Command, args []string) {
        filePath, _ := cmd.Flags().GetString("file")  // 使用标志参数
        // ...
    },
}

// 添加标志
cmd.Flags().StringP("file", "f", "", "包含资源定义的YAML文件路径")
cmd.MarkFlagRequired("file")  // 将标志标记为必需
```

## 命令树结构

Cobra 使用命令树结构，通过 `AddCommand` 添加子命令：

```go
// 添加子命令到resource命令
resourceCmd.AddCommand(resourceGetPodsCmd)
resourceCmd.AddCommand(resourceGetServicesCmd)

// 添加resource命令到根命令
rootCmd.AddCommand(resourceCmd)
```

## 常见用例

### 获取配置路径

```go
func getConfigPath(cmd *cobra.Command) string {
    configPath, _ := cmd.Flags().GetString("kubeconfig")
    if configPath == "" {
        if home := homedir.HomeDir(); home != "" {
            configPath = filepath.Join(home, ".kube", "config")
        }
    }
    return configPath
}
```

### getConfigPath 调用流程

`configPath := getConfigPath(cmd)` 的调用流程如下：

1. **函数调用**：
   - 当执行命令（如 `cluster use` 或 `resource pods`）时，会调用对应命令的 `Run` 函数
   - 在 `Run` 函数中，首先调用 `getConfigPath(cmd)` 获取 kubeconfig 文件路径

2. **获取标志值**：
   - 函数通过 `cmd.Flags().GetString("kubeconfig")` 获取命令行中 `--kubeconfig` 标志的值
   - 这个标志是在 `main.go` 的 `init()` 函数中通过 `rootCmd.PersistentFlags()` 定义的全局标志
   - 如果用户指定了这个标志（如 `--kubeconfig=/path/to/config`），则返回用户指定的值

3. **默认路径处理**：
   - 如果 `configPath` 为空（用户未指定），则使用默认路径
   - 通过 `homedir.HomeDir()` 获取用户主目录
   - 然后使用 `filepath.Join()` 拼接路径，形成默认的 kubeconfig 路径：`$HOME/.kube/config`

4. **返回结果**：
   - 函数返回最终确定的 kubeconfig 文件路径
   - 这个路径会被用于后续的集群操作，如连接集群、获取资源等

这个流程确保了用户可以通过 `--kubeconfig` 标志指定自定义的配置文件路径，如果未指定则使用标准的默认路径。

### 参数验证

```go
Args: cobra.ExactArgs(1), // 必须有且只有一个参数
```

其他常见验证：
- `cobra.MinimumNArgs(n)` - 至少需要 n 个参数
- `cobra.MaximumNArgs(n)` - 最多允许 n 个参数
- `cobra.NoArgs` - 不允许任何参数 