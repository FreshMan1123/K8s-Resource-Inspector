package node

import (
	"fmt"
	"time"

	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/models"
	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/rules"
)

// AnalysisItem 单个分析项目
type AnalysisItem struct {
	// 规则ID
	RuleID string `json:"rule_id"`
	// 规则名称
	Name string `json:"name"`
	// 类别
	Category string `json:"category"`
	// 严重程度：critical, warning, info
	Severity string `json:"severity"`
	// 检查的指标
	Metric string `json:"metric"`
	// 指标值
	Value string `json:"value"`
	// 阈值
	Threshold string `json:"threshold"`
	// 比较结果 (是否通过)
	Passed bool `json:"passed"`
	// 描述
	Description string `json:"description"`
	// 建议的修复措施
	Remediation string `json:"remediation"`
}

// AnalysisResult 分析结果
type AnalysisResult struct {
	// 节点名称
	NodeName string `json:"node_name"`
	// 分析结果项目列表
	Items []AnalysisItem `json:"items"`
	// 总体健康状态评分（0-100）
	HealthScore int `json:"health_score"`
	// 分析时间
	AnalyzedAt time.Time `json:"analyzed_at"`
}

// NodeAnalyzer 节点资源分析器
type NodeAnalyzer struct {
	rulesEngine *rules.Engine
}

// NewNodeAnalyzer 创建节点分析器
func NewNodeAnalyzer(rulesEngine *rules.Engine) *NodeAnalyzer {
	return &NodeAnalyzer{
		rulesEngine: rulesEngine,
	}
}

// AnalyzeNode 分析单个节点
func (na *NodeAnalyzer) AnalyzeNode(node *models.Node) (*AnalysisResult, error) {
	if node == nil {
		return nil, fmt.Errorf("节点为空")
	}

	// 创建分析结果
	result := &AnalysisResult{
		NodeName:   node.Name,
		Items:      make([]AnalysisItem, 0),
		AnalyzedAt: time.Now(),
	}

	// 分析CPU资源指标
	cpuItems := na.analyzeResourceMetric(node.Name, "cpu", node.CPU)
	result.Items = append(result.Items, cpuItems...)

	// 分析内存资源指标
	memoryItems := na.analyzeResourceMetric(node.Name, "memory", node.Memory)
	result.Items = append(result.Items, memoryItems...)

	// 分析临时存储资源指标
	storageItems := na.analyzeResourceMetric(node.Name, "ephemeral_storage", node.EphemeralStorage)
	result.Items = append(result.Items, storageItems...)

	// 分析Pod资源指标
	podItems := na.analyzeResourceMetric(node.Name, "pods", node.Pods)
	result.Items = append(result.Items, podItems...)

	// 分析节点压力状态
	pressureItems := na.analyzePressureStatus(node.Name, node.PressureStatus)
	result.Items = append(result.Items, pressureItems...)

	// 分析节点条件状态
	conditionItems := na.analyzeNodeConditions(node.Name, node.Ready, node.Conditions)
	result.Items = append(result.Items, conditionItems...)

	// 计算健康评分
	result.HealthScore = na.calculateHealthScore(result.Items)

	return result, nil
}

// AnalyzeNodes 分析多个节点
func (na *NodeAnalyzer) AnalyzeNodes(nodes *models.NodeList) ([]*AnalysisResult, error) {
	if nodes == nil {
		return nil, fmt.Errorf("节点列表为空")
	}

	results := make([]*AnalysisResult, 0, len(nodes.Items))

	for _, node := range nodes.Items {
		// 分析单个节点
		result, err := na.AnalyzeNode(&node)
		if err != nil {
			return nil, fmt.Errorf("分析节点 %s 失败: %w", node.Name, err)
		}

		results = append(results, result)
	}

	return results, nil
}

