package models

import "time"

// Rule 规则模型 (与现有YAML结构保持一致)
type Rule struct {
	ID          string         `yaml:"id" json:"id"`
	Name        string         `yaml:"name" json:"name"`
	Description string         `yaml:"description" json:"description"`
	Category    string         `yaml:"category" json:"category"`
	Condition   RuleCondition  `yaml:"condition" json:"condition"`
	Severity    string         `yaml:"severity" json:"severity"`
	Remediation string         `yaml:"remediation" json:"remediation"`
	Enabled     bool           `yaml:"enabled" json:"enabled"`
	CreatedAt   time.Time      `yaml:"created_at,omitempty" json:"created_at,omitempty"`
	UpdatedAt   time.Time      `yaml:"updated_at,omitempty" json:"updated_at,omitempty"`
}

// RuleCondition 规则条件
type RuleCondition struct {
	Metric    string      `yaml:"metric" json:"metric"`
	Operator  string      `yaml:"operator" json:"operator"`
	Threshold interface{} `yaml:"threshold" json:"threshold"`
	Unit      string      `yaml:"unit,omitempty" json:"unit,omitempty"`
}

// RulesConfig 规则配置文件结构
type RulesConfig struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Rules      []Rule `yaml:"rules"`
}

// RuleCategory 规则分类信息
type RuleCategory struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Count       int    `json:"count"`
}

// APIResponse 统一API响应格式
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}
