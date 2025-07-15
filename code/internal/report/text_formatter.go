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
		// 基本信息
		sb.WriteString(fmt.Sprintf("节点: %s\n", node.Name))
		
		// 角色信息
		if len(node.Roles) > 0 {
			sb.WriteString(fmt.Sprintf("角色: %s\n", strings.Join(node.Roles, ", ")))
		} else {
			sb.WriteString("角色: 未知\n")
		}
		
		// 状态信息
		sb.WriteString(fmt.Sprintf("状态: %s\n", getNodeStatusString(node.Ready)))
		sb.WriteString(fmt.Sprintf("可调度: %v\n", node.Schedulable))
		
		// 地址信息
		if len(node.Addresses) > 0 {
			sb.WriteString("地址:\n")
			for addrType, addr := range node.Addresses {
				sb.WriteString(fmt.Sprintf("  %s: %s\n", addrType, addr))
			}
		}
		
		// 创建时间
		if !node.CreationTime.IsZero() {
			sb.WriteString(fmt.Sprintf("创建时间: %s\n", node.CreationTime.Format(time.RFC3339)))
		}
		
		// 节点信息
		sb.WriteString("节点信息:\n")
		sb.WriteString(fmt.Sprintf("  内核版本: %s\n", getValueOrDefault(node.NodeInfo.KernelVersion, "未知")))
		sb.WriteString(fmt.Sprintf("  操作系统: %s\n", getValueOrDefault(node.NodeInfo.OSImage, "未知")))
		sb.WriteString(fmt.Sprintf("  容器运行时: %s\n", getValueOrDefault(node.NodeInfo.ContainerRuntimeVersion, "未知")))
		sb.WriteString(fmt.Sprintf("  Kubelet版本: %s\n", getValueOrDefault(node.NodeInfo.KubeletVersion, "未知")))
		sb.WriteString(fmt.Sprintf("  Kube-Proxy版本: %s\n", getValueOrDefault(node.NodeInfo.KubeProxyVersion, "未知")))
		sb.WriteString(fmt.Sprintf("  架构: %s\n", getValueOrDefault(node.NodeInfo.Architecture, "未知")))
		
		// 资源信息
		sb.WriteString("资源使用情况:\n")
		
		// CPU资源
		sb.WriteString("  CPU:\n")
		if node.CPU.Capacity != "" {
			sb.WriteString(fmt.Sprintf("    总量: %s\n", node.CPU.Capacity))
		}
		if node.CPU.Allocatable != "" {
			sb.WriteString(fmt.Sprintf("    可分配: %s\n", node.CPU.Allocatable))
		}
		if node.CPU.Allocated != "" {
			sb.WriteString(fmt.Sprintf("    已分配: %s\n", node.CPU.Allocated))
		}
		if node.CPU.Used != "" {
			sb.WriteString(fmt.Sprintf("    已使用: %s\n", node.CPU.Used))
		}
		sb.WriteString(fmt.Sprintf("    利用率: %.2f%%\n", node.CPU.Utilization))
		sb.WriteString(fmt.Sprintf("    分配率: %.2f%%\n", node.CPU.AllocationRate))
		
		// 内存资源
		sb.WriteString("  内存:\n")
		if node.Memory.Capacity != "" {
			sb.WriteString(fmt.Sprintf("    总量: %s\n", node.Memory.Capacity))
		}
		if node.Memory.Allocatable != "" {
			sb.WriteString(fmt.Sprintf("    可分配: %s\n", node.Memory.Allocatable))
		}
		if node.Memory.Allocated != "" {
			sb.WriteString(fmt.Sprintf("    已分配: %s\n", node.Memory.Allocated))
		}
		if node.Memory.Used != "" {
			sb.WriteString(fmt.Sprintf("    已使用: %s\n", node.Memory.Used))
		}
		sb.WriteString(fmt.Sprintf("    利用率: %.2f%%\n", node.Memory.Utilization))
		sb.WriteString(fmt.Sprintf("    分配率: %.2f%%\n", node.Memory.AllocationRate))
		
		// 临时存储资源
		sb.WriteString("  临时存储:\n")
		if node.EphemeralStorage.Capacity != "" {
			sb.WriteString(fmt.Sprintf("    总量: %s\n", node.EphemeralStorage.Capacity))
		}
		if node.EphemeralStorage.Allocatable != "" {
			sb.WriteString(fmt.Sprintf("    可分配: %s\n", node.EphemeralStorage.Allocatable))
		}
		if node.EphemeralStorage.Allocated != "" {
			sb.WriteString(fmt.Sprintf("    已分配: %s\n", node.EphemeralStorage.Allocated))
		}
		if node.EphemeralStorage.Used != "" {
			sb.WriteString(fmt.Sprintf("    已使用: %s\n", node.EphemeralStorage.Used))
		}
		sb.WriteString(fmt.Sprintf("    利用率: %.2f%%\n", node.EphemeralStorage.Utilization))
		sb.WriteString(fmt.Sprintf("    分配率: %.2f%%\n", node.EphemeralStorage.AllocationRate))
		
		// Pod信息
		sb.WriteString(fmt.Sprintf("Pod数量: %d/%d (%.2f%%)\n", node.RunningPods, node.TotalPods, 
			calculatePodPercentage(node.RunningPods, node.TotalPods)))
		
		// 压力状态
		sb.WriteString("压力状态:\n")
		sb.WriteString(fmt.Sprintf("  内存压力: %v\n", node.PressureStatus.MemoryPressure))
		sb.WriteString(fmt.Sprintf("  CPU压力: %v\n", node.PressureStatus.CPUPressure))
		sb.WriteString(fmt.Sprintf("  磁盘压力: %v\n", node.PressureStatus.DiskPressure))
		sb.WriteString(fmt.Sprintf("  网络压力: %v\n", node.PressureStatus.NetworkPressure))
		sb.WriteString(fmt.Sprintf("  PID压力: %v\n", node.PressureStatus.PIDPressure))
		
		// 健康评分
		sb.WriteString(fmt.Sprintf("健康评分: %d/100\n", node.HealthScore))
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

// getValueOrDefault 获取值或默认值
func getValueOrDefault(value string, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

// calculatePodPercentage 计算Pod百分比
func calculatePodPercentage(running, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(running) / float64(total) * 100
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