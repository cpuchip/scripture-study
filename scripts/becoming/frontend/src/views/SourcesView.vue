<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api, type DocumentSource } from '../api'
import { github } from '../services/github'

const sources = ref<DocumentSource[]>([])
const loading = ref(true)
const showForm = ref(false)
const editingId = ref<number | null>(null)
const testResult = ref<{ success: boolean; count: number; error?: string } | null>(null)
const testing = ref(false)

const form = ref({
  name: '',
  repo: '',
  branch: 'main',
  source_type: 'github_public' as DocumentSource['source_type'],
  include_paths: '',  // comma-separated globs
  exclude_paths: '',
})

async function loadSources() {
  loading.value = true
  try {
    sources.value = await api.listSources()
  } catch (e) {
    console.error('Failed to load sources:', e)
  } finally {
    loading.value = false
  }
}

function resetForm() {
  form.value = {
    name: '',
    repo: '',
    branch: 'main',
    source_type: 'github_public',
    include_paths: '',
    exclude_paths: '',
  }
  editingId.value = null
  testResult.value = null
}

function startEdit(source: DocumentSource) {
  editingId.value = source.id
  const inc = JSON.parse(source.include_paths || '[]') as string[]
  const exc = JSON.parse(source.exclude_paths || '[]') as string[]
  form.value = {
    name: source.name,
    repo: source.repo,
    branch: source.branch,
    source_type: source.source_type,
    include_paths: inc.join(', '),
    exclude_paths: exc.join(', '),
  }
  showForm.value = true
  testResult.value = null
}

function parseGlobs(input: string): string[] {
  return input
    .split(',')
    .map(s => s.trim())
    .filter(s => s.length > 0)
}

async function testConnection() {
  testing.value = true
  testResult.value = null
  try {
    const include = parseGlobs(form.value.include_paths)
    const exclude = parseGlobs(form.value.exclude_paths)
    const entries = await github.getTree(form.value.repo, form.value.branch, include, exclude)
    testResult.value = { success: true, count: entries.length }
  } catch (e: any) {
    testResult.value = { success: false, count: 0, error: e.message }
  } finally {
    testing.value = false
  }
}

async function saveSource() {
  const payload: Partial<DocumentSource> = {
    name: form.value.name,
    repo: form.value.repo,
    branch: form.value.branch,
    source_type: form.value.source_type,
    include_paths: JSON.stringify(parseGlobs(form.value.include_paths)),
    exclude_paths: JSON.stringify(parseGlobs(form.value.exclude_paths)),
  }

  try {
    if (editingId.value) {
      await api.updateSource(editingId.value, payload)
    } else {
      await api.createSource(payload)
    }
    showForm.value = false
    resetForm()
    await loadSources()
  } catch (e: any) {
    alert('Failed to save: ' + e.message)
  }
}

async function deleteSource(id: number) {
  if (!confirm('Remove this document source? Reading progress will also be deleted.')) return
  try {
    await api.deleteSource(id)
    await loadSources()
  } catch (e: any) {
    alert('Failed to delete: ' + e.message)
  }
}

function formatGlobs(jsonStr: string): string {
  try {
    const arr = JSON.parse(jsonStr) as string[]
    return arr.length ? arr.join(', ') : '(all .md files)'
  } catch {
    return jsonStr
  }
}

