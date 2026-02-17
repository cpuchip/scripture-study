<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import MarkdownIt from 'markdown-it'

export interface ReferenceTab {
  id: string
  title: string
  path: string       // display path or URL
  content: string    // markdown content (for .md files from repo)
  url: string        // external URL to load in iframe (for scripture links)
  loading: boolean
  error: string
  scrollRef?: string  // verse/section anchor from the original link
}

const props = defineProps<{
  tabs: ReferenceTab[]
  activeTabId: string
  authenticated: boolean
  memorizeLoading: boolean
}>()

const emit = defineEmits<{
  (e: 'close-tab', id: string): void
  (e: 'close-all'): void
  (e: 'select-tab', id: string): void
  (e: 'add-to-memorize', tab: ReferenceTab): void
  (e: 'close-panel'): void
  (e: 'link-click', href: string, event: MouseEvent): void
}>()

const md = new MarkdownIt({
  html: true,
  linkify: true,
  typographer: true,
  breaks: false,
})

// External links open in new tabs
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

const activeTab = computed(() => props.tabs.find(t => t.id === props.activeTabId))

const renderedContent = computed(() => {
  if (!activeTab.value?.content) return ''
  return md.render(activeTab.value.content)
})

// Detect if content looks like a scripture (has verse numbers)
const isScripture = computed(() => {
  if (!activeTab.value) return false
  // Iframe tabs pointing to church scriptures are always scripture
  if (activeTab.value.url && /churchofjesuschrist\.org\/study\/scriptures/i.test(activeTab.value.url)) return true
  if (!activeTab.value.content) return false
  const content = activeTab.value.content
  return /^\d+\s/m.test(content) || /gospel-library.*scriptures/i.test(activeTab.value.path)
})

// Whether the active tab uses an iframe (external URL)
const isIframe = computed(() => !!activeTab.value?.url)

// Iframe loading state
const iframeLoading = ref(false)

function onIframeLoad() {
  iframeLoading.value = false
}

watch(() => props.activeTabId, () => {
  if (activeTab.value?.url) {
    iframeLoading.value = true
  }
})


function handlePanelClick(event: MouseEvent) {
  const target = event.target as HTMLElement
  const link = target.closest('a')
  if (!link) return

  const href = link.getAttribute('href')
  if (!href) return

  // Skip external links (they open in new tabs)
  if (href.startsWith('http://') || href.startsWith('https://')) return

  // Delegate link handling to parent
  event.preventDefault()
  emit('link-click', href, event)
}

const memorizeAdded = ref(false)
const memorizeError = ref('')

async function handleAddToMemorize() {
  if (!activeTab.value) return
  memorizeAdded.value = false
  memorizeError.value = ''
  emit('add-to-memorize', activeTab.value)
}

watch(() => props.memorizeLoading, (loading) => {
  if (!loading) {
    memorizeAdded.value = true
    setTimeout(() => { memorizeAdded.value = false }, 3000)
  }
})
</script>

