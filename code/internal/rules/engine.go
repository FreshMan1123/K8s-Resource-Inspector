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

	// 注册安全相关验证器
	e.RegisterValidator("security_context", &SecurityContextValidator{})
	e.RegisterValidator("image_security", &ImageSecurityValidator{})
	e.RegisterValidator("network_policy", &NetworkPolicyValidator{})
	e.RegisterValidator("cis_compliance", &CISComplianceValidator{})
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

// SecurityContextValidator 安全上下文验证器
type SecurityContextValidator struct{}

// Validate 验证安全上下文配置
func (v *SecurityContextValidator) Validate(metric string, actualValue interface{}, condition RuleCondition, env string) (bool, error) {
	// 将实际值转换为map[string]interface{}
	securityContext, ok := actualValue.(map[string]interface{})
	if !ok {
		return false, fmt.Errorf("安全上下文类型断言失败，期望map[string]interface{}，实际类型：%T", actualValue)
	}

	// 获取适用的阈值
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

	// 根据指标类型进行不同的验证
	switch metric {
	case "privileged":
		return v.validatePrivileged(securityContext, thresholdValue, condition.Operator)
	case "allow_privilege_escalation":
		return v.validatePrivilegeEscalation(securityContext, thresholdValue, condition.Operator)
	case "run_as_non_root":
		return v.validateRunAsNonRoot(securityContext, thresholdValue, condition.Operator)
	case "host_network":
		return v.validateHostNetwork(securityContext, thresholdValue, condition.Operator)
	case "host_pid_ipc":
		return v.validateHostPidIpc(securityContext, thresholdValue, condition.Operator)
	case "read_only_root_filesystem":
		return v.validateReadOnlyRootFilesystem(securityContext, thresholdValue, condition.Operator)
	default:
		return false, fmt.Errorf("不支持的安全上下文指标: %s", metric)
	}
}

// validatePrivileged 验证特权容器设置
func (v *SecurityContextValidator) validatePrivileged(securityContext map[string]interface{}, threshold interface{}, operator string) (bool, error) {
	privileged, exists := securityContext["privileged"]
	if !exists {
		privileged = false // 默认为false
	}

	privilegedBool, ok := toBool(privileged)
	if !ok {
		return false, fmt.Errorf("privileged值类型转换失败: %v", privileged)
	}

	thresholdBool, ok := toBool(threshold)
	if !ok {
		return false, fmt.Errorf("privileged阈值类型转换失败: %v", threshold)
	}

	switch operator {
	case "==":
		return privilegedBool == thresholdBool, nil
	case "!=":
		return privilegedBool != thresholdBool, nil
	default:
		return false, fmt.Errorf("privileged不支持的操作符: %s", operator)
	}
}

// validatePrivilegeEscalation 验证特权升级设置
func (v *SecurityContextValidator) validatePrivilegeEscalation(securityContext map[string]interface{}, threshold interface{}, operator string) (bool, error) {
	allowPrivilegeEscalation, exists := securityContext["allowPrivilegeEscalation"]
	if !exists {
		allowPrivilegeEscalation = true // Kubernetes默认为true
	}

	escalationBool, ok := toBool(allowPrivilegeEscalation)
	if !ok {
		return false, fmt.Errorf("allowPrivilegeEscalation值类型转换失败: %v", allowPrivilegeEscalation)
	}

	thresholdBool, ok := toBool(threshold)
	if !ok {
		return false, fmt.Errorf("allowPrivilegeEscalation阈值类型转换失败: %v", threshold)
	}

	switch operator {
	case "==":
		return escalationBool == thresholdBool, nil
	case "!=":
		return escalationBool != thresholdBool, nil
	default:
		return false, fmt.Errorf("allowPrivilegeEscalation不支持的操作符: %s", operator)
	}
}

