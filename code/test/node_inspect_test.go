package test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/analyzer/node"
	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/models"
	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/rules"
	"k8s.io/apimachinery/pkg/api/resource"
)

// 节点测试数据
type TestNodeData struct {
	Name     string `json:"name"`
	Capacity struct {
		CPU    string `json:"cpu"`
		Memory string `json:"memory"`
	} `json:"capacity"`
	Metrics struct {
		CPUUsage    string `json:"cpuUsage"`
		MemoryUsage string `json:"memoryUsage"`
	} `json:"metrics"`
	Pods []struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Resources struct {
			Requests struct {
				CPU    string `json:"cpu"`
				Memory string `json:"memory"`
			} `json:"requests"`
		} `json:"resources"`
	} `json:"pods"`
}

// MockClusterClient 模拟集群客户端
type MockClusterClient struct {
	nodeData map[string]*TestNodeData
}

// NewMockClient 创建一个模拟集群客户端
func NewMockClient() *MockClusterClient {
	return &MockClusterClient{
		nodeData: make(map[string]*TestNodeData),
	}
}

// LoadNodeData 从文件加载节点数据
func (c *MockClusterClient) LoadNodeData(name, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取节点数据文件失败: %w", err)
	}

	nodeData := &TestNodeData{}
	if err := json.Unmarshal(data, nodeData); err != nil {
		return fmt.Errorf("解析节点数据失败: %w", err)
	}

	c.nodeData[name] = nodeData
	return nil
}

// GetNode 获取节点数据
func (c *MockClusterClient) GetNode(nodeName string) (*models.Node, error) {
	data, ok := c.nodeData[nodeName]
	if !ok {
		return nil, fmt.Errorf("节点 %s 不存在", nodeName)
	}

	// 解析CPU容量
	cpuCapacity, _ := resource.ParseQuantity(data.Capacity.CPU)
	
	// 解析CPU使用
	cpuUsage, _ := resource.ParseQuantity(data.Metrics.CPUUsage)
	
	// 计算CPU使用率
	cpuCapacityFloat := float64(cpuCapacity.MilliValue()) / 1000
	cpuUsageFloat := float64(cpuUsage.MilliValue()) / 1000
	cpuUtilization := (cpuUsageFloat / cpuCapacityFloat) * 100
	
	// 解析内存容量
	memCapacity, _ := resource.ParseQuantity(data.Capacity.Memory)
	
	// 解析内存使用
	memUsage, _ := resource.ParseQuantity(data.Metrics.MemoryUsage)
	
	// 计算内存使用率
	memCapacityFloat := float64(memCapacity.Value())
	memUsageFloat := float64(memUsage.Value())
	memUtilization := (memUsageFloat / memCapacityFloat) * 100
	
	// 计算分配给Pod的资源（简化）
	cpuAllocated := resource.NewQuantity(0, resource.DecimalSI)
	memAllocated := resource.NewQuantity(0, resource.DecimalSI)
	
	for _, pod := range data.Pods {
		podCPU, _ := resource.ParseQuantity(pod.Resources.Requests.CPU)
		podMem, _ := resource.ParseQuantity(pod.Resources.Requests.Memory)
		
		cpuAllocated.Add(podCPU)
		memAllocated.Add(podMem)
	}
	
	// 计算资源分配率
	cpuAllocationRate := (float64(cpuAllocated.MilliValue()) / float64(cpuCapacity.MilliValue())) * 100
	memAllocationRate := (float64(memAllocated.Value()) / float64(memCapacity.Value())) * 100

	// 创建节点模型
	node := &models.Node{
		Name:  data.Name,
		Roles: []string{"worker"},
		Addresses: map[string]string{
			"InternalIP": "192.168.1.100",
			"Hostname":   data.Name,
		},
		CreationTime: time.Now().Add(-24 * time.Hour), // 假设创建于一天前
		Ready:        true,
		Schedulable:  true,
		Labels:       map[string]string{"kubernetes.io/hostname": data.Name},
		Taints:       nil,
		NodeInfo: models.NodeInfo{
			KernelVersion:          "5.4.0",
			OSImage:                "Ubuntu 20.04 LTS",
			ContainerRuntimeVersion: "containerd://1.4.3",
			KubeletVersion:         "v1.23.0",
			KubeProxyVersion:       "v1.23.0",
			Architecture:           "amd64",
		},
		PressureStatus: models.NodePressureStatus{
			CPUPressure:     cpuUtilization > 90,
			MemoryPressure:  memUtilization > 90,
			DiskPressure:    false,
			PIDPressure:     false,
			NetworkPressure: false,
		},
		CPU: models.ResourceMetric{
			Capacity:       cpuCapacity,
			Allocatable:    cpuCapacity,
			Allocated:      *cpuAllocated,
			Used:           cpuUsage,
			Utilization:    cpuUtilization,
			AllocationRate: cpuAllocationRate,
		},
		Memory: models.ResourceMetric{
			Capacity:       memCapacity,
			Allocatable:    memCapacity,
			Allocated:      *memAllocated,
			Used:           memUsage,
			Utilization:    memUtilization,
			AllocationRate: memAllocationRate,
		},
		EphemeralStorage: models.ResourceMetric{
			Capacity:    *resource.NewQuantity(100*1024*1024*1024, resource.BinarySI), // 100Gi
			Allocatable: *resource.NewQuantity(100*1024*1024*1024, resource.BinarySI), // 100Gi
			Used:        *resource.NewQuantity(20*1024*1024*1024, resource.BinarySI),  // 20Gi
			Utilization: 20.0,                                                       // 20%
		},
		Pods: models.ResourceMetric{
			Capacity:    *resource.NewQuantity(110, resource.DecimalSI), // 110 pods
			Allocatable: *resource.NewQuantity(110, resource.DecimalSI), // 110 pods
			Used:        *resource.NewQuantity(int64(len(data.Pods)), resource.DecimalSI),
			Utilization: float64(len(data.Pods)) / 110 * 100,
		},
		RunningPods: len(data.Pods),
		Conditions:  []models.NodeConditionStatus{},
	}

	return node, nil
}

