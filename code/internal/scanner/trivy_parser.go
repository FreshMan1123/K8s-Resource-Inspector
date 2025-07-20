package scanner

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// TrivyK8sReport Trivy K8s扫描的原始输出结构
type TrivyK8sReport struct {
	SchemaVersion int                `json:"SchemaVersion"`
	ArtifactName  string             `json:"ArtifactName"`
	ArtifactType  string             `json:"ArtifactType"`
	ClusterName   string             `json:"ClusterName"`
	Results       []TrivyK8sResult   `json:"Results"`       // 镜像扫描格式
	Resources     []TrivyK8sResource `json:"Resources"`     // K8s扫描格式
	Metadata      TrivyMetadata      `json:"Metadata"`
}

// TrivyK8sResource Trivy K8s资源扫描结果
type TrivyK8sResource struct {
	Namespace string                    `json:"Namespace"`
	Kind      string                    `json:"Kind"`
	Name      string                    `json:"Name"`
	Results   []TrivyK8sResult         `json:"Results"`
	Error     string                    `json:"Error"`
	Metadata  map[string]interface{}   `json:"Metadata"`
}

// TrivyK8sResult Trivy K8s扫描结果项
type TrivyK8sResult struct {
	Target             string                    `json:"Target"`
	Class              string                    `json:"Class"`
	Type               string                    `json:"Type"`
	Vulnerabilities    []TrivyVulnerability      `json:"Vulnerabilities"`
	Misconfigurations  []TrivyMisconfiguration   `json:"Misconfigurations"`
	MisconfSummary     *TrivyMisconfSummary      `json:"MisconfSummary"`
}

// TrivyVulnerability Trivy漏洞信息
type TrivyVulnerability struct {
	VulnerabilityID  string     `json:"VulnerabilityID"`
	PkgName          string     `json:"PkgName"`
	InstalledVersion string     `json:"InstalledVersion"`
	FixedVersion     string     `json:"FixedVersion"`
	Severity         string     `json:"Severity"`
	Title            string     `json:"Title"`
	Description      string     `json:"Description"`
	References       []string   `json:"References"`
	PublishedDate    *time.Time `json:"PublishedDate"`
	LastModifiedDate *time.Time `json:"LastModifiedDate"`
	CVSS             TrivyCVSS  `json:"CVSS"`
}

// TrivyMisconfiguration Trivy配置问题
type TrivyMisconfiguration struct {
	Type        string            `json:"Type"`
	ID          string            `json:"ID"`
	Title       string            `json:"Title"`
	Description string            `json:"Description"`
	Message     string            `json:"Message"`
	Resolution  string            `json:"Resolution"`
	Severity    string            `json:"Severity"`
	Status      string            `json:"Status"`
	References  []string          `json:"References"`
	Metadata    map[string]interface{} `json:"CauseMetadata"`
}

// TrivyMisconfSummary 配置问题摘要
type TrivyMisconfSummary struct {
	Successes  int `json:"Successes"`
	Failures   int `json:"Failures"`
	Exceptions int `json:"Exceptions"`
}

// TrivyCVSS CVSS评分信息
type TrivyCVSS struct {
	Nvd    TrivyCVSSScore `json:"nvd"`
	RedHat TrivyCVSSScore `json:"redhat"`
}

// TrivyCVSSScore CVSS评分详情
type TrivyCVSSScore struct {
	V3Score  *float64 `json:"V3Score"`
	V2Score  *float64 `json:"V2Score"`
	V3Vector string   `json:"V3Vector"`
	V2Vector string   `json:"V2Vector"`
}

// TrivyMetadata Trivy元数据
type TrivyMetadata struct {
	OS           *TrivyOS      `json:"OS"`
	ImageID      string        `json:"ImageID"`
	DiffIDs      []string      `json:"DiffIDs"`
	RepoTags     []string      `json:"RepoTags"`
	RepoDigests  []string      `json:"RepoDigests"`
	ImageConfig  interface{}   `json:"ImageConfig"`
}

// TrivyOS 操作系统信息
type TrivyOS struct {
	Family string `json:"Family"`
	Name   string `json:"Name"`
}

