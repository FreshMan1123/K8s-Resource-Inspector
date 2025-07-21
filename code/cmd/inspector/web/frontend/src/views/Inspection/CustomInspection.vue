<template>
  <div class="custom-inspection">
    <h1 class="page-title">自定义巡检</h1>
    
    <el-row :gutter="20">
      <!-- 配置表单 -->
      <el-col :span="16">
        <el-card title="巡检配置">
          <template #header>
            <span>巡检配置</span>
          </template>
          
          <el-form 
            ref="formRef"
            :model="inspectionConfig"
            :rules="formRules"
            label-width="120px"
          >
            <el-form-item label="检查类型" prop="type">
              <el-select 
                v-model="inspectionConfig.type" 
                placeholder="选择检查类型"
                style="width: 100%"
                @change="handleTypeChange"
              >
                <el-option 
                  v-for="type in inspectionTypes" 
                  :key="type.value"
                  :label="type.label" 
                  :value="type.value"
                />
              </el-select>
            </el-form-item>

            <el-form-item label="命名空间" prop="namespace">
              <el-select 
                v-model="inspectionConfig.namespace" 
                placeholder="选择命名空间"
                clearable
                style="width: 100%"
              >
                <el-option label="所有命名空间" value="" />
                <el-option 
                  v-for="ns in namespaces" 
                  :key="ns"
                  :label="ns" 
                  :value="ns"
                />
              </el-select>
            </el-form-item>

            <el-form-item label="资源名称" v-if="inspectionConfig.type === 'pod'">
              <el-input 
                v-model="inspectionConfig.podName" 
                placeholder="Pod名称 (支持通配符 *)"
              />
            </el-form-item>

            <el-form-item label="规则组">
              <el-checkbox-group v-model="inspectionConfig.ruleGroups">
                <el-checkbox 
                  v-for="group in availableRuleGroups" 
                  :key="group.value"
                  :label="group.value"
                >
                  {{ group.label }}
                </el-checkbox>
              </el-checkbox-group>
            </el-form-item>

            <el-form-item label="输出选项">
              <el-checkbox v-model="inspectionConfig.onlyIssues">
                仅显示问题
              </el-checkbox>
            </el-form-item>

            <el-form-item label="自定义规则文件">
              <el-input 
                v-model="inspectionConfig.rulesFile" 
                placeholder="规则文件路径 (可选)"
              />
            </el-form-item>
          </el-form>
        </el-card>
      </el-col>

      <!-- 预览和操作 -->
      <el-col :span="8">
        <el-card title="命令预览">
          <template #header>
            <span>命令预览</span>
          </template>
          
          <div class="command-preview">
            <el-input 
              v-model="generatedCommand"
              type="textarea"
              :rows="6"
              readonly
              placeholder="配置完成后将显示生成的命令"
            />
          </div>

          <div class="preview-actions">
            <el-button 
              type="primary" 
              @click="executeInspection"
              :loading="executing"
              :disabled="!isConfigValid"
              block
            >
              <el-icon><VideoPlay /></el-icon>
              执行巡检
            </el-button>
            
            <el-button 
              @click="saveAsTemplate"
              :disabled="!isConfigValid"
              block
            >
              <el-icon><Collection /></el-icon>
              保存为模板
            </el-button>
            
            <el-button 
              @click="resetConfig"
              block
            >
              <el-icon><Refresh /></el-icon>
              重置配置
            </el-button>
          </div>
        </el-card>

        <!-- 保存的模板 -->
        <el-card title="保存的模板" style="margin-top: 20px;" v-if="savedTemplates.length > 0">
          <template #header>
            <span>保存的模板</span>
          </template>
          
          <div class="saved-templates">
            <div 
              v-for="template in savedTemplates" 
              :key="template.id"
              class="template-item"
            >
              <div class="template-info">
                <div class="template-name">{{ template.name }}</div>
                <div class="template-desc">{{ template.description }}</div>
              </div>
              <div class="template-actions">
                <el-button 
                  size="small" 
                  type="text"
                  @click="loadTemplate(template)"
                >
                  加载
                </el-button>
                <el-button 
                  size="small" 
                  type="text"
                  @click="deleteTemplate(template.id)"
                >
                  删除
                </el-button>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 执行状态 -->
    <div class="execution-status" v-if="currentExecution">
      <ExecutionStatus 
        :execution="currentExecution"
        @cancel="cancelExecution"
        @view-result="viewResult"
        @clear="clearExecution"
      />
    </div>

    <!-- 保存模板对话框 -->
    <el-dialog v-model="saveTemplateVisible" title="保存为模板" width="400px">
      <el-form :model="templateForm" label-width="80px">
        <el-form-item label="模板名称" required>
          <el-input v-model="templateForm.name" placeholder="请输入模板名称" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input 
            v-model="templateForm.description" 
            type="textarea"
            placeholder="请输入模板描述"
          />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="saveTemplateVisible = false">取消</el-button>
        <el-button type="primary" @click="confirmSaveTemplate">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { 
  VideoPlay, 
  Collection, 
  Refresh 
} from '@element-plus/icons-vue'
import { useInspectionStore } from '@/stores/inspection'
import ExecutionStatus from './ExecutionStatus.vue'

