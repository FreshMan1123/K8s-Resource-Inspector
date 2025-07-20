package models

import (
	"time"
	
	corev1 "k8s.io/api/core/v1"
)

// SecurityContext 表示Pod或容器的安全上下文
type SecurityContext struct {
	// 是否以特权模式运行
	Privileged *bool `json:"privileged,omitempty"`
	// 是否允许特权升级
	AllowPrivilegeEscalation *bool `json:"allowPrivilegeEscalation,omitempty"`
	// 是否以非root用户运行
	RunAsNonRoot *bool `json:"runAsNonRoot,omitempty"`
	// 运行用户ID
	RunAsUser *int64 `json:"runAsUser,omitempty"`
	// 运行组ID
	RunAsGroup *int64 `json:"runAsGroup,omitempty"`
	// 文件系统组ID
	FSGroup *int64 `json:"fsGroup,omitempty"`
	// 是否只读根文件系统
	ReadOnlyRootFilesystem *bool `json:"readOnlyRootFilesystem,omitempty"`
	// 安全能力
	Capabilities *SecurityCapabilities `json:"capabilities,omitempty"`
	// SELinux选项
	SELinuxOptions *corev1.SELinuxOptions `json:"seLinuxOptions,omitempty"`
	// Windows选项
	WindowsOptions *corev1.WindowsSecurityContextOptions `json:"windowsOptions,omitempty"`
	// Seccomp配置
	SeccompProfile *corev1.SeccompProfile `json:"seccompProfile,omitempty"`
}

// SecurityCapabilities 表示容器的安全能力
type SecurityCapabilities struct {
	// 添加的能力
	Add []string `json:"add,omitempty"`
	// 删除的能力
	Drop []string `json:"drop,omitempty"`
}

// ImageSecurityInfo 表示镜像安全信息
type ImageSecurityInfo struct {
	// 完整镜像名称
	FullName string `json:"fullName"`
	// 镜像仓库
	Registry string `json:"registry"`
	// 镜像仓库路径
	Repository string `json:"repository"`
	// 镜像标签
	Tag string `json:"tag"`
	// 镜像摘要
	Digest string `json:"digest,omitempty"`
	// 拉取策略
	PullPolicy string `json:"pullPolicy"`
	// 是否使用latest标签
	IsLatest bool `json:"isLatest"`
	// 是否来自可信仓库
	IsTrusted bool `json:"isTrusted"`
	// 是否有数字签名
	HasSignature bool `json:"hasSignature"`
	// 漏洞扫描结果
	Vulnerabilities []Vulnerability `json:"vulnerabilities,omitempty"`
	// 最后扫描时间
	LastScanned *time.Time `json:"lastScanned,omitempty"`
}

// Vulnerability 表示镜像漏洞信息
type Vulnerability struct {
	// 漏洞ID
	ID string `json:"id"`
	// 严重程度
	Severity string `json:"severity"`
	// 漏洞描述
	Description string `json:"description"`
	// 受影响的包
	Package string `json:"package"`
	// 修复版本
	FixedVersion string `json:"fixedVersion,omitempty"`
	// CVSS评分
	CVSSScore float64 `json:"cvssScore,omitempty"`
}

// NetworkSecurityInfo 表示网络安全信息
type NetworkSecurityInfo struct {
	// 是否存在网络策略
	HasNetworkPolicy bool `json:"hasNetworkPolicy"`
	// 网络策略数量
	NetworkPolicyCount int `json:"networkPolicyCount"`
	// 是否有默认拒绝策略
	HasDefaultDeny bool `json:"hasDefaultDeny"`
	// 允许的入站规则
	AllowedIngress []NetworkRule `json:"allowedIngress,omitempty"`
	// 允许的出站规则
	AllowedEgress []NetworkRule `json:"allowedEgress,omitempty"`
	// 是否使用主机网络
	UsesHostNetwork bool `json:"usesHostNetwork"`
	// 暴露的端口
	ExposedPorts []ExposedPort `json:"exposedPorts,omitempty"`
}

// NetworkRule 表示网络规则
type NetworkRule struct {
	// 协议
	Protocol string `json:"protocol"`
	// 端口
	Port int32 `json:"port,omitempty"`
	// 端口范围
	PortRange string `json:"portRange,omitempty"`
	// 来源/目标
	From []NetworkPeer `json:"from,omitempty"`
	To   []NetworkPeer `json:"to,omitempty"`
}

// NetworkPeer 表示网络对等体
type NetworkPeer struct {
	// 命名空间选择器
	NamespaceSelector map[string]string `json:"namespaceSelector,omitempty"`
	// Pod选择器
	PodSelector map[string]string `json:"podSelector,omitempty"`
	// IP块
	IPBlock string `json:"ipBlock,omitempty"`
}

