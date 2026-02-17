<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue'
import { useRoute } from 'vue-router'
import { api, publicApi, type DocumentSource } from '../api'
import { github, type FileTreeNode } from '../services/github'
import TreeNode from '../components/TreeNode.vue'
import ReferencePanel, { type ReferenceTab } from '../components/ReferencePanel.vue'
import MarkdownIt from 'markdown-it'

const route = useRoute()
const sourceId = computed(() => Number(route.params.id))

const source = ref<DocumentSource | null>(null)
const fileTree = ref<FileTreeNode[]>([])
const loading = ref(true)
const loadingContent = ref(false)
const error = ref('')

// Share modal state
const showShareModal = ref(false)
const shareMode = ref<'library' | 'document'>('library')
const shareUrl = ref('')
const shareShortUrl = ref('')
const shareCopied = ref(false)
const shareLoading = ref(false)

const isShareable = computed(() => source.value?.source_type === 'github_public')

// Current document state
const currentPath = ref('')
const currentContent = ref('')
const currentTitle = ref('')

// Sidebar state
const sidebarOpen = ref(true)
const expandedDirs = ref<Set<string>>(new Set())
const searchQuery = ref('')

// Reading progress
const readPaths = ref<Set<string>>(new Set())

// --- Reference Panel ---
const refPanelOpen = ref(false)
const refTabs = ref<ReferenceTab[]>([])
const refActiveTabId = ref('')
const refMemorizeLoading = ref(false)

// --- Dark Mode ---
const darkMode = ref(localStorage.getItem('reader-dark-mode') === 'true')

function toggleDarkMode() {
  darkMode.value = !darkMode.value
  localStorage.setItem('reader-dark-mode', String(darkMode.value))
}

// --- Task Creation ---
const taskCreating = ref(false)
const taskCreated = ref<Set<string>>(new Set())

// Markdown renderer
const md = new MarkdownIt({
  html: true,
  linkify: true,
  typographer: true,
  breaks: false,
})

// Custom link renderer: open external links in new tabs, mark scripture links
const defaultRender = md.renderer.rules.link_open || function(tokens: any, idx: any, options: any, _env: any, self: any) {
  return self.renderToken(tokens, idx, options)
}
md.renderer.rules.link_open = function(tokens: any, idx: any, options: any, env: any, self: any) {
  const href = tokens[idx].attrGet('href')
  if (href && (href.startsWith('http://') || href.startsWith('https://'))) {
    if (isScriptureUrl(href)) {
      tokens[idx].attrSet('data-scripture-link', 'true')
    } else {
      tokens[idx].attrSet('target', '_blank')
      tokens[idx].attrSet('rel', 'noopener noreferrer')
    }
  }
  if (href && (href.includes('gospel-library') || href.includes('general-conference'))) {
    tokens[idx].attrSet('data-ref-link', 'true')
  }
  return defaultRender(tokens, idx, options, env, self)
}

const renderedContent = computed(() => {
  if (!currentContent.value) return ''
  let html = md.render(currentContent.value)
  html = injectBecomeButtons(html)
  return html
})

const filteredTree = computed(() => {
  if (!searchQuery.value) return fileTree.value
  const q = searchQuery.value.toLowerCase()
  return filterNodes(fileTree.value, q)
})

