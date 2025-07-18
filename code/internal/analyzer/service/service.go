package service

import (
	"strings"

	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/models"
)

// ServiceAnalyzer Service 安全分析器
type ServiceAnalyzer struct{}

// NewServiceAnalyzer 创建 Service 分析器
func NewServiceAnalyzer() *ServiceAnalyzer {
	return &ServiceAnalyzer{}
}

// CheckServiceTypeSecurity 检查服务类型安全性
func (a *ServiceAnalyzer) CheckServiceTypeSecurity(service *models.Service) string {
	switch service.Type {
	case "LoadBalancer":
		return "high_risk"
	case "NodePort":
		return "medium_risk"
	case "ExternalName":
		return "medium_risk"
	default:
		return "low_risk"
	}
}

// GetMinPort 获取最小端口号
func (a *ServiceAnalyzer) GetMinPort(service *models.Service) int32 {
	if len(service.Ports) == 0 {
		return 0
	}

	minPort := service.Ports[0].Port
	for _, port := range service.Ports {
		if port.Port < minPort {
			minPort = port.Port
		}
	}
	return minPort
}

// HasReadyEndpoints 检查是否有就绪的端点
func (a *ServiceAnalyzer) HasReadyEndpoints(service *models.Service) bool {
	return service.ReadyEndpoints > 0
}

// HasMatchingPods 检查是否有匹配的 Pod
func (a *ServiceAnalyzer) HasMatchingPods(service *models.Service) bool {
	for _, pod := range service.MatchingPods {
		if pod.Ready && pod.Phase == "Running" {
			return true
		}
	}
	return false
}

// HasSensitiveAnnotations 检查是否有敏感注解
func (a *ServiceAnalyzer) HasSensitiveAnnotations(service *models.Service) bool {
	sensitiveKeys := []string{
		"password", "passwd", "pwd",
		"token", "auth", "authorization",
		"secret", "key", "credential",
		"cert", "certificate", "private",
		"api-key", "apikey", "access-key",
	}
	
	for key, value := range service.Annotations {
		keyLower := strings.ToLower(key)
		valueLower := strings.ToLower(value)
		
		for _, sensitive := range sensitiveKeys {
			if strings.Contains(keyLower, sensitive) || strings.Contains(valueLower, sensitive) {
				return true
			}
		}
	}
	return false
}

// IsLoadBalancerType 检查是否为 LoadBalancer 类型
func (a *ServiceAnalyzer) IsLoadBalancerType(service *models.Service) bool {
	return service.Type == "LoadBalancer"
}

// IsNodePortType 检查是否为 NodePort 类型
func (a *ServiceAnalyzer) IsNodePortType(service *models.Service) bool {
	return service.Type == "NodePort"
}

// HasSelector 检查是否有 Selector
func (a *ServiceAnalyzer) HasSelector(service *models.Service) bool {
	return len(service.Selector) > 0
}

// GetServiceType 获取服务类型字符串
func (a *ServiceAnalyzer) GetServiceType(service *models.Service) string {
	return service.Type
}

// GetReadyEndpointCount 获取就绪端点数量
func (a *ServiceAnalyzer) GetReadyEndpointCount(service *models.Service) int {
	return service.ReadyEndpoints
}

// GetMatchingPodCount 获取匹配的 Pod 数量
func (a *ServiceAnalyzer) GetMatchingPodCount(service *models.Service) int {
	return len(service.MatchingPods)
}

// GetReadyPodCount 获取就绪的 Pod 数量
func (a *ServiceAnalyzer) GetReadyPodCount(service *models.Service) int {
	count := 0
	for _, pod := range service.MatchingPods {
		if pod.Ready && pod.Phase == "Running" {
			count++
		}
	}
	return count
}

// HasPortNames 检查端口是否都有名称
func (a *ServiceAnalyzer) HasPortNames(service *models.Service) bool {
	for _, port := range service.Ports {
		if port.Name == "" {
			return false
		}
	}
	return len(service.Ports) > 0
}

// GetPortCount 获取端口数量
func (a *ServiceAnalyzer) GetPortCount(service *models.Service) int {
	return len(service.Ports)
}

// HasExternalTrafficPolicy 检查是否设置了外部流量策略
func (a *ServiceAnalyzer) HasExternalTrafficPolicy(service *models.Service) bool {
	// 这个需要从原始 Service 对象获取，暂时返回 false
	// 在实际实现中可能需要扩展 Service 结构
	return false
}

// GetAnnotationValue 获取指定注解的值
func (a *ServiceAnalyzer) GetAnnotationValue(service *models.Service, key string) string {
	if service.Annotations == nil {
		return ""
	}
	return service.Annotations[key]
}

// HasAnnotation 检查是否有指定注解
func (a *ServiceAnalyzer) HasAnnotation(service *models.Service, key string) bool {
	if service.Annotations == nil {
		return false
	}
	_, exists := service.Annotations[key]
	return exists
}
