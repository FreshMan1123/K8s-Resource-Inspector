package report

import (
	"strconv"
	"time"

	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/analyzer/node"
	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/analyzer/pod"
	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/rules"
)

// DefaultGenerator 实现了Generator接口
type DefaultGenerator struct {
	ClusterName string
	Namespace   string
}

// NewGenerator 创建一个新的报告生成器
func NewGenerator(clusterName, namespace string) Generator {
	return &DefaultGenerator{
		ClusterName: clusterName,
		Namespace:   namespace,
	}
}

// GenerateNodeReport 从节点分析结果创建报告
func (g *DefaultGenerator) GenerateNodeReport(results []node.AnalysisResult, rulesList []rules.Rule) *Report {
	// 创建一个新报告
	report := &Report{
		Timestamp:   time.Now(),
		ClusterName: g.ClusterName,
		Namespace:   g.Namespace,
		Findings:    make([]Finding, 0),
		NodeDetails: make([]NodeDetail, 0, len(results)),
		Summary: ReportSummary{
			TotalResources:     len(results),
			ResourcesWithIssues: 0,
			FindingCounts: map[Severity]int{
				SeverityInfo:     0,
				SeverityWarning:  0,
				SeverityError:    0,
				SeverityCritical: 0,
			},
		},
	}

	// 创建规则映射，方便查找
	rulesMap := make(map[string]rules.Rule)
	for _, rule := range rulesList {
		rulesMap[rule.ID] = rule
	}

	// 处理每个分析结果
	resourcesWithIssues := make(map[string]bool)
	
	for _, result := range results {
		// 从分析结果中获取节点详情
		nodeDetail := g.createNodeDetailFromAnalysisResult(&result)
		
		// 查找CPU、内存和Pod利用率以及Ready状态
		for _, item := range result.Items {
			switch item.Metric {
			case "cpu_utilization":
				if val, err := strconv.ParseFloat(item.Value, 64); err == nil {
					nodeDetail.CPUUtilization = val
					nodeDetail.CPU.Utilization = val
				}
			case "memory_utilization":
				if val, err := strconv.ParseFloat(item.Value, 64); err == nil {
					nodeDetail.MemoryUtilization = val
					nodeDetail.Memory.Utilization = val
				}
			case "cpu_allocation_rate":
				if val, err := strconv.ParseFloat(item.Value, 64); err == nil {
					nodeDetail.CPU.AllocationRate = val
				}
			case "memory_allocation_rate":
				if val, err := strconv.ParseFloat(item.Value, 64); err == nil {
					nodeDetail.Memory.AllocationRate = val
				}
			case "ephemeral_storage_utilization":
				if val, err := strconv.ParseFloat(item.Value, 64); err == nil {
					nodeDetail.EphemeralStorage.Utilization = val
				}
			case "ephemeral_storage_allocation_rate":
				if val, err := strconv.ParseFloat(item.Value, 64); err == nil {
					nodeDetail.EphemeralStorage.AllocationRate = val
				}
			}
		}
		
		// 添加到报告
		report.NodeDetails = append(report.NodeDetails, nodeDetail)
		
		// 查找未通过的分析项
		for _, item := range result.Items {
			if !item.Passed {
				resourcesWithIssues[result.NodeName] = true
				
				severity := mapSeverity(item.Severity)
				report.Summary.FindingCounts[severity]++
				
				finding := Finding{
					ResourceName:   result.NodeName,
					ResourceKind:   "Node",
					RuleID:         item.RuleID,
					Message:        item.Description,
					Severity:       severity,
					Recommendation: item.Remediation,
					Details:        make(map[string]interface{}),
				}
				
				// 添加资源指标到详情
				finding.Details["metric_name"] = item.Metric
				finding.Details["metric_value"] = item.Value
				finding.Details["threshold"] = item.Threshold
				
				// 如果存在，添加规则信息
				if rule, exists := rulesMap[item.RuleID]; exists {
					finding.Details["rule_description"] = rule.Description
					finding.Details["rule_category"] = rule.Category
				}
				
				// 将发现项添加到报告
				report.Findings = append(report.Findings, finding)
			}
		}
	}
	
	// 更新摘要
	report.Summary.ResourcesWithIssues = len(resourcesWithIssues)
	
	return report
}

