<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuth } from '../composables/useAuth'
import { authApi, type APIToken, type AuthProviders, type SessionInfo } from '../api'

const router = useRouter()
const { user, refresh, logout } = useAuth()

// Auth providers (is Google available?)
const providers = ref<AuthProviders | null>(null)

async function loadProviders() {
  try {
    providers.value = await authApi.providers()
  } catch { /* ignore */ }
}

// Google linking
const unlinkingGoogle = ref(false)

async function unlinkGoogle() {
  if (!confirm('Unlink your Google account? You will only be able to log in with email/password.')) return
  unlinkingGoogle.value = true
  try {
    await authApi.unlinkGoogle()
    await refresh()
  } catch (e: any) {
    alert(e.message)
  } finally {
    unlinkingGoogle.value = false
  }
}

// Profile editing
const editingName = ref(false)
const nameInput = ref('')
const savingName = ref(false)

function startEditName() {
  nameInput.value = user.value?.name || ''
  editingName.value = true
}

async function saveName() {
  if (!nameInput.value.trim()) return
  savingName.value = true
  try {
    await authApi.updateMe(nameInput.value.trim())
    await refresh()
    editingName.value = false
  } catch (e: any) {
    alert(e.message)
  } finally {
    savingName.value = false
  }
}

// Password change
const showPasswordForm = ref(false)
const currentPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const changingPassword = ref(false)
const passwordMessage = ref('')
const passwordError = ref('')

async function changePassword() {
  passwordError.value = ''
  passwordMessage.value = ''

  if (newPassword.value.length < 8) {
    passwordError.value = 'New password must be at least 8 characters'
    return
  }
  if (newPassword.value !== confirmPassword.value) {
    passwordError.value = 'Passwords do not match'
    return
  }

  changingPassword.value = true
  try {
    await authApi.changePassword(currentPassword.value, newPassword.value)
    passwordMessage.value = 'Password changed successfully. Other sessions have been logged out.'
    currentPassword.value = ''
    newPassword.value = ''
    confirmPassword.value = ''
    showPasswordForm.value = false
    loadSessions() // refresh session list
  } catch (e: any) {
    passwordError.value = e.message
  } finally {
    changingPassword.value = false
  }
}

// API Tokens
const tokens = ref<APIToken[]>([])
const loadingTokens = ref(true)
const showCreateForm = ref(false)
const newTokenName = ref('')
const creatingToken = ref(false)
const newlyCreatedToken = ref<string | null>(null)
const copiedToken = ref(false)

async function loadTokens() {
  loadingTokens.value = true
  try {
    tokens.value = await authApi.listTokens()
  } catch (e: any) {
    console.error('Failed to load tokens:', e)
  } finally {
    loadingTokens.value = false
  }
}

async function createToken() {
  if (!newTokenName.value.trim()) return
  creatingToken.value = true
  try {
    const result = await authApi.createToken(newTokenName.value.trim())
    newlyCreatedToken.value = result.token
    newTokenName.value = ''
    showCreateForm.value = false
    await loadTokens()
  } catch (e: any) {
    alert(e.message)
  } finally {
    creatingToken.value = false
  }
}

async function copyToken() {
  if (!newlyCreatedToken.value) return
  await navigator.clipboard.writeText(newlyCreatedToken.value)
  copiedToken.value = true
  setTimeout(() => (copiedToken.value = false), 2000)
}

function dismissToken() {
  newlyCreatedToken.value = null
  copiedToken.value = false
}

async function revokeToken(id: number, name: string) {
  if (!confirm(`Revoke token "${name}"? This cannot be undone.`)) return
  try {
    await authApi.deleteToken(id)
    await loadTokens()
  } catch (e: any) {
    alert(e.message)
  }
}

// Sessions
const sessions = ref<SessionInfo[]>([])
const loadingSessions = ref(true)

async function loadSessions() {
  loadingSessions.value = true
  try {
    sessions.value = await authApi.listSessions()
  } catch (e: any) {
    console.error('Failed to load sessions:', e)
  } finally {
    loadingSessions.value = false
  }
}

