<script setup lang="ts">
import { ref, computed, onMounted, nextTick } from 'vue'
import { useRoute } from 'vue-router'
import { api, type DocumentSource } from '../api'
import { github, type FileTreeNode } from '../services/github'
import TreeNode from '../components/TreeNode.vue'
import MarkdownIt from 'markdown-it'

const route = useRoute()
const sourceId = computed(() => Number(route.params.id))

const source = ref<DocumentSource | null>(null)
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

// Reading progress
const readPaths = ref<Set<string>>(new Set())

// Markdown renderer
const md = new MarkdownIt({
  html: true,
  linkify: true,
  typographer: true,
  breaks: false,
})

// Custom link renderer: open external links in new tabs
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

// Filter the file tree based on search query
const filteredTree = computed(() => {
  if (!searchQuery.value) return fileTree.value
  const q = searchQuery.value.toLowerCase()
  return filterNodes(fileTree.value, q)
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

// File name to display title (strip .md, replace hyphens/underscores)
function fileDisplayName(name: string): string {
  return name
    .replace(/\.md$/i, '')
    .replace(/[-_]/g, ' ')
    .replace(/^\d+\s*/, '')  // Strip leading number prefixes like "01_"
}

function titleFromPath(path: string): string {
  const parts = path.split('/')
  const fileName = parts[parts.length - 1] || path
  return fileDisplayName(fileName)
}

// Extract title from markdown content (first # heading)
function extractTitle(content: string): string {
  const match = content.match(/^#\s+(.+)$/m)
  return match?.[1] ?? ''
}

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

    // Auto-expand first-level directories
    for (const node of fileTree.value) {
      if (node.type === 'directory') {
        expandedDirs.value.add(node.path)
      }
    }

    // Load reading progress
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

    // Mark as read
    readPaths.value.add(path)
    try {
      await api.upsertReadingProgress(sourceId.value, path, 0)
    } catch {
      // Non-critical
    }

    // Scroll to top of content
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

// Handle internal link clicks within rendered markdown
function handleContentClick(event: MouseEvent) {
  const target = event.target as HTMLElement
  const link = target.closest('a')
  if (!link) return

  const href = link.getAttribute('href')
  if (!href) return

  // Skip external links (they open in new tabs via our custom renderer)
  if (href.startsWith('http://') || href.startsWith('https://')) return

  // Handle relative .md links — resolve relative to current file
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
    if (part === '..') {
      parts.pop()
    } else if (part !== '.') {
      parts.push(part)
    }
  }

  return parts.join('/')
}

onMounted(loadSource)
</script>

<template>
  <div class="reader-layout">
    <!-- Sidebar Toggle (when collapsed) -->
    <button
      v-if="!sidebarOpen"
      @click="toggleSidebar"
      class="fixed left-2 top-20 z-10 bg-white border border-gray-200 rounded-lg p-2 shadow-sm hover:bg-gray-50"
      title="Show sidebar"
    >
      <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-gray-600" viewBox="0 0 20 20" fill="currentColor">
        <path fill-rule="evenodd" d="M3 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm0 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm0 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1z" clip-rule="evenodd" />
      </svg>
    </button>

    <!-- Sidebar -->
    <aside v-if="sidebarOpen" class="reader-sidebar">
      <div class="flex items-center justify-between px-3 py-2 border-b border-gray-200">
        <h2 class="text-sm font-semibold text-gray-700 truncate" :title="source?.name">
          {{ source?.name || 'Loading...' }}
        </h2>
        <button @click="toggleSidebar" class="text-gray-400 hover:text-gray-600 p-1" title="Hide sidebar">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
            <path fill-rule="evenodd" d="M12.707 5.293a1 1 0 010 1.414L9.414 10l3.293 3.293a1 1 0 01-1.414 1.414l-4-4a1 1 0 010-1.414l4-4a1 1 0 011.414 0z" clip-rule="evenodd" />
          </svg>
        </button>
      </div>

      <!-- Search -->
      <div class="px-3 py-2 border-b border-gray-100">
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Filter files..."
          class="w-full border border-gray-200 rounded px-2 py-1 text-xs"
        />
      </div>

      <!-- File Tree -->
      <div class="overflow-y-auto flex-1 py-1">
        <div v-if="loading" class="px-3 py-4 text-xs text-gray-400">Loading tree...</div>
        <div v-else-if="error && !currentContent" class="px-3 py-4 text-xs text-red-500">{{ error }}</div>
        <template v-else>
          <TreeNode
            v-for="node in filteredTree"
            :key="node.path"
            :node="node"
            :depth="0"
            :expanded-dirs="expandedDirs"
            :current-path="currentPath"
            :read-paths="readPaths"
            @toggle-dir="toggleDir"
            @open-file="openFile"
          />
        </template>
      </div>

      <!-- Source link -->
      <div class="border-t border-gray-100 px-3 py-2">
        <a
          v-if="source"
          :href="`https://github.com/${source.repo}`"
          target="_blank"
          rel="noopener noreferrer"
          class="text-xs text-gray-400 hover:text-gray-600 flex items-center gap-1"
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
        <div class="text-center text-gray-400">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-12 w-12 mx-auto mb-3 opacity-50" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" />
          </svg>
          <p class="text-lg">Select a document to read</p>
          <p v-if="!sidebarOpen" class="text-sm mt-1">Open the sidebar to browse files</p>
        </div>
      </div>

      <!-- Loading content -->
      <div v-else-if="loadingContent" class="flex items-center justify-center h-full">
        <div class="text-gray-400">Loading...</div>
      </div>

      <!-- Document -->
      <div v-else-if="currentContent" class="reader-document">
        <!-- Breadcrumb -->
        <div class="text-xs text-gray-400 mb-4 font-mono">
          {{ currentPath }}
        </div>

        <!-- Rendered markdown -->
        <article class="prose prose-gray max-w-none" v-html="renderedContent" />
      </div>
    </main>
  </div>
</template>


<style scoped>
.reader-layout {
  display: flex;
  height: calc(100vh - 64px);
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

/* Prose overrides for rendered markdown */
.reader-document :deep(h1) {
  font-size: 1.75rem;
  font-weight: 700;
  margin-bottom: 1rem;
  color: #111827;
  border-bottom: 1px solid #e5e7eb;
  padding-bottom: 0.5rem;
}

.reader-document :deep(h2) {
  font-size: 1.35rem;
  font-weight: 600;
  margin-top: 2rem;
  margin-bottom: 0.75rem;
  color: #1f2937;
}

.reader-document :deep(h3) {
  font-size: 1.1rem;
  font-weight: 600;
  margin-top: 1.5rem;
  margin-bottom: 0.5rem;
  color: #374151;
}

.reader-document :deep(p) {
  line-height: 1.75;
  margin-bottom: 1rem;
  color: #374151;
}

.reader-document :deep(blockquote) {
  border-left: 3px solid #d97706;
  padding: 0.5rem 1rem;
  margin: 1rem 0;
  background: #fffbeb;
  border-radius: 0 4px 4px 0;
}

.reader-document :deep(blockquote p) {
  color: #92400e;
  margin-bottom: 0.25rem;
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

.reader-document :deep(code) {
  background: #f3f4f6;
  padding: 0.15rem 0.35rem;
  border-radius: 3px;
  font-size: 0.875em;
  color: #1f2937;
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
  color: #374151;
}

.reader-document :deep(table) {
  width: 100%;
  border-collapse: collapse;
  margin: 1rem 0;
}

.reader-document :deep(th),
.reader-document :deep(td) {
  border: 1px solid #e5e7eb;
  padding: 0.5rem 0.75rem;
  text-align: left;
  font-size: 0.875rem;
}

.reader-document :deep(th) {
  background: #f9fafb;
  font-weight: 600;
}

.reader-document :deep(hr) {
  border: none;
  border-top: 1px solid #e5e7eb;
  margin: 2rem 0;
}

.reader-document :deep(img) {
  max-width: 100%;
  border-radius: 6px;
}

/* Strong emphasis for "Become" headers */
.reader-document :deep(h2:has(+ p > strong)),
.reader-document :deep(h2:last-of-type) {
  color: #b45309;
}
</style>
