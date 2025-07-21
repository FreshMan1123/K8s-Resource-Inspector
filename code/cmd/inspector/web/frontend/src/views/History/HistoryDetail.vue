<template>
  <div class="history-detail">
    <div class="detail-header">
      <el-button @click="$router.back()" type="text">
        <el-icon><ArrowLeft /></el-icon>
        返回历史记录
      </el-button>
      <h1 class="page-title">巡检详情</h1>
    </div>
    
    <div v-if="historyDetail" class="detail-content">
      <!-- 基本信息 -->
      <el-card class="info-card">
        <template #header>
          <span>基本信息</span>
        </template>
        
        <el-descriptions :column="2" border>
          <el-descriptions-item label="检查类型">
            <el-tag>{{ getTypeLabel(historyDetail.type) }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="执行状态">
            <el-tag :type="getStatusType(historyDetail.status)">
              {{ getStatusLabel(historyDetail.status) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="开始时间">
            {{ formatTime(historyDetail.timestamp) }}
          </el-descriptions-item>
          <el-descriptions-item label="执行时长">
            {{ formatDuration(historyDetail.duration || 0) }}
          </el-descriptions-item>
          <el-descriptions-item label="触发方式">
            <el-tag :type="historyDetail.triggerType === 'manual' ? 'primary' : 'info'">
              {{ historyDetail.triggerType === 'manual' ? '手动执行' : '定时执行' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="执行ID">
            <code>{{ historyDetail.id }}</code>
          </el-descriptions-item>
        </el-descriptions>
      </el-card>

      <!-- 结果摘要 -->
      <el-card class="summary-card" v-if="historyDetail.summary">
        <template #header>
          <span>结果摘要</span>
        </template>
        
        <div class="summary-grid">
          <div class="summary-item">
            <div class="summary-icon">
              <el-icon size="24" color="#409eff"><Document /></el-icon>
            </div>
            <div class="summary-content">
              <div class="summary-value">{{ historyDetail.summary.totalResources }}</div>
              <div class="summary-label">总资源数</div>
            </div>
          </div>
          
          <div class="summary-item">
            <div class="summary-icon">
              <el-icon size="24" color="#e6a23c"><Warning /></el-icon>
            </div>
            <div class="summary-content">
              <div class="summary-value">{{ historyDetail.summary.resourcesWithIssues }}</div>
              <div class="summary-label">问题资源</div>
            </div>
          </div>
          
          <div class="summary-item">
            <div class="summary-icon">
              <el-icon size="24" color="#f56c6c"><CircleCloseFilled /></el-icon>
            </div>
            <div class="summary-content">
              <div class="summary-value critical">{{ historyDetail.summary.criticalIssues }}</div>
              <div class="summary-label">严重问题</div>
            </div>
          </div>
          
          <div class="summary-item">
            <div class="summary-icon">
              <el-icon size="24" color="#e6a23c"><WarningFilled /></el-icon>
            </div>
            <div class="summary-content">
              <div class="summary-value warning">{{ historyDetail.summary.warningIssues }}</div>
              <div class="summary-label">警告问题</div>
            </div>
          </div>
          
          <div class="summary-item">
            <div class="summary-icon">
              <el-icon size="24" color="#909399"><InfoFilled /></el-icon>
            </div>
            <div class="summary-content">
              <div class="summary-value">{{ historyDetail.summary.infoIssues }}</div>
              <div class="summary-label">信息问题</div>
            </div>
          </div>
          
          <div class="summary-item">
            <div class="summary-icon">
              <el-icon size="24" color="#67c23a"><SuccessFilled /></el-icon>
            </div>
            <div class="summary-content">
              <div class="summary-value success">{{ historyDetail.summary.passedChecks }}</div>
              <div class="summary-label">通过检查</div>
            </div>
          </div>
        </div>
        
        <!-- 健康度指示器 -->
        <div class="health-indicator">
          <div class="health-label">整体健康度</div>
          <el-progress 
            :percentage="healthPercentage" 
            :color="getHealthColor(healthPercentage)"
            :stroke-width="12"
            text-inside
          />
        </div>
      </el-card>

      <!-- 详细输出 -->
      <el-card class="output-card" v-if="historyDetail.output">
        <template #header>
          <div class="output-header">
            <span>详细输出</span>
            <div class="output-actions">
              <el-button size="small" @click="copyOutput">
                <el-icon><CopyDocument /></el-icon>
                复制
              </el-button>
              <el-button size="small" @click="downloadOutput">
                <el-icon><Download /></el-icon>
                下载
              </el-button>
            </div>
          </div>
        </template>
        
        <div class="output-content">
          <pre>{{ historyDetail.output }}</pre>
        </div>
      </el-card>

      <!-- 错误信息 -->
      <el-card class="error-card" v-if="historyDetail.error">
        <template #header>
          <span>错误信息</span>
        </template>
        
        <el-alert 
          :title="historyDetail.error" 
          type="error" 
          show-icon 
          :closable="false"
        />
      </el-card>

      <!-- 操作按钮 -->
      <div class="detail-actions">
        <el-button type="primary" @click="exportReport">
          <el-icon><Download /></el-icon>
          导出报告
        </el-button>
        
        <el-button @click="rerunInspection">
          <el-icon><Refresh /></el-icon>
          重新执行
        </el-button>
        
        <el-button @click="compareWithOther">
          <el-icon><Rank /></el-icon>
          对比分析
        </el-button>
        
        <el-button type="danger" @click="deleteRecord">
          <el-icon><Delete /></el-icon>
          删除记录
        </el-button>
      </div>
    </div>

    <!-- 加载状态 -->
    <div v-else-if="loading" class="loading-container">
      <el-skeleton :rows="8" animated />
    </div>

    <!-- 错误状态 -->
    <div v-else class="error-container">
      <el-empty description="未找到巡检记录">
        <el-button type="primary" @click="$router.back()">
          返回历史记录
        </el-button>
      </el-empty>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  ArrowLeft, 
  Document, 
  Warning, 
  CircleCloseFilled, 
  WarningFilled, 
  InfoFilled, 
  SuccessFilled,
  CopyDocument,
  Download,
  Refresh,
  Rank,
  Delete
} from '@element-plus/icons-vue'
import { useInspectionStore } from '@/stores/inspection'

const route = useRoute()
const router = useRouter()
const inspectionStore = useInspectionStore()

// 响应式数据
const loading = ref(true)

// 计算属性
const historyDetail = computed(() => inspectionStore.currentHistoryDetail)

const healthPercentage = computed(() => {
  if (!historyDetail.value?.summary) return 0
  
  const total = historyDetail.value.summary.totalResources
  const issues = historyDetail.value.summary.resourcesWithIssues
  
  if (total === 0) return 100
  return Math.round(((total - issues) / total) * 100)
})

// 方法
const getTypeLabel = (type: string) => {
  const labels: Record<string, string> = {
    pod: 'Pod检查',
    node: '节点检查',
    deployment: '部署检查',
    service: '服务检查',
    security: '安全检查',
    vulnerability: '漏洞扫描'
  }
  return labels[type] || type
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

const copyOutput = async () => {
  if (historyDetail.value?.output) {
    try {
      await navigator.clipboard.writeText(historyDetail.value.output)
      ElMessage.success('输出内容已复制到剪贴板')
    } catch (error) {
      ElMessage.error('复制失败')
    }
  }
}

const downloadOutput = () => {
  if (historyDetail.value?.output) {
    const blob = new Blob([historyDetail.value.output], { type: 'text/plain' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `inspection-output-${historyDetail.value.id}.txt`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)
    ElMessage.success('输出内容已下载')
  }
}

const exportReport = async () => {
  if (historyDetail.value) {
    try {
      await inspectionStore.exportReport(historyDetail.value.id, 'json')
      ElMessage.success('报告导出成功')
    } catch (error) {
      ElMessage.error('导出失败')
    }
  }
}

const rerunInspection = () => {
  ElMessage.info('重新执行功能开发中...')
}

const compareWithOther = () => {
  ElMessage.info('对比分析功能开发中...')
}

const deleteRecord = async () => {
  if (!historyDetail.value) return
  
  try {
    await ElMessageBox.confirm(
      '确定要删除这条巡检记录吗？此操作不可恢复。',
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    await inspectionStore.deleteHistoryItem(historyDetail.value.id)
    ElMessage.success('记录已删除')
    router.push('/history')
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

// 生命周期
onMounted(async () => {
  const id = route.params.id as string
  if (id) {
    try {
      await inspectionStore.fetchHistoryDetail(id)
    } catch (error) {
      ElMessage.error('加载巡检详情失败')
    } finally {
      loading.value = false
    }
  } else {
    loading.value = false
  }
})
</script>

<style scoped>
.history-detail {
  max-width: 1200px;
  margin: 0 auto;
}

.detail-header {
  margin-bottom: 24px;
}

.page-title {
  font-size: 24px;
  font-weight: 600;
  color: var(--text-color);
  margin: 16px 0 0 0;
}

.detail-content {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.info-card,
.summary-card,
.output-card,
.error-card {
  border-radius: 8px;
}

.summary-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
  margin-bottom: 24px;
}

.summary-item {
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

.summary-value.success {
  color: var(--success-color);
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

.output-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.output-actions {
  display: flex;
  gap: 8px;
}

.output-content {
  max-height: 400px;
  overflow-y: auto;
}

.output-content pre {
  background: #f5f5f5;
  padding: 16px;
  border-radius: 4px;
  font-family: 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.4;
  white-space: pre-wrap;
  word-break: break-all;
  margin: 0;
}

.detail-actions {
  display: flex;
  gap: 12px;
  justify-content: center;
  flex-wrap: wrap;
  margin-top: 24px;
}

.loading-container,
.error-container {
  padding: 40px 20px;
  text-align: center;
}
</style>
