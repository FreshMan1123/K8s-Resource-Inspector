package rules

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"
)

// Engine 规则引擎
type Engine struct {
	// 规则加载器
	loader *RuleLoader
	// 验证器映射
	validators map[string]Validator
	// 当前环境
	environment string
}

// NewEngine 创建规则引擎
func NewEngine(rulesFile string) (*Engine, error) {
	// 创建规则加载器
	loader := NewRuleLoader(rulesFile)
	
	// 加载规则
	if err := loader.LoadRules(); err != nil {
		return nil, err
	}

	// 创建引擎
	engine := &Engine{
		loader:      loader,
		validators:  make(map[string]Validator),
		environment: "prod", // 默认环境
	}

	// 注册默认验证器
	engine.registerDefaultValidators()

	return engine, nil
}

// SetEnvironment 设置当前环境
func (e *Engine) SetEnvironment(env string) {
	e.environment = env
}

// GetEnvironment 获取当前环境
func (e *Engine) GetEnvironment() string {
	return e.environment
}

// DetermineEnvironment 根据集群名称确定环境
func (e *Engine) DetermineEnvironment(clusterName string) string {
	return e.loader.GetEnvironment(clusterName)
}

// registerDefaultValidators 注册默认验证器
func (e *Engine) registerDefaultValidators() {
	// 注册数值验证器
	e.RegisterValidator("numeric", &NumericValidator{})
	// 注册字符串验证器
	e.RegisterValidator("string", &StringValidator{})
	// 注册布尔验证器
	e.RegisterValidator("boolean", &BooleanValidator{})
	e.RegisterValidator("map", &MapValidator{}) // 新增
}

// RegisterValidator 注册验证器
func (e *Engine) RegisterValidator(name string, validator Validator) {
	e.validators[name] = validator
}

// GetValidator 获取验证器
func (e *Engine) GetValidator(metricType string) (Validator, error) {
	validator, exists := e.validators[metricType]
	if !exists {
		return nil, fmt.Errorf("未知的指标类型: %s", metricType)
	}
	return validator, nil
}

// GetRules 获取规则
func (e *Engine) GetRules(filter RuleFilter) []Rule {
	return e.loader.GetRules(filter)
}

// EvaluateRule 评估单个规则
func (e *Engine) EvaluateRule(rule Rule, metricType string, actualValue interface{}) (*RuleResult, error) {
	// 检查规则是否启用
	if !rule.Enabled {
		return nil, fmt.Errorf("规则未启用: %s", rule.Name)
	}

	// 获取验证器
	validator, err := e.GetValidator(metricType)
	if err != nil {
		return nil, err
	}

	// 获取阈值
	threshold := e.getThresholdValue(rule.Condition, e.environment)
	
	// 验证值
	passed, err := validator.Validate(rule.Condition.Metric, actualValue, rule.Condition, e.environment)
	if err != nil {
		return nil, fmt.Errorf("验证失败: %w", err)
	}

	// 创建结果
	result := &RuleResult{
		RuleID:        rule.ID,
		RuleName:      rule.Name,
		Passed:        passed,
		ActualValue:   actualValue,
		ExpectedValue: threshold,
		Message:       e.formatResultMessage(rule, passed, validator.FormatValue(actualValue), validator.FormatValue(threshold)),
		Remediation:   rule.Remediation,
		Severity:      rule.Severity,
		EvaluatedAt:   time.Now(),
	}

	return result, nil
}

// getThresholdValue 获取适用于当前环境的阈值
func (e *Engine) getThresholdValue(condition RuleCondition, env string) interface{} {
	// 先尝试从环境特定阈值中获取
	if condition.Thresholds != nil {
		if threshold, exists := condition.Thresholds[env]; exists {
			return threshold
		}
		
		// 尝试获取默认环境的阈值
		if threshold, exists := condition.Thresholds["default"]; exists {
			return threshold
		}
	}
	
	// 返回通用阈值
	return condition.Threshold
}