<template>
  <div class="reference-panel" @click="handlePanelClick">
    <!-- Header -->
    <div class="ref-header">
      <div class="flex items-center gap-2 min-w-0 flex-1">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-orange-500 shrink-0" viewBox="0 0 20 20" fill="currentColor">
          <path d="M9 4.804A7.968 7.968 0 005.5 4c-1.255 0-2.443.29-3.5.804v10A7.969 7.969 0 015.5 14c1.669 0 3.218.51 4.5 1.385A7.962 7.962 0 0114.5 14c1.255 0 2.443.29 3.5.804v-10A7.968 7.968 0 0014.5 4c-1.255 0-2.443.29-3.5.804V15" />
        </svg>
        <span class="text-xs font-semibold text-gray-600 truncate">References</span>
      </div>
      <div class="flex items-center gap-1">
        <button
          v-if="tabs.length > 1"
          @click="emit('close-all')"
          class="text-gray-400 hover:text-red-500 p-1 text-xs"
          title="Close all tabs"
        >
          Close All
        </button>
        <button @click="emit('close-panel')" class="text-gray-400 hover:text-gray-600 p-1" title="Close panel">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
            <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd" />
          </svg>
        </button>
      </div>
    </div>

    <!-- Tabs -->
    <div v-if="tabs.length > 1" class="ref-tabs">
      <div
        v-for="tab in tabs"
        :key="tab.id"
        class="ref-tab"
        :class="{ active: tab.id === activeTabId }"
        @click="emit('select-tab', tab.id)"
      >
        <span class="truncate text-xs">{{ tab.title }}</span>
        <button
          @click.stop="emit('close-tab', tab.id)"
          class="shrink-0 text-gray-400 hover:text-red-500"
        >
          <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" viewBox="0 0 20 20" fill="currentColor">
            <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd" />
          </svg>
        </button>
      </div>
    </div>

    <!-- Content -->
    <div class="ref-content">
      <!-- Loading (markdown mode) -->
      <div v-if="activeTab?.loading && !isIframe" class="flex items-center justify-center h-32">
        <div class="text-gray-400 text-sm">Loading reference...</div>
      </div>

      <!-- Error -->
      <div v-else-if="activeTab?.error" class="p-4">
        <div class="text-red-500 text-sm">{{ activeTab.error }}</div>
      </div>

      <!-- Iframe mode (external church content) -->
      <div v-else-if="isIframe && activeTab?.url" class="ref-iframe-wrapper">
        <!-- Action bar -->
        <div class="ref-actions">
          <a
            :href="activeTab.url"
            target="_blank"
            rel="noopener noreferrer"
            class="text-xs text-orange-600 hover:text-orange-700 truncate flex-1 underline"
          >
            Open on churchofjesuschrist.org ↗
          </a>
        </div>

        <!-- Loading indicator for iframe -->
        <div v-if="iframeLoading" class="flex items-center justify-center py-8">
          <div class="text-gray-400 text-sm">Loading from churchofjesuschrist.org...</div>
        </div>

        <iframe
          :src="activeTab.url"
          class="ref-iframe"
          :class="{ 'opacity-0': iframeLoading }"
          @load="onIframeLoad"
          sandbox="allow-same-origin allow-scripts allow-popups"
          referrerpolicy="no-referrer"
          title="Scripture Reference"
        />
      </div>

      <!-- Markdown content (repo files) -->
      <div v-else-if="activeTab?.content" class="ref-document">
        <!-- Action bar -->
        <div class="ref-actions">
          <span class="text-xs text-gray-400 font-mono truncate flex-1">{{ activeTab.path }}</span>
          <div class="flex items-center gap-2 shrink-0">
            <!-- Add to memorize (authenticated only, scripture content) -->
            <button
              v-if="authenticated && isScripture"
              @click="handleAddToMemorize"
              :disabled="memorizeLoading"
              class="ref-memorize-btn"
              :class="memorizeAdded ? 'text-green-600' : 'text-orange-600 hover:text-orange-700'"
              :title="memorizeAdded ? 'Added!' : 'Add to memorize'"
            >
              <svg v-if="memorizeAdded" xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" viewBox="0 0 20 20" fill="currentColor">
                <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
              </svg>
              <svg v-else xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" viewBox="0 0 20 20" fill="currentColor">
                <path d="M5 4a2 2 0 012-2h6a2 2 0 012 2v14l-5-2.5L5 18V4z" />
              </svg>
              <span class="text-xs">{{ memorizeAdded ? 'Added' : memorizeLoading ? '...' : 'Memorize' }}</span>
            </button>
          </div>
        </div>

        <!-- Markdown output -->
        <article class="prose prose-sm prose-gray max-w-none" v-html="renderedContent" />
      </div>

      <!-- Empty -->
      <div v-else class="flex items-center justify-center h-32 text-gray-400 text-sm">
        No reference selected
      </div>
    </div>
  </div>
</template>

