kind来启动k8s集群，实际上也就是通过 yaml配置文件的形式来启动的。
第一步是安装kind
# 永久设置（用户级别）
[Environment]::SetEnvironmentVariable("GOPROXY", "https://goproxy.cn,direct", "User")

# 重新打开 PowerShell 窗口后再安装
go install sigs.k8s.io/kind@latest

下载kind
winget install -e --id Kubernetes.kind

将kind加入环境变量
$env:Path = [System.Environment]::GetEnvironmentVariable("Path","User") + ";" + [System.Environment]::GetEnvironmentVariable("Path","Machine")


第二步： 创建集群配置yaml文件
这是一个测试集群的yaml配置文件，需要注意的是，我们的kind最终是通过起docker的方式来创建集群，那么我们的docker是怎么命名的呢? 
以平面控制节点和 工作节点区分，控制平面节点则在集群name后加后缀 -control-plane，工作借点就在后面加-worker。如本例就是 dev-cluster-control-plane
```
kind: Cluster    //表明这是一个kind集群配置
apiVersion: kind.x-k8s.io/v1alpha4
name: dev-cluster         
nodes:
- role: control-plane  //一个平面控制点，单节点集群
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
  extraPortMappings:
  - containerPort: 80
    hostPort: 8080
    protocol: TCP
//可以看到这里有两个容器映射端口，它们最大的区别就是一个HTTP，一个HTTPS，也就是分流流量用的
  - containerPort: 443
    hostPort: 8443
    protocol: TCP 
```

第三步：创建集群
使用以下命令创建集群：
```
kind create cluster --config kind-dev-config.yaml
```

这个命令会：
1. 拉取必要的Docker镜像
2. 创建容器作为Kubernetes节点
3. 初始化控制平面
4. 配置kubectl访问

第四步：连接到集群
Kind会自动配置kubectl，无需SSH连接。创建集群后可以直接使用以下命令：

在只有一个集群的情况下
```
# 查看集群信息
kubectl cluster-info

# 查看节点状态
kubectl get nodes

# 查看所有命名空间的pod
kubectl get pods --all-namespaces
```

如果有多个集群，可以指定使用哪个：
```
# 查看可用的集群上下文
kubectl config get-contexts

# 切换到特定集群  其中 kind-dev-cluster是我们的目标集群
kubectl config use-context kind-dev-cluster
```

第五步：访问集群内部（如需要）
如果需要进入集群节点容器：
```
# 列出Kind创建的容器
docker ps | grep kind

# 进入控制平面节点容器
docker exec -it dev-cluster-control-plane bash
```

第六步：删除集群（使用完毕后）
```
kind delete cluster --name dev-cluster
```

第七步：限制集群资源使用
Kind使用Docker容器运行Kubernetes节点，因此可以通过Docker资源限制来控制集群的内存和CPU使用。

### 方法一：在集群配置文件中设置资源限制
修改集群配置文件，为节点添加资源限制：

```yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
name: dev-cluster
nodes:
- role: control-plane
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
  # 添加资源限制
  extraMounts:
  - containerPath: /var/lib/kubelet/config.json
    hostPath: ./kubelet-config.json
```

然后创建一个`kubelet-config.json`文件，设置资源限制：
```json
{
  "kind": "KubeletConfiguration",
  "apiVersion": "kubelet.config.k8s.io/v1beta1",
  "evictionHard": {
    "memory.available": "200Mi"
  },
  "kubeReserved": {
    "memory": "200Mi",
    "cpu": "200m"
  },
  "systemReserved": {
    "memory": "200Mi",
    "cpu": "200m"
  }
}
```

### 方法二：使用Docker资源限制（推荐）
创建集群后，可以直接对Docker容器设置资源限制：

```bash
# 限制控制平面节点内存为2GB
docker update --memory 2G --memory-swap 2G dev-cluster-control-plane

# 限制工作节点内存为1.5GB
docker update --memory 1.5G --memory-swap 1.5G dev-cluster-worker

# 限制CPU使用
docker update --cpus 2 dev-cluster-control-plane
docker update --cpus 1 dev-cluster-worker
```  

### 方法三：使用Docker Compose
如果使用Docker Compose管理Kind集群，可以在docker-compose.yml中设置资源限制：

```yaml
version: '3'
services:
  kind-control-plane:
    image: kindest/node:v1.27.3
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 2G
```

注意：资源限制设置过低可能会导致集群不稳定或无法正常工作。建议控制平面节点至少分配1.5GB内存，工作节点至少1GB内存。