// analyzeResourceMetric 分析资源指标
func (na *NodeAnalyzer) analyzeResourceMetric(nodeName string, metricName string, metric models.ResourceMetric) []AnalysisItem {
	items := make([]AnalysisItem, 0)

	// 获取所有资源相关规则
	filter := rules.RuleFilter{
		Categories: []string{"node"},
	}
	allRules := na.rulesEngine.GetRules(filter)

	// 定义要检查的指标映射
	metricChecks := map[string]interface{}{
		fmt.Sprintf("%s_utilization", metricName):      metric.Utilization,
		fmt.Sprintf("%s_allocation_rate", metricName):  metric.AllocationRate,
	}

	// 对每个指标应用适当的规则
	for metric, value := range metricChecks {
		for _, rule := range allRules {
			// 跳过不匹配的规则
			if rule.Condition.Metric != metric {
				continue
			}

			// 评估规则
			ruleResult, err := na.rulesEngine.EvaluateRule(rule, "numeric", value)
			if err != nil {
				// 记录错误并继续
				continue
			}

			// 创建分析项
			item := AnalysisItem{
				RuleID:       ruleResult.RuleID,
				Name:         ruleResult.RuleName,
				Category:     rule.Category,
				Severity:     ruleResult.Severity,
				Metric:       metric,
				Value:        fmt.Sprintf("%.2f", value.(float64)),
				Threshold:    fmt.Sprintf("%v", ruleResult.ExpectedValue),
				Passed:       ruleResult.Passed,
				Description:  rule.Description,
				Remediation:  ruleResult.Remediation,
			}

			items = append(items, item)
		}
	}

	return items
}

// analyzePressureStatus 分析节点压力状态
func (na *NodeAnalyzer) analyzePressureStatus(nodeName string, pressure models.NodePressureStatus) []AnalysisItem {
	items := make([]AnalysisItem, 0)

	// 获取所有压力相关规则
	filter := rules.RuleFilter{
		Categories: []string{"node"},
	}
	allRules := na.rulesEngine.GetRules(filter)

	// 定义要检查的指标映射
	pressureChecks := map[string]bool{
		"memory_pressure":   pressure.MemoryPressure,
		"cpu_pressure":      pressure.CPUPressure,
		"disk_pressure":     pressure.DiskPressure,
		"pid_pressure":      pressure.PIDPressure,
		"network_pressure":  pressure.NetworkPressure,
	}

	// 对每个压力状态应用适当的规则
	for metric, value := range pressureChecks {
		for _, rule := range allRules {
			// 跳过不匹配的规则
			if rule.Condition.Metric != metric {
				continue
			}

			// 评估规则
			ruleResult, err := na.rulesEngine.EvaluateRule(rule, "boolean", value)
			if err != nil {
				// 记录错误并继续
				continue
			}

			// 创建分析项
			item := AnalysisItem{
				RuleID:       ruleResult.RuleID,
				Name:         ruleResult.RuleName,
				Category:     rule.Category,
				Severity:     ruleResult.Severity,
				Metric:       metric,
				Value:        fmt.Sprintf("%v", value),
				Threshold:    fmt.Sprintf("%v", ruleResult.ExpectedValue),
				Passed:       ruleResult.Passed,
				Description:  rule.Description,
				Remediation:  ruleResult.Remediation,
			}

			items = append(items, item)
		}
	}

	return items
}

// analyzeNodeConditions 分析节点条件状态
func (na *NodeAnalyzer) analyzeNodeConditions(nodeName string, ready bool, conditions []models.NodeConditionStatus) []AnalysisItem {
	items := make([]AnalysisItem, 0)

	// 获取所有条件相关规则
	filter := rules.RuleFilter{
		Categories: []string{"node"},
	}
	allRules := na.rulesEngine.GetRules(filter)

	// 检查节点Ready状态
	for _, rule := range allRules {
		if rule.Condition.Metric == "ready" {
			// 评估规则
			ruleResult, err := na.rulesEngine.EvaluateRule(rule, "boolean", ready)
			if err != nil {
				// 记录错误并继续
				continue
			}

			// 创建分析项
			item := AnalysisItem{
				RuleID:       ruleResult.RuleID,
				Name:         ruleResult.RuleName,
				Category:     rule.Category,
				Severity:     ruleResult.Severity,
				Metric:       "ready",
				Value:        fmt.Sprintf("%v", ready),
				Threshold:    fmt.Sprintf("%v", ruleResult.ExpectedValue),
				Passed:       ruleResult.Passed,
				Description:  rule.Description,
				Remediation:  ruleResult.Remediation,
			}

			items = append(items, item)
		}
	}

	return items
}

// calculateHealthScore 计算节点健康评分
func (na *NodeAnalyzer) calculateHealthScore(items []AnalysisItem) int {
	if len(items) == 0 {
		return 100
	}

	// 基础分数
	baseScore := 100
	
	// 不同严重程度的扣分
	deductions := map[string]int{
		"critical": 20,
		"warning":  10,
		"info":     5,
	}
	
	// 计算扣分
	totalDeduction := 0
	for _, item := range items {
		if !item.Passed {
			if deduction, exists := deductions[item.Severity]; exists {
				totalDeduction += deduction
			}
		}
	}
	
	// 计算最终分数，确保不低于0
	score := baseScore - totalDeduction
	if score < 0 {
		score = 0
	}
	
	return score
} 