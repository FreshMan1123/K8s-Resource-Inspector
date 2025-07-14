package cluster

import (
	"fmt"
	"path/filepath"
	"context"

	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/models"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/metrics/pkg/client/clientset/versioned"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

// Client 表示Kubernetes集群客户端
type Client struct {
	// Clientset 是与Kubernetes API交互的客户端
	Clientset *kubernetes.Clientset
	// ConfigPath 是使用的kubeconfig文件路径
	ConfigPath string
	// ContextName 是使用的kubeconfig上下文名称
	ContextName string
	// MetricsClient 是获取指标数据的客户端
	MetricsClient *versioned.Clientset
}

// NewClient 创建一个新的Kubernetes客户端
func NewClient(configPath string, contextName string) (*Client, error) {
	// 如果未指定配置文件路径，则使用默认路径
	if configPath == "" {
		if home := homedir.HomeDir(); home != "" {
			configPath = filepath.Join(home, ".kube", "config")
		} else {
			return nil, fmt.Errorf("无法确定家目录，请明确指定kubeconfig路径")
		}
	}

	// 创建加载kubeconfig的配置
	loadingRules := &clientcmd.ClientConfigLoadingRules{
		ExplicitPath: configPath,
	}
	
	// 创建上下文覆盖配置
	overrides := &clientcmd.ConfigOverrides{}
	
	// 如果指定了上下文名称，则使用它
	if contextName != "" {
		overrides.CurrentContext = contextName
	}
	
	// 使用加载规则和覆盖配置创建clientConfig
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, overrides)
	
	// 构建rest.Config
	config, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("加载kubeconfig失败: %w", err)
	}

	// 创建clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("创建Kubernetes客户端失败: %w", err)
	}

	// 创建metrics clientset
	metricsClient, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("创建Metrics客户端失败: %w", err)
	}

	return &Client{
		Clientset: clientset,
		ConfigPath: configPath,
		ContextName: contextName,
		MetricsClient: metricsClient,
	}, nil
}

// GetServerVersion 获取Kubernetes集群版本
func (c *Client) GetServerVersion() (string, error) {
	version, err := c.Clientset.Discovery().ServerVersion()
	if err != nil {
		return "", fmt.Errorf("获取集群版本失败: %w", err)
	}
	return version.String(), nil
}

// GetCurrentContext 获取当前使用的上下文
func GetCurrentContext(configPath string) (string, error) {
	// 如果未指定配置文件路径，则使用默认路径
	if configPath == "" {
		if home := homedir.HomeDir(); home != "" {
			configPath = filepath.Join(home, ".kube", "config")
		} else {
			return "", fmt.Errorf("无法确定家目录，请明确指定kubeconfig路径")
		}
	}

	// 加载kubeconfig
	config, err := clientcmd.LoadFromFile(configPath)
	if err != nil {
		return "", fmt.Errorf("加载kubeconfig失败: %w", err)
	}

	return config.CurrentContext, nil
}

// ListContexts 列出kubeconfig中的所有上下文
func ListContexts(configPath string) ([]string, error) {
	// 如果未指定配置文件路径，则使用默认路径
	if configPath == "" {
		if home := homedir.HomeDir(); home != "" {
			configPath = filepath.Join(home, ".kube", "config")
		} else {
			return nil, fmt.Errorf("无法确定家目录，请明确指定kubeconfig路径")
		}
	}

	// 加载kubeconfig
	config, err := clientcmd.LoadFromFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("加载kubeconfig失败: %w", err)
	}

	contexts := make([]string, 0, len(config.Contexts))
	for name := range config.Contexts {
		contexts = append(contexts, name)
	}

	return contexts, nil
}

// SwitchContext 切换当前使用的上下文
func SwitchContext(configPath string, contextName string) error {
	// 如果未指定配置文件路径，则使用默认路径
	if configPath == "" {
		if home := homedir.HomeDir(); home != "" {
			configPath = filepath.Join(home, ".kube", "config")
		} else {
			return fmt.Errorf("无法确定家目录，请明确指定kubeconfig路径")
		}
	}

	// 加载kubeconfig
	config, err := clientcmd.LoadFromFile(configPath)
	if err != nil {
		return fmt.Errorf("加载kubeconfig失败: %w", err)
	}

	// 检查上下文是否存在
	if _, exists := config.Contexts[contextName]; !exists {
		return fmt.Errorf("上下文 '%s' 不存在", contextName)
	}

	// 切换上下文
	config.CurrentContext = contextName

	// 保存修改后的kubeconfig
	err = clientcmd.WriteToFile(*config, configPath)
	if err != nil {
		return fmt.Errorf("保存kubeconfig失败: %w", err)
	}

	return nil
}

