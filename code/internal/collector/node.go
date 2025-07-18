package collector

import (
	"context"
	"fmt"
	"strings"

	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/cluster"
	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/models"
	
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/metrics/pkg/client/clientset/versioned"
)

// NodeCollector 接口，统一节点采集入口
// 便于依赖注入和 mock
//
type NodeCollector interface {
	GetNodes(ctx context.Context) (*models.NodeList, error)
	GetNode(ctx context.Context, name string) (*models.Node, error)
}

// nodeCollectorImpl 节点数据收集器实现
type nodeCollectorImpl struct {
	client *cluster.Client
	//用于访问 Kubernetes Metrics API 的客户端，用于获取节点指标
	metricsClient *versioned.Clientset
}

// NewNodeCollector 创建一个新的节点收集器
func NewNodeCollector(client *cluster.Client) (NodeCollector, error) {
	// 创建metrics客户端
	metricsClient, err := versioned.NewForConfig(client.Config)
	if err != nil {
		return nil, fmt.Errorf("无法创建metrics客户端: %w", err)
	}

	return &nodeCollectorImpl{
		client: client,
		metricsClient: metricsClient,
	}, nil
}

// GetNodes 获取所有节点信息
func (nc *nodeCollectorImpl) GetNodes(ctx context.Context) (*models.NodeList, error) {
	// 通过 cluster 层获取原生 Node
	nodes, err := nc.client.ListRawNodes(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取节点列表失败: %w", err)
	}
	// 通过 cluster 层获取原生 Node metrics
	nodeMetricsList, err := nc.client.ListRawNodeMetrics(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取节点指标失败: %w", err)
	}
	metricsMap := make(map[string]corev1.ResourceList)
	for _, metric := range nodeMetricsList {
		metricsMap[metric.Name] = metric.Usage
	}
	// 通过 cluster 层获取所有 Pod
	pods, err := nc.client.ListRawPods(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("获取Pod列表失败: %w", err)
	}
	
	// 计算每个节点上已分配的资源
	nodeAllocatedResources := make(map[string]map[corev1.ResourceName]resource.Quantity)
	// 计算每个节点上的Pod数量
	nodeTotalPods := make(map[string]int)
	
	for _, pod := range pods {
		nodeName := pod.Spec.NodeName
		if nodeName == "" {
			continue
		}
		
		// 统计每个节点上的总Pod数量
		if _, exists := nodeTotalPods[nodeName]; !exists {
			nodeTotalPods[nodeName] = 0
		}
		nodeTotalPods[nodeName]++
		
		// 忽略已完成的Pod进行资源计算
		if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
			continue
		}
		
		// 初始化节点资源映射
		if _, exists := nodeAllocatedResources[nodeName]; !exists {
			nodeAllocatedResources[nodeName] = make(map[corev1.ResourceName]resource.Quantity)
			nodeAllocatedResources[nodeName][corev1.ResourceCPU] = resource.Quantity{}
			nodeAllocatedResources[nodeName][corev1.ResourceMemory] = resource.Quantity{}
			nodeAllocatedResources[nodeName][corev1.ResourceEphemeralStorage] = resource.Quantity{}
			nodeAllocatedResources[nodeName]["pods"] = resource.Quantity{}
		}
		
		// 累加Pod请求的资源
		for _, container := range pod.Spec.Containers {
			if cpu, ok := container.Resources.Requests[corev1.ResourceCPU]; ok {
				cpuQuant := nodeAllocatedResources[nodeName][corev1.ResourceCPU]
				cpuQuant.Add(cpu)
				nodeAllocatedResources[nodeName][corev1.ResourceCPU] = cpuQuant
			}
			
			if memory, ok := container.Resources.Requests[corev1.ResourceMemory]; ok {
				memoryQuant := nodeAllocatedResources[nodeName][corev1.ResourceMemory]
				memoryQuant.Add(memory)
				nodeAllocatedResources[nodeName][corev1.ResourceMemory] = memoryQuant
			}
			
			if storage, ok := container.Resources.Requests[corev1.ResourceEphemeralStorage]; ok {
				storageQuant := nodeAllocatedResources[nodeName][corev1.ResourceEphemeralStorage]
				storageQuant.Add(storage)
				nodeAllocatedResources[nodeName][corev1.ResourceEphemeralStorage] = storageQuant
			}
		}
		
		// 增加Pod计数
		podsQuant := nodeAllocatedResources[nodeName]["pods"]
		podsQuant.Add(*resource.NewQuantity(1, resource.DecimalSI))
		nodeAllocatedResources[nodeName]["pods"] = podsQuant
	}
	
	// 转换为内部节点模型
	nodeList := &models.NodeList{
		Items: make([]models.Node, 0, len(nodes)),
	}
	
	for _, node := range nodes {
		modelNode := convertNodeToModel(&node, metricsMap[node.Name], nodeAllocatedResources[node.Name])
		// 设置总Pod数量
		modelNode.TotalPods = nodeTotalPods[node.Name]
		nodeList.Items = append(nodeList.Items, modelNode)
		
		// 更新统计信息
		nodeList.TotalCount++
		if modelNode.Ready {
			nodeList.ReadyCount++
		}
		if !modelNode.Schedulable {
			nodeList.NotSchedulableCount++
		}
	}
	
	return nodeList, nil
}