// TestNormalNodeInspection 测试正常节点的分析
func TestNormalNodeInspection(t *testing.T) {
	// 创建模拟客户端
	mockClient := NewMockClient()
	
	// 加载测试数据
	err := mockClient.LoadNodeData("test-node", "testdata/node_normal.json")
	if err != nil {
		t.Fatalf("加载测试数据失败: %v", err)
	}

	// 加载测试规则
	rulesPath := filepath.Join("testdata", "rules_test.yaml")
	rulesEngine, err := rules.NewEngine(rulesPath)
	if err != nil {
		t.Fatalf("加载规则引擎失败: %v", err)
	}

	// 创建分析器
	analyzer := node.NewNodeAnalyzer(rulesEngine)

	// 获取节点数据
	nodeData, err := mockClient.GetNode("test-node")
	if err != nil {
		t.Fatalf("获取节点数据失败: %v", err)
	}

	// 分析节点
	result, err := analyzer.AnalyzeNode(nodeData)
	if err != nil {
		t.Fatalf("分析节点失败: %v", err)
	}

	// 计算未通过的项目数量
	failedItems := 0
	for _, item := range result.Items {
		if !item.Passed {
			failedItems++
		}
	}

	// 验证结果 - 正常节点不应该有未通过的项目
	if failedItems > 0 {
		t.Errorf("正常节点不应有未通过项目，实际发现 %d 个", failedItems)
		for _, item := range result.Items {
			if !item.Passed {
				t.Logf("未通过: %s - %s", item.Name, item.Description)
			}
		}
	}
}

// TestHighLoadNodeInspection 测试高负载节点的分析
func TestHighLoadNodeInspection(t *testing.T) {
	// 创建模拟客户端
	mockClient := NewMockClient()
	
	// 加载测试数据
	err := mockClient.LoadNodeData("test-node", "testdata/node_critical.json")
	if err != nil {
		t.Fatalf("加载测试数据失败: %v", err)
	}

	// 加载测试规则
	rulesPath := filepath.Join("testdata", "rules_test.yaml")
	rulesEngine, err := rules.NewEngine(rulesPath)
	if err != nil {
		t.Fatalf("加载规则引擎失败: %v", err)
	}

	// 创建分析器
	analyzer := node.NewNodeAnalyzer(rulesEngine)

	// 获取节点数据
	nodeData, err := mockClient.GetNode("test-node")
	if err != nil {
		t.Fatalf("获取节点数据失败: %v", err)
	}

	// 分析节点
	result, err := analyzer.AnalyzeNode(nodeData)
	if err != nil {
		t.Fatalf("分析节点失败: %v", err)
	}

	// 计算未通过的项目数量
	failedItems := 0
	hasCPUUtilizationWarning := false
	for _, item := range result.Items {
		if !item.Passed {
			failedItems++
			if item.Metric == "cpu_utilization" {
				hasCPUUtilizationWarning = true
			}
		}
	}

	// 验证结果 - 高负载节点应该有未通过的项目
	if failedItems == 0 {
		t.Errorf("高负载节点应有未通过项目，实际未发现")
	}

	// 检查是否包含CPU使用率过高的警告
	if !hasCPUUtilizationWarning {
		t.Errorf("应该检测到CPU使用率过高")
	}
}

// TestRuleEvaluation 测试规则评估逻辑
func TestRuleEvaluation(t *testing.T) {
	// 加载测试规则
	rulesPath := filepath.Join("testdata", "rules_test.yaml")
	rulesEngine, err := rules.NewEngine(rulesPath)
	if err != nil {
		t.Fatalf("加载规则引擎失败: %v", err)
	}

	// 获取所有规则
	filter := rules.RuleFilter{}
	allRules := rulesEngine.GetRules(filter)
	
	// 验证规则数量
	if len(allRules) != 2 {
		t.Errorf("应加载2条规则，实际加载了%d条", len(allRules))
	}

	// 验证规则内容
	for _, rule := range allRules {
		switch rule.ID {
		case "node-high-cpu":
			if rule.Severity != "critical" {
				t.Errorf("node-high-cpu规则严重性应为critical，实际为%s", rule.Severity)
			}
		case "node-high-memory":
			if rule.Severity != "warning" {
				t.Errorf("node-high-memory规则严重性应为warning，实际为%s", rule.Severity)
			}
		default:
			t.Errorf("未知规则ID: %s", rule.ID)
		}
	}
} 