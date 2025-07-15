package models

import (
	"time"
	
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// Pod 表示Kubernetes Pod及其资源使用情况
type Pod struct {
	// Pod名称
	Name string
	// 命名空间
	Namespace string
	// Pod状态
	Phase corev1.PodPhase
	// Pod状态原因
	Reason string
	// Pod创建时间
	CreationTime time.Time
	// Pod IP地址
	IP string
	// 所在节点名称
	NodeName string
	// 标签
	Labels map[string]string
	// 注解
	Annotations map[string]string
	// 容器列表
	Containers []Container
	// 初始化容器列表
	InitContainers []Container
	// 重启次数（所有容器总和）
	TotalRestarts int
	// 运行时长
	RunningDuration time.Duration
	// 相关事件
	Events []Event
	// 最近日志（可选）
	RecentLogs map[string][]string
	// 是否有就绪探针
	HasReadinessProbe bool
	// 是否有存活探针
	HasLivenessProbe bool
	// 是否有启动探针
	HasStartupProbe bool
	// QoS类别
	QOSClass corev1.PodQOSClass
	// 优先级
	Priority int32
	// 调度到节点的时间
	ScheduledTime *time.Time
}

// Container 表示容器及其资源使用情况
type Container struct {
	// 容器名称
	Name string
	// 容器镜像
	Image string
	// 容器状态
	State corev1.ContainerState
	// 上次状态
	LastState corev1.ContainerState
	// 容器就绪状态
	Ready bool
	// 重启次数
	RestartCount int
	// CPU资源指标
	CPU ResourceMetric
	// 内存资源指标
	Memory ResourceMetric
	// 是否有存活探针
	HasLivenessProbe bool
	// 是否有就绪探针
	HasReadinessProbe bool
	// 是否有启动探针
	HasStartupProbe bool
	// 资源请求
	Requests corev1.ResourceList
	// 资源限制
	Limits corev1.ResourceList
}

// Event 表示与Pod相关的事件
type Event struct {
	// 事件类型
	Type string
	// 事件原因
	Reason string
	// 事件消息
	Message string
	// 事件时间
	Time time.Time
	// 事件计数
	Count int
}

// PodList 表示Pod列表
type PodList struct {
	// Pod列表
	Items []Pod
	// Pod总数
	TotalCount int
	// 运行中Pod数量
	RunningCount int
	// 失败Pod数量
	FailedCount int
	// 挂起Pod数量
	PendingCount int
	// 成功Pod数量
	SucceededCount int
	// 未知状态Pod数量
	UnknownCount int
}

// PodConditionStatus 表示Pod条件状态
type PodConditionStatus struct {
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

// PodStatus 表示Pod状态的详细信息
type PodStatus struct {
	// Pod阶段
	Phase corev1.PodPhase
	// Pod状态原因
	Reason string
	// Pod状态消息
	Message string
	// Pod条件列表
	Conditions []PodConditionStatus
	// Pod就绪状态
	Ready bool
	// Pod调度状态
	Scheduled bool
	// Pod初始化状态
	Initialized bool
} 