// validateRunAsNonRoot 验证非root用户运行设置
func (v *SecurityContextValidator) validateRunAsNonRoot(securityContext map[string]interface{}, threshold interface{}, operator string) (bool, error) {
	runAsNonRoot, exists := securityContext["runAsNonRoot"]
	if !exists {
		runAsNonRoot = false // 默认为false
	}

	nonRootBool, ok := toBool(runAsNonRoot)
	if !ok {
		return false, fmt.Errorf("runAsNonRoot值类型转换失败: %v", runAsNonRoot)
	}

	thresholdBool, ok := toBool(threshold)
	if !ok {
		return false, fmt.Errorf("runAsNonRoot阈值类型转换失败: %v", threshold)
	}

	switch operator {
	case "==":
		return nonRootBool == thresholdBool, nil
	case "!=":
		return nonRootBool != thresholdBool, nil
	default:
		return false, fmt.Errorf("runAsNonRoot不支持的操作符: %s", operator)
	}
}

// validateHostNetwork 验证主机网络设置
func (v *SecurityContextValidator) validateHostNetwork(securityContext map[string]interface{}, threshold interface{}, operator string) (bool, error) {
	hostNetwork, exists := securityContext["hostNetwork"]
	if !exists {
		hostNetwork = false // 默认为false
	}

	hostNetworkBool, ok := toBool(hostNetwork)
	if !ok {
		return false, fmt.Errorf("hostNetwork值类型转换失败: %v", hostNetwork)
	}

	thresholdBool, ok := toBool(threshold)
	if !ok {
		return false, fmt.Errorf("hostNetwork阈值类型转换失败: %v", threshold)
	}

	switch operator {
	case "==":
		return hostNetworkBool == thresholdBool, nil
	case "!=":
		return hostNetworkBool != thresholdBool, nil
	default:
		return false, fmt.Errorf("hostNetwork不支持的操作符: %s", operator)
	}
}

// validateHostPidIpc 验证主机PID/IPC设置
func (v *SecurityContextValidator) validateHostPidIpc(securityContext map[string]interface{}, threshold interface{}, operator string) (bool, error) {
	hostPID, pidExists := securityContext["hostPID"]
	hostIPC, ipcExists := securityContext["hostIPC"]

	if !pidExists {
		hostPID = false
	}
	if !ipcExists {
		hostIPC = false
	}

	hostPIDBool, ok := toBool(hostPID)
	if !ok {
		return false, fmt.Errorf("hostPID值类型转换失败: %v", hostPID)
	}

	hostIPCBool, ok := toBool(hostIPC)
	if !ok {
		return false, fmt.Errorf("hostIPC值类型转换失败: %v", hostIPC)
	}

	thresholdBool, ok := toBool(threshold)
	if !ok {
		return false, fmt.Errorf("hostPID/IPC阈值类型转换失败: %v", threshold)
	}

	// 检查是否有任何一个为true
	hasHostPidOrIpc := hostPIDBool || hostIPCBool

	switch operator {
	case "==":
		return hasHostPidOrIpc == thresholdBool, nil
	case "!=":
		return hasHostPidOrIpc != thresholdBool, nil
	default:
		return false, fmt.Errorf("hostPID/IPC不支持的操作符: %s", operator)
	}
}

// validateReadOnlyRootFilesystem 验证只读根文件系统设置
func (v *SecurityContextValidator) validateReadOnlyRootFilesystem(securityContext map[string]interface{}, threshold interface{}, operator string) (bool, error) {
	readOnlyRootFilesystem, exists := securityContext["readOnlyRootFilesystem"]
	if !exists {
		readOnlyRootFilesystem = false // 默认为false
	}

	readOnlyBool, ok := toBool(readOnlyRootFilesystem)
	if !ok {
		return false, fmt.Errorf("readOnlyRootFilesystem值类型转换失败: %v", readOnlyRootFilesystem)
	}

	thresholdBool, ok := toBool(threshold)
	if !ok {
		return false, fmt.Errorf("readOnlyRootFilesystem阈值类型转换失败: %v", threshold)
	}

	switch operator {
	case "==":
		return readOnlyBool == thresholdBool, nil
	case "!=":
		return readOnlyBool != thresholdBool, nil
	default:
		return false, fmt.Errorf("readOnlyRootFilesystem不支持的操作符: %s", operator)
	}
}

// FormatValue 格式化安全上下文值
func (v *SecurityContextValidator) FormatValue(value interface{}) string {
	return fmt.Sprintf("%v", value)
}

