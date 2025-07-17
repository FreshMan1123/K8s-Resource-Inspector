package rules

import (
	"fmt"
	"os"

	"time"

	"gopkg.in/yaml.v2"
)

// RuleLoader 规则加载器
type RuleLoader struct {
	// 规则文件路径
	rulesFile string
	// 上次修改时间
	lastModified time.Time
	// 已加载的规则配置
	config *RulesConfig
}

// NewRuleLoader 创建规则加载器
func NewRuleLoader(rulesFile string) *RuleLoader {
	return &RuleLoader{
		rulesFile: rulesFile,
	}
}

// LoadRules 加载规则
func (rl *RuleLoader) LoadRules() error {
	// 检查文件是否存在
	info, err := os.Stat(rl.rulesFile)
	if os.IsNotExist(err) {
		return fmt.Errorf("规则文件不存在: %s", rl.rulesFile)
	}
	if err != nil {
		return fmt.Errorf("检查规则文件失败: %w", err)
	}

	// 检查文件是否已修改
	if rl.config != nil && info.ModTime().Equal(rl.lastModified) {
		// 文件未修改，使用已加载的配置
		return nil
	}

	// 读取文件内容
	data, err := os.ReadFile(rl.rulesFile)
	if err != nil {
		return fmt.Errorf("读取规则文件失败: %w", err)
	}

	// 解析YAML
	var config RulesConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("解析规则文件YAML失败: %w", err)
	}

	// 验证配置
	if err := validateConfig(&config); err != nil {
		return fmt.Errorf("验证规则配置失败: %w", err)
	}

	// 更新配置和修改时间
	rl.config = &config
	rl.lastModified = info.ModTime()

	return nil
}

// GetRulesConfig 获取规则配置
func (rl *RuleLoader) GetRulesConfig() *RulesConfig {
	return rl.config
}

// GetRules 获取规则列表，可以根据过滤条件筛选
func (rl *RuleLoader) GetRules(filter RuleFilter) []Rule {
	if rl.config == nil {
		return nil
	}

	var result []Rule
	
	// 遍历所有规则
	for _, rule := range rl.config.Rules {
		// 应用过滤条件
		if matchesFilter(rule, filter) {
			result = append(result, rule)
		}
	}

	return result
}

// GetEnvironment 获取环境设置
func (rl *RuleLoader) GetEnvironment(clusterName string) string {
	if rl.config == nil {
		return "prod" // 默认使用生产环境
	}

	// 从集群环境映射中查找
	if env, exists := rl.config.ClusterEnvironments[clusterName]; exists {
		return env
	}

	// 返回默认环境
	if defaultEnv, exists := rl.config.ClusterEnvironments["default"]; exists {
		return defaultEnv
	}

	// 如果没有默认环境，使用配置中的环境
	if rl.config.Config.Environment != "" {
		return rl.config.Config.Environment
	}

	// 最终默认值
	return "prod"
}

// matchesFilter 检查规则是否符合过滤条件
func matchesFilter(rule Rule, filter RuleFilter) bool {
	// 检查类别
	if len(filter.Categories) > 0 {
		categoryMatch := false
		for _, category := range filter.Categories {
			if rule.Category == category {
				categoryMatch = true
				break
			}
		}
		if !categoryMatch {
			return false
		}
	}

	// 检查严重程度
	if len(filter.Severities) > 0 {
		severityMatch := false
		for _, severity := range filter.Severities {
			if rule.Severity == severity {
				severityMatch = true
				break
			}
		}
		if !severityMatch {
			return false
		}
	}

	// 检查是否启用
	if filter.Enabled != nil && rule.Enabled != *filter.Enabled {
		return false
	}

	return true
}

// validateConfig 验证规则配置
func validateConfig(config *RulesConfig) error {
	// 检查API版本和类型
	if config.APIVersion == "" {
		return fmt.Errorf("缺少apiVersion字段")
	}
	if config.Kind == "" {
		return fmt.Errorf("缺少kind字段")
	}
	if config.Kind != "RulesConfig" {
		return fmt.Errorf("不支持的kind: %s", config.Kind)
	}

	// 检查每条规则
	for i, rule := range config.Rules {
		if rule.ID == "" {
			return fmt.Errorf("第 %d 条规则缺少ID", i+1)
		}
		if rule.Name == "" {
			return fmt.Errorf("规则 '%s' 缺少名称", rule.ID)
		}
		if rule.Category == "" {
			return fmt.Errorf("规则 '%s' 缺少类别", rule.ID)
		}
		if rule.Severity == "" {
			return fmt.Errorf("规则 '%s' 缺少严重程度", rule.ID)
		}
		if rule.Condition.Metric == "" {
			return fmt.Errorf("规则 '%s' 缺少指标", rule.ID)
		}
		if rule.Condition.Operator == "" {
			return fmt.Errorf("规则 '%s' 缺少操作符", rule.ID)
		}
		if rule.Condition.Threshold == nil && len(rule.Condition.Thresholds) == 0 {
			return fmt.Errorf("规则 '%s' 缺少阈值", rule.ID)
		}
		
		// 验证操作符是否支持
		if !isValidOperator(rule.Condition.Operator) {
			return fmt.Errorf("规则 '%s' 包含不支持的操作符: %s", rule.ID, rule.Condition.Operator)
		}
	}

	return nil
}

// isValidOperator 检查操作符是否有效
func isValidOperator(op string) bool {
	validOps := map[string]bool{
		">":            true,
		">=":           true,
		"<":            true,
		"<=":           true,
		"==":           true,
		"!=":           true,
		"contains":     true,
		"matches":      true,
		"has_non_empty": true,
	}
	return validOps[op]
}