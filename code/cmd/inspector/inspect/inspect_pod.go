package inspect

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/analyzer/pod"
	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/cluster"
	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/report"
	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/rules"
	"github.com/spf13/cobra"
)

var (
	// Pod命令特有的选项
	namespace  string
	fetchLogs  bool
	logLines   int
	liveLogs   bool
)

// NewPodCommand 创建Pod检查命令
func NewPodCommand(kubecfg, ctx, outFmt *string, noClr, onlyIss *bool, rFile, outFile *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pod [pod名称] [-n 命名空间]",
		Short: "检查Pod资源并生成报告",
		Long:  `检查Kubernetes集群中的Pod资源状态并生成详细报告，包括状态、资源使用情况和潜在问题。`,
		Run: func(cmd *cobra.Command, args []string) {
			// 检查是否指定了Pod名称
			podName := ""
			if len(args) > 0 {
				podName = args[0]
			}

			if err := runPodInspect(podName, namespace, *kubecfg, *contextName, *outputFormat, *noColor, *onlyIssues, *rulesFile, *outputFile, fetchLogs, logLines, liveLogs); err != nil {
				fmt.Fprintf(os.Stderr, "检查Pod失败: %v\n", err)
				os.Exit(1)
			}
		},
	}

	// 添加Pod特有的标志
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "要检查的命名空间")
	cmd.Flags().BoolVar(&fetchLogs, "fetch-logs", false, "获取Pod日志")
	cmd.Flags().IntVar(&logLines, "log-lines", 50, "获取的日志行数")
	cmd.Flags().BoolVar(&liveLogs, "live-logs", false, "对问题Pod实时获取最新日志")

	return cmd
}

// runPodInspect 执行Pod检查逻辑
func runPodInspect(podName, namespace, kubeconfig, contextName, outputFormat string, noColor, onlyIssues bool, rulesFile, outputFile string, fetchLogs bool, logLines int, liveLogs bool) error {
	// 创建集群客户端
	client, err := cluster.NewClient(kubeconfig, contextName)
	if err != nil {
		return fmt.Errorf("创建集群客户端失败: %w", err)
	}

	// 获取集群信息
	clusterName := "default-cluster"
	if contextName != "" {
		clusterName = contextName
	}

	// 加载规则配置
	var rulesEngine *rules.Engine
	if rulesFile != "" {
		// 使用用户提供的规则文件
		rulesEngine, err = rules.NewEngine(rulesFile)
	} else {
		// 使用默认规则文件
		defaultRulesPath := filepath.Join("configs", "rules", "pod.yaml")
		rulesEngine, err = rules.NewEngine(defaultRulesPath)
	}

	if err != nil {
		return fmt.Errorf("加载规则引擎失败: %w", err)
	}

	// 创建分析器并设置客户端
	analyzer := pod.NewPodAnalyzer(rulesEngine)
	analyzer.SetClient(client)

	// 分析Pod
	var results []*pod.AnalysisResult
	if podName != "" {
		// 分析单个Pod
		result, err := analyzer.AnalyzePodByName(namespace, podName)
		if err != nil {
			return fmt.Errorf("分析Pod %s/%s 失败: %w", namespace, podName, err)
		}
		results = []*pod.AnalysisResult{result}

		// 如果需要获取日志
		if fetchLogs {
			for _, container := range result.Containers {
				logs, err := client.GetPodLogs(namespace, podName, container.Name, logLines)
				if err != nil {
					fmt.Printf("警告: 获取容器 %s 日志失败: %v\n", container.Name, err)
					continue
				}
				fmt.Printf("容器 %s 日志:\n", container.Name)
				for _, line := range logs {
					fmt.Println(line)
				}
			}
		}
	} else {
		// 分析命名空间中的所有Pod
		results, err = analyzer.AnalyzePodsInNamespace(namespace)
		if err != nil {
			return fmt.Errorf("分析命名空间 %s 中的Pod失败: %w", namespace, err)
		}
	}

	// 过滤结果（如果只显示有问题的资源）
	if onlyIssues {
		filteredResults := []*pod.AnalysisResult{}
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

	// 获取规则列表
	filter := rules.RuleFilter{}
	rulesList := rulesEngine.GetRules(filter)

	// 创建报告生成器
	reportGenerator := report.NewGenerator(clusterName, namespace)
	podReport := reportGenerator.GeneratePodReport(results, rulesList)

	// 创建格式化器
	var formatter report.Formatter
	switch outputFormat {
	case "text":
		formatter = report.NewTextFormatter(!noColor)
	default:
		return fmt.Errorf("不支持的输出格式: %s", outputFormat)
	}

	// 格式化报告
	output := formatter.Format(podReport)

	// 输出报告
	if outputFile != "" {
		// 写入文件
		err = os.WriteFile(outputFile, []byte(output), 0644)
		if err != nil {
			return fmt.Errorf("写入报告到文件失败: %w", err)
		}
		fmt.Printf("报告已写入文件: %s\n", outputFile)
	} else {
		// 输出到标准输出
		fmt.Println(output)
	}

	// 如果启用了实时日志并且有问题Pod
	if liveLogs && onlyIssues && len(results) > 0 {
		fmt.Println("\n=== 问题Pod的实时日志 ===")
		for _, result := range results {
			for _, item := range result.Items {
				if !item.Passed {
					fmt.Printf("\nPod %s/%s 有问题: %s\n", result.Namespace, result.PodName, item.Description)
					
					// 获取该Pod的所有容器
					for _, container := range result.Containers {
						logs, err := client.GetPodLogs(result.Namespace, result.PodName, container.Name, logLines)
						if err != nil {
							fmt.Printf("警告: 获取容器 %s 日志失败: %v\n", container.Name, err)
							continue
						}
						fmt.Printf("容器 %s 最新日志:\n", container.Name)
						for _, line := range logs {
							fmt.Println(line)
						}
						fmt.Println()
					}
					
					break
				}
			}
		}
	}

	return nil
} 