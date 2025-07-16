package inspect

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/analyzer/node"
	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/collector"
	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/cluster"
	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/report"
	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/rules"
	"github.com/spf13/cobra"
)

// 共享的配置选项
var (
	kubeconfig   *string
	contextName  *string
	outputFormat *string
	noColor      *bool
	rulesFile    *string
	outputFile   *string
	onlyIssues   *bool
)

// NewNodeCommand 创建节点检查命令
func NewNodeCommand(kubecfg, ctx, outFmt *string, noClr, onlyIss *bool, rFile, outFile *string) *cobra.Command {
	// 保存引用，供命令执行时使用
	kubeconfig = kubecfg
	contextName = ctx
	outputFormat = outFmt
	noColor = noClr
	rulesFile = rFile
	outputFile = outFile
	onlyIssues = onlyIss

	cmd := &cobra.Command{
		Use:   "node [节点名称]",
		Short: "检查节点资源并生成报告",
		Long:  `检查Kubernetes集群中的节点资源状态并生成详细报告，包括CPU、内存使用情况和潜在问题。`,
		Run: func(cmd *cobra.Command, args []string) {
			// 检查是否指定了节点名称
			nodeName := ""
			if len(args) > 0 {
				nodeName = args[0]
			}

			if err := runNodeInspect(nodeName); err != nil {
				fmt.Fprintf(os.Stderr, "检查节点失败: %v\n", err)
				os.Exit(1)
			}
		},
	}

	return cmd
}

// runNodeInspect 执行节点检查逻辑
func runNodeInspect(nodeName string) error {
	// 创建集群客户端
	client, err := cluster.NewClient(*kubeconfig, *contextName)
	if err != nil {
		return fmt.Errorf("创建集群客户端失败: %w", err)
	}

	// 创建节点采集器
	collectorInst, err := collector.NewNodeCollector(client)
	if err != nil {
		return fmt.Errorf("创建节点采集器失败: %w", err)
	}

	// 获取集群信息
	clusterName := "default-cluster"
	if *contextName != "" {
		clusterName = *contextName
	}

	// 加载规则配置
	var rulesEngine *rules.Engine
	if *rulesFile != "" {
		// 使用用户提供的规则文件
		rulesEngine, err = rules.NewEngine(*rulesFile)
	} else {
		// 使用默认规则文件
		defaultRulesPath := filepath.Join("configs", "rules", "node.yaml")
		rulesEngine, err = rules.NewEngine(defaultRulesPath)
	}

	if err != nil {
		return fmt.Errorf("加载规则引擎失败: %w", err)
	}

	// 创建分析器并注入采集器
	analyzer := node.NewNodeAnalyzer(rulesEngine, collectorInst)

	// 分析节点
	var results []node.AnalysisResult
	if nodeName != "" {
		// 分析单个节点
		result, err := analyzer.AnalyzeNodeByName(nodeName)
		if err != nil {
			return fmt.Errorf("分析节点 %s 失败: %w", nodeName, err)
		}
		results = []node.AnalysisResult{*result}
	} else {
		// 分析所有节点
		results, err = analyzer.AnalyzeAllNodes()
		if err != nil {
			return fmt.Errorf("分析节点失败: %w", err)
		}
	}

	// 过滤结果（如果只显示有问题的资源）
	if *onlyIssues {
		filteredResults := []node.AnalysisResult{}
		for _, result := range results {
			// 检查是否有未通过的分析项
			hasIssues := false
			for _, item := range result.Items {
				if !item.Passed {
					hasIssues = true
					break
				}
			}
			if hasIssues {
				filteredResults = append(filteredResults, result)
			}
		}
		results = filteredResults
	}

	// 获取规则列表 - 添加空的过滤器参数
	filter := rules.RuleFilter{}
	rulesList := rulesEngine.GetRules(filter)

	// 创建报告生成器
	reportGenerator := report.NewGenerator(clusterName, "")
	nodeReport := reportGenerator.GenerateNodeReport(results, rulesList)

	// 创建格式化器
	var formatter report.Formatter
	switch *outputFormat {
	case "text":
		formatter = report.NewTextFormatter(!*noColor)
	default:
		return fmt.Errorf("不支持的输出格式: %s", *outputFormat)
	}

	// 格式化报告
	output := formatter.Format(nodeReport)

	// 输出报告
	if *outputFile != "" {
		// 写入文件
		err = os.WriteFile(*outputFile, []byte(output), 0644)
		if err != nil {
			return fmt.Errorf("写入报告到文件失败: %w", err)
		}
		fmt.Printf("报告已写入文件: %s\n", *outputFile)
	} else {
		// 输出到标准输出
		fmt.Println(output)
	}

	return nil
} 