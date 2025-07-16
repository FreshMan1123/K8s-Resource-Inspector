package models

import (
	"time"
	
	corev1 "k8s.io/api/core/v1"
	//"k8s.io/apimachinery/pkg/api/resource"
)

// ResourceMetric 表示资源指标
type ResourceMetric struct {
	// 总量（Mi）
	Capacity float64 `json:"capacity"`
	// 可分配量（Mi）
	Allocatable float64 `json:"allocatable"`
	// 已分配量（Mi）
	Allocated float64 `json:"allocated"`
	// 已使用量（Mi）
	Used float64 `json:"used"`
	// 利用率（百分比）
	Utilization float64 `json:"utilization"`
	// 分配率（百分比）
	AllocationRate float64 `json:"allocationRate"`
}

// NodeConditionStatus 表示节点条件状态
type NodeConditionStatus struct {
	// 条件类型
	Type string
	// 条件状态
	Status corev1.ConditionStatus
	// 最后一次转换为此状态的时间
	LastTransitionTime time.Time
	// 原因
	Reason string
	// 消息
	Message string
}

// NodePressureStatus 表示节点压力状态
type NodePressureStatus struct {
	// CPU压力状态
	CPUPressure bool
	// 内存压力状态
	MemoryPressure bool
	// 磁盘压力状态
	DiskPressure bool
	// 网络压力状态
	NetworkPressure bool
	// PID压力状态
	PIDPressure bool
}

// NodeInfo 表示节点信息
type NodeInfo struct {
	// 内核版本
	KernelVersion string
	// 操作系统
	OSImage string
	// 容器运行时版本
	ContainerRuntimeVersion string
	// Kubelet版本
	KubeletVersion string
	// Kube-Proxy版本
	KubeProxyVersion string
	// 架构
	Architecture string
}

// CustomMetric 表示自定义指标
type CustomMetric struct {
	// 指标名称
	Name string
	// 指标值
	Value string
	// 指标类型(gauge, counter等)
	Type string
	// 指标来源
	Source string
	// 单位
	Unit string
	// 是否为关键指标
	IsCritical bool
}

// Node 表示Kubernetes节点及其资源使用情况
type Node struct {
	// 节点名称
	Name string
	// 节点角色（master/worker）
	Roles []string
	// 节点IP地址
	Addresses map[string]string
	// 节点创建时间
	CreationTime time.Time
	// 节点就绪状态
	Ready bool
	// 节点是否可调度
	Schedulable bool
	// 节点标签
	Labels map[string]string
	// 节点污点
	Taints []corev1.Taint
	// 节点信息
	NodeInfo NodeInfo
	// 节点压力状态
	PressureStatus NodePressureStatus
	
	// CPU资源指标
	CPU ResourceMetric
	// 内存资源指标
	Memory ResourceMetric
	// 临时存储资源指标
	EphemeralStorage ResourceMetric
	// Pods数量指标
	Pods ResourceMetric
	
	// 节点上运行的Pod数量
	RunningPods int
	// 节点上的总Pod数量（包括运行中和已完成的）
	TotalPods int
	// 节点条件状态列表
	Conditions []NodeConditionStatus
	
	// 自定义指标
	CustomMetrics map[string]CustomMetric
}

// NodeList 表示节点列表
type NodeList struct {
	// 节点列表
	Items []Node
	// 节点总数
	TotalCount int
	// 就绪节点数量
	ReadyCount int
	// 不可调度节点数量
	NotSchedulableCount int
} 