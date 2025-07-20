package analyzer

import (
	"context"
	"fmt"

	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/collector"
	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/models"
	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/rules"
	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/scanner"
)

// SecurityAnalyzer 安全分析器
type SecurityAnalyzer struct {
	podCollector         *collector.PodCollector
	engine              *rules.Engine
	vulnerabilityScanner scanner.VulnerabilityScanner // 新增漏洞扫描器
}

// NewSecurityAnalyzer 创建新的安全分析器
func NewSecurityAnalyzer(podCollector *collector.PodCollector, engine *rules.Engine) *SecurityAnalyzer {
	return &SecurityAnalyzer{
		podCollector: podCollector,
		engine:       engine,
	}
}

// NewSecurityAnalyzerWithVulnScanner 创建带漏洞扫描器的安全分析器
func NewSecurityAnalyzerWithVulnScanner(podCollector *collector.PodCollector, engine *rules.Engine, vulnScanner scanner.VulnerabilityScanner) *SecurityAnalyzer {
	return &SecurityAnalyzer{
		podCollector:         podCollector,
		engine:              engine,
		vulnerabilityScanner: vulnScanner,
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

// AnalyzeAllNamespacesSecurity 分析所有命名空间的安全配置
func (sa *SecurityAnalyzer) AnalyzeAllNamespacesSecurity(ctx context.Context) ([]*models.SecurityAnalysisResult, error) {
	// 定义要检查的命名空间
	namespaces := []string{"default", "kube-system", "kube-public", "kube-node-lease"}

	allResults := make([]*models.SecurityAnalysisResult, 0)

	// 分析每个命名空间
	for _, namespace := range namespaces {
		results, err := sa.AnalyzeNamespaceSecurity(ctx, namespace)
		if err != nil {
			// 记录错误但继续处理其他命名空间
			fmt.Printf("警告: 分析命名空间 %s 失败: %v\n", namespace, err)
			continue
		}
		allResults = append(allResults, results...)
	}

	return allResults, nil
}

// AnalyzeWithVulnerabilities 分析安全配置并包含漏洞扫描
func (sa *SecurityAnalyzer) AnalyzeWithVulnerabilities(ctx context.Context, namespace string) ([]*models.SecurityAnalysisResult, error) {
	// 先执行常规安全分析
	results, err := sa.AnalyzeNamespaceSecurity(ctx, namespace)
	if err != nil {
		return nil, err
	}

	// 如果有漏洞扫描器，执行漏洞扫描
	if sa.vulnerabilityScanner != nil && sa.vulnerabilityScanner.IsAvailable() {
		fmt.Printf("🔍 正在执行漏洞扫描...\n")

		vulnReport, err := sa.vulnerabilityScanner.ScanNamespace(ctx, namespace)
		if err != nil {
			fmt.Printf("警告: 漏洞扫描失败: %v\n", err)
		} else {
			// 将漏洞信息集成到安全分析结果中
			sa.integrateVulnerabilityResults(results, vulnReport)
		}
	}

	return results, nil
}

// AnalyzeVulnerabilitiesOnly 仅执行漏洞扫描
func (sa *SecurityAnalyzer) AnalyzeVulnerabilitiesOnly(ctx context.Context, namespace string) (*scanner.VulnerabilityReport, error) {
	if sa.vulnerabilityScanner == nil {
		return nil, fmt.Errorf("漏洞扫描器未配置")
	}

	if !sa.vulnerabilityScanner.IsAvailable() {
		return nil, fmt.Errorf("漏洞扫描器不可用")
	}

	return sa.vulnerabilityScanner.ScanNamespace(ctx, namespace)
}

// integrateVulnerabilityResults 将漏洞扫描结果集成到安全分析结果中
func (sa *SecurityAnalyzer) integrateVulnerabilityResults(results []*models.SecurityAnalysisResult, vulnReport *scanner.VulnerabilityReport) {
	// 为每个安全分析结果添加漏洞信息
	for _, result := range results {
		// 查找对应的Pod漏洞报告
		if podVulnReport, exists := vulnReport.PodReports[result.PodName]; exists {
			// 将漏洞转换为安全分析项
			vulnItems := sa.convertVulnerabilityToSecurityItems(podVulnReport)
			result.Items = append(result.Items, vulnItems...)
		}
	}
}

// convertVulnerabilityToSecurityItems 将漏洞报告转换为安全分析项
func (sa *SecurityAnalyzer) convertVulnerabilityToSecurityItems(podVulnReport *scanner.PodVulnReport) []models.SecurityAnalysisItem {
	var items []models.SecurityAnalysisItem

	// 检查严重漏洞
	if podVulnReport.Summary.Critical > 0 {
		items = append(items, models.SecurityAnalysisItem{
			RuleID:      "vuln-critical",
			Metric:      "critical_vulnerabilities",
			Value:       podVulnReport.Summary.Critical,
			Threshold:   0,
			Passed:      false,
			Severity:    "critical",
			Description: fmt.Sprintf("发现 %d 个严重漏洞", podVulnReport.Summary.Critical),
			Remediation: "立即更新镜像到安全版本",
		})
	}

	// 检查高危漏洞
	if podVulnReport.Summary.High > 5 {
		items = append(items, models.SecurityAnalysisItem{
			RuleID:      "vuln-high",
			Metric:      "high_vulnerabilities",
			Value:       podVulnReport.Summary.High,
			Threshold:   5,
			Passed:      false,
			Severity:    "error",
			Description: fmt.Sprintf("发现 %d 个高危漏洞 (超过阈值 5)", podVulnReport.Summary.High),
			Remediation: "尽快更新镜像版本",
		})
	}

	// 检查中危漏洞
	if podVulnReport.Summary.Medium > 10 {
		items = append(items, models.SecurityAnalysisItem{
			RuleID:      "vuln-medium",
			Metric:      "medium_vulnerabilities",
			Value:       podVulnReport.Summary.Medium,
			Threshold:   10,
			Passed:      false,
			Severity:    "warning",
			Description: fmt.Sprintf("发现 %d 个中危漏洞 (超过阈值 10)", podVulnReport.Summary.Medium),
			Remediation: "建议更新镜像版本",
		})
	}

	// 检查不可修复的漏洞
	if podVulnReport.Summary.Unfixable > 0 {
		items = append(items, models.SecurityAnalysisItem{
			RuleID:      "vuln-unfixable",
			Metric:      "unfixable_vulnerabilities",
			Value:       podVulnReport.Summary.Unfixable,
			Threshold:   0,
			Passed:      false,
			Severity:    "warning",
			Description: fmt.Sprintf("发现 %d 个不可修复的漏洞", podVulnReport.Summary.Unfixable),
			Remediation: "考虑使用替代镜像或接受风险",
		})
	}

	return items
}

// analyzePodSecurityWithRules 使用规则引擎分析Pod安全配置
func (sa *SecurityAnalyzer) analyzePodSecurityWithRules(podSecurity *models.PodSecurityInfo) []models.SecurityAnalysisItem {
	items := make([]models.SecurityAnalysisItem, 0)

	// 获取安全相关规则
	enabled := true
	filter := rules.RuleFilter{
		Categories: []string{"pod", "image"},
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
		metricType = "string"
	case "host_network":
		actualValue = podSecurity.HostNetwork
		metricType = "boolean"
	case "host_pid_ipc":
		actualValue = podSecurity.HostPID || podSecurity.HostIPC
		metricType = "boolean"
	case "automount_service_account_token":
		if podSecurity.AutomountServiceAccountToken != nil {
			actualValue = *podSecurity.AutomountServiceAccountToken
		} else {
			actualValue = true // Kubernetes默认值
		}
		metricType = "boolean"
	case "mounts_sensitive_host_paths":
		actualValue = sa.hasSensitiveHostPathMounts(podSecurity.HostPathMounts)
		metricType = "boolean"
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
		metricType = "boolean"
	case "allow_privilege_escalation":
		if container.SecurityContext != nil && container.SecurityContext.AllowPrivilegeEscalation != nil {
			actualValue = *container.SecurityContext.AllowPrivilegeEscalation
		} else {
			actualValue = true // Kubernetes默认值
		}
		metricType = "boolean"
	case "run_as_non_root":
		if container.SecurityContext != nil && container.SecurityContext.RunAsNonRoot != nil {
			actualValue = *container.SecurityContext.RunAsNonRoot
		} else {
			actualValue = false
		}
		metricType = "boolean"
	case "read_only_root_filesystem":
		if container.SecurityContext != nil && container.SecurityContext.ReadOnlyRootFilesystem != nil {
			actualValue = *container.SecurityContext.ReadOnlyRootFilesystem
		} else {
			actualValue = false
		}
		metricType = "boolean"
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






