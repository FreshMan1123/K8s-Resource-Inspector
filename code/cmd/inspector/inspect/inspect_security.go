package inspect

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/analyzer"
	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/cluster"
	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/collector"
	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/report"
	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/rules"
)

// NewSecurityCommand 创建安全检查命令
func NewSecurityCommand(kubecfg, ctx, outFmt *string, noClr, onlyIss *bool, rFile, outFile *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "security [资源类型]",
		Short: "检查Kubernetes资源的安全配置",
		Long: `基于CIS基准和企业安全策略检查资源配置合规性

支持的资源类型:
  pod       检查Pod安全配置
  cis       CIS Kubernetes基准检查
  image     镜像安全策略检查
  all       综合安全检查 (默认)

示例:
  # 检查所有Pod的安全配置
  inspector inspect security pod

  # CIS基准检查
  inspector inspect security cis --rules-file=code/configs/rules/cis-kubernetes-v1.8.yaml

  # 镜像安全检查
  inspector inspect security image --rules-file=configs/rules/image-security.yaml

  # 综合安全检查
  inspector inspect security all --output=json --output-file=security-report.json`,
		Run: func(cmd *cobra.Command, args []string) {
			resourceType := "all"
			if len(args) > 0 {
				resourceType = args[0]
			}

			if err := runSecurityInspect(resourceType, *kubecfg, *ctx, *outFmt, *noClr, *onlyIss, *rFile, *outFile); err != nil {
				fmt.Fprintf(os.Stderr, "安全检查失败: %v\n", err)
				os.Exit(1)
			}
		},
	}

	// 添加子命令
	cmd.AddCommand(NewCISCommand(kubecfg, ctx, outFmt, noClr, onlyIss, rFile, outFile))
	cmd.AddCommand(NewImageSecurityCommand(kubecfg, ctx, outFmt, noClr, onlyIss, rFile, outFile))
	cmd.AddCommand(NewPodSecurityCommand(kubecfg, ctx, outFmt, noClr, onlyIss, rFile, outFile))

	return cmd
}

// NewCISCommand 创建CIS基准检查命令
func NewCISCommand(kubecfg, ctx, outFmt *string, noClr, onlyIss *bool, rFile, outFile *string) *cobra.Command {
	return &cobra.Command{
		Use:   "cis",
		Short: "CIS Kubernetes基准检查",
		Long:  `基于CIS Kubernetes基准进行安全配置检查`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := runCISInspect(*kubecfg, *ctx, *outFmt, *noClr, *onlyIss, *rFile, *outFile); err != nil {
				fmt.Fprintf(os.Stderr, "CIS基准检查失败: %v\n", err)
				os.Exit(1)
			}
		},
	}
}

// NewImageSecurityCommand 创建镜像安全检查命令
func NewImageSecurityCommand(kubecfg, ctx, outFmt *string, noClr, onlyIss *bool, rFile, outFile *string) *cobra.Command {
	return &cobra.Command{
		Use:   "image",
		Short: "镜像安全策略检查",
		Long:  `检查容器镜像的安全配置和策略合规性`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := runImageSecurityInspect(*kubecfg, *ctx, *outFmt, *noClr, *onlyIss, *rFile, *outFile); err != nil {
				fmt.Fprintf(os.Stderr, "镜像安全检查失败: %v\n", err)
				os.Exit(1)
			}
		},
	}
}

// NewPodSecurityCommand 创建Pod安全检查命令
func NewPodSecurityCommand(kubecfg, ctx, outFmt *string, noClr, onlyIss *bool, rFile, outFile *string) *cobra.Command {
	return &cobra.Command{
		Use:   "pod",
		Short: "Pod安全策略检查",
		Long:  `检查Pod和容器的安全配置`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := runPodSecurityInspect(*kubecfg, *ctx, *outFmt, *noClr, *onlyIss, *rFile, *outFile); err != nil {
				fmt.Fprintf(os.Stderr, "Pod安全检查失败: %v\n", err)
				os.Exit(1)
			}
		},
	}
}

// runSecurityInspect 执行安全检查
func runSecurityInspect(resourceType, kubecfg, contextName, outputFormat string, noColor, onlyIssues bool, rulesFile, outputFile string) error {
	switch resourceType {
	case "pod":
		return runPodSecurityInspect(kubecfg, contextName, outputFormat, noColor, onlyIssues, rulesFile, outputFile)
	case "cis":
		return runCISInspect(kubecfg, contextName, outputFormat, noColor, onlyIssues, rulesFile, outputFile)
	case "image":
		return runImageSecurityInspect(kubecfg, contextName, outputFormat, noColor, onlyIssues, rulesFile, outputFile)
	case "all":
		return runComprehensiveSecurityInspect(kubecfg, contextName, outputFormat, noColor, onlyIssues, rulesFile, outputFile)
	default:
		return fmt.Errorf("不支持的资源类型: %s", resourceType)
	}
}