// GetNode 获取单个节点的详细信息
func (c *Client) GetNode(nodeName string) (*models.Node, error) {
	ctx := context.Background()

	// 获取节点基本信息
	node, err := c.Clientset.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取节点信息失败: %w", err)
	}

	// 获取节点指标
	nodeMetrics, err := c.MetricsClient.MetricsV1beta1().NodeMetricses().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取节点指标失败: %w", err)
	}

	// 获取节点上的Pod列表
	pods, err := c.Clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("spec.nodeName=%s", nodeName),
	})
	if err != nil {
		return nil, fmt.Errorf("获取节点上的Pod列表失败: %w", err)
	}

	// 构建节点模型
	nodeModel := buildNodeModel(node, nodeMetrics, pods)

	return nodeModel, nil
}

// ListNodes 获取所有节点的详细信息
func (c *Client) ListNodes() (*models.NodeList, error) {
	ctx := context.Background()

	// 获取所有节点
	nodes, err := c.Clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取节点列表失败: %w", err)
	}

	// 获取所有节点指标
	nodeMetricsList, err := c.MetricsClient.MetricsV1beta1().NodeMetricses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取节点指标列表失败: %w", err)
	}

	// 创建节点指标映射，方便查找
	metricsMap := make(map[string]*metricsv1beta1.NodeMetrics)
	for i, metrics := range nodeMetricsList.Items {
		metricsMap[metrics.Name] = &nodeMetricsList.Items[i]
	}

	// 获取所有Pod
	allPods, err := c.Clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取Pod列表失败: %w", err)
	}

	// 按节点分组Pod
	podsByNode := make(map[string][]v1.Pod)
	for _, pod := range allPods.Items {
		nodeName := pod.Spec.NodeName
		if nodeName != "" {
			podsByNode[nodeName] = append(podsByNode[nodeName], pod)
		}
	}

	// 构建节点列表
	nodeList := &models.NodeList{
		Items:               make([]models.Node, 0, len(nodes.Items)),
		TotalCount:          len(nodes.Items),
		ReadyCount:          0,
		NotSchedulableCount: 0,
	}

	for _, node := range nodes.Items {
		// 获取节点对应的指标
		nodeMetrics, exists := metricsMap[node.Name]
		if !exists {
			// 如果没有找到指标，使用空指标
			nodeMetrics = &metricsv1beta1.NodeMetrics{}
		}

		// 获取节点上的Pod
		nodePods := podsByNode[node.Name]
		podsList := &v1.PodList{
			Items: nodePods,
		}

		// 构建节点模型
		nodeModel := buildNodeModel(&node, nodeMetrics, podsList)
		
		// 添加到列表
		nodeList.Items = append(nodeList.Items, *nodeModel)
		
		// 更新统计信息
		if nodeModel.Ready {
			nodeList.ReadyCount++
		}
		if !nodeModel.Schedulable {
			nodeList.NotSchedulableCount++
		}
	}

	return nodeList, nil
}

