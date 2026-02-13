<script setup lang="ts">
import { computed } from 'vue'
import { useAuth } from './composables/useAuth'
import { useRouter } from 'vue-router'

const { user, isAuthenticated, logout } = useAuth()
const router = useRouter()

const siteName = computed(() => {
  const host = window.location.hostname
  if (host.includes('webeco')) return 'We Become'
  return 'I Become'
})

async function handleLogout() {
  await logout()
  router.push('/login')
}
</script>

<template>
  <div class="min-h-screen bg-gray-50 text-gray-900">
    <!-- Navigation (only when authenticated) -->
    <nav v-if="isAuthenticated" class="bg-white border-b border-gray-200 px-4 py-3">
      <div class="max-w-4xl mx-auto flex items-center justify-between">
        <router-link to="/" class="text-xl font-bold text-orange-600 flex items-center gap-1.5">
          <img src="/ibecome-icon.png" alt="" class="h-7 w-7" />
          {{ siteName }}
        </router-link>
        <div class="flex items-center gap-4 text-sm">
          <router-link to="/" class="hover:text-orange-600" active-class="text-orange-600 font-semibold">Today</router-link>
          <router-link to="/memorize" class="hover:text-orange-600" active-class="text-orange-600 font-semibold">Memorize</router-link>
          <router-link to="/practices" class="hover:text-orange-600" active-class="text-orange-600 font-semibold">Practices</router-link>
          <router-link to="/reports" class="hover:text-orange-600" active-class="text-orange-600 font-semibold">Reports</router-link>
          <router-link to="/pillars" class="hover:text-orange-600" active-class="text-orange-600 font-semibold">Pillars</router-link>
          <router-link to="/notes" class="hover:text-orange-600" active-class="text-orange-600 font-semibold">Notes</router-link>
          <router-link to="/reflections" class="hover:text-orange-600" active-class="text-orange-600 font-semibold">Reflect</router-link>
          <router-link to="/tasks" class="hover:text-orange-600" active-class="text-orange-600 font-semibold">Tasks</router-link>
          <span class="border-l border-gray-300 pl-4 flex items-center gap-2">
            <router-link to="/settings" class="text-gray-600 hover:text-orange-600" title="Settings">{{ user?.name }}</router-link>
            <button
              @click="handleLogout"
              class="text-gray-400 hover:text-red-500 transition-colors"
              title="Sign out"
            >
              <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
                <path fill-rule="evenodd" d="M3 3a1 1 0 00-1 1v12a1 1 0 001 1h6a1 1 0 100-2H4V5h5a1 1 0 100-2H3zm11.707 4.293a1 1 0 010 1.414L13.414 10l1.293 1.293a1 1 0 01-1.414 1.414l-2-2a1 1 0 010-1.414l2-2a1 1 0 011.414 0z" clip-rule="evenodd" />
                <path fill-rule="evenodd" d="M10 10a1 1 0 011-1h5a1 1 0 110 2h-5a1 1 0 01-1-1z" clip-rule="evenodd" />
              </svg>
            </button>
          </span>
        </div>
      </div>
    </nav>

    <!-- Content -->
    <main :class="isAuthenticated ? 'max-w-4xl mx-auto px-4 py-6' : ''">
      <router-view />
    </main>
  </div>
</template>