// runPodSecurityInspect 执行Pod安全检查
func runPodSecurityInspect(kubecfg, contextName, outputFormat string, noColor, onlyIssues bool, rulesFile, outputFile string) error {
	// 创建集群客户端
	client, err := cluster.NewClient(kubecfg, contextName)
	if err != nil {
		return fmt.Errorf("创建集群客户端失败: %w", err)
	}

	// 创建Pod收集器
	podCollector, err := collector.NewPodCollector(client)
	if err != nil {
		return fmt.Errorf("创建Pod收集器失败: %w", err)
	}

	// 加载规则引擎
	var engine *rules.Engine
	if rulesFile != "" {
		engine, err = rules.NewEngine(rulesFile)
		if err != nil {
			return fmt.Errorf("加载规则文件失败: %w", err)
		}
	} else {
		// 使用默认的Pod安全规则
		engine, err = rules.NewEngine("code/configs/rules/pod-security.yaml")
		if err != nil {
			return fmt.Errorf("无法加载默认规则文件: %w", err)
		}
	}

	// 创建安全分析器
	securityAnalyzer := analyzer.NewSecurityAnalyzer(podCollector, engine)

	// 执行安全分析 - 检查所有命名空间
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 获取所有命名空间的Pod进行安全分析
	allResults, err := securityAnalyzer.AnalyzeAllNamespacesSecurity(ctx)
	if err != nil {
		return fmt.Errorf("安全分析失败: %w", err)
	}

	// 获取规则列表
	filter := rules.RuleFilter{}
	rulesList := engine.GetRules(filter)

	// 创建报告生成器
	clusterName := "default-cluster"
	if contextName != "" {
		clusterName = contextName
	} else {
		// 如果没有指定context，尝试获取当前上下文
		if currentContext, err := cluster.GetCurrentContext(kubecfg); err == nil {
			clusterName = currentContext
		}
	}
	reportGenerator := report.NewGenerator(clusterName, "all-namespaces")
	securityReport := reportGenerator.GenerateSecurityReport(allResults, rulesList)

	// 创建格式化器
	var formatter report.Formatter
	switch outputFormat {
	case "text":
		formatter = report.NewTextFormatter(!noColor)
	default:
		return fmt.Errorf("不支持的输出格式: %s", outputFormat)
	}

	// 格式化报告
	output := formatter.Format(securityReport)

	// 输出报告
	if outputFile != "" {
		// 写入文件
		err = os.WriteFile(outputFile, []byte(output), 0644)
		if err != nil {
			return fmt.Errorf("写入报告到文件失败: %w", err)
		}
		fmt.Printf("报告已写入文件: %s\n", outputFile)
		// 同时也输出到控制台
		fmt.Println("\n检查结果:")
		fmt.Println(output)
	} else {
		// 输出到标准输出
		fmt.Println(output)
	}

	return nil
}

// runCISInspect 执行CIS基准检查
func runCISInspect(kubecfg, contextName, outputFormat string, noColor, onlyIssues bool, rulesFile, outputFile string) error {
	if rulesFile == "" {
		rulesFile = "code/configs/rules/cis-kubernetes-v1.8.yaml"
	}
	return runPodSecurityInspect(kubecfg, contextName, outputFormat, noColor, onlyIssues, rulesFile, outputFile)
}

// runImageSecurityInspect 执行镜像安全检查
func runImageSecurityInspect(kubecfg, contextName, outputFormat string, noColor, onlyIssues bool, rulesFile, outputFile string) error {
	if rulesFile == "" {
		rulesFile = "code/configs/rules/image-security.yaml"
	}
	return runPodSecurityInspect(kubecfg, contextName, outputFormat, noColor, onlyIssues, rulesFile, outputFile)
}

// runComprehensiveSecurityInspect 执行综合安全检查
func runComprehensiveSecurityInspect(kubecfg, contextName, outputFormat string, noColor, onlyIssues bool, rulesFile, outputFile string) error {
	// 依次执行各种安全检查
	fmt.Println("执行综合安全检查...")

	// CIS基准检查
	fmt.Println("1. CIS基准检查...")
	if err := runCISInspect(kubecfg, contextName, outputFormat, noColor, onlyIssues, "", ""); err != nil {
		fmt.Printf("CIS基准检查失败: %v\n", err)
	} else {
		fmt.Println("CIS基准检查完成")
	}

	// 镜像安全检查
	fmt.Println("2. 镜像安全检查...")
	if err := runImageSecurityInspect(kubecfg, contextName, outputFormat, noColor, onlyIssues, "", ""); err != nil {
		fmt.Printf("镜像安全检查失败: %v\n", err)
	} else {
		fmt.Println("镜像安全检查完成")
	}

	// Pod安全检查
	fmt.Println("3. Pod安全检查...")
	if err := runPodSecurityInspect(kubecfg, contextName, outputFormat, noColor, onlyIssues, "", ""); err != nil {
		fmt.Printf("Pod安全检查失败: %v\n", err)
	} else {
		fmt.Println("Pod安全检查完成")
	}

	fmt.Println("综合安全检查完成")
	return nil
}


