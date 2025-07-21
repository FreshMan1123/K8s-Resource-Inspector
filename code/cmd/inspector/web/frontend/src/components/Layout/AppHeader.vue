<template>
  <div class="layout-header">
    <div class="header-content">
      <!-- Logo和标题 -->
      <div class="header-left">
        <div class="logo">
          <el-icon size="24" color="#409eff">
            <Monitor />
          </el-icon>
          <span class="title">K8s Resource Inspector</span>
        </div>
      </div>

      <!-- 中间状态信息 -->
      <div class="header-center">
        <div class="status-info">
          <el-tag :type="connectionStatus.type" size="small">
            <el-icon><Connection /></el-icon>
            {{ connectionStatus.text }}
          </el-tag>
        </div>
      </div>

      <!-- 右侧操作区 -->
      <div class="header-right">
        <div class="system-info">
          <el-tooltip content="系统版本" placement="bottom">
            <el-tag size="small" type="info">v1.0.0</el-tag>
          </el-tooltip>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { Monitor, Connection } from '@element-plus/icons-vue'

// 连接状态
const connected = ref(true)

const connectionStatus = computed(() => {
  return connected.value 
    ? { type: 'success', text: '已连接集群' }
    : { type: 'danger', text: '连接断开' }
})
</script>

<style scoped>
.layout-header {
  height: 60px;
  background: #fff;
  border-bottom: 1px solid var(--border-color);
  box-shadow: 0 1px 4px rgba(0,21,41,.08);
}

.header-content {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
}

.header-left {
  display: flex;
  align-items: center;
}

.logo {
  display: flex;
  align-items: center;
  gap: 8px;
}

.title {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-color);
}

.header-center {
  display: flex;
  align-items: center;
}

.status-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 16px;
}

.system-info {
  display: flex;
  align-items: center;
  gap: 8px;
}
</style>