// When filtering, auto-expand all directories so matches are visible
const displayExpandedDirs = computed(() => {
  if (!searchQuery.value) return expandedDirs.value
  const allDirs = new Set(expandedDirs.value)
  const collectDirs = (nodes: FileTreeNode[]) => {
    for (const node of nodes) {
      if (node.type === 'directory') {
        allDirs.add(node.path)
        if (node.children) collectDirs(node.children)
      }
    }
  }
  collectDirs(filteredTree.value)
  return allDirs
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

function fileDisplayName(name: string): string {
  return name.replace(/\.md$/i, '').replace(/[-_]/g, ' ').replace(/^\d+\s*/, '')
}

function titleFromPath(path: string): string {
  const parts = path.split('/')
  const fileName = parts[parts.length - 1] || path
  return fileDisplayName(fileName)
}

function extractTitle(content: string): string {
  const match = content.match(/^#\s+(.+)$/m)
  return match?.[1] ?? ''
}

// --- Scripture link detection ---

function isScriptureUrl(url: string): boolean {
  return /churchofjesuschrist\.org\/study\/(scriptures|general-conference)/i.test(url)
}

function churchUrlToPath(url: string): string | null {
  try {
    const u = new URL(url)
    const match = u.pathname.match(/\/study\/(scriptures\/.+|general-conference\/.+)/)
    if (!match) return null
    let path = match[1]!
    path = path.replace(/\?.*$/, '')
    if (!path.endsWith('.md')) path += '.md'
    return `gospel-library/eng/${path}`
  } catch {
    return null
  }
}

function titleFromLink(linkEl: HTMLElement, href: string): string {
  const text = linkEl.textContent?.trim()
  if (text && text.length > 0 && text.length < 100) return text
  return titleFromPath(href)
}

// --- Become section detection ---

function injectBecomeButtons(html: string): string {
  const becomePattern = /(<h[23][^>]*>(?:[^<]*(?:Become|Application|Apply|Action|Commitment)[^<]*)<\/h[23]>)([\s\S]*?)(?=<h[23]|$)/gi
  return html.replace(becomePattern, (_match, heading, body) => {
    const updatedBody = body.replace(
      /<li>([\s\S]*?)<\/li>/g,
      (_liMatch: string, liContent: string) => {
        const plainText = liContent.replace(/<[^>]+>/g, '').trim().substring(0, 120)
        const encodedText = plainText.replace(/"/g, '&quot;')
        return `<li class="become-item">${liContent} <button class="become-task-btn" data-task-text="${encodedText}" title="Create task">+ Task</button></li>`
      }
    )
    return heading + updatedBody
  })
}

// --- Source & File loading ---

async function loadSource() {
  loading.value = true
  error.value = ''
  try {
    source.value = await api.getSource(sourceId.value)
    if (!source.value) {
      error.value = 'Source not found'
      return
    }

    const include = JSON.parse(source.value.include_paths || '[]') as string[]
    const exclude = JSON.parse(source.value.exclude_paths || '[]') as string[]
    const entries = await github.getTree(source.value.repo, source.value.branch, include, exclude)
    fileTree.value = github.buildFileTree(entries)

    for (const node of fileTree.value) {
      if (node.type === 'directory') {
        expandedDirs.value.add(node.path)
      }
    }

    try {
      const progress = await api.listReadingProgress(sourceId.value)
      readPaths.value = new Set(progress.map(p => p.file_path))
    } catch {
      // Non-critical
    }
  } catch (e: any) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

async function openFile(path: string) {
  if (!source.value) return
  loadingContent.value = true
  try {
    const content = await github.getContent(source.value.repo, source.value.branch, path)
    currentPath.value = path
    currentContent.value = content
    currentTitle.value = extractTitle(content) || titleFromPath(path)

    readPaths.value.add(path)
    try {
      await api.upsertReadingProgress(sourceId.value, path, 0)
    } catch {
      // Non-critical
    }

    await nextTick()
    const contentEl = document.getElementById('reader-content')
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

// --- Link interception ---

function handleContentClick(event: MouseEvent) {
  const target = event.target as HTMLElement

  // Handle "Create Task" buttons from Become sections
  const taskBtn = target.closest('.become-task-btn') as HTMLElement
  if (taskBtn) {
    event.preventDefault()
    const text = taskBtn.getAttribute('data-task-text') || ''
    createTaskFromBecome(text)
    return
  }

  const link = target.closest('a') as HTMLElement
  if (!link) return

  const href = link.getAttribute('href')
  if (!href) return

  // Check for scripture/reference links → open in reference panel
  if (link.hasAttribute('data-scripture-link')) {
    event.preventDefault()
    const path = churchUrlToPath(href)
    if (path) {
      openInReferencePanel(path, titleFromLink(link, href))
    }
    return
  }

  if (link.hasAttribute('data-ref-link')) {
    event.preventDefault()
    const mdPath = href.split('#')[0] ?? href
    const resolvedPath = resolvePath(currentPath.value, mdPath)
    openInReferencePanel(resolvedPath, titleFromLink(link, href))
    return
  }

  // Skip external links (they open in new tabs)
  if (href.startsWith('http://') || href.startsWith('https://')) return

  // Handle relative .md links — navigate in main reader
  if (href.endsWith('.md') || href.includes('.md#')) {
    event.preventDefault()
    const mdPath = href.split('#')[0] ?? href
    const resolvedPath = resolvePath(currentPath.value, mdPath)

    // If it looks like a gospel-library reference, open in panel
    if (resolvedPath.includes('gospel-library') || resolvedPath.includes('general-conference')) {
      openInReferencePanel(resolvedPath, titleFromLink(link, href))
    } else {
      openFile(resolvedPath)
    }
  }
}

function resolvePath(currentFile: string, relativePath: string): string {
  const currentDir = currentFile.substring(0, currentFile.lastIndexOf('/'))
  const parts = currentDir.split('/').filter(Boolean)
  const relParts = relativePath.split('/')

  for (const part of relParts) {
    if (part === '..') {
      parts.pop()
    } else if (part !== '.') {
      parts.push(part)
    }
  }

  return parts.join('/')
}

// --- Reference Panel operations ---

async function openInReferencePanel(path: string, title: string) {
  if (!source.value) return

  // Reuse existing tab if path matches
  const existing = refTabs.value.find(t => t.path === path)
  if (existing) {
    refActiveTabId.value = existing.id
    refPanelOpen.value = true
    return
  }

  const id = `ref-${Date.now()}-${Math.random().toString(36).substr(2, 5)}`
  const newTab: ReferenceTab = {
    id,
    title,
    path,
    content: '',
    loading: true,
    error: '',
  }
  refTabs.value.push(newTab)
  refActiveTabId.value = id
  refPanelOpen.value = true

  try {
    const content = await github.getContent(source.value.repo, source.value.branch, path)
    const tab = refTabs.value.find(t => t.id === id)
    if (tab) {
      tab.content = content
      tab.title = extractTitle(content) || title
      tab.loading = false
    }
  } catch (e: any) {
    const tab = refTabs.value.find(t => t.id === id)
    if (tab) {
      tab.error = `Could not load: ${e.message}`
      tab.loading = false
    }
  }
}

function closeRefTab(id: string) {
  const idx = refTabs.value.findIndex(t => t.id === id)
  if (idx === -1) return
  refTabs.value.splice(idx, 1)
  if (refTabs.value.length === 0) {
    refPanelOpen.value = false
    refActiveTabId.value = ''
  } else if (refActiveTabId.value === id) {
    refActiveTabId.value = refTabs.value[Math.max(0, idx - 1)]?.id ?? ''
  }
}

function closeAllRefTabs() {
  refTabs.value = []
  refPanelOpen.value = false
  refActiveTabId.value = ''
}

function handleRefLinkClick(href: string, _event: MouseEvent) {
  if (!source.value) return
  const mdPath = href.split('#')[0] ?? href
  const activeTab = refTabs.value.find(t => t.id === refActiveTabId.value)
  const basePath = activeTab?.path || currentPath.value
  const resolvedPath = resolvePath(basePath, mdPath)
  const title = titleFromPath(resolvedPath)
  openInReferencePanel(resolvedPath, title)
}

// --- Add to Memorize ---

async function addToMemorize(tab: ReferenceTab) {
  refMemorizeLoading.value = true
  try {
    // Extract a short excerpt for the description (first meaningful paragraph)
    const lines = tab.content.split('\n').filter(l => l.trim() && !l.startsWith('#'))
    const excerpt = lines.slice(0, 3).join('\n').substring(0, 500)

    await api.createPractice({
      name: tab.title,
      type: 'memorize',
      description: excerpt,
      source_doc: tab.path,
      source_path: tab.path,
      category: 'scripture',
    })
  } catch (e: any) {
    error.value = `Failed to add to memorize: ${e.message}`
  } finally {
    refMemorizeLoading.value = false
  }
}

// --- Become task creation ---

async function createTaskFromBecome(text: string) {
  if (taskCreated.value.has(text)) return
  taskCreating.value = true
  try {
    await api.createTask({
      title: text,
      description: `From study: ${currentTitle.value}`,
      source_doc: currentPath.value,
      scripture: '',
      type: 'become',
      status: 'active',
    })
    taskCreated.value.add(text)
  } catch (e: any) {
    error.value = `Failed to create task: ${e.message}`
  } finally {
    taskCreating.value = false
  }
}

// --- Share ---

async function openShareModal(mode: 'library' | 'document') {
  if (!source.value) return
  shareMode.value = mode
  shareCopied.value = false
  shareLoading.value = true
  showShareModal.value = true

  const s = source.value
  const includes = JSON.parse(s.include_paths || '[]') as string[]
  const docFilter = includes[0] || '**/*.md'

  const base = window.location.origin
  const params = new URLSearchParams()
  params.set('p', 'gh')
  params.set('r', s.repo)
  if (s.branch !== 'main') params.set('b', s.branch)
  if (docFilter !== '**/*.md') params.set('d', docFilter)
  if (mode === 'document' && currentPath.value) params.set('f', currentPath.value)
  shareUrl.value = `${base}/share?${params.toString()}`

  try {
    const link = await publicApi.createShareLink({
      repo: s.repo,
      branch: s.branch,
      doc_filter: docFilter,
      file_path: mode === 'document' ? currentPath.value : undefined,
      source_id: s.id,
    })
    shareShortUrl.value = `${base}/s/${link.code}`
  } catch {
    shareShortUrl.value = ''
  } finally {
    shareLoading.value = false
  }
}

async function copyShareUrl(url: string) {
  try {
    await navigator.clipboard.writeText(url)
    shareCopied.value = true
    setTimeout(() => { shareCopied.value = false }, 2000)
  } catch {
    // fallback
  }
}

// --- Mobile sidebar ---
const isMobile = ref(window.innerWidth < 768)

function handleResize() {
  isMobile.value = window.innerWidth < 768
  if (isMobile.value && sidebarOpen.value && currentContent.value) {
    sidebarOpen.value = false
  }
}

onMounted(() => {
  loadSource()
  window.addEventListener('resize', handleResize)
  handleResize()
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
})
</script>

<template>
  <div class="reader-layout" :class="{ dark: darkMode }">
    <!-- Mobile sidebar overlay -->
    <div
      v-if="isMobile && sidebarOpen"
      class="sidebar-overlay"
      @click="sidebarOpen = false"
    />

    <!-- Sidebar Toggle (when collapsed) -->
    <button
      v-if="!sidebarOpen"
      @click="toggleSidebar"
      class="sidebar-toggle-btn"
      title="Show sidebar"
    >
      <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
        <path fill-rule="evenodd" d="M3 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm0 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm0 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1z" clip-rule="evenodd" />
      </svg>
    </button>

    <!-- Sidebar -->
    <aside v-if="sidebarOpen" class="reader-sidebar" :class="{ 'mobile-sidebar': isMobile }">
      <div class="flex items-center justify-between px-3 py-2 border-b" style="border-color: var(--border-color, #e5e7eb)">
        <h2 class="text-sm font-semibold truncate" style="color: var(--text-secondary, #374151)" :title="source?.name">
          {{ source?.name || 'Loading...' }}
        </h2>
        <div class="flex items-center gap-1">
          <!-- Dark mode toggle -->
          <button
            @click="toggleDarkMode"
            class="p-1 rounded hover:bg-gray-100 dark:hover:bg-gray-700"
            :title="darkMode ? 'Light mode' : 'Dark mode'"
          >
            <svg v-if="darkMode" xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-yellow-400" viewBox="0 0 20 20" fill="currentColor">
              <path fill-rule="evenodd" d="M10 2a1 1 0 011 1v1a1 1 0 11-2 0V3a1 1 0 011-1zm4 8a4 4 0 11-8 0 4 4 0 018 0zm-.464 4.95l.707.707a1 1 0 001.414-1.414l-.707-.707a1 1 0 00-1.414 1.414zm2.12-10.607a1 1 0 010 1.414l-.706.707a1 1 0 11-1.414-1.414l.707-.707a1 1 0 011.414 0zM17 11a1 1 0 100-2h-1a1 1 0 100 2h1zm-7 4a1 1 0 011 1v1a1 1 0 11-2 0v-1a1 1 0 011-1zM5.05 6.464A1 1 0 106.465 5.05l-.708-.707a1 1 0 00-1.414 1.414l.707.707zm1.414 8.486l-.707.707a1 1 0 01-1.414-1.414l.707-.707a1 1 0 011.414 1.414zM4 11a1 1 0 100-2H3a1 1 0 000 2h1z" clip-rule="evenodd" />
            </svg>
            <svg v-else xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-gray-400" viewBox="0 0 20 20" fill="currentColor">
              <path d="M17.293 13.293A8 8 0 016.707 2.707a8.001 8.001 0 1010.586 10.586z" />
            </svg>
          </button>
          <button
            v-if="isShareable"
            @click="openShareModal('library')"
            class="p-1" style="color: var(--text-muted, #9ca3af)"
            title="Share"
          >
            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
              <path d="M15 8a3 3 0 10-2.977-2.63l-4.94 2.47a3 3 0 100 4.319l4.94 2.47a3 3 0 10.895-1.789l-4.94-2.47a3.027 3.027 0 000-.74l4.94-2.47C13.456 7.68 14.19 8 15 8z" />
            </svg>
          </button>
          <button @click="toggleSidebar" class="p-1" style="color: var(--text-muted, #9ca3af)" title="Hide sidebar">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
              <path fill-rule="evenodd" d="M12.707 5.293a1 1 0 010 1.414L9.414 10l3.293 3.293a1 1 0 01-1.414 1.414l-4-4a1 1 0 010-1.414l4-4a1 1 0 011.414 0z" clip-rule="evenodd" />
            </svg>
          </button>
        </div>
      </div>

      <!-- Search -->
      <div class="px-3 py-2" style="border-bottom: 1px solid var(--border-light, #f3f4f6)">
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Filter files..."
          class="w-full rounded px-2 py-1 text-xs"
          style="border: 1px solid var(--border-color, #e5e7eb); background: var(--bg-input, white); color: var(--text-primary, #111827)"
        />
      </div>

      <!-- File Tree -->
      <div class="overflow-y-auto flex-1 py-1">
        <div v-if="loading" class="px-3 py-4 text-xs" style="color: var(--text-muted, #9ca3af)">Loading tree...</div>
        <div v-else-if="error && !currentContent" class="px-3 py-4 text-xs text-red-500">{{ error }}</div>
        <template v-else>
          <TreeNode
            v-for="node in filteredTree"
            :key="node.path"
            :node="node"
            :depth="0"
            :expanded-dirs="displayExpandedDirs"
            :current-path="currentPath"
            :read-paths="readPaths"
            @toggle-dir="toggleDir"
            @open-file="(p) => { openFile(p); if (isMobile) sidebarOpen = false }"
          />
        </template>
      </div>

      <!-- Source link -->
      <div style="border-top: 1px solid var(--border-light, #f3f4f6)" class="px-3 py-2">
        <a
          v-if="source"
          :href="`https://github.com/${source.repo}`"
          target="_blank"
          rel="noopener noreferrer"
          class="text-xs flex items-center gap-1"
          style="color: var(--text-muted, #9ca3af)"
        >
          <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" viewBox="0 0 16 16" fill="currentColor">
            <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z"/>
          </svg>
          {{ source.repo }}
        </a>
      </div>
    </aside>

    <!-- Main Content -->
    <main class="reader-content" id="reader-content" @click="handleContentClick">
      <!-- Empty state -->
      <div v-if="!currentContent && !loadingContent && !loading" class="flex items-center justify-center h-full">
        <div class="text-center" style="color: var(--text-muted, #9ca3af)">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-12 w-12 mx-auto mb-3 opacity-50" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" />
          </svg>
          <p class="text-lg">Select a document to read</p>
          <p v-if="!sidebarOpen" class="text-sm mt-1">Open the sidebar to browse files</p>
        </div>
      </div>

      <!-- Loading content -->
      <div v-else-if="loadingContent" class="flex items-center justify-center h-full">
        <div style="color: var(--text-muted, #9ca3af)">Loading...</div>
      </div>

      <!-- Document -->
      <div v-else-if="currentContent" class="reader-document">
        <!-- Breadcrumb -->
        <div class="flex items-center justify-between text-xs mb-4" style="color: var(--text-muted, #9ca3af)">
          <span class="font-mono">{{ currentPath }}</span>
          <button
            v-if="isShareable"
            @click="openShareModal('document')"
            class="flex items-center gap-1 hover:text-orange-600"
          >
            <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" viewBox="0 0 20 20" fill="currentColor">
              <path d="M15 8a3 3 0 10-2.977-2.63l-4.94 2.47a3 3 0 100 4.319l4.94 2.47a3 3 0 10.895-1.789l-4.94-2.47a3.027 3.027 0 000-.74l4.94-2.47C13.456 7.68 14.19 8 15 8z" />
            </svg>
            Share
          </button>
        </div>

        <!-- Rendered markdown -->
        <article class="prose prose-gray max-w-none" v-html="renderedContent" />
      </div>
    </main>

    <!-- Reference Panel -->
    <ReferencePanel
      v-if="refPanelOpen && refTabs.length > 0"
      :tabs="refTabs"
      :active-tab-id="refActiveTabId"
      :authenticated="true"
      :memorize-loading="refMemorizeLoading"
      @close-tab="closeRefTab"
      @close-all="closeAllRefTabs"
      @select-tab="(id) => refActiveTabId = id"
      @add-to-memorize="addToMemorize"
      @close-panel="refPanelOpen = false"
      @link-click="handleRefLinkClick"
    />

    <!-- Share Modal -->
    <div v-if="showShareModal" class="fixed inset-0 bg-black/40 flex items-center justify-center z-50" @click.self="showShareModal = false">
      <div class="rounded-lg p-6 w-[440px] shadow-xl" style="background: var(--bg-surface, white)">
        <h3 class="text-lg font-semibold mb-1" style="color: var(--text-primary, #111827)">Share {{ shareMode === 'document' ? 'Document' : 'Library' }}</h3>
        <p class="text-sm mb-4" style="color: var(--text-muted, #6b7280)">
          {{ shareMode === 'document' ? 'Share a link to this specific document.' : 'Share a link to this entire study library.' }}
        </p>

        <div v-if="shareLoading" class="py-6 text-center" style="color: var(--text-muted, #9ca3af)">Generating link...</div>
        <div v-else class="space-y-3">
          <div v-if="shareShortUrl">
            <label class="block text-xs font-medium mb-1" style="color: var(--text-secondary, #4b5563)">Short Link</label>
            <div class="flex gap-2">
              <input
                :value="shareShortUrl"
                readonly
                class="flex-1 rounded px-3 py-2 text-sm font-mono"
                style="border: 1px solid var(--border-color, #e5e7eb); background: var(--bg-muted, #f9fafb); color: var(--text-primary, #111827)"
                @click="($event.target as HTMLInputElement).select()"
              />
              <button
                @click="copyShareUrl(shareShortUrl)"
                class="px-3 py-2 text-sm bg-orange-600 text-white rounded hover:bg-orange-700"
              >
                {{ shareCopied ? 'Copied!' : 'Copy' }}
              </button>
            </div>
          </div>

          <div>
            <label class="block text-xs font-medium mb-1" style="color: var(--text-secondary, #4b5563)">{{ shareShortUrl ? 'Full Link' : 'Share Link' }}</label>
            <div class="flex gap-2">
              <input
                :value="shareUrl"
                readonly
                class="flex-1 rounded px-3 py-2 text-sm font-mono text-xs"
                style="border: 1px solid var(--border-color, #e5e7eb); background: var(--bg-muted, #f9fafb); color: var(--text-primary, #111827)"
                @click="($event.target as HTMLInputElement).select()"
              />
              <button
                @click="copyShareUrl(shareUrl)"
                class="px-3 py-2 text-sm text-orange-600 rounded hover:bg-orange-50"
                style="border: 1px solid #fed7aa"
              >
                Copy
              </button>
            </div>
          </div>

          <div v-if="currentPath && shareMode === 'document'" class="pt-2" style="border-top: 1px solid var(--border-light, #f3f4f6)">
            <button @click="openShareModal('library')" class="text-xs hover:text-orange-600" style="color: var(--text-muted, #6b7280)">
              Or share the entire library instead
            </button>
          </div>
          <div v-else-if="currentPath && shareMode === 'library'" class="pt-2" style="border-top: 1px solid var(--border-light, #f3f4f6)">
            <button @click="openShareModal('document')" class="text-xs hover:text-orange-600" style="color: var(--text-muted, #6b7280)">
              Or share just this document instead
            </button>
          </div>
        </div>

        <div class="flex justify-end mt-4">
          <button @click="showShareModal = false" class="px-4 py-2 text-sm" style="color: var(--text-secondary, #4b5563)">Close</button>
        </div>
      </div>
    </div>
  </div>
</template>


<style scoped>
/* --- CSS Custom Properties for dark mode --- */
.reader-layout {
  display: flex;
  height: calc(100vh - 64px);
  width: 100%;

  --bg-page: #fafafa;
  --bg-surface: white;
  --bg-muted: #f9fafb;
  --bg-hover: #f3f4f6;
  --bg-input: white;
  --bg-code: #f3f4f6;
  --bg-quote: #fffbeb;
  --text-primary: #111827;
  --text-secondary: #374151;
  --text-body: #374151;
  --text-muted: #9ca3af;
  --border-color: #e5e7eb;
  --border-light: #f3f4f6;
}

.reader-layout.dark {
  --bg-page: #111827;
  --bg-surface: #1f2937;
  --bg-muted: #1a2332;
  --bg-hover: #374151;
  --bg-input: #374151;
  --bg-code: #374151;
  --bg-quote: #292524;
  --text-primary: #f3f4f6;
  --text-secondary: #d1d5db;
  --text-body: #d1d5db;
  --text-muted: #6b7280;
  --border-color: #374151;
  --border-light: #1f2937;
}

.sidebar-toggle-btn {
  position: fixed;
  left: 8px;
  top: 80px;
  z-index: 10;
  padding: 8px;
  border-radius: 8px;
  border: 1px solid var(--border-color);
  background: var(--bg-surface);
  color: var(--text-secondary);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
  cursor: pointer;
}

.sidebar-toggle-btn:hover {
  background: var(--bg-hover);
}

.sidebar-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.4);
  z-index: 19;
}

.reader-sidebar {
  width: 280px;
  min-width: 280px;
  border-right: 1px solid var(--border-color);
  background: var(--bg-surface);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.reader-sidebar.mobile-sidebar {
  position: fixed;
  top: 64px;
  left: 0;
  bottom: 0;
  z-index: 20;
  box-shadow: 4px 0 12px rgba(0, 0, 0, 0.15);
}

.reader-content {
  flex: 1;
  overflow-y: auto;
  padding: 24px 32px;
  background: var(--bg-page);
}

.reader-document {
  max-width: 800px;
  margin: 0 auto;
  background: var(--bg-surface);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 32px 40px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
}

/* Prose overrides for rendered markdown */
.reader-document :deep(h1) {
  font-size: 1.75rem;
  font-weight: 700;
  margin-bottom: 1rem;
  color: var(--text-primary);
  border-bottom: 1px solid var(--border-color);
  padding-bottom: 0.5rem;
}

.reader-document :deep(h2) {
  font-size: 1.35rem;
  font-weight: 600;
  margin-top: 2rem;
  margin-bottom: 0.75rem;
  color: var(--text-primary);
}

.reader-document :deep(h3) {
  font-size: 1.1rem;
  font-weight: 600;
  margin-top: 1.5rem;
  margin-bottom: 0.5rem;
  color: var(--text-secondary);
}

.reader-document :deep(p) {
  line-height: 1.75;
  margin-bottom: 1rem;
  color: var(--text-body);
}

.reader-document :deep(blockquote) {
  border-left: 3px solid #d97706;
  padding: 0.5rem 1rem;
  margin: 1rem 0;
  background: var(--bg-quote);
  border-radius: 0 4px 4px 0;
}

.reader-document :deep(blockquote p) {
  color: #92400e;
  margin-bottom: 0.25rem;
}

.dark .reader-document :deep(blockquote p) {
  color: #fbbf24;
}

.reader-document :deep(a) {
  color: #ea580c;
  text-decoration: underline;
  text-decoration-color: #fed7aa;
  text-underline-offset: 2px;
}

.reader-document :deep(a:hover) {
  color: #c2410c;
  text-decoration-color: #ea580c;
}

.reader-document :deep(a[data-ref-link]),
.reader-document :deep(a[data-scripture-link]) {
  cursor: pointer;
  border-bottom: 2px solid #d97706;
  text-decoration: none;
}

.reader-document :deep(a[data-ref-link]:hover),
.reader-document :deep(a[data-scripture-link]:hover) {
  background: rgba(217, 119, 6, 0.1);
  border-radius: 2px;
}

.reader-document :deep(code) {
  background: var(--bg-code);
  padding: 0.15rem 0.35rem;
  border-radius: 3px;
  font-size: 0.875em;
  color: var(--text-primary);
}

.reader-document :deep(pre) {
  background: #1f2937;
  color: #e5e7eb;
  padding: 1rem;
  border-radius: 6px;
  overflow-x: auto;
  margin: 1rem 0;
}

.reader-document :deep(pre code) {
  background: none;
  color: inherit;
  padding: 0;
}

.reader-document :deep(ul),
.reader-document :deep(ol) {
  padding-left: 1.5rem;
  margin-bottom: 1rem;
}

.reader-document :deep(li) {
  line-height: 1.75;
  margin-bottom: 0.25rem;
  color: var(--text-body);
}

.reader-document :deep(table) {
  width: 100%;
  border-collapse: collapse;
  margin: 1rem 0;
}

.reader-document :deep(th),
.reader-document :deep(td) {
  border: 1px solid var(--border-color);
  padding: 0.5rem 0.75rem;
  text-align: left;
  font-size: 0.875rem;
  color: var(--text-body);
}

.reader-document :deep(th) {
  background: var(--bg-muted);
  font-weight: 600;
}

.reader-document :deep(hr) {
  border: none;
  border-top: 1px solid var(--border-color);
  margin: 2rem 0;
}

.reader-document :deep(img) {
  max-width: 100%;
  border-radius: 6px;
}

/* Become section styling */
.reader-document :deep(.become-item) {
  position: relative;
}

.reader-document :deep(.become-task-btn) {
  display: inline-flex;
  align-items: center;
  margin-left: 8px;
  padding: 2px 8px;
  font-size: 0.7rem;
  font-weight: 600;
  color: #ea580c;
  border: 1px solid #fed7aa;
  border-radius: 4px;
  background: transparent;
  cursor: pointer;
  vertical-align: middle;
  white-space: nowrap;
  transition: all 0.15s;
}

.reader-document :deep(.become-task-btn:hover) {
  background: #fff7ed;
  border-color: #ea580c;
}

/* Mobile responsive */
@media (max-width: 768px) {
  .reader-content {
    padding: 12px 8px;
  }

  .reader-document {
    padding: 16px 16px;
    border-radius: 0;
    border-left: none;
    border-right: none;
  }

  .reader-document :deep(h1) {
    font-size: 1.4rem;
  }

  .reader-document :deep(h2) {
    font-size: 1.15rem;
  }

  .reader-document :deep(table) {
    display: block;
    overflow-x: auto;
  }
}
</style>
