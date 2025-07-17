package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/cluster"
	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/kubeconfig"
	"k8s.io/client-go/util/homedir"
)

var (
	// 集群命令的配置选项
	clusterConfigPath string
	clusterName       string
)

// clusterCmd 表示集群管理命令，cobra.Command是一个结构体类型，取地址符使其返回一个指针，来在不同函数之间使用。
var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "管理Kubernetes集群连接",
	Long:  `管理Kubernetes集群连接，包括添加、列出、切换和删除集群配置。`,
	Run: func(cmd *cobra.Command, args []string) {
		// 默认显示帮助信息
		if err := cmd.Help(); err != nil {
			fmt.Printf("显示帮助信息失败: %v\n", err)
		}
	},
}

// clusterListCmd 表示列出集群命令
var clusterListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出所有可用的集群",
	Long:  `列出kubeconfig中所有可用的集群上下文。`,
	Run: func(cmd *cobra.Command, args []string) {
		configPath := getConfigPath(cmd)
		
		// 获取当前上下文
		currentContext, err := cluster.GetCurrentContext(configPath)
		if err != nil {
			fmt.Printf("获取当前上下文失败: %v\n", err)
			os.Exit(1)
		}
		
		// 列出所有上下文
		contexts, err := cluster.ListContexts(configPath)
		if err != nil {
			fmt.Printf("列出上下文失败: %v\n", err)
			os.Exit(1)
		}
		
		fmt.Println("可用的集群上下文:")
		for _, ctx := range contexts {
			if ctx == currentContext {
				fmt.Printf("* %s (当前)\n", ctx)
			} else {
				fmt.Printf("  %s\n", ctx)
			}
		}
	},
}

// clusterUseCmd 表示切换集群命令
var clusterUseCmd = &cobra.Command{
	Use:   "use [context-name]",
	Short: "切换到指定的集群上下文",
	Long:  `切换当前活动的Kubernetes集群上下文。`,
	Args:  cobra.ExactArgs(1), //指定传入参数数量，必须得是一个，否则会报错
	Run: func(cmd *cobra.Command, args []string) {
		configPath := getConfigPath(cmd)
		//Args验证确保args切片的长度为1，验证通过后，Run函数才会被调用
		contextName := args[0]
		
		// 切换上下文
		err := cluster.SwitchContext(configPath, contextName)
		if err != nil {
			fmt.Printf("切换上下文失败: %v\n", err)
			os.Exit(1)
		}
		
		fmt.Printf("已切换到上下文 '%s'\n", contextName)
	},
}

// clusterAddCmd 表示添加集群命令
var clusterAddCmd = &cobra.Command{
	Use:   "add",
	Short: "添加新的集群配置",
	Long:  `添加新的Kubernetes集群配置到安全存储。`,
	Run: func(cmd *cobra.Command, args []string) {
		// 获取安全存储目录
		secureDir := filepath.Join("code", "internal", "config", "secure")
		if err := os.MkdirAll(secureDir, 0700); err != nil {
			fmt.Printf("创建安全存储目录失败: %v\n", err)
			os.Exit(1)
		}
		
		// 创建kubeconfig管理器,由package包导入，这个kubeconfig是由我们自己创建并上推到github仓库的
		// 所以需要我们自己导入
		manager, err := kubeconfig.NewManager(secureDir)
		if err != nil {
			fmt.Printf("创建kubeconfig管理器失败: %v\n", err)
			os.Exit(1)
		}
		
		// 从源文件读取kubeconfig内容
		sourcePath := clusterConfigPath
		if sourcePath == "" {
			if home := homedir.HomeDir(); home != "" {
				// 如果用户没有指定路径，则直接走默认路径
				sourcePath = filepath.Join(home, ".kube", "config")
			} else {
				fmt.Println("无法确定家目录，请明确指定kubeconfig路径")
				os.Exit(1)
			}
		}
		
		content, err := os.ReadFile(sourcePath)
		if err != nil {
			fmt.Printf("读取kubeconfig文件失败: %v\n", err)
			os.Exit(1)
		}
		
		// 保存到安全存储
		if err := manager.SaveKubeconfig(clusterName, content); err != nil {
			fmt.Printf("保存kubeconfig失败: %v\n", err)
			os.Exit(1)
		}
		
		fmt.Printf("集群配置 '%s' 已添加\n", clusterName)
	},
}

// clusterInfoCmd 表示获取集群信息命令
var clusterInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "显示当前集群信息",
	Long:  `显示当前连接的Kubernetes集群的详细信息。`,
	Run: func(cmd *cobra.Command, args []string) {
		configPath := getConfigPath(cmd)
		contextName, _ := cmd.Flags().GetString("contextName")
		
		// 创建集群客户端
		client, err := cluster.NewClient(configPath, contextName)
		if err != nil {
			fmt.Printf("创建集群客户端失败: %v\n", err)
			os.Exit(1)
		}
		
		// 获取集群版本
		version, err := client.GetServerVersion()
		if err != nil {
			fmt.Printf("获取集群版本失败: %v\n", err)
			os.Exit(1)
		}
		
		// 获取当前上下文
		currentContext, err := cluster.GetCurrentContext(configPath)
		if err != nil {
			fmt.Printf("获取当前上下文失败: %v\n", err)
			os.Exit(1)
		}
		
		fmt.Println("集群信息:")
		fmt.Printf("当前上下文: %s\n", currentContext)
		fmt.Printf("配置文件: %s\n", configPath)
		fmt.Printf("集群版本: %s\n", version)
	},
}

func init() {
	// 添加子命令到cluster命令
	clusterCmd.AddCommand(clusterListCmd)
	clusterCmd.AddCommand(clusterUseCmd)
	clusterCmd.AddCommand(clusterAddCmd)
	clusterCmd.AddCommand(clusterInfoCmd)
	
	// 添加cluster add命令的标志
	clusterAddCmd.Flags().StringVarP(&clusterConfigPath, "file", "f", "", "要添加的kubeconfig文件路径")
	clusterAddCmd.Flags().StringVarP(&clusterName, "name", "n", "", "集群的名称")
	if err := clusterAddCmd.MarkFlagRequired("name"); err != nil {
		fmt.Printf("标记必需标志失败: %v\n", err)
		os.Exit(1)
	}
}

// getConfigPath 获取kubeconfig文件路径
func getConfigPath(cmd *cobra.Command) string {
	configPath, _ := cmd.Flags().GetString("kubeconfig")
	if configPath == "" {
		if home := homedir.HomeDir(); home != "" {
			configPath = filepath.Join(home, ".kube", "config")
		}
	}
	return configPath
} 