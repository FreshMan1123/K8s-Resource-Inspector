package test

import (
	"context"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/cluster"
	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/collector"
	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/analyzer/deployment"
)

func int32Ptr(i int32) *int32 { return &i }

func TestDeploymentCollectorWithFakeClient(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset(
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-dep",
				Namespace: "default",
				Labels:    map[string]string{"app": "web"},
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: int32Ptr(2),
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "c1",
								Image: "nginx:latest",
								ImagePullPolicy: corev1.PullIfNotPresent,
								Resources: corev1.ResourceRequirements{
									Limits:   corev1.ResourceList{"cpu": resourceQuantity("500m")},
									Requests: corev1.ResourceList{"cpu": resourceQuantity("100m")},
								},
							},
						},
					},
				},
			},
		},
	)

	cli := &cluster.Client{Clientset: fakeClientset}
	depCollector := collector.NewDeploymentCollector(cli)
	deployments, err := depCollector.GetDeployments(context.TODO(), "default")
	if err != nil {
		t.Fatalf("采集失败: %v", err)
	}
	if len(deployments) != 1 {
		t.Fatalf("期望采集到1个Deployment，实际: %d", len(deployments))
	}
	dep := deployments[0]
	if !deployment.CheckMinReplicas(dep, 2) {
		t.Errorf("副本数检查失败，应通过")
	}
	if !deployment.AllContainersHaveResourceLimits(dep) {
		t.Errorf("资源限制检查失败，应通过")
	}
	if !deployment.AllContainersImagePullPolicy(dep, "IfNotPresent") {
		t.Errorf("镜像拉取策略检查失败，应通过")
	}
	if !deployment.HasLabels(dep, map[string]string{"app": "web"}) {
		t.Errorf("标签检查失败，应通过")
	}
}

func resourceQuantity(val string) resource.Quantity {
	return resource.MustParse(val)
} 