// ImageSecurityValidator 镜像安全验证器
type ImageSecurityValidator struct{}

// Validate 验证镜像安全配置
func (v *ImageSecurityValidator) Validate(metric string, actualValue interface{}, condition RuleCondition, env string) (bool, error) {
	// 获取适用的阈值
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

	// 根据指标类型进行不同的验证
	switch metric {
	case "image_tag":
		return v.validateImageTag(actualValue, thresholdValue, condition.Operator)
	case "image_registry":
		return v.validateImageRegistry(actualValue, thresholdValue, condition.Operator)
	case "image_pull_policy":
		return v.validateImagePullPolicy(actualValue, thresholdValue, condition.Operator)
	case "image_digest":
		return v.validateImageDigest(actualValue, thresholdValue, condition.Operator)
	default:
		return false, fmt.Errorf("不支持的镜像安全指标: %s", metric)
	}
}

// validateImageTag 验证镜像标签
func (v *ImageSecurityValidator) validateImageTag(actualValue, threshold interface{}, operator string) (bool, error) {
	actualStr, ok := toString(actualValue)
	if !ok {
		return false, fmt.Errorf("镜像标签类型转换失败: %v", actualValue)
	}

	thresholdStr, ok := toString(threshold)
	if !ok {
		return false, fmt.Errorf("镜像标签阈值类型转换失败: %v", threshold)
	}

	switch operator {
	case "==":
		return actualStr == thresholdStr, nil
	case "!=":
		return actualStr != thresholdStr, nil
	case "contains":
		return strings.Contains(actualStr, thresholdStr), nil
	case "matches":
		matched, err := regexp.MatchString(thresholdStr, actualStr)
		if err != nil {
			return false, fmt.Errorf("镜像标签正则匹配失败: %v", err)
		}
		return matched, nil
	default:
		return false, fmt.Errorf("镜像标签不支持的操作符: %s", operator)
	}
}

// validateImageRegistry 验证镜像仓库
func (v *ImageSecurityValidator) validateImageRegistry(actualValue, threshold interface{}, operator string) (bool, error) {
	actualStr, ok := toString(actualValue)
	if !ok {
		return false, fmt.Errorf("镜像仓库类型转换失败: %v", actualValue)
	}

	switch operator {
	case "==":
		thresholdStr, ok := toString(threshold)
		if !ok {
			return false, fmt.Errorf("镜像仓库阈值类型转换失败: %v", threshold)
		}
		return actualStr == thresholdStr, nil
	case "!=":
		thresholdStr, ok := toString(threshold)
		if !ok {
			return false, fmt.Errorf("镜像仓库阈值类型转换失败: %v", threshold)
		}
		return actualStr != thresholdStr, nil
	case "contains":
		thresholdStr, ok := toString(threshold)
		if !ok {
			return false, fmt.Errorf("镜像仓库阈值类型转换失败: %v", threshold)
		}
		return strings.Contains(actualStr, thresholdStr), nil
	case "matches":
		thresholdStr, ok := toString(threshold)
		if !ok {
			return false, fmt.Errorf("镜像仓库阈值类型转换失败: %v", threshold)
		}
		matched, err := regexp.MatchString(thresholdStr, actualStr)
		if err != nil {
			return false, fmt.Errorf("镜像仓库正则匹配失败: %v", err)
		}
		return matched, nil
	case "in":
		// 支持检查是否在允许的仓库列表中
		thresholdSlice, ok := threshold.([]interface{})
		if !ok {
			return false, fmt.Errorf("镜像仓库阈值应为数组类型: %v", threshold)
		}
		for _, item := range thresholdSlice {
			itemStr, ok := toString(item)
			if !ok {
				continue
			}
			if strings.Contains(actualStr, itemStr) {
				return true, nil
			}
		}
		return false, nil
	default:
		return false, fmt.Errorf("镜像仓库不支持的操作符: %s", operator)
	}
}

