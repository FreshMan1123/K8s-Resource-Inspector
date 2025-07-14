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