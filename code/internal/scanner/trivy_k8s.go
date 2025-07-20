package scanner

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// TrivyScanner Trivy扫描器实现
type TrivyScanner struct {
	binaryManager *TrivyBinaryManager
	cache         *VulnerabilityCache
	kubeconfig    string
	context       string
	options       *ScanOptions
}

// NewTrivyScanner 创建新的Trivy扫描器
func NewTrivyScanner(kubeconfig, context string, options *ScanOptions) (*TrivyScanner, error) {
	if options == nil {
		options = DefaultScanOptions()
	}

	binaryManager := NewTrivyBinaryManager()
	if err := binaryManager.ExtractBinary(); err != nil {
		return nil, fmt.Errorf("初始化Trivy二进制失败: %w", err)
	}

	var cache *VulnerabilityCache
	if !options.NoCache {
		cacheDir := filepath.Join(filepath.Dir(kubeconfig), ".inspector-cache")
		cache = NewVulnerabilityCache(cacheDir, options.CacheTTL)
	}

	return &TrivyScanner{
		binaryManager: binaryManager,
		cache:         cache,
		kubeconfig:    kubeconfig,
		context:       context,
		options:       options,
	}, nil
}

// ScanNamespace 扫描指定命名空间
func (ts *TrivyScanner) ScanNamespace(ctx context.Context, namespace string) (*VulnerabilityReport, error) {
	target := fmt.Sprintf("namespace/%s", namespace)
	
	// 检查缓存
	if ts.cache != nil {
		cacheKey := GenerateCacheKey("namespace", namespace, ts.options)
		if cachedReport, found := ts.cache.Get(cacheKey); found {
			return cachedReport, nil
		}
	}

	// 执行K8s扫描
	report, err := ts.executeK8sScan(ctx, "--namespace", namespace)
	if err != nil {
		return nil, fmt.Errorf("扫描命名空间 %s 失败: %w", namespace, err)
	}

	report.Target = target

	// 缓存结果
	if ts.cache != nil {
		cacheKey := GenerateCacheKey("namespace", namespace, ts.options)
		if err := ts.cache.Set(cacheKey, report); err != nil {
			// 缓存失败不影响主要功能，只记录警告
			fmt.Printf("警告: 缓存扫描结果失败: %v\n", err)
		}
	}

	return report, nil
}

// ScanPod 扫描指定Pod
func (ts *TrivyScanner) ScanPod(ctx context.Context, namespace, podName string) (*VulnerabilityReport, error) {
	target := fmt.Sprintf("pod/%s", podName)
	
	// 检查缓存
	if ts.cache != nil {
		cacheKey := GenerateCacheKey("pod", fmt.Sprintf("%s/%s", namespace, podName), ts.options)
		if cachedReport, found := ts.cache.Get(cacheKey); found {
			return cachedReport, nil
		}
	}

	// 执行K8s扫描
	report, err := ts.executeK8sScan(ctx, "--namespace", namespace, fmt.Sprintf("pod/%s", podName))
	if err != nil {
		return nil, fmt.Errorf("扫描Pod %s/%s 失败: %w", namespace, podName, err)
	}

	report.Target = target

	// 缓存结果
	if ts.cache != nil {
		cacheKey := GenerateCacheKey("pod", fmt.Sprintf("%s/%s", namespace, podName), ts.options)
		if err := ts.cache.Set(cacheKey, report); err != nil {
			fmt.Printf("警告: 缓存扫描结果失败: %v\n", err)
		}
	}

	return report, nil
}

// ScanCluster 扫描整个集群
func (ts *TrivyScanner) ScanCluster(ctx context.Context) (*VulnerabilityReport, error) {
	target := "cluster"
	
	// 检查缓存
	if ts.cache != nil {
		cacheKey := GenerateCacheKey("cluster", "all", ts.options)
		if cachedReport, found := ts.cache.Get(cacheKey); found {
			return cachedReport, nil
		}
	}

	// 执行K8s扫描
	report, err := ts.executeK8sScan(ctx, "cluster")
	if err != nil {
		return nil, fmt.Errorf("扫描集群失败: %w", err)
	}

	report.Target = target

	// 缓存结果
	if ts.cache != nil {
		cacheKey := GenerateCacheKey("cluster", "all", ts.options)
		if err := ts.cache.Set(cacheKey, report); err != nil {
			fmt.Printf("警告: 缓存扫描结果失败: %v\n", err)
		}
	}

	return report, nil
}

