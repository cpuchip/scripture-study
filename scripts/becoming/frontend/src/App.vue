<script setup lang="ts">
import { computed, ref } from 'vue'
import { useAuth } from './composables/useAuth'
import { useTheme } from './composables/useTheme'
import { useRouter } from 'vue-router'

const { user, isAuthenticated, logout } = useAuth()
const { darkMode, toggleDarkMode } = useTheme()
const router = useRouter()
const mobileMenuOpen = ref(false)

const siteName = computed(() => {
  const host = window.location.hostname
  if (host.includes('webeco')) return 'We Become'
  return 'I Become'
})

async function handleLogout() {
  await logout()
  router.push('/login')
}

// Close mobile menu on route change
router.afterEach(() => { mobileMenuOpen.value = false })
</script>

<template>
  <div class="min-h-screen bg-gray-50 text-gray-900">
    <!-- Navigation (only when authenticated) -->
    <nav v-if="isAuthenticated" class="bg-white border-b border-gray-200 px-4 py-3">
      <div class="max-w-4xl mx-auto flex items-center justify-between">
        <router-link to="/today" class="text-xl font-bold text-orange-600 flex items-center gap-1.5">
          <img src="/ibecome-icon.png" alt="" class="h-7 w-7" />
          {{ siteName }}
        </router-link>

        <!-- Desktop nav links -->
        <div class="hidden md:flex items-center gap-4 text-sm">
          <router-link to="/today" class="hover:text-orange-600" active-class="text-orange-600 font-semibold">Today</router-link>
          <router-link to="/memorize" class="hover:text-orange-600" active-class="text-orange-600 font-semibold">Memorize</router-link>
          <router-link to="/practices" class="hover:text-orange-600" active-class="text-orange-600 font-semibold">Practices</router-link>
          <router-link to="/reports" class="hover:text-orange-600" active-class="text-orange-600 font-semibold">Reports</router-link>
          <router-link to="/pillars" class="hover:text-orange-600" active-class="text-orange-600 font-semibold">Pillars</router-link>
          <router-link to="/notes" class="hover:text-orange-600" active-class="text-orange-600 font-semibold">Notes</router-link>
          <router-link to="/reflections" class="hover:text-orange-600" active-class="text-orange-600 font-semibold">Reflect</router-link>
          <router-link to="/sources" class="hover:text-orange-600" active-class="text-orange-600 font-semibold">Library</router-link>
          <router-link to="/tasks" class="hover:text-orange-600" active-class="text-orange-600 font-semibold">Tasks</router-link>
          <span class="border-l border-gray-300 pl-4 flex items-center gap-2">
            <button
              @click="toggleDarkMode"
              class="text-gray-400 hover:text-orange-500 transition-colors"
              :title="darkMode ? 'Light mode' : 'Dark mode'"
            >
              <svg v-if="darkMode" xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-yellow-400" viewBox="0 0 20 20" fill="currentColor">
                <path fill-rule="evenodd" d="M10 2a1 1 0 011 1v1a1 1 0 11-2 0V3a1 1 0 011-1zm4 8a4 4 0 11-8 0 4 4 0 018 0zm-.464 4.95l.707.707a1 1 0 001.414-1.414l-.707-.707a1 1 0 00-1.414 1.414zm2.12-10.607a1 1 0 010 1.414l-.706.707a1 1 0 11-1.414-1.414l.707-.707a1 1 0 011.414 0zM17 11a1 1 0 100-2h-1a1 1 0 100 2h1zm-7 4a1 1 0 011 1v1a1 1 0 11-2 0v-1a1 1 0 011-1zM5.05 6.464A1 1 0 106.465 5.05l-.708-.707a1 1 0 00-1.414 1.414l.707.707zm1.414 8.486l-.707.707a1 1 0 01-1.414-1.414l.707-.707a1 1 0 011.414 1.414zM4 11a1 1 0 100-2H3a1 1 0 000 2h1z" clip-rule="evenodd" />
              </svg>
              <svg v-else xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
                <path d="M17.293 13.293A8 8 0 016.707 2.707a8.001 8.001 0 1010.586 10.586z" />
              </svg>
            </button>
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

        <!-- Mobile: theme toggle + hamburger -->
        <div class="flex items-center gap-2 md:hidden">
          <button
            @click="toggleDarkMode"
            class="text-gray-400 hover:text-orange-500 transition-colors p-1"
            :title="darkMode ? 'Light mode' : 'Dark mode'"
          >
            <svg v-if="darkMode" xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-yellow-400" viewBox="0 0 20 20" fill="currentColor">
              <path fill-rule="evenodd" d="M10 2a1 1 0 011 1v1a1 1 0 11-2 0V3a1 1 0 011-1zm4 8a4 4 0 11-8 0 4 4 0 018 0zm-.464 4.95l.707.707a1 1 0 001.414-1.414l-.707-.707a1 1 0 00-1.414 1.414zm2.12-10.607a1 1 0 010 1.414l-.706.707a1 1 0 11-1.414-1.414l.707-.707a1 1 0 011.414 0zM17 11a1 1 0 100-2h-1a1 1 0 100 2h1zm-7 4a1 1 0 011 1v1a1 1 0 11-2 0v-1a1 1 0 011-1zM5.05 6.464A1 1 0 106.465 5.05l-.708-.707a1 1 0 00-1.414 1.414l.707.707zm1.414 8.486l-.707.707a1 1 0 01-1.414-1.414l.707-.707a1 1 0 011.414 1.414zM4 11a1 1 0 100-2H3a1 1 0 000 2h1z" clip-rule="evenodd" />
            </svg>
            <svg v-else xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
              <path d="M17.293 13.293A8 8 0 016.707 2.707a8.001 8.001 0 1010.586 10.586z" />
            </svg>
          </button>
          <button @click="mobileMenuOpen = !mobileMenuOpen" class="text-gray-600 hover:text-orange-600 p-1" title="Menu">
            <!-- Hamburger / X icon -->
            <svg v-if="!mobileMenuOpen" xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M4 6h16M4 12h16M4 18h16" />
            </svg>
            <svg v-else xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
      </div>

      <!-- Mobile dropdown -->
      <div v-if="mobileMenuOpen" class="md:hidden mt-3 pt-3 border-t border-gray-200 space-y-1">
        <router-link to="/today" class="block px-2 py-2 rounded text-sm hover:bg-gray-100" active-class="text-orange-600 font-semibold bg-orange-50">Today</router-link>
        <router-link to="/memorize" class="block px-2 py-2 rounded text-sm hover:bg-gray-100" active-class="text-orange-600 font-semibold bg-orange-50">Memorize</router-link>
        <router-link to="/practices" class="block px-2 py-2 rounded text-sm hover:bg-gray-100" active-class="text-orange-600 font-semibold bg-orange-50">Practices</router-link>
        <router-link to="/reports" class="block px-2 py-2 rounded text-sm hover:bg-gray-100" active-class="text-orange-600 font-semibold bg-orange-50">Reports</router-link>
        <router-link to="/pillars" class="block px-2 py-2 rounded text-sm hover:bg-gray-100" active-class="text-orange-600 font-semibold bg-orange-50">Pillars</router-link>
        <router-link to="/notes" class="block px-2 py-2 rounded text-sm hover:bg-gray-100" active-class="text-orange-600 font-semibold bg-orange-50">Notes</router-link>
        <router-link to="/reflections" class="block px-2 py-2 rounded text-sm hover:bg-gray-100" active-class="text-orange-600 font-semibold bg-orange-50">Reflect</router-link>
        <router-link to="/sources" class="block px-2 py-2 rounded text-sm hover:bg-gray-100" active-class="text-orange-600 font-semibold bg-orange-50">Library</router-link>
        <router-link to="/tasks" class="block px-2 py-2 rounded text-sm hover:bg-gray-100" active-class="text-orange-600 font-semibold bg-orange-50">Tasks</router-link>
        <div class="flex items-center justify-between pt-2 mt-2 border-t border-gray-200 px-2">
          <router-link to="/settings" class="text-sm text-gray-600 hover:text-orange-600">{{ user?.name }}</router-link>
          <button @click="handleLogout" class="text-sm text-gray-400 hover:text-red-500" title="Sign out">Sign out</button>
        </div>
      </div>
    </nav>

    <!-- Content -->
    <main :class="isAuthenticated && !['reader', 'public-reader', 'short-link'].includes($route.name as string) ? 'max-w-4xl mx-auto px-4 py-6' : ''">
      <router-view />
    </main>
  </div>