// ExposedPort 表示暴露的端口
type ExposedPort struct {
	// 端口号
	Port int32 `json:"port"`
	// 协议
	Protocol string `json:"protocol"`
	// 服务类型
	ServiceType string `json:"serviceType,omitempty"`
	// 是否对外暴露
	External bool `json:"external"`
}

// PodSecurityInfo 表示Pod安全信息
type PodSecurityInfo struct {
	// Pod名称
	Name string `json:"name"`
	// 命名空间
	Namespace string `json:"namespace"`
	// 服务账户名称
	ServiceAccountName string `json:"serviceAccountName"`
	// 是否自动挂载服务账户令牌
	AutomountServiceAccountToken *bool `json:"automountServiceAccountToken,omitempty"`
	// Pod安全上下文
	SecurityContext *SecurityContext `json:"securityContext,omitempty"`
	// 容器安全信息
	Containers []ContainerSecurityInfo `json:"containers"`
	// 初始化容器安全信息
	InitContainers []ContainerSecurityInfo `json:"initContainers,omitempty"`
	// 是否使用主机网络
	HostNetwork bool `json:"hostNetwork"`
	// 是否使用主机PID
	HostPID bool `json:"hostPID"`
	// 是否使用主机IPC
	HostIPC bool `json:"hostIPC"`
	// 主机路径挂载
	HostPathMounts []HostPathMount `json:"hostPathMounts,omitempty"`
	// 网络安全信息
	NetworkSecurity *NetworkSecurityInfo `json:"networkSecurity,omitempty"`
}

// ContainerSecurityInfo 表示容器安全信息
type ContainerSecurityInfo struct {
	// 容器名称
	Name string `json:"name"`
	// 镜像安全信息
	Image *ImageSecurityInfo `json:"image"`
	// 容器安全上下文
	SecurityContext *SecurityContext `json:"securityContext,omitempty"`
	// 资源限制
	ResourceLimits map[string]string `json:"resourceLimits,omitempty"`
	// 资源请求
	ResourceRequests map[string]string `json:"resourceRequests,omitempty"`
	// 存活探针
	LivenessProbe bool `json:"livenessProbe"`
	// 就绪探针
	ReadinessProbe bool `json:"readinessProbe"`
	// 启动探针
	StartupProbe bool `json:"startupProbe"`
}

// HostPathMount 表示主机路径挂载
type HostPathMount struct {
	// 主机路径
	HostPath string `json:"hostPath"`
	// 挂载路径
	MountPath string `json:"mountPath"`
	// 挂载类型
	Type string `json:"type,omitempty"`
	// 是否只读
	ReadOnly bool `json:"readOnly"`
	// 是否敏感路径
	IsSensitive bool `json:"isSensitive"`
}

// CISComplianceInfo 表示CIS合规性信息
type CISComplianceInfo struct {
	// CIS基准版本
	Version string `json:"version"`
	// 合规性检查结果
	ComplianceResults []CISComplianceResult `json:"complianceResults"`
	// 总体合规性评分
	OverallScore float64 `json:"overallScore"`
	// 通过的检查数量
	PassedChecks int `json:"passedChecks"`
	// 失败的检查数量
	FailedChecks int `json:"failedChecks"`
	// 跳过的检查数量
	SkippedChecks int `json:"skippedChecks"`
}

// CISComplianceResult 表示单个CIS合规性检查结果
type CISComplianceResult struct {
	// CIS控制项ID
	ControlID string `json:"controlId"`
	// 控制项标题
	Title string `json:"title"`
	// 检查结果
	Status string `json:"status"` // PASS, FAIL, SKIP
	// 严重程度
	Severity string `json:"severity"`
	// 检查描述
	Description string `json:"description"`
	// 修复建议
	Remediation string `json:"remediation"`
	// 检查时间
	CheckedAt time.Time `json:"checkedAt"`
}

// SecurityAnalysisResult 表示安全分析结果，遵循现有AnalysisResult模式
type SecurityAnalysisResult struct {
	// Pod名称
	PodName string `json:"podName"`
	// 命名空间
	Namespace string `json:"namespace"`
	// 分析项列表
	Items []SecurityAnalysisItem `json:"items"`
	// 漏洞报告（可选）
	VulnerabilityReport interface{} `json:"vulnerabilityReport,omitempty"`
}

// SecurityAnalysisItem 表示单个安全分析项
type SecurityAnalysisItem struct {
	// 规则ID
	RuleID string `json:"ruleId"`
	// 指标名称
	Metric string `json:"metric"`
	// 实际值
	Value interface{} `json:"value"`
	// 阈值
	Threshold interface{} `json:"threshold"`
	// 是否通过
	Passed bool `json:"passed"`
	// 严重程度
	Severity string `json:"severity"`
	// 描述
	Description string `json:"description"`
	// 修复建议
	Remediation string `json:"remediation"`
}
