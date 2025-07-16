package cluster

import (
	"fmt"
	"path/filepath"
	"context"
	"strings"
	"time"

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

	// 获取节点上的Pod列表（所有命名空间）
	pods, err := c.Clientset.CoreV1().Pods(metav1.NamespaceAll).List(ctx, metav1.ListOptions{
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

	// 获取所有Pod（所有命名空间）
	allPods, err := c.Clientset.CoreV1().Pods(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
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
		TotalPods:   len(pods.Items), // 设置总Pod数量
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

// GetPod 获取单个Pod的详细信息
func (c *Client) GetPod(namespace, podName string) (*models.Pod, error) {
	ctx := context.Background()

	// 获取Pod基本信息
	pod, err := c.Clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取Pod信息失败: %w", err)
	}

	// 获取Pod指标
	var podMetrics *metricsv1beta1.PodMetrics
	var podMetricsMap map[string]v1.ResourceList // containerName -> metrics
	
	podMetrics, err = c.MetricsClient.MetricsV1beta1().PodMetricses(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		// 记录错误但继续执行，因为指标不是必需的
		fmt.Printf("警告: 获取Pod指标失败: %v\n", err)
		podMetricsMap = make(map[string]v1.ResourceList)
	} else {
		podMetricsMap = make(map[string]v1.ResourceList)
		for _, container := range podMetrics.Containers {
			podMetricsMap[container.Name] = container.Usage
		}
	}

	// 获取Pod相关事件
	events, err := c.getPodEvents(ctx, namespace, podName)
	if err != nil {
		// 记录错误但继续执行，因为事件不是必需的
		fmt.Printf("警告: 获取Pod事件失败: %v\n", err)
		events = []models.Event{}
	}

	// 构建Pod模型
	podModel := buildPodModel(pod, podMetricsMap, events)

	return &podModel, nil
}

// ListPods 获取指定命名空间中的所有Pod
func (c *Client) ListPods(namespace string) (*models.PodList, error) {
	ctx := context.Background()

	// 获取Pod列表
	pods, err := c.Clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取Pod列表失败: %w", err)
	}

	// 获取Pod指标
	var podMetrics *metricsv1beta1.PodMetricsList
	podMetricsMap := make(map[string]map[string]v1.ResourceList) // podName -> containerName -> metrics
	
	podMetrics, err = c.MetricsClient.MetricsV1beta1().PodMetricses(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		// 记录错误但继续执行，因为指标不是必需的
		fmt.Printf("警告: 获取Pod指标失败: %v\n", err)
	} else {
		for _, metric := range podMetrics.Items {
			if _, exists := podMetricsMap[metric.Name]; !exists {
				podMetricsMap[metric.Name] = make(map[string]v1.ResourceList)
			}
			
			for _, container := range metric.Containers {
				podMetricsMap[metric.Name][container.Name] = container.Usage
			}
		}
	}

	// 构建Pod列表模型
	podList := &models.PodList{
		Items: make([]models.Pod, 0, len(pods.Items)),
	}

	for _, pod := range pods.Items {
		// 获取Pod相关事件
		events, err := c.getPodEvents(ctx, pod.Namespace, pod.Name)
		if err != nil {
			// 记录错误但继续执行，因为事件不是必需的
			fmt.Printf("警告: 获取Pod %s/%s 的事件失败: %v\n", pod.Namespace, pod.Name, err)
			events = []models.Event{}
		}

		// 构建Pod模型
		podModel := buildPodModel(&pod, podMetricsMap[pod.Name], events)
		podList.Items = append(podList.Items, podModel)

		// 更新统计信息
		podList.TotalCount++
		switch pod.Status.Phase {
		case v1.PodRunning:
			podList.RunningCount++
		case v1.PodFailed:
			podList.FailedCount++
		case v1.PodPending:
			podList.PendingCount++
		case v1.PodSucceeded:
			podList.SucceededCount++
		default:
			podList.UnknownCount++
		}
	}

	return podList, nil
}

// GetPodLogs 获取Pod日志
func (c *Client) GetPodLogs(namespace, podName, containerName string, lines int) ([]string, error) {
	ctx := context.Background()

	// 设置日志选项
	podLogOptions := v1.PodLogOptions{
		Container: containerName,
		TailLines: int64Ptr(int64(lines)),
	}

	// 获取日志
	req := c.Clientset.CoreV1().Pods(namespace).GetLogs(podName, &podLogOptions)
	podLogs, err := req.Stream(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取Pod日志失败: %w", err)
	}
	defer podLogs.Close()

	// 读取日志
	buf := make([]byte, 2048)
	var logContent strings.Builder
	for {
		n, err := podLogs.Read(buf)
		if err != nil {
			break
		}
		logContent.Write(buf[:n])
	}

	// 按行分割日志
	logs := strings.Split(logContent.String(), "\n")
	if len(logs) > 0 && logs[len(logs)-1] == "" {
		logs = logs[:len(logs)-1] // 移除最后一个空行
	}

	return logs, nil
}

// getPodEvents 获取与Pod相关的事件
func (c *Client) getPodEvents(ctx context.Context, namespace, podName string) ([]models.Event, error) {
	// 创建字段选择器，只选择与指定Pod相关的事件
	fieldSelector := fmt.Sprintf("involvedObject.kind=Pod,involvedObject.name=%s,involvedObject.namespace=%s", podName, namespace)

	// 获取事件
	events, err := c.Clientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fieldSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("获取Pod事件失败: %w", err)
	}

	// 转换为内部事件模型
	modelEvents := make([]models.Event, 0, len(events.Items))
	for _, event := range events.Items {
		modelEvents = append(modelEvents, models.Event{
			Type:    string(event.Type),
			Reason:  event.Reason,
			Message: event.Message,
			Time:    event.LastTimestamp.Time,
			Count:   int(event.Count),
		})
	}

	return modelEvents, nil
}

// buildPodModel 构建Pod模型
func buildPodModel(pod *v1.Pod, containerMetrics map[string]v1.ResourceList, events []models.Event) models.Pod {
	// 计算总重启次数
	totalRestarts := 0
	for _, containerStatus := range pod.Status.ContainerStatuses {
		totalRestarts += int(containerStatus.RestartCount)
	}

	// 计算运行时长
	var runningDuration time.Duration
	if pod.Status.Phase == v1.PodRunning {
		for _, condition := range pod.Status.Conditions {
			if condition.Type == v1.PodReady && condition.Status == v1.ConditionTrue {
				runningDuration = time.Since(condition.LastTransitionTime.Time)
				break
			}
		}
	}

	// 检查探针配置
	hasReadinessProbe := false
	hasLivenessProbe := false
	hasStartupProbe := false

	for _, container := range pod.Spec.Containers {
		if container.ReadinessProbe != nil {
			hasReadinessProbe = true
		}
		if container.LivenessProbe != nil {
			hasLivenessProbe = true
		}
		if container.StartupProbe != nil {
			hasStartupProbe = true
		}
	}

	// 获取Pod调度时间
	var scheduledTime *time.Time
	for _, condition := range pod.Status.Conditions {
		if condition.Type == v1.PodScheduled && condition.Status == v1.ConditionTrue {
			scheduledTime = &condition.LastTransitionTime.Time
			break
		}
	}

	// 创建Pod模型
	modelPod := models.Pod{
		Name:              pod.Name,
		Namespace:         pod.Namespace,
		Phase:             pod.Status.Phase,
		Reason:            pod.Status.Reason,
		CreationTime:      pod.CreationTimestamp.Time,
		IP:                pod.Status.PodIP,
		NodeName:          pod.Spec.NodeName,
		Labels:            pod.Labels,
		Annotations:       pod.Annotations,
		TotalRestarts:     totalRestarts,
		RunningDuration:   runningDuration,
		Events:            events,
		RecentLogs:        make(map[string][]string),
		HasReadinessProbe: hasReadinessProbe,
		HasLivenessProbe:  hasLivenessProbe,
		HasStartupProbe:   hasStartupProbe,
		QOSClass:          pod.Status.QOSClass,
		Priority:          getPodPriority(pod),
		ScheduledTime:     scheduledTime,
	}

	// 转换容器状态
	modelPod.Containers = buildContainers(pod, pod.Status.ContainerStatuses, containerMetrics, false)
	modelPod.InitContainers = buildContainers(pod, pod.Status.InitContainerStatuses, containerMetrics, true)

	return modelPod
}

// buildContainers 构建容器模型列表
func buildContainers(pod *v1.Pod, containerStatuses []v1.ContainerStatus, containerMetrics map[string]v1.ResourceList, isInit bool) []models.Container {
	containers := make([]models.Container, 0, len(containerStatuses))

	// 创建容器规格映射，用于获取资源请求和限制
	containerSpecMap := make(map[string]v1.Container)
	var containerSpecs []v1.Container

	if isInit {
		containerSpecs = pod.Spec.InitContainers
	} else {
		containerSpecs = pod.Spec.Containers
	}

	for _, container := range containerSpecs {
		containerSpecMap[container.Name] = container
	}

	// 转换容器状态
	for _, status := range containerStatuses {
		container := models.Container{
			Name:         status.Name,
			Image:        status.Image,
			State:        status.State,
			Ready:        status.Ready,
			RestartCount: int(status.RestartCount),
		}

		// 设置资源请求和限制
		if spec, exists := containerSpecMap[status.Name]; exists {
			container.Requests = spec.Resources.Requests
			container.Limits = spec.Resources.Limits

			// 检查探针配置
			container.HasReadinessProbe = spec.ReadinessProbe != nil
			container.HasLivenessProbe = spec.LivenessProbe != nil
			container.HasStartupProbe = spec.StartupProbe != nil
		}

		// 设置资源使用情况
		if metrics, exists := containerMetrics[status.Name]; exists {
			// CPU使用情况
			if cpuUsage, exists := metrics[v1.ResourceCPU]; exists {
				cpuLimit := container.Limits.Cpu()
				cpuRequest := container.Requests.Cpu()

				// 处理可能为nil的情况
				var cpuLimitValue, cpuRequestValue resource.Quantity
				if cpuLimit != nil {
					cpuLimitValue = *cpuLimit
				}
				if cpuRequest != nil {
					cpuRequestValue = *cpuRequest
				}

				container.CPU = models.ResourceMetric{
					Used: cpuUsage,
				}

				// 计算利用率
				if !cpuLimitValue.IsZero() {
					container.CPU.Utilization = float64(cpuUsage.MilliValue()) / float64(cpuLimitValue.MilliValue()) * 100
				} else if !cpuRequestValue.IsZero() {
					container.CPU.Utilization = float64(cpuUsage.MilliValue()) / float64(cpuRequestValue.MilliValue()) * 100
				}
			}

			// 内存使用情况
			if memoryUsage, exists := metrics[v1.ResourceMemory]; exists {
				memoryLimit := container.Limits.Memory()
				memoryRequest := container.Requests.Memory()

				// 处理可能为nil的情况
				var memoryLimitValue, memoryRequestValue resource.Quantity
				if memoryLimit != nil {
					memoryLimitValue = *memoryLimit
				}
				if memoryRequest != nil {
					memoryRequestValue = *memoryRequest
				}

				container.Memory = models.ResourceMetric{
					Used: memoryUsage,
				}

				// 计算利用率
				if !memoryLimitValue.IsZero() {
					container.Memory.Utilization = float64(memoryUsage.Value()) / float64(memoryLimitValue.Value()) * 100
				} else if !memoryRequestValue.IsZero() {
					container.Memory.Utilization = float64(memoryUsage.Value()) / float64(memoryRequestValue.Value()) * 100
				}
			}
		}

		containers = append(containers, container)
	}

	return containers
}

// getPodPriority 获取Pod优先级
func getPodPriority(pod *v1.Pod) int32 {
	if pod.Spec.Priority != nil {
		return *pod.Spec.Priority
	}
	return 0
}

// int64Ptr 返回int64指针
func int64Ptr(i int64) *int64 {
	return &i
} 