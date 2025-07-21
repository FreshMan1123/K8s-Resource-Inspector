package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/FreshMan1123/k8s-resource-inspector/code/web/backend/models"
	"github.com/FreshMan1123/k8s-resource-inspector/code/web/backend/services"
)

// InspectionHandler 巡检处理器
type InspectionHandler struct {
	commandService *services.CommandService
}

// NewInspectionHandler 创建巡检处理器实例
func NewInspectionHandler() *InspectionHandler {
	return &InspectionHandler{
		commandService: services.NewCommandService(),
	}
}

// GetCommandTemplates 获取命令模板
func (h *InspectionHandler) GetCommandTemplates(c *gin.Context) {
	templates := h.commandService.GetCommandTemplates()
	
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    templates,
	})
}

// ExecuteCommand 执行命令
func (h *InspectionHandler) ExecuteCommand(c *gin.Context) {
	var req models.ExecutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "无效的请求参数: " + err.Error(),
		})
		return
	}

	// 创建超时上下文
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	// 执行命令
	result, err := h.commandService.ExecuteCommand(ctx, req.TemplateID, req.Parameters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    result,
	})
}
