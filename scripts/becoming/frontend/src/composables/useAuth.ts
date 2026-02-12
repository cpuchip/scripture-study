import { ref, computed, readonly } from 'vue'
import { authApi, setUnauthorizedHandler, type User } from '../api'

const user = ref<User | null>(null)
const loading = ref(true)
const initialized = ref(false)

export function useAuth() {
  const isAuthenticated = computed(() => !!user.value)

  async function init() {
    if (initialized.value) return
    initialized.value = true
    loading.value = true

    // Register global 401 handler
    setUnauthorizedHandler(() => {
      user.value = null
    })

    try {
      user.value = await authApi.me()
    } catch {
      user.value = null
    } finally {
      loading.value = false
    }
  }

  async function login(email: string, password: string) {
    const res = await authApi.login(email, password)
    user.value = res.user
    return res.user
  }

  async function register(email: string, password: string, name?: string) {
    const res = await authApi.register(email, password, name)
    user.value = res.user
    return res.user
  }

  async function logout() {
    try {
      await authApi.logout()
    } finally {
      user.value = null
    }
  }

  async function refresh() {
    try {
      user.value = await authApi.me()
    } catch {
      user.value = null
    }
  }

  return {
    user: readonly(user),
    loading: readonly(loading),
    isAuthenticated,
    init,
    login,
    register,
    logout,
    refresh,
  }
}
