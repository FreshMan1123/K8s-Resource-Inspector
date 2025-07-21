package services

import (
	"fmt"
	"io/ioutil"
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
	return &RuleService{
		configPath: "code/configs/rules",
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
	
	rulesConfig, err := rs.loadRulesFromFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("加载规则文件失败: %w", err)
	}
	
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
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}
	
	var rulesConfig models.RulesConfig
	err = yaml.Unmarshal(data, &rulesConfig)
	if err != nil {
		return nil, fmt.Errorf("解析YAML失败: %w", err)
	}
	
	return &rulesConfig, nil
}
