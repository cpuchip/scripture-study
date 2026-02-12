<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAuth } from '../composables/useAuth'
import { authApi, type APIToken } from '../api'

const { user, refresh } = useAuth()

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

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  })
}

onMounted(loadTokens)
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
  </div>
</template>
