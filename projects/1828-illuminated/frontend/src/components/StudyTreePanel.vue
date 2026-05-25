<script setup lang="ts">
// StudyTreePanel — slide-out drawer that shows the branching study tree.
//
// Available on every page where study happens (mounted in App.vue as a
// fixed-position overlay). Toggles open/closed via the panelOpen ref
// in useStudyTree. Recursion handled via TreeBranch.vue (recursive SFC).

import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import {
  useStudyTree,
  exportTreeMarkdown,
  panelPinned,
  type StudyNode,
} from '@/composables/useStudyTree'
import TreeBranch from './TreeBranch.vue'

const signInUrl = computed(() => {
  const currentUrl = typeof window !== 'undefined' ? window.location.href : 'https://1828.ibeco.me/'
  return `https://ibeco.me/login?redirect=${encodeURIComponent(currentUrl)}`
})

// `inline=true` when mounted as the second column of the page layout
// (pinned mode). `inline=false`/omitted when mounted as a viewport-fixed
// overlay drawer (unpinned mode). The styling + Teleport behavior switch
// on this prop.
const props = defineProps<{ inline?: boolean }>()

const router = useRouter()
const {
  roots,
  childrenOf,
  activeNodeId,
  navigateTo,
  removeSubtree,
  clearAll,
  nodeCount,
  panelOpen,
  session,
  saveTreeToCloud,
  fetchCloudTrees,
  loadTreeFromCloud,
} = useStudyTree()

function togglePin() {
  panelPinned.value = !panelPinned.value
}

const confirmClear = ref(false)
const copiedHint = ref(false)

function jumpTo(node: StudyNode) {
  navigateTo(node.id)
  if (node.kind === 'word' && node.payload.kind === 'word') {
    router.push({ name: 'word-detail', params: { word: node.payload.word } })
  } else if (node.kind === 'chapter' && node.payload.kind === 'chapter') {
    const [book, chapter] = node.payload.abbrRef.split('/')
    if (book && chapter) {
      router.push({
        name: 'verse-explorer',
        query: {
          mode: 'canon',
          v: bookToVolume(book),
          b: book,
          c: chapter,
          ...(node.payload.range ? { r: node.payload.range } : {}),
        },
      })
    }
  } else if (node.kind === 'verse' && node.payload.kind === 'verse') {
    const [book, chapter] = node.payload.abbrRef.split('/')
    if (book && chapter) {
      router.push({
        name: 'verse-explorer',
        query: {
          mode: 'canon',
          v: bookToVolume(book),
          b: book,
          c: chapter,
          r: String(node.payload.verse),
        },
      })
    }
  }
  // render + note nodes don't route — they're viewing-only inside the tree.
}

function bookToVolume(abbr: string): string {
  const OT = ['gen','ex','lev','num','deut','josh','judg','ruth','1-sam','2-sam','1-kgs','2-kgs','1-chr','2-chr','ezra','neh','esth','job','ps','prov','eccl','song','isa','jer','lam','ezek','dan','hosea','joel','amos','obad','jonah','micah','nahum','hab','zeph','hag','zech','mal']
  const NT = ['matt','mark','luke','john','acts','rom','1-cor','2-cor','gal','eph','philip','col','1-thes','2-thes','1-tim','2-tim','titus','philem','heb','james','1-pet','2-pet','1-jn','2-jn','3-jn','jude','rev']
  const BoM = ['1-ne','2-ne','jacob','enos','jarom','omni','w-of-m','mosiah','alma','hel','3-ne','4-ne','morm','ether','moro']
  const PGP = ['moses','abr','js-m','js-h','a-of-f']
  if (OT.includes(abbr)) return 'ot'
  if (NT.includes(abbr)) return 'nt'
  if (BoM.includes(abbr)) return 'bofm'
  if (abbr === 'dc') return 'dc'
  if (PGP.includes(abbr)) return 'pgp'
  return 'bofm'
}

function exportToClipboard() {
  const md = exportTreeMarkdown()
  navigator.clipboard?.writeText(md).then(() => {
    copiedHint.value = true
    setTimeout(() => { copiedHint.value = false }, 1500)
  }).catch(() => { /* clipboard blocked; silent */ })
}

function onClearClick() {
  if (!confirmClear.value) {
    confirmClear.value = true
    setTimeout(() => { confirmClear.value = false }, 4000)
    return
  }
  clearAll()
  confirmClear.value = false
}

const saveTitle = ref('')
const cloudTrees = ref<any[]>([])
const showSaveModal = ref(false)
const showLoadList = ref(false)
const cloudSaveSuccess = ref(false)

