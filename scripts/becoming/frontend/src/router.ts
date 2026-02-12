import { createRouter, createWebHistory } from 'vue-router'
import DailyView from './views/DailyView.vue'
import { useAuth } from './composables/useAuth'

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    // Public routes (no auth required)
    { path: '/login', name: 'login', component: () => import('./views/LoginView.vue'), meta: { public: true } },
    { path: '/register', name: 'register', component: () => import('./views/RegisterView.vue'), meta: { public: true } },

    // Protected routes
    { path: '/', name: 'daily', component: DailyView },
    { path: '/practices', name: 'practices', component: () => import('./views/PracticesView.vue') },
    { path: '/practices/:id/history', name: 'history', component: () => import('./views/HistoryView.vue') },
    { path: '/memorize', name: 'memorize', component: () => import('./views/MemorizeView.vue') },
    { path: '/tasks', name: 'tasks', component: () => import('./views/TasksView.vue') },
    { path: '/notes', name: 'notes', component: () => import('./views/NotesView.vue') },
    { path: '/reflections', name: 'reflections', component: () => import('./views/ReflectionsView.vue') },
    { path: '/pillars', name: 'pillars', component: () => import('./views/PillarsView.vue') },
    { path: '/reports', name: 'reports', component: () => import('./views/ReportsView.vue') },
  ],
})

// Auth guard — redirect to /login if not authenticated
router.beforeEach(async (to) => {
  const { isAuthenticated, loading, init } = useAuth()

  // Ensure auth is initialized (calls /api/me once)
  await init()

  // Wait for loading to finish
  if (loading.value) return

  // Allow public routes
  if (to.meta.public) {
    // If already authenticated, redirect away from login/register
    if (isAuthenticated.value) return { path: '/' }
    return
  }

  // Protected routes: redirect to login if not authenticated
  if (!isAuthenticated.value) {
    return { path: '/login', query: { redirect: to.fullPath } }
  }
})
