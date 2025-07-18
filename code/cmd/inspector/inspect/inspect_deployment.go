package inspect

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/analyzer/deployment"
	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/collector"
	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/cluster"
	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/rules"
	"github.com/spf13/cobra"
)

// 共享配置选项
var (
	depKubeconfig   *string
	depContextName  *string
	depRulesFile    *string
	depNoColor      *bool
)

// 颜色对象
var (
	redColor    = color.New(color.FgRed, color.Bold)
	greenColor  = color.New(color.FgGreen, color.Bold)
	yellowColor = color.New(color.FgYellow, color.Bold)
)

// 颜色工具函数
func coloredFail(text string) string {
	if depNoColor != nil && *depNoColor {
		return text
	}
	return redColor.Sprint(text)
}

func coloredSuccess(text string) string {
	if depNoColor != nil && *depNoColor {
		return text
	}
	return greenColor.Sprint(text)
}

func coloredWarning(text string) string {
	if depNoColor != nil && *depNoColor {
		return text
	}
	return yellowColor.Sprint(text)
}

func NewDeploymentCommand(kubecfg, ctx *string, rFile *string, noColor *bool) *cobra.Command {
	depKubeconfig = kubecfg
	depContextName = ctx
	depRulesFile = rFile
	depNoColor = noColor

	cmd := &cobra.Command{
		Use:   "deployment",
		Short: "检查Deployment资源并生成报告",
		Long:  `检查Kubernetes集群中的Deployment资源配置与合规性，并生成详细报告。`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := runDeploymentInspect(); err != nil {
				fmt.Fprintf(os.Stderr, "检查Deployment失败: %v\n", err)
				os.Exit(1)
			}
		},
	}
	return cmd
}

func runDeploymentInspect() error {
	client, err := cluster.NewClient(*depKubeconfig, *depContextName)
	if err != nil {
		return fmt.Errorf("创建集群客户端失败: %w", err)
	}
	collectorInst := collector.NewDeploymentCollector(client)
	// analyzer := deployment.NewDeploymentAnalyzer(collectorInst) // 已声明未用，删除

	// 加载规则
	var rulesEngine *rules.Engine
	if *depRulesFile != "" {
		rulesEngine, err = rules.NewEngine(*depRulesFile)
	} else {
		defaultRulesPath := filepath.Join("code", "configs", "rules", "deployment.yaml")
		rulesEngine, err = rules.NewEngine(defaultRulesPath)
	}
	if err != nil {
		return fmt.Errorf("加载规则引擎失败: %w", err)
	}
	filter := rules.RuleFilter{}
	rulesList := rulesEngine.GetRules(filter)

	// 采集所有Deployment
	deployments, err := collectorInst.GetDeployments(cmdContext(), "")
	if err != nil {
		return fmt.Errorf("采集Deployment失败: %w", err)
	}

	// 分析与规则适配
	for _, dep := range deployments {
		hasIssues := false
		var failedChecks []string

		for _, rule := range rulesList {
			var actualValue interface{}
			var metricType string
			switch rule.Condition.Metric {
			case "replicas":
				actualValue = dep.Replicas
				metricType = "numeric"
			case "has_resource_limits":
				actualValue = deployment.AllContainersHaveResourceLimits(dep)
				metricType = "boolean"
			case "image_pull_policy":
				actualValue = deployment.GetImagePullPolicy(dep)
				metricType = "string"
			case "has_labels":
				actualValue = dep.Labels
				metricType = "map"
			default:
				continue
			}

			result, err := rulesEngine.EvaluateRule(rule, metricType, actualValue)
			if err != nil {
				fmt.Fprintf(os.Stderr, "规则评估失败: %v\n", err)
				continue
			}

			// 只记录失败的检查
			if !result.Passed {
				hasIssues = true
				// 由于result.Message已经包含了rule.Name，我们直接使用Message
				// 或者提取Message中除了rule.Name之外的部分
				message := result.Message
				if len(message) > len(rule.Name)+2 && message[:len(rule.Name)+2] == rule.Name+": " {
					// 如果消息以"规则名: "开头，则去掉这部分
					message = message[len(rule.Name)+2:]
				}
				failedChecks = append(failedChecks, fmt.Sprintf("  %s %s: %s", coloredFail("[FAIL]"), rule.Name, message))
			}
		}

		// 输出结果
		if hasIssues {
			fmt.Printf("\nDeployment %s/%s 检查问题:\n", dep.Namespace, dep.Name)
			for _, check := range failedChecks {
				fmt.Println(check)
			}
		} else {
			fmt.Printf("Deployment %s/%s: %s\n", dep.Namespace, dep.Name, coloredSuccess("所有检查通过"))
		}
	}
	return nil
}

func cmdContext() context.Context {
	return context.TODO()
}
