# K8s巡检管理平台 - 技术设计文档

## 一、技术架构

### 1.1 整体架构
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   前端界面      │    │   后端API       │    │   CLI工具       │
│  Vue 3 +        │◄──►│  Go + Gin       │◄──►│  inspector.exe  │
│  Element Plus   │    │                 │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   浏览器存储    │    │   文件系统      │    │   K8s集群       │
│  LocalStorage   │    │  YAML + JSON    │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### 1.2 技术栈选择
```
前端技术栈:
├── Vue 3: 现代化前端框架，组合式API
├── Element Plus: 企业级UI组件库
├── Axios: HTTP客户端
├── Chart.js: 图表展示 (阶段3)
└── Vue Router: 单页面路由

后端技术栈:
├── Go: 高性能后端语言
├── Gin: 轻量级Web框架
├── Viper: 配置管理
├── Cron: 定时任务调度 (阶段4)
└── YAML: 配置文件解析

存储方案:
├── YAML文件: 规则配置 (保持现有结构)
├── JSON文件: 历史数据和任务配置
└── 文件系统: 按时间分目录存储
```

## 二、数据模型设计

### 2.1 规则数据模型
```go
// 规则模型 (与现有YAML结构保持一致)
type Rule struct {
    ID          string         `yaml:"id" json:"id"`
    Name        string         `yaml:"name" json:"name"`
    Description string         `yaml:"description" json:"description"`
    Category    string         `yaml:"category" json:"category"`
    Condition   RuleCondition  `yaml:"condition" json:"condition"`
    Severity    string         `yaml:"severity" json:"severity"`
    Remediation string         `yaml:"remediation" json:"remediation"`
    Enabled     bool           `yaml:"enabled" json:"enabled"`
    CreatedAt   time.Time      `yaml:"created_at,omitempty" json:"created_at,omitempty"`
    UpdatedAt   time.Time      `yaml:"updated_at,omitempty" json:"updated_at,omitempty"`
}

type RuleCondition struct {
    Metric    string      `yaml:"metric" json:"metric"`
    Operator  string      `yaml:"operator" json:"operator"`
    Threshold interface{} `yaml:"threshold" json:"threshold"`
    Unit      string      `yaml:"unit,omitempty" json:"unit,omitempty"`
}

// 规则配置文件结构
type RulesConfig struct {
    APIVersion string `yaml:"apiVersion"`
    Kind       string `yaml:"kind"`
    Rules      []Rule `yaml:"rules"`
}
```

### 2.2 巡检数据模型
```go
// 巡检历史模型
type InspectionHistory struct {
    ID          string                 `json:"id"`
    Timestamp   time.Time             `json:"timestamp"`
    Type        string                `json:"type"`        // node/pod/deployment/service/security
    Command     string                `json:"command"`     // 执行的CLI命令
    Duration    time.Duration         `json:"duration"`    // 执行时长
    Status      string                `json:"status"`      // success/failed/running
    Summary     InspectionSummary     `json:"summary"`     // 结果摘要
    Results     interface{}           `json:"results"`     // 详细结果
    TriggerType string                `json:"trigger_type"` // manual/scheduled
    TaskID      string                `json:"task_id,omitempty"` // 定时任务ID
    Error       string                `json:"error,omitempty"`   // 错误信息
}

// 巡检结果摘要
type InspectionSummary struct {
    TotalResources      int `json:"total_resources"`
    ResourcesWithIssues int `json:"resources_with_issues"`
    CriticalIssues      int `json:"critical_issues"`
    WarningIssues       int `json:"warning_issues"`
    InfoIssues          int `json:"info_issues"`
    PassedChecks        int `json:"passed_checks"`
    FailedChecks        int `json:"failed_checks"`
}

// 命令执行结果
type CommandResult struct {
    Command   string    `json:"command"`
    Output    string    `json:"output"`
    Success   bool      `json:"success"`
    Error     string    `json:"error,omitempty"`
    Timestamp time.Time `json:"timestamp"`
    Duration  time.Duration `json:"duration"`
}
```

