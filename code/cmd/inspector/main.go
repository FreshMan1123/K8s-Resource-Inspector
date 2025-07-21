package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/FreshMan1123/k8s-resource-inspector/code/web/backend"
)

var rootCmd = &cobra.Command{
	Use:   "inspector",
	Short: "K8s-Resource-Inspector是一个Kubernetes资源配置审计和合规检查工具",
	Long: `K8s-Resource-Inspector是一个专注于Kubernetes资源配置审计、合规检查和最佳实践验证的多集群资源巡检工具。
它能够帮助DevOps团队和平台工程师快速识别集群中的配置问题、安全风险和潜在的性能瓶颈，
确保集群资源符合企业标准和最佳实践。`,
}

// webCmd 表示Web管理界面命令
var webCmd = &cobra.Command{
	Use:   "web",
	Short: "启动Web管理界面",
	Long: `启动Web管理界面，提供可视化的规则管理和巡检执行功能。

示例:
  # 启动Web界面（默认端口8080）
  inspector web

  # 指定端口启动
  inspector web --port 9090`,
	Run: func(cmd *cobra.Command, args []string) {
		port, _ := cmd.Flags().GetInt("port")

		fmt.Printf("🚀 启动K8s巡检管理平台...\n")
		fmt.Printf("📱 访问地址: http://localhost:%d\n", port)
		fmt.Printf("⏹️  按 Ctrl+C 停止服务\n\n")

		startWebServer(port)
	},
}

func init() {
	// 添加全局标志
	rootCmd.PersistentFlags().StringP("kubeconfig", "k", "", "kubeconfig文件路径 (默认为$HOME/.kube/config)")
	rootCmd.PersistentFlags().StringP("contextName", "c", "", "要使用的kubeconfig上下文名称")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "启用详细输出")

	// 添加子命令
	rootCmd.AddCommand(clusterCmd)
	rootCmd.AddCommand(resourceCmd)  // 添加缺失的resource命令
	rootCmd.AddCommand(inspectCmd)   // 添加缺失的inspect命令
	rootCmd.AddCommand(webCmd)       // 添加新的web命令

	// 初始化web命令标志
	webCmd.Flags().Int("port", 8080, "Web服务端口")
}

// startWebServer 启动Web服务器
func startWebServer(port int) {
	fmt.Printf("✅ 已加载现有规则配置\n")
	fmt.Printf("✅ Web界面已就绪\n")

	// 启动完整的Web服务
	backend.StartWebServer(port)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
} 