// parseTrivyK8sOutput 解析Trivy K8s扫描输出
func (ts *TrivyScanner) parseTrivyK8sOutput(output []byte) (*VulnerabilityReport, error) {
	// 调试：输出原始JSON数据
	fmt.Printf("🔍 调试: Trivy原始输出长度: %d bytes\n", len(output))
	if len(output) > 0 && len(output) < 10000 {
		fmt.Printf("🔍 调试: Trivy原始输出内容:\n%s\n", string(output))
	}

	var trivyReport TrivyK8sReport
	if err := json.Unmarshal(output, &trivyReport); err != nil {
		return nil, fmt.Errorf("解析Trivy K8s输出失败: %w", err)
	}

	// 调试：输出解析后的结构
	fmt.Printf("🔍 调试: 解析后的结果数量: %d\n", len(trivyReport.Results))

	report := &VulnerabilityReport{
		ScanTime:     time.Now(),
		Scanner:      ts.GetScannerInfo(),
		Status:       "completed",
		PodReports:   make(map[string]*PodVulnReport),
		ConfigIssues: make([]ConfigurationIssue, 0),
	}

	// 解析结果 - 处理两种格式
	// 1. 处理 Results 数组 (镜像扫描格式)
	for i, result := range trivyReport.Results {
		fmt.Printf("🔍 调试: 处理Results结果 %d, Target: %s, Class: %s, Type: %s\n",
			i, result.Target, result.Class, result.Type)
		fmt.Printf("🔍 调试: 漏洞数量: %d, 配置问题数量: %d\n",
			len(result.Vulnerabilities), len(result.Misconfigurations))

		if err := ts.processK8sResult(&result, report); err != nil {
			fmt.Printf("警告: 处理扫描结果失败: %v\n", err)
			continue
		}
	}

	// 2. 处理 Resources 数组 (K8s扫描格式)
	for i, resource := range trivyReport.Resources {
		fmt.Printf("🔍 调试: 处理Resources资源 %d, Namespace: %s, Kind: %s, Name: %s\n",
			i, resource.Namespace, resource.Kind, resource.Name)
		fmt.Printf("🔍 调试: 资源错误: %s, 结果数量: %d\n",
			resource.Error, len(resource.Results))

		if resource.Error != "" {
			fmt.Printf("警告: 资源扫描失败: %s\n", resource.Error)
			continue
		}

		// 处理资源中的每个结果
		for j, result := range resource.Results {
			fmt.Printf("🔍 调试: 处理资源结果 %d-%d, Target: %s, Class: %s, Type: %s\n",
				i, j, result.Target, result.Class, result.Type)

			// 为K8s资源构造目标字符串
			if result.Target == "" {
				result.Target = fmt.Sprintf("%s/%s", resource.Namespace, resource.Name)
			}

			if err := ts.processK8sResult(&result, report); err != nil {
				fmt.Printf("警告: 处理资源扫描结果失败: %v\n", err)
				continue
			}
		}
	}

	// 计算统计摘要
	ts.calculateSummary(report)

	return report, nil
}

// parseTrivyImageOutput 解析Trivy镜像扫描输出
func (ts *TrivyScanner) parseTrivyImageOutput(output []byte, imageRef string) (*VulnerabilityReport, error) {
	var trivyReport TrivyK8sReport // 使用相同的结构
	if err := json.Unmarshal(output, &trivyReport); err != nil {
		return nil, fmt.Errorf("解析Trivy镜像输出失败: %w", err)
	}

	report := &VulnerabilityReport{
		ScanTime:     time.Now(),
		Scanner:      ts.GetScannerInfo(),
		Status:       "completed",
		PodReports:   make(map[string]*PodVulnReport),
		ConfigIssues: make([]ConfigurationIssue, 0),
	}

	// 为镜像扫描创建虚拟Pod报告
	podReport := &PodVulnReport{
		PodName:    "image-scan",
		Namespace:  "virtual",
		Containers: make(map[string]*ContainerVulnReport),
	}

	containerReport := &ContainerVulnReport{
		ContainerName:   "image",
		Image:          imageRef,
		Vulnerabilities: make([]Vulnerability, 0),
	}

	// 处理漏洞
	for _, result := range trivyReport.Results {
		for _, vuln := range result.Vulnerabilities {
			vulnerability := ts.convertTrivyVulnerability(&vuln)
			containerReport.Vulnerabilities = append(containerReport.Vulnerabilities, vulnerability)
		}
	}

	// 计算容器统计
	ts.calculateContainerSummary(containerReport)
	podReport.Containers["image"] = containerReport
	
	// 计算Pod统计
	ts.calculatePodSummary(podReport)
	report.PodReports["image-scan"] = podReport

	// 计算总体统计
	ts.calculateSummary(report)

	return report, nil
}