<style scoped>
.reference-panel {
  width: 400px;
  min-width: 320px;
  max-width: 500px;
  border-left: 1px solid var(--border-color, #e5e7eb);
  background: var(--bg-surface, white);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.ref-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  border-bottom: 1px solid var(--border-color, #e5e7eb);
  background: var(--bg-surface, white);
}

.ref-tabs {
  display: flex;
  overflow-x: auto;
  border-bottom: 1px solid var(--border-color, #e5e7eb);
  background: var(--bg-muted, #f9fafb);
}

.ref-tab {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 6px 10px;
  cursor: pointer;
  border-bottom: 2px solid transparent;
  max-width: 180px;
  white-space: nowrap;
  color: var(--text-secondary, #6b7280);
}

.ref-tab:hover {
  background: var(--bg-hover, #f3f4f6);
}

.ref-tab.active {
  border-bottom-color: #ea580c;
  color: var(--text-primary, #111827);
  background: var(--bg-surface, white);
}

.ref-content {
  flex: 1;
  overflow-y: auto;
  background: var(--bg-page, #fafafa);
}

.ref-iframe-wrapper {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-height: 0;
}

.ref-iframe {
  flex: 1;
  width: 100%;
  border: none;
  min-height: 400px;
}

.ref-document {
  padding: 16px 20px;
}

.ref-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
  padding-bottom: 8px;
  border-bottom: 1px solid var(--border-color, #e5e7eb);
}

.ref-memorize-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  border: 1px solid currentColor;
  border-radius: 4px;
  font-size: 0.75rem;
  white-space: nowrap;
  transition: all 0.15s;
}

.ref-memorize-btn:hover:not(:disabled) {
  background: #fff7ed;
}

.ref-memorize-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Prose overrides for the panel — slightly smaller than main reader */
.ref-document :deep(h1) { font-size: 1.35rem; font-weight: 700; margin-bottom: 0.75rem; color: var(--text-primary, #111827); border-bottom: 1px solid var(--border-color, #e5e7eb); padding-bottom: 0.4rem; }
.ref-document :deep(h2) { font-size: 1.15rem; font-weight: 600; margin-top: 1.5rem; margin-bottom: 0.5rem; color: var(--text-primary, #1f2937); }
.ref-document :deep(h3) { font-size: 1rem; font-weight: 600; margin-top: 1rem; margin-bottom: 0.4rem; color: var(--text-secondary, #374151); }
.ref-document :deep(p) { line-height: 1.7; margin-bottom: 0.75rem; color: var(--text-body, #374151); font-size: 0.875rem; }
.ref-document :deep(blockquote) { border-left: 3px solid #d97706; padding: 0.4rem 0.75rem; margin: 0.75rem 0; background: var(--bg-quote, #fffbeb); border-radius: 0 4px 4px 0; }
.ref-document :deep(blockquote p) { color: #92400e; margin-bottom: 0.2rem; }
.ref-document :deep(a) { color: #ea580c; text-decoration: underline; text-decoration-color: #fed7aa; text-underline-offset: 2px; }
.ref-document :deep(a:hover) { color: #c2410c; text-decoration-color: #ea580c; }
.ref-document :deep(code) { background: var(--bg-code, #f3f4f6); padding: 0.1rem 0.3rem; border-radius: 3px; font-size: 0.8em; }
.ref-document :deep(pre) { background: #1f2937; color: #e5e7eb; padding: 0.75rem; border-radius: 6px; overflow-x: auto; margin: 0.75rem 0; font-size: 0.8rem; }
.ref-document :deep(pre code) { background: none; color: inherit; padding: 0; }
.ref-document :deep(ul), .ref-document :deep(ol) { padding-left: 1.25rem; margin-bottom: 0.75rem; }
.ref-document :deep(li) { line-height: 1.7; margin-bottom: 0.2rem; color: var(--text-body, #374151); font-size: 0.875rem; }
.ref-document :deep(table) { width: 100%; border-collapse: collapse; margin: 0.75rem 0; font-size: 0.8rem; }
.ref-document :deep(th), .ref-document :deep(td) { border: 1px solid var(--border-color, #e5e7eb); padding: 0.4rem 0.6rem; text-align: left; }
.ref-document :deep(th) { background: var(--bg-muted, #f9fafb); font-weight: 600; }
.ref-document :deep(hr) { border: none; border-top: 1px solid var(--border-color, #e5e7eb); margin: 1.5rem 0; }

/* Mobile: overlay mode */
@media (max-width: 768px) {
  .reference-panel {
    position: fixed;
    top: 0;
    right: 0;
    bottom: 0;
    width: 100%;
    max-width: 100%;
    z-index: 40;
    box-shadow: -4px 0 12px rgba(0, 0, 0, 0.15);
  }
}
</style>
