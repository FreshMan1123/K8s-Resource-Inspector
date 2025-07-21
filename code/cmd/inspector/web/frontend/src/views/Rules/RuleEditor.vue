<template>
  <div class="rule-editor">
    <el-form 
      ref="formRef"
      :model="formData"
      :rules="formRules"
      label-width="120px"
      label-position="left"
    >
      <el-form-item label="规则名称" prop="name">
        <el-input v-model="formData.name" placeholder="请输入规则名称" />
      </el-form-item>

      <el-form-item label="规则描述" prop="description">
        <el-input 
          v-model="formData.description" 
          type="textarea" 
          :rows="3"
          placeholder="请输入规则描述"
        />
      </el-form-item>

      <el-form-item label="严重程度" prop="severity">
        <el-select v-model="formData.severity" placeholder="请选择严重程度">
          <el-option label="严重" value="critical" />
          <el-option label="错误" value="error" />
          <el-option label="警告" value="warning" />
          <el-option label="信息" value="info" />
        </el-select>
      </el-form-item>

      <el-form-item label="检查指标" prop="condition.metric">
        <el-input 
          v-model="formData.condition.metric" 
          placeholder="请输入检查指标"
        />
      </el-form-item>

      <el-form-item label="操作符" prop="condition.operator">
        <el-select v-model="formData.condition.operator" placeholder="请选择操作符">
          <el-option label="等于 (==)" value="==" />
          <el-option label="不等于 (!=)" value="!=" />
          <el-option label="大于 (>)" value=">" />
          <el-option label="大于等于 (>=)" value=">=" />
          <el-option label="小于 (<)" value="<" />
          <el-option label="小于等于 (<=)" value="<=" />
          <el-option label="包含 (contains)" value="contains" />
          <el-option label="不包含 (not_contains)" value="not_contains" />
        </el-select>
      </el-form-item>

      <el-form-item label="阈值" prop="condition.threshold">
        <el-input 
          v-model="thresholdInput"
          placeholder="请输入阈值 (支持数字、字符串、JSON对象)"
          @blur="parseThreshold"
        />
        <div class="threshold-help">
          <el-text size="small" type="info">
            支持格式: 数字(80)、字符串("value")、JSON对象({"key":"value"})
          </el-text>
        </div>
      </el-form-item>

      <el-form-item label="修复建议" prop="remediation">
        <el-input 
          v-model="formData.remediation" 
          type="textarea" 
          :rows="4"
          placeholder="请输入修复建议"
        />
      </el-form-item>

      <el-form-item label="启用状态">
        <el-switch 
          v-model="formData.enabled" 
          active-text="启用" 
          inactive-text="禁用"
        />
      </el-form-item>
    </el-form>

    <div class="editor-actions">
      <el-button @click="handleCancel">取消</el-button>
      <el-button type="primary" @click="handleSave" :loading="saving">
        保存
      </el-button>
      <el-button type="info" @click="handleValidate">
        验证规则
      </el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import type { Rule } from '@/api/rules'
import { useRulesStore } from '@/stores/rules'

interface Props {
  rule: Rule
}

interface Emits {
  (e: 'save', rule: Rule): void
  (e: 'cancel'): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const rulesStore = useRulesStore()

// 响应式数据
const formRef = ref<FormInstance>()
const saving = ref(false)
const thresholdInput = ref('')

const formData = reactive<Rule>({
  id: '',
  name: '',
  description: '',
  category: '',
  severity: 'warning',
  enabled: true,
  condition: {
    metric: '',
    operator: '==',
    threshold: ''
  },
  remediation: ''
})

// 表单验证规则
const formRules: FormRules = {
  name: [
    { required: true, message: '请输入规则名称', trigger: 'blur' }
  ],
  description: [
    { required: true, message: '请输入规则描述', trigger: 'blur' }
  ],
  severity: [
    { required: true, message: '请选择严重程度', trigger: 'change' }
  ],
  'condition.metric': [
    { required: true, message: '请输入检查指标', trigger: 'blur' }
  ],
  'condition.operator': [
    { required: true, message: '请选择操作符', trigger: 'change' }
  ],
  remediation: [
    { required: true, message: '请输入修复建议', trigger: 'blur' }
  ]
}

// 方法
const parseThreshold = () => {
  const input = thresholdInput.value.trim()
  if (!input) {
    formData.condition.threshold = ''
    return
  }

  try {
    // 尝试解析为数字
    if (!isNaN(Number(input))) {
      formData.condition.threshold = Number(input)
      return
    }

    // 尝试解析为JSON
    if (input.startsWith('{') || input.startsWith('[')) {
      formData.condition.threshold = JSON.parse(input)
      return
    }

    // 作为字符串处理
    formData.condition.threshold = input
  } catch (error) {
    ElMessage.warning('阈值格式不正确，将作为字符串处理')
    formData.condition.threshold = input
  }
}

const formatThresholdForInput = (threshold: any) => {
  if (typeof threshold === 'object') {
    return JSON.stringify(threshold)
  }
  return String(threshold)
}

const handleSave = async () => {
  if (!formRef.value) return

  try {
    await formRef.value.validate()
    saving.value = true
    
    parseThreshold() // 确保阈值被正确解析
    emit('save', { ...formData })
  } catch (error) {
    ElMessage.error('请检查表单输入')
  } finally {
    saving.value = false
  }
}

const handleCancel = () => {
  emit('cancel')
}

const handleValidate = async () => {
  if (!formRef.value) return

  try {
    await formRef.value.validate()
    parseThreshold()
    
    const result = await rulesStore.validateRule(formData)
    if (result.valid) {
      ElMessage.success('规则验证通过')
    } else {
      ElMessage.error(`规则验证失败: ${result.errors?.join(', ')}`)
    }
  } catch (error) {
    ElMessage.error('验证失败，请检查表单输入')
  }
}

// 初始化表单数据
onMounted(() => {
  Object.assign(formData, props.rule)
  thresholdInput.value = formatThresholdForInput(props.rule.condition.threshold)
})
</script>

<style scoped>
.rule-editor {
  padding: 20px;
}

.threshold-help {
  margin-top: 4px;
}

.editor-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 24px;
  padding-top: 16px;
  border-top: 1px solid var(--border-color);
}
</style>
