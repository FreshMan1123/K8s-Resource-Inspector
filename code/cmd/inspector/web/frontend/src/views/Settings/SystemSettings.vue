<template>
  <div class="system-settings">
    <h1 class="page-title">系统设置</h1>
    
    <el-row :gutter="20">
      <!-- 基本设置 -->
      <el-col :span="12">
        <el-card title="基本设置">
          <template #header>
            <span>基本设置</span>
          </template>
          
          <el-form :model="basicSettings" label-width="120px">
            <el-form-item label="系统名称">
              <el-input v-model="basicSettings.systemName" />
            </el-form-item>
            
            <el-form-item label="默认命名空间">
              <el-select v-model="basicSettings.defaultNamespace" style="width: 100%">
                <el-option label="default" value="default" />
                <el-option label="kube-system" value="kube-system" />
                <el-option label="所有命名空间" value="" />
              </el-select>
            </el-form-item>
            
            <el-form-item label="自动刷新间隔">
              <el-select v-model="basicSettings.refreshInterval" style="width: 100%">
                <el-option label="不自动刷新" :value="0" />
                <el-option label="30秒" :value="30" />
                <el-option label="1分钟" :value="60" />
                <el-option label="5分钟" :value="300" />
                <el-option label="10分钟" :value="600" />
              </el-select>
            </el-form-item>
            
            <el-form-item label="历史记录保留">
              <el-select v-model="basicSettings.historyRetention" style="width: 100%">
                <el-option label="7天" :value="7" />
                <el-option label="30天" :value="30" />
                <el-option label="90天" :value="90" />
                <el-option label="永久保留" :value="0" />
              </el-select>
            </el-form-item>
          </el-form>
        </el-card>
      </el-col>

      <!-- 显示设置 -->
      <el-col :span="12">
        <el-card title="显示设置">
          <template #header>
            <span>显示设置</span>
          </template>
          
          <el-form :model="displaySettings" label-width="120px">
            <el-form-item label="主题模式">
              <el-radio-group v-model="displaySettings.theme">
                <el-radio label="light">浅色主题</el-radio>
                <el-radio label="dark">深色主题</el-radio>
                <el-radio label="auto">跟随系统</el-radio>
              </el-radio-group>
            </el-form-item>
            
            <el-form-item label="语言设置">
              <el-select v-model="displaySettings.language" style="width: 100%">
                <el-option label="简体中文" value="zh-CN" />
                <el-option label="English" value="en-US" />
              </el-select>
            </el-form-item>
            
            <el-form-item label="时间格式">
              <el-select v-model="displaySettings.timeFormat" style="width: 100%">
                <el-option label="24小时制" value="24h" />
                <el-option label="12小时制" value="12h" />
              </el-select>
            </el-form-item>
            
            <el-form-item label="显示选项">
              <el-checkbox-group v-model="displaySettings.showOptions">
                <el-checkbox label="showIcons">显示图标</el-checkbox>
                <el-checkbox label="showTooltips">显示提示</el-checkbox>
                <el-checkbox label="compactMode">紧凑模式</el-checkbox>
              </el-checkbox-group>
            </el-form-item>
          </el-form>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" style="margin-top: 20px;">
      <!-- 通知设置 -->
      <el-col :span="12">
        <el-card title="通知设置">
          <template #header>
            <span>通知设置</span>
          </template>
          
          <el-form :model="notificationSettings" label-width="120px">
            <el-form-item label="启用通知">
              <el-switch v-model="notificationSettings.enabled" />
            </el-form-item>
            
            <el-form-item label="通知类型" v-if="notificationSettings.enabled">
              <el-checkbox-group v-model="notificationSettings.types">
                <el-checkbox label="success">执行成功</el-checkbox>
                <el-checkbox label="error">执行失败</el-checkbox>
                <el-checkbox label="warning">发现问题</el-checkbox>
              </el-checkbox-group>
            </el-form-item>
            
            <el-form-item label="通知方式" v-if="notificationSettings.enabled">
              <el-checkbox-group v-model="notificationSettings.methods">
                <el-checkbox label="browser">浏览器通知</el-checkbox>
                <el-checkbox label="sound">声音提醒</el-checkbox>
              </el-checkbox-group>
            </el-form-item>
          </el-form>
        </el-card>
      </el-col>

      <!-- 高级设置 -->
      <el-col :span="12">
        <el-card title="高级设置">
          <template #header>
            <span>高级设置</span>
          </template>
          
          <el-form :model="advancedSettings" label-width="120px">
            <el-form-item label="调试模式">
              <el-switch v-model="advancedSettings.debugMode" />
            </el-form-item>
            
            <el-form-item label="API超时时间">
              <el-input-number 
                v-model="advancedSettings.apiTimeout" 
                :min="5" 
                :max="300" 
                :step="5"
                style="width: 100%"
              />
              <span style="margin-left: 8px; color: var(--text-color-secondary);">秒</span>
            </el-form-item>
            
            <el-form-item label="并发执行数">
              <el-input-number 
                v-model="advancedSettings.maxConcurrency" 
                :min="1" 
                :max="10" 
                style="width: 100%"
              />
            </el-form-item>
            
            <el-form-item label="缓存设置">
              <el-checkbox-group v-model="advancedSettings.cacheOptions">
                <el-checkbox label="enableCache">启用缓存</el-checkbox>
                <el-checkbox label="cacheRules">缓存规则</el-checkbox>
                <el-checkbox label="cacheResults">缓存结果</el-checkbox>
              </el-checkbox-group>
            </el-form-item>
          </el-form>
        </el-card>
      </el-col>
    </el-row>

    <!-- 系统信息 -->
    <el-row style="margin-top: 20px;">
      <el-col :span="24">
        <el-card title="系统信息">
          <template #header>
            <span>系统信息</span>
          </template>
          
          <el-descriptions :column="3" border>
            <el-descriptions-item label="系统版本">
              {{ systemInfo.version }}
            </el-descriptions-item>
            <el-descriptions-item label="构建时间">
              {{ systemInfo.buildTime }}
            </el-descriptions-item>
            <el-descriptions-item label="运行时间">
              {{ systemInfo.uptime }}
            </el-descriptions-item>
            <el-descriptions-item label="Go版本">
              {{ systemInfo.goVersion }}
            </el-descriptions-item>
            <el-descriptions-item label="平台">
              {{ systemInfo.platform }}
            </el-descriptions-item>
            <el-descriptions-item label="架构">
              {{ systemInfo.arch }}
            </el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>
    </el-row>

    <!-- 操作按钮 -->
    <div class="settings-actions">
      <el-button type="primary" @click="saveSettings" :loading="saving">
        <el-icon><Check /></el-icon>
        保存设置
      </el-button>
      
      <el-button @click="resetSettings">
        <el-icon><Refresh /></el-icon>
        重置为默认
      </el-button>
      
      <el-button @click="exportSettings">
        <el-icon><Download /></el-icon>
        导出配置
      </el-button>
      
      <el-button @click="importSettings">
        <el-icon><Upload /></el-icon>
        导入配置
      </el-button>
    </div>

    <!-- 导入配置对话框 -->
    <el-dialog v-model="importDialogVisible" title="导入配置" width="400px">
      <el-upload
        ref="uploadRef"
        :auto-upload="false"
        :show-file-list="false"
        accept=".json"
        :on-change="handleFileChange"
      >
        <el-button type="primary">
          <el-icon><Upload /></el-icon>
          选择配置文件
        </el-button>
      </el-upload>
      
      <div v-if="importFile" style="margin-top: 16px;">
        <p>已选择文件: {{ importFile.name }}</p>
      </div>
      
      <template #footer>
        <el-button @click="importDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="confirmImport" :disabled="!importFile">
          导入
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Check, Refresh, Download, Upload } from '@element-plus/icons-vue'

