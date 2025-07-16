package test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/analyzer/pod"
	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/models"
	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/rules"
	
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// Pod测试数据
type TestPodData struct {
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	Phase      string `json:"phase"`
	IP         string `json:"ip"`
	NodeName   string `json:"nodeName"`
	QOSClass   string `json:"qosClass"`
	Containers []struct {
		Name         string `json:"name"`
		Image        string `json:"image"`
		Ready        bool   `json:"ready"`
		RestartCount int    `json:"restartCount"`
		Resources    struct {
			Requests struct {
				CPU    string `json:"cpu"`
				Memory string `json:"memory"`
			} `json:"requests"`
			Limits struct {
				CPU    string `json:"cpu"`
				Memory string `json:"memory"`
			} `json:"limits"`
		} `json:"resources"`
		Metrics struct {
			CPUUsage    string `json:"cpuUsage"`
			MemoryUsage string `json:"memoryUsage"`
		} `json:"metrics"`
		HasProbes bool `json:"hasProbes"`
	} `json:"containers"`
	Events []struct {
		Type    string `json:"type"`
		Reason  string `json:"reason"`
		Message string `json:"message"`
		Count   int    `json:"count"`
	} `json:"events"`
	CreationTime time.Time `json:"creationTime"`
}

// MockPodClient 模拟Pod客户端
type MockPodClient struct {
	podData map[string]*TestPodData
}

// NewMockPodClient 创建模拟Pod客户端
func NewMockPodClient() *MockPodClient {
	return &MockPodClient{
		podData: make(map[string]*TestPodData),
	}
}

// LoadPodData 从文件加载Pod数据
func (c *MockPodClient) LoadPodData(key, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取Pod数据文件失败: %w", err)
	}

	podData := &TestPodData{}
	if err := json.Unmarshal(data, podData); err != nil {
		return fmt.Errorf("解析Pod数据失败: %w", err)
	}

	c.podData[key] = podData
	return nil
}

// GetPod 获取Pod数据
func (c *MockPodClient) GetPod(key string) (*models.Pod, error) {
	data, ok := c.podData[key]
	if !ok {
		return nil, fmt.Errorf("Pod %s 不存在", key)
	}

	// 创建Pod模型
	pod := &models.Pod{
		Name:         data.Name,
		Namespace:    data.Namespace,
		Phase:        corev1.PodPhase(data.Phase),
		IP:           data.IP,
		NodeName:     data.NodeName,
		QOSClass:     corev1.PodQOSClass(data.QOSClass),
		CreationTime: data.CreationTime,
		RunningDuration: time.Since(data.CreationTime),
		Containers:   make([]models.Container, 0, len(data.Containers)),
		TotalRestarts: 0,
		Events:       make([]models.Event, 0, len(data.Events)),
	}

	// 添加容器
	for _, c := range data.Containers {
		cpuLimits, _ := resource.ParseQuantity(c.Resources.Limits.CPU)
		cpuUsage, _ := resource.ParseQuantity(c.Metrics.CPUUsage)
		
		memLimits, _ := resource.ParseQuantity(c.Resources.Limits.Memory)
		memUsage, _ := resource.ParseQuantity(c.Metrics.MemoryUsage)
		
		// 计算CPU利用率
		cpuUtilization := 0.0
		if cpuLimits.MilliValue() > 0 {
			cpuUtilization = float64(cpuUsage.MilliValue()) / float64(cpuLimits.MilliValue()) * 100
		}
		
		// 计算内存利用率
		memUtilization := 0.0
		if memLimits.Value() > 0 {
			memUtilization = float64(memUsage.Value()) / float64(memLimits.Value()) * 100
		}
		
		container := models.Container{
			Name:         c.Name,
			Image:        c.Image,
			Ready:        c.Ready,
			RestartCount: c.RestartCount,
			CPU: models.ResourceMetric{
				Used:        cpuUsage.AsApproximateFloat64(),
				Utilization: cpuUtilization,
			},
			Memory: models.ResourceMetric{
				Used:        memUsage.AsApproximateFloat64() / 1024 / 1024,
				Utilization: memUtilization,
			},
			HasLivenessProbe:  c.HasProbes,
			HasReadinessProbe: c.HasProbes,
		}
		
		pod.Containers = append(pod.Containers, container)
		pod.TotalRestarts += c.RestartCount
	}

	// 添加事件
	for _, e := range data.Events {
		event := models.Event{
			Type:    e.Type,
			Reason:  e.Reason,
			Message: e.Message,
			Count:   e.Count,
			Time:    time.Now().Add(-10 * time.Minute), // 假设事件发生在10分钟前
		}
		pod.Events = append(pod.Events, event)
	}

	return pod, nil
}

