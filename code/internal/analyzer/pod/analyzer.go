package pod

import (
	"fmt"
	"time"

	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/cluster"
	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/models"
	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/rules"
	
	corev1 "k8s.io/api/core/v1"
)

// RulesEngine 规则引擎接口
type RulesEngine interface {
	// GetRules 获取规则
	GetRules(filter rules.RuleFilter) []rules.Rule
	// EvaluateRule 评估单个规则
	EvaluateRule(rule rules.Rule, metricType string, actualValue interface{}) (*rules.RuleResult, error)
	// SetEnvironment 设置当前环境
	SetEnvironment(env string)
	// GetEnvironment 获取当前环境
	GetEnvironment() string
	// DetermineEnvironment 根据集群名称确定环境
	DetermineEnvironment(clusterName string) string
	// RegisterValidator 注册验证器
	RegisterValidator(name string, validator rules.Validator)
}

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
	// Pod名称
	PodName string `json:"pod_name"`
	// Pod命名空间
	Namespace string `json:"namespace"`
	// 分析结果项目列表
	Items []AnalysisItem `json:"items"`
	// 总体健康状态评分（0-100）
	HealthScore int `json:"health_score"`
	// 分析时间
	AnalyzedAt time.Time `json:"analyzed_at"`
	
	// Pod基本信息
	PodBasicInfo struct {
		// Pod状态
		Phase string `json:"phase"`
		// Pod创建时间
		CreationTime time.Time `json:"creation_time"`
		// Pod运行时长
		RunningDuration string `json:"running_duration"`
		// Pod IP地址
		IP string `json:"ip"`
		// 所在节点名称
		NodeName string `json:"node_name"`
		// 重启次数
		TotalRestarts int `json:"total_restarts"`
		// QoS类别
		QOSClass string `json:"qos_class"`
	} `json:"pod_basic_info"`
	
	// 容器信息
	Containers []struct {
		// 容器名称
		Name string `json:"name"`
		// 容器镜像
		Image string `json:"image"`
		// 容器状态
		State string `json:"state"`
		// 容器就绪状态
		Ready bool `json:"ready"`
		// 重启次数
		RestartCount int `json:"restart_count"`
		// CPU资源
		CPU struct {
			// 请求量
			Request string `json:"request"`
			// 限制量
			Limit string `json:"limit"`
			// 实际使用量
			Used string `json:"used"`
			// 利用率
			Utilization float64 `json:"utilization"`
		} `json:"cpu"`
		// 内存资源
		Memory struct {
			// 请求量
			Request string `json:"request"`
			// 限制量
			Limit string `json:"limit"`
			// 实际使用量
			Used string `json:"used"`
			// 利用率
			Utilization float64 `json:"utilization"`
		} `json:"memory"`
		// 是否有健康检查
		HasProbes bool `json:"has_probes"`
	} `json:"containers"`
	
	// 相关事件
	Events []struct {
		// 事件类型
		Type string `json:"type"`
		// 事件原因
		Reason string `json:"reason"`
		// 事件消息
		Message string `json:"message"`
		// 事件时间
		Time time.Time `json:"time"`
		// 事件计数
		Count int `json:"count"`
	} `json:"events"`
}

// PodAnalyzer Pod资源分析器
type PodAnalyzer struct {
	rulesEngine RulesEngine
	client      *cluster.Client
}

// NewPodAnalyzer 创建Pod分析器
func NewPodAnalyzer(rulesEngine RulesEngine) *PodAnalyzer {
	return &PodAnalyzer{
		rulesEngine: rulesEngine,
	}
}

// NewPodAnalyzerWithClient 创建带有集群客户端的Pod分析器
func NewPodAnalyzerWithClient(rulesEngine RulesEngine, client *cluster.Client) *PodAnalyzer {
	return &PodAnalyzer{
		rulesEngine: rulesEngine,
		client:      client,
	}
}

// SetClient 设置集群客户端
func (pa *PodAnalyzer) SetClient(client *cluster.Client) {
	pa.client = client
}

