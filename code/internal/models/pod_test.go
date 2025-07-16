package models

import (
	"testing"
	"time"
	
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// TestPodStructure 测试Pod结构的基本属性
func TestPodStructure(t *testing.T) {
	// 创建一个测试Pod
	pod := Pod{
		Name:      "test-pod",
		Namespace: "default",
		Phase:     corev1.PodRunning,
		Reason:    "Started",
		IP:        "10.0.0.1",
		NodeName:  "test-node",
		Labels: map[string]string{
			"app": "test",
		},
		Annotations: map[string]string{
			"description": "测试Pod",
		},
	}
	
	// 验证基本属性
	if pod.Name != "test-pod" {
		t.Errorf("期望Pod名称为 'test-pod'，实际为 '%s'", pod.Name)
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
	
	// 验证标签
	if app, exists := pod.Labels["app"]; !exists || app != "test" {
		t.Errorf("期望Pod标签 'app' 为 'test'，实际为 '%s'", app)
	}
	
	// 验证注解
	if desc, exists := pod.Annotations["description"]; !exists || desc != "测试Pod" {
		t.Errorf("期望Pod注解 'description' 为 '测试Pod'，实际为 '%s'", desc)
	}
}

// TestPodContainers 测试Pod容器相关功能
func TestPodContainers(t *testing.T) {
	// 创建测试容器
	containers := []Container{
		{
			Name:  "container-1",
			Image: "nginx:latest",
			Ready: true,
			RestartCount: 0,
			CPU: ResourceMetric{
				Allocated:   *resource.NewQuantity(100, resource.DecimalSI),
				Capacity:    *resource.NewQuantity(200, resource.DecimalSI),
				Used:        *resource.NewQuantity(50, resource.DecimalSI),
				Utilization: 50.0,
			},
			Memory: ResourceMetric{
				Allocated:   *resource.NewQuantity(256*1024*1024, resource.BinarySI),
				Capacity:    *resource.NewQuantity(512*1024*1024, resource.BinarySI),
				Used:        *resource.NewQuantity(128*1024*1024, resource.BinarySI),
				Utilization: 50.0,
			},
			HasLivenessProbe:  true,
			HasReadinessProbe: true,
		},
		{
			Name:  "container-2",
			Image: "redis:latest",
			Ready: false,
			RestartCount: 2,
			CPU: ResourceMetric{
				Allocated:   *resource.NewQuantity(200, resource.DecimalSI),
				Capacity:    *resource.NewQuantity(400, resource.DecimalSI),
				Used:        *resource.NewQuantity(300, resource.DecimalSI),
				Utilization: 75.0,
			},
			Memory: ResourceMetric{
				Allocated:   *resource.NewQuantity(512*1024*1024, resource.BinarySI),
				Capacity:    *resource.NewQuantity(1024*1024*1024, resource.BinarySI),
				Used:        *resource.NewQuantity(768*1024*1024, resource.BinarySI),
				Utilization: 75.0,
			},
			HasLivenessProbe: false,
		},
	}
	
	// 创建测试Pod
	pod := Pod{
		Name:       "test-pod",
		Namespace:  "default",
		Containers: containers,
		TotalRestarts: 2, // 容器2的重启次数
	}
	
	// 验证容器数量
	if len(pod.Containers) != 2 {
		t.Errorf("期望容器数量为 2，实际为 %d", len(pod.Containers))
	}
	
	// 验证第一个容器
	container1 := pod.Containers[0]
	if container1.Name != "container-1" {
		t.Errorf("期望容器1名称为 'container-1'，实际为 '%s'", container1.Name)
	}
	
	if !container1.Ready {
		t.Errorf("期望容器1就绪状态为 true，实际为 %v", container1.Ready)
	}
	
	if !container1.HasLivenessProbe {
		t.Errorf("期望容器1存活探针状态为 true，实际为 %v", container1.HasLivenessProbe)
	}
	
	// 验证第二个容器
	container2 := pod.Containers[1]
	if container2.RestartCount != 2 {
		t.Errorf("期望容器2重启次数为 2，实际为 %d", container2.RestartCount)
	}
	
	if container2.HasLivenessProbe {
		t.Errorf("期望容器2存活探针状态为 false，实际为 %v", container2.HasLivenessProbe)
	}
	
	// 验证总重启次数
	if pod.TotalRestarts != 2 {
		t.Errorf("期望Pod总重启次数为 2，实际为 %d", pod.TotalRestarts)
	}
}

// TestPodStatus 测试Pod状态相关功能
func TestPodStatus(t *testing.T) {
	// 创建测试Pod状态
	podStatus := PodStatus{
		Phase:   corev1.PodRunning,
		Reason:  "Started",
		Message: "Pod已启动",
		Conditions: []PodConditionStatus{
			{
				Type:               "Ready",
				Status:             corev1.ConditionTrue,
				LastTransitionTime: time.Now(),
				Reason:             "PodReady",
				Message:            "Pod就绪",
			},
			{
				Type:               "Initialized",
				Status:             corev1.ConditionTrue,
				LastTransitionTime: time.Now(),
				Reason:             "PodInitialized",
				Message:            "Pod已初始化",
			},
		},
		Ready:      true,
		Scheduled:  true,
		Initialized: true,
	}
	
	// 验证Pod状态
	if podStatus.Phase != corev1.PodRunning {
		t.Errorf("期望Pod状态为 'Running'，实际为 '%s'", podStatus.Phase)
	}
	
	if !podStatus.Ready {
		t.Errorf("期望Pod就绪状态为 true，实际为 %v", podStatus.Ready)
	}
	
	if !podStatus.Scheduled {
		t.Errorf("期望Pod调度状态为 true，实际为 %v", podStatus.Scheduled)
	}
	
	// 验证Pod条件
	if len(podStatus.Conditions) != 2 {
		t.Errorf("期望Pod条件数量为 2，实际为 %d", len(podStatus.Conditions))
	}
	
	readyCondition := podStatus.Conditions[0]
	if readyCondition.Type != "Ready" || readyCondition.Status != corev1.ConditionTrue {
		t.Errorf("期望Ready条件状态为 True，实际为 %v", readyCondition.Status)
	}
}

// TestPodList 测试Pod列表功能
func TestPodList(t *testing.T) {
	// 创建测试Pod列表
	podList := PodList{
		Items: []Pod{
			{
				Name:      "pod-1",
				Namespace: "default",
				Phase:     corev1.PodRunning,
			},
			{
				Name:      "pod-2",
				Namespace: "default",
				Phase:     corev1.PodPending,
			},
			{
				Name:      "pod-3",
				Namespace: "kube-system",
				Phase:     corev1.PodRunning,
			},
			{
				Name:      "pod-4",
				Namespace: "default",
				Phase:     corev1.PodFailed,
			},
			{
				Name:      "pod-5",
				Namespace: "default",
				Phase:     corev1.PodSucceeded,
			},
		},
		TotalCount:     5,
		RunningCount:   2,
		PendingCount:   1,
		FailedCount:    1,
		SucceededCount: 1,
		UnknownCount:   0,
	}
	
	// 验证Pod列表
	if len(podList.Items) != 5 {
		t.Errorf("期望Pod数量为 5，实际为 %d", len(podList.Items))
	}
	
	if podList.TotalCount != 5 {
		t.Errorf("期望总Pod数量为 5，实际为 %d", podList.TotalCount)
	}
	
	if podList.RunningCount != 2 {
		t.Errorf("期望运行中Pod数量为 2，实际为 %d", podList.RunningCount)
	}
	
	if podList.PendingCount != 1 {
		t.Errorf("期望挂起Pod数量为 1，实际为 %d", podList.PendingCount)
	}
	
	if podList.FailedCount != 1 {
		t.Errorf("期望失败Pod数量为 1，实际为 %d", podList.FailedCount)
	}
	
	if podList.SucceededCount != 1 {
		t.Errorf("期望成功Pod数量为 1，实际为 %d", podList.SucceededCount)
	}
	
	if podList.UnknownCount != 0 {
		t.Errorf("期望未知状态Pod数量为 0，实际为 %d", podList.UnknownCount)
	}
} 