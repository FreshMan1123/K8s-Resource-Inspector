package main

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/yourusername/k8s-resource-inspector/code/internal/cluster"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

// resourceGetPodsCmd 表示获取Pod资源命令
var resourceGetPodsCmd = &cobra.Command{
	Use:   "pods",
	Short: "获取Pod资源列表",
	Long:  `获取并显示Kubernetes集群中的Pod资源列表。`,
	Run: func(cmd *cobra.Command, args []string) {
		configPath := getConfigPath(cmd)
		contextName, _ := cmd.Flags().GetString("contextName")
		
		// 创建集群客户端
		client, err := cluster.NewClient(configPath, contextName)
		if err != nil {
			fmt.Printf("创建集群客户端失败: %v\n", err)
			os.Exit(1)
		}
		
		// 获取Pod列表
		var podList *corev1.PodList
		var listErr error
		
		if allNamespaces {
			podList, listErr = client.Clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
		} else {
			// 如果未指定命名空间，则使用"default"
			if namespace == "" {
				namespace = "default"
			}
			podList, listErr = client.Clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
		}
		
		if listErr != nil {
			fmt.Printf("获取Pod列表失败: %v\n", listErr)
			os.Exit(1)
		}
		
		// 显示Pod列表
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAMESPACE\tNAME\tSTATUS\tAGE\tIP")
		
		for _, pod := range podList.Items {
			// 计算Pod存在时间
			age := formatAge(time.Since(pod.CreationTimestamp.Time))
			
			// 获取Pod状态
			status := getPodStatus(&pod)
			
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", 
				pod.Namespace,
				pod.Name,
				status,
				age,
				pod.Status.PodIP)
		}
		w.Flush()
	},
}

// resourceGetServicesCmd 表示获取Service资源命令
var resourceGetServicesCmd = &cobra.Command{
	Use:   "services",
	Short: "获取Service资源列表",
	Long:  `获取并显示Kubernetes集群中的Service资源列表。`,
	Run: func(cmd *cobra.Command, args []string) {
		configPath := getConfigPath(cmd)
		contextName, _ := cmd.Flags().GetString("contextName")
		
		// 创建集群客户端
		client, err := cluster.NewClient(configPath, contextName)
		if err != nil {
			fmt.Printf("创建集群客户端失败: %v\n", err)
			os.Exit(1)
		}
		
		// 获取Service列表
		var serviceList *corev1.ServiceList
		var listErr error
		
		if allNamespaces {
			serviceList, listErr = client.Clientset.CoreV1().Services("").List(context.TODO(), metav1.ListOptions{})
		} else {
			// 如果未指定命名空间，则使用"default"
			if namespace == "" {
				namespace = "default"
			}
			serviceList, listErr = client.Clientset.CoreV1().Services(namespace).List(context.TODO(), metav1.ListOptions{})
		}
		
		if listErr != nil {
			fmt.Printf("获取Service列表失败: %v\n", listErr)
			os.Exit(1)
		}
		
		// 显示Service列表
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAMESPACE\tNAME\tTYPE\tCLUSTER-IP\tEXTERNAL-IP\tPORTS\tAGE")
		
		for _, svc := range serviceList.Items {
			// 计算Service存在时间
			age := formatAge(time.Since(svc.CreationTimestamp.Time))
			
			// 获取外部IP
			externalIP := "<none>"
			if len(svc.Status.LoadBalancer.Ingress) > 0 {
				externalIP = svc.Status.LoadBalancer.Ingress[0].IP
				if externalIP == "" {
					externalIP = svc.Status.LoadBalancer.Ingress[0].Hostname
				}
			}
			
			// 获取端口信息
			ports := ""
			for i, port := range svc.Spec.Ports {
				if i > 0 {
					ports += ", "
				}
				ports += fmt.Sprintf("%d/%s", port.Port, port.Protocol)
			}
			
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n", 
				svc.Namespace,
				svc.Name,
				svc.Spec.Type,
				svc.Spec.ClusterIP,
				externalIP,
				ports,
				age)
		}
		w.Flush()
	},
}

