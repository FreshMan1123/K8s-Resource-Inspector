package kubeconfig

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	// DefaultPermissions 设置为0600，确保只有文件所有者可以读写
	DefaultPermissions os.FileMode = 0600
)

// Manager 处理kubeconfig文件的安全存储和加载
type Manager struct {
	// ConfigDir 是存储kubeconfig文件的目录
	ConfigDir string
}

// NewManager 创建一个新的kubeconfig管理器，返回值返回一个 Manager 指针结构体，这样不会因为 go的副本机制而导致错误。
func NewManager(configDir string) (*Manager, error) {
	// 确保配置目录存在， MkdirAll类似于mkdir，如果目录不存在则创建，存在则正常运行。也就是当 目录没被正常创建
	// 同时又不存在时，它会不返回一个nil，而是返回一个error，那么我们就会结束此函数
	// DefaultPermissions是我们在前面配置的权限0600，只有文件所有者可以读写。
	if err := os.MkdirAll(configDir, DefaultPermissions); err != nil {
		return nil, fmt.Errorf("创建配置目录失败: %w", err)
	}
	
	return &Manager{
		//将 configDir赋给 Manger结构体的ConfigDir字段
		ConfigDir: configDir,
	}, nil
}

// SaveKubeconfig 安全地保存kubeconfig内容到指定的文件，指针接收者，指向Manager结构体，使用指针是直接对传入的 manager进行修改，而不是创建副本
func (m *Manager) SaveKubeconfig(name string, content []byte) error {
	// 构建完整的文件路径，name也就是我们的集群名字
	filePath := filepath.Join(m.ConfigDir, fmt.Sprintf("%s.yaml", name))
	
	// 使用安全权限写入文件，只有文件写入者有读写权限
	if err := ioutil.WriteFile(filePath, content, DefaultPermissions); err != nil {
		return fmt.Errorf("保存kubeconfig文件失败: %w", err)
	}
	
	return nil
}

// LoadKubeconfig 从指定的文件加载kubeconfig内容
func (m *Manager) LoadKubeconfig(name string) ([]byte, error) {
	// 构建完整的文件路径
	filePath := filepath.Join(m.ConfigDir, fmt.Sprintf("%s.yaml", name))
	
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("kubeconfig文件不存在: %s", filePath)
	}
	
	// 读取文件内容
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取kubeconfig文件失败: %w", err)
	}
	
	return content, nil
}

// ListKubeconfigs 列出所有保存的kubeconfig文件
func (m *Manager) ListKubeconfigs() ([]string, error) {
	var configs []string
	
	// 读取目录中的所有文件
	files, err := ioutil.ReadDir(m.ConfigDir)
	if err != nil {
		return nil, fmt.Errorf("读取配置目录失败: %w", err)
	}
	
	// 过滤出.yaml文件并提取名称
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".yaml" {
			// 去掉.yaml扩展名
			name := file.Name()[:len(file.Name())-5]
			configs = append(configs, name)
		}
	}
	
	return configs, nil
}

// DeleteKubeconfig 删除指定的kubeconfig文件
func (m *Manager) DeleteKubeconfig(name string) error {
	// 构建完整的文件路径
	filePath := filepath.Join(m.ConfigDir, fmt.Sprintf("%s.yaml", name))
	
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("kubeconfig文件不存在: %s", filePath)
	}
	
	// 删除文件
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("删除kubeconfig文件失败: %w", err)
	}
	
	return nil
}