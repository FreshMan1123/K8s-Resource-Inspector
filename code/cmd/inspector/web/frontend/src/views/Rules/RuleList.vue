<template>
  <div class="rule-list">
    <h1 class="page-title">规则管理</h1>
    
    <!-- 规则分类标签页 -->
    <el-tabs v-model="activeCategory" @tab-change="handleCategoryChange">
      <el-tab-pane 
        v-for="category in categories" 
        :key="category.name"
        :label="category.displayName"
        :name="category.name"
      >
        <!-- 搜索和操作栏 -->
        <div class="toolbar">
          <div class="search-box">
            <el-input
              v-model="searchText"
              placeholder="搜索规则..."
              clearable
              style="width: 300px"
            >
              <template #prefix>
                <el-icon><Search /></el-icon>
              </template>
            </el-input>
          </div>
          
          <div class="filter-box">
            <el-select v-model="severityFilter" placeholder="严重程度" clearable style="width: 120px">
              <el-option label="严重" value="critical" />
              <el-option label="错误" value="error" />
              <el-option label="警告" value="warning" />
              <el-option label="信息" value="info" />
            </el-select>
            
            <el-select v-model="statusFilter" placeholder="状态" clearable style="width: 100px">
              <el-option label="启用" value="enabled" />
              <el-option label="禁用" value="disabled" />
            </el-select>
          </div>
        </div>

        <!-- 规则列表 -->
        <div class="rules-grid" v-loading="loading">
          <RuleCard
            v-for="rule in filteredRules"
            :key="rule.id"
            :rule="rule"
            @view-detail="handleViewDetail"
            @edit-rule="handleEditRule"
            @toggle-status="handleToggleStatus"
          />
        </div>

        <!-- 空状态 -->
        <el-empty 
          v-if="!loading && filteredRules.length === 0"
          description="没有找到匹配的规则"
        />
      </el-tab-pane>
    </el-tabs>

    <!-- 规则详情抽屉 -->
    <el-drawer
      v-model="detailDrawerVisible"
      title="规则详情"
      size="50%"
      direction="rtl"
    >
      <RuleDetail 
        v-if="currentRule"
        :rule="currentRule"
        @edit="handleEditRule"
        @close="detailDrawerVisible = false"
      />
    </el-drawer>

    <!-- 规则编辑对话框 -->
    <el-dialog
      v-model="editDialogVisible"
      title="编辑规则"
      width="60%"
      :before-close="handleEditCancel"
    >
      <RuleEditor
        v-if="editingRule"
        :rule="editingRule"
        @save="handleSaveRule"
        @cancel="handleEditCancel"
      />
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import { useRulesStore } from '@/stores/rules'
import type { Rule } from '@/api/rules'
import RuleCard from '@/components/Business/RuleCard.vue'
import RuleDetail from './RuleDetail.vue'
import RuleEditor from './RuleEditor.vue'

const rulesStore = useRulesStore()

// 响应式数据
const activeCategory = ref('')
const searchText = ref('')
const severityFilter = ref('')
const statusFilter = ref('')
const detailDrawerVisible = ref(false)
const editDialogVisible = ref(false)
const editingRule = ref<Rule | null>(null)

// 计算属性
const categories = computed(() => rulesStore.categories)
const rules = computed(() => rulesStore.rules)
const currentRule = computed(() => rulesStore.currentRule)
const loading = computed(() => rulesStore.loading)

const filteredRules = computed(() => {
  let filtered = rules.value

  // 搜索过滤
  if (searchText.value) {
    const search = searchText.value.toLowerCase()
    filtered = filtered.filter(rule => 
      rule.name.toLowerCase().includes(search) ||
      rule.description.toLowerCase().includes(search) ||
      rule.id.toLowerCase().includes(search)
    )
  }

  // 严重程度过滤
  if (severityFilter.value) {
    filtered = filtered.filter(rule => rule.severity === severityFilter.value)
  }

  // 状态过滤
  if (statusFilter.value) {
    const enabled = statusFilter.value === 'enabled'
    filtered = filtered.filter(rule => rule.enabled === enabled)
  }

  return filtered
})

// 方法
const handleCategoryChange = async (category: string) => {
  if (category) {
    await rulesStore.fetchRulesByCategory(category)
  }
}

const handleViewDetail = async (rule: Rule) => {
  try {
    await rulesStore.fetchRuleDetail(activeCategory.value, rule.id)
    detailDrawerVisible.value = true
  } catch (error) {
    ElMessage.error('获取规则详情失败')
  }
}

const handleEditRule = (rule: Rule) => {
  editingRule.value = { ...rule }
  editDialogVisible.value = true
}

const handleToggleStatus = async (rule: Rule) => {
  try {
    const action = rule.enabled ? '禁用' : '启用'
    await ElMessageBox.confirm(
      `确定要${action}规则"${rule.name}"吗？`,
      '确认操作',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    await rulesStore.updateRule(activeCategory.value, rule.id, {
      enabled: !rule.enabled
    })

    ElMessage.success(`规则已${action}`)
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('操作失败')
    }
  }
}

const handleSaveRule = async (rule: Rule) => {
  try {
    await rulesStore.updateRule(activeCategory.value, rule.id, rule)
    editDialogVisible.value = false
    editingRule.value = null
    ElMessage.success('规则保存成功')
  } catch (error) {
    ElMessage.error('保存失败')
  }
}

const handleEditCancel = () => {
  editDialogVisible.value = false
  editingRule.value = null
}

// 监听分类变化
watch(activeCategory, (newCategory) => {
  if (newCategory) {
    handleCategoryChange(newCategory)
  }
})

// 生命周期
onMounted(async () => {
  try {
    await rulesStore.fetchCategories()
    if (categories.value.length > 0) {
      activeCategory.value = categories.value[0].name
    }
  } catch (error) {
    ElMessage.error('加载规则分类失败')
  }
})
</script>

<style scoped>
.rule-list {
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

.filter-box {
  display: flex;
  gap: 12px;
}

.rules-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
  gap: 20px;
  margin-bottom: 20px;
}
</style>
