import { createRouter, createWebHashHistory } from 'vue-router'
import Home from './pages/Home.vue'
import Registration from './pages/Registration.vue'
import Schedule from './pages/Schedule.vue'
import Score from './pages/Score.vue'
import Display from './pages/Display.vue'
import Results from './pages/Results.vue'

export const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    { path: '/', component: Home },
    { path: '/race/:id/registration', component: Registration, props: true },
    { path: '/race/:id/schedule', component: Schedule, props: true },
    { path: '/race/:id/score', component: Score, props: true },
    { path: '/race/:id/score/:heat', component: Score, props: true },
    { path: '/race/:id/display', component: Display, props: true },
    { path: '/race/:id/results', component: Results, props: true },
  ]
})