// GetNode 获取单个节点信息
func (nc *nodeCollectorImpl) GetNode(ctx context.Context, name string) (*models.Node, error) {
	// 通过 cluster 层获取单个 Node
	node, err := nc.client.GetRawNode(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("获取节点失败: %w", err)
	}
	// 通过 cluster 层获取 Node metrics
	nodeMetric, err := nc.client.GetRawNodeMetrics(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("获取节点指标失败: %w", err)
	}
	// 通过 cluster 层获取所有 Pod，再过滤出调度到该节点的 Pod
	pods, err := nc.client.ListRawPods(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("获取Pod列表失败: %w", err)
	}
	var nodePods []corev1.Pod
	for _, pod := range pods {
		if pod.Spec.NodeName == name {
			nodePods = append(nodePods, pod)
		}
	}
	podList := &corev1.PodList{Items: nodePods}
	// 计算已分配资源
	allocatedResources := make(map[corev1.ResourceName]resource.Quantity)
	allocatedResources[corev1.ResourceCPU] = resource.Quantity{}
	allocatedResources[corev1.ResourceMemory] = resource.Quantity{}
	allocatedResources[corev1.ResourceEphemeralStorage] = resource.Quantity{}
	allocatedResources["pods"] = resource.Quantity{}
	runningPods := 0
	totalPods := len(podList.Items)
	for _, pod := range podList.Items {
		if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
			continue
		}
		runningPods++
		for _, container := range pod.Spec.Containers {
			if cpu, ok := container.Resources.Requests[corev1.ResourceCPU]; ok {
				cpuQuant := allocatedResources[corev1.ResourceCPU]
				cpuQuant.Add(cpu)
				allocatedResources[corev1.ResourceCPU] = cpuQuant
			}
			if memory, ok := container.Resources.Requests[corev1.ResourceMemory]; ok {
				memoryQuant := allocatedResources[corev1.ResourceMemory]
				memoryQuant.Add(memory)
				allocatedResources[corev1.ResourceMemory] = memoryQuant
			}
			if storage, ok := container.Resources.Requests[corev1.ResourceEphemeralStorage]; ok {
				storageQuant := allocatedResources[corev1.ResourceEphemeralStorage]
				storageQuant.Add(storage)
				allocatedResources[corev1.ResourceEphemeralStorage] = storageQuant
			}
		}
	}
	allocatedResources["pods"] = *resource.NewQuantity(int64(runningPods), resource.DecimalSI)
	modelNode := convertNodeToModel(node, nodeMetric.Usage, allocatedResources)
	modelNode.TotalPods = totalPods
	return &modelNode, nil
}

