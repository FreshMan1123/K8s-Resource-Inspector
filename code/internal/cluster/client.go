package cluster

import (
	"fmt"
	"path/filepath"
	"context"
	"strings"


	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/api/core/v1"

	"k8s.io/metrics/pkg/client/clientset/versioned"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	"k8s.io/client-go/rest"
)

// Client 表示Kubernetes集群客户端
type Client struct {
	// Clientset 是与Kubernetes API交互的客户端
	Clientset *kubernetes.Clientset
	// ConfigPath 是使用的kubeconfig文件路径
	ConfigPath string
	// ContextName 是使用的kubeconfig上下文名称
	ContextName string
	// MetricsClient 是获取指标数据的客户端
	MetricsClient *versioned.Clientset
	Config *rest.Config // 新增字段
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

	// 创建metrics clientset
	metricsClient, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("创建Metrics客户端失败: %w", err)
	}

	return &Client{
		Clientset: clientset,
		ConfigPath: configPath,
		ContextName: contextName,
		MetricsClient: metricsClient,
		Config: config, // 赋值
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



// ======= 删除业务聚合相关接口和函数 =======
// 删除 GetPod、ListPods、GetPodLogs（返回 models.Pod/PodList 的版本）
// 删除 buildPodModel、buildContainers、getPodEvents
// ======= 保留原生对象 API 封装接口 =======
// GetRawPod, ListRawPods, GetRawPodMetrics, ListRawPodMetrics, GetRawPodEvents, GetRawPodLogs 保持不变

// 获取所有 Node 原生对象
func (c *Client) ListRawNodes(ctx context.Context) ([]v1.Node, error) {
	nodes, err := c.Clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return nodes.Items, nil
}

// 获取单个 Pod 原生对象
func (c *Client) GetRawPod(ctx context.Context, namespace, name string) (*v1.Pod, error) {
	return c.Clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
}

// 获取所有 Pod 原生对象
func (c *Client) ListRawPods(ctx context.Context, namespace string) ([]v1.Pod, error) {
	pods, err := c.Clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return pods.Items, nil
}

// 获取所有 Node 原生 metrics
func (c *Client) ListRawNodeMetrics(ctx context.Context) ([]metricsv1beta1.NodeMetrics, error) {
	metrics, err := c.MetricsClient.MetricsV1beta1().NodeMetricses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return metrics.Items, nil
} 

// 获取单个 Node 原生对象
func (c *Client) GetRawNode(ctx context.Context, name string) (*v1.Node, error) {
	return c.Clientset.CoreV1().Nodes().Get(ctx, name, metav1.GetOptions{})
}

// 获取单个 Node 原生 metrics
func (c *Client) GetRawNodeMetrics(ctx context.Context, name string) (*metricsv1beta1.NodeMetrics, error) {
	return c.MetricsClient.MetricsV1beta1().NodeMetricses().Get(ctx, name, metav1.GetOptions{})
} 

// 获取单个 Pod 原生 metrics
func (c *Client) GetRawPodMetrics(ctx context.Context, namespace, name string) (*metricsv1beta1.PodMetrics, error) {
	return c.MetricsClient.MetricsV1beta1().PodMetricses(namespace).Get(ctx, name, metav1.GetOptions{})
}

// 获取所有 Pod 原生 metrics
func (c *Client) ListRawPodMetrics(ctx context.Context, namespace string) ([]metricsv1beta1.PodMetrics, error) {
	metrics, err := c.MetricsClient.MetricsV1beta1().PodMetricses(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return metrics.Items, nil
}

// 获取 Pod 相关事件
func (c *Client) GetRawPodEvents(ctx context.Context, namespace, name string) ([]v1.Event, error) {
	fieldSelector := fmt.Sprintf("involvedObject.kind=Pod,involvedObject.name=%s,involvedObject.namespace=%s", name, namespace)
	events, err := c.Clientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fieldSelector,
	})
	if err != nil {
		return nil, err
	}
	return events.Items, nil
}

// 获取 Pod 日志
func (c *Client) GetRawPodLogs(ctx context.Context, namespace, name, container string, lines int) ([]string, error) {
	podLogOptions := v1.PodLogOptions{
		Container: container,
		TailLines: int64Ptr(int64(lines)),
	}
	req := c.Clientset.CoreV1().Pods(namespace).GetLogs(name, &podLogOptions)
	podLogs, err := req.Stream(ctx)
	if err != nil {
		return nil, err
	}
	defer podLogs.Close()
	buf := make([]byte, 2048)
	var logContent strings.Builder
	for {
		n, err := podLogs.Read(buf)
		if err != nil {
			break
		}
		logContent.Write(buf[:n])
	}
	logs := strings.Split(logContent.String(), "\n")
	if len(logs) > 0 && logs[len(logs)-1] == "" {
		logs = logs[:len(logs)-1]
	}
	return logs, nil
} 

// ======= 保留 int64Ptr 辅助函数，供 GetRawPodLogs 使用 =======
func int64Ptr(i int64) *int64 {
	return &i
} 