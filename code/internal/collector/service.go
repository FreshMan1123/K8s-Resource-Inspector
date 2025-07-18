package collector

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"


	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/cluster"
	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/models"
)

// ServiceInfo Service 完整信息
type ServiceInfo struct {
	Name        string
	Namespace   string
	Type        v1.ServiceType
	Ports       []v1.ServicePort
	Selector    map[string]string
	Labels      map[string]string
	Annotations map[string]string

	// 连通性相关
	Endpoints      []EndpointInfo
	MatchingPods   []PodInfo
	ReadyEndpoints int
}

// EndpointInfo 端点信息
type EndpointInfo struct {
	IP    string
	Port  int32
	Ready bool
}

// PodInfo 关联的 Pod 信息
type PodInfo struct {
	Name      string
	Namespace string
	Ready     bool
	Phase     v1.PodPhase
}

// ServiceCollector Service 数据收集器
type ServiceCollector struct {
	client *cluster.Client
}

// NewServiceCollector 创建 Service 收集器
func NewServiceCollector(client *cluster.Client) *ServiceCollector {
	return &ServiceCollector{
		client: client,
	}
}

// GetServices 获取指定命名空间的所有 Service 信息
func (c *ServiceCollector) GetServices(ctx context.Context, namespace string) ([]models.Service, error) {
	// 通过 cluster 层获取 Service 列表
	services, err := c.client.ListRawServices(ctx, namespace)
	if err != nil {
		return nil, fmt.Errorf("获取 Service 列表失败: %w", err)
	}

	var serviceInfos []models.Service
	for _, service := range services {
		serviceInfo, err := c.buildServiceInfo(ctx, &service)
		if err != nil {
			// 记录错误但继续处理其他 Service
			fmt.Printf("处理 Service %s/%s 失败: %v\n", service.Namespace, service.Name, err)
			continue
		}
		serviceInfos = append(serviceInfos, serviceInfo)
	}

	return serviceInfos, nil
}

// buildServiceInfo 构建完整的 Service 模型
func (c *ServiceCollector) buildServiceInfo(ctx context.Context, service *v1.Service) (models.Service, error) {
	// 使用 models 包中的转换函数
	serviceInfo := models.FromK8sService(service)

	// 获取 Endpoints 信息
	endpoints, err := c.getEndpointsForService(ctx, service)
	if err != nil {
		return models.Service{}, fmt.Errorf("获取 Endpoints 失败: %w", err)
	}
	serviceInfo.Endpoints = endpoints

	// 计算就绪的端点数量
	readyCount := 0
	for _, ep := range endpoints {
		if ep.Ready {
			readyCount++
		}
	}
	serviceInfo.ReadyEndpoints = readyCount

	// 获取匹配的 Pod 信息
	if len(service.Spec.Selector) > 0 {
		pods, err := c.getMatchingPods(ctx, service)
		if err != nil {
			return models.Service{}, fmt.Errorf("获取匹配的 Pod 失败: %w", err)
		}
		serviceInfo.MatchingPods = pods
	}

	return serviceInfo, nil
}

// getEndpointsForService 获取 Service 对应的 Endpoints
func (c *ServiceCollector) getEndpointsForService(ctx context.Context, service *v1.Service) ([]models.Endpoint, error) {
	endpoints, err := c.client.GetEndpoints(ctx, service.Namespace, service.Name)
	if err != nil {
		// Endpoints 不存在是正常情况（比如没有匹配的 Pod）
		return []models.Endpoint{}, nil
	}

	var endpointInfos []models.Endpoint
	for _, subset := range endpoints.Subsets {
		// 处理就绪的端点
		for _, addr := range subset.Addresses {
			for _, port := range subset.Ports {
				endpointInfos = append(endpointInfos, models.Endpoint{
					IP:    addr.IP,
					Port:  port.Port,
					Ready: true,
				})
			}
		}

		// 处理未就绪的端点
		for _, addr := range subset.NotReadyAddresses {
			for _, port := range subset.Ports {
				endpointInfos = append(endpointInfos, models.Endpoint{
					IP:    addr.IP,
					Port:  port.Port,
					Ready: false,
				})
			}
		}
	}

	return endpointInfos, nil
}

// getMatchingPods 根据 selector 获取匹配的 Pod
func (c *ServiceCollector) getMatchingPods(ctx context.Context, service *v1.Service) ([]models.ServicePod, error) {
	// 通过 cluster 层获取匹配的 Pod
	pods, err := c.client.GetPodsBySelector(ctx, service.Namespace, service.Spec.Selector)
	if err != nil {
		return nil, fmt.Errorf("查询匹配的 Pod 失败: %w", err)
	}

	var podInfos []models.ServicePod
	for _, pod := range pods.Items {
		// 检查 Pod 是否就绪
		ready := false
		for _, condition := range pod.Status.Conditions {
			if condition.Type == v1.PodReady && condition.Status == v1.ConditionTrue {
				ready = true
				break
			}
		}

		podInfos = append(podInfos, models.ServicePod{
			Name:      pod.Name,
			Namespace: pod.Namespace,
			Ready:     ready,
			Phase:     string(pod.Status.Phase),
		})
	}

	return podInfos, nil
}