// formatResultMessage 格式化结果消息
func (e *Engine) formatResultMessage(rule Rule, passed bool, formattedValue string, formattedThreshold string) string {
	if passed {
		return fmt.Sprintf("%s: 检查通过 (值: %s)", rule.Name, formattedValue)
	}
	
	// 根据操作符生成不同的消息
	var expectation string
	switch rule.Condition.Operator {
	case ">":
		expectation = fmt.Sprintf("应大于 %s", formattedThreshold)
	case ">=":
		expectation = fmt.Sprintf("应大于等于 %s", formattedThreshold)
	case "<":
		expectation = fmt.Sprintf("应小于 %s", formattedThreshold)
	case "<=":
		expectation = fmt.Sprintf("应小于等于 %s", formattedThreshold)
	case "==":
		expectation = fmt.Sprintf("应等于 %s", formattedThreshold)
	case "!=":
		expectation = fmt.Sprintf("不应等于 %s", formattedThreshold)
	case "contains":
		expectation = fmt.Sprintf("应包含 %s", formattedThreshold)
	case "has_non_empty":
		// 对于has_non_empty操作符，提供更具体的错误信息
		if threshold, ok := rule.Condition.Threshold.(map[string]interface{}); ok {
			var missingKeys []string
			for key := range threshold {
				missingKeys = append(missingKeys, key)
			}
			if len(missingKeys) == 1 {
				expectation = fmt.Sprintf("应包含标签 '%s' 且值不为空", missingKeys[0])
			} else {
				expectation = fmt.Sprintf("应包含标签 %v 且值不为空", missingKeys)
			}
		} else {
			expectation = "应包含指定标签且值不为空"
		}
	case "matches":
		expectation = fmt.Sprintf("应匹配正则表达式 %s", formattedThreshold)
	default:
		expectation = fmt.Sprintf("不满足条件 %s %s", rule.Condition.Operator, formattedThreshold)
	}

	return fmt.Sprintf("%s: 检查失败, 值 %s %s", rule.Name, formattedValue, expectation)
}

// NumericValidator 数值验证器
type NumericValidator struct{}

// Validate 验证数值
func (v *NumericValidator) Validate(metric string, actualValue interface{}, condition RuleCondition, env string) (bool, error) {
	// 将实际值转换为float64
	actualFloat, err := toFloat64(actualValue)
	if err != nil {
		return false, fmt.Errorf("无法将实际值转换为数字: %v", err)
	}

	// 获取适用的阈值
	var thresholdValue interface{}
	if len(condition.Thresholds) > 0 {
		// 尝试获取环境特定阈值
		if val, exists := condition.Thresholds[env]; exists {
			thresholdValue = val
		} else if val, exists := condition.Thresholds["default"]; exists {
			// 尝试获取默认环境阈值
			thresholdValue = val
		} else {
			// 使用通用阈值
			thresholdValue = condition.Threshold
		}
	} else {
		// 使用通用阈值
		thresholdValue = condition.Threshold
	}

	// 将阈值转换为float64
	thresholdFloat, err := toFloat64(thresholdValue)
	if err != nil {
		return false, fmt.Errorf("无法将阈值转换为数字: %v", err)
	}

	// 根据操作符比较
	switch condition.Operator {
	case ">":
		return actualFloat > thresholdFloat, nil
	case ">=":
		return actualFloat >= thresholdFloat, nil
	case "<":
		return actualFloat < thresholdFloat, nil
	case "<=":
		return actualFloat <= thresholdFloat, nil
	case "==":
		return actualFloat == thresholdFloat, nil
	case "!=":
		return actualFloat != thresholdFloat, nil
	default:
		return false, fmt.Errorf("数值类型不支持的操作符: %s", condition.Operator)
	}
}

// FormatValue 格式化数值
func (v *NumericValidator) FormatValue(value interface{}) string {
	if f, err := toFloat64(value); err == nil {
		if f == float64(int(f)) {
			return fmt.Sprintf("%d", int(f))
		}
		return fmt.Sprintf("%.2f", f)
	}
	return fmt.Sprintf("%v", value)
}

// StringValidator 字符串验证器
type StringValidator struct{}

