package services

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
	"github.com/FreshMan1123/k8s-resource-inspector/code/web/backend/models"
)

// RuleService 规则管理服务
type RuleService struct {
	configPath string
}

// NewRuleService 创建规则服务实例
func NewRuleService() *RuleService {
	// 尝试多个可能的路径
	possiblePaths := []string{
		"code/configs/rules",        // 相对于项目根目录
		"configs/rules",             // 相对于code目录
		"../configs/rules",          // 相对于web目录
		"../../configs/rules",       // 相对于web/backend目录
	}

	// 选择第一个存在的路径
	configPath := ""
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			configPath = path
			log.Printf("找到规则配置目录: %s", path)
			break
		}
	}

	// 如果没有找到，使用默认路径
	if configPath == "" {
		configPath = "code/configs/rules"
		log.Printf("未找到规则配置目录，使用默认路径: %s", configPath)
	}

	return &RuleService{
		configPath: configPath,
	}
}

// GetAllRules 获取所有规则，按分类组织
func (rs *RuleService) GetAllRules() (map[string][]models.Rule, error) {
	categories := []string{
		"node", 
		"pod", 
		"deployment", 
		"service", 
		"cis-kubernetes-v1.8", 
		"image-security", 
		"pod-security",
	}
	
	allRules := make(map[string][]models.Rule)
	
	for _, category := range categories {
		rules, err := rs.GetRulesByCategory(category)
		if err != nil {
			// 跳过不存在的文件，继续处理其他分类
			continue
		}
		allRules[category] = rules
	}
	
	return allRules, nil
}

// GetRulesByCategory 根据分类获取规则
func (rs *RuleService) GetRulesByCategory(category string) ([]models.Rule, error) {
	filename := fmt.Sprintf("%s.yaml", category)
	filePath := filepath.Join(rs.configPath, filename)

	log.Printf("尝试加载规则文件: %s", filePath)

	rulesConfig, err := rs.loadRulesFromFile(filePath)
	if err != nil {
		log.Printf("加载规则文件失败: %v", err)
		return nil, fmt.Errorf("加载规则文件失败: %w", err)
	}

	log.Printf("成功加载 %d 条规则从文件: %s", len(rulesConfig.Rules), filePath)

	// 为每个规则设置分类信息
	for i := range rulesConfig.Rules {
		rulesConfig.Rules[i].Category = category
	}

	return rulesConfig.Rules, nil
}

// GetRuleByID 根据ID获取具体规则
func (rs *RuleService) GetRuleByID(category, id string) (*models.Rule, error) {
	rules, err := rs.GetRulesByCategory(category)
	if err != nil {
		return nil, err
	}
	
	for _, rule := range rules {
		if rule.ID == id {
			return &rule, nil
		}
	}
	
	return nil, fmt.Errorf("规则不存在: %s", id)
}

// GetCategories 获取所有规则分类
func (rs *RuleService) GetCategories() ([]models.RuleCategory, error) {
	allRules, err := rs.GetAllRules()
	if err != nil {
		return nil, err
	}
	
	categories := []models.RuleCategory{
		{Key: "node", Name: "节点规则", Description: "节点资源和状态检查规则"},
		{Key: "pod", Name: "Pod规则", Description: "Pod配置和运行状态检查规则"},
		{Key: "deployment", Name: "部署规则", Description: "Deployment配置检查规则"},
		{Key: "service", Name: "服务规则", Description: "Service配置检查规则"},
		{Key: "cis-kubernetes-v1.8", Name: "CIS基准", Description: "CIS Kubernetes基准检查规则"},
		{Key: "image-security", Name: "镜像安全", Description: "容器镜像安全检查规则"},
		{Key: "pod-security", Name: "Pod安全", Description: "Pod安全配置检查规则"},
	}
	
	// 统计每个分类的规则数量
	for i, category := range categories {
		if rules, exists := allRules[category.Key]; exists {
			categories[i].Count = len(rules)
		}
	}
	
	return categories, nil
}

// loadRulesFromFile 从文件加载规则配置
func (rs *RuleService) loadRulesFromFile(filePath string) (*models.RulesConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}

	// 先解析为通用的map结构，以适应现有的YAML格式
	var rawConfig map[string]interface{}
	err = yaml.Unmarshal(data, &rawConfig)
	if err != nil {
		return nil, fmt.Errorf("解析YAML失败: %w", err)
	}

	// 创建RulesConfig结构
	rulesConfig := &models.RulesConfig{}

	// 提取基本信息
	if apiVersion, ok := rawConfig["apiVersion"].(string); ok {
		rulesConfig.APIVersion = apiVersion
	}
	if kind, ok := rawConfig["kind"].(string); ok {
		rulesConfig.Kind = kind
	}

	// 提取规则列表
	if rulesData, ok := rawConfig["rules"].([]interface{}); ok {
		for _, ruleData := range rulesData {
			if ruleMap, ok := ruleData.(map[interface{}]interface{}); ok {
				rule := rs.parseRule(ruleMap)
				rulesConfig.Rules = append(rulesConfig.Rules, rule)
			}
		}
	}

	return rulesConfig, nil
}

// parseRule 解析单个规则
func (rs *RuleService) parseRule(ruleMap map[interface{}]interface{}) models.Rule {
	rule := models.Rule{}

	// 解析基本字段
	if id, ok := ruleMap["id"].(string); ok {
		rule.ID = id
	}
	if name, ok := ruleMap["name"].(string); ok {
		rule.Name = name
	}
	if description, ok := ruleMap["description"].(string); ok {
		rule.Description = description
	}
	if category, ok := ruleMap["category"].(string); ok {
		rule.Category = category
	}
	if severity, ok := ruleMap["severity"].(string); ok {
		rule.Severity = severity
	}
	if remediation, ok := ruleMap["remediation"].(string); ok {
		rule.Remediation = remediation
	}
	if enabled, ok := ruleMap["enabled"].(bool); ok {
		rule.Enabled = enabled
	}

	// 解析条件
	if conditionData, ok := ruleMap["condition"].(map[interface{}]interface{}); ok {
		condition := models.RuleCondition{}
		if metric, ok := conditionData["metric"].(string); ok {
			condition.Metric = metric
		}
		if operator, ok := conditionData["operator"].(string); ok {
			condition.Operator = operator
		}
		if threshold := conditionData["threshold"]; threshold != nil {
			// 确保threshold是JSON可序列化的类型
			switch v := threshold.(type) {
			case int:
				condition.Threshold = v
			case float64:
				condition.Threshold = v
			case string:
				condition.Threshold = v
			case bool:
				condition.Threshold = v
			default:
				// 对于其他类型，转换为字符串
				condition.Threshold = fmt.Sprintf("%v", v)
			}
		}
		if unit, ok := conditionData["unit"].(string); ok {
			condition.Unit = unit
		}
		rule.Condition = condition
	}

	return rule
}
