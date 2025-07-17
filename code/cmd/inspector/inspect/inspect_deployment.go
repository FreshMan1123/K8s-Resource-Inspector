package inspect

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/analyzer/deployment"
	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/collector"
	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/cluster"
	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/rules"
	"github.com/spf13/cobra"
)

// 共享配置选项（与node一致）
var (
	depKubeconfig   *string
	depContextName  *string
	depOutputFormat *string
	depNoColor      *bool
	depRulesFile    *string
	depOutputFile   *string
	depOnlyIssues   *bool
)

func NewDeploymentCommand(kubecfg, ctx, outFmt *string, noClr, onlyIss *bool, rFile, outFile *string) *cobra.Command {
	depKubeconfig = kubecfg
	depContextName = ctx
	depOutputFormat = outFmt
	depNoColor = noClr
	depRulesFile = rFile
	depOutputFile = outFile
	depOnlyIssues = onlyIss

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
			// 这里可收集result，后续生成报告
			fmt.Printf("Deployment %s/%s 规则[%s] 检查结果: %v\n", dep.Namespace, dep.Name, rule.ID, result.Passed)
		}
	}
	return nil
}

func cmdContext() context.Context {
	return context.TODO()
}
