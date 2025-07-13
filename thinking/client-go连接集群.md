获取集群的kube-config
使用kind命令获取， --name 指定集群名称
kind get kubeconfig --name dev-cluster

哪怕本次部署是在本地的，但为了模拟企业开发环境，我们需要对kubeconfig的保存进行最佳实践