<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api, publicApi, type DocumentSource, type Pillar } from '../api'
import { github, type FileTreeNode } from '../services/github'
import TreeNode from '../components/TreeNode.vue'
import ReferencePanel, { type ReferenceTab } from '../components/ReferencePanel.vue'
import MarkdownIt from 'markdown-it'

const route = useRoute()
const router = useRouter()
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

// --- Text Selection Trigger ---
const selectionTrigger = ref<{ show: boolean; x: number; y: number; text: string }>({ show: false, x: 0, y: 0, text: '' })

// --- Practice Creation Form ---
const presetCategories = ['spiritual', 'scripture', 'pt', 'fitness', 'study', 'health']
const showPracticeForm = ref(false)
const practiceForm = ref({
  name: '',
  description: '',
  category: '',
  pillarIds: [] as number[],
  reps: 1,
  startDate: '',
  endDate: '',
})
const allPillars = ref<Pillar[]>([])
const practiceFormSaving = ref(false)
const practiceFormError = ref('')
const practiceFormSuccess = ref(false)

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

// Convert a relative gospel-library path to a church website URL
function pathToChurchUrl(path: string): string | null {
  // e.g. gospel-library/eng/scriptures/nt/john/1.md → https://www.churchofjesuschrist.org/study/scriptures/nt/john/1?lang=eng
  const match = path.match(/gospel-library\/eng\/(scriptures\/.+|general-conference\/.+)/)
  if (!match) return null
  let segment = match[1]!
  segment = segment.replace(/\.md$/, '')
  return `https://www.churchofjesuschrist.org/study/${segment}?lang=eng`
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

// Internal file loader — updates state without touching the URL
async function loadFile(path: string) {
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

// Navigate to a file — pushes a history entry so back button works
async function openFile(path: string) {
  // Push the file path into the URL query so the browser tracks it
  router.push({ params: route.params, query: { ...route.query, f: path } })
}

// Watch for route query changes (back/forward button)
watch(
  () => route.query.f as string | undefined,
  async (newPath) => {
    if (newPath && newPath !== currentPath.value) {
      await loadFile(newPath)
    } else if (!newPath && currentPath.value) {
      // Navigated back to before any file was opened
      currentPath.value = ''
      currentContent.value = ''
      currentTitle.value = ''
    }
  },
)

function toggleDir(path: string) {
  if (expandedDirs.value.has(path)) {
    expandedDirs.value.delete(path)
  } else {
    expandedDirs.value.add(path)
  }
}

// Expand all ancestor directories for a given file path so it's visible in the tree.
function expandToPath(filePath: string) {
  const segments = filePath.split('/')
  // Build cumulative directory paths: "public", "public/study", etc.
  for (let i = 1; i < segments.length; i++) {
    const dirPath = segments.slice(0, i).join('/')
    expandedDirs.value.add(dirPath)
  }
}

// Scroll the sidebar so the currently highlighted (active) file is visible.
function scrollSidebarToActive() {
  const sidebar = document.querySelector('[data-sidebar]')
  if (!sidebar) return
  const activeEl = sidebar.querySelector('.bg-orange-50')
  if (activeEl) {
    activeEl.scrollIntoView({ block: 'center', behavior: 'smooth' })
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

  // Check for scripture/reference links → open in reference panel as iframe
  if (link.hasAttribute('data-scripture-link')) {
    event.preventDefault()
    // Open the church website directly in an iframe — no local file fetch
    openIframeReference(href, titleFromLink(link, href))
    return
  }

  if (link.hasAttribute('data-ref-link')) {
    event.preventDefault()
    const mdPath = href.split('#')[0] ?? href
    const resolvedPath = resolvePath(currentPath.value, mdPath)
    // Convert gospel-library path to church URL for iframe
    const churchUrl = pathToChurchUrl(resolvedPath)
    if (churchUrl) {
      openIframeReference(churchUrl, titleFromLink(link, href))
    } else {
      // Non-scripture reference (e.g. study aid) — fetch from repo
      openInReferencePanel(resolvedPath, titleFromLink(link, href))
    }
    return
  }

  // Skip external links (they open in new tabs)
  if (href.startsWith('http://') || href.startsWith('https://')) return

  // Handle relative .md links — navigate in main reader
  if (href.endsWith('.md') || href.includes('.md#')) {
    event.preventDefault()
    const mdPath = href.split('#')[0] ?? href
    const resolvedPath = resolvePath(currentPath.value, mdPath)

    // If it looks like a gospel-library reference, open as iframe
    if (resolvedPath.includes('gospel-library') || resolvedPath.includes('general-conference')) {
      const churchUrl = pathToChurchUrl(resolvedPath)
      if (churchUrl) {
        openIframeReference(churchUrl, titleFromLink(link, href))
      } else {
        openInReferencePanel(resolvedPath, titleFromLink(link, href))
      }
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

// Open an external URL (e.g. church website) in an iframe tab
function openIframeReference(url: string, title: string) {
  // Reuse existing tab if URL matches
  const existing = refTabs.value.find(t => t.url === url)
  if (existing) {
    refActiveTabId.value = existing.id
    refPanelOpen.value = true
    return
  }

  const id = `ref-${Date.now()}-${Math.random().toString(36).substr(2, 5)}`
  const newTab: ReferenceTab = {
    id,
    title,
    path: url,      // display the URL
    url,             // iframe source
    content: '',
    loading: false,  // iframe handles its own loading
    error: '',
  }
  refTabs.value.push(newTab)
  refActiveTabId.value = id
  refPanelOpen.value = true
}

// Open a repo file (markdown) in a content tab
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
    url: '',         // no iframe — content will be fetched from repo
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
  // If it's a church URL, open as iframe
  if (isScriptureUrl(href)) {
    openIframeReference(href, titleFromPath(href))
    return
  }
  const mdPath = href.split('#')[0] ?? href
  const activeTab = refTabs.value.find(t => t.id === refActiveTabId.value)
  const basePath = activeTab?.path || currentPath.value
  const resolvedPath = resolvePath(basePath, mdPath)
  const churchUrl = pathToChurchUrl(resolvedPath)
  if (churchUrl) {
    openIframeReference(churchUrl, titleFromPath(resolvedPath))
  } else {
    openInReferencePanel(resolvedPath, titleFromPath(resolvedPath))
  }
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

// --- Text Selection handlers ---

function handleSelectionChange() {
  // Don't interfere while the practice form is open
  if (showPracticeForm.value) return

  const sel = window.getSelection()
  if (!sel || sel.isCollapsed || !sel.toString().trim()) {
    // Delay hiding so the trigger button can be clicked
    setTimeout(() => {
      if (showPracticeForm.value) return
      const sel2 = window.getSelection()
      if (!sel2 || sel2.isCollapsed || !sel2.toString().trim()) {
        selectionTrigger.value.show = false
      }
    }, 200)
    return
  }

  const text = sel.toString().trim()
  if (text.length < 5) return // Too short to be meaningful

  // Only show trigger if selection is within the reader content area
  const contentEl = document.getElementById('reader-content')
  const anchorNode = sel.anchorNode
  if (!contentEl || !anchorNode || !contentEl.contains(anchorNode)) return

  const range = sel.getRangeAt(0)
  const rect = range.getBoundingClientRect()

  // Use fixed viewport coordinates — no scroll offset math needed
  selectionTrigger.value = {
    show: true,
    x: rect.left + rect.width / 2,
    y: rect.top - 8,
    text,
  }
}

function localDateStr(d: Date = new Date()): string {
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

// Detect scripture references in text (e.g. "D&C 93:36", "1 Nephi 3:7", "Alma 32:21", "Moses 1:39")
function detectScriptureRef(text: string): string | null {
  const patterns = [
    // D&C / DC / Doctrine and Covenants
    /D&C\s+(\d+:\d+(?:[–-]\d+)?)/i,
    /DC\s+(\d+:\d+(?:[–-]\d+)?)/i,
    /Doctrine\s+and\s+Covenants\s+(\d+:\d+(?:[–-]\d+)?)/i,
    // Book of Mormon books
    /(\d\s+Nephi\s+\d+:\d+(?:[–-]\d+)?)/i,
    /(Nephi\s+\d+:\d+(?:[–-]\d+)?)/i,
    /(Alma\s+\d+:\d+(?:[–-]\d+)?)/i,
    /(Helaman\s+\d+:\d+(?:[–-]\d+)?)/i,
    /(Mosiah\s+\d+:\d+(?:[–-]\d+)?)/i,
    /(Ether\s+\d+:\d+(?:[–-]\d+)?)/i,
    /(Moroni\s+\d+:\d+(?:[–-]\d+)?)/i,
    /(Mormon\s+\d+:\d+(?:[–-]\d+)?)/i,
    /(Jacob\s+\d+:\d+(?:[–-]\d+)?)/i,
    /(Enos\s+\d+:\d+(?:[–-]\d+)?)/i,
    /(Omni\s+\d+:\d+(?:[–-]\d+)?)/i,
    /(Jarom\s+\d+:\d+(?:[–-]\d+)?)/i,
    /(3\s+Nephi\s+\d+:\d+(?:[–-]\d+)?)/i,
    /(4\s+Nephi\s+\d+:\d+(?:[–-]\d+)?)/i,
    // Pearl of Great Price
    /(Moses\s+\d+:\d+(?:[–-]\d+)?)/i,
    /(Abraham\s+\d+:\d+(?:[–-]\d+)?)/i,
    /(JS[–-]H\s+\d+:\d+(?:[–-]\d+)?)/i,
    /(Articles\s+of\s+Faith\s+\d+:\d+(?:[–-]\d+)?)/i,
    // Old and New Testament
    /(Genesis\s+\d+:\d+(?:[–-]\d+)?)/i,
    /(Exodus\s+\d+:\d+(?:[–-]\d+)?)/i,
    /(Isaiah\s+\d+:\d+(?:[–-]\d+)?)/i,
    /(Psalms?\s+\d+:\d+(?:[–-]\d+)?)/i,
    /(Proverbs\s+\d+:\d+(?:[–-]\d+)?)/i,
    /(Matthew\s+\d+:\d+(?:[–-]\d+)?)/i,
    /(Mark\s+\d+:\d+(?:[–-]\d+)?)/i,
    /(Luke\s+\d+:\d+(?:[–-]\d+)?)/i,
    /(John\s+\d+:\d+(?:[–-]\d+)?)/i,
    /(Romans\s+\d+:\d+(?:[–-]\d+)?)/i,
    /(Hebrews\s+\d+:\d+(?:[–-]\d+)?)/i,
    /(James\s+\d+:\d+(?:[–-]\d+)?)/i,
    /(Revelation\s+\d+:\d+(?:[–-]\d+)?)/i,
    /(\d\s+Corinthians\s+\d+:\d+(?:[–-]\d+)?)/i,
    /(\d\s+Timothy\s+\d+:\d+(?:[–-]\d+)?)/i,
    /(\d\s+Peter\s+\d+:\d+(?:[–-]\d+)?)/i,
    // Catch-all: "Book Chapter:Verse" pattern
    /—\s*([A-Z]\w+(?:\s+\d+)?\s+\d+:\d+(?:[–-]\d+)?)/,
  ]

  for (const pattern of patterns) {
    const match = text.match(pattern)
    if (match) {
      // For D&C patterns, normalize the prefix
      if (/^D&C|^DC|^Doctrine/i.test(pattern.source)) {
        return `D&C ${match[1]}`
      }
      return match[1] || match[0]
    }
  }
  return null
}

// Look at the DOM context around the current selection for scripture references.
// Checks: parent blockquote/paragraph text, nearby sibling elements, link text.
function detectRefFromSelectionContext(): string | null {
  const sel = window.getSelection()
  if (!sel || sel.rangeCount === 0) return null

  const range = sel.getRangeAt(0)
  let container: Node | null = range.commonAncestorContainer

  // Walk up to find the nearest block-level element (blockquote, p, li, div)
  const blockTags = new Set(['BLOCKQUOTE', 'P', 'LI', 'DIV', 'ARTICLE', 'SECTION'])
  let blockEl: HTMLElement | null = null
  let node: Node | null = container
  while (node && node !== document.body) {
    if (node.nodeType === Node.ELEMENT_NODE && blockTags.has((node as HTMLElement).tagName)) {
      blockEl = node as HTMLElement
      break
    }
    node = node.parentNode
  }

  if (blockEl) {
    // Check the full text content of the containing block
    const blockText = blockEl.textContent || ''
    const ref = detectScriptureRef(blockText)
    if (ref) return ref

    // Check link text within the block (e.g. <a>D&C 93:24</a>)
    const links = blockEl.querySelectorAll('a')
    for (const link of links) {
      const linkText = link.textContent || ''
      const linkRef = detectScriptureRef(linkText)
      if (linkRef) return linkRef
      // Also check href for scripture paths (e.g. /scriptures/dc-testament/dc/93)
      const href = link.getAttribute('href') || ''
      const pathRef = detectRefFromPath(href)
      if (pathRef) return pathRef
    }

    // Check the next sibling block (citation often follows the quote)
    let nextSibling = blockEl.nextElementSibling
    for (let i = 0; i < 2 && nextSibling; i++) {
      const sibText = nextSibling.textContent || ''
      const sibRef = detectScriptureRef(sibText)
      if (sibRef) return sibRef
      nextSibling = nextSibling.nextElementSibling
    }
  }

  return null
}

// Extract a scripture reference from a URL path (e.g. /scriptures/dc-testament/dc/93 → D&C 93)
function detectRefFromPath(href: string): string | null {
  // D&C
  const dcMatch = href.match(/dc-testament\/dc\/(\d+)/)
  if (dcMatch) return `D&C ${dcMatch[1]}`
  // Book of Mormon
  const bofmMatch = href.match(/bofm\/([\w-]+)\/(\d+)/)
  if (bofmMatch) {
    const bookMap: Record<string, string> = {
      '1-ne': '1 Nephi', '2-ne': '2 Nephi', '3-ne': '3 Nephi', '4-ne': '4 Nephi',
      'jacob': 'Jacob', 'enos': 'Enos', 'jarom': 'Jarom', 'omni': 'Omni',
      'mosiah': 'Mosiah', 'alma': 'Alma', 'hel': 'Helaman',
      'morm': 'Mormon', 'ether': 'Ether', 'moro': 'Moroni',
    }
    const name = bookMap[bofmMatch[1]!] || bofmMatch[1]
    return `${name} ${bofmMatch[2]}`
  }
  // OT/NT
  const otntMatch = href.match(/scriptures\/(?:ot|nt)\/([\w-]+)\/(\d+)/)
  if (otntMatch) {
    const bookMap: Record<string, string> = {
      'gen': 'Genesis', 'ex': 'Exodus', 'lev': 'Leviticus', 'deut': 'Deuteronomy',
      'isa': 'Isaiah', 'jer': 'Jeremiah', 'ps': 'Psalms', 'prov': 'Proverbs',
      'matt': 'Matthew', 'mark': 'Mark', 'luke': 'Luke', 'john': 'John',
      'acts': 'Acts', 'rom': 'Romans', 'heb': 'Hebrews', 'james': 'James',
      'rev': 'Revelation', '1-cor': '1 Corinthians', '2-cor': '2 Corinthians',
      '1-tim': '1 Timothy', '2-tim': '2 Timothy', '1-pet': '1 Peter', '2-pet': '2 Peter',
    }
    const name = bookMap[otntMatch[1]!] || otntMatch[1]
    return `${name} ${otntMatch[2]}`
  }
  // PGP
  const pgpMatch = href.match(/pgp\/(moses|abr|js-h|a-of-f)\/(\d+)/)
  if (pgpMatch) {
    const bookMap: Record<string, string> = {
      'moses': 'Moses', 'abr': 'Abraham', 'js-h': 'JS—H', 'a-of-f': 'Articles of Faith',
    }
    return `${bookMap[pgpMatch[1]!] || pgpMatch[1]} ${pgpMatch[2]}`
  }
  return null
}

function openPracticeForm() {
  const text = selectionTrigger.value.text
  if (!text) return

  // Try to detect a scripture reference:
  // 1. First check the selected text itself
  let ref = detectScriptureRef(text)

  // 2. If not found, look at surrounding DOM context (nearby links, parent elements)
  if (!ref) {
    ref = detectRefFromSelectionContext()
  }

  // Build a sensible default name
  let name: string
  if (ref) {
    name = ref
  } else {
    const preview = text.substring(0, 40).replace(/\n/g, ' ')
    name = currentTitle.value
      ? `${currentTitle.value} — "${preview}${text.length > 40 ? '…' : ''}"`
      : `"${preview}${text.length > 40 ? '…' : ''}"`
  }

  // Default dates: start today, end in 1 week
  const today = new Date()
  const nextWeek = new Date(today)
  nextWeek.setDate(nextWeek.getDate() + 7)

  practiceForm.value = {
    name: name.substring(0, 120),
    description: text.substring(0, 2000),
    category: 'scripture',
    pillarIds: [],
    reps: 3,
    startDate: localDateStr(today),
    endDate: localDateStr(nextWeek),
  }
  practiceFormError.value = ''
  practiceFormSuccess.value = false
  showPracticeForm.value = true
  selectionTrigger.value.show = false
}

function closePracticeForm() {
  showPracticeForm.value = false
  practiceFormSuccess.value = false
  practiceFormError.value = ''
  window.getSelection()?.removeAllRanges()
}

function toggleFormCategory(cat: string) {
  if (practiceForm.value.category === cat) {
    practiceForm.value.category = ''
  } else {
    practiceForm.value.category = cat
  }
}

function toggleFormPillar(id: number) {
  const idx = practiceForm.value.pillarIds.indexOf(id)
  if (idx >= 0) {
    practiceForm.value.pillarIds.splice(idx, 1)
  } else {
    practiceForm.value.pillarIds.push(id)
  }
}

async function submitPracticeForm() {
  const f = practiceForm.value
  if (!f.name.trim()) {
    practiceFormError.value = 'Title is required'
    return
  }

  practiceFormSaving.value = true
  practiceFormError.value = ''
  try {
    const practice = await api.createPractice({
      name: f.name.trim(),
      type: 'memorize',
      description: f.description,
      source_doc: currentPath.value,
      source_path: currentPath.value,
      category: f.category || 'scripture',
      config: JSON.stringify({ target_daily_reps: f.reps || 1 }),
      start_date: f.startDate || undefined,
      end_date: f.endDate || undefined,
    })

    if (f.pillarIds.length > 0 && practice?.id) {
      await api.setPracticePillars(practice.id, f.pillarIds)
    }

    practiceFormSuccess.value = true
    setTimeout(() => {
      closePracticeForm()
    }, 1200)
  } catch (e: any) {
    practiceFormError.value = e.message || 'Failed to create practice'
  } finally {
    practiceFormSaving.value = false
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

onMounted(async () => {
  window.addEventListener('resize', handleResize)
  document.addEventListener('selectionchange', handleSelectionChange)
  handleResize()

  // Load pillars for the practice creation form (non-blocking)
  api.listPillarsFlat().then(p => { allPillars.value = p }).catch(() => {})

  // Load the source first, then check for deep-link file param
  await loadSource()
  const initialFile = route.query.f as string | undefined
  if (initialFile) {
    expandToPath(initialFile)
    await loadFile(initialFile)
    // Scroll the sidebar to the active file after Vue renders
    await nextTick()
    scrollSidebarToActive()
  }
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  document.removeEventListener('selectionchange', handleSelectionChange)
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
    <aside v-if="sidebarOpen" class="reader-sidebar" :class="{ 'mobile-sidebar': isMobile }" data-sidebar>
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

      <!-- Text Selection Trigger Button (fixed to viewport) -->
      <Transition name="popup-fade">
        <div
          v-if="selectionTrigger.show"
          class="selection-trigger"
          :style="{ left: selectionTrigger.x + 'px', top: selectionTrigger.y + 'px' }"
        >
          <button @click.stop="openPracticeForm" class="selection-trigger-btn">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" viewBox="0 0 20 20" fill="currentColor">
              <path d="M5 4a2 2 0 012-2h6a2 2 0 012 2v14l-5-2.5L5 18V4z" />
            </svg>
            Memorize
          </button>
          <div class="selection-trigger-arrow"></div>
        </div>
      </Transition>
    </main>

    <!-- Practice Creation Modal -->
    <Teleport to="body">
      <Transition name="modal-fade">
        <div v-if="showPracticeForm" class="practice-modal-overlay" @click.self="closePracticeForm">
          <div class="practice-modal" :class="{ dark: darkMode }">
            <!-- Success state -->
            <div v-if="practiceFormSuccess" class="practice-modal-success">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-10 w-10 text-emerald-500" viewBox="0 0 20 20" fill="currentColor">
                <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
              </svg>
              <p class="text-sm font-semibold mt-2" style="color: var(--text-primary, #111827)">Practice created!</p>
            </div>

            <!-- Form -->
            <template v-else>
              <div class="flex items-center justify-between mb-4">
                <h3 class="text-base font-semibold" style="color: var(--text-primary, #111827)">
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 inline-block mr-1 -mt-0.5 text-amber-500" viewBox="0 0 20 20" fill="currentColor">
                    <path d="M5 4a2 2 0 012-2h6a2 2 0 012 2v14l-5-2.5L5 18V4z" />
                  </svg>
                  New Memorize Practice
                </h3>
                <button @click="closePracticeForm" class="p-1 rounded hover:bg-gray-200 dark:hover:bg-gray-700" style="color: var(--text-muted, #9ca3af)">
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
                    <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd" />
                  </svg>
                </button>
              </div>

              <!-- Error -->
              <div v-if="practiceFormError" class="text-xs text-red-600 bg-red-50 dark:bg-red-900/20 px-3 py-2 rounded mb-3">
                {{ practiceFormError }}
              </div>

              <div class="space-y-3">
                <!-- Title -->
                <div>
                  <label class="practice-form-label">Title</label>
                  <input
                    v-model="practiceForm.name"
                    type="text"
                    class="practice-form-input"
                    placeholder="Practice name"
                    maxlength="120"
                  />
                </div>

                <!-- Text / Description -->
                <div>
                  <label class="practice-form-label">Text</label>
                  <textarea
                    v-model="practiceForm.description"
                    class="practice-form-input practice-form-textarea"
                    rows="4"
                    maxlength="2000"
                    placeholder="Selected text"
                  />
                </div>

                <!-- Category -->
                <div>
                  <label class="practice-form-label">Category</label>
                  <div class="flex flex-wrap gap-1.5">
                    <button
                      v-for="cat in presetCategories"
                      :key="cat"
                      @click="toggleFormCategory(cat)"
                      class="practice-form-chip"
                      :class="{ active: practiceForm.category === cat }"
                    >
                      {{ cat }}
                    </button>
                  </div>
                </div>

                <!-- Pillars -->
                <div v-if="allPillars.length > 0">
                  <label class="practice-form-label">Pillars</label>
                  <div class="flex flex-wrap gap-1.5">
                    <button
                      v-for="p in allPillars"
                      :key="p.id"
                      @click="toggleFormPillar(p.id)"
                      class="practice-form-chip"
                      :class="{ active: practiceForm.pillarIds.includes(p.id) }"
                    >
                      {{ p.icon }} {{ p.name }}
                    </button>
                  </div>
                </div>

                <!-- Reps + Dates row -->
                <div class="grid grid-cols-3 gap-2">
                  <div>
                    <label class="practice-form-label">Daily Reps</label>
                    <input
                      v-model.number="practiceForm.reps"
                      type="number"
                      min="1"
                      max="20"
                      class="practice-form-input text-center"
                    />
                  </div>
                  <div>
                    <label class="practice-form-label">Start</label>
                    <input
                      v-model="practiceForm.startDate"
                      type="date"
                      class="practice-form-input"
                    />
                  </div>
                  <div>
                    <label class="practice-form-label">End</label>
                    <input
                      v-model="practiceForm.endDate"
                      type="date"
                      class="practice-form-input"
                    />
                  </div>
                </div>
              </div>

              <!-- Actions -->
              <div class="flex justify-end gap-2 mt-4 pt-3" style="border-top: 1px solid var(--border-color, #e5e7eb)">
                <button
                  @click="closePracticeForm"
                  class="px-3 py-1.5 text-xs font-medium rounded-md hover:bg-gray-100 dark:hover:bg-gray-700"
                  style="color: var(--text-secondary, #6b7280)"
                >
                  Cancel
                </button>
                <button
                  @click="submitPracticeForm"
                  :disabled="practiceFormSaving"
                  class="px-4 py-1.5 text-xs font-semibold rounded-md text-white bg-amber-600 hover:bg-amber-700 disabled:opacity-50 disabled:cursor-wait"
                >
                  {{ practiceFormSaving ? 'Creating...' : 'Create Practice' }}
                </button>
              </div>
            </template>
          </div>
        </div>
      </Transition>
    </Teleport>

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

/* --- Text Selection Trigger --- */

.selection-trigger {
  position: fixed;
  transform: translateX(-50%) translateY(-100%);
  z-index: 40;
  background: #1f2937;
  border-radius: 6px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
  padding: 2px;
  white-space: nowrap;
  pointer-events: auto;
}

.selection-trigger-btn {
  display: flex;
  align-items: center;
  gap: 5px;
  padding: 5px 10px;
  font-size: 0.75rem;
  font-weight: 600;
  color: #fbbf24;
  background: none;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  transition: background 0.15s;
}

.selection-trigger-btn:hover {
  background: rgba(251, 191, 36, 0.15);
}

.selection-trigger-arrow {
  position: absolute;
  bottom: -5px;
  left: 50%;
  transform: translateX(-50%);
  width: 0;
  height: 0;
  border-left: 6px solid transparent;
  border-right: 6px solid transparent;
  border-top: 6px solid #1f2937;
}

/* --- Practice Creation Modal --- */

.practice-modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.4);
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding-top: 10vh;
  z-index: 50;
}

.practice-modal {
  width: 420px;
  max-width: 95vw;
  max-height: 80vh;
  overflow-y: auto;
  background: var(--bg-surface, white);
  border: 1px solid var(--border-color, #e5e7eb);
  border-radius: 12px;
  padding: 20px 24px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.15);
}

.practice-modal-success {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 24px 0;
}

.practice-form-label {
  display: block;
  font-size: 0.7rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.03em;
  margin-bottom: 4px;
  color: var(--text-muted, #9ca3af);
}

.practice-form-input {
  width: 100%;
  padding: 6px 10px;
  font-size: 0.8rem;
  border: 1px solid var(--border-color, #d1d5db);
  border-radius: 6px;
  background: var(--bg-page, #f9fafb);
  color: var(--text-primary, #111827);
  outline: none;
  transition: border-color 0.15s;
}

.practice-form-input:focus {
  border-color: #f59e0b;
  box-shadow: 0 0 0 2px rgba(245, 158, 11, 0.15);
}

.practice-form-textarea {
  resize: vertical;
  min-height: 60px;
  line-height: 1.5;
}

.practice-form-chip {
  padding: 3px 10px;
  font-size: 0.7rem;
  font-weight: 500;
  border: 1px solid var(--border-color, #d1d5db);
  border-radius: 999px;
  background: transparent;
  color: var(--text-secondary, #6b7280);
  cursor: pointer;
  transition: all 0.15s;
}

.practice-form-chip:hover {
  border-color: #f59e0b;
  color: #d97706;
}

.practice-form-chip.active {
  background: #fffbeb;
  border-color: #f59e0b;
  color: #b45309;
  font-weight: 600;
}

/* Dark mode overrides for the modal */
.practice-modal.dark .practice-form-chip.active {
  background: rgba(245, 158, 11, 0.15);
  color: #fbbf24;
}

/* Popup trigger transition */
.popup-fade-enter-active,
.popup-fade-leave-active {
  transition: opacity 0.15s, transform 0.15s;
}
.popup-fade-enter-from,
.popup-fade-leave-to {
  opacity: 0;
  transform: translateX(-50%) translateY(-100%) scale(0.9);
}

/* Modal transition */
.modal-fade-enter-active,
.modal-fade-leave-active {
  transition: opacity 0.2s ease;
}
.modal-fade-enter-active .practice-modal,
.modal-fade-leave-active .practice-modal {
  transition: transform 0.2s ease, opacity 0.2s ease;
}
.modal-fade-enter-from,
.modal-fade-leave-to {
  opacity: 0;
}
.modal-fade-enter-from .practice-modal,
.modal-fade-leave-to .practice-modal {
  transform: translateY(-10px) scale(0.98);
  opacity: 0;
}
</style>