// validateImagePullPolicy 验证镜像拉取策略
func (v *ImageSecurityValidator) validateImagePullPolicy(actualValue, threshold interface{}, operator string) (bool, error) {
	actualStr, ok := toString(actualValue)
	if !ok {
		return false, fmt.Errorf("镜像拉取策略类型转换失败: %v", actualValue)
	}

	thresholdStr, ok := toString(threshold)
	if !ok {
		return false, fmt.Errorf("镜像拉取策略阈值类型转换失败: %v", threshold)
	}

	switch operator {
	case "==":
		return actualStr == thresholdStr, nil
	case "!=":
		return actualStr != thresholdStr, nil
	case "matches":
		matched, err := regexp.MatchString(thresholdStr, actualStr)
		if err != nil {
			return false, fmt.Errorf("镜像拉取策略正则匹配失败: %v", err)
		}
		return matched, nil
	default:
		return false, fmt.Errorf("镜像拉取策略不支持的操作符: %s", operator)
	}
}

// validateImageDigest 验证镜像摘要
func (v *ImageSecurityValidator) validateImageDigest(actualValue, threshold interface{}, operator string) (bool, error) {
	actualStr, ok := toString(actualValue)
	if !ok {
		return false, fmt.Errorf("镜像摘要类型转换失败: %v", actualValue)
	}

	switch operator {
	case "exists":
		// 检查是否存在摘要（非空）
		return strings.TrimSpace(actualStr) != "", nil
	case "not_exists":
		// 检查是否不存在摘要（为空）
		return strings.TrimSpace(actualStr) == "", nil
	case "==":
		thresholdStr, ok := toString(threshold)
		if !ok {
			return false, fmt.Errorf("镜像摘要阈值类型转换失败: %v", threshold)
		}
		return actualStr == thresholdStr, nil
	case "!=":
		thresholdStr, ok := toString(threshold)
		if !ok {
			return false, fmt.Errorf("镜像摘要阈值类型转换失败: %v", threshold)
		}
		return actualStr != thresholdStr, nil
	default:
		return false, fmt.Errorf("镜像摘要不支持的操作符: %s", operator)
	}
}

// FormatValue 格式化镜像安全值
func (v *ImageSecurityValidator) FormatValue(value interface{}) string {
	if str, ok := toString(value); ok {
		return str
	}
	return fmt.Sprintf("%v", value)
}

// NetworkPolicyValidator 网络策略验证器
type NetworkPolicyValidator struct{}

// Validate 验证网络策略配置
func (v *NetworkPolicyValidator) Validate(metric string, actualValue interface{}, condition RuleCondition, env string) (bool, error) {
	// 获取适用的阈值
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

	// 根据指标类型进行不同的验证
	switch metric {
	case "has_network_policy":
		return v.validateHasNetworkPolicy(actualValue, thresholdValue, condition.Operator)
	case "default_deny":
		return v.validateDefaultDeny(actualValue, thresholdValue, condition.Operator)
	default:
		return false, fmt.Errorf("不支持的网络策略指标: %s", metric)
	}
}

// validateHasNetworkPolicy 验证是否存在网络策略
func (v *NetworkPolicyValidator) validateHasNetworkPolicy(actualValue, threshold interface{}, operator string) (bool, error) {
	actualBool, ok := toBool(actualValue)
	if !ok {
		return false, fmt.Errorf("网络策略存在性类型转换失败: %v", actualValue)
	}

	thresholdBool, ok := toBool(threshold)
	if !ok {
		return false, fmt.Errorf("网络策略存在性阈值类型转换失败: %v", threshold)
	}

	switch operator {
	case "==":
		return actualBool == thresholdBool, nil
	case "!=":
		return actualBool != thresholdBool, nil
	default:
		return false, fmt.Errorf("网络策略存在性不支持的操作符: %s", operator)
	}
}

// validateDefaultDeny 验证默认拒绝策略
func (v *NetworkPolicyValidator) validateDefaultDeny(actualValue, threshold interface{}, operator string) (bool, error) {
	actualBool, ok := toBool(actualValue)
	if !ok {
		return false, fmt.Errorf("默认拒绝策略类型转换失败: %v", actualValue)
	}

	thresholdBool, ok := toBool(threshold)
	if !ok {
		return false, fmt.Errorf("默认拒绝策略阈值类型转换失败: %v", threshold)
	}

	switch operator {
	case "==":
		return actualBool == thresholdBool, nil
	case "!=":
		return actualBool != thresholdBool, nil
	default:
		return false, fmt.Errorf("默认拒绝策略不支持的操作符: %s", operator)
	}
}

