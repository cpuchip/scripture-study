<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, computed } from 'vue'
import { useRouter } from 'vue-router'
import {
  useStudyTree,
  childrenOf,
  navigateTo,
  type StudyNode,
} from '@/composables/useStudyTree'

const router = useRouter()
const { activePath, activeNodeId, roots } = useStudyTree()

const openDropdownId = ref<string | null>(null)

function toggleDropdown(id: string, event: MouseEvent) {
  event.stopPropagation()
  if (openDropdownId.value === id) {
    openDropdownId.value = null
  } else {
    openDropdownId.value = id
  }
}

function closeAllDropdowns() {
  openDropdownId.value = null
}

onMounted(() => {
  window.addEventListener('click', closeAllDropdowns)
})
onBeforeUnmount(() => {
  window.removeEventListener('click', closeAllDropdowns)
})

function jumpTo(node: StudyNode) {
  navigateTo(node.id)
  closeAllDropdowns()
  
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

const KIND_GLYPH: Record<StudyNode['kind'], string> = {
  word: '📚', verse: '📖', chapter: '§', render: '↻', note: '✎',
}

const KIND_COLOR: Record<StudyNode['kind'], string> = {
  word: 'text-amber-700',
  verse: 'text-sky-700',
  chapter: 'text-sky-800',
  render: 'text-violet-700',
  note: 'text-stone-600',
}

// Helper to determine sibling nodes at each level
function getSiblings(node: StudyNode, index: number): StudyNode[] {
  if (index === 0) {
    return roots.value
  }
  const parent = activePath.value[index - 1]
  if (!parent) return []
  return childrenOf(parent.id)
}
</script>

<template>
  <div v-if="activePath.length" class="bg-[var(--paper-2)] border border-stone-200 rounded-lg px-4 py-2 text-xs flex items-center gap-2 overflow-x-auto select-none shadow-sm mb-6">
    <span class="text-stone-500 font-mono uppercase tracking-wider font-semibold">Study Path:</span>
    <div class="flex items-center gap-1.5 whitespace-nowrap">
      <div v-for="(node, index) in activePath" :key="node.id" class="flex items-center gap-1.5">
        <!-- Arrow connector between breadcrumbs -->
        <span v-if="index > 0" class="text-stone-400">➔</span>
        
        <!-- Breadcrumb node container -->
        <div class="relative flex items-center">
          <!-- Main node button -->
          <button
            @click="jumpTo(node)"
            :class="[
              'inline-flex items-center gap-1 px-2 py-1 rounded transition border text-left',
              activeNodeId === node.id
                ? 'bg-amber-100/80 border-amber-300 text-amber-900 font-medium'
                : 'bg-white border-stone-200 text-stone-700 hover:bg-stone-50 hover:border-stone-300'
            ]"
            :title="`Jump to ${node.label}`"
          >
            <span :class="['text-xs', KIND_COLOR[node.kind]]">{{ KIND_GLYPH[node.kind] }}</span>
            <span class="font-serif text-xs max-w-[120px] truncate">{{ node.label }}</span>
          </button>

          <!-- Sibling branch trigger dropdown button -->
          <button
            v-if="getSiblings(node, index).length > 1"
            @click="(e) => toggleDropdown(node.id, e)"
            class="breadcrumb-dropdown-trigger px-1 py-1 text-stone-400 hover:text-stone-700 hover:bg-stone-100 rounded transition border-t border-b border-r border-stone-200 bg-white -ml-px"
            title="Alternative sibling branches"
          >
            <span class="text-[10px]">▾</span>
          </button>

          <!-- Dropdown menu -->
          <transition name="fade">
            <div
              v-if="openDropdownId === node.id"
              class="breadcrumb-dropdown-menu absolute top-full left-0 mt-1 bg-white border border-stone-300 rounded shadow-lg py-1 z-30 min-w-[180px] max-h-60 overflow-y-auto"
            >
              <div class="px-2 py-1 text-[10px] text-stone-400 uppercase tracking-wider font-mono border-b border-stone-100 mb-1">
                Branch Options
              </div>
              <button
                v-for="sib in getSiblings(node, index)"
                :key="sib.id"
                @click="jumpTo(sib)"
                :class="[
                  'w-full text-left px-3 py-1.5 text-xs hover:bg-amber-50 flex items-center gap-1.5 transition font-serif',
                  sib.id === node.id ? 'bg-amber-50/50 text-amber-900 font-semibold' : 'text-stone-700'
                ]"
              >
                <span :class="KIND_COLOR[sib.kind]">{{ KIND_GLYPH[sib.kind] }}</span>
                <span class="truncate">{{ sib.label }}</span>
              </button>
            </div>
          </transition>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.fade-enter-active, .fade-leave-active {
  transition: opacity 0.15s ease, transform 0.15s ease;
}
.fade-enter-from, .fade-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}
</style>
