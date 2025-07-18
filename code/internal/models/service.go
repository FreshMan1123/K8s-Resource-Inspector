package models

import v1 "k8s.io/api/core/v1"

// Service 表示 Kubernetes Service 的简化模型
type Service struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	Type        string            `json:"type"`
	Ports       []ServicePort     `json:"ports"`
	Selector    map[string]string `json:"selector"`
	
	// 连通性相关信息
	Endpoints      []Endpoint `json:"endpoints"`
	MatchingPods   []ServicePod `json:"matchingPods"`
	ReadyEndpoints int        `json:"readyEndpoints"`
}

// ServicePort 表示 Service 端口配置
type ServicePort struct {
	Name       string `json:"name"`
	Protocol   string `json:"protocol"`
	Port       int32  `json:"port"`
	TargetPort string `json:"targetPort"`
	NodePort   int32  `json:"nodePort,omitempty"`
}

// Endpoint 表示服务端点信息
type Endpoint struct {
	IP    string `json:"ip"`
	Port  int32  `json:"port"`
	Ready bool   `json:"ready"`
}

// ServicePod 表示 Service 关联的 Pod 信息（简化版）
type ServicePod struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Ready     bool   `json:"ready"`
	Phase     string `json:"phase"`
}

// ServiceSummary 表示 Service 检查结果摘要
type ServiceSummary struct {
	TotalServices    int `json:"totalServices"`
	HealthyServices  int `json:"healthyServices"`
	UnhealthyServices int `json:"unhealthyServices"`
	SecurityRisks    int `json:"securityRisks"`
	ConnectivityIssues int `json:"connectivityIssues"`
}

// ServiceCheckResult 表示单个 Service 的检查结果
type ServiceCheckResult struct {
	Service     Service           `json:"service"`
	ChecksPassed int              `json:"checksPassed"`
	ChecksFailed int              `json:"checksFailed"`
	Issues      []ServiceIssue    `json:"issues"`
	Status      string            `json:"status"` // "healthy", "warning", "error"
}

// ServiceIssue 表示 Service 检查中发现的问题
type ServiceIssue struct {
	RuleID      string `json:"ruleId"`
	RuleName    string `json:"ruleName"`
	Severity    string `json:"severity"`
	Message     string `json:"message"`
	Remediation string `json:"remediation"`
	ActualValue interface{} `json:"actualValue"`
	ExpectedValue interface{} `json:"expectedValue"`
}

// ServiceReport 表示完整的 Service 巡检报告
type ServiceReport struct {
	ClusterName string               `json:"clusterName"`
	Timestamp   string               `json:"timestamp"`
	Summary     ServiceSummary       `json:"summary"`
	Results     []ServiceCheckResult `json:"results"`
}

// 转换函数：从 collector.ServiceInfo 转换为 models.Service
func FromCollectorServiceInfo(info interface{}) Service {
	// 这里需要根据实际的 collector.ServiceInfo 结构进行转换
	// 暂时返回空的 Service，实际使用时需要实现具体的转换逻辑
	return Service{}
}

// 转换函数：从 Kubernetes API 对象转换为 models.Service
func FromK8sService(k8sService *v1.Service) Service {
	service := Service{
		Name:        k8sService.Name,
		Namespace:   k8sService.Namespace,
		Labels:      k8sService.Labels,
		Annotations: k8sService.Annotations,
		Type:        string(k8sService.Spec.Type),
		Selector:    k8sService.Spec.Selector,
	}

	// 转换端口信息
	for _, port := range k8sService.Spec.Ports {
		servicePort := ServicePort{
			Name:       port.Name,
			Protocol:   string(port.Protocol),
			Port:       port.Port,
			TargetPort: port.TargetPort.String(),
		}
		if port.NodePort != 0 {
			servicePort.NodePort = port.NodePort
		}
		service.Ports = append(service.Ports, servicePort)
	}

	return service
}

// 辅助方法：检查 Service 是否健康
func (s *Service) IsHealthy() bool {
	// 基本健康检查逻辑
	if len(s.Selector) == 0 && s.Type != "ExternalName" {
		return false
	}
	if s.ReadyEndpoints == 0 {
		return false
	}
	return true
}

// 辅助方法：获取 Service 的风险等级
func (s *Service) GetRiskLevel() string {
	switch s.Type {
	case "LoadBalancer":
		return "high"
	case "NodePort":
		return "medium"
	case "ExternalName":
		return "medium"
	default:
		return "low"
	}
}

// 辅助方法：检查是否使用了特权端口
func (s *Service) HasPrivilegedPorts() bool {
	for _, port := range s.Ports {
		if port.Port < 1024 {
			return true
		}
	}
	return false
}

// 辅助方法：获取最小端口号
func (s *Service) GetMinPort() int32 {
	if len(s.Ports) == 0 {
		return 0
	}
	
	minPort := s.Ports[0].Port
	for _, port := range s.Ports {
		if port.Port < minPort {
			minPort = port.Port
		}
	}
	return minPort
}
