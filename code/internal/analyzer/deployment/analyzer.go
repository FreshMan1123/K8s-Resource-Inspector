package deployment

import (

	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/collector"
	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/models"
)

type DeploymentAnalyzer struct {
	collector *collector.DeploymentCollector
}

func NewDeploymentAnalyzer(collector *collector.DeploymentCollector) *DeploymentAnalyzer {
	return &DeploymentAnalyzer{collector: collector}
}

// AnalysisResult 表示单个Deployment的分析结果
// 可根据后续报告结构扩展

// HasLabels 检查Deployment是否包含所有指定标签
func HasLabels(deployment models.Deployment, required map[string]string) bool {
	for k, v := range required {
		if val, ok := deployment.Labels[k]; !ok || val != v {
			return false
		}
	}
	return true
}

// 检查副本数是否大于等于min
func CheckMinReplicas(deployment models.Deployment, min int32) bool {
	return deployment.Replicas >= min
}

// 检查所有容器是否都设置了资源限制
func AllContainersHaveResourceLimits(deployment models.Deployment) bool {
	for _, c := range deployment.Containers {
		if len(c.Resources.Limits) == 0 || len(c.Resources.Requests) == 0 {
			return false
		}
	}
	return true
}

// 检查所有容器的ImagePullPolicy是否等于指定值
func AllContainersImagePullPolicy(deployment models.Deployment, policy string) bool {
	for _, c := range deployment.Containers {
		if c.ImagePullPolicy != policy {
			return false
		}
	}
	return true
}

// 获取所有容器的统一ImagePullPolicy（如不一致则返回Mixed）
func GetImagePullPolicy(deployment models.Deployment) string {
	if len(deployment.Containers) == 0 {
		return ""
	}
	policy := deployment.Containers[0].ImagePullPolicy
	for _, c := range deployment.Containers[1:] {
		if c.ImagePullPolicy != policy {
			return "Mixed"
		}
	}
	return policy
} 