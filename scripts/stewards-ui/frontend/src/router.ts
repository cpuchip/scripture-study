import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'

// Lazy-loaded views — keeps the initial bundle small as more pages
// land. Phase 1: only Dashboard exists; the others render the
// Placeholder view via dynamic resolution. As phases land,
// replace each with the real view import.
const Dashboard = () => import('./views/Dashboard.vue')
const Placeholder = () => import('./views/Placeholder.vue')

const routes: RouteRecordRaw[] = [
  { path: '/',           name: 'dashboard',  component: Dashboard },
  { path: '/studies',    name: 'studies',    component: Placeholder, meta: { title: 'Studies' } },
  { path: '/studies/:slug', name: 'study-detail', component: Placeholder, meta: { title: 'Study detail' }, props: true },
  { path: '/work-items', name: 'work-items', component: Placeholder, meta: { title: 'Work items' } },
  { path: '/work-items/:id', name: 'work-item-detail', component: Placeholder, meta: { title: 'Work item detail' }, props: true },
  { path: '/sessions',   name: 'sessions',   component: Placeholder, meta: { title: 'Sessions' } },
  { path: '/watchman',   name: 'watchman',   component: Placeholder, meta: { title: 'Watchman' } },
  { path: '/bridge',     name: 'bridge',     component: Placeholder, meta: { title: 'Bridge state' } },
  { path: '/graph',      name: 'graph',      component: Placeholder, meta: { title: 'Graph' } },
  { path: '/new',        name: 'new-work',   component: Placeholder, meta: { title: 'New work' } },
]

export default createRouter({
  history: createWebHistory(),
  routes,
})
