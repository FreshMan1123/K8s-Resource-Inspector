package collector

import (
	"context"
	"testing"
	"time"
	
	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/cluster"
	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/models"
	
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	metricsfake "k8s.io/metrics/pkg/client/clientset/versioned/fake"
)

// MockPodCollector 创建用于测试的Pod收集器
type MockPodCollector struct {
	PodCollector
}

// 创建一个测试用的Pod
func createTestPod(name string, namespace string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			CreationTimestamp: metav1.Time{
				Time: time.Now().Add(-1 * time.Hour),
			},
			Labels: map[string]string{
				"app": "test",
			},
			Annotations: map[string]string{
				"description": "测试Pod",
			},
		},
		Spec: corev1.PodSpec{
			NodeName: "test-node",
			Containers: []corev1.Container{
				{
					Name:  "container-1",
					Image: "nginx:latest",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("100m"),
							corev1.ResourceMemory: resource.MustParse("256Mi"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("200m"),
							corev1.ResourceMemory: resource.MustParse("512Mi"),
						},
					},
					LivenessProbe:  &corev1.Probe{},
					ReadinessProbe: &corev1.Probe{},
				},
				{
					Name:  "container-2",
					Image: "redis:latest",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("200m"),
							corev1.ResourceMemory: resource.MustParse("512Mi"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("400m"),
							corev1.ResourceMemory: resource.MustParse("1Gi"),
						},
					},
				},
			},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			PodIP: "10.0.0.1",
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name:  "container-1",
					Ready: true,
					State: corev1.ContainerState{
						Running: &corev1.ContainerStateRunning{
							StartedAt: metav1.Time{
								Time: time.Now().Add(-30 * time.Minute),
							},
						},
					},
					RestartCount: 0,
				},
				{
					Name:  "container-2",
					Ready: false,
					State: corev1.ContainerState{
						Running: &corev1.ContainerStateRunning{
							StartedAt: metav1.Time{
								Time: time.Now().Add(-15 * time.Minute),
							},
						},
					},
					RestartCount: 2,
				},
			},
			QOSClass: corev1.PodQOSBurstable,
		},
	}
}

// 创建测试用的Pod指标
func createTestPodMetrics(name string, namespace string) *metricsv1beta1.PodMetrics {
	return &metricsv1beta1.PodMetrics{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Containers: []metricsv1beta1.ContainerMetrics{
			{
				Name: "container-1",
				Usage: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("50m"),
					corev1.ResourceMemory: resource.MustParse("128Mi"),
				},
			},
			{
				Name: "container-2",
				Usage: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("150m"),
					corev1.ResourceMemory: resource.MustParse("384Mi"),
				},
			},
		},
	}
}

// 创建测试用的事件
func createTestEvents(podName string, namespace string) []corev1.Event {
	return []corev1.Event{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "event-1",
				Namespace: namespace,
			},
			InvolvedObject: corev1.ObjectReference{
				Kind:      "Pod",
				Name:      podName,
				Namespace: namespace,
			},
			Type:    "Normal",
			Reason:  "Started",
			Message: "Pod已启动",
			Count:   1,
			LastTimestamp: metav1.Time{
				Time: time.Now().Add(-20 * time.Minute),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "event-2",
				Namespace: namespace,
			},
			InvolvedObject: corev1.ObjectReference{
				Kind:      "Pod",
				Name:      podName,
				Namespace: namespace,
			},
			Type:    "Warning",
			Reason:  "Unhealthy",
			Message: "容器重启",
			Count:   2,
			LastTimestamp: metav1.Time{
				Time: time.Now().Add(-10 * time.Minute),
			},
		},
	}
}

// 创建一个Mock的Pod收集器
func createMockPodCollector() (*MockPodCollector, error) {
	// 创建测试Pod
	pod1 := createTestPod("test-pod-1", "default")
	pod2 := createTestPod("test-pod-2", "default")
	pod3 := createTestPod("test-pod-3", "kube-system")
	
	// 创建测试Pod指标
	metrics1 := createTestPodMetrics("test-pod-1", "default")
	metrics2 := createTestPodMetrics("test-pod-2", "default")
	metrics3 := createTestPodMetrics("test-pod-3", "kube-system")
	
	// 创建测试事件
	events1 := createTestEvents("test-pod-1", "default")
	events2 := createTestEvents("test-pod-2", "default")
	events3 := createTestEvents("test-pod-3", "kube-system")
	
	// 创建假的Kubernetes客户端
	clientset := fake.NewSimpleClientset(
		pod1, pod2, pod3,
		&events1[0], &events1[1],
		&events2[0], &events2[1],
		&events3[0], &events3[1],
	)
	
	// 创建假的Metrics客户端
	metricsClient := metricsfake.NewSimpleClientset(metrics1, metrics2, metrics3)
	
	// 创建Mock收集器
	collector := &MockPodCollector{}
	collector.client = &cluster.Client{
		Clientset: clientset,
	}
	collector.metricsClient = metricsClient.MetricsV1beta1().RESTClient()
	
	return collector, nil
}

