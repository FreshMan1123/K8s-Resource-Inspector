package rules

import (
	"time"
)

// Rule 表示一条检查规则
type Rule struct {
	// 规则ID
	ID string `yaml:"id" json:"id"`
	// 规则名称
	Name string `yaml:"name" json:"name"`
	// 规则描述
	Description string `yaml:"description" json:"description"`
	// 规则类别：node, pod, deployment等
	Category string `yaml:"category" json:"category"`
	// 严重程度：critical, warning, info
	Severity string `yaml:"severity" json:"severity"`
	// 触发条件
	Condition RuleCondition `yaml:"condition" json:"condition"`
	// 修复建议
	Remediation string `yaml:"remediation" json:"remediation"`
	// 是否启用
	Enabled bool `yaml:"enabled" json:"enabled"`
	// 创建时间
	CreatedAt time.Time `yaml:"created_at,omitempty" json:"created_at,omitempty"`
	// 更新时间
	UpdatedAt time.Time `yaml:"updated_at,omitempty" json:"updated_at,omitempty"`
}

// RuleCondition 表示规则的触发条件
type RuleCondition struct {
	// 要检查的指标
	Metric string `yaml:"metric" json:"metric"`
	// 比较操作符：>, >=, <, <=, ==, !=, contains
	Operator string `yaml:"operator" json:"operator"`
	// 默认阈值，如不使用环境特定阈值则使用此值
	Threshold interface{} `yaml:"threshold" json:"threshold"`
	// 环境特定阈值
	Thresholds map[string]interface{} `yaml:"thresholds,omitempty" json:"thresholds,omitempty"`
	// 持续时间（可选，用于某些需要持续一段时间的条件）
	Duration *time.Duration `yaml:"duration,omitempty" json:"duration,omitempty"`
}

// RuleResult 表示规则评估结果
type RuleResult struct {
	// 规则ID
	RuleID string `json:"rule_id"`
	// 规则名称
	RuleName string `json:"rule_name"`
	// 检查结果：通过(true)或失败(false)
	// Passed=true 表示检查通过，规则条件未被触发（如CPU使用率<阈值）
	// Passed=false 表示检查失败，规则条件被触发（如CPU使用率>=阈值）
	Passed bool `json:"passed"`
	// 实际值
	ActualValue interface{} `json:"actual_value"`
	// 期望值
	ExpectedValue interface{} `json:"expected_value"`
	// 评估消息
	Message string `json:"message"`
	// 修复建议
	Remediation string `json:"remediation"`
	// 严重程度
	Severity string `json:"severity"`
	// 评估时间
	EvaluatedAt time.Time `json:"evaluated_at"`
}

// RuleSet 表示一组规则
type RuleSet struct {
	// 规则集名称
	Name string `yaml:"name" json:"name"`
	// 规则列表
	Rules []Rule `yaml:"rules" json:"rules"`
}

// RulesConfig 表示规则配置文件
type RulesConfig struct {
	// API版本
	APIVersion string `yaml:"apiVersion" json:"apiVersion"`
	// 类型
	Kind string `yaml:"kind" json:"kind"`
	// 通用配置
	Config struct {
		// 是否自动重载配置
		AutoReload bool `yaml:"autoReload" json:"autoReload"`
		// 重载间隔
		ReloadInterval string `yaml:"reloadInterval" json:"reloadInterval"`
		// 当前环境
		Environment string `yaml:"environment" json:"environment"`
	} `yaml:"config" json:"config"`
	// 集群环境映射
	ClusterEnvironments map[string]string `yaml:"clusterEnvironments" json:"clusterEnvironments"`
	// 规则列表
	Rules []Rule `yaml:"rules" json:"rules"`
}

// RuleFilter 用于过滤规则
type RuleFilter struct {
	Categories []string
	Severities []string
	Enabled    *bool
}

// Validator 指标验证接口
type Validator interface {
	// Validate 验证指标值是否符合规则条件
	Validate(metric string, actualValue interface{}, condition RuleCondition, env string) (bool, error)
	// FormatValue 格式化值用于显示
	FormatValue(value interface{}) string
} 