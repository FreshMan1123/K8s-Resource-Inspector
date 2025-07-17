package collector

import (
	"context"
	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/cluster"
	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/models"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

type DeploymentCollector struct {
	client *cluster.Client
}

func NewDeploymentCollector(client *cluster.Client) *DeploymentCollector {
	return &DeploymentCollector{client: client}
}

func (dc *DeploymentCollector) GetDeployments(ctx context.Context, namespace string) ([]models.Deployment, error) {
	deployments, err := dc.client.ListRawDeployments(ctx, namespace)
	if err != nil {
		return nil, err
	}
	result := make([]models.Deployment, 0, len(deployments))
	for _, d := range deployments {
		result = append(result, convertDeploymentToModel(&d))
	}
	return result, nil
}

func convertDeploymentToModel(d *appsv1.Deployment) models.Deployment {
	containers := make([]models.DeploymentContainer, 0, len(d.Spec.Template.Spec.Containers))
	for _, c := range d.Spec.Template.Spec.Containers {
		containers = append(containers, models.DeploymentContainer{
			Name:  c.Name,
			Image: c.Image,
			ImagePullPolicy: string(c.ImagePullPolicy),
			Resources: models.ResourceSpec{
				Limits:   resourceListToMap(c.Resources.Limits),
				Requests: resourceListToMap(c.Resources.Requests),
			},
		})
	}
	return models.Deployment{
		Name:        d.Name,
		Namespace:   d.Namespace,
		Labels:      d.Labels,
		Annotations: d.Annotations,
		Replicas:    getInt32(d.Spec.Replicas),
		AvailableReplicas: d.Status.AvailableReplicas,
		Strategy:    string(d.Spec.Strategy.Type),
		Containers:  containers,
	}
}

func getInt32(p *int32) int32 {
	if p == nil {
		return 1
	}
	return *p
}

func resourceListToMap(rl corev1.ResourceList) map[string]string {
	m := make(map[string]string)
	for k, v := range rl {
		m[string(k)] = v.String()
	}
	return m
} 