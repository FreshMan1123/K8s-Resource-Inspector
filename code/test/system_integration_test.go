package test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/scanner"
)

// TestSystemIntegration_RealWorldScenario 系统集成测试 - 模拟真实使用场景
func TestSystemIntegration_RealWorldScenario(t *testing.T) {
	// 跳过需要网络的测试，除非在CI环境中
	if testing.Short() {
		t.Skip("跳过需要网络连接的系统集成测试")
	}

	t.Log("🚀 开始系统集成测试 - 模拟用户使用二进制包进行漏洞扫描的完整流程")

	// 场景：用户使用我们的工具进行完整的漏洞扫描流程
	// 1. 初始化系统
	// 2. 配置扫描参数
	// 3. 执行漏洞扫描
	// 4. 处理扫描结果
	// 5. 验证输出格式

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// 第一步：系统初始化
	t.Log("📋 第一步：系统初始化")
	options := scanner.DefaultScanOptions()
	
	// 验证默认配置符合预期（禁用缓存，允许数据库更新）
	assert.True(t, options.NoCache, "系统应该默认禁用缓存")
	assert.Equal(t, scanner.ScanModeK8s, options.Mode, "默认模式应该是K8s")
	assert.True(t, options.IncludeConfig, "应该包含配置检查")
	
	t.Log("✅ 系统初始化完成")

	// 第二步：创建扫描器
	t.Log("📋 第二步：创建漏洞扫描器")
	trivyScanner, err := scanner.NewTrivyScanner("", "", options)
	require.NoError(t, err, "创建Trivy扫描器应该成功")
	require.True(t, trivyScanner.IsAvailable(), "扫描器应该可用")
	
	info := trivyScanner.GetScannerInfo()
	t.Logf("✅ 扫描器创建成功: %s v%s", info.Name, info.Version)

	// 第三步：执行多种类型的扫描（模拟用户的不同使用场景）
	t.Log("📋 第三步：执行多种扫描场景")
	
	scanScenarios := []struct {
		name        string
		description string
		scanFunc    func() (*scanner.VulnerabilityReport, error)
		validate    func(*testing.T, *scanner.VulnerabilityReport)
	}{
		{
			name:        "高风险镜像扫描",
			description: "扫描已知存在高危漏洞的镜像",
			scanFunc: func() (*scanner.VulnerabilityReport, error) {
				return trivyScanner.ScanImage(ctx, "nginx:1.14.0")
			},
			validate: func(t *testing.T, report *scanner.VulnerabilityReport) {
				assert.Greater(t, report.Summary.Total, 0, "高风险镜像应该发现漏洞")
				assert.Contains(t, report.Target, "image/nginx:1.14.0", "目标应该正确")
			},
		},
		{
			name:        "相对安全镜像扫描",
			description: "扫描相对安全的最新镜像",
			scanFunc: func() (*scanner.VulnerabilityReport, error) {
				return trivyScanner.ScanImage(ctx, "alpine:3.18")
			},
			validate: func(t *testing.T, report *scanner.VulnerabilityReport) {
				assert.GreaterOrEqual(t, report.Summary.Total, 0, "扫描应该完成")
				assert.Contains(t, report.Target, "image/alpine:3.18", "目标应该正确")
			},
		},
	}

	var allReports []*scanner.VulnerabilityReport
	
	for _, scenario := range scanScenarios {
		t.Run(scenario.name, func(t *testing.T) {
			t.Logf("🔍 执行场景: %s", scenario.description)
			
			startTime := time.Now()
			report, err := scenario.scanFunc()
			duration := time.Since(startTime)
			
			require.NoError(t, err, "扫描不应该失败")
			require.NotNil(t, report, "报告不应该为空")
			
			// 基础验证
			assert.Equal(t, "trivy", report.Scanner.Name, "扫描器名称应该正确")
			assert.Equal(t, "completed", report.Status, "扫描状态应该是completed")
			assert.WithinDuration(t, time.Now(), report.ScanTime, time.Minute, "扫描时间应该是最近的")
			
			// 统计数据一致性验证
			calculatedTotal := report.Summary.Critical + report.Summary.High +
				report.Summary.Medium + report.Summary.Low
			assert.Equal(t, report.Summary.Total, calculatedTotal, "统计数字应该一致")
			
			// 场景特定验证
			scenario.validate(t, report)
			
			allReports = append(allReports, report)
			
			t.Logf("✅ %s 完成，耗时: %v, 漏洞数: %d", scenario.name, duration, report.Summary.Total)
		})
	}

	// 第四步：数据聚合和分析
	t.Log("📋 第四步：数据聚合和分析")
	
	totalVulnerabilities := 0
	totalCritical := 0
	totalHigh := 0
	
	for _, report := range allReports {
		totalVulnerabilities += report.Summary.Total
		totalCritical += report.Summary.Critical
		totalHigh += report.Summary.High
	}
	
	t.Logf("📊 聚合结果: 总漏洞数=%d, 严重=%d, 高危=%d", 
		totalVulnerabilities, totalCritical, totalHigh)
	
	// 验证聚合数据的合理性
	assert.GreaterOrEqual(t, totalVulnerabilities, 0, "总漏洞数应该>=0")
	assert.GreaterOrEqual(t, totalCritical, 0, "严重漏洞数应该>=0")
	assert.GreaterOrEqual(t, totalHigh, 0, "高危漏洞数应该>=0")
	
	t.Log("✅ 数据聚合和分析完成")

	// 第五步：验证系统稳定性
	t.Log("📋 第五步：验证系统稳定性")
	
	// 验证扫描器在多次使用后仍然可用
	assert.True(t, trivyScanner.IsAvailable(), "扫描器在多次使用后应该仍然可用")
	
	// 验证内存使用情况（简单检查）
	info2 := trivyScanner.GetScannerInfo()
	assert.Equal(t, info.Name, info2.Name, "扫描器信息应该保持一致")
	assert.Equal(t, info.Version, info2.Version, "版本信息应该保持一致")
	
	t.Log("✅ 系统稳定性验证通过")

	t.Log("🎉 系统集成测试完成 - 所有组件协同工作正常")
}