// resourceGetDeploymentsCmd 表示获取Deployment资源命令
var resourceGetDeploymentsCmd = &cobra.Command{
	Use:   "deployments",
	Short: "获取Deployment资源列表",
	Long:  `获取并显示Kubernetes集群中的Deployment资源列表。`,
	Run: func(cmd *cobra.Command, args []string) {
		configPath := getConfigPath(cmd)
		contextName, _ := cmd.Flags().GetString("contextName")
		
		// 创建集群客户端
		client, err := cluster.NewClient(configPath, contextName)
		if err != nil {
			fmt.Printf("创建集群客户端失败: %v\n", err)
			os.Exit(1)
		}
		
		// 获取Deployment列表
		var deploymentList *appsv1.DeploymentList
		var listErr error
		
		if allNamespaces {
			deploymentList, listErr = client.Clientset.AppsV1().Deployments("").List(context.TODO(), metav1.ListOptions{})
		} else {
			// 如果未指定命名空间，则使用"default"
			if namespace == "" {
				namespace = "default"
			}
			deploymentList, listErr = client.Clientset.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
		}
		
		if listErr != nil {
			fmt.Printf("获取Deployment列表失败: %v\n", listErr)
			os.Exit(1)
		}
		
		// 显示Deployment列表
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAMESPACE\tNAME\tREADY\tUP-TO-DATE\tAVAILABLE\tAGE")
		
		for _, deploy := range deploymentList.Items {
			// 计算Deployment存在时间
			age := formatAge(time.Since(deploy.CreationTimestamp.Time))
			
			fmt.Fprintf(w, "%s\t%s\t%d/%d\t%d\t%d\t%s\n", 
				deploy.Namespace,
				deploy.Name,
				deploy.Status.ReadyReplicas, 
				deploy.Status.Replicas,
				deploy.Status.UpdatedReplicas,
				deploy.Status.AvailableReplicas,
				age)
		}
		w.Flush()
	},
}

// getPodStatus 获取Pod的状态
func getPodStatus(pod *corev1.Pod) string {
	if pod.Status.Phase == corev1.PodSucceeded {
		return "Completed"
	}
	if pod.Status.Phase == corev1.PodFailed {
		return "Error"
	}

	if pod.DeletionTimestamp != nil {
		return "Terminating"
	}
	
	// 检查容器状态
	for _, container := range pod.Status.ContainerStatuses {
		if container.State.Waiting != nil {
			return container.State.Waiting.Reason
		}
		if container.State.Terminated != nil {
			return container.State.Terminated.Reason
		}
	}
	
	return string(pod.Status.Phase)
}

// formatAge 格式化时间间隔为人类可读的形式
func formatAge(d time.Duration) string {
	d = d.Round(time.Second)
	
	if d.Hours() > 24*365 {
		years := int(d.Hours() / (24 * 365))
		return fmt.Sprintf("%dy", years)
	}
	if d.Hours() > 24*30 {
		months := int(d.Hours() / (24 * 30))
		return fmt.Sprintf("%dm", months)
	}
	if d.Hours() > 24 {
		days := int(d.Hours() / 24)
		return fmt.Sprintf("%dd", days)
	}
	if d.Hours() > 0 {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	if d.Minutes() > 0 {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	
	return fmt.Sprintf("%ds", int(d.Seconds()))
}

func init() {
	// 添加子命令到resource命令
	resourceCmd.AddCommand(resourceGetPodsCmd)
	resourceCmd.AddCommand(resourceGetServicesCmd)
	resourceCmd.AddCommand(resourceGetDeploymentsCmd)
	
	// 添加标志
	resourceCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "要查询的命名空间")
	resourceCmd.PersistentFlags().BoolVarP(&allNamespaces, "all-namespaces", "A", false, "是否查询所有命名空间")
	
	// 添加resource命令到根命令
	rootCmd.AddCommand(resourceCmd)
} 