// Validate 验证字符串
func (v *StringValidator) Validate(metric string, actualValue interface{}, condition RuleCondition, env string) (bool, error) {
	// 将实际值转换为字符串
	actualStr, ok := toString(actualValue)
	if !ok {
		return false, fmt.Errorf("无法将实际值转换为字符串: %v", actualValue)
	}

	// 获取适用的阈值
	var thresholdValue interface{}
	if len(condition.Thresholds) > 0 {
		// 尝试获取环境特定阈值
		if val, exists := condition.Thresholds[env]; exists {
			thresholdValue = val
		} else if val, exists := condition.Thresholds["default"]; exists {
			// 尝试获取默认环境阈值
			thresholdValue = val
		} else {
			// 使用通用阈值
			thresholdValue = condition.Threshold
		}
	} else {
		// 使用通用阈值
		thresholdValue = condition.Threshold
	}

	// 将阈值转换为字符串
	thresholdStr, ok := toString(thresholdValue)
	if !ok {
		return false, fmt.Errorf("无法将阈值转换为字符串: %v", thresholdValue)
	}

	// 根据操作符比较
	switch condition.Operator {
	case "==":
		return actualStr == thresholdStr, nil
	case "!=":
		return actualStr != thresholdStr, nil
	case "contains":
		return strings.Contains(actualStr, thresholdStr), nil
	case "matches":
		// 正则表达式匹配
		matched, err := regexp.MatchString(thresholdStr, actualStr)
		if err != nil {
			return false, fmt.Errorf("正则表达式匹配失败: %v", err)
		}
		return matched, nil
	default:
		return false, fmt.Errorf("字符串类型不支持的操作符: %s", condition.Operator)
	}
}

// FormatValue 格式化字符串
func (v *StringValidator) FormatValue(value interface{}) string {
	if str, ok := toString(value); ok {
		return str
	}
	return fmt.Sprintf("%v", value)
}

// BooleanValidator 布尔验证器
type BooleanValidator struct{}

// Validate 验证布尔值
func (v *BooleanValidator) Validate(metric string, actualValue interface{}, condition RuleCondition, env string) (bool, error) {
	// 将实际值转换为布尔值
	actualBool, ok := toBool(actualValue)
	if !ok {
		return false, fmt.Errorf("无法将实际值转换为布尔值: %v", actualValue)
	}

	// 获取适用的阈值
	var thresholdValue interface{}
	if len(condition.Thresholds) > 0 {
		// 尝试获取环境特定阈值
		if val, exists := condition.Thresholds[env]; exists {
			thresholdValue = val
		} else if val, exists := condition.Thresholds["default"]; exists {
			// 尝试获取默认环境阈值
			thresholdValue = val
		} else {
			// 使用通用阈值
			thresholdValue = condition.Threshold
		}
	} else {
		// 使用通用阈值
		thresholdValue = condition.Threshold
	}

	// 将阈值转换为布尔值
	thresholdBool, ok := toBool(thresholdValue)
	if !ok {
		return false, fmt.Errorf("无法将阈值转换为布尔值: %v", thresholdValue)
	}

	// 根据操作符比较
	switch condition.Operator {
	case "==":
		return actualBool == thresholdBool, nil
	case "!=":
		return actualBool != thresholdBool, nil
	default:
		return false, fmt.Errorf("布尔类型不支持的操作符: %s", condition.Operator)
	}
}

// FormatValue 格式化布尔值
func (v *BooleanValidator) FormatValue(value interface{}) string {
	if b, ok := toBool(value); ok {
		if b {
			return "true"
		}
		return "false"
	}
	return fmt.Sprintf("%v", value)
}

// MapValidator 用于验证map类型的标签
type MapValidator struct{}

