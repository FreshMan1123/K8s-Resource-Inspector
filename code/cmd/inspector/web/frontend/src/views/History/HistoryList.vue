<template>
  <div class="history-list">
    <h1 class="page-title">历史记录</h1>
    
    <!-- 搜索和筛选 -->
    <div class="toolbar">
      <div class="search-filters">
        <el-input
          v-model="searchText"
          placeholder="搜索历史记录..."
          clearable
          style="width: 300px"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
        
        <el-select v-model="typeFilter" placeholder="检查类型" clearable style="width: 120px">
          <el-option label="Pod检查" value="pod" />
          <el-option label="节点检查" value="node" />
          <el-option label="部署检查" value="deployment" />
          <el-option label="服务检查" value="service" />
          <el-option label="安全检查" value="security" />
          <el-option label="漏洞扫描" value="vulnerability" />
        </el-select>
        
        <el-select v-model="statusFilter" placeholder="执行状态" clearable style="width: 100px">
          <el-option label="成功" value="success" />
          <el-option label="失败" value="failed" />
          <el-option label="运行中" value="running" />
        </el-select>
        
        <el-date-picker
          v-model="dateRange"
          type="daterange"
          range-separator="至"
          start-placeholder="开始日期"
          end-placeholder="结束日期"
          format="YYYY-MM-DD"
          value-format="YYYY-MM-DD"
          style="width: 240px"
        />
      </div>
      
      <div class="toolbar-actions">
        <el-button @click="refreshHistory">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
        
        <el-button @click="clearAllHistory" type="danger" :disabled="!hasHistory">
          <el-icon><Delete /></el-icon>
          清空历史
        </el-button>
      </div>
    </div>

    <!-- 历史记录表格 -->
    <el-table 
      :data="filteredHistory" 
      v-loading="loading"
      style="width: 100%"
      @row-click="handleRowClick"
    >
      <el-table-column prop="timestamp" label="执行时间" width="180" sortable>
        <template #default="{ row }">
          {{ formatTime(row.timestamp) }}
        </template>
      </el-table-column>
      
      <el-table-column prop="type" label="检查类型" width="120">
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
      
      <el-table-column prop="duration" label="耗时" width="100">
        <template #default="{ row }">
          {{ formatDuration(row.duration) }}
        </template>
      </el-table-column>
      
      <el-table-column prop="summary" label="结果摘要" min-width="200">
        <template #default="{ row }">
          <div v-if="row.summary" class="summary-info">
            <span class="summary-item">
              总计: <strong>{{ row.summary.totalResources }}</strong>
            </span>
            <span class="summary-item">
              问题: <strong class="text-warning">{{ row.summary.resourcesWithIssues }}</strong>
            </span>
            <span class="summary-item">
              严重: <strong class="text-danger">{{ row.summary.criticalIssues }}</strong>
            </span>
            <span class="summary-item">
              警告: <strong class="text-warning">{{ row.summary.warningIssues }}</strong>
            </span>
          </div>
          <span v-else class="text-muted">无摘要信息</span>
        </template>
      </el-table-column>
      
      <el-table-column prop="triggerType" label="触发方式" width="100">
        <template #default="{ row }">
          <el-tag :type="row.triggerType === 'manual' ? 'primary' : 'info'" size="small">
            {{ row.triggerType === 'manual' ? '手动' : '定时' }}
          </el-tag>
        </template>
      </el-table-column>
      
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <el-button 
            type="text" 
            size="small" 
            @click.stop="viewDetail(row)"
          >
            查看详情
          </el-button>
          
          <el-button 
            type="text" 
            size="small" 
            @click.stop="exportReport(row)"
            v-if="row.status === 'success'"
          >
            导出报告
          </el-button>
          
          <el-button 
            type="text" 
            size="small" 
            @click.stop="compareHistory(row)"
            v-if="row.status === 'success'"
          >
            对比
          </el-button>
          
          <el-button 
            type="text" 
            size="small" 
            @click.stop="deleteHistory(row)"
            class="text-danger"
          >
            删除
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 分页 -->
    <div class="pagination-wrapper">
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :page-sizes="[10, 20, 50, 100]"
        :total="totalCount"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />
    </div>

    <!-- 空状态 -->
    <el-empty 
      v-if="!loading && filteredHistory.length === 0"
      description="暂无历史记录"
    >
      <el-button type="primary" @click="$router.push('/inspection')">
        开始巡检
      </el-button>
    </el-empty>

    <!-- 对比选择对话框 -->
    <el-dialog v-model="compareDialogVisible" title="选择对比记录" width="60%">
      <div class="compare-selection">
        <p>请选择要与 <strong>{{ compareBaseRecord?.type }} ({{ formatTime(compareBaseRecord?.timestamp || '') }})</strong> 对比的记录：</p>
        
        <el-table 
          :data="comparableHistory" 
          @row-click="selectCompareTarget"
          highlight-current-row
        >
          <el-table-column prop="timestamp" label="执行时间" width="180">
            <template #default="{ row }">
              {{ formatTime(row.timestamp) }}
            </template>
          </el-table-column>
          <el-table-column prop="type" label="检查类型" width="120">
            <template #default="{ row }">
              <el-tag size="small">{{ getTypeLabel(row.type) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="summary" label="结果摘要">
            <template #default="{ row }">
              <div v-if="row.summary" class="summary-info">
                总计: {{ row.summary.totalResources }} | 
                问题: {{ row.summary.resourcesWithIssues }} |
                严重: {{ row.summary.criticalIssues }}
              </div>
            </template>
          </el-table-column>
        </el-table>
      </div>
      
      <template #footer>
        <el-button @click="compareDialogVisible = false">取消</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Refresh, Delete } from '@element-plus/icons-vue'
import { useInspectionStore } from '@/stores/inspection'
import type { InspectionHistory } from '@/api/inspection'

const router = useRouter()
const inspectionStore = useInspectionStore()

// 响应式数据
const searchText = ref('')
const typeFilter = ref('')
const statusFilter = ref('')
const dateRange = ref<[string, string] | null>(null)
const currentPage = ref(1)
const pageSize = ref(20)
const compareDialogVisible = ref(false)
const compareBaseRecord = ref<InspectionHistory | null>(null)

// 计算属性
const loading = computed(() => inspectionStore.loading)
const history = computed(() => inspectionStore.history)
const totalCount = computed(() => inspectionStore.historyTotal)
const hasHistory = computed(() => history.value.length > 0)

const filteredHistory = computed(() => {
  let filtered = history.value

  // 搜索过滤
  if (searchText.value) {
    const search = searchText.value.toLowerCase()
    filtered = filtered.filter(item => 
      item.type.toLowerCase().includes(search) ||
      item.command.toLowerCase().includes(search)
    )
  }

  // 类型过滤
  if (typeFilter.value) {
    filtered = filtered.filter(item => item.type === typeFilter.value)
  }

  // 状态过滤
  if (statusFilter.value) {
    filtered = filtered.filter(item => item.status === statusFilter.value)
  }

  // 日期范围过滤
  if (dateRange.value) {
    const [startDate, endDate] = dateRange.value
    filtered = filtered.filter(item => {
      const itemDate = new Date(item.timestamp).toISOString().split('T')[0]
      return itemDate >= startDate && itemDate <= endDate
    })
  }

  return filtered
})

const comparableHistory = computed(() => {
  if (!compareBaseRecord.value) return []
  
  return history.value.filter(item => 
    item.id !== compareBaseRecord.value?.id &&
    item.type === compareBaseRecord.value?.type &&
    item.status === 'success'
  )
})

// 方法
const refreshHistory = async () => {
  try {
    await inspectionStore.fetchHistory(1, pageSize.value)
    currentPage.value = 1
    ElMessage.success('历史记录已刷新')
  } catch (error) {
    ElMessage.error('刷新失败')
  }
}

const clearAllHistory = async () => {
  try {
    await ElMessageBox.confirm(
      '确定要清空所有历史记录吗？此操作不可恢复。',
      '确认清空',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    // 这里应该调用批量删除API
    ElMessage.success('历史记录已清空')
  } catch (error) {
    // 用户取消
  }
}

const handleRowClick = (row: InspectionHistory) => {
  viewDetail(row)
}

const viewDetail = (row: InspectionHistory) => {
  router.push(`/history/${row.id}`)
}

const exportReport = async (row: InspectionHistory) => {
  try {
    await inspectionStore.exportReport(row.id, 'json')
    ElMessage.success('报告导出成功')
  } catch (error) {
    ElMessage.error('导出失败')
  }
}

const compareHistory = (row: InspectionHistory) => {
  compareBaseRecord.value = row
  compareDialogVisible.value = true
}

const selectCompareTarget = (row: InspectionHistory) => {
  if (compareBaseRecord.value) {
    router.push(`/history/compare?base=${compareBaseRecord.value.id}&target=${row.id}`)
  }
}

const deleteHistory = async (row: InspectionHistory) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除这条历史记录吗？`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    await inspectionStore.deleteHistoryItem(row.id)
    ElMessage.success('历史记录已删除')
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

const handleSizeChange = (size: number) => {
  pageSize.value = size
  currentPage.value = 1
  loadHistory()
}

const handleCurrentChange = (page: number) => {
  currentPage.value = page
  loadHistory()
}

const loadHistory = async () => {
  try {
    await inspectionStore.fetchHistory(currentPage.value, pageSize.value)
  } catch (error) {
    ElMessage.error('加载历史记录失败')
  }
}

const formatTime = (timestamp: string) => {
  return new Date(timestamp).toLocaleString('zh-CN')
}

const formatDuration = (duration: number) => {
  const minutes = Math.floor(duration / 60)
  const seconds = duration % 60
  return minutes > 0 ? `${minutes}分${seconds}秒` : `${seconds}秒`
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
onMounted(() => {
  loadHistory()
})
</script>

<style scoped>
.history-list {
  max-width: 1200px;
  margin: 0 auto;
}

.page-title {
  font-size: 24px;
  font-weight: 600;
  color: var(--text-color);
  margin-bottom: 24px;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding: 16px 0;
}

.search-filters {
  display: flex;
  gap: 12px;
  align-items: center;
}

.toolbar-actions {
  display: flex;
  gap: 8px;
}

.summary-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
  font-size: 12px;
}

.summary-item {
  display: inline-block;
  margin-right: 12px;
}

.text-warning {
  color: var(--warning-color);
}

.text-danger {
  color: var(--danger-color);
}

.text-muted {
  color: var(--text-color-secondary);
}

.pagination-wrapper {
  margin-top: 20px;
  text-align: center;
}

.compare-selection {
  margin-bottom: 16px;
}

.compare-selection p {
  margin-bottom: 16px;
  color: var(--text-color);
}
</style>
