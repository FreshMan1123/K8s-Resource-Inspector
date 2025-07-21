import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'Layout',
    component: () => import('@/components/Layout/AppLayout.vue'),
    redirect: '/dashboard',
    children: [
      {
        path: '/dashboard',
        name: 'Dashboard',
        component: () => import('@/views/Dashboard.vue'),
        meta: { title: '仪表板', icon: 'House' }
      },
      {
        path: '/rules',
        name: 'Rules',
        component: () => import('@/views/Rules/RuleList.vue'),
        meta: { title: '规则管理', icon: 'Document' }
      },
      {
        path: '/rules/:category',
        name: 'RuleDetail',
        component: () => import('@/views/Rules/RuleDetail.vue'),
        meta: { title: '规则详情', hidden: true }
      },
      {
        path: '/inspection',
        name: 'Inspection',
        component: () => import('@/views/Inspection/QuickInspection.vue'),
        meta: { title: '快速巡检', icon: 'Search' }
      },
      {
        path: '/inspection/custom',
        name: 'CustomInspection',
        component: () => import('@/views/Inspection/CustomInspection.vue'),
        meta: { title: '自定义巡检', hidden: true }
      },
      {
        path: '/history',
        name: 'History',
        component: () => import('@/views/History/HistoryList.vue'),
        meta: { title: '历史记录', icon: 'Clock' }
      },
      {
        path: '/history/:id',
        name: 'HistoryDetail',
        component: () => import('@/views/History/HistoryDetail.vue'),
        meta: { title: '历史详情', hidden: true }
      },
      {
        path: '/settings',
        name: 'Settings',
        component: () => import('@/views/Settings/SystemSettings.vue'),
        meta: { title: '系统设置', icon: 'Setting' }
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router