// FormatValue 格式化网络策略值
func (v *NetworkPolicyValidator) FormatValue(value interface{}) string {
	if b, ok := toBool(value); ok {
		if b {
			return "true"
		}
		return "false"
	}
	return fmt.Sprintf("%v", value)
}

// CISComplianceValidator CIS合规性验证器
type CISComplianceValidator struct{}

// Validate 验证CIS合规性
func (v *CISComplianceValidator) Validate(metric string, actualValue interface{}, condition RuleCondition, env string) (bool, error) {
	// 获取适用的阈值
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

	// 根据指标类型进行不同的验证
	switch metric {
	case "service_account_name":
		return v.validateServiceAccountName(actualValue, thresholdValue, condition.Operator)
	case "rbac_wildcard_usage":
		return v.validateRBACWildcardUsage(actualValue, thresholdValue, condition.Operator)
	case "automount_service_account_token":
		return v.validateAutomountServiceAccountToken(actualValue, thresholdValue, condition.Operator)
	default:
		return false, fmt.Errorf("不支持的CIS合规性指标: %s", metric)
	}
}

// validateServiceAccountName 验证服务账户名称
func (v *CISComplianceValidator) validateServiceAccountName(actualValue, threshold interface{}, operator string) (bool, error) {
	actualStr, ok := toString(actualValue)
	if !ok {
		return false, fmt.Errorf("服务账户名称类型转换失败: %v", actualValue)
	}

	thresholdStr, ok := toString(threshold)
	if !ok {
		return false, fmt.Errorf("服务账户名称阈值类型转换失败: %v", threshold)
	}

	switch operator {
	case "==":
		return actualStr == thresholdStr, nil
	case "!=":
		return actualStr != thresholdStr, nil
	case "contains":
		return strings.Contains(actualStr, thresholdStr), nil
	default:
		return false, fmt.Errorf("服务账户名称不支持的操作符: %s", operator)
	}
}

// validateRBACWildcardUsage 验证RBAC通配符使用
func (v *CISComplianceValidator) validateRBACWildcardUsage(actualValue, threshold interface{}, operator string) (bool, error) {
	actualBool, ok := toBool(actualValue)
	if !ok {
		return false, fmt.Errorf("RBAC通配符使用类型转换失败: %v", actualValue)
	}

	thresholdBool, ok := toBool(threshold)
	if !ok {
		return false, fmt.Errorf("RBAC通配符使用阈值类型转换失败: %v", threshold)
	}

	switch operator {
	case "==":
		return actualBool == thresholdBool, nil
	case "!=":
		return actualBool != thresholdBool, nil
	default:
		return false, fmt.Errorf("RBAC通配符使用不支持的操作符: %s", operator)
	}
}

// validateAutomountServiceAccountToken 验证自动挂载服务账户令牌
func (v *CISComplianceValidator) validateAutomountServiceAccountToken(actualValue, threshold interface{}, operator string) (bool, error) {
	actualBool, ok := toBool(actualValue)
	if !ok {
		return false, fmt.Errorf("自动挂载服务账户令牌类型转换失败: %v", actualValue)
	}

	thresholdBool, ok := toBool(threshold)
	if !ok {
		return false, fmt.Errorf("自动挂载服务账户令牌阈值类型转换失败: %v", threshold)
	}

	switch operator {
	case "==":
		return actualBool == thresholdBool, nil
	case "!=":
		return actualBool != thresholdBool, nil
	default:
		return false, fmt.Errorf("自动挂载服务账户令牌不支持的操作符: %s", operator)
	}
}

// FormatValue 格式化CIS合规性值
func (v *CISComplianceValidator) FormatValue(value interface{}) string {
	if str, ok := toString(value); ok {
		return str
	}
	if b, ok := toBool(value); ok {
		if b {
			return "true"
		}
		return "false"
	}
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