// buildNodeModel 从Kubernetes API返回的数据构建节点模型
func buildNodeModel(node *v1.Node, metrics *metricsv1beta1.NodeMetrics, pods *v1.PodList) *models.Node {
	// 提取节点基本信息
	name := node.Name
	ready := isNodeReady(node)
	schedulable := !node.Spec.Unschedulable
	
	// 提取节点角色
	roles := getNodeRoles(node)
	
	// 提取节点地址
	addresses := make(map[string]string)
	for _, addr := range node.Status.Addresses {
		addresses[string(addr.Type)] = addr.Address
	}
	
	// 提取节点标签
	labels := make(map[string]string)
	for k, v := range node.Labels {
		labels[k] = v
	}
	
	// 提取节点污点
	taints := node.Spec.Taints
	
	// 提取节点信息
	nodeInfo := models.NodeInfo{
		KernelVersion:          node.Status.NodeInfo.KernelVersion,
		OSImage:                node.Status.NodeInfo.OSImage,
		ContainerRuntimeVersion: node.Status.NodeInfo.ContainerRuntimeVersion,
		KubeletVersion:         node.Status.NodeInfo.KubeletVersion,
		KubeProxyVersion:       node.Status.NodeInfo.KubeProxyVersion,
		Architecture:           node.Status.NodeInfo.Architecture,
	}
	
	// 提取节点压力状态
	pressureStatus := models.NodePressureStatus{}
	for _, condition := range node.Status.Conditions {
		switch condition.Type {
		case v1.NodeMemoryPressure:
			pressureStatus.MemoryPressure = (condition.Status == v1.ConditionTrue)
		case v1.NodeDiskPressure:
			pressureStatus.DiskPressure = (condition.Status == v1.ConditionTrue)
		case v1.NodePIDPressure:
			pressureStatus.PIDPressure = (condition.Status == v1.ConditionTrue)
		case v1.NodeNetworkUnavailable:
			pressureStatus.NetworkPressure = (condition.Status == v1.ConditionTrue)
		}
	}
	
	// 计算CPU资源指标
	cpuCapacity := node.Status.Capacity.Cpu()
	cpuAllocatable := node.Status.Allocatable.Cpu()
	cpuAllocated := calculateAllocatedCPU(pods)
	cpuUsed := metrics.Usage.Cpu()
	
	cpuUtilization := calculateUtilization(cpuUsed, cpuCapacity)
	cpuAllocationRate := calculateAllocationRate(cpuAllocated, cpuAllocatable)
	
	// 计算内存资源指标
	memCapacity := node.Status.Capacity.Memory()
	memAllocatable := node.Status.Allocatable.Memory()
	memAllocated := calculateAllocatedMemory(pods)
	memUsed := metrics.Usage.Memory()
	
	memUtilization := calculateUtilization(memUsed, memCapacity)
	memAllocationRate := calculateAllocationRate(memAllocated, memAllocatable)
	
	// 计算临时存储资源指标
	storageCapacity := node.Status.Capacity.StorageEphemeral()
	storageAllocatable := node.Status.Allocatable.StorageEphemeral()
	storageAllocated := calculateAllocatedStorage(pods)
	
	// 临时存储使用量需要从其他来源获取，这里简化处理
	storageUsed := resource.NewQuantity(0, resource.BinarySI)
	storageUtilization := 0.0
	
	// 计算Pod资源指标
	podCapacity := node.Status.Capacity.Pods()
	podAllocatable := node.Status.Allocatable.Pods()
	podUsed := resource.NewQuantity(int64(len(pods.Items)), resource.DecimalSI)
	podUtilization := calculateUtilization(podUsed, podCapacity)
	podAllocationRate := calculateAllocationRate(podUsed, podAllocatable)
	
	// 提取节点条件状态
	conditions := make([]models.NodeConditionStatus, 0, len(node.Status.Conditions))
	for _, condition := range node.Status.Conditions {
		conditions = append(conditions, models.NodeConditionStatus{
			Type:               string(condition.Type),
			Status:             condition.Status,
			LastTransitionTime: condition.LastTransitionTime.Time,
			Reason:             condition.Reason,
			Message:            condition.Message,
		})
	}
	
	// 检查CPU压力
	pressureStatus.CPUPressure = (cpuUtilization > 90.0)
	
	// 构建节点模型
	nodeModel := &models.Node{
		Name:         name,
		Roles:        roles,
		Addresses:    addresses,
		CreationTime: node.CreationTimestamp.Time,
		Ready:        ready,
		Schedulable:  schedulable,
		Labels:       labels,
		Taints:       taints,
		NodeInfo:     nodeInfo,
		PressureStatus: pressureStatus,
		CPU: models.ResourceMetric{
			Capacity:       *cpuCapacity,
			Allocatable:    *cpuAllocatable,
			Allocated:      *cpuAllocated,
			Used:           *cpuUsed,
			Utilization:    cpuUtilization,
			AllocationRate: cpuAllocationRate,
		},
		Memory: models.ResourceMetric{
			Capacity:       *memCapacity,
			Allocatable:    *memAllocatable,
			Allocated:      *memAllocated,
			Used:           *memUsed,
			Utilization:    memUtilization,
			AllocationRate: memAllocationRate,
		},
		EphemeralStorage: models.ResourceMetric{
			Capacity:       *storageCapacity,
			Allocatable:    *storageAllocatable,
			Allocated:      *storageAllocated,
			Used:           *storageUsed,
			Utilization:    storageUtilization,
			AllocationRate: calculateAllocationRate(storageAllocated, storageAllocatable),
		},
		Pods: models.ResourceMetric{
			Capacity:       *podCapacity,
			Allocatable:    *podAllocatable,
			Used:           *podUsed,
			Utilization:    podUtilization,
			AllocationRate: podAllocationRate,
		},
		RunningPods: countRunningPods(pods),
		Conditions:  conditions,
	}
	
	return nodeModel
}

