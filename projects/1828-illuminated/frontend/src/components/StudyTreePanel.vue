<script setup lang="ts">
// StudyTreePanel — slide-out drawer that shows the branching study tree.
//
// Available on every page where study happens (mounted in App.vue as a
// fixed-position overlay). Toggles open/closed via the panelOpen ref
// in useStudyTree. Recursion handled via TreeBranch.vue (recursive SFC).

import { ref } from 'vue'
import { useRouter } from 'vue-router'
import {
  useStudyTree,
  exportTreeMarkdown,
  type StudyNode,
} from '@/composables/useStudyTree'
import TreeBranch from './TreeBranch.vue'

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
} = useStudyTree()

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
</script>

<template>
  <Teleport to="body">
    <!-- Backdrop (mobile: tap to close) -->
    <div
      v-if="panelOpen"
      class="fixed inset-0 bg-black/20 z-30 lg:hidden"
      @click="panelOpen = false"
    />

    <!-- Slide-out drawer -->
    <aside
      :class="[
        'fixed top-0 right-0 h-full w-full sm:w-96 bg-[var(--paper)] border-l border-stone-300 shadow-2xl z-40 transition-transform duration-200 ease-out flex flex-col',
        panelOpen ? 'translate-x-0' : 'translate-x-full',
      ]"
      aria-label="Study tree panel"
    >
      <header class="px-4 py-3 border-b border-stone-300 flex items-baseline justify-between gap-2">
        <div>
          <h2 class="font-serif text-lg">Study tree</h2>
          <p class="text-xs text-stone-500">{{ nodeCount }} node{{ nodeCount === 1 ? '' : 's' }}</p>
        </div>
        <button
          @click="panelOpen = false"
          class="text-stone-400 hover:text-stone-900 text-2xl leading-none"
          aria-label="Close study tree panel"
        >×</button>
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

      <!-- Footer actions -->
      <footer class="px-4 py-3 border-t border-stone-300 flex items-center justify-between gap-2 text-xs">
        <button
          @click="exportToClipboard"
          class="px-3 py-1.5 rounded border border-stone-300 hover:border-amber-500 hover:bg-amber-50 transition"
          :disabled="!nodeCount"
          :class="!nodeCount ? 'opacity-40 cursor-not-allowed' : ''"
        >
          <span v-if="copiedHint" class="text-amber-700">✓ copied as markdown</span>
          <span v-else>Export ↻</span>
        </button>
        <button
          @click="onClearClick"
          class="px-3 py-1.5 rounded border transition"
          :class="confirmClear
            ? 'border-red-400 bg-red-50 text-red-700'
            : 'border-stone-300 text-stone-500 hover:border-stone-400'"
          :disabled="!nodeCount"
        >
          <span v-if="confirmClear">Confirm clear {{ nodeCount }}</span>
          <span v-else>Start fresh</span>
        </button>
      </footer>
    </aside>

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