// AnalyzePod 分析单个Pod
func (pa *PodAnalyzer) AnalyzePod(pod *models.Pod) (*AnalysisResult, error) {
	if pod == nil {
		return nil, fmt.Errorf("Pod为空")
	}

	// 创建分析结果
	result := &AnalysisResult{
		PodName:    pod.Name,
		Namespace:  pod.Namespace,
		Items:      make([]AnalysisItem, 0),
		AnalyzedAt: time.Now(),
	}
	
	// 填充Pod基本信息
	result.PodBasicInfo.Phase = string(pod.Phase)
	result.PodBasicInfo.CreationTime = pod.CreationTime
	result.PodBasicInfo.RunningDuration = formatDuration(pod.RunningDuration)
	result.PodBasicInfo.IP = pod.IP
	result.PodBasicInfo.NodeName = pod.NodeName
	result.PodBasicInfo.TotalRestarts = pod.TotalRestarts
	result.PodBasicInfo.QOSClass = string(pod.QOSClass)
	
	// 填充容器信息
	for _, container := range pod.Containers {
		containerInfo := struct {
			// 容器名称
			Name string `json:"name"`
			// 容器镜像
			Image string `json:"image"`
			// 容器状态
			State string `json:"state"`
			// 容器就绪状态
			Ready bool `json:"ready"`
			// 重启次数
			RestartCount int `json:"restart_count"`
			// CPU资源
			CPU struct {
				// 请求量
				Request string `json:"request"`
				// 限制量
				Limit string `json:"limit"`
				// 实际使用量
				Used string `json:"used"`
				// 利用率
				Utilization float64 `json:"utilization"`
			} `json:"cpu"`
			// 内存资源
			Memory struct {
				// 请求量
				Request string `json:"request"`
				// 限制量
				Limit string `json:"limit"`
				// 实际使用量
				Used string `json:"used"`
				// 利用率
				Utilization float64 `json:"utilization"`
			} `json:"memory"`
			// 是否有健康检查
			HasProbes bool `json:"has_probes"`
		}{
			Name:         container.Name,
			Image:        container.Image,
			State:        getContainerState(container.State),
			Ready:        container.Ready,
			RestartCount: container.RestartCount,
			HasProbes:    container.HasLivenessProbe || container.HasReadinessProbe || container.HasStartupProbe,
		}
		
		// 填充CPU信息
		if request := container.Requests.Cpu(); request != nil {
			containerInfo.CPU.Request = request.String()
		}
		if limit := container.Limits.Cpu(); limit != nil {
			containerInfo.CPU.Limit = limit.String()
		}
		containerInfo.CPU.Used = container.CPU.Used.String()
		containerInfo.CPU.Utilization = container.CPU.Utilization
		
		// 填充内存信息
		if request := container.Requests.Memory(); request != nil {
			containerInfo.Memory.Request = request.String()
		}
		if limit := container.Limits.Memory(); limit != nil {
			containerInfo.Memory.Limit = limit.String()
		}
		containerInfo.Memory.Used = container.Memory.Used.String()
		containerInfo.Memory.Utilization = container.Memory.Utilization
		
		result.Containers = append(result.Containers, containerInfo)
	}
	
	// 填充事件信息
	for _, event := range pod.Events {
		eventInfo := struct {
			// 事件类型
			Type string `json:"type"`
			// 事件原因
			Reason string `json:"reason"`
			// 事件消息
			Message string `json:"message"`
			// 事件时间
			Time time.Time `json:"time"`
			// 事件计数
			Count int `json:"count"`
		}{
			Type:    event.Type,
			Reason:  event.Reason,
			Message: event.Message,
			Time:    event.Time,
			Count:   event.Count,
		}
		
		result.Events = append(result.Events, eventInfo)
	}

	// 分析Pod状态
	statusItems := pa.analyzePodStatus(pod)
	result.Items = append(result.Items, statusItems...)

	// 分析Pod资源使用情况
	resourceItems := pa.analyzePodResources(pod)
	result.Items = append(result.Items, resourceItems...)

	// 分析Pod稳定性
	stabilityItems := pa.analyzePodStability(pod)
	result.Items = append(result.Items, stabilityItems...)

	// 分析Pod配置
	configItems := pa.analyzePodConfig(pod)
	result.Items = append(result.Items, configItems...)

	// 计算健康评分
	result.HealthScore = pa.calculateHealthScore(result.Items)

	return result, nil
}

// AnalyzePodByName 根据Pod名称分析Pod
func (pa *PodAnalyzer) AnalyzePodByName(namespace, name string) (*AnalysisResult, error) {
	if pa.client == nil {
		return nil, fmt.Errorf("未设置集群客户端")
	}

	// 获取Pod数据
	pod, err := pa.client.GetPod(namespace, name)
	if err != nil {
		return nil, fmt.Errorf("获取Pod数据失败: %w", err)
	}

	// 分析Pod
	return pa.AnalyzePod(pod)
}

// AnalyzePodsInNamespace 分析命名空间中的所有Pod
func (pa *PodAnalyzer) AnalyzePodsInNamespace(namespace string) ([]*AnalysisResult, error) {
	if pa.client == nil {
		return nil, fmt.Errorf("未设置集群客户端")
	}

	// 获取Pod列表
	podList, err := pa.client.ListPods(namespace)
	if err != nil {
		return nil, fmt.Errorf("获取Pod列表失败: %w", err)
	}

	// 分析所有Pod
	results := make([]*AnalysisResult, 0, len(podList.Items))
	for i := range podList.Items {
		result, err := pa.AnalyzePod(&podList.Items[i])
		if err != nil {
			return nil, fmt.Errorf("分析Pod %s/%s 失败: %w", podList.Items[i].Namespace, podList.Items[i].Name, err)
		}
		results = append(results, result)
	}

	return results, nil
}

