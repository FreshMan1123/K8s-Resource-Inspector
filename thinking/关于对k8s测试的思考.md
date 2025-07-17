B. 企业最佳实践：分层测试+Mock+集成环境
单元测试：只测纯逻辑，mock 掉所有外部依赖（如你现在这样，直接传 model）
集成测试：用 mock/fake K8s API server 或本地 kind 集群，测试 collector/analyzer/engine 的真实数据流
端到端测试（E2E）：在专门的测试集群上，部署完整系统，跑全流程巡检
3. 企业CI环境的常见做法
A. 本地/云端临时K8s集群
用 kind、minikube、k3d 在 CI 里启动一个临时 K8s 集群
在 CI 脚本里自动 apply 测试资源、跑巡检、清理环境
这种方式最接近真实环境，但 CI 时间会变长
B. 专用测试集群
企业会有一套“CI专用K8s集群”，所有集成测试都连这个集群
需要做好资源隔离、权限控制、数据清理
C. Mock/Fake Client
用 client-go 的 fake client 替代真实 K8s API
只测 collector/analyzer 的数据流，不依赖真实集群
适合大部分集成测试，速度快、无副作用
D. GitHub Actions Runner 环境
可以用自托管 runner，部署在和开发环境一致的服务器上
也可以用云 runner + kind/k3s 等临时集群
4. 最佳实践建议
单元测试：mock 一切外部依赖，保证纯逻辑覆盖率
集成测试：用 fake client 或 kind 集群，覆盖 collector/analyzer/engine 的真实交互
E2E测试：在专用测试集群或自托管 runner 上，跑全流程
CI/CD：根据项目重要性和资源，选择合适的测试深度和环境