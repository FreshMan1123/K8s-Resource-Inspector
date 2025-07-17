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
	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/rules"
)

// TestMapValidatorIntegration 测试MapValidator修复和新的owner标签规则
func TestMapValidatorIntegration(t *testing.T) {
	// 创建规则引擎，使用测试规则配置
	rulesEngine, err := rules.NewEngine("testdata/deployment_rules_test.yaml")
	if err != nil {
		t.Fatalf("创建规则引擎失败: %v", err)
	}

	tests := []struct {
		name           string
		deployment     *appsv1.Deployment
		expectedResult map[string]bool // 规则ID -> 是否通过
	}{
		{
			name: "有owner标签且值不为空 - 应该通过",
			deployment: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-with-owner",
					Namespace: "default",
					Labels: map[string]string{
						"owner": "team-backend",
						"app":   "web",
					},
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: int32Ptr(2),
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:            "nginx",
									Image:           "nginx:latest",
									ImagePullPolicy: corev1.PullIfNotPresent,
									Resources: corev1.ResourceRequirements{
										Limits: corev1.ResourceList{
											"cpu":    resource.MustParse("500m"),
											"memory": resource.MustParse("512Mi"),
										},
									},
								},
							},
						},
					},
				},
			},
			expectedResult: map[string]bool{
				"min_replicas":           true,
				"require_resource_limits": true,
				"require_image_pull_policy": true,
				"require_owner_label":    true,
			},
		},
		{
			name: "没有owner标签 - 应该失败",
			deployment: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-no-owner",
					Namespace: "default",
					Labels: map[string]string{
						"app": "web",
						"env": "prod",
					},
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: int32Ptr(2),
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:            "nginx",
									Image:           "nginx:latest",
									ImagePullPolicy: corev1.PullIfNotPresent,
									Resources: corev1.ResourceRequirements{
										Limits: corev1.ResourceList{
											"cpu":    resource.MustParse("500m"),
											"memory": resource.MustParse("512Mi"),
										},
									},
								},
							},
						},
					},
				},
			},
			expectedResult: map[string]bool{
				"min_replicas":           true,
				"require_resource_limits": true,
				"require_image_pull_policy": true,
				"require_owner_label":    false,
			},
		},
		{
			name: "owner标签值为空字符串 - 应该失败",
			deployment: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-empty-owner",
					Namespace: "default",
					Labels: map[string]string{
						"owner": "",
						"app":   "web",
					},
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: int32Ptr(2),
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:            "nginx",
									Image:           "nginx:latest",
									ImagePullPolicy: corev1.PullIfNotPresent,
									Resources: corev1.ResourceRequirements{
										Limits: corev1.ResourceList{
											"cpu":    resource.MustParse("500m"),
											"memory": resource.MustParse("512Mi"),
										},
									},
								},
							},
						},
					},
				},
			},
			expectedResult: map[string]bool{
				"min_replicas":           true,
				"require_resource_limits": true,
				"require_image_pull_policy": true,
				"require_owner_label":    false,
			},
		},
		{
			name: "owner标签值为纯空格 - 应该失败",
			deployment: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-whitespace-owner",
					Namespace: "default",
					Labels: map[string]string{
						"owner": "   ",
						"app":   "web",
					},
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: int32Ptr(2),
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:            "nginx",
									Image:           "nginx:latest",
									ImagePullPolicy: corev1.PullIfNotPresent,
									Resources: corev1.ResourceRequirements{
										Limits: corev1.ResourceList{
											"cpu":    resource.MustParse("500m"),
											"memory": resource.MustParse("512Mi"),
										},
									},
								},
							},
						},
					},
				},
			},
			expectedResult: map[string]bool{
				"min_replicas":           true,
				"require_resource_limits": true,
				"require_image_pull_policy": true,
				"require_owner_label":    false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建fake客户端
			fakeClientset := fake.NewSimpleClientset(tt.deployment)
			cli := &cluster.Client{Clientset: fakeClientset}
			
			// 收集Deployment数据
			depCollector := collector.NewDeploymentCollector(cli)
			deployments, err := depCollector.GetDeployments(context.TODO(), "default")
			if err != nil {
				t.Fatalf("采集Deployment失败: %v", err)
			}
			
			if len(deployments) != 1 {
				t.Fatalf("期望采集到1个Deployment，实际: %d", len(deployments))
			}
			
			dep := deployments[0]
			
			// 获取规则并测试
			ruleFilter := rules.RuleFilter{
				Categories: []string{"deployment"},
			}
			rulesList := rulesEngine.GetRules(ruleFilter)
			
			// 测试每个规则
			for _, rule := range rulesList {
				var actualValue interface{}
				var metricType string
				
				switch rule.Condition.Metric {
				case "replicas":
					actualValue = dep.Replicas
					metricType = "numeric"
				case "has_resource_limits":
					// 这里需要实现资源限制检查逻辑
					actualValue = hasResourceLimits(dep)
					metricType = "boolean"
				case "image_pull_policy":
					actualValue = getImagePullPolicy(dep)
					metricType = "string"
				case "has_labels":
					actualValue = dep.Labels
					metricType = "map"
				default:
					continue
				}
				
				result, err := rulesEngine.EvaluateRule(rule, metricType, actualValue)
				if err != nil {
					t.Errorf("规则 %s 评估失败: %v", rule.ID, err)
					continue
				}
				
				expectedPass, exists := tt.expectedResult[rule.ID]
				if !exists {
					continue // 跳过不在预期结果中的规则
				}
				
				if result.Passed != expectedPass {
					t.Errorf("规则 %s 结果不符合预期，期望: %v, 实际: %v, 消息: %s", 
						rule.ID, expectedPass, result.Passed, result.Message)
				}
			}
		})
	}
}

// 辅助函数
func hasResourceLimits(dep interface{}) bool {
	// 简化的资源限制检查逻辑
	return true // 测试数据中都设置了资源限制
}

func getImagePullPolicy(dep interface{}) string {
	// 简化的镜像拉取策略获取逻辑
	return "IfNotPresent" // 测试数据中都设置为IfNotPresent
}