// processK8sResult 处理K8s扫描结果项
func (ts *TrivyScanner) processK8sResult(result *TrivyK8sResult, report *VulnerabilityReport) error {
	// 解析目标信息 (例如: "default/nginx-deployment-xxx (nginx)")
	podName, containerName := ts.parseK8sTarget(result.Target)
	fmt.Printf("🔍 调试: parseK8sTarget结果 - podName: '%s', containerName: '%s'\n", podName, containerName)

	if podName == "" {
		fmt.Printf("🔍 调试: podName为空，检查是否为配置检查结果\n")
		// 如果是配置检查结果
		if result.Class == "config" || result.Class == "config-misconfig" {
			ts.processConfigIssues(result, report)
		}
		return nil
	}

	// 获取或创建Pod报告
	podReport, exists := report.PodReports[podName]
	if !exists {
		podReport = &PodVulnReport{
			PodName:    podName,
			Namespace:  ts.extractNamespace(result.Target),
			Containers: make(map[string]*ContainerVulnReport),
		}
		report.PodReports[podName] = podReport
	}

	// 获取或创建容器报告
	containerReport, exists := podReport.Containers[containerName]
	if !exists {
		containerReport = &ContainerVulnReport{
			ContainerName:   containerName,
			Image:          ts.extractImageName(result.Target),
			Vulnerabilities: make([]Vulnerability, 0),
		}
		podReport.Containers[containerName] = containerReport
	}

	// 处理漏洞
	for _, vuln := range result.Vulnerabilities {
		vulnerability := ts.convertTrivyVulnerability(&vuln)
		containerReport.Vulnerabilities = append(containerReport.Vulnerabilities, vulnerability)
	}

	// 处理配置问题
	ts.processConfigIssues(result, report)

	return nil
}

// processConfigIssues 处理配置问题
func (ts *TrivyScanner) processConfigIssues(result *TrivyK8sResult, report *VulnerabilityReport) {
	for _, misconfig := range result.Misconfigurations {
		if misconfig.Status == "FAIL" {
			configIssue := ConfigurationIssue{
				ID:         misconfig.ID,
				Title:      misconfig.Title,
				Severity:   misconfig.Severity,
				Resource:   result.Target,
				Message:    misconfig.Message,
				Resolution: misconfig.Resolution,
			}
			report.ConfigIssues = append(report.ConfigIssues, configIssue)
		}
	}
}

// convertTrivyVulnerability 转换Trivy漏洞为标准格式
func (ts *TrivyScanner) convertTrivyVulnerability(vuln *TrivyVulnerability) Vulnerability {
	// 获取CVSS评分
	cvssScore := 0.0
	if vuln.CVSS.Nvd.V3Score != nil {
		cvssScore = *vuln.CVSS.Nvd.V3Score
	} else if vuln.CVSS.Nvd.V2Score != nil {
		cvssScore = *vuln.CVSS.Nvd.V2Score
	} else if vuln.CVSS.RedHat.V3Score != nil {
		cvssScore = *vuln.CVSS.RedHat.V3Score
	} else if vuln.CVSS.RedHat.V2Score != nil {
		cvssScore = *vuln.CVSS.RedHat.V2Score
	}

	return Vulnerability{
		ID:               vuln.VulnerabilityID,
		Title:            vuln.Title,
		Severity:         vuln.Severity,
		CVSS:             cvssScore,
		PkgName:          vuln.PkgName,
		InstalledVersion: vuln.InstalledVersion,
		FixedVersion:     vuln.FixedVersion,
		PublishedDate:    vuln.PublishedDate,
		References:       vuln.References,
	}
}

// parseK8sTarget 解析K8s目标字符串
func (ts *TrivyScanner) parseK8sTarget(target string) (podName, containerName string) {
	fmt.Printf("🔍 调试: 解析目标字符串: '%s'\n", target)

	// 处理不同的目标格式
	if strings.Contains(target, " (") {
		// 格式: "default/nginx-deployment-xxx (nginx)"
		parts := strings.Split(target, " ")
		if len(parts) >= 2 {
			// 提取Pod名称
			namespacePod := parts[0]
			if strings.Contains(namespacePod, "/") {
				podName = strings.Split(namespacePod, "/")[1]
			} else {
				podName = namespacePod
			}

			// 提取容器名称
			containerPart := parts[1]
			containerName = strings.Trim(containerPart, "()")
		}
	} else if strings.Contains(target, "/") {
		// 格式: "default/nginx-deployment-xxx"
		parts := strings.Split(target, "/")
		if len(parts) >= 2 {
			podName = parts[1]
			containerName = "unknown" // 默认容器名
		}
	} else {
		// 格式: "nginx-deployment-xxx"
		podName = target
		containerName = "unknown"
	}

	fmt.Printf("🔍 调试: 解析结果 - podName: '%s', containerName: '%s'\n", podName, containerName)
	return podName, containerName
}

