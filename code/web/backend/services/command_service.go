package services

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/FreshMan1123/k8s-resource-inspector/code/web/backend/models"
)

// CommandService CLI命令执行服务
type CommandService struct {
	binaryPath string
}

// NewCommandService 创建命令服务实例
func NewCommandService() *CommandService {
	return &CommandService{
		binaryPath: "./inspector.exe", // 当前二进制文件路径
	}
}

// GetCommandTemplates 获取预定义命令模板
func (cs *CommandService) GetCommandTemplates() []models.CommandTemplate {
	return []models.CommandTemplate{
		// 集群管理命令
		{
			ID:          "cluster_list",
			Name:        "列出集群",
			Description: "列出所有可用的集群上下文",
			Category:    "cluster",
			Command:     []string{"cluster", "list"},
			Parameters:  []models.CommandParam{},
		},
		{
			ID:          "cluster_info",
			Name:        "集群信息",
			Description: "显示当前集群信息",
			Category:    "cluster",
			Command:     []string{"cluster", "info"},
			Parameters:  []models.CommandParam{},
		},
		
		// 资源管理命令
		{
			ID:          "resource_get_pods",
			Name:        "获取Pods",
			Description: "获取Pod资源列表",
			Category:    "resource",
			Command:     []string{"resource", "get", "pods"},
			Parameters: []models.CommandParam{
				{
					Name:        "namespace",
					Type:        "string",
					Required:    false,
					Default:     "",
					Description: "命名空间（留空为默认）",
				},
			},
		},
		
		// 巡检命令
		{
			ID:          "inspect_node",
			Name:        "巡检节点",
			Description: "检查节点资源状态",
			Category:    "inspect",
			Command:     []string{"inspect", "node"},
			Parameters: []models.CommandParam{
				{
					Name:        "node_name",
					Type:        "string",
					Required:    false,
					Default:     "",
					Description: "节点名称（留空检查所有节点）",
				},
			},
		},
		{
			ID:          "inspect_pod",
			Name:        "巡检Pod",
			Description: "检查Pod资源状态",
			Category:    "inspect",
			Command:     []string{"inspect", "pod"},
			Parameters: []models.CommandParam{
				{
					Name:        "pod_name",
					Type:        "string",
					Required:    false,
					Default:     "",
					Description: "Pod名称（留空检查所有Pod）",
				},
				{
					Name:        "namespace",
					Type:        "string",
					Required:    false,
					Default:     "default",
					Description: "命名空间",
				},
			},
		},
		{
			ID:          "inspect_deployment",
			Name:        "巡检Deployment",
			Description: "检查Deployment配置合规性",
			Category:    "inspect",
			Command:     []string{"inspect", "deployment"},
			Parameters:  []models.CommandParam{},
		},
		{
			ID:          "inspect_service",
			Name:        "巡检Service",
			Description: "检查Service配置",
			Category:    "inspect",
			Command:     []string{"inspect", "service"},
			Parameters:  []models.CommandParam{},
		},
		{
			ID:          "inspect_security",
			Name:        "安全检查",
			Description: "执行安全配置检查",
			Category:    "inspect",
			Command:     []string{"inspect", "security"},
			Parameters: []models.CommandParam{
				{
					Name:        "security_type",
					Type:        "select",
					Required:    false,
					Default:     "all",
					Options:     []string{"pod", "cis", "image", "all"},
					Description: "安全检查类型",
				},
			},
		},
		{
			ID:          "inspect_vulnerability",
			Name:        "漏洞扫描",
			Description: "扫描容器镜像和K8s资源漏洞",
			Category:    "inspect",
			Command:     []string{"inspect", "vulnerability"},
			Parameters: []models.CommandParam{
				{
					Name:        "scan_target",
					Type:        "select",
					Required:    true,
					Default:     "cluster",
					Options:     []string{"namespace", "pod", "image", "cluster"},
					Description: "扫描目标类型",
				},
			},
		},
	}
}

// ExecuteCommand 执行CLI命令
func (cs *CommandService) ExecuteCommand(ctx context.Context, templateID string, params map[string]string) (*models.CommandResult, error) {
	// 查找命令模板
	template := cs.findTemplate(templateID)
	if template == nil {
		return nil, fmt.Errorf("命令模板不存在: %s", templateID)
	}
	
	// 构建命令参数
	command := cs.buildCommand(*template, params)
	
	// 创建命令结果
	result := &models.CommandResult{
		ID:        fmt.Sprintf("%s_%d", templateID, time.Now().Unix()),
		Command:   strings.Join(command, " "),
		Timestamp: time.Now(),
		Status:    "running",
	}
	
	// 执行命令
	startTime := time.Now()
	cmd := exec.CommandContext(ctx, cs.binaryPath, command...)
	output, err := cmd.CombinedOutput()
	result.Duration = time.Since(startTime)
	
	// 设置结果
	result.Output = string(output)
	if err != nil {
		result.Success = false
		result.Error = err.Error()
		result.Status = "failed"
	} else {
		result.Success = true
		result.Status = "completed"
	}
	
	return result, nil
}

// findTemplate 查找命令模板
func (cs *CommandService) findTemplate(templateID string) *models.CommandTemplate {
	templates := cs.GetCommandTemplates()
	for _, template := range templates {
		if template.ID == templateID {
			return &template
		}
	}
	return nil
}

// buildCommand 构建命令参数
func (cs *CommandService) buildCommand(template models.CommandTemplate, params map[string]string) []string {
	command := make([]string, len(template.Command))
	copy(command, template.Command)
	
	// 根据模板ID添加特定参数
	switch template.ID {
	case "resource_get_pods":
		if namespace := params["namespace"]; namespace != "" {
			command = append(command, "-n", namespace)
		}
		
	case "inspect_node":
		if nodeName := params["node_name"]; nodeName != "" {
			command = append(command, nodeName)
		}
		
	case "inspect_pod":
		if podName := params["pod_name"]; podName != "" {
			command = append(command, podName)
		}
		if namespace := params["namespace"]; namespace != "" {
			command = append(command, "-n", namespace)
		}
		
	case "inspect_security":
		if securityType := params["security_type"]; securityType != "" && securityType != "all" {
			command = append(command, securityType)
		}
		
	case "inspect_vulnerability":
		target := params["scan_target"]
		if target == "" {
			// 如果没有指定扫描目标，默认使用 cluster 进行全集群扫描
			target = "cluster"
		}

		switch target {
		case "cluster":
			command = append(command, "--cluster")
		case "namespace":
			if namespace := params["namespace"]; namespace != "" {
				command = append(command, "--namespace", namespace)
			} else {
				command = append(command, "--namespace", "default")
			}
		case "pod":
			if podName := params["pod_name"]; podName != "" {
				command = append(command, "--pod", podName)
			}
		case "image":
			if imageName := params["image_name"]; imageName != "" {
				command = append(command, "--image", imageName)
			}
		}
	}
	
	return command
}
