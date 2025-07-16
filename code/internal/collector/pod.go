package collector

import (
	"context"
	"fmt"
	"time"
	"strings"

	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/cluster"
	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/models"
	
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/metrics/pkg/client/clientset/versioned"
	"k8s.io/apimachinery/pkg/api/resource"
)

// PodCollector Pod数据收集器
type PodCollector struct {
	client *cluster.Client
	// 用于访问 Kubernetes Metrics API 的客户端，用于获取Pod指标
	metricsClient *versioned.Clientset
}

// NewPodCollector 创建一个新的Pod收集器
func NewPodCollector(client *cluster.Client) (*PodCollector, error) {
	// 创建metrics客户端
	metricsClient, err := versioned.NewForConfig(client.Config)
	if err != nil {
		return nil, fmt.Errorf("无法创建metrics客户端: %w", err)
	}

	return &PodCollector{
		client:       client,
		metricsClient: metricsClient,
	}, nil
}

// GetPods 获取指定命名空间中的所有Pod信息
func (pc *PodCollector) GetPods(ctx context.Context, namespace string) (*models.PodList, error) {
	// 获取Pod列表
	pods, err := pc.client.Clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取Pod列表失败: %w", err)
	}
	
	// 获取Pod指标
	var podMetrics *versioned.PodMetricsList
	podMetricsMap := make(map[string]map[string]corev1.ResourceList) // namespace/podName -> containerName -> metrics
	
	podMetrics, err = pc.metricsClient.MetricsV1beta1().PodMetricses(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		// 记录错误但继续执行，因为指标不是必需的
		fmt.Printf("警告: 获取Pod指标失败: %v\n", err)
	} else {
		// 创建指标映射，方便查找
		for _, metric := range podMetrics.Items {
			key := fmt.Sprintf("%s/%s", metric.Namespace, metric.Name)
			if _, exists := podMetricsMap[key]; !exists {
				podMetricsMap[key] = make(map[string]corev1.ResourceList)
			}
			
			for _, container := range metric.Containers {
				podMetricsMap[key][container.Name] = container.Usage
			}
		}
	}
	
	// 转换为内部Pod模型
	podList := &models.PodList{
		Items: make([]models.Pod, 0, len(pods.Items)),
	}
	
	for _, pod := range pods.Items {
		// 获取Pod相关事件
		events, err := pc.getPodEvents(ctx, pod.Namespace, pod.Name)
		if err != nil {
			// 记录错误但继续执行，因为事件不是必需的
			fmt.Printf("警告: 获取Pod %s/%s 的事件失败: %v\n", pod.Namespace, pod.Name, err)
		}
		
		// 转换Pod
		modelPod := convertPodToModel(&pod, podMetricsMap, events)
		podList.Items = append(podList.Items, modelPod)
		
		// 更新统计信息
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
	// 获取单个Pod
	pod, err := pc.client.Clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取Pod失败: %w", err)
	}
	
	// 获取Pod指标
	var podMetric *versioned.PodMetrics
	podMetricsMap := make(map[string]map[string]corev1.ResourceList) // namespace/podName -> containerName -> metrics
	
	podMetric, err = pc.metricsClient.MetricsV1beta1().PodMetricses(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		// 记录错误但继续执行，因为指标不是必需的
		fmt.Printf("警告: 获取Pod指标失败: %v\n", err)
	} else {
		key := fmt.Sprintf("%s/%s", podMetric.Namespace, podMetric.Name)
		podMetricsMap[key] = make(map[string]corev1.ResourceList)
		
		for _, container := range podMetric.Containers {
			podMetricsMap[key][container.Name] = container.Usage
		}
	}
	
	// 获取Pod相关事件
	events, err := pc.getPodEvents(ctx, namespace, name)
	if err != nil {
		// 记录错误但继续执行，因为事件不是必需的
		fmt.Printf("警告: 获取Pod事件失败: %v\n", err)
	}
	
	// 转换为内部Pod模型
	modelPod := convertPodToModel(pod, podMetricsMap, events)
	
	return &modelPod, nil
}

// GetPodLogs 获取Pod日志
func (pc *PodCollector) GetPodLogs(ctx context.Context, namespace, name string, containerName string, lines int) ([]string, error) {
	// 设置日志选项
	podLogOptions := corev1.PodLogOptions{
		Container: containerName,
		TailLines: int64Ptr(int64(lines)),
	}
	
	// 获取日志
	req := pc.client.Clientset.CoreV1().Pods(namespace).GetLogs(name, &podLogOptions)
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
func (pc *PodCollector) getPodEvents(ctx context.Context, namespace, name string) ([]models.Event, error) {
	// 创建字段选择器，只选择与指定Pod相关的事件
	fieldSelector := fmt.Sprintf("involvedObject.kind=Pod,involvedObject.name=%s,involvedObject.namespace=%s", name, namespace)
	
	// 获取事件
	events, err := pc.client.Clientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
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

// int64Ptr 返回int64指针
func int64Ptr(i int64) *int64 {
	return &i
} 

// getPodPriority 获取Pod优先级
func getPodPriority(pod *corev1.Pod) int32 {
	if pod.Spec.Priority != nil {
		return *pod.Spec.Priority
	}
	return 0
} 