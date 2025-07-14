package report

import (
	"time"

	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/analyzer/node"
	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/rules"
)

// Severity 定义报告发现项的重要性级别
type Severity string

// 严重性级别常量
const (
	SeverityInfo     Severity = "INFO"     // 信息级别
	SeverityWarning  Severity = "WARNING"  // 警告级别
	SeverityError    Severity = "ERROR"    // 错误级别
	SeverityCritical Severity = "CRITICAL" // 严重级别
)

// Finding 表示分析过程中发现的单个问题
type Finding struct {
	// ResourceName 是有问题的资源名称
	ResourceName string `json:"resourceName"`
	// ResourceKind 表示资源类型（Node、Pod等）
	ResourceKind string `json:"resourceKind"`
	// RuleID 违反的规则ID
	RuleID string `json:"ruleID"`
	// Message 描述问题
	Message string `json:"message"`
	// Severity 表示问题的严重性
	Severity Severity `json:"severity"`
	// Recommendation 提供解决问题的建议
	Recommendation string `json:"recommendation,omitempty"`
	// Details 包含关于问题的额外上下文信息
	Details map[string]interface{} `json:"details,omitempty"`
}

// Report 表示完整的分析报告
type Report struct {
	// Timestamp 报告生成的时间戳
	Timestamp time.Time `json:"timestamp"`
	// ClusterName 被分析的集群名称
	ClusterName string `json:"clusterName,omitempty"`
	// Namespace 被分析的命名空间，如果适用
	Namespace string `json:"namespace,omitempty"`
	// Findings 包含所有检测到的问题
	Findings []Finding `json:"findings"`
	// Summary 包含报告的汇总统计信息
	Summary ReportSummary `json:"summary"`
}

// ReportSummary 包含报告的汇总信息
type ReportSummary struct {
	// TotalResources 分析的资源总数
	TotalResources int `json:"totalResources"`
	// ResourcesWithIssues 有至少一个问题的资源数量
	ResourcesWithIssues int `json:"resourcesWithIssues"`
	// FindingCounts 按严重性级别统计的问题数量
	FindingCounts map[Severity]int `json:"findingCounts"`
}

// Generator 定义报告生成器的接口
type Generator interface {
	// GenerateNodeReport 从节点分析结果创建报告
	GenerateNodeReport(results []node.AnalysisResult, rules []rules.Rule) *Report
}

// Formatter 定义报告输出格式化的接口
type Formatter interface {
	// Format 将报告转换为字符串表示
	Format(report *Report) string
} 