// TestGetPod 测试获取单个Pod信息
func TestGetPod(t *testing.T) {
	// 创建Mock收集器
	collector, err := createMockPodCollector()
	if err != nil {
		t.Fatalf("创建Mock收集器失败: %v", err)
	}
	
	// 获取Pod信息
	ctx := context.Background()
	pod, err := collector.GetPod(ctx, "default", "test-pod-1")
	
	// 验证结果
	if err != nil {
		t.Errorf("获取Pod信息失败: %v", err)
	}
	
	if pod == nil {
		t.Fatal("获取的Pod为空")
	}
	
	// 验证基本信息
	if pod.Name != "test-pod-1" {
		t.Errorf("期望Pod名称为 'test-pod-1'，实际为 '%s'", pod.Name)
	}
	
	if pod.Namespace != "default" {
		t.Errorf("期望Pod命名空间为 'default'，实际为 '%s'", pod.Namespace)
	}
	
	if pod.Phase != corev1.PodRunning {
		t.Errorf("期望Pod状态为 'Running'，实际为 '%s'", pod.Phase)
	}
	
	if pod.IP != "10.0.0.1" {
		t.Errorf("期望Pod IP为 '10.0.0.1'，实际为 '%s'", pod.IP)
	}
	
	if pod.NodeName != "test-node" {
		t.Errorf("期望Pod所在节点为 'test-node'，实际为 '%s'", pod.NodeName)
	}
	
	// 验证容器
	if len(pod.Containers) != 2 {
		t.Errorf("期望容器数量为 2，实际为 %d", len(pod.Containers))
	}
	
	// 验证容器1
	container1 := pod.Containers[0]
	if container1.Name != "container-1" {
		t.Errorf("期望容器1名称为 'container-1'，实际为 '%s'", container1.Name)
	}
	
	// 验证总重启次数
	if pod.TotalRestarts != 2 { // container-2 的重启次数
		t.Errorf("期望Pod总重启次数为 2，实际为 %d", pod.TotalRestarts)
	}
}

// TestGetPods 测试获取Pod列表
func TestGetPods(t *testing.T) {
	// 创建Mock收集器
	collector, err := createMockPodCollector()
	if err != nil {
		t.Fatalf("创建Mock收集器失败: %v", err)
	}
	
	// 获取默认命名空间的Pod列表
	ctx := context.Background()
	podList, err := collector.GetPods(ctx, "default")
	
	// 验证结果
	if err != nil {
		t.Errorf("获取Pod列表失败: %v", err)
	}
	
	if podList == nil {
		t.Fatal("获取的Pod列表为空")
	}
	
	// 验证Pod数量
	if len(podList.Items) != 2 { // default命名空间有2个Pod
		t.Errorf("期望Pod数量为 2，实际为 %d", len(podList.Items))
	}
	
	// 验证运行中的Pod数量
	if podList.RunningCount != 2 {
		t.Errorf("期望运行中Pod数量为 2，实际为 %d", podList.RunningCount)
	}
	
	// 获取所有命名空间的Pod列表
	podList, err = collector.GetPods(ctx, "")
	
	// 验证结果
	if err != nil {
		t.Errorf("获取所有命名空间Pod列表失败: %v", err)
	}
	
	// 验证Pod数量
	if len(podList.Items) != 3 { // 总共有3个Pod
		t.Errorf("期望Pod总数量为 3，实际为 %d", len(podList.Items))
	}
}

// TestGetPodLogs 测试获取Pod日志
func TestGetPodLogs(t *testing.T) {
	// 这个测试需要模拟Pod日志流，比较复杂，可以先跳过
	// 或者使用更高级的模拟技术
	t.Skip("跳过Pod日志测试，需要更复杂的模拟")
} 