const router = useRouter()
const inspectionStore = useInspectionStore()

// 响应式数据
const formRef = ref<FormInstance>()
const saveTemplateVisible = ref(false)

const inspectionConfig = reactive({
  type: '',
  namespace: '',
  podName: '',
  ruleGroups: [] as string[],
  onlyIssues: false,
  rulesFile: ''
})

const templateForm = reactive({
  name: '',
  description: ''
})

const savedTemplates = ref([
  {
    id: '1',
    name: '生产环境Pod检查',
    description: '检查生产环境Pod的资源限制和安全配置',
    config: {
      type: 'pod',
      namespace: 'production',
      ruleGroups: ['resource_limits', 'security_config'],
      onlyIssues: true
    }
  }
])

// 基础数据
const inspectionTypes = [
  { label: 'Pod检查', value: 'pod' },
  { label: '节点检查', value: 'node' },
  { label: '部署检查', value: 'deployment' },
  { label: '服务检查', value: 'service' },
  { label: '安全检查', value: 'security' },
  { label: '漏洞扫描', value: 'vulnerability' }
]

const namespaces = [
  'default',
  'kube-system',
  'kube-public',
  'production',
  'staging',
  'development'
]

const ruleGroupsMap: Record<string, Array<{label: string, value: string}>> = {
  pod: [
    { label: '资源限制', value: 'resource_limits' },
    { label: '安全配置', value: 'security_config' },
    { label: '重启检查', value: 'restart_check' },
    { label: '镜像策略', value: 'image_policy' }
  ],
  node: [
    { label: '资源使用', value: 'resource_usage' },
    { label: '节点状态', value: 'node_status' },
    { label: '污点检查', value: 'taint_check' }
  ],
  deployment: [
    { label: '副本配置', value: 'replica_config' },
    { label: '更新策略', value: 'update_strategy' },
    { label: '资源配置', value: 'resource_config' }
  ],
  security: [
    { label: 'CIS基准', value: 'cis_benchmark' },
    { label: 'RBAC检查', value: 'rbac_check' },
    { label: '网络策略', value: 'network_policy' }
  ]
}

// 表单验证规则
const formRules: FormRules = {
  type: [
    { required: true, message: '请选择检查类型', trigger: 'change' }
  ]
}

// 计算属性
const executing = computed(() => inspectionStore.executing)
const currentExecution = computed(() => inspectionStore.currentExecution)

const availableRuleGroups = computed(() => {
  return ruleGroupsMap[inspectionConfig.type] || []
})

const isConfigValid = computed(() => {
  return inspectionConfig.type !== ''
})

