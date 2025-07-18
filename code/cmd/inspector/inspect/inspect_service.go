package inspect

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/analyzer/service"
	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/collector"
	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/cluster"
	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/rules"
	"github.com/spf13/cobra"
)

// 共享配置选项
var (
	svcKubeconfig   *string
	svcContextName  *string
	svcRulesFile    *string
	svcNoColor      *bool
)

// 颜色对象
var (
	svcRedColor    = color.New(color.FgRed, color.Bold)
	svcGreenColor  = color.New(color.FgGreen, color.Bold)
	svcYellowColor = color.New(color.FgYellow, color.Bold)
)

// 颜色工具函数
func svcColoredFail(text string) string {
	if svcNoColor != nil && *svcNoColor {
		return text
	}
	return svcRedColor.Sprint(text)
}

func svcColoredSuccess(text string) string {
	if svcNoColor != nil && *svcNoColor {
		return text
	}
	return svcGreenColor.Sprint(text)
}

func svcColoredWarning(text string) string {
	if svcNoColor != nil && *svcNoColor {
		return text
	}
	return svcYellowColor.Sprint(text)
}

func NewServiceCommand(kubecfg, ctx *string, rFile *string, noColor *bool) *cobra.Command {
	svcKubeconfig = kubecfg
	svcContextName = ctx
	svcRulesFile = rFile
	svcNoColor = noColor

	cmd := &cobra.Command{
		Use:   "service",
		Short: "检查Service资源并生成报告",
		Long:  `检查Kubernetes集群中的Service资源配置与合规性，并生成详细报告。`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := runServiceInspect(); err != nil {
				fmt.Fprintf(os.Stderr, "检查Service失败: %v\n", err)
				os.Exit(1)
			}
		},
	}
	return cmd
}

func runServiceInspect() error {
	client, err := cluster.NewClient(*svcKubeconfig, *svcContextName)
	if err != nil {
		return fmt.Errorf("创建集群客户端失败: %w", err)
	}
	collectorInst := collector.NewServiceCollector(client)

	// 加载规则
	var rulesEngine *rules.Engine
	if *svcRulesFile != "" {
		rulesEngine, err = rules.NewEngine(*svcRulesFile)
	} else {
		// 使用默认规则文件
		defaultRulesPath := filepath.Join("code", "configs", "rules", "service.yaml")
		rulesEngine, err = rules.NewEngine(defaultRulesPath)
	}
	if err != nil {
		return fmt.Errorf("加载规则引擎失败: %w", err)
	}

	// 获取所有命名空间的Service
	namespaces := []string{"default", "kube-system", "kube-public", "kube-node-lease"}
	
	// 获取规则列表
	ruleFilter := rules.RuleFilter{
		Categories: []string{"service"},
	}
	rulesList := rulesEngine.GetRules(ruleFilter)

	for _, namespace := range namespaces {
		services, err := collectorInst.GetServices(context.TODO(), namespace)
		if err != nil {
			fmt.Fprintf(os.Stderr, "获取命名空间 %s 的Service失败: %v\n", namespace, err)
			continue
		}

		if len(services) == 0 {
			continue
		}

		// 分析与规则适配
		for _, svc := range services {
			hasIssues := false
			var failedChecks []string
			
			for _, rule := range rulesList {
				var actualValue interface{}
				var metricType string
				
				analyzer := service.NewServiceAnalyzer()
				
				switch rule.Condition.Metric {
				case "is_loadbalancer_type":
					actualValue = analyzer.IsLoadBalancerType(svc)
					metricType = "boolean"
				case "is_nodeport_type":
					actualValue = analyzer.IsNodePortType(svc)
					metricType = "boolean"
				case "min_port":
					actualValue = analyzer.GetMinPort(svc)
					metricType = "numeric"
				case "has_sensitive_annotations":
					actualValue = analyzer.HasSensitiveAnnotations(svc)
					metricType = "boolean"
				case "has_ready_endpoints":
					actualValue = analyzer.HasReadyEndpoints(svc)
					metricType = "boolean"
				case "has_matching_pods":
					actualValue = analyzer.HasMatchingPods(svc)
					metricType = "boolean"
				case "has_labels":
					actualValue = svc.Labels
					metricType = "map"
				case "has_selector":
					actualValue = analyzer.HasSelector(svc)
					metricType = "boolean"
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
					// 处理消息重复问题
					message := result.Message
					if len(message) > len(rule.Name)+2 && message[:len(rule.Name)+2] == rule.Name+": " {
						message = message[len(rule.Name)+2:]
					}
					failedChecks = append(failedChecks, fmt.Sprintf("  %s %s: %s", svcColoredFail("[FAIL]"), rule.Name, message))
				}
			}
			
			// 输出结果
			if hasIssues {
				fmt.Printf("\nService %s/%s 检查问题:\n", svc.Namespace, svc.Name)
				for _, check := range failedChecks {
					fmt.Println(check)
				}
			} else {
				fmt.Printf("Service %s/%s: %s\n", svc.Namespace, svc.Name, svcColoredSuccess("所有检查通过"))
			}
		}
	}

	return nil
}