function openSaveModal() {
  saveTitle.value = 'My Study Path - ' + new Date().toLocaleDateString()
  showSaveModal.value = true
}

async function handleCloudSave() {
  if (!saveTitle.value.trim()) return
  const result = await saveTreeToCloud(saveTitle.value)
  if (result) {
    cloudSaveSuccess.value = true
    setTimeout(() => {
      cloudSaveSuccess.value = false
      showSaveModal.value = false
    }, 1500)
    await refreshCloudTrees()
  }
}

async function refreshCloudTrees() {
  cloudTrees.value = await fetchCloudTrees()
}

async function toggleLoadList() {
  showLoadList.value = !showLoadList.value
  if (showLoadList.value) {
    await refreshCloudTrees()
  }
}

async function handleLoadTree(treeId: string) {
  const success = await loadTreeFromCloud(treeId)
  if (success) {
    showLoadList.value = false
  }
}
</script>

<template>
  <Teleport to="body" :disabled="!!props.inline">
    <!-- Backdrop (only when overlay mode AND on mobile; pinned/inline doesn't dim) -->
    <div
      v-if="panelOpen && !panelPinned && !props.inline"
      class="fixed inset-0 bg-black/20 z-30 lg:hidden"
      @click="panelOpen = false"
    />

    <!-- The panel. Three render modes:
         - INLINE (props.inline=true): mounted by App.vue as the second column of
           the centered layout when pinned. Teleport disabled. Sticky-positioned
           so it scrolls with the page but stays visible. No translate animation.
         - OVERLAY DRAWER (default, panelOpen=true): Teleport'd to body, full-height
           slide-out from the right edge of the viewport, z-40.
         - HIDDEN: panelOpen=false → slid off-screen via translate-x-full. -->
    <aside
      :class="[
        'bg-[var(--paper)] border border-stone-300 flex flex-col',
        props.inline
          ? 'rounded-lg shadow-sm sticky top-6 max-h-[calc(100vh-3rem)] w-full'
          : [
              'fixed right-0 w-full sm:w-96 shadow-2xl transition-transform duration-200 ease-out border-l',
              'top-0 h-full z-40',
              panelOpen ? 'translate-x-0' : 'translate-x-full',
            ],
      ]"
      aria-label="Study tree panel"
    >
      <header class="px-4 py-3 border-b border-stone-300 flex items-baseline justify-between gap-2">
        <div>
          <h2 class="font-serif text-lg">Study tree</h2>
          <p class="text-xs text-stone-500">{{ nodeCount }} node{{ nodeCount === 1 ? '' : 's' }}</p>
        </div>
        <div class="flex items-center gap-1">
          <button
            @click="togglePin"
            :class="[
              'text-stone-400 hover:text-stone-900 text-base leading-none px-1.5 py-1 rounded transition',
              panelPinned ? 'bg-amber-100 text-amber-800 hover:bg-amber-200' : 'hover:bg-stone-100',
            ]"
            :title="panelPinned ? 'Unpin (close to a drawer that overlays content)' : 'Pin (anchor below header so it stays put + nav stays reachable)'"
            :aria-label="panelPinned ? 'Unpin study tree panel' : 'Pin study tree panel'"
          >📌</button>
          <button
            @click="panelOpen = false"
            class="text-stone-400 hover:text-stone-900 text-2xl leading-none px-1"
            aria-label="Close study tree panel"
          >×</button>
        </div>
      </header>

      <!-- Tree body -->
      <div class="flex-1 overflow-y-auto px-2 py-3 text-sm">
        <div v-if="!roots.length" class="text-stone-500 italic px-2 py-8 text-center">
          Empty. Click any word, verse, or chapter — it'll appear here. Clicks from there branch the tree.
        </div>

        <ul v-else class="space-y-1">
          <TreeBranch
            v-for="root in roots"
            :key="root.id"
            :node="root"
            :children-of="childrenOf"
            :active-id="activeNodeId"
            @jump="jumpTo"
            @remove="removeSubtree"
          />
        </ul>
      </div>

      <!-- Cloud Sync Section -->
      <div class="px-4 py-2.5 border-t border-stone-200 bg-amber-50/10 text-xs">
        <div v-if="session.authenticated" class="flex flex-col gap-2">
          <div class="flex items-center justify-between">
            <span class="text-stone-500 font-mono">Cloud Account</span>
            <button @click="toggleLoadList" class="text-amber-700 hover:text-amber-900 font-medium select-none">
              {{ showLoadList ? 'Hide Cloud' : 'Load Tree ▾' }}
            </button>
          </div>
          <button
            @click="openSaveModal"
            class="w-full px-3 py-1.5 bg-amber-700/80 text-white rounded text-center hover:bg-amber-800 transition font-medium"
            :disabled="!nodeCount"
          >
            Save Tree to Cloud
          </button>
          
          <!-- Cloud Trees list -->
          <div v-if="showLoadList" class="mt-2 border-t border-stone-200 pt-2 space-y-1 max-h-32 overflow-y-auto">
            <div v-if="!cloudTrees.length" class="text-stone-400 italic py-1 text-center">No saved trees on cloud.</div>
            <button
              v-for="tree in cloudTrees"
              :key="tree.id"
              @click="handleLoadTree(tree.id)"
              class="w-full text-left px-2 py-1 rounded hover:bg-amber-100/50 flex justify-between text-stone-700 truncate"
            >
              <span class="font-serif truncate mr-2">{{ tree.title }}</span>
              <span class="text-[10px] text-stone-400 shrink-0">{{ new Date(tree.updated_at).toLocaleDateString() }}</span>
            </button>
          </div>
        </div>
        <div v-else class="text-stone-500 italic text-center py-1">
          <a :href="signInUrl" class="underline text-amber-700 font-medium">Sign in with Becoming</a> to sync your trees.
        </div>
      </div>

      <!-- Footer actions -->
      <footer class="px-4 py-3 border-t border-stone-300 flex items-center justify-between gap-2 text-xs">
        <button
          @click="exportToClipboard"
          class="px-3 py-1.5 rounded border border-stone-300 hover:border-amber-500 hover:bg-amber-50 transition"
          :disabled="!nodeCount"
          :class="!nodeCount ? 'opacity-40 cursor-not-allowed' : ''"
        >
          <span v-if="copiedHint" class="text-amber-700 font-semibold">✓ copied as markdown</span>
          <span v-else>Export ↻</span>
        </button>
        <button
          @click="onClearClick"
          class="px-3 py-1.5 rounded border transition"
          :class="confirmClear
            ? 'border-red-400 bg-red-50 text-red-700 font-medium'
            : 'border-stone-300 text-stone-500 hover:border-stone-400'"
          :disabled="!nodeCount"
        >
          <span v-if="confirmClear">Confirm clear {{ nodeCount }}</span>
          <span v-else>Start fresh</span>
        </button>
      </footer>
    </aside>

    <!-- Save Modal overlay -->
    <div v-if="showSaveModal" class="fixed inset-0 bg-black/60 z-50 flex items-center justify-center p-4" @click.self="showSaveModal = false">
      <div class="bg-white rounded-lg p-6 max-w-sm w-full border border-stone-300 shadow-xl space-y-4">
        <h3 class="font-serif text-lg font-bold text-stone-850">Save Study Tree</h3>
        <p class="text-xs text-stone-500">Give your study tree a name to sync it to your cloud account.</p>
        <input
          v-model="saveTitle"
          type="text"
          class="w-full px-3 py-2 border border-stone-300 rounded text-sm focus:outline-none focus:border-amber-500 font-serif"
          placeholder="Study tree title"
          @keydown.enter="handleCloudSave"
        />
        <div class="flex justify-end gap-2 text-xs">
          <button @click="showSaveModal = false" class="px-3 py-1.5 border rounded hover:bg-stone-50">Cancel</button>
          <button
            @click="handleCloudSave"
            class="px-3 py-1.5 bg-amber-700 text-white rounded hover:bg-amber-800 transition"
            :disabled="!saveTitle.trim()"
          >
            <span v-if="cloudSaveSuccess">Saved!</span>
            <span v-else>Save</span>
          </button>
        </div>
      </div>
    </div>

    <!-- Floating toggle pill (visible when panel is closed) -->
    <button
      v-if="!panelOpen"
      @click="panelOpen = true"
      class="fixed bottom-6 right-6 z-30 px-4 py-2 rounded-full bg-stone-800 text-white text-xs font-medium shadow-lg hover:bg-amber-700 transition flex items-center gap-2"
      :title="`Study tree (${nodeCount} node${nodeCount === 1 ? '' : 's'})`"
      aria-label="Open study tree"
    >
      <span class="inline-block w-2 h-2 rounded-full bg-amber-400"></span>
      Study tree
      <span v-if="nodeCount" class="ml-1 px-1.5 py-0.5 rounded-full bg-amber-600 text-[10px]">{{ nodeCount }}</span>
    </button>
  </Teleport>
</template>