// ScanImage 扫描指定镜像（备用方法）
func (ts *TrivyScanner) ScanImage(ctx context.Context, imageRef string) (*VulnerabilityReport, error) {
	target := fmt.Sprintf("image/%s", imageRef)
	
	// 检查缓存
	if ts.cache != nil {
		cacheKey := GenerateCacheKey("image", imageRef, ts.options)
		if cachedReport, found := ts.cache.Get(cacheKey); found {
			return cachedReport, nil
		}
	}

	// 执行镜像扫描
	report, err := ts.executeImageScan(ctx, imageRef)
	if err != nil {
		return nil, fmt.Errorf("扫描镜像 %s 失败: %w", imageRef, err)
	}

	report.Target = target

	// 缓存结果
	if ts.cache != nil {
		cacheKey := GenerateCacheKey("image", imageRef, ts.options)
		if err := ts.cache.Set(cacheKey, report); err != nil {
			fmt.Printf("警告: 缓存扫描结果失败: %v\n", err)
		}
	}

	return report, nil
}

// GetScannerInfo 获取扫描器信息
func (ts *TrivyScanner) GetScannerInfo() ScannerInfo {
	return ScannerInfo{
		Name:    "trivy",
		Version: ts.binaryManager.GetVersion(),
		// DatabaseVersion 需要从实际扫描中获取
	}
}

// IsAvailable 检查扫描器是否可用
func (ts *TrivyScanner) IsAvailable() bool {
	return ts.binaryManager.IsAvailable()
}

// executeK8sScan 执行K8s扫描
func (ts *TrivyScanner) executeK8sScan(ctx context.Context, args ...string) (*VulnerabilityReport, error) {
	binaryPath, err := ts.binaryManager.GetBinaryPath()
	if err != nil {
		return nil, fmt.Errorf("获取Trivy二进制路径失败: %w", err)
	}

	// 构建命令参数
	cmdArgs := []string{"k8s", "--format", "json", "--quiet"}
	
	// 添加kubeconfig和context
	if ts.kubeconfig != "" {
		cmdArgs = append(cmdArgs, "--kubeconfig", ts.kubeconfig)
	}
	if ts.context != "" {
		cmdArgs = append(cmdArgs, "--context", ts.context)
	}

	// 添加扫描器选择
	scanners := []string{"vuln"}
	if ts.options.IncludeConfig {
		scanners = append(scanners, "config")
	}
	cmdArgs = append(cmdArgs, "--scanners", strings.Join(scanners, ","))

	// 添加严重程度过滤
	if len(ts.options.SeverityFilter) > 0 {
		cmdArgs = append(cmdArgs, "--severity", strings.Join(ts.options.SeverityFilter, ","))
	}

	// 添加用户指定的参数
	cmdArgs = append(cmdArgs, args...)

	// 创建命令
	cmd := exec.CommandContext(ctx, binaryPath, cmdArgs...)
	
	// 执行命令
	fmt.Printf("🔍 调试: 执行Trivy命令: %s %v\n", binaryPath, cmdArgs)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			fmt.Printf("🔍 调试: Trivy命令失败，stderr: %s\n", string(exitErr.Stderr))
			return nil, fmt.Errorf("Trivy扫描失败 (退出码: %d): %s",
				exitErr.ExitCode(), string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("执行Trivy命令失败: %w", err)
	}

	fmt.Printf("🔍 调试: Trivy命令成功，输出长度: %d bytes\n", len(output))

	// 解析结果
	return ts.parseTrivyK8sOutput(output)
}

// executeImageScan 执行镜像扫描
func (ts *TrivyScanner) executeImageScan(ctx context.Context, imageRef string) (*VulnerabilityReport, error) {
	binaryPath, err := ts.binaryManager.GetBinaryPath()
	if err != nil {
		return nil, fmt.Errorf("获取Trivy二进制路径失败: %w", err)
	}

	// 构建命令参数
	cmdArgs := []string{"image", "--format", "json", "--quiet"}
	
	// 添加严重程度过滤
	if len(ts.options.SeverityFilter) > 0 {
		cmdArgs = append(cmdArgs, "--severity", strings.Join(ts.options.SeverityFilter, ","))
	}
	
	cmdArgs = append(cmdArgs, imageRef)

	// 创建命令
	cmd := exec.CommandContext(ctx, binaryPath, cmdArgs...)
	
	// 执行命令
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("Trivy镜像扫描失败 (退出码: %d): %s", 
				exitErr.ExitCode(), string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("执行Trivy镜像扫描失败: %w", err)
	}

	// 解析结果
	return ts.parseTrivyImageOutput(output, imageRef)
}
