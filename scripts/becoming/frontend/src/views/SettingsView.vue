<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuth } from '../composables/useAuth'
import { authApi, type APIToken, type SessionInfo } from '../api'

const router = useRouter()
const { user, refresh, logout } = useAuth()

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

    <!-- Change Password Section (email users only) -->
    <section v-if="user?.provider === 'email'" class="bg-white rounded-lg border border-gray-200 p-6">
      <div class="flex items-center justify-between mb-4">
        <div>
          <h2 class="text-lg font-semibold">Password</h2>
          <p class="text-sm text-gray-500 mt-1">Change your account password.</p>
        </div>
        <button
          v-if="!showPasswordForm"
          @click="showPasswordForm = true"
          class="text-sm text-indigo-600 hover:text-indigo-800"
        >
          Change Password
        </button>
      </div>

      <div v-if="showPasswordForm" class="space-y-3 max-w-sm">
        <div v-if="passwordMessage" class="p-3 bg-green-50 border border-green-200 rounded text-sm text-green-800">
          {{ passwordMessage }}
        </div>
        <div v-if="passwordError" class="p-3 bg-red-50 border border-red-200 rounded text-sm text-red-800">
          {{ passwordError }}
        </div>
        <div>
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
            :disabled="changingPassword || !currentPassword || !newPassword || !confirmPassword"
            class="bg-indigo-600 text-white text-sm px-4 py-2 rounded hover:bg-indigo-700 disabled:opacity-50"
          >
            {{ changingPassword ? 'Changing...' : 'Change Password' }}
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
        <div v-if="user?.provider === 'email'">
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
            :disabled="deleting || deleteConfirmText !== 'DELETE MY ACCOUNT' || (user?.provider === 'email' && !deletePassword)"
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
