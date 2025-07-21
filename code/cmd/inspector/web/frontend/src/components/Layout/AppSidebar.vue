<template>
  <div class="layout-sidebar">
    <el-menu
      :default-active="activeMenu"
      class="sidebar-menu"
      router
      :collapse="false"
    >
      <el-menu-item index="/dashboard">
        <el-icon><House /></el-icon>
        <span>仪表板</span>
      </el-menu-item>

      <el-menu-item index="/rules">
        <el-icon><Document /></el-icon>
        <span>规则管理</span>
      </el-menu-item>

      <el-sub-menu index="inspection">
        <template #title>
          <el-icon><Search /></el-icon>
          <span>巡检管理</span>
        </template>
        <el-menu-item index="/inspection">
          <el-icon><Lightning /></el-icon>
          <span>快速巡检</span>
        </el-menu-item>
        <el-menu-item index="/inspection/custom">
          <el-icon><Tools /></el-icon>
          <span>自定义巡检</span>
        </el-menu-item>
      </el-sub-menu>

      <el-menu-item index="/history">
        <el-icon><Clock /></el-icon>
        <span>历史记录</span>
      </el-menu-item>

      <el-menu-item index="/settings">
        <el-icon><Setting /></el-icon>
        <span>系统设置</span>
      </el-menu-item>
    </el-menu>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { 
  House, 
  Document, 
  Search, 
  Lightning, 
  Tools, 
  Clock, 
  Setting 
} from '@element-plus/icons-vue'

const route = useRoute()

// 当前激活的菜单项
const activeMenu = computed(() => {
  const path = route.path
  if (path.startsWith('/inspection')) {
    return '/inspection'
  }
  return path
})
</script>

<style scoped>
.layout-sidebar {
  width: 200px;
  background: #fff;
  border-right: 1px solid var(--border-color);
}

.sidebar-menu {
  border-right: none;
  height: 100%;
}

.sidebar-menu .el-menu-item {
  height: 50px;
  line-height: 50px;
}

.sidebar-menu .el-sub-menu .el-menu-item {
  height: 45px;
  line-height: 45px;
  padding-left: 50px !important;
}
</style>
