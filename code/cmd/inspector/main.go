package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "inspector",
	Short: "K8s-Resource-Inspector是一个Kubernetes资源配置审计和合规检查工具",
	Long: `K8s-Resource-Inspector是一个专注于Kubernetes资源配置审计、合规检查和最佳实践验证的多集群资源巡检工具。
它能够帮助DevOps团队和平台工程师快速识别集群中的配置问题、安全风险和潜在的性能瓶颈，
确保集群资源符合企业标准和最佳实践。`,
}

func init() {
	// 添加全局标志
	rootCmd.PersistentFlags().StringP("kubeconfig", "k", "", "kubeconfig文件路径 (默认为$HOME/.kube/config)")
	rootCmd.PersistentFlags().StringP("contextName", "c", "", "要使用的kubeconfig上下文名称")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "启用详细输出")

	// 添加子命令
	rootCmd.AddCommand(clusterCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
} 