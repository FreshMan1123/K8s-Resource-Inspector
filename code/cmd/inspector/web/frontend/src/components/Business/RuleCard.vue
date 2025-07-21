<template>
  <el-card class="rule-card" :class="{ 'disabled': !rule.enabled }">
    <template #header>
      <div class="rule-header">
        <div class="rule-title">
          <span class="rule-name">{{ rule.name }}</span>
          <el-tag 
            :type="rule.enabled ? 'success' : 'info'" 
            size="small"
          >
            {{ rule.enabled ? '启用' : '禁用' }}
          </el-tag>
        </div>
        <div class="rule-id">ID: {{ rule.id }}</div>
      </div>
    </template>

    <div class="rule-content">
      <p class="rule-description">{{ rule.description }}</p>
      
      <div class="rule-meta">
        <div class="meta-item">
          <span class="meta-label">严重程度:</span>
          <el-tag :type="getSeverityType(rule.severity)" size="small">
            {{ getSeverityLabel(rule.severity) }}
          </el-tag>
        </div>
        
        <div class="meta-item">
          <span class="meta-label">检查指标:</span>
          <el-tag type="info" size="small">{{ rule.condition.metric }}</el-tag>
        </div>
        
        <div class="meta-item">
          <span class="meta-label">条件:</span>
          <span class="condition-text">
            {{ rule.condition.operator }} {{ formatThreshold(rule.condition.threshold) }}
          </span>
        </div>
      </div>
    </div>

    <template #footer>
      <div class="rule-actions">
        <el-button size="small" @click="$emit('viewDetail', rule)">
          <el-icon><View /></el-icon>
          查看详情
        </el-button>
        
        <el-button size="small" type="primary" @click="$emit('editRule', rule)">
          <el-icon><Edit /></el-icon>
          编辑
        </el-button>
        
        <el-button 
          size="small" 
          :type="rule.enabled ? 'warning' : 'success'"
          @click="$emit('toggleStatus', rule)"
        >
          <el-icon v-if="rule.enabled"><Close /></el-icon>
          <el-icon v-else><Check /></el-icon>
          {{ rule.enabled ? '禁用' : '启用' }}
        </el-button>
      </div>
    </template>
  </el-card>
</template>

<script setup lang="ts">
import { View, Edit, Close, Check } from '@element-plus/icons-vue'
import type { Rule } from '@/api/rules'

interface Props {
  rule: Rule
}

interface Emits {
  (e: 'viewDetail', rule: Rule): void
  (e: 'editRule', rule: Rule): void
  (e: 'toggleStatus', rule: Rule): void
}

defineProps<Props>()
defineEmits<Emits>()

// 方法
const getSeverityType = (severity: string) => {
  const types: Record<string, string> = {
    critical: 'danger',
    error: 'danger',
    warning: 'warning',
    info: 'info'
  }
  return types[severity] || 'info'
}

const getSeverityLabel = (severity: string) => {
  const labels: Record<string, string> = {
    critical: '严重',
    error: '错误',
    warning: '警告',
    info: '信息'
  }
  return labels[severity] || severity
}

const formatThreshold = (threshold: any) => {
  if (typeof threshold === 'object') {
    return JSON.stringify(threshold)
  }
  return String(threshold)
}
</script>

<style scoped>
.rule-card {
  height: 100%;
  transition: all 0.3s ease;
  border-radius: 8px;
}

.rule-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  transform: translateY(-2px);
}

.rule-card.disabled {
  opacity: 0.7;
}

.rule-header {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.rule-title {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.rule-name {
  font-weight: 600;
  color: var(--text-color);
  font-size: 16px;
}

.rule-id {
  font-size: 12px;
  color: var(--text-color-secondary);
}

.rule-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
  min-height: 120px;
}

.rule-description {
  color: var(--text-color-secondary);
  line-height: 1.5;
  margin: 0;
}

.rule-meta {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
}

.meta-label {
  color: var(--text-color-secondary);
  min-width: 70px;
}

.condition-text {
  color: var(--text-color);
  font-family: 'Courier New', monospace;
  font-size: 13px;
}

.rule-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.rule-actions .el-button {
  flex: 1;
  min-width: 80px;
}
</style>
