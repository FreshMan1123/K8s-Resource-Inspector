<template>
  <div class="quick-inspection">
    <h1 class="page-title">快速巡检</h1>
    
    <!-- 巡检模板 -->
    <div class="inspection-templates">
      <h2 class="section-title">选择巡检类型</h2>
      <div class="templates-grid">
        <el-card 
          v-for="template in templates" 
          :key="template.type"
          class="template-card"
          :class="{ 'active': selectedTemplate === template.type }"
          @click="selectTemplate(template.type)"
        >
          <div class="template-content">
            <div class="template-icon">
              <el-icon :size="32" :color="template.color">
                <component :is="template.icon" />
              </el-icon>
            </div>
            <div class="template-info">
              <h3>{{ template.name }}</h3>
              <p>{{ template.description }}</p>
              <div class="template-meta">
                <el-tag size="small">{{ template.ruleCount }} 条规则</el-tag>
                <el-tag size="small" type="info">{{ template.estimatedTime }}</el-tag>
              </div>
            </div>
          </div>
        </el-card>
      </div>
    </div>

    <!-- 执行选项 -->
    <div class="execution-options" v-if="selectedTemplate">
      <h2 class="section-title">执行选项</h2>
      <el-form :model="executionForm" label-width="120px">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="命名空间">
              <el-select 
                v-model="executionForm.namespace" 
                placeholder="选择命名空间"
                clearable
                style="width: 100%"
              >
                <el-option label="所有命名空间" value="" />
                <el-option label="default" value="default" />
                <el-option label="kube-system" value="kube-system" />
                <el-option label="kube-public" value="kube-public" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="输出选项">
              <el-checkbox v-model="executionForm.onlyIssues">仅显示问题</el-checkbox>
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
    </div>

    <!-- 执行按钮 -->
    <div class="execution-actions" v-if="selectedTemplate">
      <el-button 
        type="primary" 
        size="large" 
        @click="executeInspection"
        :loading="executing"
        :disabled="!selectedTemplate"
      >
        <el-icon><VideoPlay /></el-icon>
        开始巡检
      </el-button>
      
      <el-button 
        size="large" 
        @click="$router.push('/inspection/custom')"
      >
        <el-icon><Tools /></el-icon>
        自定义巡检
      </el-button>
    </div>

    <!-- 执行状态 -->
    <div class="execution-status" v-if="currentExecution">
      <h2 class="section-title">执行状态</h2>
      <el-card>
        <div class="status-content">
          <div class="status-header">
            <div class="status-info">
              <h3>{{ getTemplateByType(currentExecution.type)?.name }} 巡检</h3>
              <el-tag :type="getStatusType(currentExecution.status)">
                {{ getStatusLabel(currentExecution.status) }}
              </el-tag>
            </div>
            <div class="status-actions" v-if="currentExecution.status === 'running'">
              <el-button 
                type="danger" 
                size="small" 
                @click="cancelExecution"
              >
                取消执行
              </el-button>
            </div>
          </div>

          <div class="status-progress" v-if="currentExecution.status === 'running'">
            <el-progress 
              :percentage="executionProgress" 
              :status="currentExecution.status === 'failed' ? 'exception' : undefined"
            />
            <p class="progress-text">正在执行巡检，请稍候...</p>
          </div>

          <div class="status-result" v-if="currentExecution.status !== 'running'">
            <div v-if="currentExecution.summary" class="result-summary">
              <div class="summary-item">
                <span class="summary-label">总资源数:</span>
                <span class="summary-value">{{ currentExecution.summary.totalResources }}</span>
              </div>
              <div class="summary-item">
                <span class="summary-label">问题资源:</span>
                <span class="summary-value">{{ currentExecution.summary.resourcesWithIssues }}</span>
              </div>
              <div class="summary-item">
                <span class="summary-label">严重问题:</span>
                <span class="summary-value critical">{{ currentExecution.summary.criticalIssues }}</span>
              </div>
              <div class="summary-item">
                <span class="summary-label">警告问题:</span>
                <span class="summary-value warning">{{ currentExecution.summary.warningIssues }}</span>
              </div>
            </div>

            <div class="result-actions">
              <el-button 
                type="primary" 
                @click="viewDetailResult"
                v-if="currentExecution.status === 'success'"
              >
                查看详细结果
              </el-button>
              <el-button @click="clearExecution">
                清除状态
              </el-button>
            </div>
          </div>

          <div class="error-message" v-if="currentExecution.error">
            <el-alert 
              :title="currentExecution.error" 
              type="error" 
              show-icon 
              :closable="false"
            />
          </div>
        </div>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { 
  VideoPlay, 
  Tools, 
  Box, 
  Monitor, 
  Lock, 
  Document, 
  Service 
} from '@element-plus/icons-vue'
import { useInspectionStore } from '@/stores/inspection'

const router = useRouter()
const inspectionStore = useInspectionStore()

// 响应式数据
const selectedTemplate = ref('')
const executionProgress = ref(0)

const executionForm = reactive({
  namespace: '',
  onlyIssues: false
})