// analyzePodStatus 分析Pod状态
func (pa *PodAnalyzer) analyzePodStatus(pod *models.Pod) []AnalysisItem {
	items := make([]AnalysisItem, 0)

	// 获取所有Pod状态相关规则
	filter := rules.RuleFilter{
		Categories: []string{"pod"},
	}
	allRules := pa.rulesEngine.GetRules(filter)

	// 检查Pod是否处于Running状态
	if pod.Phase != corev1.PodRunning {
		// 计算Pod处于非Running状态的时长（分钟）
		notRunningDuration := time.Since(pod.CreationTime).Minutes()
		
		// 查找适用的规则
		for _, rule := range allRules {
			if rule.Condition.Metric == "pod_not_running_duration" {
				// 评估规则
				ruleResult, err := pa.rulesEngine.EvaluateRule(rule, "numeric", notRunningDuration)
				if err != nil {
					// 记录错误并继续
					continue
				}
				
				// 创建分析项
				item := AnalysisItem{
					RuleID:      ruleResult.RuleID,
					Name:        ruleResult.RuleName,
					Category:    rule.Category,
					Severity:    ruleResult.Severity,
					Metric:      "pod_not_running_duration",
					Value:       fmt.Sprintf("%.1f", notRunningDuration),
					Threshold:   fmt.Sprintf("%v", ruleResult.ExpectedValue),
					Passed:      !ruleResult.Passed, // 反转结果
					Description: fmt.Sprintf("Pod处于%s状态%s", pod.Phase, formatDuration(time.Since(pod.CreationTime))),
					Remediation: ruleResult.Remediation,
				}
				
				items = append(items, item)
			}
		}
	}

	return items
}

// analyzePodResources 分析Pod资源使用情况
func (pa *PodAnalyzer) analyzePodResources(pod *models.Pod) []AnalysisItem {
	items := make([]AnalysisItem, 0)

	// 获取所有Pod资源相关规则
	filter := rules.RuleFilter{
		Categories: []string{"pod"},
	}
	allRules := pa.rulesEngine.GetRules(filter)

	// 检查每个容器的资源使用情况
	for _, container := range pod.Containers {
		// 检查CPU使用率
		if container.CPU.Utilization > 0 {
			for _, rule := range allRules {
				if rule.Condition.Metric == "pod_cpu_utilization" {
					// 评估规则
					ruleResult, err := pa.rulesEngine.EvaluateRule(rule, "numeric", container.CPU.Utilization)
					if err != nil {
						// 记录错误并继续
						continue
					}
					
					// 创建分析项
					item := AnalysisItem{
						RuleID:      ruleResult.RuleID,
						Name:        ruleResult.RuleName,
						Category:    rule.Category,
						Severity:    ruleResult.Severity,
						Metric:      "pod_cpu_utilization",
						Value:       fmt.Sprintf("%.2f", container.CPU.Utilization),
						Threshold:   fmt.Sprintf("%v", ruleResult.ExpectedValue),
						Passed:      !ruleResult.Passed, // 反转结果
						Description: fmt.Sprintf("容器 %s CPU使用率为 %.2f%%", container.Name, container.CPU.Utilization),
						Remediation: ruleResult.Remediation,
					}
					
					items = append(items, item)
				}
			}
		}
		
		// 检查内存使用率
		if container.Memory.Utilization > 0 {
			for _, rule := range allRules {
				if rule.Condition.Metric == "pod_memory_utilization" {
					// 评估规则
					ruleResult, err := pa.rulesEngine.EvaluateRule(rule, "numeric", container.Memory.Utilization)
					if err != nil {
						// 记录错误并继续
						continue
					}
					
					// 创建分析项
					item := AnalysisItem{
						RuleID:      ruleResult.RuleID,
						Name:        ruleResult.RuleName,
						Category:    rule.Category,
						Severity:    ruleResult.Severity,
						Metric:      "pod_memory_utilization",
						Value:       fmt.Sprintf("%.2f", container.Memory.Utilization),
						Threshold:   fmt.Sprintf("%v", ruleResult.ExpectedValue),
						Passed:      !ruleResult.Passed, // 反转结果
						Description: fmt.Sprintf("容器 %s 内存使用率为 %.2f%%", container.Name, container.Memory.Utilization),
						Remediation: ruleResult.Remediation,
					}
					
					items = append(items, item)
				}
			}
		}
		
		// 检查是否缺少资源限制
		cpuLimit := container.Limits.Cpu()
		memoryLimit := container.Limits.Memory()
		
		if (cpuLimit == nil || cpuLimit.IsZero()) && (memoryLimit == nil || memoryLimit.IsZero()) {
			for _, rule := range allRules {
				if rule.Condition.Metric == "pod_missing_resource_limits" {
					// 评估规则
					ruleResult, err := pa.rulesEngine.EvaluateRule(rule, "boolean", true)
					if err != nil {
						// 记录错误并继续
						continue
					}
					
					// 创建分析项
					item := AnalysisItem{
						RuleID:      ruleResult.RuleID,
						Name:        ruleResult.RuleName,
						Category:    rule.Category,
						Severity:    ruleResult.Severity,
						Metric:      "pod_missing_resource_limits",
						Value:       "true",
						Threshold:   "false",
						Passed:      !ruleResult.Passed, // 反转结果
						Description: fmt.Sprintf("容器 %s 缺少资源限制", container.Name),
						Remediation: ruleResult.Remediation,
					}
					
					items = append(items, item)
				}
			}
		}
	}

	return items
}

