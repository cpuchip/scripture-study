import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'

const Home = () => import('./views/Home.vue')
const WordSearch = () => import('./views/WordSearch.vue')
const VerseExplorer = () => import('./views/VerseExplorer.vue')
const WordDetail = () => import('./views/WordDetail.vue')
const About = () => import('./views/About.vue')

const routes: RouteRecordRaw[] = [
  { path: '/', name: 'home', component: Home },
  { path: '/word', name: 'word-search', component: WordSearch },
  { path: '/word/:word', name: 'word-detail', component: WordDetail, props: true },
  { path: '/verse', name: 'verse-explorer', component: VerseExplorer },
  { path: '/about', name: 'about', component: About },
]

export default createRouter({
  history: createWebHistory(),
  routes,
})
