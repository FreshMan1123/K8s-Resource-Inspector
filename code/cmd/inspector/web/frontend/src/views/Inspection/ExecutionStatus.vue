<template>
  <div class="execution-status">
    <h2 class="section-title">执行状态</h2>
    
    <el-card>
      <div class="status-content">
        <!-- 状态头部 -->
        <div class="status-header">
          <div class="status-info">
            <h3>{{ execution.type }} 巡检</h3>
            <el-tag :type="getStatusType(execution.status)" size="large">
              <el-icon><component :is="getStatusIcon(execution.status)" /></el-icon>
              {{ getStatusLabel(execution.status) }}
            </el-tag>
          </div>
          
          <div class="status-meta">
            <div class="meta-item">
              <span class="meta-label">开始时间:</span>
              <span class="meta-value">{{ formatTime(execution.timestamp) }}</span>
            </div>
            <div class="meta-item" v-if="execution.duration">
              <span class="meta-label">执行时长:</span>
              <span class="meta-value">{{ formatDuration(execution.duration) }}</span>
            </div>
          </div>
        </div>

        <!-- 执行中状态 -->
        <div class="status-progress" v-if="execution.status === 'running'">
          <el-progress 
            :percentage="progressPercentage" 
            :status="execution.status === 'failed' ? 'exception' : undefined"
            :stroke-width="8"
          />
          <div class="progress-info">
            <p class="progress-text">正在执行巡检，请稍候...</p>
            <div class="progress-details" v-if="progressDetails">
              <span>{{ progressDetails }}</span>
            </div>
          </div>
          
          <div class="running-actions">
            <el-button 
              type="danger" 
              @click="$emit('cancel')"
              :loading="cancelling"
            >
              <el-icon><Close /></el-icon>
              取消执行
            </el-button>
          </div>
        </div>

        <!-- 完成状态 -->
        <div class="status-result" v-if="execution.status !== 'running'">
          <!-- 成功结果 -->
          <div v-if="execution.status === 'success' && execution.summary" class="result-summary">
            <div class="summary-grid">
              <div class="summary-card">
                <div class="summary-icon">
                  <el-icon size="24" color="#409eff"><Document /></el-icon>
                </div>
                <div class="summary-content">
                  <div class="summary-value">{{ execution.summary.totalResources }}</div>
                  <div class="summary-label">总资源数</div>
                </div>
              </div>
              
              <div class="summary-card">
                <div class="summary-icon">
                  <el-icon size="24" color="#e6a23c"><Warning /></el-icon>
                </div>
                <div class="summary-content">
                  <div class="summary-value">{{ execution.summary.resourcesWithIssues }}</div>
                  <div class="summary-label">问题资源</div>
                </div>
              </div>
              
              <div class="summary-card">
                <div class="summary-icon">
                  <el-icon size="24" color="#f56c6c"><CircleCloseFilled /></el-icon>
                </div>
                <div class="summary-content">
                  <div class="summary-value critical">{{ execution.summary.criticalIssues }}</div>
                  <div class="summary-label">严重问题</div>
                </div>
              </div>
              
              <div class="summary-card">
                <div class="summary-icon">
                  <el-icon size="24" color="#e6a23c"><WarningFilled /></el-icon>
                </div>
                <div class="summary-content">
                  <div class="summary-value warning">{{ execution.summary.warningIssues }}</div>
                  <div class="summary-label">警告问题</div>
                </div>
              </div>
            </div>
            
            <!-- 健康度指示器 -->
            <div class="health-indicator">
              <div class="health-label">集群健康度</div>
              <el-progress 
                :percentage="healthPercentage" 
                :color="getHealthColor(healthPercentage)"
                :stroke-width="12"
                text-inside
              />
            </div>
          </div>

          <!-- 失败结果 -->
          <div v-if="execution.status === 'failed'" class="result-error">
            <el-alert 
              title="执行失败" 
              type="error" 
              show-icon 
              :closable="false"
            >
              <template #default>
                <div v-if="execution.error">{{ execution.error }}</div>
                <div v-else>巡检执行过程中发生未知错误</div>
              </template>
            </el-alert>
          </div>

          <!-- 操作按钮 -->
          <div class="result-actions">
            <el-button 
              type="primary" 
              @click="$emit('viewResult')"
              v-if="execution.status === 'success'"
            >
              <el-icon><View /></el-icon>
              查看详细结果
            </el-button>
            
            <el-button 
              @click="downloadReport"
              v-if="execution.status === 'success'"
            >
              <el-icon><Download /></el-icon>
              下载报告
            </el-button>
            
            <el-button 
              @click="retryExecution"
              v-if="execution.status === 'failed'"
            >
              <el-icon><Refresh /></el-icon>
              重新执行
            </el-button>
            
            <el-button @click="$emit('clear')">
              <el-icon><Delete /></el-icon>
              清除状态
            </el-button>
          </div>
        </div>

        <!-- 输出日志 -->
        <div class="execution-output" v-if="execution.output">
          <el-collapse>
            <el-collapse-item title="查看执行日志" name="output">
              <pre class="output-content">{{ execution.output }}</pre>
            </el-collapse-item>
          </el-collapse>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { 
  Close, 
  Document, 
  Warning, 
  CircleCloseFilled, 
  WarningFilled,
  View, 
  Download, 
  Refresh, 
  Delete,
  Loading,
  CircleCheckFilled,
  CircleCloseFilled as CircleCloseFilledIcon
} from '@element-plus/icons-vue'
import type { InspectionResult } from '@/api/inspection'