// GeneratePodReport 从Pod分析结果创建报告
func (g *DefaultGenerator) GeneratePodReport(results []*pod.AnalysisResult, rules []rules.Rule) *Report {
	report := &Report{
		Timestamp:   time.Now(),
		ClusterName: g.ClusterName,
		Namespace:   g.Namespace,
		Findings:    make([]Finding, 0),
		Summary: ReportSummary{
			FindingCounts: make(map[Severity]int),
		},
	}

	// 添加Pod详情
	for _, result := range results {
		// 添加发现项
		for _, item := range result.Items {
			if !item.Passed {
				severity := mapSeverity(item.Severity)
				finding := Finding{
					ResourceName: result.PodName,
					ResourceKind: "Pod",
					RuleID:       item.RuleID,
					Message:      item.Description,
					Severity:     severity,
					Recommendation: item.Remediation,
					Details: map[string]interface{}{
						"metric":    item.Metric,
						"value":     item.Value,
						"threshold": item.Threshold,
						"namespace": result.Namespace,
					},
				}

				report.Findings = append(report.Findings, finding)
				report.Summary.FindingCounts[severity]++
			}
		}
	}

	// 更新统计信息
	report.Summary.TotalResources = len(results)
	report.Summary.ResourcesWithIssues = countResourcesWithIssues(results)

	return report
}

// countResourcesWithIssues 计算有问题的Pod数量
func countResourcesWithIssues(results []*pod.AnalysisResult) int {
	count := 0
	for _, result := range results {
		hasIssues := false
		for _, item := range result.Items {
			if !item.Passed {
				hasIssues = true
				break
			}
		}
		if hasIssues {
			count++
		}
	}
	return count
}

// createNodeDetailFromAnalysisResult 从分析结果创建节点详情
func (g *DefaultGenerator) createNodeDetailFromAnalysisResult(result *node.AnalysisResult) NodeDetail {
	// 创建节点详情
	nodeDetail := NodeDetail{
		Name:            result.NodeName,
		HealthScore:     result.HealthScore,
		Ready:           result.NodeBasicInfo.Ready,
		RunningPods:     result.NodeBasicInfo.RunningPods,
		TotalPods:       result.NodeBasicInfo.TotalPods,
		MaxPods:         result.NodeBasicInfo.MaxPods,
		PodUtilization:  result.NodeBasicInfo.PodUtilization,
		CreationTime:    result.CreationTime,
		Schedulable:     result.Schedulable,
		Roles:           result.Roles,
		Addresses:       result.Addresses,
	}

	// 填充节点信息
	nodeDetail.NodeInfo.KernelVersion = result.NodeInfo.KernelVersion
	nodeDetail.NodeInfo.OSImage = result.NodeInfo.OSImage
	nodeDetail.NodeInfo.ContainerRuntimeVersion = result.NodeInfo.ContainerRuntimeVersion
	nodeDetail.NodeInfo.KubeletVersion = result.NodeInfo.KubeletVersion
	nodeDetail.NodeInfo.KubeProxyVersion = result.NodeInfo.KubeProxyVersion
	nodeDetail.NodeInfo.Architecture = result.NodeInfo.Architecture
	
	// 填充资源信息
	nodeDetail.CPU.Capacity = result.Resources.CPU.Capacity
	nodeDetail.CPU.Allocatable = result.Resources.CPU.Allocatable
	nodeDetail.CPU.Allocated = result.Resources.CPU.Allocated
	nodeDetail.CPU.Used = result.Resources.CPU.Used
	
	nodeDetail.Memory.Capacity = result.Resources.Memory.Capacity
	nodeDetail.Memory.Allocatable = result.Resources.Memory.Allocatable
	nodeDetail.Memory.Allocated = result.Resources.Memory.Allocated
	nodeDetail.Memory.Used = result.Resources.Memory.Used
	
	nodeDetail.EphemeralStorage.Capacity = result.Resources.EphemeralStorage.Capacity
	nodeDetail.EphemeralStorage.Allocatable = result.Resources.EphemeralStorage.Allocatable
	nodeDetail.EphemeralStorage.Allocated = result.Resources.EphemeralStorage.Allocated
	nodeDetail.EphemeralStorage.Used = result.Resources.EphemeralStorage.Used

	// 查找节点信息项
	for _, item := range result.Items {
		// 检查节点压力状态
		switch item.Metric {
		case "memory_pressure":
			if val, err := strconv.ParseBool(item.Value); err == nil {
				nodeDetail.PressureStatus.MemoryPressure = val
			}
		case "cpu_pressure":
			if val, err := strconv.ParseBool(item.Value); err == nil {
				nodeDetail.PressureStatus.CPUPressure = val
			}
		case "disk_pressure":
			if val, err := strconv.ParseBool(item.Value); err == nil {
				nodeDetail.PressureStatus.DiskPressure = val
			}
		case "pid_pressure":
			if val, err := strconv.ParseBool(item.Value); err == nil {
				nodeDetail.PressureStatus.PIDPressure = val
			}
		case "network_pressure":
			if val, err := strconv.ParseBool(item.Value); err == nil {
				nodeDetail.PressureStatus.NetworkPressure = val
			}
		}
	}

	return nodeDetail
}

// mapSeverity 将分析器严重性转换为报告严重性
func mapSeverity(severity string) Severity {
	switch severity {
	case "info":
		return SeverityInfo
	case "warning":
		return SeverityWarning
	case "error":
		return SeverityError
	case "critical":
		return SeverityCritical
	default:
		return SeverityInfo
	}
} 