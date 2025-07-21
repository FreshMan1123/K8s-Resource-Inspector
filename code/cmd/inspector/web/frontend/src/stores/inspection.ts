import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { inspectionApi, type InspectionRequest, type InspectionResult, type InspectionHistory } from '@/api/inspection'

export const useInspectionStore = defineStore('inspection', () => {
  // 状态
  const executing = ref(false)
  const currentExecution = ref<InspectionResult | null>(null)
  const history = ref<InspectionHistory[]>([])
  const historyTotal = ref(0)
  const currentHistoryDetail = ref<InspectionResult | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  // 计算属性
  const isExecuting = computed(() => executing.value)
  const hasHistory = computed(() => history.value.length > 0)
  const recentHistory = computed(() => history.value.slice(0, 5))

  // 操作
  const executeInspection = async (request: InspectionRequest) => {
    try {
      executing.value = true
      error.value = null
      currentExecution.value = await inspectionApi.execute(request)
      return currentExecution.value
    } catch (err) {
      error.value = err instanceof Error ? err.message : '执行巡检失败'
      throw err
    } finally {
      executing.value = false
    }
  }

  const getExecutionStatus = async (id: string) => {
    try {
      const result = await inspectionApi.getStatus(id)
      if (currentExecution.value && currentExecution.value.id === id) {
        currentExecution.value = result
      }
      return result
    } catch (err) {
      error.value = err instanceof Error ? err.message : '获取执行状态失败'
      throw err
    }
  }

  const cancelExecution = async (id: string) => {
    try {
      await inspectionApi.cancel(id)
      if (currentExecution.value && currentExecution.value.id === id) {
        currentExecution.value.status = 'failed'
        currentExecution.value.error = '用户取消'
      }
      executing.value = false
    } catch (err) {
      error.value = err instanceof Error ? err.message : '取消执行失败'
      throw err
    }
  }

  const fetchHistory = async (page = 1, size = 20) => {
    try {
      loading.value = true
      error.value = null
      const response = await inspectionApi.getHistory(page, size)
      if (page === 1) {
        history.value = response.items
      } else {
        history.value.push(...response.items)
      }
      historyTotal.value = response.total
    } catch (err) {
      error.value = err instanceof Error ? err.message : '获取历史记录失败'
      throw err
    } finally {
      loading.value = false
    }
  }

  const fetchHistoryDetail = async (id: string) => {
    try {
      loading.value = true
      error.value = null
      currentHistoryDetail.value = await inspectionApi.getHistoryDetail(id)
    } catch (err) {
      error.value = err instanceof Error ? err.message : '获取历史详情失败'
      throw err
    } finally {
      loading.value = false
    }
  }

  const deleteHistoryItem = async (id: string) => {
    try {
      await inspectionApi.deleteHistory(id)
      // 从本地状态中移除
      const index = history.value.findIndex(item => item.id === id)
      if (index !== -1) {
        history.value.splice(index, 1)
        historyTotal.value--
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : '删除历史记录失败'
      throw err
    }
  }

  const exportReport = async (id: string, format = 'json') => {
    try {
      const blob = await inspectionApi.exportReport(id, format)
      // 创建下载链接
      const url = window.URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = url
      link.download = `inspection-report-${id}.${format}`
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
      window.URL.revokeObjectURL(url)
    } catch (err) {
      error.value = err instanceof Error ? err.message : '导出报告失败'
      throw err
    }
  }

  const clearError = () => {
    error.value = null
  }

  const clearCurrentExecution = () => {
    currentExecution.value = null
    executing.value = false
  }

  return {
    // 状态
    executing,
    currentExecution,
    history,
    historyTotal,
    currentHistoryDetail,
    loading,
    error,
    // 计算属性
    isExecuting,
    hasHistory,
    recentHistory,
    // 操作
    executeInspection,
    getExecutionStatus,
    cancelExecution,
    fetchHistory,
    fetchHistoryDetail,
    deleteHistoryItem,
    exportReport,
    clearError,
    clearCurrentExecution
  }
})
