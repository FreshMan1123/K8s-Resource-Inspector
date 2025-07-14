package resource

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/cluster"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewGetCommand 创建get命令
func NewGetCommand(namespace *string, allNamespaces *bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [resource-type] [name]",
		Short: "获取Kubernetes资源",
		Long:  `获取并显示Kubernetes集群中的资源信息。支持的资源类型: pods, services, deployments`,
		Args:  cobra.RangeArgs(1, 2),
		Run: func(cmd *cobra.Command, args []string) {
			resourceType := args[0]
			resourceName := ""
			if len(args) > 1 {
				resourceName = args[1]
			}

			configPath, _ := cmd.Flags().GetString("kubeconfig")
			contextName, _ := cmd.Flags().GetString("contextName")
			
			// 创建集群客户端
			client, err := cluster.NewClient(configPath, contextName)
			if err != nil {
				fmt.Printf("创建集群客户端失败: %v\n", err)
				os.Exit(1)
			}

			// 根据资源类型调用相应的处理函数
			switch resourceType {
			case "pod", "pods":
				getPods(client, resourceName, *namespace, *allNamespaces)
			case "service", "services", "svc":
				getServices(client, resourceName, *namespace, *allNamespaces)
			case "deployment", "deployments", "deploy":
				getDeployments(client, resourceName, *namespace, *allNamespaces)
			default:
				fmt.Printf("不支持的资源类型: %s\n", resourceType)
				fmt.Println("支持的资源类型: pods, services, deployments")
				os.Exit(1)
			}
		},
	}

	return cmd
}

// getPods 获取Pod资源
func getPods(client *cluster.Client, podName string, namespace string, allNamespaces bool) {
	var podList *corev1.PodList
	var err error
	
	// 确定命名空间
	ns := DetermineNamespace(allNamespaces, namespace)
	
	ctx := context.TODO()
	
	// 根据是否指定资源名称决定获取单个资源还是列表
	if podName != "" {
		// 获取单个Pod
		pod, err := client.Clientset.CoreV1().Pods(ns).Get(ctx, podName, metav1.GetOptions{})
		if err != nil {
			fmt.Printf("获取Pod失败: %v\n", err)
			os.Exit(1)
		}
		
		// 创建只包含一个Pod的列表
		podList = &corev1.PodList{
			Items: []corev1.Pod{*pod},
		}
	} else {
		// 获取Pod列表
		podList, err = client.Clientset.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{})
		if err != nil {
			fmt.Printf("获取Pod列表失败: %v\n", err)
			os.Exit(1)
		}
	}
	
	// 显示Pod信息
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAMESPACE\tNAME\tSTATUS\tAGE\tIP")
	
	for _, pod := range podList.Items {
		// 计算Pod存在时间
		age := FormatAge(time.Since(pod.CreationTimestamp.Time))
		
		// 获取Pod状态
		status := GetPodStatus(&pod)
		
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", 
			pod.Namespace,
			pod.Name,
			status,
			age,
			pod.Status.PodIP)
	}
	w.Flush()
}

// getServices 获取Service资源
func getServices(client *cluster.Client, serviceName string, namespace string, allNamespaces bool) {
	var serviceList *corev1.ServiceList
	var err error
	
	// 确定命名空间
	ns := DetermineNamespace(allNamespaces, namespace)
	
	ctx := context.TODO()
	
	// 根据是否指定资源名称决定获取单个资源还是列表
	if serviceName != "" {
		// 获取单个Service
		service, err := client.Clientset.CoreV1().Services(ns).Get(ctx, serviceName, metav1.GetOptions{})
		if err != nil {
			fmt.Printf("获取Service失败: %v\n", err)
			os.Exit(1)
		}
		
		// 创建只包含一个Service的列表
		serviceList = &corev1.ServiceList{
			Items: []corev1.Service{*service},
		}
	} else {
		// 获取Service列表
		serviceList, err = client.Clientset.CoreV1().Services(ns).List(ctx, metav1.ListOptions{})
		if err != nil {
			fmt.Printf("获取Service列表失败: %v\n", err)
			os.Exit(1)
		}
	}
	
	// 显示Service信息
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAMESPACE\tNAME\tTYPE\tCLUSTER-IP\tEXTERNAL-IP\tPORTS\tAGE")
	
	for _, svc := range serviceList.Items {
		// 计算Service存在时间
		age := FormatAge(time.Since(svc.CreationTimestamp.Time))
		
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
}

// getDeployments 获取Deployment资源
func getDeployments(client *cluster.Client, deploymentName string, namespace string, allNamespaces bool) {
	var deploymentList *appsv1.DeploymentList
	var err error
	
	// 确定命名空间
	ns := DetermineNamespace(allNamespaces, namespace)
	
	ctx := context.TODO()
	
	// 根据是否指定资源名称决定获取单个资源还是列表
	if deploymentName != "" {
		// 获取单个Deployment
		deployment, err := client.Clientset.AppsV1().Deployments(ns).Get(ctx, deploymentName, metav1.GetOptions{})
		if err != nil {
			fmt.Printf("获取Deployment失败: %v\n", err)
			os.Exit(1)
		}
		
		// 创建只包含一个Deployment的列表
		deploymentList = &appsv1.DeploymentList{
			Items: []appsv1.Deployment{*deployment},
		}
	} else {
		// 获取Deployment列表
		deploymentList, err = client.Clientset.AppsV1().Deployments(ns).List(ctx, metav1.ListOptions{})
		if err != nil {
			fmt.Printf("获取Deployment列表失败: %v\n", err)
			os.Exit(1)
		}
	}
	
	// 显示Deployment信息
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAMESPACE\tNAME\tREADY\tUP-TO-DATE\tAVAILABLE\tAGE")
	
	for _, deploy := range deploymentList.Items {
		// 计算Deployment存在时间
		age := FormatAge(time.Since(deploy.CreationTimestamp.Time))
		
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
} 