// 响应式数据
const saving = ref(false)
const importDialogVisible = ref(false)
const importFile = ref<File | null>(null)

const basicSettings = reactive({
  systemName: 'K8s Resource Inspector',
  defaultNamespace: 'default',
  refreshInterval: 60,
  historyRetention: 30
})

const displaySettings = reactive({
  theme: 'light',
  language: 'zh-CN',
  timeFormat: '24h',
  showOptions: ['showIcons', 'showTooltips']
})

const notificationSettings = reactive({
  enabled: true,
  types: ['error', 'warning'],
  methods: ['browser']
})

const advancedSettings = reactive({
  debugMode: false,
  apiTimeout: 30,
  maxConcurrency: 3,
  cacheOptions: ['enableCache', 'cacheRules']
})

const systemInfo = reactive({
  version: 'v1.0.0',
  buildTime: '2024-01-15 14:30:00',
  uptime: '2天3小时45分钟',
  goVersion: 'go1.21.0',
  platform: 'windows',
  arch: 'amd64'
})

// 方法
const saveSettings = async () => {
  try {
    saving.value = true
    
    // 模拟保存设置
    await new Promise(resolve => setTimeout(resolve, 1000))
    
    ElMessage.success('设置保存成功')
  } catch (error) {
    ElMessage.error('保存设置失败')
  } finally {
    saving.value = false
  }
}