interface Props {
  execution: InspectionResult
}

interface Emits {
  (e: 'cancel'): void
  (e: 'viewResult'): void
  (e: 'clear'): void
}

defineProps<Props>()
defineEmits<Emits>()

// 响应式数据
const cancelling = ref(false)
const progressPercentage = ref(0)
const progressDetails = ref('')

// 计算属性
const healthPercentage = computed(() => {
  const { execution } = defineProps<Props>()
  if (!execution.summary) return 0
  
  const total = execution.summary.totalResources
  const issues = execution.summary.resourcesWithIssues
  
  if (total === 0) return 100
  return Math.round(((total - issues) / total) * 100)
})

// 方法
const getStatusType = (status: string) => {
  const types: Record<string, string> = {
    success: 'success',
    failed: 'danger',
    running: 'warning'
  }
  return types[status] || 'info'
}

const getStatusIcon = (status: string) => {
  const icons: Record<string, string> = {
    success: 'CircleCheckFilled',
    failed: 'CircleCloseFilled',
    running: 'Loading'
  }
  return icons[status] || 'Loading'
}

const getStatusLabel = (status: string) => {
  const labels: Record<string, string> = {
    success: '执行成功',
    failed: '执行失败',
    running: '执行中'
  }
  return labels[status] || status
}

const getHealthColor = (percentage: number) => {
  if (percentage >= 90) return '#67c23a'
  if (percentage >= 70) return '#e6a23c'
  return '#f56c6c'
}

const formatTime = (timestamp: string) => {
  return new Date(timestamp).toLocaleString('zh-CN')
}

const formatDuration = (duration: number) => {
  const minutes = Math.floor(duration / 60)
  const seconds = duration % 60
  return minutes > 0 ? `${minutes}分${seconds}秒` : `${seconds}秒`
}

const downloadReport = () => {
  ElMessage.info('下载功能开发中...')
}

const retryExecution = () => {
  ElMessage.info('重新执行功能开发中...')
}
</script>

<style scoped>
.execution-status {
  margin-bottom: 32px;
}

.section-title {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-color);
  margin-bottom: 16px;
}

.status-content {
  padding: 24px;
}

.status-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 24px;
  padding-bottom: 16px;
  border-bottom: 1px solid var(--border-color);
}

.status-info h3 {
  margin: 0 0 12px 0;
  color: var(--text-color);
  font-size: 20px;
}

.status-meta {
  display: flex;
  flex-direction: column;
  gap: 8px;
  text-align: right;
}

.meta-item {
  display: flex;
  gap: 8px;
}

.meta-label {
  color: var(--text-color-secondary);
  font-size: 14px;
}

.meta-value {
  color: var(--text-color);
  font-size: 14px;
  font-weight: 500;
}

.status-progress {
  margin-bottom: 24px;
}

.progress-info {
  margin: 16px 0;
  text-align: center;
}

.progress-text {
  margin: 0 0 8px 0;
  color: var(--text-color);
  font-size: 16px;
}

.progress-details {
  color: var(--text-color-secondary);
  font-size: 14px;
}

.running-actions {
  text-align: center;
  margin-top: 16px;
}

.result-summary {
  margin-bottom: 24px;
}

.summary-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
  margin-bottom: 24px;
}

.summary-card {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px;
  background: var(--bg-color);
  border-radius: 8px;
  border: 1px solid var(--border-color);
}

.summary-icon {
  flex-shrink: 0;
}

.summary-content {
  flex: 1;
}

.summary-value {
  font-size: 24px;
  font-weight: 600;
  color: var(--text-color);
  line-height: 1;
}

.summary-value.critical {
  color: var(--danger-color);
}

.summary-value.warning {
  color: var(--warning-color);
}

.summary-label {
  font-size: 14px;
  color: var(--text-color-secondary);
  margin-top: 4px;
}

.health-indicator {
  margin-top: 24px;
}

.health-label {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-color);
  margin-bottom: 12px;
}

.result-error {
  margin-bottom: 24px;
}

.result-actions {
  display: flex;
  gap: 12px;
  justify-content: center;
  flex-wrap: wrap;
}

.execution-output {
  margin-top: 24px;
  border-top: 1px solid var(--border-color);
  padding-top: 16px;
}

.output-content {
  background: #f5f5f5;
  padding: 16px;
  border-radius: 4px;
  font-family: 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.4;
  max-height: 300px;
  overflow-y: auto;
  white-space: pre-wrap;
  word-break: break-all;
}
</style>