### 2.3 定时任务模型 (阶段4)
```go
// 定时任务模型
type ScheduledTask struct {
    ID           string            `json:"id"`
    Name         string            `json:"name"`
    Description  string            `json:"description"`
    Enabled      bool              `json:"enabled"`
    Schedule     string            `json:"schedule"`     // Cron表达式
    Config       InspectionConfig  `json:"config"`       // 巡检配置
    Notification NotificationConfig `json:"notification"` // 通知配置
    CreatedAt    time.Time         `json:"created_at"`
    UpdatedAt    time.Time         `json:"updated_at"`
    LastRun      *time.Time        `json:"last_run,omitempty"`
    NextRun      *time.Time        `json:"next_run,omitempty"`
}

// 巡检配置
type InspectionConfig struct {
    Type        string   `json:"type"`         // node/pod/deployment/service/security/all
    Namespaces  []string `json:"namespaces"`   // 目标命名空间
    Resources   []string `json:"resources"`    // 具体资源名称
    RuleGroups  []string `json:"rule_groups"`  // 应用的规则分组
    OutputFormat string  `json:"output_format"` // text/json/yaml
    OnlyIssues  bool     `json:"only_issues"`  // 只显示问题
}

// 通知配置
type NotificationConfig struct {
    Enabled     bool     `json:"enabled"`
    OnlyOnIssues bool    `json:"only_on_issues"` // 仅在有问题时通知
    Email       EmailConfig    `json:"email,omitempty"`
    Webhook     WebhookConfig  `json:"webhook,omitempty"`
    Recipients  []string `json:"recipients"`
}
```

## 三、API接口设计

### 3.1 规则管理API
```
GET    /api/rules                    # 获取所有规则
GET    /api/rules/:category          # 获取指定分类的规则
GET    /api/rules/:category/:id      # 获取具体规则详情
POST   /api/rules                    # 创建新规则
PUT    /api/rules/:category/:id      # 更新规则
DELETE /api/rules/:category/:id      # 删除规则
POST   /api/rules/apply              # 应用规则变更
POST   /api/rules/validate           # 验证规则格式
GET    /api/rules/metrics/:category  # 获取分类的可用指标
POST   /api/rules/check-duplicate    # 检查规则重复
```

### 3.2 巡检执行API
```
GET    /api/inspection/templates     # 获取命令模板
POST   /api/inspection/execute       # 执行巡检命令
GET    /api/inspection/status/:id    # 获取执行状态
POST   /api/inspection/cancel/:id    # 取消执行
GET    /api/inspection/history       # 获取历史记录
GET    /api/inspection/history/:id   # 获取具体历史详情
DELETE /api/inspection/history/:id   # 删除历史记录
POST   /api/inspection/compare       # 对比历史结果
GET    /api/inspection/export/:id    # 导出报告
```

### 3.3 定时任务API (阶段4)
```
GET    /api/tasks                    # 获取所有定时任务
POST   /api/tasks                    # 创建定时任务
PUT    /api/tasks/:id                # 更新定时任务
DELETE /api/tasks/:id                # 删除定时任务
POST   /api/tasks/:id/enable         # 启用任务
POST   /api/tasks/:id/disable        # 禁用任务
POST   /api/tasks/:id/run            # 手动执行任务
GET    /api/tasks/:id/history        # 获取任务执行历史
```

## 四、文件存储设计

### 4.1 目录结构
```
项目根目录/
├── code/configs/rules/           # 现有规则文件 (保持不变)
│   ├── node.yaml
│   ├── pod.yaml
│   ├── deployment.yaml
│   ├── service.yaml
│   └── cis-kubernetes-v1.8.yaml
├── code/web/data/               # Web平台数据
│   ├── inspection_history/      # 巡检历史
│   │   └── 2024/               # 按年分目录
│   │       └── 01/             # 按月分目录
│   │           └── 15/         # 按日分目录
│   │               ├── 09-00-00_scheduled_full.json
│   │               ├── 14-30-00_manual_pod.json
│   │               └── ...
│   ├── scheduled_tasks/         # 定时任务配置
│   │   ├── tasks.json          # 任务定义
│   │   └── execution_history.json # 执行历史
│   ├── audit_logs/             # 审计日志
│   │   ├── 2024-01.log        # 按月分文件
│   │   └── ...
│   └── notifications/          # 通知历史
│       └── sent_notifications.json
```

### 4.2 文件命名规范
```
巡检历史文件命名:
格式: {时间}_{触发方式}_{类型}.json
示例: 2024-01-15_14-30-00_manual_pod.json

定时任务文件:
├── tasks.json: 任务配置列表
└── execution_history.json: 执行历史记录

审计日志文件:
格式: {年}-{月}.log
示例: 2024-01.log
```

