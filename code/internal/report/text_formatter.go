package report

import (
	"fmt"
	"strings"
	"time"
)

// TextFormatter 实现了用于文本输出的Formatter接口
type TextFormatter struct {
	// ColorEnabled 决定是否使用终端颜色代码
	ColorEnabled bool
}

// NewTextFormatter 创建一个新的文本格式化器
func NewTextFormatter(colorEnabled bool) Formatter {
	return &TextFormatter{
		ColorEnabled: colorEnabled,
	}
}

// Format 将报告转换为人类可读的文本表示
func (f *TextFormatter) Format(report *Report) string {
	var sb strings.Builder

	// 添加报告头部
	f.writeHeader(&sb, report)
	
	// 添加节点详细信息部分
	f.writeNodeDetails(&sb, report)
	
	// 添加摘要部分
	f.writeSummary(&sb, report)
	
	// 添加发现项部分
	f.writeFindings(&sb, report)
	
	return sb.String()
}

// writeHeader 添加报告头部部分到字符串构建器
func (f *TextFormatter) writeHeader(sb *strings.Builder, report *Report) {
	sb.WriteString("========================================\n")
	sb.WriteString("    K8S RESOURCE INSPECTOR REPORT\n")
	sb.WriteString("========================================\n\n")
	
	sb.WriteString(fmt.Sprintf("Generated: %s\n", report.Timestamp.Format(time.RFC1123)))
	
	if report.ClusterName != "" {
		sb.WriteString(fmt.Sprintf("Cluster:   %s\n", report.ClusterName))
	}
	
	if report.Namespace != "" {
		sb.WriteString(fmt.Sprintf("Namespace: %s\n", report.Namespace))
	}
	
	sb.WriteString("\n")
}

// writeNodeDetails 添加节点详细信息部分到字符串构建器
func (f *TextFormatter) writeNodeDetails(sb *strings.Builder, report *Report) {
	if len(report.NodeDetails) == 0 {
		return
	}
	
	sb.WriteString("NODE DETAILS\n")
	sb.WriteString("----------------------------------------\n\n")
	
	for _, node := range report.NodeDetails {
		sb.WriteString(fmt.Sprintf("Node: %s\n", node.Name))
		sb.WriteString(fmt.Sprintf("Status: %s\n", getNodeStatusString(node.Ready)))
		sb.WriteString(fmt.Sprintf("CPU Utilization: %.2f%%\n", node.CPUUtilization))
		sb.WriteString(fmt.Sprintf("Memory Utilization: %.2f%%\n", node.MemoryUtilization))
		if node.PodUtilization > 0 {
			sb.WriteString(fmt.Sprintf("Pod Count: %d/%d (%.2f%%)\n", node.RunningPods, node.MaxPods, node.PodUtilization))
		}
		sb.WriteString(fmt.Sprintf("Health Score: %d/100\n", node.HealthScore))
		sb.WriteString("\n")
	}
}

// getNodeStatusString 根据节点就绪状态返回状态字符串
func getNodeStatusString(ready bool) string {
	if ready {
		return "Ready"
	}
	return "NotReady"
}

// writeSummary 添加摘要部分到字符串构建器
func (f *TextFormatter) writeSummary(sb *strings.Builder, report *Report) {
	sb.WriteString("SUMMARY\n")
	sb.WriteString("----------------------------------------\n")
	sb.WriteString(fmt.Sprintf("Total resources analyzed:    %d\n", report.Summary.TotalResources))
	sb.WriteString(fmt.Sprintf("Resources with issues:       %d\n", report.Summary.ResourcesWithIssues))
	sb.WriteString("\n")
	
	sb.WriteString("Issue severity breakdown:\n")
	f.writeSeverityCount(sb, "CRITICAL", report.Summary.FindingCounts[SeverityCritical])
	f.writeSeverityCount(sb, "ERROR", report.Summary.FindingCounts[SeverityError])
	f.writeSeverityCount(sb, "WARNING", report.Summary.FindingCounts[SeverityWarning])
	f.writeSeverityCount(sb, "INFO", report.Summary.FindingCounts[SeverityInfo])
	
	sb.WriteString("\n")
}

// writeSeverityCount 格式化并写入特定严重性级别的计数
func (f *TextFormatter) writeSeverityCount(sb *strings.Builder, level string, count int) {
	prefix := fmt.Sprintf("  %-8s", level)
	
	if f.ColorEnabled {
		var colorCode string
		switch level {
		case "CRITICAL":
			colorCode = "\033[1;31m" // 粗体红色
		case "ERROR":
			colorCode = "\033[31m" // 红色
		case "WARNING":
			colorCode = "\033[33m" // 黄色
		case "INFO":
			colorCode = "\033[36m" // 青色
		}
		
		if colorCode != "" {
			prefix = fmt.Sprintf("%s%s\033[0m", colorCode, prefix)
		}
	}
	
	sb.WriteString(fmt.Sprintf("%s %d\n", prefix, count))
}

// writeFindings 添加发现项部分到字符串构建器
func (f *TextFormatter) writeFindings(sb *strings.Builder, report *Report) {
	if len(report.Findings) == 0 {
		sb.WriteString("No issues found.\n")
		return
	}
	
	sb.WriteString("FINDINGS\n")
	sb.WriteString("----------------------------------------\n\n")
	
	// 按资源对发现项进行分组
	resourceFindings := make(map[string][]Finding)
	for _, finding := range report.Findings {
		key := fmt.Sprintf("%s/%s", finding.ResourceKind, finding.ResourceName)
		resourceFindings[key] = append(resourceFindings[key], finding)
	}
	
	// 打印每个资源的发现项
	resourceCount := 0
	for resource, findings := range resourceFindings {
		if resourceCount > 0 {
			sb.WriteString("\n")
		}
		
		sb.WriteString(fmt.Sprintf("Resource: %s\n", resource))
		sb.WriteString(fmt.Sprintf("%s\n", strings.Repeat("-", len("Resource: "+resource))))
		
		for i, finding := range findings {
			if i > 0 {
				sb.WriteString("\n")
			}
			
			// 带可选颜色写入严重性
			severityStr := string(finding.Severity)
			if f.ColorEnabled {
				var colorCode string
				switch finding.Severity {
				case SeverityCritical:
					colorCode = "\033[1;31m" // 粗体红色
				case SeverityError:
					colorCode = "\033[31m" // 红色
				case SeverityWarning:
					colorCode = "\033[33m" // 黄色
				case SeverityInfo:
					colorCode = "\033[36m" // 青色
				}
				
				if colorCode != "" {
					severityStr = fmt.Sprintf("%s%s\033[0m", colorCode, severityStr)
				}
			}
			
			sb.WriteString(fmt.Sprintf("[%s] Rule: %s\n", severityStr, finding.RuleID))
			sb.WriteString(fmt.Sprintf("Message: %s\n", finding.Message))
			
			if finding.Recommendation != "" {
				sb.WriteString(fmt.Sprintf("Recommendation: %s\n", finding.Recommendation))
			}
			
			// 添加相关详情
			if cpuUtil, ok := finding.Details["cpu_utilization"]; ok {
				sb.WriteString(fmt.Sprintf("CPU Utilization: %.1f%%\n", cpuUtil))
			}
			
			if memUtil, ok := finding.Details["memory_utilization"]; ok {
				sb.WriteString(fmt.Sprintf("Memory Utilization: %.1f%%\n", memUtil))
			}
		}
		
		resourceCount++
	}
} 