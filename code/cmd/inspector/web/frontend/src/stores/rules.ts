import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { rulesApi, type RuleCategory, type Rule } from '@/api/rules'

export const useRulesStore = defineStore('rules', () => {
  // 状态
  const categories = ref<RuleCategory[]>([])
  const currentCategory = ref<string>('')
  const rules = ref<Rule[]>([])
  const currentRule = ref<Rule | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  // 计算属性
  const enabledRules = computed(() => rules.value.filter(rule => rule.enabled))
  const disabledRules = computed(() => rules.value.filter(rule => !rule.enabled))
  const rulesBySeverity = computed(() => {
    const result: Record<string, Rule[]> = {}
    rules.value.forEach(rule => {
      if (!result[rule.severity]) {
        result[rule.severity] = []
      }
      result[rule.severity].push(rule)
    })
    return result
  })

  // 操作
  const fetchCategories = async () => {
    try {
      loading.value = true
      error.value = null
      categories.value = await rulesApi.getCategories()
    } catch (err) {
      error.value = err instanceof Error ? err.message : '获取规则分类失败'
      throw err
    } finally {
      loading.value = false
    }
  }

  const fetchRulesByCategory = async (category: string) => {
    try {
      loading.value = true
      error.value = null
      currentCategory.value = category
      const response = await rulesApi.getRulesByCategory(category)
      rules.value = response.rules || []
    } catch (err) {
      error.value = err instanceof Error ? err.message : '获取规则失败'
      throw err
    } finally {
      loading.value = false
    }
  }

  const fetchRuleDetail = async (category: string, id: string) => {
    try {
      loading.value = true
      error.value = null
      currentRule.value = await rulesApi.getRuleDetail(category, id)
    } catch (err) {
      error.value = err instanceof Error ? err.message : '获取规则详情失败'
      throw err
    } finally {
      loading.value = false
    }
  }

  const updateRule = async (category: string, id: string, rule: Partial<Rule>) => {
    try {
      loading.value = true
      error.value = null
      await rulesApi.updateRule(category, id, rule)
      // 更新本地状态
      const index = rules.value.findIndex(r => r.id === id)
      if (index !== -1) {
        rules.value[index] = { ...rules.value[index], ...rule }
      }
      if (currentRule.value && currentRule.value.id === id) {
        currentRule.value = { ...currentRule.value, ...rule }
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : '更新规则失败'
      throw err
    } finally {
      loading.value = false
    }
  }

  const validateRule = async (rule: Partial<Rule>) => {
    try {
      return await rulesApi.validateRule(rule)
    } catch (err) {
      error.value = err instanceof Error ? err.message : '验证规则失败'
      throw err
    }
  }

  const clearError = () => {
    error.value = null
  }

  return {
    // 状态
    categories,
    currentCategory,
    rules,
    currentRule,
    loading,
    error,
    // 计算属性
    enabledRules,
    disabledRules,
    rulesBySeverity,
    // 操作
    fetchCategories,
    fetchRulesByCategory,
    fetchRuleDetail,
    updateRule,
    validateRule,
    clearError
  }
})