// convertNodeToModel 将Kubernetes节点转换为内部节点模型
func convertNodeToModel(node *corev1.Node, usage corev1.ResourceList, allocated map[corev1.ResourceName]resource.Quantity) models.Node {
	// 提取节点角色
	roles := extractNodeRoles(node.Labels)
	
	// 提取节点地址
	addresses := make(map[string]string)
	for _, addr := range node.Status.Addresses {
		addresses[string(addr.Type)] = addr.Address
	}
	
	// 检查节点是否就绪
	ready := false
	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeReady {
			ready = condition.Status == corev1.ConditionTrue
			break
		}
	}
	
	// 检查节点是否可调度
	schedulable := !node.Spec.Unschedulable
	
	// 初始化Node模型
	podsQ := allocated["pods"]
	modelNode := models.Node{
		Name:         node.Name,
		Roles:        roles,
		Addresses:    addresses,
		CreationTime: node.CreationTimestamp.Time,
		Ready:        ready,
		Schedulable:  schedulable,
		Labels:       node.Labels,
		Taints:       node.Spec.Taints,
		RunningPods:  int(podsQ.AsApproximateFloat64()),
		CustomMetrics: make(map[string]models.CustomMetric),
		NodeInfo: models.NodeInfo{
			KernelVersion:           node.Status.NodeInfo.KernelVersion,
			OSImage:                 node.Status.NodeInfo.OSImage,
			ContainerRuntimeVersion: node.Status.NodeInfo.ContainerRuntimeVersion,
			KubeletVersion:          node.Status.NodeInfo.KubeletVersion,
			// KubeProxyVersion 已被废弃，不再使用
			Architecture:            node.Status.NodeInfo.Architecture,
		},
	}
	
	// 设置节点条件
	modelNode.Conditions = make([]models.NodeConditionStatus, 0, len(node.Status.Conditions))
	for _, condition := range node.Status.Conditions {
		modelNode.Conditions = append(modelNode.Conditions, models.NodeConditionStatus{
			Type:               string(condition.Type),
			Status:             condition.Status,
			LastTransitionTime: condition.LastTransitionTime.Time,
			Reason:             condition.Reason,
			Message:            condition.Message,
		})
		
		// 设置节点压力状态
		switch condition.Type {
		case corev1.NodeMemoryPressure:
			modelNode.PressureStatus.MemoryPressure = condition.Status == corev1.ConditionTrue
		case corev1.NodeDiskPressure:
			modelNode.PressureStatus.DiskPressure = condition.Status == corev1.ConditionTrue
		case corev1.NodePIDPressure:
			modelNode.PressureStatus.PIDPressure = condition.Status == corev1.ConditionTrue
		case corev1.NodeNetworkUnavailable:
			modelNode.PressureStatus.NetworkPressure = condition.Status == corev1.ConditionTrue
		}
	}
	
	// 处理CPU资源
	modelNode.CPU = calculateResourceMetric(
		node.Status.Capacity[corev1.ResourceCPU],
		node.Status.Allocatable[corev1.ResourceCPU],
		allocated[corev1.ResourceCPU],
		usage[corev1.ResourceCPU],
	)
	
	// 处理内存资源（全部转为 Mi 单位，float64 存储与展示）
	// 节点声明的物理内存总量（Mi）
	memCapacityQ := node.Status.Capacity[corev1.ResourceMemory]
	memCapacityMi := (&memCapacityQ).AsApproximateFloat64() / 1024 / 1024

	// 节点可分配给 Pod 的内存总量（Mi）
	memAllocatableQ := node.Status.Allocatable[corev1.ResourceMemory]
	memAllocatableMi := (&memAllocatableQ).AsApproximateFloat64() / 1024 / 1024

	// 当前已分配给所有 Pod 的内存总量（Mi）
	memAllocatedQ := allocated[corev1.ResourceMemory]
	memAllocatedMi := (&memAllocatedQ).AsApproximateFloat64() / 1024 / 1024

	// 节点实际使用的内存总量（Mi）
	memUsedQ := usage[corev1.ResourceMemory]
	memUsedMi := (&memUsedQ).AsApproximateFloat64() / 1024 / 1024

	modelNode.Memory = models.ResourceMetric{
		Capacity:    memCapacityMi,
		Allocatable: memAllocatableMi,
		Allocated:   memAllocatedMi,
		Used:        memUsedMi,
		Utilization: func() float64 {
			if memCapacityMi > 0 {
				return memUsedMi / memCapacityMi * 100
			}
			return 0
		}(),
		AllocationRate: func() float64 {
			if memAllocatableMi > 0 {
				return memAllocatedMi / memAllocatableMi * 100
			}
			return 0
		}(),
	}
	
	// 处理临时存储资源
	modelNode.EphemeralStorage = calculateResourceMetric(
		node.Status.Capacity[corev1.ResourceEphemeralStorage],
		node.Status.Allocatable[corev1.ResourceEphemeralStorage],
		allocated[corev1.ResourceEphemeralStorage],
		usage[corev1.ResourceEphemeralStorage],
	)
	
	// 处理Pod资源
	modelNode.Pods = calculateResourceMetric(
		node.Status.Capacity[corev1.ResourcePods],
		node.Status.Allocatable[corev1.ResourcePods],
		*resource.NewQuantity(int64(modelNode.RunningPods), resource.DecimalSI),
		*resource.NewQuantity(int64(modelNode.RunningPods), resource.DecimalSI),
	)
	
	return modelNode
}

// calculateResourceMetric 计算资源指标
func calculateResourceMetric(capacity, allocatable, allocated, used resource.Quantity) models.ResourceMetric {
	metric := models.ResourceMetric{
		Capacity:    capacity.AsApproximateFloat64(),
		Allocatable: allocatable.AsApproximateFloat64(),
		Allocated:   allocated.AsApproximateFloat64(),
		Used:        used.AsApproximateFloat64(),
	}
	
	// 计算资源利用率 (使用量/已分配量)
	if !allocated.IsZero() {
		utilization := float64(used.MilliValue()) / float64(allocated.MilliValue()) * 100
		metric.Utilization = utilization
	}
	
	// 计算资源分配率 (已分配量/可分配量)
	if !allocatable.IsZero() {
		allocationRate := float64(allocated.MilliValue()) / float64(allocatable.MilliValue()) * 100
		metric.AllocationRate = allocationRate
	}
	
	return metric
}

// extractNodeRoles 从节点标签中提取角色
func extractNodeRoles(labels map[string]string) []string {
	roles := make([]string, 0)
	
	// 检查是否是控制平面节点
	if _, isControlPlane := labels["node-role.kubernetes.io/control-plane"]; isControlPlane {
		roles = append(roles, "control-plane")
	}
	if _, isMaster := labels["node-role.kubernetes.io/master"]; isMaster {
		roles = append(roles, "master")
	}
	
	// 检查其他角色标签
	for label := range labels {
		if strings.HasPrefix(label, "node-role.kubernetes.io/") {
			role := strings.TrimPrefix(label, "node-role.kubernetes.io/")
			if role != "master" && role != "control-plane" {
				roles = append(roles, role)
			}
		}
	}
	
	// 如果没有角色，则为worker
	if len(roles) == 0 {
		roles = append(roles, "worker")
	}
	
	return roles
} 