import { createRouter, createWebHistory } from 'vue-router'
import DailyView from './views/DailyView.vue'

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'daily', component: DailyView },
    { path: '/practices', name: 'practices', component: () => import('./views/PracticesView.vue') },
    { path: '/practices/:id/history', name: 'history', component: () => import('./views/HistoryView.vue') },
    { path: '/memorize', name: 'memorize', component: () => import('./views/MemorizeView.vue') },
    { path: '/tasks', name: 'tasks', component: () => import('./views/TasksView.vue') },
    { path: '/reports', name: 'reports', component: () => import('./views/ReportsView.vue') },
  ],
})
