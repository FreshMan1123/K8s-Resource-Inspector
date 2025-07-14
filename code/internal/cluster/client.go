package cluster

import (
	"fmt"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// Client 表示Kubernetes集群客户端
type Client struct {
	// Clientset 是与Kubernetes API交互的客户端
	Clientset *kubernetes.Clientset
	// ConfigPath 是使用的kubeconfig文件路径
	ConfigPath string
	// ContextName 是使用的kubeconfig上下文名称
	ContextName string
}

// NewClient 创建一个新的Kubernetes客户端
func NewClient(configPath string, contextName string) (*Client, error) {
	// 如果未指定配置文件路径，则使用默认路径
	if configPath == "" {
		if home := homedir.HomeDir(); home != "" {
			configPath = filepath.Join(home, ".kube", "config")
		} else {
			return nil, fmt.Errorf("无法确定家目录，请明确指定kubeconfig路径")
		}
	}

	// 创建加载kubeconfig的配置
	loadingRules := &clientcmd.ClientConfigLoadingRules{
		ExplicitPath: configPath,
	}
	
	// 创建上下文覆盖配置
	overrides := &clientcmd.ConfigOverrides{}
	
	// 如果指定了上下文名称，则使用它
	if contextName != "" {
		overrides.CurrentContext = contextName
	}
	
	// 使用加载规则和覆盖配置创建clientConfig
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, overrides)
	
	// 构建rest.Config
	config, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("加载kubeconfig失败: %w", err)
	}

	// 创建clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("创建Kubernetes客户端失败: %w", err)
	}

	return &Client{
		Clientset: clientset,
		ConfigPath: configPath,
		ContextName: contextName,
	}, nil
}

// GetServerVersion 获取Kubernetes集群版本
func (c *Client) GetServerVersion() (string, error) {
	version, err := c.Clientset.Discovery().ServerVersion()
	if err != nil {
		return "", fmt.Errorf("获取集群版本失败: %w", err)
	}
	return version.String(), nil
}

// GetCurrentContext 获取当前使用的上下文
func GetCurrentContext(configPath string) (string, error) {
	// 如果未指定配置文件路径，则使用默认路径
	if configPath == "" {
		if home := homedir.HomeDir(); home != "" {
			configPath = filepath.Join(home, ".kube", "config")
		} else {
			return "", fmt.Errorf("无法确定家目录，请明确指定kubeconfig路径")
		}
	}

	// 加载kubeconfig
	config, err := clientcmd.LoadFromFile(configPath)
	if err != nil {
		return "", fmt.Errorf("加载kubeconfig失败: %w", err)
	}

	return config.CurrentContext, nil
}

// ListContexts 列出kubeconfig中的所有上下文
func ListContexts(configPath string) ([]string, error) {
	// 如果未指定配置文件路径，则使用默认路径
	if configPath == "" {
		if home := homedir.HomeDir(); home != "" {
			configPath = filepath.Join(home, ".kube", "config")
		} else {
			return nil, fmt.Errorf("无法确定家目录，请明确指定kubeconfig路径")
		}
	}

	// 加载kubeconfig
	config, err := clientcmd.LoadFromFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("加载kubeconfig失败: %w", err)
	}

	contexts := make([]string, 0, len(config.Contexts))
	for name := range config.Contexts {
		contexts = append(contexts, name)
	}

	return contexts, nil
}

// SwitchContext 切换当前使用的上下文
func SwitchContext(configPath string, contextName string) error {
	// 如果未指定配置文件路径，则使用默认路径
	if configPath == "" {
		if home := homedir.HomeDir(); home != "" {
			configPath = filepath.Join(home, ".kube", "config")
		} else {
			return fmt.Errorf("无法确定家目录，请明确指定kubeconfig路径")
		}
	}

	// 加载kubeconfig
	config, err := clientcmd.LoadFromFile(configPath)
	if err != nil {
		return fmt.Errorf("加载kubeconfig失败: %w", err)
	}

	// 检查上下文是否存在
	if _, exists := config.Contexts[contextName]; !exists {
		return fmt.Errorf("上下文 '%s' 不存在", contextName)
	}

	// 切换上下文
	config.CurrentContext = contextName

	// 保存修改后的kubeconfig
	err = clientcmd.WriteToFile(*config, configPath)
	if err != nil {
		return fmt.Errorf("保存kubeconfig失败: %w", err)
	}

	return nil
} 