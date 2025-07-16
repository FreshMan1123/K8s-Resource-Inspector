package report

import (
	"time"

	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/analyzer/node"
	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/rules"
	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/analyzer/pod"
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

// NodeDetail 表示节点的详细信息
type NodeDetail struct {
	// 节点名称
	Name string `json:"name"`
	// 节点就绪状态
	Ready bool `json:"ready"`
	// 节点是否可调度
	Schedulable bool `json:"schedulable"`
	// 节点角色（master/worker）
	Roles []string `json:"roles"`
	// 节点IP地址
	Addresses map[string]string `json:"addresses"`
	// 节点创建时间
	CreationTime time.Time `json:"creationTime"`
	
	// 节点信息
	NodeInfo struct {
		// 内核版本
		KernelVersion string `json:"kernelVersion"`
		// 操作系统
		OSImage string `json:"osImage"`
		// 容器运行时版本
		ContainerRuntimeVersion string `json:"containerRuntimeVersion"`
		// Kubelet版本
		KubeletVersion string `json:"kubeletVersion"`
		// Kube-Proxy版本
		KubeProxyVersion string `json:"kubeProxyVersion"`
		// 架构
		Architecture string `json:"architecture"`
	} `json:"nodeInfo"`
	
	// 节点压力状态
	PressureStatus struct {
		// CPU压力状态
		CPUPressure bool `json:"cpuPressure"`
		// 内存压力状态
		MemoryPressure bool `json:"memoryPressure"`
		// 磁盘压力状态
		DiskPressure bool `json:"diskPressure"`
		// 网络压力状态
		NetworkPressure bool `json:"networkPressure"`
		// PID压力状态
		PIDPressure bool `json:"pidPressure"`
	} `json:"pressureStatus"`
	
	// CPU资源指标
	CPU struct {
		// 资源总量
		Capacity string `json:"capacity"`
		// 可分配资源总量
		Allocatable string `json:"allocatable"`
		// 已分配给Pod的资源量
		Allocated string `json:"allocated"`
		// 实际使用的资源量
		Used string `json:"used"`
		// 资源利用率
		Utilization float64 `json:"utilization"`
		// 资源分配率
		AllocationRate float64 `json:"allocationRate"`
	} `json:"cpu"`
	
	// 内存资源指标
	Memory struct {
		// 资源总量
		Capacity string `json:"capacity"`
		// 可分配资源总量
		Allocatable string `json:"allocatable"`
		// 已分配给Pod的资源量
		Allocated string `json:"allocated"`
		// 实际使用的资源量
		Used string `json:"used"`
		// 资源利用率
		Utilization float64 `json:"utilization"`
		// 资源分配率
		AllocationRate float64 `json:"allocationRate"`
	} `json:"memory"`
	
	// 临时存储资源指标
	EphemeralStorage struct {
		// 资源总量
		Capacity string `json:"capacity"`
		// 可分配资源总量
		Allocatable string `json:"allocatable"`
		// 已分配给Pod的资源量
		Allocated string `json:"allocated"`
		// 实际使用的资源量
		Used string `json:"used"`
		// 资源利用率
		Utilization float64 `json:"utilization"`
		// 资源分配率
		AllocationRate float64 `json:"allocationRate"`
	} `json:"ephemeralStorage"`
	
	// CPU利用率 (保留向后兼容)
	CPUUtilization float64 `json:"cpuUtilization"`
	// 内存利用率 (保留向后兼容)
	MemoryUtilization float64 `json:"memoryUtilization"`
	// 运行中的Pod数量
	RunningPods int `json:"runningPods"`
	// 最大Pod数量
	MaxPods int `json:"maxPods"`
	// 总Pod数量（包括运行中和已完成的）
	TotalPods int `json:"totalPods"`
	// Pod利用率
	PodUtilization float64 `json:"podUtilization"`
	// 健康评分
	HealthScore int `json:"healthScore"`
}

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
	// NodeDetails 包含所有节点的详细信息
	NodeDetails []NodeDetail `json:"nodeDetails,omitempty"`
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
	// GeneratePodReport 从Pod分析结果创建报告
	GeneratePodReport(results []*pod.AnalysisResult, rules []rules.Rule) *Report
}

// Formatter 定义报告输出格式化的接口
type Formatter interface {
	// Format 将报告转换为字符串表示
	Format(report *Report) string
} 