onMounted(loadSources)
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold">Document Sources</h1>
      <button
        v-if="!showForm"
        @click="showForm = true; resetForm()"
        class="bg-orange-500 text-white px-4 py-2 rounded hover:bg-orange-600 text-sm"
      >
        + Add Source
      </button>
    </div>

    <!-- Form -->
    <div v-if="showForm" class="bg-white rounded-lg border border-gray-200 p-6 mb-6">
      <h2 class="text-lg font-semibold mb-4">
        {{ editingId ? 'Edit Source' : 'Add Document Source' }}
      </h2>

      <div class="space-y-4">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Name</label>
          <input
            v-model="form.name"
            type="text"
            placeholder="My Studies"
            class="w-full border border-gray-300 rounded px-3 py-2 text-sm"
          />
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">GitHub Repository</label>
          <input
            v-model="form.repo"
            type="text"
            placeholder="owner/repo"
            class="w-full border border-gray-300 rounded px-3 py-2 text-sm font-mono"
          />
          <p class="text-xs text-gray-500 mt-1">Public GitHub repo in owner/repo format</p>
        </div>

        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Branch</label>
            <input
              v-model="form.branch"
              type="text"
              placeholder="main"
              class="w-full border border-gray-300 rounded px-3 py-2 text-sm font-mono"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Type</label>
            <select
              v-model="form.source_type"
              class="w-full border border-gray-300 rounded px-3 py-2 text-sm"
            >
              <option value="github_public">Public</option>
              <option value="github_private">Private (requires PAT)</option>
            </select>
          </div>
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Include Paths</label>
          <input
            v-model="form.include_paths"
            type="text"
            placeholder="public/study/**/*.md, public/lessons/**/*.md"
            class="w-full border border-gray-300 rounded px-3 py-2 text-sm font-mono"
          />
          <p class="text-xs text-gray-500 mt-1">
            Comma-separated glob patterns. Empty = all .md files.
            Use <code>**</code> for recursive matching.
          </p>
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Exclude Paths</label>
          <input
            v-model="form.exclude_paths"
            type="text"
            placeholder="**/README.md, **/_*"
            class="w-full border border-gray-300 rounded px-3 py-2 text-sm font-mono"
          />
        </div>

        <!-- Test Connection -->
        <div class="flex items-center gap-3">
          <button
            @click="testConnection"
            :disabled="!form.repo || testing"
            class="bg-gray-100 text-gray-700 px-4 py-2 rounded text-sm hover:bg-gray-200 disabled:opacity-50"
          >
            {{ testing ? 'Testing...' : 'Test Connection' }}
          </button>
          <div v-if="testResult" class="text-sm">
            <span v-if="testResult.success" class="text-green-600">
              ✓ Found {{ testResult.count }} markdown file{{ testResult.count === 1 ? '' : 's' }}
            </span>
            <span v-else class="text-red-600">
              ✗ {{ testResult.error }}
            </span>
          </div>
        </div>
      </div>

      <div class="flex gap-2 mt-6">
        <button
          @click="saveSource"
          :disabled="!form.name || !form.repo"
          class="bg-orange-500 text-white px-4 py-2 rounded hover:bg-orange-600 text-sm disabled:opacity-50"
        >
          {{ editingId ? 'Update' : 'Add Source' }}
        </button>
        <button
          @click="showForm = false; resetForm()"
          class="bg-gray-100 text-gray-700 px-4 py-2 rounded hover:bg-gray-200 text-sm"
        >
          Cancel
        </button>
      </div>
    </div>

    <!-- Sources List -->
    <div v-if="loading" class="text-gray-500 text-center py-8">Loading...</div>

    <div v-else-if="sources.length === 0 && !showForm" class="text-center py-12">
      <p class="text-gray-400 text-lg mb-2">No document sources configured</p>
      <p class="text-gray-400 text-sm mb-4">Add a GitHub repo to start reading study documents</p>
      <button
        @click="showForm = true; resetForm()"
        class="bg-orange-500 text-white px-4 py-2 rounded hover:bg-orange-600 text-sm"
      >
        + Add Your First Source
      </button>
    </div>

    <div v-else class="space-y-3">
      <div
        v-for="source in sources"
        :key="source.id"
        class="bg-white rounded-lg border border-gray-200 p-4"
      >
        <div class="flex items-start justify-between">
          <div>
            <h3 class="font-semibold text-gray-900">{{ source.name }}</h3>
            <p class="text-sm text-gray-500 font-mono mt-0.5">
              {{ source.repo }}
              <span v-if="source.branch !== 'main'" class="text-orange-600">@{{ source.branch }}</span>
            </p>
            <div class="mt-2 space-y-1">
              <p class="text-xs text-gray-400">
                <span class="font-medium">Include:</span>
                {{ formatGlobs(source.include_paths) }}
              </p>
              <p v-if="JSON.parse(source.exclude_paths || '[]').length > 0" class="text-xs text-gray-400">
                <span class="font-medium">Exclude:</span>
                {{ formatGlobs(source.exclude_paths) }}
              </p>
            </div>
          </div>
          <div class="flex items-center gap-2">
            <router-link
              :to="{ name: 'reader', params: { id: source.id } }"
              class="bg-orange-50 text-orange-600 px-3 py-1.5 rounded text-sm hover:bg-orange-100"
            >
              Read →
            </router-link>
            <button
              @click="startEdit(source)"
              class="text-gray-400 hover:text-gray-600 text-sm"
            >
              Edit
            </button>
            <button
              @click="deleteSource(source.id)"
              class="text-gray-400 hover:text-red-500 text-sm"
            >
              Delete
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
