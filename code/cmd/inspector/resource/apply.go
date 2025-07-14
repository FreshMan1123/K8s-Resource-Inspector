package resource

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/cluster"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/kubernetes/scheme"
)

// NewApplyCommand 创建apply命令
func NewApplyCommand() *cobra.Command {
	ioStreams := genericclioptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}

	cmd := &cobra.Command{
		Use:   "apply -f [file]",
		Short: "从文件创建或更新资源",
		Long:  `从YAML文件创建或更新Kubernetes资源。支持单个资源文件或包含多个资源的文件。`,
		Run: func(cmd *cobra.Command, args []string) {
			// 获取文件路径
			filePath, _ := cmd.Flags().GetString("file")
			if filePath == "" {
				fmt.Println("错误: 必须使用 -f 指定文件路径")
				os.Exit(1)
			}

			// 获取kubeconfig和context
			configPath, _ := cmd.Flags().GetString("kubeconfig")
			contextName, _ := cmd.Flags().GetString("contextName")

			// 创建集群客户端
			client, err := cluster.NewClient(configPath, contextName)
			if err != nil {
				fmt.Printf("创建集群客户端失败: %v\n", err)
				os.Exit(1)
			}

			// 创建资源构建器配置
			configFlags := genericclioptions.NewConfigFlags(true)
			if configPath != "" {
				configFlags.KubeConfig = &configPath
			}
			if contextName != "" {
				configFlags.Context = &contextName
			}

			// 创建资源构建器
			builder := resource.NewBuilder(configFlags).
				Unstructured().
				Schema(scheme.Scheme).
				ContinueOnError().
				NamespaceParam("").DefaultNamespace().
				FilenameParam(false, &resource.FilenameOptions{
					Filenames: []string{filePath},
				})

			// 构建对象
			result := builder.Do()
			if err := result.Err(); err != nil {
				fmt.Printf("解析资源文件失败: %v\n", err)
				os.Exit(1)
			}

			// 处理每个对象
			count := 0
			err = result.Visit(func(info *resource.Info, err error) error {
				if err != nil {
					return err
				}

				// 获取对象数据
				obj := info.Object
				name := info.Name
				namespace := info.Namespace
				kind := info.Mapping.GroupVersionKind.Kind

				// 尝试创建或更新资源
				helper := resource.NewHelper(info.Client, info.Mapping)
				
				// 首先尝试获取资源
				_, err = helper.Get(namespace, name)
				if err != nil {
					// 资源不存在，创建它
					_, err = helper.Create(namespace, true, obj)
					if err != nil {
						return fmt.Errorf("创建资源 %s '%s' 失败: %w", kind, name, err)
					}
					fmt.Printf("已创建 %s '%s'%s\n", kind, name, namespaceInfo(namespace))
				} else {
					// 资源存在，更新它
					_, err = helper.Replace(namespace, name, true, obj)
					if err != nil {
						return fmt.Errorf("更新资源 %s '%s' 失败: %w", kind, name, err)
					}
					fmt.Printf("已更新 %s '%s'%s\n", kind, name, namespaceInfo(namespace))
				}

				count++
				return nil
			})

			if err != nil {
				fmt.Printf("应用资源失败: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("成功应用了 %d 个资源\n", count)
		},
	}

	// 添加文件标志
	cmd.Flags().StringP("file", "f", "", "包含资源定义的YAML文件路径")
	cmd.MarkFlagRequired("file")

	return cmd
}

// namespaceInfo 返回命名空间信息的格式化字符串
func namespaceInfo(namespace string) string {
	if namespace == "" || namespace == "default" {
		return ""
	}
	return fmt.Sprintf(" (namespace: %s)", namespace)
} 