// TestNormalPodInspection 测试正常Pod的分析
func TestNormalPodInspection(t *testing.T) {
	// 创建模拟客户端
	mockClient := NewMockPodClient()
	
	// 加载测试数据（这里需要先创建测试数据文件）
	err := mockClient.LoadPodData("test-pod-normal", "testdata/pod_normal.json")
	if err != nil {
		t.Skipf("跳过测试，原因：%v", err)
		return
	}

	// 加载测试规则
	rulesPath := filepath.Join("testdata", "rules_test.yaml")
	rulesEngine, err := rules.NewEngine(rulesPath)
	if err != nil {
		t.Fatalf("加载规则引擎失败: %v", err)
	}

	// 创建分析器
	analyzer := pod.NewPodAnalyzer(rulesEngine)

	// 获取Pod数据
	podData, err := mockClient.GetPod("test-pod-normal")
	if err != nil {
		t.Fatalf("获取Pod数据失败: %v", err)
	}

	// 分析Pod
	result, err := analyzer.AnalyzePod(podData)
	if err != nil {
		t.Fatalf("分析Pod失败: %v", err)
	}

	// 验证分析结果
	if result.PodName != podData.Name {
		t.Errorf("期望Pod名称为 %s，实际为 %s", podData.Name, result.PodName)
	}

	// 检查健康评分
	if result.HealthScore < 80 {
		t.Errorf("期望健康评分大于等于80，实际为 %d", result.HealthScore)
	}

	// 检查是否有严重问题
	hasCritical := false
	for _, item := range result.Items {
		if item.Severity == "critical" && !item.Passed {
			hasCritical = true
			break
		}
	}
	if hasCritical {
		t.Errorf("正常Pod不应该有严重问题")
	}
}

// TestProblemPodInspection 测试有问题的Pod分析
func TestProblemPodInspection(t *testing.T) {
	// 创建模拟客户端
	mockClient := NewMockPodClient()
	
	// 加载测试数据（这里需要先创建测试数据文件）
	err := mockClient.LoadPodData("test-pod-problem", "testdata/pod_problem.json")
	if err != nil {
		t.Skipf("跳过测试，原因：%v", err)
		return
	}

	// 加载测试规则
	rulesPath := filepath.Join("testdata", "rules_test.yaml")
	rulesEngine, err := rules.NewEngine(rulesPath)
	if err != nil {
		t.Fatalf("加载规则引擎失败: %v", err)
	}

	// 创建分析器
	analyzer := pod.NewPodAnalyzer(rulesEngine)

	// 获取Pod数据
	podData, err := mockClient.GetPod("test-pod-problem")
	if err != nil {
		t.Fatalf("获取Pod数据失败: %v", err)
	}

	// 分析Pod
	result, err := analyzer.AnalyzePod(podData)
	if err != nil {
		t.Fatalf("分析Pod失败: %v", err)
	}

	// 验证分析结果
	if result.PodName != podData.Name {
		t.Errorf("期望Pod名称为 %s，实际为 %s", podData.Name, result.PodName)
	}

	// 检查健康评分
	if result.HealthScore > 70 {
		t.Errorf("期望健康评分小于等于70，实际为 %d", result.HealthScore)
	}

	// 检查是否有问题
	hasIssues := false
	for _, item := range result.Items {
		if !item.Passed {
			hasIssues = true
			break
		}
	}
	if !hasIssues {
		t.Errorf("问题Pod应该有检测到的问题")
	}
}

// TestPodRuleEvaluation 测试Pod规则评估
func TestPodRuleEvaluation(t *testing.T) {
	// 加载测试规则
	rulesPath := filepath.Join("testdata", "rules_test.yaml")
	rulesEngine, err := rules.NewEngine(rulesPath)
	if err != nil {
		t.Fatalf("加载规则引擎失败: %v", err)
	}

	// 创建测试Pod
	testPod := &models.Pod{
		Name:      "test-pod",
		Namespace: "default",
		Phase:     corev1.PodRunning,
		Containers: []models.Container{
			{
				Name:  "high-cpu-container",
				Image: "test:latest",
				CPU: models.ResourceMetric{
					Used:        950.0,
					Utilization: 95.0, // 95% CPU使用率，应该触发告警
				},
				Memory: models.ResourceMetric{
					Used:        512.0,
					Utilization: 50.0, // 50% 内存使用率，应该正常
				},
				RestartCount: 0,
			},
		},
		TotalRestarts: 0,
	}

	// 创建分析器
	analyzer := pod.NewPodAnalyzer(rulesEngine)

	// 分析Pod
	result, err := analyzer.AnalyzePod(testPod)
	if err != nil {
		t.Fatalf("分析Pod失败: %v", err)
	}

	// 检查CPU高使用率规则是否被触发
	cpuRuleTriggered := false
	for _, item := range result.Items {
		if item.Metric == "pod_cpu_utilization" && !item.Passed {
			cpuRuleTriggered = true
			break
		}
	}

	if !cpuRuleTriggered {
		t.Errorf("CPU高使用率规则应该被触发")
	}

	// 检查内存使用率规则是否正常（不应该被触发）
	memoryRuleTriggered := false
	for _, item := range result.Items {
		if item.Metric == "pod_memory_utilization" && !item.Passed {
			memoryRuleTriggered = true
			break
		}
	}

	if memoryRuleTriggered {
		t.Errorf("内存使用率规则不应该被触发")
	}
} 