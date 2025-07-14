package resource

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"
	
	"github.com/spf13/cobra"
	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/cluster"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewNamespaceCommand 创建namespace命令
func NewNamespaceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "namespace",
		Aliases: []string{"ns"},
		Short:   "查看Kubernetes命名空间",
		Long:    `查看并显示Kubernetes集群中的命名空间信息。`,
		Run: func(cmd *cobra.Command, args []string) {
			configPath, _ := cmd.Flags().GetString("kubeconfig")
			contextName, _ := cmd.Flags().GetString("contextName")
			
			// 创建集群客户端
			client, err := cluster.NewClient(configPath, contextName)
			if err != nil {
				fmt.Printf("创建集群客户端失败: %v\n", err)
				os.Exit(1)
			}

			// 获取命名空间列表
			namespaceList, err := client.Clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				fmt.Printf("获取命名空间列表失败: %v\n", err)
				os.Exit(1)
			}
			
			// 显示命名空间信息
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tSTATUS\tAGE")
			
			for _, ns := range namespaceList.Items {
				// 计算命名空间存在时间
				age := FormatAge(time.Since(ns.CreationTimestamp.Time))
				
				// 获取命名空间状态
				status := string(ns.Status.Phase)
				
				fmt.Fprintf(w, "%s\t%s\t%s\n", 
					ns.Name,
					status,
					age)
			}
			w.Flush()
		},
	}

	return cmd
} 