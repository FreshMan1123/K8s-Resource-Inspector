package resource

import (
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
)

// GetPodStatus 获取Pod的状态
func GetPodStatus(pod *corev1.Pod) string {
	if pod.Status.Phase == corev1.PodSucceeded {
		return "Completed"
	}
	if pod.Status.Phase == corev1.PodFailed {
		return "Error"
	}

	if pod.DeletionTimestamp != nil {
		return "Terminating"
	}
	
	// 检查容器状态
	for _, container := range pod.Status.ContainerStatuses {
		if container.State.Waiting != nil {
			return container.State.Waiting.Reason
		}
		if container.State.Terminated != nil {
			return container.State.Terminated.Reason
		}
	}
	
	return string(pod.Status.Phase)
}

// FormatAge 格式化时间间隔为人类可读的形式
func FormatAge(d time.Duration) string {
	d = d.Round(time.Second)
	
	if d.Hours() > 24*365 {
		years := int(d.Hours() / (24 * 365))
		return fmt.Sprintf("%dy", years)
	}
	if d.Hours() > 24*30 {
		months := int(d.Hours() / (24 * 30))
		return fmt.Sprintf("%dm", months)
	}
	if d.Hours() > 24 {
		days := int(d.Hours() / 24)
		return fmt.Sprintf("%dd", days)
	}
	if d.Hours() > 0 {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	if d.Minutes() > 0 {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	
	return fmt.Sprintf("%ds", int(d.Seconds()))
}

// DetermineNamespace 确定要使用的命名空间
func DetermineNamespace(allNamespaces bool, namespace string) string {
	if allNamespaces {
		return ""
	}
	
	if namespace == "" {
		return "default"
	}
	
	return namespace
} 