async function revokeSession(id: string) {
  if (!confirm('Revoke this session? The device will be logged out.')) return
  try {
    await authApi.revokeSession(id)
    await loadSessions()
  } catch (e: any) {
    alert(e.message)
  }
}

async function revokeOtherSessions() {
  if (!confirm('Revoke all other sessions? All other devices will be logged out.')) return
  try {
    await authApi.revokeOtherSessions()
    await loadSessions()
  } catch (e: any) {
    alert(e.message)
  }
}

function parseUserAgent(ua: string): string {
  if (!ua) return 'Unknown device'
  // Simple browser detection
  if (ua.includes('Chrome') && !ua.includes('Edg')) return 'Chrome'
  if (ua.includes('Edg')) return 'Edge'
  if (ua.includes('Firefox')) return 'Firefox'
  if (ua.includes('Safari') && !ua.includes('Chrome')) return 'Safari'
  if (ua.includes('curl')) return 'curl'
  return ua.substring(0, 40)
}

function formatRelative(dateStr: string): string {
  const date = new Date(dateStr)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const minutes = Math.floor(diff / 60000)
  if (minutes < 1) return 'just now'
  if (minutes < 60) return `${minutes}m ago`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours}h ago`
  const days = Math.floor(hours / 24)
  if (days < 30) return `${days}d ago`
  return formatDate(dateStr)
}

// Data export
const exporting = ref(false)

async function exportData() {
  exporting.value = true
  try {
    await authApi.exportData()
  } catch (e: any) {
    alert(e.message)
  } finally {
    exporting.value = false
  }
}

// Delete account
const showDeleteConfirm = ref(false)
const deletePassword = ref('')
const deleteConfirmText = ref('')
const deleting = ref(false)

async function deleteAccount() {
  if (deleteConfirmText.value !== 'DELETE MY ACCOUNT') return
  deleting.value = true
  try {
    await authApi.deleteAccount(deletePassword.value)
    await logout()
    router.push('/login')
  } catch (e: any) {
    alert(e.message)
  } finally {
    deleting.value = false
  }
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  })
}

onMounted(() => {
  loadProviders()
  loadTokens()
  loadSessions()
})
</script>

<template>
  <div class="space-y-8">
    <h1 class="text-2xl font-bold">Settings</h1>

    <!-- Profile Section -->
    <section class="bg-white rounded-lg border border-gray-200 p-6">
      <h2 class="text-lg font-semibold mb-4">Profile</h2>
      <div class="space-y-3">
        <div class="flex items-center justify-between">
          <div>
            <span class="text-sm text-gray-500">Email</span>
            <p class="font-medium">{{ user?.email }}</p>
          </div>
        </div>
        <div class="flex items-center justify-between">
          <div v-if="!editingName">
            <span class="text-sm text-gray-500">Name</span>
            <p class="font-medium">{{ user?.name || 'Not set' }}</p>
          </div>
          <div v-else class="flex-1 mr-4">
            <label class="text-sm text-gray-500">Name</label>
            <input
              v-model="nameInput"
              @keyup.enter="saveName"
              @keyup.escape="editingName = false"
              class="mt-1 w-full border border-gray-300 rounded px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
              placeholder="Your name"
              autofocus
            />
          </div>
          <div>
            <button
              v-if="!editingName"
              @click="startEditName"
              class="text-sm text-indigo-600 hover:text-indigo-800"
            >
              Edit
            </button>
            <div v-else class="flex gap-2">
              <button
                @click="saveName"
                :disabled="savingName"
                class="text-sm bg-indigo-600 text-white px-3 py-1 rounded hover:bg-indigo-700 disabled:opacity-50"
              >
                Save
              </button>
              <button
                @click="editingName = false"
                class="text-sm text-gray-500 hover:text-gray-700"
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
        <div>
          <span class="text-sm text-gray-500">Member since</span>
          <p class="font-medium">{{ user ? formatDate(user.created_at) : '' }}</p>
        </div>
      </div>
    </section>

    <!-- Password Section -->
    <section class="bg-white rounded-lg border border-gray-200 p-6">
      <div class="flex items-center justify-between mb-4">
        <div>
          <h2 class="text-lg font-semibold">Password</h2>
          <p v-if="user?.has_password" class="text-sm text-gray-500 mt-1">Change your account password.</p>
          <p v-else class="text-sm text-gray-500 mt-1">Set a password so you can also log in with email.</p>
        </div>
        <button
          v-if="!showPasswordForm"
          @click="showPasswordForm = true"
          class="text-sm text-indigo-600 hover:text-indigo-800"
        >
          {{ user?.has_password ? 'Change Password' : 'Set Password' }}
        </button>
      </div>

      <div v-if="showPasswordForm" class="space-y-3 max-w-sm">
        <div v-if="passwordMessage" class="p-3 bg-green-50 border border-green-200 rounded text-sm text-green-800">
          {{ passwordMessage }}
        </div>
        <div v-if="passwordError" class="p-3 bg-red-50 border border-red-200 rounded text-sm text-red-800">
          {{ passwordError }}
        </div>
        <div v-if="user?.has_password">
          <label class="block text-sm font-medium text-gray-700 mb-1">Current Password</label>
          <input
            v-model="currentPassword"
            type="password"
            class="w-full border border-gray-300 rounded px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">New Password</label>
          <input
            v-model="newPassword"
            type="password"
            class="w-full border border-gray-300 rounded px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
            placeholder="At least 8 characters"
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Confirm New Password</label>
          <input
            v-model="confirmPassword"
            type="password"
            @keyup.enter="changePassword"
            class="w-full border border-gray-300 rounded px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
          />
        </div>
        <div class="flex gap-2 pt-1">
          <button
            @click="changePassword"
            :disabled="changingPassword || (user?.has_password && !currentPassword) || !newPassword || !confirmPassword"
            class="bg-indigo-600 text-white text-sm px-4 py-2 rounded hover:bg-indigo-700 disabled:opacity-50"
          >
            {{ changingPassword ? 'Saving...' : (user?.has_password ? 'Change Password' : 'Set Password') }}
          </button>
          <button
            @click="showPasswordForm = false; passwordError = ''; passwordMessage = ''"
            class="text-sm text-gray-500 hover:text-gray-700 px-3 py-2"
          >
            Cancel
          </button>
        </div>
      </div>
    </section>

    <!-- Connected Accounts Section -->
    <section class="bg-white rounded-lg border border-gray-200 p-6">
      <h2 class="text-lg font-semibold mb-4">Connected Accounts</h2>
      <div class="space-y-3">
        <!-- Google -->
        <div class="flex items-center justify-between py-2">
          <div class="flex items-center gap-3">
            <svg class="w-5 h-5" viewBox="0 0 24 24"><path d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92a5.06 5.06 0 0 1-2.2 3.32v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.1z" fill="#4285F4"/><path d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" fill="#34A853"/><path d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z" fill="#FBBC05"/><path d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" fill="#EA4335"/></svg>
            <div>
              <p class="text-sm font-medium">Google</p>
              <p v-if="user?.google_linked" class="text-xs text-green-600">Connected</p>
              <p v-else class="text-xs text-gray-400">Not connected</p>
            </div>
          </div>
          <div>
            <button
              v-if="user?.google_linked && user?.has_password"
              @click="unlinkGoogle"
              :disabled="unlinkingGoogle"
              class="text-sm text-red-600 hover:text-red-800 disabled:opacity-50"
            >
              {{ unlinkingGoogle ? 'Unlinking...' : 'Unlink' }}
            </button>
            <a
              v-else-if="!user?.google_linked && providers?.google"
              :href="'/auth/google/login'"
              class="text-sm text-indigo-600 hover:text-indigo-800"
            >
              Link Google Account
            </a>
            <span v-else-if="user?.google_linked && !user?.has_password" class="text-xs text-gray-400">
              Set a password first to unlink
            </span>
          </div>
        </div>
      </div>
    </section>

    <!-- API Tokens Section -->
    <section class="bg-white rounded-lg border border-gray-200 p-6">
      <div class="flex items-center justify-between mb-4">
        <div>
          <h2 class="text-lg font-semibold">API Tokens</h2>
          <p class="text-sm text-gray-500 mt-1">
            Use tokens to access the Becoming API from scripts or the MCP server.
          </p>
        </div>
        <button
          v-if="!showCreateForm"
          @click="showCreateForm = true"
          class="bg-indigo-600 text-white text-sm px-4 py-2 rounded hover:bg-indigo-700"
        >
          New Token
        </button>
      </div>

      <!-- Newly created token banner -->
      <div
        v-if="newlyCreatedToken"
        class="mb-4 p-4 bg-green-50 border border-green-200 rounded-lg"
      >
        <p class="text-sm text-green-800 font-medium mb-2">
          Token created! Copy it now — you won't be able to see it again.
        </p>
        <div class="flex items-center gap-2">
          <code class="flex-1 bg-white border border-green-300 rounded px-3 py-2 text-sm font-mono break-all select-all">
            {{ newlyCreatedToken }}
          </code>
          <button
            @click="copyToken"
            class="shrink-0 bg-green-600 text-white text-sm px-3 py-2 rounded hover:bg-green-700"
          >
            {{ copiedToken ? 'Copied!' : 'Copy' }}
          </button>
        </div>
        <button
          @click="dismissToken"
          class="mt-2 text-sm text-green-600 hover:text-green-800"
        >
          I've saved it, dismiss
        </button>
      </div>

      <!-- Create form -->
      <div v-if="showCreateForm" class="mb-4 p-4 bg-gray-50 rounded-lg border border-gray-200">
        <label class="block text-sm font-medium text-gray-700 mb-1">Token name</label>
        <div class="flex gap-2">
          <input
            v-model="newTokenName"
            @keyup.enter="createToken"
            @keyup.escape="showCreateForm = false"
            class="flex-1 border border-gray-300 rounded px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
            placeholder="e.g., MCP Server, Automation Script"
            autofocus
          />
          <button
            @click="createToken"
            :disabled="creatingToken || !newTokenName.trim()"
            class="bg-indigo-600 text-white text-sm px-4 py-1.5 rounded hover:bg-indigo-700 disabled:opacity-50"
          >
            Create
          </button>
          <button
            @click="showCreateForm = false"
            class="text-sm text-gray-500 hover:text-gray-700"
          >
            Cancel
          </button>
        </div>
      </div>

      <!-- Token list -->
      <div v-if="loadingTokens" class="text-sm text-gray-500">Loading tokens...</div>
      <div v-else-if="tokens.length === 0" class="text-sm text-gray-500">
        No API tokens yet. Create one to get started.
      </div>
      <div v-else class="divide-y divide-gray-100">
        <div
          v-for="token in tokens"
          :key="token.id"
          class="flex items-center justify-between py-3"
        >
          <div>
            <p class="font-medium text-sm">{{ token.name }}</p>
            <p class="text-xs text-gray-500">
              <code class="bg-gray-100 px-1.5 py-0.5 rounded">{{ token.prefix }}...</code>
              &middot; Created {{ formatDate(token.created_at) }}
              <template v-if="token.last_used">
                &middot; Last used {{ formatDate(token.last_used) }}
              </template>
            </p>
          </div>
          <button
            @click="revokeToken(token.id, token.name)"
            class="text-sm text-red-500 hover:text-red-700"
          >
            Revoke
          </button>
        </div>
      </div>
    </section>

    <!-- Active Sessions -->
    <section class="bg-white rounded-lg border border-gray-200 p-6">
      <div class="flex items-center justify-between mb-4">
        <div>
          <h2 class="text-lg font-semibold">Active Sessions</h2>
          <p class="text-sm text-gray-500 mt-1">Devices where you're currently logged in.</p>
        </div>
        <button
          v-if="sessions.filter(s => !s.is_current).length > 0"
          @click="revokeOtherSessions"
          class="text-sm text-red-500 hover:text-red-700"
        >
          Revoke All Others
        </button>
      </div>

      <div v-if="loadingSessions" class="text-sm text-gray-500">Loading sessions...</div>
      <div v-else-if="sessions.length === 0" class="text-sm text-gray-500">No active sessions.</div>
      <div v-else class="divide-y divide-gray-100">
        <div
          v-for="session in sessions"
          :key="session.id"
          class="flex items-center justify-between py-3"
        >
          <div>
            <p class="font-medium text-sm">
              {{ parseUserAgent(session.user_agent) }}
              <span v-if="session.is_current" class="ml-2 text-xs bg-green-100 text-green-800 px-2 py-0.5 rounded-full">
                This device
              </span>
            </p>
            <p class="text-xs text-gray-500">
              {{ session.ip_address }} &middot; Active {{ formatRelative(session.last_active) }}
              &middot; Created {{ formatDate(session.created_at) }}
            </p>
          </div>
          <button
            v-if="!session.is_current"
            @click="revokeSession(session.id.replace('...', ''))"
            class="text-sm text-red-500 hover:text-red-700"
          >
            Revoke
          </button>
        </div>
      </div>
    </section>

    <!-- Data Export -->
    <section class="bg-white rounded-lg border border-gray-200 p-6">
      <div class="flex items-center justify-between">
        <div>
          <h2 class="text-lg font-semibold">Export Data</h2>
          <p class="text-sm text-gray-500 mt-1">
            Download all your data (practices, logs, tasks, notes, reflections) as JSON.
          </p>
        </div>
        <button
          @click="exportData"
          :disabled="exporting"
          class="bg-gray-100 text-gray-700 text-sm px-4 py-2 rounded hover:bg-gray-200 disabled:opacity-50"
        >
          {{ exporting ? 'Exporting...' : 'Download JSON' }}
        </button>
      </div>
    </section>

    <!-- Danger Zone -->
    <section class="bg-white rounded-lg border-2 border-red-200 p-6">
      <h2 class="text-lg font-semibold text-red-700 mb-4">Danger Zone</h2>

      <div v-if="!showDeleteConfirm">
        <p class="text-sm text-gray-600 mb-3">
          Permanently delete your account and all associated data. This action cannot be undone.
        </p>
        <button
          @click="showDeleteConfirm = true"
          class="bg-red-600 text-white text-sm px-4 py-2 rounded hover:bg-red-700"
        >
          Delete Account
        </button>
      </div>

      <div v-else class="space-y-3 max-w-sm">
        <p class="text-sm text-red-800 font-medium">
          This will permanently delete your account, all practices, logs, tasks, notes, reflections, and pillars. This cannot be undone.
        </p>
        <div v-if="user?.has_password">
          <label class="block text-sm font-medium text-gray-700 mb-1">Enter your password</label>
          <input
            v-model="deletePassword"
            type="password"
            class="w-full border border-red-300 rounded px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-red-500"
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">
            Type <code class="bg-red-50 px-1 rounded">DELETE MY ACCOUNT</code> to confirm
          </label>
          <input
            v-model="deleteConfirmText"
            class="w-full border border-red-300 rounded px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-red-500"
            placeholder="DELETE MY ACCOUNT"
          />
        </div>
        <div class="flex gap-2 pt-1">
          <button
            @click="deleteAccount"
            :disabled="deleting || deleteConfirmText !== 'DELETE MY ACCOUNT' || (user?.has_password && !deletePassword)"
            class="bg-red-600 text-white text-sm px-4 py-2 rounded hover:bg-red-700 disabled:opacity-50"
          >
            {{ deleting ? 'Deleting...' : 'Permanently Delete Account' }}
          </button>
          <button
            @click="showDeleteConfirm = false; deletePassword = ''; deleteConfirmText = ''"
            class="text-sm text-gray-500 hover:text-gray-700 px-3 py-2"
          >
            Cancel
          </button>
        </div>
      </div>
    </section>
  </div>
</template>