func (v *MapValidator) Validate(metric string, actualValue interface{}, condition RuleCondition, env string) (bool, error) {
	// 将实际值转换为map[string]string
	actual, ok := actualValue.(map[string]string)
	if !ok {
		return false, fmt.Errorf("actualValue类型断言失败，期望map[string]string，实际类型：%T", actualValue)
	}

	// 获取适用的阈值（与其他验证器保持一致）
	var thresholdValue interface{}
	if len(condition.Thresholds) > 0 {
		if val, exists := condition.Thresholds[env]; exists {
			thresholdValue = val
		} else if val, exists := condition.Thresholds["default"]; exists {
			thresholdValue = val
		} else {
			thresholdValue = condition.Threshold
		}
	} else {
		thresholdValue = condition.Threshold
	}

	// 转换threshold为map[string]string
	expected, err := v.convertToStringMap(thresholdValue)
	if err != nil {
		return false, fmt.Errorf("threshold类型转换失败: %w", err)
	}

	// 根据操作符执行不同的验证逻辑
	switch condition.Operator {
	case "==":
		return v.validateEquals(actual, expected), nil
	case "contains":
		return v.validateContains(actual, expected), nil
	case "has_non_empty":
		return v.validateHasNonEmpty(actual, expected), nil
	default:
		return false, fmt.Errorf("map类型不支持的操作符: %s", condition.Operator)
	}
}

// validateEquals 验证map完全相等
func (v *MapValidator) validateEquals(actual, expected map[string]string) bool {
	for k, v := range expected {
		if actualVal, exists := actual[k]; !exists || actualVal != v {
			return false
		}
	}
	return true
}

// validateContains 验证map包含指定的键值对
func (v *MapValidator) validateContains(actual, expected map[string]string) bool {
	for k, v := range expected {
		if actualVal, exists := actual[k]; !exists || actualVal != v {
			return false
		}
	}
	return true
}

// validateHasNonEmpty 验证map包含指定的键且值不为空
func (v *MapValidator) validateHasNonEmpty(actual, expected map[string]string) bool {
	for k := range expected {
		actualVal, exists := actual[k]
		if !exists {
			return false
		}
		// 检查值是否为空（空字符串或纯空格）
		if strings.TrimSpace(actualVal) == "" {
			return false
		}
	}
	return true
}

// convertToStringMap 将不同类型的map转换为map[string]string
func (v *MapValidator) convertToStringMap(value interface{}) (map[string]string, error) {
	switch v := value.(type) {
	case map[string]string:
		return v, nil
	case map[string]interface{}:
		result := make(map[string]string)
		for k, val := range v {
			result[k] = fmt.Sprintf("%v", val)
		}
		return result, nil
	case map[interface{}]interface{}:
		result := make(map[string]string)
		for k, val := range v {
			keyStr := fmt.Sprintf("%v", k)
			valStr := fmt.Sprintf("%v", val)
			result[keyStr] = valStr
		}
		return result, nil
	default:
		return nil, fmt.Errorf("不支持的threshold类型: %T", value)
	}
}

func (v *MapValidator) FormatValue(value interface{}) string {
	return fmt.Sprintf("%v", value)
}

// toFloat64 将值转换为float64
func toFloat64(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case string:
		// 尝试将字符串解析为浮点数
		var f float64
		if _, err := fmt.Sscanf(v, "%f", &f); err == nil {
			return f, nil
		}
		return 0, fmt.Errorf("无法将字符串转换为浮点数: %s", v)
	}
	
	// 尝试使用反射
	val := reflect.ValueOf(value)
	switch val.Kind() {
	case reflect.Float32, reflect.Float64:
		return val.Float(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(val.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(val.Uint()), nil
	}

	return 0, fmt.Errorf("不支持的类型: %T", value)
}

// toString 将值转换为字符串
func toString(value interface{}) (string, bool) {
	if value == nil {
		return "", false
	}
	
	switch v := value.(type) {
	case string:
		return v, true
	case []byte:
		return string(v), true
	default:
		return fmt.Sprintf("%v", value), true
	}
}

// toBool 将值转换为布尔值
func toBool(value interface{}) (bool, bool) {
	if value == nil {
		return false, false
	}
	
	switch v := value.(type) {
	case bool:
		return v, true
	case int:
		return v != 0, true
	case string:
		lower := strings.ToLower(v)
		if lower == "true" || lower == "yes" || lower == "1" {
			return true, true
		}
		if lower == "false" || lower == "no" || lower == "0" {
			return false, true
		}
		return false, false
	default:
		return false, false
	}
} 