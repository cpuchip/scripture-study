<script setup lang="ts">
// UX 1-2 punch (2026-05-29): the always-visible unified search. One header
// input that detects intent and routes:
//   - a scripture reference ("1 Ne 3:7", "John 3:16", "Alma 32") → Verse Explorer
//   - a 1828 word ("charity") → word detail (full class-E lookup happens there)
// Free-text full-text scripture search is specced but deferred — there is no
// results surface for it yet (see .spec/proposals/ux-1-2-punch.md).
//
// Keyboard: "/" or Cmd/Ctrl-K focuses from anywhere; ArrowUp/Down move the
// highlight; Enter activates; Esc closes/blurs.
import { ref, computed, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { CANON } from '@/data/canon-books'
import { useWordData } from '@/composables/useWordData'

const router = useRouter()
const { searchPrefix } = useWordData()

const q = ref('')
const open = ref(false)
const highlight = ref(0)
const inputEl = ref<HTMLInputElement | null>(null)

// ── reference parsing ──────────────────────────────────────────────────
const normalize = (s: string) => s.toLowerCase().replace(/[\s.\-]/g, '')

interface RefMatch {
  vol: string
  book: string
  name: string
  chapter: number
  range: string
  label: string
}

function parseReference(input: string): RefMatch | null {
  // "<book> <chapter>[:<verse>[-<verse>]]" — book may contain spaces/numbers.
  const m = input.trim().match(/^(.+?)\s+(\d+)(?::(\d+)(?:\s*-\s*(\d+))?)?$/)
  if (!m) return null
  const bookKey = normalize(m[1]!)
  const chapter = parseInt(m[2]!, 10)
  const vStart = m[3] ? parseInt(m[3], 10) : null
  const vEnd = m[4] ? parseInt(m[4], 10) : null
  for (const vol of CANON) {
    for (const b of vol.books) {
      if (normalize(b.name) === bookKey || normalize(b.abbr) === bookKey) {
        if (chapter >= 1 && chapter <= b.chapters) {
          const range = vStart ? (vEnd ? `${vStart}-${vEnd}` : `${vStart}`) : ''
          const vLabel = vStart ? `:${vEnd ? `${vStart}-${vEnd}` : vStart}` : ''
          return { vol: vol.id, book: b.abbr, name: b.name, chapter, range, label: `${b.name} ${chapter}${vLabel}` }
        }
      }
    }
  }
  return null
}

// ── suggestions ────────────────────────────────────────────────────────
type Suggestion =
  | { kind: 'ref'; label: string; ref: RefMatch }
  | { kind: 'word'; label: string; word: string }

const suggestions = computed<Suggestion[]>(() => {
  const text = q.value.trim()
  if (!text) return []
  const out: Suggestion[] = []

  const refMatch = parseReference(text)
  if (refMatch) out.push({ kind: 'ref', label: refMatch.label, ref: refMatch })

  // Word path. A single token is the common dictionary case; always offer the
  // direct lookup (word-detail does the full class-E + stem fallback), then a
  // few instant prefix matches from the curated tiers.
  const firstToken = text.split(/\s+/)[0]!
  if (/^[a-zA-Z'-]+$/.test(firstToken)) {
    out.push({ kind: 'word', label: `Look up “${firstToken}” in 1828`, word: firstToken })
    for (const tw of searchPrefix(firstToken, 5)) {
      if (tw.word.toLowerCase() !== firstToken.toLowerCase()) {
        out.push({ kind: 'word', label: tw.word, word: tw.word })
      }
    }
  }
  return out.slice(0, 8)
})

function activate(s: Suggestion) {
  if (s.kind === 'ref') {
    const query: Record<string, string> = {
      mode: 'canon', v: s.ref.vol, b: s.ref.book, c: String(s.ref.chapter),
    }
    if (s.ref.range) query.r = s.ref.range
    router.push({ name: 'verse-explorer', query }).catch(() => {})
  } else {
    router.push({ name: 'word-detail', params: { word: s.word } }).catch(() => {})
  }
  close()
}

function onEnter() {
  const list = suggestions.value
  if (list.length) {
    activate(list[Math.min(highlight.value, list.length - 1)]!)
  }
}

function close() {
  open.value = false
  q.value = ''
  highlight.value = 0
  inputEl.value?.blur()
}

function move(delta: number) {
  const n = suggestions.value.length
  if (!n) return
  highlight.value = (highlight.value + delta + n) % n
}

// ── global focus shortcut (/ and Cmd/Ctrl-K) ───────────────────────────
function onGlobalKey(e: KeyboardEvent) {
  const tag = (e.target as HTMLElement | null)?.tagName
  const typing = tag === 'INPUT' || tag === 'TEXTAREA' || (e.target as HTMLElement | null)?.isContentEditable
  if ((e.key === 'k' && (e.metaKey || e.ctrlKey)) || (e.key === '/' && !typing)) {
    e.preventDefault()
    inputEl.value?.focus()
    open.value = true
  }
}
onMounted(() => window.addEventListener('keydown', onGlobalKey))
onBeforeUnmount(() => window.removeEventListener('keydown', onGlobalKey))

function onFocus() { open.value = true }
function onBlur() { setTimeout(() => { open.value = false }, 120) } // allow click
</script>

<template>
  <div class="relative w-full max-w-xs">
    <input
      ref="inputEl"
      v-model="q"
      type="search"
      placeholder="Search a word or reference…"
      class="w-full rounded-md border border-stone-300 bg-[var(--paper)] px-3 py-1.5 text-sm
             placeholder:text-stone-400 focus:border-amber-500 focus:outline-none focus:ring-1 focus:ring-amber-400"
      @focus="onFocus"
      @blur="onBlur"
      @keydown.down.prevent="move(1)"
      @keydown.up.prevent="move(-1)"
      @keydown.enter.prevent="onEnter"
      @keydown.esc.prevent="close"
      aria-label="Search words or scripture references"
    />
    <kbd
      v-if="!open && !q"
      class="pointer-events-none absolute right-2 top-1/2 -translate-y-1/2 rounded border border-stone-300 px-1 text-[10px] text-stone-400"
    >/</kbd>

    <ul
      v-if="open && suggestions.length"
      class="absolute z-30 mt-1 w-full overflow-hidden rounded-md border border-stone-300 bg-[var(--paper-2)] shadow-lg"
    >
      <li
        v-for="(s, i) in suggestions"
        :key="s.kind + ':' + s.label"
        :class="[
          'flex items-baseline gap-2 px-3 py-1.5 text-sm cursor-pointer',
          i === highlight ? 'bg-amber-100 text-stone-900' : 'text-stone-700 hover:bg-stone-100',
        ]"
        @mousedown.prevent="activate(s)"
        @mouseenter="highlight = i"
      >
        <span class="text-xs">{{ s.kind === 'ref' ? '📖' : '📚' }}</span>
        <span>{{ s.label }}</span>
        <span v-if="s.kind === 'ref'" class="ml-auto text-[10px] uppercase tracking-wide text-stone-400">verse</span>
      </li>
    </ul>
  </div>
</template>
