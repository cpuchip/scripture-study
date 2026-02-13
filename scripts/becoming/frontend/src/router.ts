import { createRouter, createWebHistory } from 'vue-router'
import DailyView from './views/DailyView.vue'
import { useAuth } from './composables/useAuth'

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    // Public routes (no auth required)
    { path: '/', name: 'landing', component: () => import('./views/LandingView.vue'), meta: { public: true } },
    { path: '/login', name: 'login', component: () => import('./views/LoginView.vue'), meta: { public: true } },
    { path: '/register', name: 'register', component: () => import('./views/RegisterView.vue'), meta: { public: true } },
    { path: '/privacy', name: 'privacy', component: () => import('./views/PrivacyView.vue'), meta: { public: true } },
    { path: '/terms', name: 'terms', component: () => import('./views/TermsView.vue'), meta: { public: true } },

    // Protected routes
    { path: '/today', name: 'daily', component: DailyView },
    { path: '/onboarding', name: 'onboarding', component: () => import('./views/OnboardingView.vue') },
    { path: '/practices', name: 'practices', component: () => import('./views/PracticesView.vue') },
    { path: '/practices/:id/history', name: 'history', component: () => import('./views/HistoryView.vue') },
    { path: '/memorize', name: 'memorize', component: () => import('./views/MemorizeView.vue') },
    { path: '/tasks', name: 'tasks', component: () => import('./views/TasksView.vue') },
    { path: '/notes', name: 'notes', component: () => import('./views/NotesView.vue') },
    { path: '/reflections', name: 'reflections', component: () => import('./views/ReflectionsView.vue') },
    { path: '/pillars', name: 'pillars', component: () => import('./views/PillarsView.vue') },
    { path: '/reports', name: 'reports', component: () => import('./views/ReportsView.vue') },
    { path: '/settings', name: 'settings', component: () => import('./views/SettingsView.vue') },
  ],
})

// Dynamic page title based on hostname + route
function getSitePrefix(): string {
  const host = window.location.hostname
  if (host.includes('webeco')) return 'We Become'
  return 'I Become'
}

const routeTitles: Record<string, string> = {
  landing: '',
  daily: 'Today',
  login: 'Login',
  register: 'Register',
  privacy: 'Privacy Policy',
  terms: 'Terms of Service',
  onboarding: 'Get Started',
  practices: 'Practices',
  history: 'History',
  memorize: 'Memorize',
  tasks: 'Tasks',
  notes: 'Notes',
  reflections: 'Reflect',
  pillars: 'Pillars',
  reports: 'Reports',
  settings: 'Settings',
}

router.afterEach((to) => {
  const prefix = getSitePrefix()
  const page = routeTitles[to.name as string] || ''
  document.title = page ? `${prefix} — ${page}` : prefix
})

// Auth guard — redirect to /login if not authenticated, /onboarding if new user
let onboardingChecked = false

router.beforeEach(async (to) => {
  const { isAuthenticated, loading, init } = useAuth()

  // Ensure auth is initialized (calls /api/me once)
  await init()

  // Wait for loading to finish
  if (loading.value) return

  // Allow public routes
  if (to.meta.public) {
    // If already authenticated, redirect away from login/register/landing
    if (isAuthenticated.value && (to.name === 'login' || to.name === 'register' || to.name === 'landing')) {
      return { path: '/today' }
    }
    return
  }

  // Protected routes: redirect to landing if not authenticated
  if (!isAuthenticated.value) {
    return { path: '/', query: { redirect: to.fullPath } }
  }

  // Onboarding redirect: check once per session for new users
  if (!onboardingChecked && to.name !== 'onboarding') {
    onboardingChecked = true

    // Skip if already completed onboarding
    if (localStorage.getItem('onboarding_complete') === 'true') return

    try {
      const has = await fetch('/api/pillars/has-pillars', { credentials: 'same-origin' })
      if (has.ok) {
        const data = await has.json()
        if (!data.has_pillars) {
          return { path: '/onboarding' }
        }
        // Has pillars — mark onboarding complete
        localStorage.setItem('onboarding_complete', 'true')
      }
    } catch {
      // If check fails, don't block navigation
    }
  }
})
