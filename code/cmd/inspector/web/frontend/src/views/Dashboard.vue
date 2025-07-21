<template>
  <div class="dashboard">
    <h1 class="page-title">系统概览</h1>
    
    <!-- 统计卡片 -->
    <div class="stats-grid">
      <el-card class="stat-card">
        <div class="stat-content">
          <div class="stat-icon">
            <el-icon size="32" color="#409eff"><Document /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ totalRules }}</div>
            <div class="stat-label">总规则数</div>
          </div>
        </div>
      </el-card>

      <el-card class="stat-card">
        <div class="stat-content">
          <div class="stat-icon">
            <el-icon size="32" color="#e6a23c"><Warning /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ activeIssues }}</div>
            <div class="stat-label">活跃问题</div>
          </div>
        </div>
      </el-card>

      <el-card class="stat-card">
        <div class="stat-content">
          <div class="stat-icon">
            <el-icon size="32" color="#67c23a"><SuccessFilled /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ healthyResources }}%</div>
            <div class="stat-label">健康资源</div>
          </div>
        </div>
      </el-card>

      <el-card class="stat-card">
        <div class="stat-content">
          <div class="stat-icon">
            <el-icon size="32" color="#909399"><Clock /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ lastCheckTime }}</div>
            <div class="stat-label">最后检查</div>
          </div>
        </div>
      </el-card>
    </div>

    <!-- 快速操作 -->
    <div class="quick-actions">
      <h2 class="section-title">快速操作</h2>
      <div class="action-buttons">
        <el-button 
          type="primary" 
          size="large" 
          @click="quickInspection('pod')"
          :loading="loading"
        >
          <el-icon><Box /></el-icon>
          检查Pod
        </el-button>
        
        <el-button 
          type="success" 
          size="large" 
          @click="quickInspection('node')"
          :loading="loading"
        >
          <el-icon><Monitor /></el-icon>
          检查节点
        </el-button>
        
        <el-button 
          type="warning" 
          size="large" 
          @click="quickInspection('security')"
          :loading="loading"
        >
          <el-icon><Lock /></el-icon>
          安全扫描
        </el-button>
        
        <el-button 
          type="info" 
          size="large" 
          @click="$router.push('/history')"
        >
          <el-icon><Clock /></el-icon>
          查看历史
        </el-button>
      </div>
    </div>

    <!-- 最近巡检结果 -->
    <div class="recent-results" v-if="recentHistory.length > 0">
      <h2 class="section-title">最近巡检结果</h2>
      <el-table :data="recentHistory" style="width: 100%">
        <el-table-column prop="timestamp" label="时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.timestamp) }}
          </template>
        </el-table-column>
        <el-table-column prop="type" label="类型" width="120">
          <template #default="{ row }">
            <el-tag size="small">{{ getTypeLabel(row.type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)" size="small">
              {{ getStatusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="summary" label="结果摘要">
          <template #default="{ row }">
            <span v-if="row.summary">
              总计: {{ row.summary.totalResources }} | 
              问题: {{ row.summary.resourcesWithIssues }} |
              严重: {{ row.summary.criticalIssues }} |
              警告: {{ row.summary.warningIssues }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120">
          <template #default="{ row }">
            <el-button 
              type="text" 
              size="small" 
              @click="$router.push(`/history/${row.id}`)"
            >
              查看详情
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { 
  Document, 
  Warning, 
  SuccessFilled, 
  Clock, 
  Box, 
  Monitor, 
  Lock 
} from '@element-plus/icons-vue'
import { useRulesStore } from '@/stores/rules'
import { useInspectionStore } from '@/stores/inspection'

const router = useRouter()
const rulesStore = useRulesStore()
const inspectionStore = useInspectionStore()

// 响应式数据
const loading = ref(false)
const totalRules = ref(156)
const activeIssues = ref(23)
const healthyResources = ref(89)
const lastCheckTime = ref('2小时前')

// 计算属性
const recentHistory = computed(() => inspectionStore.recentHistory)

// 方法
const quickInspection = async (type: string) => {
  try {
    loading.value = true
    ElMessage.info(`正在执行${getTypeLabel(type)}检查...`)
    
    await inspectionStore.executeInspection({
      type,
      onlyIssues: false
    })
    
    ElMessage.success('检查完成！')
    router.push('/history')
  } catch (error) {
    ElMessage.error('检查失败，请重试')
  } finally {
    loading.value = false
  }
}

const formatTime = (timestamp: string) => {
  return new Date(timestamp).toLocaleString('zh-CN')
}

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
    success: '成功',
    failed: '失败',
    running: '运行中'
  }
  return labels[status] || status
}

// 生命周期
onMounted(async () => {
  try {
    await Promise.all([
      rulesStore.fetchCategories(),
      inspectionStore.fetchHistory(1, 5)
    ])
  } catch (error) {
    console.error('初始化数据失败:', error)
  }
})
</script>

<style scoped>
.dashboard {
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

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 20px;
  margin-bottom: 32px;
}

.stat-card {
  border-radius: 8px;
}

.stat-content {
  display: flex;
  align-items: center;
  gap: 16px;
}

.stat-icon {
  flex-shrink: 0;
}

.stat-info {
  flex: 1;
}

.stat-value {
  font-size: 24px;
  font-weight: 600;
  color: var(--text-color);
  line-height: 1;
}

.stat-label {
  font-size: 14px;
  color: var(--text-color-secondary);
  margin-top: 4px;
}

.quick-actions {
  margin-bottom: 32px;
}

.action-buttons {
  display: flex;
  gap: 16px;
  flex-wrap: wrap;
}

.action-buttons .el-button {
  min-width: 120px;
}

.recent-results {
  margin-bottom: 32px;
}
</style>
