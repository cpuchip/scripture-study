<script setup lang="ts">
import { ref, computed, onMounted, nextTick, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { publicApi, api, type SharedLink } from '../api'
import { github, type FileTreeNode } from '../services/github'
import { useAuth } from '../composables/useAuth'
import TreeNode from '../components/TreeNode.vue'
import MarkdownIt from 'markdown-it'

const route = useRoute()
const router = useRouter()
const { isAuthenticated } = useAuth()

// Resolve parameters from query string or short code
const provider = ref('gh')
const repo = ref('')
const branch = ref('main')
const docFilter = ref('**/*.md')
const initialFile = ref('')
const shortCode = ref('')
const resolvedLink = ref<SharedLink | null>(null)

const fileTree = ref<FileTreeNode[]>([])
const loading = ref(true)
const loadingContent = ref(false)
const error = ref('')

// Current document state
const currentPath = ref('')
const currentContent = ref('')
const currentTitle = ref('')

// Sidebar state
const sidebarOpen = ref(true)
const expandedDirs = ref<Set<string>>(new Set())
const searchQuery = ref('')

// Save to library
const showSaveModal = ref(false)
const saveName = ref('')
const saving = ref(false)
const saved = ref(false)

// Share link display
const shareUrl = ref('')

// Markdown renderer
const md = new MarkdownIt({
  html: true,
  linkify: true,
  typographer: true,
  breaks: false,
})

const defaultRender = md.renderer.rules.link_open || function(tokens: any, idx: any, options: any, _env: any, self: any) {
  return self.renderToken(tokens, idx, options)
}
md.renderer.rules.link_open = function(tokens: any, idx: any, options: any, env: any, self: any) {
  const href = tokens[idx].attrGet('href')
  if (href && (href.startsWith('http://') || href.startsWith('https://'))) {
    tokens[idx].attrSet('target', '_blank')
    tokens[idx].attrSet('rel', 'noopener noreferrer')
  }
  return defaultRender(tokens, idx, options, env, self)
}

const renderedContent = computed(() => {
  if (!currentContent.value) return ''
  return md.render(currentContent.value)
})

const filteredTree = computed(() => {
  if (!searchQuery.value) return fileTree.value
  const q = searchQuery.value.toLowerCase()
  return filterNodes(fileTree.value, q)
})

const repoDisplayName = computed(() => {
  if (!repo.value) return ''
  const parts = repo.value.split('/')
  return parts[parts.length - 1] || repo.value
})

const sourceName = computed(() => {
  return resolvedLink.value ? `${repoDisplayName.value}` : repoDisplayName.value
})

function filterNodes(nodes: FileTreeNode[], query: string): FileTreeNode[] {
  const result: FileTreeNode[] = []
  for (const node of nodes) {
    if (node.type === 'file') {
      if (node.name.toLowerCase().includes(query) || node.path.toLowerCase().includes(query)) {
        result.push(node)
      }
    } else if (node.children) {
      const filtered = filterNodes(node.children, query)
      if (filtered.length > 0) {
        result.push({ ...node, children: filtered })
      }
    }
  }
  return result
}

function extractTitle(content: string): string {
  const match = content.match(/^#\s+(.+)$/m)
  return match?.[1] ?? ''
}

function titleFromPath(path: string): string {
  const parts = path.split('/')
  const fileName = parts[parts.length - 1] || path
  return fileName.replace(/\.md$/i, '').replace(/[-_]/g, ' ').replace(/^\d+\s*/, '')
}

async function resolveParams() {
  // Check for short code first: /s/:code
  const code = route.params.code as string
  if (code) {
    shortCode.value = code
    try {
      const link = await publicApi.resolveShareLink(code)
      resolvedLink.value = link
      repo.value = link.repo
      branch.value = link.branch
      docFilter.value = link.doc_filter
      if (link.file_path) initialFile.value = link.file_path
    } catch (e: any) {
      error.value = `Link not found: ${e.message}`
      loading.value = false
      return
    }
  } else {
    // Read from query params: /share?p=gh&r=owner/repo&b=main&d=glob&f=file.md
    repo.value = (route.query.r as string) || ''
    branch.value = (route.query.b as string) || 'main'
    docFilter.value = (route.query.d as string) || '**/*.md'
    initialFile.value = (route.query.f as string) || ''
    provider.value = (route.query.p as string) || 'gh'
  }

  if (!repo.value) {
    error.value = 'No repository specified'
    loading.value = false
    return
  }

  await loadTree()
}

async function loadTree() {
  loading.value = true
  error.value = ''
  try {
    const includes = docFilter.value ? [docFilter.value] : []
    const entries = await github.getTree(repo.value, branch.value, includes, [])
    fileTree.value = github.buildFileTree(entries)

    // Auto-expand first-level directories
    for (const node of fileTree.value) {
      if (node.type === 'directory') {
        expandedDirs.value.add(node.path)
      }
    }

    // Open initial file if specified
    if (initialFile.value) {
      await openFile(initialFile.value)
    }
  } catch (e: any) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

async function openFile(path: string) {
  loadingContent.value = true
  try {
    const content = await github.getContent(repo.value, branch.value, path)
    currentPath.value = path
    currentContent.value = content
    currentTitle.value = extractTitle(content) || titleFromPath(path)

    // Update URL with file param (without full page reload)
    const query = { ...route.query, f: path }
    router.replace({ query })

    await nextTick()
    const contentEl = document.getElementById('public-reader-content')
    if (contentEl) contentEl.scrollTop = 0
  } catch (e: any) {
    error.value = `Failed to load ${path}: ${e.message}`
  } finally {
    loadingContent.value = false
  }
}

function toggleDir(path: string) {
  if (expandedDirs.value.has(path)) {
    expandedDirs.value.delete(path)
  } else {
    expandedDirs.value.add(path)
  }
}

function toggleSidebar() {
  sidebarOpen.value = !sidebarOpen.value
}

function handleContentClick(event: MouseEvent) {
  const target = event.target as HTMLElement
  const link = target.closest('a')
  if (!link) return

  const href = link.getAttribute('href')
  if (!href) return

  if (href.startsWith('http://') || href.startsWith('https://')) return

  if (href.endsWith('.md') || href.includes('.md#')) {
    event.preventDefault()
    const mdPath = href.split('#')[0] ?? href
    const resolvedPath = resolvePath(currentPath.value, mdPath)
    openFile(resolvedPath)
  }
}

function resolvePath(currentFile: string, relativePath: string): string {
  const currentDir = currentFile.substring(0, currentFile.lastIndexOf('/'))
  const parts = currentDir.split('/').filter(Boolean)
  const relParts = relativePath.split('/')
  for (const part of relParts) {
    if (part === '..') parts.pop()
    else if (part !== '.') parts.push(part)
  }
  return parts.join('/')
}

// --- Save to Library ---

function openSaveModal() {
  saveName.value = repoDisplayName.value || 'Shared Study'
  showSaveModal.value = true
}

async function saveToLibrary() {
  saving.value = true
  try {
    await api.createSource({
      name: saveName.value,
      source_type: 'github_public',
      repo: repo.value,
      branch: branch.value,
      include_paths: JSON.stringify(docFilter.value ? [docFilter.value] : []),
      exclude_paths: '[]',
    })
    saved.value = true
    showSaveModal.value = false
  } catch (e: any) {
    error.value = `Failed to save: ${e.message}`
  } finally {
    saving.value = false
  }
}

function goToRegister() {
  // Pass share URL as redirect so after registration they can save
  const currentUrl = window.location.pathname + window.location.search
  router.push({ path: '/register', query: { redirect: currentUrl } })
}

// Build the full share URL for display
function buildShareUrl(): string {
  const base = window.location.origin
  if (shortCode.value) {
    return `${base}/s/${shortCode.value}`
  }
  const params = new URLSearchParams()
  params.set('p', provider.value)
  params.set('r', repo.value)
  if (branch.value !== 'main') params.set('b', branch.value)
  if (docFilter.value !== '**/*.md') params.set('d', docFilter.value)
  if (currentPath.value) params.set('f', currentPath.value)
  return `${base}/share?${params.toString()}`
}

async function copyShareUrl() {
  shareUrl.value = buildShareUrl()
  try {
    await navigator.clipboard.writeText(shareUrl.value)
  } catch {
    // Fallback: select text
  }
}

onMounted(resolveParams)

// Update page title when document changes
watch(currentTitle, (title) => {
  if (title) {
    document.title = `${title} — Shared Study`
  } else {
    document.title = `${repoDisplayName.value || 'Shared Study'}`
  }
})
</script>

<template>
  <div class="public-reader">
    <!-- Minimal header -->
    <header class="public-header">
      <div class="flex items-center gap-3">
        <router-link to="/" class="text-orange-600 font-bold text-sm flex items-center gap-1">
          <img src="/ibecome-icon.png" alt="" class="h-5 w-5" />
          I Become
        </router-link>
        <span class="text-gray-300">|</span>
        <span class="text-sm text-gray-600 font-medium">{{ sourceName }}</span>
      </div>
      <div class="flex items-center gap-2">
        <!-- Copy share link -->
        <button
          @click="copyShareUrl"
          class="text-xs text-gray-500 hover:text-orange-600 flex items-center gap-1 px-2 py-1 rounded hover:bg-gray-100"
          title="Copy share link"
        >
          <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" viewBox="0 0 20 20" fill="currentColor">
            <path d="M12.586 4.586a2 2 0 112.828 2.828l-3 3a2 2 0 01-2.828 0 1 1 0 00-1.414 1.414 4 4 0 005.656 0l3-3a4 4 0 00-5.656-5.656l-1.5 1.5a1 1 0 101.414 1.414l1.5-1.5zm-5 5a2 2 0 012.828 0 1 1 0 101.414-1.414 4 4 0 00-5.656 0l-3 3a4 4 0 105.656 5.656l1.5-1.5a1 1 0 10-1.414-1.414l-1.5 1.5a2 2 0 11-2.828-2.828l3-3z" />
          </svg>
          Share
        </button>

        <!-- Save to Library -->
        <template v-if="isAuthenticated">
          <button
            v-if="!saved"
            @click="openSaveModal"
            class="text-xs bg-orange-600 text-white px-3 py-1 rounded hover:bg-orange-700"
          >
            + Save to Library
          </button>
          <span v-else class="text-xs text-green-600 flex items-center gap-1">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" viewBox="0 0 20 20" fill="currentColor">
              <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
            </svg>
            Saved
          </span>
        </template>
        <template v-else>
          <button
            @click="goToRegister"
            class="text-xs bg-orange-600 text-white px-3 py-1 rounded hover:bg-orange-700"
          >
            Sign up to save
          </button>
        </template>
      </div>
    </header>

    <!-- Error state -->
    <div v-if="error && !currentContent && !loading" class="flex items-center justify-center h-[calc(100vh-48px)]">
      <div class="text-center">
        <p class="text-red-500 text-lg mb-2">{{ error }}</p>
        <router-link to="/" class="text-orange-600 hover:underline text-sm">Go to I Become</router-link>
      </div>
    </div>

    <!-- Reader layout -->
    <div v-else class="reader-layout">
      <!-- Sidebar Toggle -->
      <button
        v-if="!sidebarOpen"
        @click="toggleSidebar"
        class="fixed left-2 top-16 z-10 bg-white border border-gray-200 rounded-lg p-2 shadow-sm hover:bg-gray-50"
        title="Show sidebar"
      >
        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-gray-600" viewBox="0 0 20 20" fill="currentColor">
          <path fill-rule="evenodd" d="M3 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm0 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm0 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1z" clip-rule="evenodd" />
        </svg>
      </button>

      <!-- Sidebar -->
      <aside v-if="sidebarOpen" class="reader-sidebar">
        <div class="flex items-center justify-between px-3 py-2 border-b border-gray-200">
          <h2 class="text-sm font-semibold text-gray-700 truncate">{{ sourceName }}</h2>
          <button @click="toggleSidebar" class="text-gray-400 hover:text-gray-600 p-1" title="Hide sidebar">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
              <path fill-rule="evenodd" d="M12.707 5.293a1 1 0 010 1.414L9.414 10l3.293 3.293a1 1 0 01-1.414 1.414l-4-4a1 1 0 010-1.414l4-4a1 1 0 011.414 0z" clip-rule="evenodd" />
            </svg>
          </button>
        </div>

        <div class="px-3 py-2 border-b border-gray-100">
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Filter files..."
            class="w-full border border-gray-200 rounded px-2 py-1 text-xs"
          />
        </div>

        <div class="overflow-y-auto flex-1 py-1">
          <div v-if="loading" class="px-3 py-4 text-xs text-gray-400">Loading tree...</div>
          <template v-else>
            <TreeNode
              v-for="node in filteredTree"
              :key="node.path"
              :node="node"
              :depth="0"
              :expanded-dirs="expandedDirs"
              :current-path="currentPath"
              :read-paths="new Set()"
              @toggle-dir="toggleDir"
              @open-file="openFile"
            />
          </template>
        </div>

        <div class="border-t border-gray-100 px-3 py-2">
          <a
            :href="`https://github.com/${repo}`"
            target="_blank"
            rel="noopener noreferrer"
            class="text-xs text-gray-400 hover:text-gray-600 flex items-center gap-1"
          >
            <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" viewBox="0 0 16 16" fill="currentColor">
              <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z"/>
            </svg>
            {{ repo }}
          </a>
        </div>
      </aside>

      <!-- Main Content -->
      <main class="reader-content" id="public-reader-content" @click="handleContentClick">
        <div v-if="!currentContent && !loadingContent && !loading" class="flex items-center justify-center h-full">
          <div class="text-center text-gray-400">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-12 w-12 mx-auto mb-3 opacity-50" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" />
            </svg>
            <p class="text-lg">Select a document to read</p>
            <p v-if="!sidebarOpen" class="text-sm mt-1">Open the sidebar to browse files</p>
          </div>
        </div>

        <div v-else-if="loadingContent" class="flex items-center justify-center h-full">
          <div class="text-gray-400">Loading...</div>
        </div>

        <div v-else-if="currentContent" class="reader-document">
          <div class="text-xs text-gray-400 mb-4 font-mono">{{ currentPath }}</div>
          <article class="prose prose-gray max-w-none" v-html="renderedContent" />
        </div>
      </main>
    </div>

    <!-- Save to Library modal -->
    <div v-if="showSaveModal" class="fixed inset-0 bg-black/40 flex items-center justify-center z-50" @click.self="showSaveModal = false">
      <div class="bg-white rounded-lg p-6 w-96 shadow-xl">
        <h3 class="text-lg font-semibold mb-4">Save to Library</h3>
        <p class="text-sm text-gray-600 mb-3">Add this study source to your personal library for reading progress tracking.</p>
        <div class="mb-4">
          <label class="block text-sm font-medium text-gray-700 mb-1">Name</label>
          <input v-model="saveName" type="text" class="w-full border border-gray-300 rounded px-3 py-2 text-sm" />
        </div>
        <div class="text-xs text-gray-400 mb-4">
          <span class="font-mono">{{ repo }}</span> · {{ branch }} · {{ docFilter }}
        </div>
        <div class="flex justify-end gap-2">
          <button @click="showSaveModal = false" class="px-4 py-2 text-sm text-gray-600 hover:text-gray-900">Cancel</button>
          <button
            @click="saveToLibrary"
            :disabled="saving || !saveName"
            class="px-4 py-2 text-sm bg-orange-600 text-white rounded hover:bg-orange-700 disabled:opacity-50"
          >
            {{ saving ? 'Saving...' : 'Save' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.public-reader {
  min-height: 100vh;
  background: #fafafa;
}

.public-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 16px;
  background: white;
  border-bottom: 1px solid #e5e7eb;
  height: 48px;
}

.reader-layout {
  display: flex;
  height: calc(100vh - 48px);
  width: 100%;
}

.reader-sidebar {
  width: 280px;
  min-width: 280px;
  border-right: 1px solid #e5e7eb;
  background: white;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.reader-content {
  flex: 1;
  overflow-y: auto;
  padding: 24px 32px;
  background: #fafafa;
}

.reader-document {
  max-width: 800px;
  margin: 0 auto;
  background: white;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  padding: 32px 40px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
}

/* Same prose styles as ReaderView */
.reader-document :deep(h1) { font-size: 1.75rem; font-weight: 700; margin-bottom: 1rem; color: #111827; border-bottom: 1px solid #e5e7eb; padding-bottom: 0.5rem; }
.reader-document :deep(h2) { font-size: 1.35rem; font-weight: 600; margin-top: 2rem; margin-bottom: 0.75rem; color: #1f2937; }
.reader-document :deep(h3) { font-size: 1.1rem; font-weight: 600; margin-top: 1.5rem; margin-bottom: 0.5rem; color: #374151; }
.reader-document :deep(p) { line-height: 1.75; margin-bottom: 1rem; color: #374151; }
.reader-document :deep(blockquote) { border-left: 3px solid #d97706; padding: 0.5rem 1rem; margin: 1rem 0; background: #fffbeb; border-radius: 0 4px 4px 0; }
.reader-document :deep(blockquote p) { color: #92400e; margin-bottom: 0.25rem; }
.reader-document :deep(a) { color: #ea580c; text-decoration: underline; text-decoration-color: #fed7aa; text-underline-offset: 2px; }
.reader-document :deep(a:hover) { color: #c2410c; text-decoration-color: #ea580c; }
.reader-document :deep(code) { background: #f3f4f6; padding: 0.15rem 0.35rem; border-radius: 3px; font-size: 0.875em; color: #1f2937; }
.reader-document :deep(pre) { background: #1f2937; color: #e5e7eb; padding: 1rem; border-radius: 6px; overflow-x: auto; margin: 1rem 0; }
.reader-document :deep(pre code) { background: none; color: inherit; padding: 0; }
.reader-document :deep(ul), .reader-document :deep(ol) { padding-left: 1.5rem; margin-bottom: 1rem; }
.reader-document :deep(li) { line-height: 1.75; margin-bottom: 0.25rem; color: #374151; }
.reader-document :deep(table) { width: 100%; border-collapse: collapse; margin: 1rem 0; }
.reader-document :deep(th), .reader-document :deep(td) { border: 1px solid #e5e7eb; padding: 0.5rem 0.75rem; text-align: left; font-size: 0.875rem; }
.reader-document :deep(th) { background: #f9fafb; font-weight: 600; }
.reader-document :deep(hr) { border: none; border-top: 1px solid #e5e7eb; margin: 2rem 0; }
.reader-document :deep(img) { max-width: 100%; border-radius: 6px; }
.reader-document :deep(h2:has(+ p > strong)), .reader-document :deep(h2:last-of-type) { color: #b45309; }
</style>
