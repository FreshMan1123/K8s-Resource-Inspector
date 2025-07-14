package main

import (


	"github.com/spf13/cobra"
	"github.com/FreshMan1123/k8s-resource-inspector/code/cmd/inspector/resource"
)

var (
	// 资源命令的配置选项
	namespace string
	allNamespaces bool
)

// resourceCmd 表示资源管理命令
var resourceCmd = &cobra.Command{
	Use:   "resource",
	Short: "管理Kubernetes资源",
	Long:  `获取和显示Kubernetes集群中的资源信息。`,
	Run: func(cmd *cobra.Command, args []string) {
		// 默认显示帮助信息
		cmd.Help()
	},
}

func init() {
	// 添加标志
	resourceCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "要查询的命名空间")
	resourceCmd.PersistentFlags().BoolVarP(&allNamespaces, "all-namespaces", "A", false, "是否查询所有命名空间")
	
	// 添加子命令
	resourceCmd.AddCommand(resource.NewGetCommand(&namespace, &allNamespaces))
	resourceCmd.AddCommand(resource.NewNamespaceCommand())
	resourceCmd.AddCommand(resource.NewApplyCommand())
	
	// 添加resource命令到根命令
	rootCmd.AddCommand(resourceCmd)
} 