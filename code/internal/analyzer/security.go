package analyzer

import (
	"context"
	"fmt"

	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/collector"
	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/models"
	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/rules"
)

// SecurityAnalyzer 安全分析器
type SecurityAnalyzer struct {
	podCollector *collector.PodCollector
	engine       *rules.Engine
}

// NewSecurityAnalyzer 创建新的安全分析器
func NewSecurityAnalyzer(podCollector *collector.PodCollector, engine *rules.Engine) *SecurityAnalyzer {
	return &SecurityAnalyzer{
		podCollector: podCollector,
		engine:       engine,
	}
}

// AnalyzePodSecurity 分析Pod安全配置
func (sa *SecurityAnalyzer) AnalyzePodSecurity(ctx context.Context, namespace, podName string) (*models.SecurityAnalysisResult, error) {
	// 获取Pod安全信息
	podSecurityInfo, err := sa.podCollector.GetPodSecurityInfo(ctx, namespace, podName)
	if err != nil {
		return nil, fmt.Errorf("获取Pod安全信息失败: %w", err)
	}

	// 创建分析结果
	result := &models.SecurityAnalysisResult{
		PodName:   podName,
		Namespace: namespace,
		Items:     make([]models.SecurityAnalysisItem, 0),
	}

	// 分析Pod安全配置并转换为标准分析项
	result.Items = sa.analyzePodSecurityWithRules(podSecurityInfo)

	return result, nil
}

// AnalyzeNamespaceSecurity 分析命名空间安全配置
func (sa *SecurityAnalyzer) AnalyzeNamespaceSecurity(ctx context.Context, namespace string) ([]*models.SecurityAnalysisResult, error) {
	// 获取命名空间中所有Pod的安全信息
	podsSecurityInfo, err := sa.podCollector.GetPodsSecurityInfo(ctx, namespace)
	if err != nil {
		return nil, fmt.Errorf("获取命名空间Pod安全信息失败: %w", err)
	}

	// 创建分析结果列表
	results := make([]*models.SecurityAnalysisResult, 0, len(podsSecurityInfo))

	// 分析每个Pod的安全配置
	for _, podSecurity := range podsSecurityInfo {
		result := &models.SecurityAnalysisResult{
			PodName:   podSecurity.Name,
			Namespace: podSecurity.Namespace,
			Items:     sa.analyzePodSecurityWithRules(&podSecurity),
		}
		results = append(results, result)
	}

	return results, nil
}

// analyzePodSecurityWithRules 使用规则引擎分析Pod安全配置
func (sa *SecurityAnalyzer) analyzePodSecurityWithRules(podSecurity *models.PodSecurityInfo) []models.SecurityAnalysisItem {
	items := make([]models.SecurityAnalysisItem, 0)

	// 获取安全相关规则
	enabled := true
	filter := rules.RuleFilter{
		Categories: []string{"pod"},
		Enabled:    &enabled,
	}
	securityRules := sa.engine.GetRules(filter)

	// 为每个规则评估Pod安全配置
	for _, rule := range securityRules {
		item := sa.evaluateSecurityRule(rule, podSecurity)
		if item != nil {
			items = append(items, *item)
		}
	}

	return items
}

// evaluateSecurityRule 评估单个安全规则
func (sa *SecurityAnalyzer) evaluateSecurityRule(rule rules.Rule, podSecurity *models.PodSecurityInfo) *models.SecurityAnalysisItem {
	var actualValue interface{}
	var metricType string

	// 根据规则条件获取实际值
	switch rule.Condition.Metric {
	case "service_account_name":
		actualValue = podSecurity.ServiceAccountName
		metricType = "cis_compliance"
	case "host_network":
		actualValue = podSecurity.HostNetwork
		metricType = "security_context"
	case "host_pid_ipc":
		actualValue = podSecurity.HostPID || podSecurity.HostIPC
		metricType = "security_context"
	case "automount_service_account_token":
		if podSecurity.AutomountServiceAccountToken != nil {
			actualValue = *podSecurity.AutomountServiceAccountToken
		} else {
			actualValue = true // Kubernetes默认值
		}
		metricType = "cis_compliance"
	case "mounts_sensitive_host_paths":
		actualValue = sa.hasSensitiveHostPathMounts(podSecurity.HostPathMounts)
		metricType = "security_context"
	default:
		// 检查容器级别的安全配置
		return sa.evaluateContainerSecurityRule(rule, podSecurity)
	}

	// 使用规则引擎评估
	result, err := sa.engine.EvaluateRule(rule, metricType, actualValue)
	if err != nil {
		// 记录错误但继续处理其他规则
		fmt.Printf("警告: 评估规则 %s 失败: %v\n", rule.ID, err)
		return nil
	}

	// 转换为SecurityAnalysisItem
	return &models.SecurityAnalysisItem{
		RuleID:      rule.ID,
		Metric:      rule.Condition.Metric,
		Value:       actualValue,
		Threshold:   result.ExpectedValue,
		Passed:      result.Passed,
		Severity:    rule.Severity,
		Description: rule.Description,
		Remediation: rule.Remediation,
	}
}

