package test

import (
	"testing"

	"github.com/FreshMan1123/k8s-resource-inspector/code/internal/rules"
)

// TestMapValidatorFix 测试MapValidator修复后的功能
func TestMapValidatorFix(t *testing.T) {
	// 创建规则引擎
	rulesEngine, err := rules.NewEngine("testdata/deployment_rules_test.yaml")
	if err != nil {
		t.Fatalf("创建规则引擎失败: %v", err)
	}

	// 获取owner标签规则
	ruleFilter := rules.RuleFilter{
		Categories: []string{"deployment"},
	}
	rulesList := rulesEngine.GetRules(ruleFilter)
	
	var ownerRule rules.Rule
	found := false
	for _, rule := range rulesList {
		if rule.ID == "require_owner_label" {
			ownerRule = rule
			found = true
			break
		}
	}
	
	if !found {
		t.Fatalf("未找到require_owner_label规则")
	}

	tests := []struct {
		name         string
		actualValue  interface{}
		expectedPass bool
		expectError  bool
	}{
		{
			name: "正常情况 - 有owner标签且值不为空",
			actualValue: map[string]string{
				"owner": "team-backend",
				"app":   "web",
			},
			expectedPass: true,
			expectError:  false,
		},
		{
			name: "缺少owner标签",
			actualValue: map[string]string{
				"app": "web",
				"env": "prod",
			},
			expectedPass: false,
			expectError:  false,
		},
		{
			name: "owner标签值为空字符串",
			actualValue: map[string]string{
				"owner": "",
				"app":   "web",
			},
			expectedPass: false,
			expectError:  false,
		},
		{
			name: "owner标签值为纯空格",
			actualValue: map[string]string{
				"owner": "   ",
				"app":   "web",
			},
			expectedPass: false,
			expectError:  false,
		},
		{
			name: "owner标签值包含空格但不全是空格",
			actualValue: map[string]string{
				"owner": " team-backend ",
				"app":   "web",
			},
			expectedPass: true,
			expectError:  false,
		},
		{
			name: "测试类型转换 - 应该不再出现map类型断言失败",
			actualValue: map[string]string{
				"owner": "team-frontend",
			},
			expectedPass: true,
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := rulesEngine.EvaluateRule(ownerRule, "map", tt.actualValue)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("期望出现错误，但没有错误")
				}
				return
			}
			
			if err != nil {
				t.Errorf("不期望出现错误，但出现了错误: %v", err)
				return
			}
			
			if result.Passed != tt.expectedPass {
				t.Errorf("验证结果不符合期望，期望: %v, 实际: %v, 消息: %s", 
					tt.expectedPass, result.Passed, result.Message)
			}
			
			// 验证结果的其他字段
			if result.RuleID != ownerRule.ID {
				t.Errorf("规则ID不匹配，期望: %s, 实际: %s", ownerRule.ID, result.RuleID)
			}
			
			if result.RuleName != ownerRule.Name {
				t.Errorf("规则名称不匹配，期望: %s, 实际: %s", ownerRule.Name, result.RuleName)
			}
			
			if result.Severity != ownerRule.Severity {
				t.Errorf("严重程度不匹配，期望: %s, 实际: %s", ownerRule.Severity, result.Severity)
			}
		})
	}
}

// TestMapValidatorTypeConversion 专门测试类型转换功能
func TestMapValidatorTypeConversion(t *testing.T) {
	// 创建规则引擎来获取MapValidator
	rulesEngine, err := rules.NewEngine("testdata/deployment_rules_test.yaml")
	if err != nil {
		t.Fatalf("创建规则引擎失败: %v", err)
	}

	// 获取MapValidator
	validator, err := rulesEngine.GetValidator("map")
	if err != nil {
		t.Fatalf("获取MapValidator失败: %v", err)
	}

	// 创建测试条件
	condition := rules.RuleCondition{
		Metric:   "has_labels",
		Operator: "has_non_empty",
		Threshold: map[string]interface{}{
			"owner": "",
		},
	}
	
	tests := []struct {
		name         string
		actualValue  interface{}
		expectedPass bool
		expectError  bool
		errorContains string
	}{
		{
			name: "map[string]string类型",
			actualValue: map[string]string{
				"owner": "team-a",
			},
			expectedPass: true,
			expectError:  false,
		},
		{
			name: "错误的actualValue类型",
			actualValue: "not a map",
			expectedPass: false,
			expectError:  true,
			errorContains: "actualValue类型断言失败",
		},
		{
			name: "threshold为map[string]interface{}类型",
			actualValue: map[string]string{
				"owner": "team-b",
			},
			expectedPass: true,
			expectError:  false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validator.Validate(condition.Metric, tt.actualValue, condition, "test")
			
			if tt.expectError {
				if err == nil {
					t.Errorf("期望出现错误，但没有错误")
					return
				}
				if tt.errorContains != "" && !containsString(err.Error(), tt.errorContains) {
					t.Errorf("错误信息不包含期望的内容，期望包含: %s, 实际错误: %s", 
						tt.errorContains, err.Error())
				}
				return
			}
			
			if err != nil {
				t.Errorf("不期望出现错误，但出现了错误: %v", err)
				return
			}
			
			if result != tt.expectedPass {
				t.Errorf("验证结果不符合期望，期望: %v, 实际: %v", tt.expectedPass, result)
			}
		})
	}
}

// 辅助函数
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
