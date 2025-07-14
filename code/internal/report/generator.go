package report

import (
	"strconv"
	"time"

	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/analyzer/node"
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
		// 添加节点详情
		nodeDetail := NodeDetail{
			Name:        result.NodeName,
			HealthScore: result.HealthScore,
		}
		
		// 查找CPU、内存和Pod利用率
		for _, item := range result.Items {
			switch item.Metric {
			case "cpu_utilization":
				if val, err := strconv.ParseFloat(item.Value, 64); err == nil {
					nodeDetail.CPUUtilization = val
				}
			case "memory_utilization":
				if val, err := strconv.ParseFloat(item.Value, 64); err == nil {
					nodeDetail.MemoryUtilization = val
				}
			case "pods_utilization":
				if val, err := strconv.ParseFloat(item.Value, 64); err == nil {
					nodeDetail.PodUtilization = val
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