// extractNamespace 从目标字符串中提取命名空间
func (ts *TrivyScanner) extractNamespace(target string) string {
	parts := strings.Split(target, "/")
	if len(parts) >= 2 {
		return parts[0]
	}
	return "default"
}

// extractImageName 从目标字符串中提取镜像名称（简化实现）
func (ts *TrivyScanner) extractImageName(target string) string {
	// 这里简化处理，实际可能需要更复杂的解析
	return "unknown"
}

// calculateSummary 计算总体统计摘要
func (ts *TrivyScanner) calculateSummary(report *VulnerabilityReport) {
	summary := VulnerabilitySummary{}
	var maxCVSS float64
	var totalCVSS float64
	var cvssCount int

	// 遍历所有Pod报告
	for _, podReport := range report.PodReports {
		ts.calculatePodSummary(podReport)

		// 累加统计
		summary.Total += podReport.Summary.Total
		summary.Critical += podReport.Summary.Critical
		summary.High += podReport.Summary.High
		summary.Medium += podReport.Summary.Medium
		summary.Low += podReport.Summary.Low
		summary.Fixable += podReport.Summary.Fixable
		summary.Unfixable += podReport.Summary.Unfixable

		// 更新最大CVSS
		if podReport.Summary.MaxCVSS > maxCVSS {
			maxCVSS = podReport.Summary.MaxCVSS
		}

		// 累加CVSS用于计算平均值
		totalCVSS += podReport.Summary.AvgCVSS * float64(podReport.Summary.Total)
		cvssCount += podReport.Summary.Total
	}

	// 计算平均CVSS
	if cvssCount > 0 {
		summary.AvgCVSS = totalCVSS / float64(cvssCount)
	}
	summary.MaxCVSS = maxCVSS

	report.Summary = summary
}

// calculatePodSummary 计算Pod级别统计摘要
func (ts *TrivyScanner) calculatePodSummary(podReport *PodVulnReport) {
	summary := VulnerabilitySummary{}
	var maxCVSS float64
	var totalCVSS float64
	var cvssCount int

	// 遍历所有容器报告
	for _, containerReport := range podReport.Containers {
		ts.calculateContainerSummary(containerReport)

		// 累加统计
		summary.Total += containerReport.Summary.Total
		summary.Critical += containerReport.Summary.Critical
		summary.High += containerReport.Summary.High
		summary.Medium += containerReport.Summary.Medium
		summary.Low += containerReport.Summary.Low
		summary.Fixable += containerReport.Summary.Fixable
		summary.Unfixable += containerReport.Summary.Unfixable

		// 更新最大CVSS
		if containerReport.Summary.MaxCVSS > maxCVSS {
			maxCVSS = containerReport.Summary.MaxCVSS
		}

		// 累加CVSS用于计算平均值
		totalCVSS += containerReport.Summary.AvgCVSS * float64(containerReport.Summary.Total)
		cvssCount += containerReport.Summary.Total
	}

	// 计算平均CVSS
	if cvssCount > 0 {
		summary.AvgCVSS = totalCVSS / float64(cvssCount)
	}
	summary.MaxCVSS = maxCVSS

	podReport.Summary = summary
}

// calculateContainerSummary 计算容器级别统计摘要
func (ts *TrivyScanner) calculateContainerSummary(containerReport *ContainerVulnReport) {
	summary := VulnerabilitySummary{
		Total: len(containerReport.Vulnerabilities),
	}

	var maxCVSS float64
	var totalCVSS float64
	var cvssCount int

	// 统计各级别漏洞
	for _, vuln := range containerReport.Vulnerabilities {
		switch vuln.Severity {
		case "CRITICAL":
			summary.Critical++
		case "HIGH":
			summary.High++
		case "MEDIUM":
			summary.Medium++
		case "LOW":
			summary.Low++
		}

		// 统计可修复性
		if vuln.FixedVersion != "" {
			summary.Fixable++
		} else {
			summary.Unfixable++
		}

		// 统计CVSS
		if vuln.CVSS > 0 {
			if vuln.CVSS > maxCVSS {
				maxCVSS = vuln.CVSS
			}
			totalCVSS += vuln.CVSS
			cvssCount++
		}
	}

	// 计算平均CVSS
	if cvssCount > 0 {
		summary.AvgCVSS = totalCVSS / float64(cvssCount)
	}
	summary.MaxCVSS = maxCVSS

	containerReport.Summary = summary
}