// evaluateContainerSecurityRule 评估容器级别的安全规则
func (sa *SecurityAnalyzer) evaluateContainerSecurityRule(rule rules.Rule, podSecurity *models.PodSecurityInfo) *models.SecurityAnalysisItem {
	// 检查所有容器（包括初始化容器）
	allContainers := append(podSecurity.Containers, podSecurity.InitContainers...)

	for _, container := range allContainers {
		item := sa.evaluateContainerRule(rule, &container, podSecurity)
		if item != nil && !item.Passed {
			// 如果任何容器不通过规则，返回失败结果
			return item
		}
	}

	// 所有容器都通过规则
	if len(allContainers) > 0 {
		return &models.SecurityAnalysisItem{
			RuleID:      rule.ID,
			Metric:      rule.Condition.Metric,
			Value:       "all_containers_compliant",
			Threshold:   rule.Condition.Threshold,
			Passed:      true,
			Severity:    rule.Severity,
			Description: rule.Description,
			Remediation: rule.Remediation,
		}
	}

	return nil
}

// evaluateContainerRule 评估单个容器的安全规则
func (sa *SecurityAnalyzer) evaluateContainerRule(rule rules.Rule, container *models.ContainerSecurityInfo, podSecurity *models.PodSecurityInfo) *models.SecurityAnalysisItem {
	var actualValue interface{}
	var metricType string

	switch rule.Condition.Metric {
	case "privileged":
		if container.SecurityContext != nil && container.SecurityContext.Privileged != nil {
			actualValue = *container.SecurityContext.Privileged
		} else {
			actualValue = false
		}
		metricType = "security_context"
	case "allow_privilege_escalation":
		if container.SecurityContext != nil && container.SecurityContext.AllowPrivilegeEscalation != nil {
			actualValue = *container.SecurityContext.AllowPrivilegeEscalation
		} else {
			actualValue = true // Kubernetes默认值
		}
		metricType = "security_context"
	case "run_as_non_root":
		if container.SecurityContext != nil && container.SecurityContext.RunAsNonRoot != nil {
			actualValue = *container.SecurityContext.RunAsNonRoot
		} else {
			actualValue = false
		}
		metricType = "security_context"
	case "read_only_root_filesystem":
		if container.SecurityContext != nil && container.SecurityContext.ReadOnlyRootFilesystem != nil {
			actualValue = *container.SecurityContext.ReadOnlyRootFilesystem
		} else {
			actualValue = false
		}
		metricType = "security_context"
	case "image_tag":
		if container.Image != nil {
			actualValue = container.Image.Tag
		} else {
			actualValue = ""
		}
		metricType = "image_security"
	case "image_registry":
		if container.Image != nil {
			actualValue = container.Image.Registry
		} else {
			actualValue = ""
		}
		metricType = "image_security"
	case "has_resource_limits":
		actualValue = len(container.ResourceLimits) > 0
		metricType = "boolean"
	case "has_liveness_probe":
		actualValue = container.LivenessProbe
		metricType = "boolean"
	case "has_readiness_probe":
		actualValue = container.ReadinessProbe
		metricType = "boolean"
	default:
		return nil
	}

	// 使用规则引擎评估
	result, err := sa.engine.EvaluateRule(rule, metricType, actualValue)
	if err != nil {
		fmt.Printf("警告: 评估容器规则 %s 失败: %v\n", rule.ID, err)
		return nil
	}

	// 转换为SecurityAnalysisItem
	return &models.SecurityAnalysisItem{
		RuleID:      rule.ID,
		Metric:      rule.Condition.Metric,
		Value:       actualValue,
		Threshold:   result.ExpectedValue,
		Passed:      result.Passed,
		Severity:    rule.Severity,
		Description: fmt.Sprintf("%s (容器: %s)", rule.Description, container.Name),
		Remediation: rule.Remediation,
	}
}

// hasSensitiveHostPathMounts 检查是否有敏感的主机路径挂载
func (sa *SecurityAnalyzer) hasSensitiveHostPathMounts(mounts []models.HostPathMount) bool {
	for _, mount := range mounts {
		if mount.IsSensitive {
			return true
		}
	}
	return false
}






