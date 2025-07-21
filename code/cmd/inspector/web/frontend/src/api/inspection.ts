import request from './index'

// 巡检请求接口
export interface InspectionRequest {
  type: string
  namespace?: string
  podName?: string
  kubeconfig?: string
  contextName?: string
  onlyIssues?: boolean
  rulesFile?: string
}

// 巡检结果接口
export interface InspectionResult {
  id: string
  timestamp: string
  type: string
  status: 'running' | 'success' | 'failed'
  duration?: number
  summary?: {
    totalResources: number
    resourcesWithIssues: number
    criticalIssues: number
    warningIssues: number
    infoIssues: number
  }
  output?: string
  error?: string
}

// 巡检历史接口
export interface InspectionHistory {
  id: string
  timestamp: string
  type: string
  command: string
  duration: number
  status: string
  summary: {
    totalResources: number
    resourcesWithIssues: number
    criticalIssues: number
    warningIssues: number
    infoIssues: number
    passedChecks: number
    failedChecks: number
  }
  triggerType: string
}

// 巡检API
export const inspectionApi = {
  // 执行巡检
  execute(requestData: InspectionRequest): Promise<InspectionResult> {
    // 转换为后端期望的格式
    const backendRequest = {
      template_id: requestData.type,
      parameters: {
        namespace: requestData.namespace || '',
        pod_name: requestData.podName || '',
        kubeconfig: requestData.kubeconfig || '',
        context_name: requestData.contextName || '',
        only_issues: requestData.onlyIssues ? 'true' : 'false',
        rules_file: requestData.rulesFile || ''
      }
    }
    return request.post('/api/inspection/execute', backendRequest)
  },

  // 获取执行状态
  getStatus(id: string): Promise<InspectionResult> {
    return request.get(`/inspection/status/${id}`)
  },

  // 取消执行
  cancel(id: string): Promise<void> {
    return request.post(`/inspection/cancel/${id}`)
  },

  // 获取历史记录
  getHistory(page = 1, size = 20): Promise<{ items: InspectionHistory[]; total: number }> {
    return request.get('/inspection/history', { params: { page, size } })
  },

  // 获取历史详情
  getHistoryDetail(id: string): Promise<InspectionResult> {
    return request.get(`/inspection/history/${id}`)
  },

  // 删除历史记录
  deleteHistory(id: string): Promise<void> {
    return request.delete(`/inspection/history/${id}`)
  },

  // 导出报告
  exportReport(id: string, format = 'json'): Promise<Blob> {
    return request.get(`/inspection/export/${id}`, {
      params: { format },
      responseType: 'blob'
    })
  }
}
