<template>
  <div class="layout-container">
    <!-- 顶部导航 -->
    <AppHeader />
    
    <!-- 主体内容 -->
    <div class="layout-main">
      <!-- 侧边栏 -->
      <AppSidebar />
      
      <!-- 内容区域 -->
      <div class="layout-content">
        <router-view />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import AppHeader from './AppHeader.vue'
import AppSidebar from './AppSidebar.vue'
</script>

<style scoped>
.layout-container {
  height: 100vh;
  display: flex;
  flex-direction: column;
}

.layout-main {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.layout-content {
  flex: 1;
  background: var(--bg-color);
  padding: 20px;
  overflow-y: auto;
}
</style>