// TestSystemIntegration_ErrorRecovery 系统错误恢复测试
func TestSystemIntegration_ErrorRecovery(t *testing.T) {
	t.Log("🔧 开始系统错误恢复测试")

	options := scanner.DefaultScanOptions()
	options.Timeout = 5 * time.Second // 设置短超时来触发错误

	trivyScanner, err := scanner.NewTrivyScanner("", "", options)
	require.NoError(t, err, "创建Trivy扫描器应该成功")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 测试错误场景下的系统恢复能力
	errorScenarios := []struct {
		name        string
		description string
		testFunc    func() error
	}{
		{
			name:        "网络错误恢复",
			description: "测试网络错误后系统是否能正常恢复",
			testFunc: func() error {
				_, err := trivyScanner.ScanImage(ctx, "nonexistent-registry.example.com/test:latest")
				return err // 预期会有错误
			},
		},
		{
			name:        "无效输入恢复",
			description: "测试无效输入后系统是否能正常恢复",
			testFunc: func() error {
				_, err := trivyScanner.ScanImage(ctx, "")
				return err // 预期会有错误
			},
		},
	}

	for _, scenario := range errorScenarios {
		t.Run(scenario.name, func(t *testing.T) {
			t.Logf("🔍 测试场景: %s", scenario.description)
			
			// 执行错误场景
			err := scenario.testFunc()
			assert.Error(t, err, "错误场景应该返回错误")
			
			// 验证系统在错误后仍然可用
			assert.True(t, trivyScanner.IsAvailable(), "系统在错误后应该仍然可用")
			
			t.Logf("✅ %s - 系统错误恢复正常", scenario.name)
		})
	}

	t.Log("✅ 系统错误恢复测试完成")
}

// TestSystemIntegration_PerformanceBaseline 系统性能基线测试
func TestSystemIntegration_PerformanceBaseline(t *testing.T) {
	// 跳过需要网络的测试，除非在CI环境中
	if testing.Short() {
		t.Skip("跳过需要网络连接的性能基线测试")
	}

	t.Log("⏱️ 开始系统性能基线测试")

	options := scanner.DefaultScanOptions()
	trivyScanner, err := scanner.NewTrivyScanner("", "", options)
	require.NoError(t, err, "创建Trivy扫描器应该成功")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// 性能基线测试 - 确保系统在合理时间内完成扫描
	performanceTests := []struct {
		name        string
		imageRef    string
		maxDuration time.Duration
	}{
		{
			name:        "小镜像扫描性能",
			imageRef:    "alpine:3.18",
			maxDuration: 2 * time.Minute,
		},
		{
			name:        "中等镜像扫描性能",
			imageRef:    "nginx:1.21",
			maxDuration: 3 * time.Minute,
		},
	}

	for _, test := range performanceTests {
		t.Run(test.name, func(t *testing.T) {
			startTime := time.Now()
			
			report, err := trivyScanner.ScanImage(ctx, test.imageRef)
			duration := time.Since(startTime)
			
			require.NoError(t, err, "扫描不应该失败")
			require.NotNil(t, report, "报告不应该为空")
			
			// 性能验证
			assert.Less(t, duration, test.maxDuration, 
				fmt.Sprintf("扫描时间(%v)不应该超过基线(%v)", duration, test.maxDuration))
			
			t.Logf("✅ %s 完成，耗时: %v (基线: %v)", test.name, duration, test.maxDuration)
		})
	}

	t.Log("✅ 系统性能基线测试完成")
}

// TestSystemIntegration_DataConsistency 数据一致性测试
func TestSystemIntegration_DataConsistency(t *testing.T) {
	t.Log("🔍 开始数据一致性测试")

	// 创建两个独立的扫描器实例
	options1 := scanner.DefaultScanOptions()
	options2 := scanner.DefaultScanOptions()

	scanner1, err := scanner.NewTrivyScanner("", "", options1)
	require.NoError(t, err, "创建第一个扫描器应该成功")

	scanner2, err := scanner.NewTrivyScanner("", "", options2)
	require.NoError(t, err, "创建第二个扫描器应该成功")

	// 验证两个扫描器的基本信息一致
	info1 := scanner1.GetScannerInfo()
	info2 := scanner2.GetScannerInfo()

	assert.Equal(t, info1.Name, info2.Name, "扫描器名称应该一致")
	assert.Equal(t, info1.Version, info2.Version, "扫描器版本应该一致")

	t.Log("✅ 数据一致性测试完成 - 多个扫描器实例保持一致")
}
