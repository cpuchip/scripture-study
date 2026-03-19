import { createApp } from 'vue'
import './style.css'
import App from './App.vue'
import { router } from './router'

createApp(App).use(router).mount('#app')

// Register service worker for push notifications
if ('serviceWorker' in navigator) {
  navigator.serviceWorker.register('/sw.js').catch((err) => {
    console.warn('Service worker registration failed:', err)
  })
}