// analyzePodStability 分析Pod稳定性
func (pa *PodAnalyzer) analyzePodStability(pod *models.Pod) []AnalysisItem {
	items := make([]AnalysisItem, 0)

	// 获取所有Pod稳定性相关规则
	filter := rules.RuleFilter{
		Categories: []string{"pod"},
	}
	allRules := pa.rulesEngine.GetRules(filter)

	// 检查重启次数
	for _, rule := range allRules {
		if rule.Condition.Metric == "pod_restart_count" {
			// 评估规则
			ruleResult, err := pa.rulesEngine.EvaluateRule(rule, "numeric", float64(pod.TotalRestarts))
			if err != nil {
				// 记录错误并继续
				continue
			}
			
			// 创建分析项
			item := AnalysisItem{
				RuleID:      ruleResult.RuleID,
				Name:        ruleResult.RuleName,
				Category:    rule.Category,
				Severity:    ruleResult.Severity,
				Metric:      "pod_restart_count",
				Value:       fmt.Sprintf("%d", pod.TotalRestarts),
				Threshold:   fmt.Sprintf("%v", ruleResult.ExpectedValue),
				Passed:      !ruleResult.Passed, // 反转结果
				Description: fmt.Sprintf("Pod总重启次数为 %d", pod.TotalRestarts),
				Remediation: ruleResult.Remediation,
			}
			
			items = append(items, item)
		}
	}

	return items
}

// analyzePodConfig 分析Pod配置
func (pa *PodAnalyzer) analyzePodConfig(pod *models.Pod) []AnalysisItem {
	items := make([]AnalysisItem, 0)

	// 获取所有Pod配置相关规则
	filter := rules.RuleFilter{
		Categories: []string{"pod"},
	}
	allRules := pa.rulesEngine.GetRules(filter)

	// 检查是否缺少健康检查
	if !pod.HasLivenessProbe && !pod.HasReadinessProbe {
		for _, rule := range allRules {
			if rule.Condition.Metric == "pod_missing_probes" {
				// 评估规则
				ruleResult, err := pa.rulesEngine.EvaluateRule(rule, "boolean", true)
				if err != nil {
					// 记录错误并继续
					continue
				}
				
				// 创建分析项
				item := AnalysisItem{
					RuleID:      ruleResult.RuleID,
					Name:        ruleResult.RuleName,
					Category:    rule.Category,
					Severity:    ruleResult.Severity,
					Metric:      "pod_missing_probes",
					Value:       "true",
					Threshold:   "false",
					Passed:      !ruleResult.Passed, // 反转结果
					Description: "Pod缺少健康检查探针",
					Remediation: ruleResult.Remediation,
				}
				
				items = append(items, item)
			}
		}
	}

	return items
}

// calculateHealthScore 计算Pod健康评分
func (pa *PodAnalyzer) calculateHealthScore(items []AnalysisItem) int {
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

// formatDuration 格式化时间间隔
func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	
	days := d / (24 * time.Hour)
	d -= days * 24 * time.Hour
	
	hours := d / time.Hour
	d -= hours * time.Hour
	
	minutes := d / time.Minute
	d -= minutes * time.Minute
	
	seconds := d / time.Second
	
	if days > 0 {
		return fmt.Sprintf("%dd%dh%dm", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh%dm%ds", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm%ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

// getContainerState 获取容器状态描述
func getContainerState(state corev1.ContainerState) string {
	if state.Running != nil {
		return "Running"
	}
	if state.Waiting != nil {
		if state.Waiting.Reason != "" {
			return state.Waiting.Reason
		}
		return "Waiting"
	}
	if state.Terminated != nil {
		if state.Terminated.Reason != "" {
			return state.Terminated.Reason
		}
		return fmt.Sprintf("Terminated (exit code: %d)", state.Terminated.ExitCode)
	}
	return "Unknown"
} 