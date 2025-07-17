package collector

import (
	"context"
	"fmt"
	"time"

	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/cluster"
	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/models"
	
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// PodCollector Pod数据收集器
// 只保留 cluster.Client 引用
// 移除 metricsClient 字段
type PodCollector struct {
	client *cluster.Client
}

// NewPodCollector 创建一个新的Pod收集器
func NewPodCollector(client *cluster.Client) (*PodCollector, error) {
	return &PodCollector{
		client: client,
	}, nil
}

// GetPods 获取指定命名空间中的所有Pod信息
func (pc *PodCollector) GetPods(ctx context.Context, namespace string) (*models.PodList, error) {
	// 通过 cluster 层接口获取 Pod 列表
	pods, err := pc.client.ListRawPods(ctx, namespace)
	if err != nil {
		return nil, fmt.Errorf("获取Pod列表失败: %w", err)
	}

	// 通过 cluster 层接口获取 Pod 指标
	podMetricsList, err := pc.client.ListRawPodMetrics(ctx, namespace)
	podMetricsMap := make(map[string]map[string]corev1.ResourceList) // namespace/podName -> containerName -> metrics
	if err != nil {
		fmt.Printf("警告: 获取Pod指标失败: %v\n", err)
	} else {
		for _, metric := range podMetricsList {
			key := fmt.Sprintf("%s/%s", metric.Namespace, metric.Name)
			if _, exists := podMetricsMap[key]; !exists {
				podMetricsMap[key] = make(map[string]corev1.ResourceList)
			}
			for _, container := range metric.Containers {
				podMetricsMap[key][container.Name] = container.Usage
			}
		}
	}

	podList := &models.PodList{
		Items: make([]models.Pod, 0, len(pods)),
	}

	for _, pod := range pods {
		// 通过 cluster 层接口获取事件
		events, err := pc.client.GetRawPodEvents(ctx, pod.Namespace, pod.Name)
		modelEvents := make([]models.Event, 0, len(events))
		if err != nil {
			fmt.Printf("警告: 获取Pod %s/%s 的事件失败: %v\n", pod.Namespace, pod.Name, err)
		} else {
			for _, event := range events {
				modelEvents = append(modelEvents, models.Event{
					Type:    string(event.Type),
					Reason:  event.Reason,
					Message: event.Message,
					Time:    event.LastTimestamp.Time,
					Count:   int(event.Count),
				})
			}
		}
		modelPod := convertPodToModel(&pod, podMetricsMap, modelEvents)
		podList.Items = append(podList.Items, modelPod)
		podList.TotalCount++
		switch pod.Status.Phase {
		case corev1.PodRunning:
			podList.RunningCount++
		case corev1.PodFailed:
			podList.FailedCount++
		case corev1.PodPending:
			podList.PendingCount++
		case corev1.PodSucceeded:
			podList.SucceededCount++
		default:
			podList.UnknownCount++
		}
	}
	return podList, nil
}

// GetPod 获取单个Pod信息
func (pc *PodCollector) GetPod(ctx context.Context, namespace, name string) (*models.Pod, error) {
	pod, err := pc.client.GetRawPod(ctx, namespace, name)
	if err != nil {
		return nil, fmt.Errorf("获取Pod失败: %w", err)
	}
	podMetric, err := pc.client.GetRawPodMetrics(ctx, namespace, name)
	podMetricsMap := make(map[string]map[string]corev1.ResourceList)
	if err != nil {
		fmt.Printf("警告: 获取Pod指标失败: %v\n", err)
	} else if podMetric != nil {
		key := fmt.Sprintf("%s/%s", podMetric.Namespace, podMetric.Name)
		podMetricsMap[key] = make(map[string]corev1.ResourceList)
		for _, container := range podMetric.Containers {
			podMetricsMap[key][container.Name] = container.Usage
		}
	}
	// 通过 cluster 层接口获取事件
	modelEvents := []models.Event{}
	events, err := pc.client.GetRawPodEvents(ctx, namespace, name)
	if err != nil {
		fmt.Printf("警告: 获取Pod事件失败: %v\n", err)
	} else {
		for _, event := range events {
			modelEvents = append(modelEvents, models.Event{
				Type:    string(event.Type),
				Reason:  event.Reason,
				Message: event.Message,
				Time:    event.LastTimestamp.Time,
				Count:   int(event.Count),
			})
		}
	}
	modelPod := convertPodToModel(pod, podMetricsMap, modelEvents)
	return &modelPod, nil
}

// GetPodLogs 获取Pod日志
func (pc *PodCollector) GetPodLogs(ctx context.Context, namespace, name string, containerName string, lines int) ([]string, error) {
	return pc.client.GetRawPodLogs(ctx, namespace, name, containerName, lines)
}

// convertPodToModel 将Kubernetes Pod转换为内部Pod模型
func convertPodToModel(pod *corev1.Pod, metricsMap map[string]map[string]corev1.ResourceList, events []models.Event) models.Pod {
	// 计算总重启次数
	totalRestarts := 0
	for _, containerStatus := range pod.Status.ContainerStatuses {
		totalRestarts += int(containerStatus.RestartCount)
	}
	
	// 计算运行时长
	var runningDuration time.Duration
	if pod.Status.Phase == corev1.PodRunning {
		for _, condition := range pod.Status.Conditions {
			if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
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
		if condition.Type == corev1.PodScheduled && condition.Status == corev1.ConditionTrue {
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
	modelPod.Containers = convertContainers(pod, pod.Status.ContainerStatuses, metricsMap, false)
	modelPod.InitContainers = convertContainers(pod, pod.Status.InitContainerStatuses, metricsMap, true)
	
	return modelPod
}

// convertContainers 转换容器列表
func convertContainers(pod *corev1.Pod, containerStatuses []corev1.ContainerStatus, metricsMap map[string]map[string]corev1.ResourceList, isInit bool) []models.Container {
	containers := make([]models.Container, 0, len(containerStatuses))
	
	// 创建容器规格映射，用于获取资源请求和限制
	containerSpecMap := make(map[string]corev1.Container)
	var containerSpecs []corev1.Container
	
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
		key := fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)
		if containerMetrics, exists := metricsMap[key]; exists {
			if metrics, exists := containerMetrics[status.Name]; exists {
				// CPU使用情况
				if cpuUsage, exists := metrics[corev1.ResourceCPU]; exists {
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
					
					container.CPU = calculateResourceMetric(
						cpuLimitValue,
						cpuLimitValue,
						cpuRequestValue,
						cpuUsage,
					)
				}
				
				// 内存使用情况
				if memoryUsage, exists := metrics[corev1.ResourceMemory]; exists {
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
					
					container.Memory = calculateResourceMetric(
						memoryLimitValue,
						memoryLimitValue,
						memoryRequestValue,
						memoryUsage,
					)
				}
			}
		}
		
		containers = append(containers, container)
	}
	
	return containers
}

// getPodPriority 获取Pod优先级
func getPodPriority(pod *corev1.Pod) int32 {
	if pod.Spec.Priority != nil {
		return *pod.Spec.Priority
	}
	return 0
} 