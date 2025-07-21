package models

import "time"

// CommandTemplate 命令模板
type CommandTemplate struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Category    string           `json:"category"`
	Command     []string         `json:"command"`
	Parameters  []CommandParam   `json:"parameters"`
}

// CommandParam 命令参数
type CommandParam struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`        // "string", "select", "boolean"
	Required    bool     `json:"required"`
	Default     string   `json:"default"`
	Options     []string `json:"options,omitempty"` // for select type
	Description string   `json:"description"`
}

// CommandResult 命令执行结果
type CommandResult struct {
	ID        string        `json:"id"`
	Command   string        `json:"command"`
	Output    string        `json:"output"`
	Success   bool          `json:"success"`
	Error     string        `json:"error,omitempty"`
	Timestamp time.Time     `json:"timestamp"`
	Duration  time.Duration `json:"duration"`
	Status    string        `json:"status"` // "running", "completed", "failed"
}

// ExecutionRequest 执行请求
type ExecutionRequest struct {
	TemplateID string            `json:"template_id"`
	Parameters map[string]string `json:"parameters"`
}