// isNodeReady 检查节点是否就绪
func isNodeReady(node *v1.Node) bool {
	for _, condition := range node.Status.Conditions {
		if condition.Type == v1.NodeReady {
			return condition.Status == v1.ConditionTrue
		}
	}
	return false
}

// getNodeRoles 获取节点角色
func getNodeRoles(node *v1.Node) []string {
	roles := make([]string, 0)
	
	// 检查常见的角色标签
	if _, ok := node.Labels["node-role.kubernetes.io/master"]; ok {
		roles = append(roles, "master")
	} else if _, ok := node.Labels["node-role.kubernetes.io/control-plane"]; ok {
		roles = append(roles, "master")
	}
	
	// 检查其他角色标签
	for label := range node.Labels {
		if label == "kubernetes.io/role" {
			roles = append(roles, node.Labels[label])
		} else if len(label) > 20 && label[:20] == "node-role.kubernetes" {
			role := label[21:] // 截取"node-role.kubernetes.io/"后的部分
			roles = append(roles, role)
		}
	}
	
	// 如果没有找到角色，默认为worker
	if len(roles) == 0 {
		roles = append(roles, "worker")
	}
	
	return roles
}

// calculateAllocatedCPU 计算分配给Pod的CPU资源
func calculateAllocatedCPU(pods *v1.PodList) *resource.Quantity {
	total := resource.NewQuantity(0, resource.DecimalSI)
	
	for _, pod := range pods.Items {
		if pod.Status.Phase == v1.PodRunning || pod.Status.Phase == v1.PodPending {
			for _, container := range pod.Spec.Containers {
				if container.Resources.Requests.Cpu() != nil {
					total.Add(*container.Resources.Requests.Cpu())
				}
			}
		}
	}
	
	return total
}

// calculateAllocatedMemory 计算分配给Pod的内存资源
func calculateAllocatedMemory(pods *v1.PodList) *resource.Quantity {
	total := resource.NewQuantity(0, resource.BinarySI)
	
	for _, pod := range pods.Items {
		if pod.Status.Phase == v1.PodRunning || pod.Status.Phase == v1.PodPending {
			for _, container := range pod.Spec.Containers {
				if container.Resources.Requests.Memory() != nil {
					total.Add(*container.Resources.Requests.Memory())
				}
			}
		}
	}
	
	return total
}

// calculateAllocatedStorage 计算分配给Pod的临时存储资源
func calculateAllocatedStorage(pods *v1.PodList) *resource.Quantity {
	total := resource.NewQuantity(0, resource.BinarySI)
	
	for _, pod := range pods.Items {
		if pod.Status.Phase == v1.PodRunning || pod.Status.Phase == v1.PodPending {
			for _, container := range pod.Spec.Containers {
				if container.Resources.Requests.StorageEphemeral() != nil {
					total.Add(*container.Resources.Requests.StorageEphemeral())
				}
			}
		}
	}
	
	return total
}

// countRunningPods 计算正在运行的Pod数量
func countRunningPods(pods *v1.PodList) int {
	count := 0
	
	for _, pod := range pods.Items {
		if pod.Status.Phase == v1.PodRunning {
			count++
		}
	}
	
	return count
}

// calculateUtilization 计算资源利用率
func calculateUtilization(used, capacity *resource.Quantity) float64 {
	if capacity.IsZero() {
		return 0.0
	}
	
	utilization := float64(used.MilliValue()) / float64(capacity.MilliValue()) * 100.0
	return utilization
}

// calculateAllocationRate 计算资源分配率
func calculateAllocationRate(allocated, allocatable *resource.Quantity) float64 {
	if allocatable.IsZero() {
		return 0.0
	}
	
	allocationRate := float64(allocated.MilliValue()) / float64(allocatable.MilliValue()) * 100.0
	return allocationRate
} 