</template>

<style>
/* ============================================
   Global Dark Mode Overrides
   Applied when useTheme sets .dark-mode on <html>
   ============================================ */

html.dark-mode {
  background-color: #111827;
  color-scheme: dark;
}

/* --- Backgrounds --- */
html.dark-mode .bg-white { background-color: #1f2937 !important; }
html.dark-mode .bg-gray-50 { background-color: #111827 !important; }
html.dark-mode .bg-gray-100 { background-color: #1f2937 !important; }
html.dark-mode .bg-gray-200 { background-color: #374151 !important; }
html.dark-mode .hover\:bg-gray-50:hover { background-color: #1f2937 !important; }
html.dark-mode .hover\:bg-gray-100:hover { background-color: #374151 !important; }
html.dark-mode .hover\:bg-gray-200:hover { background-color: #4b5563 !important; }

/* --- Text --- */
html.dark-mode .text-gray-900 { color: #f3f4f6 !important; }
html.dark-mode .text-gray-800 { color: #e5e7eb !important; }
html.dark-mode .text-gray-700 { color: #d1d5db !important; }
html.dark-mode .text-gray-600 { color: #9ca3af !important; }
html.dark-mode .text-gray-500 { color: #6b7280 !important; }
html.dark-mode .text-gray-400 { color: #6b7280 !important; }
html.dark-mode .text-gray-300 { color: #9ca3af !important; }

/* --- Borders --- */
html.dark-mode .border-gray-200 { border-color: #374151 !important; }
html.dark-mode .border-gray-300 { border-color: #4b5563 !important; }
html.dark-mode .border-gray-100 { border-color: #1f2937 !important; }
html.dark-mode .divide-gray-100 > :not([hidden]) ~ :not([hidden]) { border-color: #374151 !important; }

/* --- Nav bar --- */
html.dark-mode nav.bg-white {
  background-color: #1a2332 !important;
  border-color: #334155 !important;
}
html.dark-mode nav a { color: #cbd5e1; }
html.dark-mode nav a:hover,
html.dark-mode nav a.text-orange-600 { color: #f59e0b !important; }
html.dark-mode nav .text-xl { color: #f59e0b !important; }

/* --- Accent backgrounds (soften for dark) --- */
html.dark-mode .bg-orange-50 { background-color: rgba(234, 88, 12, 0.1) !important; }
html.dark-mode .bg-orange-100 { background-color: rgba(234, 88, 12, 0.15) !important; }
html.dark-mode .bg-amber-50 { background-color: rgba(245, 158, 11, 0.1) !important; }
html.dark-mode .bg-amber-100 { background-color: rgba(245, 158, 11, 0.15) !important; }
html.dark-mode .bg-green-50 { background-color: rgba(34, 197, 94, 0.1) !important; }
html.dark-mode .bg-green-100 { background-color: rgba(34, 197, 94, 0.15) !important; }
html.dark-mode .bg-red-50 { background-color: rgba(239, 68, 68, 0.1) !important; }
html.dark-mode .bg-red-100 { background-color: rgba(239, 68, 68, 0.15) !important; }
html.dark-mode .bg-indigo-50 { background-color: rgba(99, 102, 241, 0.1) !important; }
html.dark-mode .bg-indigo-100 { background-color: rgba(99, 102, 241, 0.15) !important; }
html.dark-mode .bg-purple-50 { background-color: rgba(168, 85, 247, 0.1) !important; }
html.dark-mode .bg-purple-100 { background-color: rgba(168, 85, 247, 0.15) !important; }
html.dark-mode .bg-blue-50 { background-color: rgba(59, 130, 246, 0.1) !important; }

/* --- Accent borders (soften) --- */
html.dark-mode .border-amber-200 { border-color: rgba(245, 158, 11, 0.3) !important; }
html.dark-mode .border-green-200 { border-color: rgba(34, 197, 94, 0.3) !important; }
html.dark-mode .border-red-200,
html.dark-mode .border-red-300 { border-color: rgba(239, 68, 68, 0.3) !important; }
html.dark-mode .border-indigo-200,
html.dark-mode .border-indigo-300 { border-color: rgba(99, 102, 241, 0.3) !important; }
html.dark-mode .border-indigo-400 { border-color: rgba(99, 102, 241, 0.4) !important; }

/* --- Accent text (lighten for readability) --- */
html.dark-mode .text-amber-600 { color: #fbbf24 !important; }
html.dark-mode .text-amber-700 { color: #fcd34d !important; }
html.dark-mode .text-green-600 { color: #4ade80 !important; }
html.dark-mode .text-green-700 { color: #86efac !important; }
html.dark-mode .text-green-800 { color: #86efac !important; }
html.dark-mode .text-red-400 { color: #f87171 !important; }
html.dark-mode .text-red-500 { color: #f87171 !important; }
html.dark-mode .text-red-600 { color: #fca5a5 !important; }
html.dark-mode .text-red-700 { color: #fca5a5 !important; }
html.dark-mode .text-indigo-500 { color: #a5b4fc !important; }
html.dark-mode .text-indigo-600 { color: #a5b4fc !important; }
html.dark-mode .text-indigo-700 { color: #c7d2fe !important; }
html.dark-mode .text-orange-600 { color: #fb923c !important; }
html.dark-mode .text-orange-700 { color: #fdba74 !important; }
html.dark-mode .text-purple-600 { color: #c084fc !important; }
html.dark-mode .text-blue-600 { color: #93c5fd !important; }
html.dark-mode .text-rose-600 { color: #fda4af !important; }

/* --- Form inputs --- */
html.dark-mode input,
html.dark-mode textarea,
html.dark-mode select {
  background-color: #374151 !important;
  color: #f3f4f6 !important;
  border-color: #4b5563 !important;
}
html.dark-mode input::placeholder,
html.dark-mode textarea::placeholder {
  color: #6b7280 !important;
}
html.dark-mode input:focus,
html.dark-mode textarea:focus,
html.dark-mode select:focus {
  border-color: #f59e0b !important;
  box-shadow: 0 0 0 2px rgba(245, 158, 11, 0.15) !important;
}

/* --- Buttons --- */
html.dark-mode .bg-indigo-600 { background-color: #4f46e5 !important; }
html.dark-mode .bg-orange-500 { background-color: #ea580c !important; }
html.dark-mode .bg-orange-600 { background-color: #ea580c !important; }
html.dark-mode .bg-green-600 { background-color: #16a34a !important; }
html.dark-mode .bg-red-600 { background-color: #dc2626 !important; }

/* --- Gradient override (landing page) --- */
html.dark-mode .bg-gradient-to-b { background: #111827 !important; }

/* --- Ring colors --- */
html.dark-mode .ring-indigo-400 { --tw-ring-color: rgba(99, 102, 241, 0.4) !important; }
html.dark-mode .ring-offset-1 { --tw-ring-offset-color: #1f2937 !important; }

/* --- Scrollbar --- */
html.dark-mode ::-webkit-scrollbar { width: 8px; }
html.dark-mode ::-webkit-scrollbar-track { background: #1f2937; }
html.dark-mode ::-webkit-scrollbar-thumb { background: #4b5563; border-radius: 4px; }
html.dark-mode ::-webkit-scrollbar-thumb:hover { background: #6b7280; }
</style>
