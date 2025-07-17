package models

type Deployment struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	Replicas    int32             `json:"replicas"`
	AvailableReplicas int32       `json:"availableReplicas"`
	Strategy    string            `json:"strategy"`
	Containers  []DeploymentContainer       `json:"containers"`
}

type DeploymentContainer struct {
	Name      string            `json:"name"`
	Image     string            `json:"image"`
	ImagePullPolicy string      `json:"imagePullPolicy"`
	Resources ResourceSpec      `json:"resources"`
}

type ResourceSpec struct {
	Limits   map[string]string  `json:"limits"`
	Requests map[string]string  `json:"requests"`
} 