// 巡检模板
const templates = ref([
  {
    type: 'pod',
    name: 'Pod健康检查',
    description: '检查Pod配置、资源限制、重启次数等',
    icon: 'Box',
    color: '#409eff',
    ruleCount: 23,
    estimatedTime: '1-2分钟'
  },
  {
    type: 'node',
    name: '节点状态检查',
    description: '检查节点资源使用率、状态和可调度性',
    icon: 'Monitor',
    color: '#67c23a',
    ruleCount: 18,
    estimatedTime: '30秒-1分钟'
  },
  {
    type: 'deployment',
    name: '部署配置检查',
    description: '检查Deployment副本数、更新策略等',
    icon: 'Document',
    color: '#e6a23c',
    ruleCount: 15,
    estimatedTime: '1-2分钟'
  },
  {
    type: 'service',
    name: '服务配置检查',
    description: '检查Service端口、选择器等配置',
    icon: 'Service',
    color: '#f56c6c',
    ruleCount: 12,
    estimatedTime: '30秒-1分钟'
  },
  {
    type: 'security',
    name: '安全合规检查',
    description: 'CIS基准检查和安全策略验证',
    icon: 'Lock',
    color: '#909399',
    ruleCount: 45,
    estimatedTime: '3-5分钟'
  }
])

// 计算属性
const executing = computed(() => inspectionStore.executing)
const currentExecution = computed(() => inspectionStore.currentExecution)

// 方法
const selectTemplate = (type: string) => {
  selectedTemplate.value = type
}

const getTemplateByType = (type: string) => {
  return templates.value.find(t => t.type === type)
}

const executeInspection = async () => {
  if (!selectedTemplate.value) return

  try {
    const request = {
      type: selectedTemplate.value,
      namespace: executionForm.namespace || undefined,
      onlyIssues: executionForm.onlyIssues
    }

    await inspectionStore.executeInspection(request)
    
    // 模拟进度更新
    const progressInterval = setInterval(() => {
      if (executionProgress.value < 90) {
        executionProgress.value += Math.random() * 20
      }
    }, 500)

    // 检查执行状态
    const statusInterval = setInterval(async () => {
      if (currentExecution.value) {
        try {
          await inspectionStore.getExecutionStatus(currentExecution.value.id)
          if (currentExecution.value.status !== 'running') {
            clearInterval(progressInterval)
            clearInterval(statusInterval)
            executionProgress.value = 100
          }
        } catch (error) {
          clearInterval(progressInterval)
          clearInterval(statusInterval)
        }
      }
    }, 1000)

    ElMessage.success('巡检已开始执行')
  } catch (error) {
    ElMessage.error('启动巡检失败')
  }
}

const cancelExecution = async () => {
  if (currentExecution.value) {
    try {
      await inspectionStore.cancelExecution(currentExecution.value.id)
      ElMessage.info('巡检已取消')
    } catch (error) {
      ElMessage.error('取消巡检失败')
    }
  }
}

const viewDetailResult = () => {
  if (currentExecution.value) {
    router.push(`/history/${currentExecution.value.id}`)
  }
}

const clearExecution = () => {
  inspectionStore.clearCurrentExecution()
  executionProgress.value = 0
}

const getStatusType = (status: string) => {
  const types: Record<string, string> = {
    success: 'success',
    failed: 'danger',
    running: 'warning'
  }
  return types[status] || 'info'
}

const getStatusLabel = (status: string) => {
  const labels: Record<string, string> = {
    success: '执行成功',
    failed: '执行失败',
    running: '执行中'
  }
  return labels[status] || status
}

// 生命周期
onMounted(() => {
  // 清除之前的执行状态
  clearExecution()
})
</script>

<style scoped>
.quick-inspection {
  max-width: 1200px;
  margin: 0 auto;
}

.page-title {
  font-size: 24px;
  font-weight: 600;
  color: var(--text-color);
  margin-bottom: 24px;
}

.section-title {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-color);
  margin-bottom: 16px;
}

.inspection-templates {
  margin-bottom: 32px;
}

.templates-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 20px;
}

.template-card {
  cursor: pointer;
  transition: all 0.3s ease;
  border-radius: 8px;
}

.template-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  transform: translateY(-2px);
}

.template-card.active {
  border-color: var(--primary-color);
  box-shadow: 0 0 0 2px rgba(64, 158, 255, 0.2);
}

.template-content {
  display: flex;
  gap: 16px;
}

.template-icon {
  flex-shrink: 0;
}

.template-info h3 {
  margin: 0 0 8px 0;
  color: var(--text-color);
}

.template-info p {
  margin: 0 0 12px 0;
  color: var(--text-color-secondary);
  line-height: 1.5;
}

.template-meta {
  display: flex;
  gap: 8px;
}

.execution-options {
  margin-bottom: 32px;
}

.execution-actions {
  margin-bottom: 32px;
  text-align: center;
}

.execution-actions .el-button {
  margin: 0 8px;
}

.execution-status {
  margin-bottom: 32px;
}

.status-content {
  padding: 20px;
}

.status-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.status-info h3 {
  margin: 0 0 8px 0;
  color: var(--text-color);
}

.status-progress {
  margin-bottom: 20px;
}

.progress-text {
  margin: 8px 0 0 0;
  color: var(--text-color-secondary);
  text-align: center;
}

.result-summary {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 16px;
  margin-bottom: 20px;
}

.summary-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.summary-label {
  font-size: 14px;
  color: var(--text-color-secondary);
}

.summary-value {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-color);
}

.summary-value.critical {
  color: var(--danger-color);
}

.summary-value.warning {
  color: var(--warning-color);
}

.result-actions {
  display: flex;
  gap: 12px;
  justify-content: center;
}

.error-message {
  margin-top: 16px;
}
</style>
