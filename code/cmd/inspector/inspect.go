package main

import (
	"fmt"
	"github.com/FreshMan1123/k8s-resource-inspector/code/cmd/inspector/inspect"
	"github.com/spf13/cobra"
)

var (
	// inspect命令的配置选项
	inspectKubeconfig  string
	inspectContextName string
	inspectOutputFormat string
	inspectNoColor     bool
	inspectRulesFile   string
	inspectOutputFile  string
	inspectOnlyIssues  bool
)

// inspectCmd 表示资源检查命令
var inspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "检查Kubernetes资源",
	Long:  `检查Kubernetes集群中的资源状态并生成详细报告，可以检测资源配置问题和潜在风险。`,
	Run: func(cmd *cobra.Command, args []string) {
		// 默认显示帮助信息
		if err := cmd.Help(); err != nil {
			fmt.Printf("显示帮助信息失败: %v\n", err)
		}
	},
}

func init() {
	// 添加标志
	inspectCmd.PersistentFlags().StringVar(&inspectKubeconfig, "kubeconfig", "", "kubeconfig文件路径")
	inspectCmd.PersistentFlags().StringVar(&inspectContextName, "context", "", "要使用的kubeconfig上下文")
	inspectCmd.PersistentFlags().StringVar(&inspectOutputFormat, "output", "text", "报告输出格式 (text, json, yaml)")
	inspectCmd.PersistentFlags().BoolVar(&inspectNoColor, "no-color", false, "禁用颜色输出")
	inspectCmd.PersistentFlags().StringVar(&inspectRulesFile, "rules-file", "", "自定义规则配置文件路径")
	inspectCmd.PersistentFlags().StringVarP(&inspectOutputFile, "output-file", "o", "", "将报告写入文件而不是标准输出")
	inspectCmd.PersistentFlags().BoolVar(&inspectOnlyIssues, "only-issues", false, "只显示有问题的资源")
	
	// 添加子命令 - 使用inspect包中的NewNodeCommand函数
	inspectCmd.AddCommand(inspect.NewNodeCommand(
		&inspectKubeconfig, 
		&inspectContextName,
		&inspectOutputFormat,
		&inspectNoColor,
		&inspectOnlyIssues,
		&inspectRulesFile,
		&inspectOutputFile,
	))
	
	// 添加Pod检查命令
	inspectCmd.AddCommand(inspect.NewPodCommand(
		&inspectKubeconfig, 
		&inspectContextName,
		&inspectOutputFormat,
		&inspectNoColor,
		&inspectOnlyIssues,
		&inspectRulesFile,
		&inspectOutputFile,
	))
	
	// 添加inspect命令到根命令
	rootCmd.AddCommand(inspectCmd)
} 