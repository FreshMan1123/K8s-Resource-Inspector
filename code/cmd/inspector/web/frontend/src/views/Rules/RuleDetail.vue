<template>
  <div class="rule-detail" v-if="rule">
    <div class="detail-header">
      <h2>{{ rule.name }}</h2>
      <el-tag :type="rule.enabled ? 'success' : 'info'">
        {{ rule.enabled ? '启用' : '禁用' }}
      </el-tag>
    </div>

    <div class="detail-content">
      <el-descriptions :column="1" border>
        <el-descriptions-item label="规则ID">
          <el-tag type="info" size="small">{{ rule.id }}</el-tag>
        </el-descriptions-item>
        
        <el-descriptions-item label="规则名称">
          {{ rule.name }}
        </el-descriptions-item>
        
        <el-descriptions-item label="描述">
          {{ rule.description }}
        </el-descriptions-item>
        
        <el-descriptions-item label="分类">
          <el-tag size="small">{{ rule.category }}</el-tag>
        </el-descriptions-item>
        
        <el-descriptions-item label="严重程度">
          <el-tag :type="getSeverityType(rule.severity)">
            {{ getSeverityLabel(rule.severity) }}
          </el-tag>
        </el-descriptions-item>
        
        <el-descriptions-item label="检查条件">
          <div class="condition-detail">
            <div><strong>指标:</strong> {{ rule.condition.metric }}</div>
            <div><strong>操作符:</strong> {{ rule.condition.operator }}</div>
            <div><strong>阈值:</strong> {{ formatThreshold(rule.condition.threshold) }}</div>
          </div>
        </el-descriptions-item>
        
        <el-descriptions-item label="修复建议">
          <div class="remediation-text">{{ rule.remediation }}</div>
        </el-descriptions-item>
        
        <el-descriptions-item label="状态">
          <el-switch 
            v-model="rule.enabled" 
            active-text="启用" 
            inactive-text="禁用"
            @change="handleStatusChange"
          />
        </el-descriptions-item>
      </el-descriptions>
    </div>

    <div class="detail-actions">
      <el-button type="primary" @click="$emit('edit', rule)">
        <el-icon><Edit /></el-icon>
        编辑规则
      </el-button>
      
      <el-button @click="handleTestRule">
        <el-icon><VideoPlay /></el-icon>
        测试规则
      </el-button>
      
      <el-button @click="$emit('close')">
        关闭
      </el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ElMessage } from 'element-plus'
import { Edit, VideoPlay } from '@element-plus/icons-vue'
import type { Rule } from '@/api/rules'

interface Props {
  rule: Rule
}

interface Emits {
  (e: 'edit', rule: Rule): void
  (e: 'close'): void
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
    return JSON.stringify(threshold, null, 2)
  }
  return String(threshold)
}

const handleStatusChange = (enabled: boolean) => {
  ElMessage.success(`规则已${enabled ? '启用' : '禁用'}`)
}

const handleTestRule = () => {
  ElMessage.info('测试规则功能开发中...')
}
</script>

<style scoped>
.rule-detail {
  padding: 20px;
}

.detail-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  padding-bottom: 16px;
  border-bottom: 1px solid var(--border-color);
}

.detail-header h2 {
  margin: 0;
  color: var(--text-color);
}

.detail-content {
  margin-bottom: 24px;
}

.condition-detail {
  display: flex;
  flex-direction: column;
  gap: 8px;
  font-family: 'Courier New', monospace;
  font-size: 14px;
}

.remediation-text {
  line-height: 1.6;
  color: var(--text-color-secondary);
}

.detail-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
  padding-top: 16px;
  border-top: 1px solid var(--border-color);
}
</style>