## 五、前端组件设计

### 5.1 组件层次结构
```
App.vue
├── Layout/
│   ├── Header.vue              # 顶部导航
│   ├── Sidebar.vue             # 侧边栏菜单
│   └── Main.vue                # 主内容区
├── Pages/
│   ├── Dashboard.vue           # 仪表板页面
│   ├── Rules/                  # 规则管理页面
│   │   ├── RuleList.vue        # 规则列表
│   │   ├── RuleEditor.vue      # 规则编辑器
│   │   └── RuleDetail.vue      # 规则详情
│   ├── Inspection/             # 巡检页面
│   │   ├── QuickInspection.vue # 快速巡检
│   │   ├── CustomInspection.vue # 自定义巡检
│   │   └── ExecutionStatus.vue # 执行状态
│   └── History/                # 历史页面
│       ├── HistoryList.vue     # 历史列表
│       ├── HistoryDetail.vue   # 历史详情
│       └── ComparisonView.vue  # 对比视图
└── Components/
    ├── RuleCard.vue            # 规则卡片
    ├── MetricSelector.vue      # 指标选择器
    ├── CommandButton.vue       # 命令按钮
    ├── StatusIndicator.vue     # 状态指示器
    └── ResultDisplay.vue       # 结果显示
```

### 5.2 状态管理设计
```javascript
// Vuex/Pinia状态结构
const store = {
  state: {
    // 规则相关状态
    rules: {
      list: [],              // 规则列表
      categories: [],        // 分类列表
      currentRule: null,     // 当前编辑的规则
      loading: false,        // 加载状态
      error: null           // 错误信息
    },
    
    // 巡检相关状态
    inspection: {
      executing: false,      // 是否正在执行
      currentExecution: null, // 当前执行信息
      history: [],          // 历史记录
      templates: []         // 命令模板
    },
    
    // 系统状态
    system: {
      connected: true,      // 后端连接状态
      version: '',          // 系统版本
      config: {}           // 系统配置
    }
  },
  
  mutations: {
    // 规则相关mutations
    SET_RULES(state, rules) { state.rules.list = rules },
    SET_CURRENT_RULE(state, rule) { state.rules.currentRule = rule },
    
    // 巡检相关mutations
    SET_EXECUTION_STATUS(state, status) { state.inspection.executing = status },
    ADD_HISTORY_ITEM(state, item) { state.inspection.history.unshift(item) }
  },
  
  actions: {
    // 异步操作
    async fetchRules({ commit }) {
      const response = await api.getRules()
      commit('SET_RULES', response.data)
    },
    
    async executeInspection({ commit }, config) {
      commit('SET_EXECUTION_STATUS', true)
      const result = await api.executeInspection(config)
      commit('SET_EXECUTION_STATUS', false)
      return result
    }
  }
}
```

## 六、安全考虑

### 6.1 输入验证
```
前端验证:
├── 表单字段格式验证
├── 文件上传类型限制
└── XSS防护 (Element Plus内置)

后端验证:
├── API参数验证
├── 文件路径验证 (防止路径遍历)
├── 命令注入防护
└── YAML解析安全检查
```

### 6.2 文件系统安全
```
安全措施:
├── 限制文件操作范围 (仅在指定目录)
├── 文件权限控制 (读写权限分离)
├── 备份机制 (操作前自动备份)
└── 操作日志记录 (审计追踪)
```

## 七、性能优化

### 7.1 前端优化
```
优化策略:
├── 组件懒加载
├── 虚拟滚动 (大列表)
├── 防抖节流 (搜索、输入)
├── 缓存策略 (API响应缓存)
└── 代码分割 (路由级别)
```

### 7.2 后端优化
```
优化策略:
├── 并发控制 (限制同时执行的巡检数量)
├── 缓存机制 (规则配置缓存)
├── 分页查询 (历史记录)
├── 异步处理 (长时间运行的巡检)
└── 资源清理 (定期清理历史数据)
```

---

**总结**: 这个技术设计确保了系统的可扩展性、安全性和性能，同时保持了与现有CLI工具的完全兼容性。通过模块化设计，我们可以分阶段实施，每个阶段都能交付可用的功能。
