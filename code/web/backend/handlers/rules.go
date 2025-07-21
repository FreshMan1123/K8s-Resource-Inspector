package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/FreshMan1123/k8s-resource-inspector/code/web/backend/models"
	"github.com/FreshMan1123/k8s-resource-inspector/code/web/backend/services"
)

// RuleHandler 规则管理处理器
type RuleHandler struct {
	ruleService *services.RuleService
}

// NewRuleHandler 创建规则处理器实例
func NewRuleHandler() *RuleHandler {
	return &RuleHandler{
		ruleService: services.NewRuleService(),
	}
}

// GetAllRules 获取所有规则
func (h *RuleHandler) GetAllRules(c *gin.Context) {
	rules, err := h.ruleService.GetAllRules()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    rules,
	})
}

// GetRulesByCategory 获取指定分类的规则
func (h *RuleHandler) GetRulesByCategory(c *gin.Context) {
	category := c.Param("category")
	
	rules, err := h.ruleService.GetRulesByCategory(category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    rules,
	})
}

// GetRuleByID 获取具体规则详情
func (h *RuleHandler) GetRuleByID(c *gin.Context) {
	category := c.Param("category")
	id := c.Param("id")
	
	rule, err := h.ruleService.GetRuleByID(category, id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    rule,
	})
}

// GetCategories 获取规则分类
func (h *RuleHandler) GetCategories(c *gin.Context) {
	categories, err := h.ruleService.GetCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    categories,
	})
}