const resetSettings = async () => {
  try {
    await ElMessageBox.confirm(
      '确定要重置所有设置为默认值吗？',
      '确认重置',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    // 重置为默认值
    Object.assign(basicSettings, {
      systemName: 'K8s Resource Inspector',
      defaultNamespace: 'default',
      refreshInterval: 60,
      historyRetention: 30
    })
    
    Object.assign(displaySettings, {
      theme: 'light',
      language: 'zh-CN',
      timeFormat: '24h',
      showOptions: ['showIcons', 'showTooltips']
    })
    
    Object.assign(notificationSettings, {
      enabled: true,
      types: ['error', 'warning'],
      methods: ['browser']
    })
    
    Object.assign(advancedSettings, {
      debugMode: false,
      apiTimeout: 30,
      maxConcurrency: 3,
      cacheOptions: ['enableCache', 'cacheRules']
    })
    
    ElMessage.success('设置已重置为默认值')
  } catch (error) {
    // 用户取消
  }
}

const exportSettings = () => {
  const settings = {
    basic: basicSettings,
    display: displaySettings,
    notification: notificationSettings,
    advanced: advancedSettings
  }
  
  const blob = new Blob([JSON.stringify(settings, null, 2)], {
    type: 'application/json'
  })
  
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = 'k8s-inspector-settings.json'
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
  
  ElMessage.success('配置已导出')
}

const importSettings = () => {
  importFile.value = null
  importDialogVisible.value = true
}

const handleFileChange = (file: any) => {
  importFile.value = file.raw
}

const confirmImport = async () => {
  if (!importFile.value) return
  
  try {
    const text = await importFile.value.text()
    const settings = JSON.parse(text)
    
    // 验证配置格式
    if (!settings.basic || !settings.display || !settings.notification || !settings.advanced) {
      throw new Error('配置文件格式不正确')
    }
    
    // 应用配置
    Object.assign(basicSettings, settings.basic)
    Object.assign(displaySettings, settings.display)
    Object.assign(notificationSettings, settings.notification)
    Object.assign(advancedSettings, settings.advanced)
    
    importDialogVisible.value = false
    ElMessage.success('配置导入成功')
  } catch (error) {
    ElMessage.error('配置文件格式错误或导入失败')
  }
}

// 生命周期
onMounted(() => {
  // 加载系统信息
  // 这里可以调用API获取实际的系统信息
})
</script>

<style scoped>
.system-settings {
  max-width: 1200px;
  margin: 0 auto;
}

.page-title {
  font-size: 24px;
  font-weight: 600;
  color: var(--text-color);
  margin-bottom: 24px;
}

.settings-actions {
  margin-top: 32px;
  text-align: center;
}

.settings-actions .el-button {
  margin: 0 8px;
}
</style>
