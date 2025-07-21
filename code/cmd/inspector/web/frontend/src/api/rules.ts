import request from './index'

// 规则分类接口
export interface RuleCategory {
  name: string
  displayName: string
  description: string
  filePath: string
}

// 规则接口
export interface Rule {
  id: string
  name: string
  description: string
  category: string
  severity: string
  enabled: boolean
  condition: {
    metric: string
    operator: string
    threshold: any
  }
  remediation: string
}

// 规则API
export const rulesApi = {
  // 获取所有规则分类
  getCategories(): Promise<RuleCategory[]> {
    return request.get('/rules')
  },

  // 获取指定分类的规则
  getRulesByCategory(category: string): Promise<{ rules: Rule[] }> {
    return request.get(`/rules/${category}`)
  },

  // 获取具体规则详情
  getRuleDetail(category: string, id: string): Promise<Rule> {
    return request.get(`/rules/${category}/${id}`)
  },

  // 更新规则
  updateRule(category: string, id: string, rule: Partial<Rule>): Promise<void> {
    return request.put(`/rules/${category}/${id}`, rule)
  },

  // 验证规则
  validateRule(rule: Partial<Rule>): Promise<{ valid: boolean; errors?: string[] }> {
    return request.post('/rules/validate', rule)
  }
}
