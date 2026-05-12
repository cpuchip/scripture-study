import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'

// Lazy-loaded views — keeps the initial bundle small as more pages
// land. Phase 1: only Dashboard exists; the others render the
// Placeholder view via dynamic resolution. As phases land,
// replace each with the real view import.
const Dashboard = () => import('./views/Dashboard.vue')
const Studies = () => import('./views/Studies.vue')
const StudyDetail = () => import('./views/StudyDetail.vue')
const WorkItems = () => import('./views/WorkItems.vue')
const WorkItemDetail = () => import('./views/WorkItemDetail.vue')
const Sessions = () => import('./views/Sessions.vue')
const Watchman = () => import('./views/Watchman.vue')
const BridgeState = () => import('./views/BridgeState.vue')
const NewWork = () => import('./views/NewWork.vue')
const Graph = () => import('./views/Graph.vue')
const Intents = () => import('./views/Intents.vue')
const Covenants = () => import('./views/Covenants.vue')
const Sabbath = () => import('./views/Sabbath.vue')
const Lessons = () => import('./views/Lessons.vue')
const Trust = () => import('./views/Trust.vue')
const Councils = () => import('./views/Councils.vue')
const CouncilDetail = () => import('./views/CouncilDetail.vue')
const Projects = () => import('./views/Projects.vue')

const routes: RouteRecordRaw[] = [
  { path: '/',           name: 'dashboard',  component: Dashboard },
  { path: '/studies',    name: 'studies',    component: Studies, meta: { title: 'Studies' } },
  { path: '/studies/:slug', name: 'study-detail', component: StudyDetail, meta: { title: 'Study detail' }, props: true },
  { path: '/work-items', name: 'work-items', component: WorkItems, meta: { title: 'Work items' } },
  { path: '/work-items/:id', name: 'work-item-detail', component: WorkItemDetail, meta: { title: 'Work item detail' }, props: true },
  { path: '/sessions',   name: 'sessions',   component: Sessions, meta: { title: 'Sessions' } },
  { path: '/sessions/:sid', name: 'session-detail', component: Sessions, meta: { title: 'Session' }, props: true },
  { path: '/watchman',   name: 'watchman',   component: Watchman, meta: { title: 'Watchman' } },
  { path: '/bridge',     name: 'bridge',     component: BridgeState, meta: { title: 'Bridge state' } },
  { path: '/graph',      name: 'graph',      component: Graph, meta: { title: 'Graph' } },
  { path: '/new',        name: 'new-work',   component: NewWork, meta: { title: 'New work' } },
  { path: '/intents',    name: 'intents',    component: Intents, meta: { title: 'Intents' } },
  { path: '/covenants',  name: 'covenants',  component: Covenants, meta: { title: 'Covenant' } },
  { path: '/sabbath',    name: 'sabbath',    component: Sabbath, meta: { title: 'Sabbath log' } },
  { path: '/lessons',    name: 'lessons',    component: Lessons, meta: { title: 'Lessons' } },
  { path: '/trust',      name: 'trust',      component: Trust, meta: { title: 'Trust matrix' } },
  { path: '/councils',   name: 'councils',   component: Councils, meta: { title: 'Councils' } },
  { path: '/councils/:id', name: 'council-detail', component: CouncilDetail, meta: { title: 'Council' }, props: true },
  { path: '/projects',   name: 'projects',   component: Projects, meta: { title: 'Projects' } },
]

export default createRouter({
  history: createWebHistory(),
  routes,
})