const generatedCommand = computed(() => {
  if (!inspectionConfig.type) return ''
  
  let command = `inspector inspect ${inspectionConfig.type}`
  
  if (inspectionConfig.namespace) {
    command += ` --namespace=${inspectionConfig.namespace}`
  }
  
  if (inspectionConfig.podName && inspectionConfig.type === 'pod') {
    command += ` --pod-name=${inspectionConfig.podName}`
  }
  
  if (inspectionConfig.onlyIssues) {
    command += ' --only-issues'
  }
  
  if (inspectionConfig.rulesFile) {
    command += ` --rules-file=${inspectionConfig.rulesFile}`
  }
  
  return command
})

// 方法
const handleTypeChange = () => {
  inspectionConfig.ruleGroups = []
}

const executeInspection = async () => {
  if (!formRef.value) return
  
  try {
    await formRef.value.validate()
    
    const request = {
      type: inspectionConfig.type,
      namespace: inspectionConfig.namespace || undefined,
      podName: inspectionConfig.podName || undefined,
      onlyIssues: inspectionConfig.onlyIssues,
      rulesFile: inspectionConfig.rulesFile || undefined
    }
    
    await inspectionStore.executeInspection(request)
    ElMessage.success('自定义巡检已开始执行')
  } catch (error) {
    ElMessage.error('启动巡检失败')
  }
}

const saveAsTemplate = () => {
  templateForm.name = ''
  templateForm.description = ''
  saveTemplateVisible.value = true
}

const confirmSaveTemplate = () => {
  if (!templateForm.name) {
    ElMessage.warning('请输入模板名称')
    return
  }
  
  const template = {
    id: Date.now().toString(),
    name: templateForm.name,
    description: templateForm.description,
    config: { ...inspectionConfig }
  }
  
  savedTemplates.value.push(template)
  saveTemplateVisible.value = false
  ElMessage.success('模板保存成功')
}

const loadTemplate = (template: any) => {
  Object.assign(inspectionConfig, template.config)
  ElMessage.success(`已加载模板: ${template.name}`)
}

const deleteTemplate = async (id: string) => {
  try {
    await ElMessageBox.confirm('确定要删除这个模板吗？', '确认删除', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    const index = savedTemplates.value.findIndex(t => t.id === id)
    if (index !== -1) {
      savedTemplates.value.splice(index, 1)
      ElMessage.success('模板删除成功')
    }
  } catch (error) {
    // 用户取消删除
  }
}

const resetConfig = () => {
  Object.assign(inspectionConfig, {
    type: '',
    namespace: '',
    podName: '',
    ruleGroups: [],
    onlyIssues: false,
    rulesFile: ''
  })
}

const cancelExecution = async () => {
  if (currentExecution.value) {
    try {
      await inspectionStore.cancelExecution(currentExecution.value.id)
      ElMessage.info('巡检已取消')
    } catch (error) {
      ElMessage.error('取消巡检失败')
    }
  }
}

const viewResult = () => {
  if (currentExecution.value) {
    router.push(`/history/${currentExecution.value.id}`)
  }
}

const clearExecution = () => {
  inspectionStore.clearCurrentExecution()
}
</script>

<style scoped>
.custom-inspection {
  max-width: 1200px;
  margin: 0 auto;
}

.page-title {
  font-size: 24px;
  font-weight: 600;
  color: var(--text-color);
  margin-bottom: 24px;
}

.command-preview {
  margin-bottom: 16px;
}

.preview-actions {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.saved-templates {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.template-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px;
  border: 1px solid var(--border-color);
  border-radius: 4px;
}

.template-info {
  flex: 1;
}

.template-name {
  font-weight: 600;
  color: var(--text-color);
  margin-bottom: 4px;
}

.template-desc {
  font-size: 12px;
  color: var(--text-color-secondary);
}

.template-actions {
  display: flex;
  gap: 8px;
}

.execution-status {
  margin-top: 32px;
}
</style>
