package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/FreshMan1123/k8s-resource-inspector/code/cmd/inspector/web"
)

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "启动Web管理界面",
	Long: `启动K8s巡检管理平台的Web界面，提供可视化的规则管理和巡检执行功能。

示例:
  # 启动Web服务，默认端口8080
  inspector web

  # 指定端口启动
  inspector web --port 9090

  # 启动后访问浏览器
  http://localhost:8080`,
	Run: func(cmd *cobra.Command, args []string) {
		port, _ := cmd.Flags().GetInt("port")
		if err := web.StartServer(port); err != nil {
			fmt.Printf("启动Web服务失败: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	webCmd.Flags().IntP("port", "p", 8080, "Web服务端口")
	rootCmd.AddCommand(webCmd)
}
