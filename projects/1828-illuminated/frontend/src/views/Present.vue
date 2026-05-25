<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, watch } from 'vue'
import { useRoute, useRouter, RouterLink } from 'vue-router'
import {
  tokenize,
  useWordData,
  type Def1828Entry,
  type ModernEntry,
} from '@/composables/useWordData'
import { useStudyTree } from '@/composables/useStudyTree'
import { ancestorsOf } from '@/composables/useStudyTree'
import { CANON, buildChurchUrl } from '@/data/canon-books'
import { apiUrl } from '@/composables/useApiBase'
import LinkedDefinition from '@/components/LinkedDefinition.vue'
import demoData from '@/data/demo-verses.json'

interface Slide {
  id: string
  ref: string
  blurb?: string
  text?: string
  church_url?: string
  kind?: 'word' | 'verse' | 'chapter' | 'render' | 'note'
  payload?: any
}

type DemoVerse = (typeof demoData.verses)[number]
const demoVerses: DemoVerse[] = demoData.verses

const route = useRoute()
const router = useRouter()
const data = useWordData()

const slides = ref<Slide[]>([])
const loading = ref(false)
const error = ref<string>('')
const verseIdx = ref<number>(0)

// Load slides based on query parameters
async function loadSlides() {
  const mode = route.query.mode as string | undefined
  slides.value = []
  error.value = ''

  if (!mode || mode === 'demo') {
    // Demo Mode
    slides.value = demoVerses.map(d => ({
      id: d.id,
      ref: d.ref,
      blurb: d.blurb,
      text: d.text,
      church_url: d.church_url
    }))
  } else if (mode === 'canon') {
    // Canon Mode
    const b = route.query.b as string | undefined
    const c = route.query.c as string | undefined
    const r = route.query.r as string | undefined
    if (!b || !c) {
      error.value = 'Invalid book or chapter parameter for canon presentation'
      return
    }

    loading.value = true
    try {
      const bookAbbr = decodeURIComponent(b)
      const chapterNum = parseInt(c, 10)
      
      let url: string
      if (r && r.trim()) {
        url = apiUrl(`/scripture/${encodeURIComponent(bookAbbr)}/${chapterNum}:${encodeURIComponent(r.trim())}?highlight=1`)
      } else {
        url = apiUrl(`/scripture/chapter/${encodeURIComponent(bookAbbr)}/${chapterNum}?highlight=1`)
      }

      const resp = await fetch(url)
      if (!resp.ok) {
        error.value = `HTTP ${resp.status} — scripture not found.`
        return
      }
      const json = await resp.json()
      
      const foundVol = CANON.find(v => v.books.some(bk => bk.abbr === bookAbbr))
      const foundBook = foundVol?.books.find(bk => bk.abbr === bookAbbr)
      
      const verses = json.verses ?? []
      slides.value = verses.map((v: any) => {
        const verseNum = v.verse
        const churchUrl = (foundVol && foundBook) 
          ? buildChurchUrl(foundVol.urlVolume, foundBook.urlPath, chapterNum, verseNum, verseNum) 
          : ''
        return {
          id: `${bookAbbr}-${chapterNum}-${verseNum}`,
          ref: `${json.ref.replace(/:.*/, '')}:${verseNum}`,
          blurb: `Verse ${verseNum}`,
          text: v.text,
          church_url: churchUrl
        }
      })
    } catch (e: any) {
      error.value = e.message || String(e)
    } finally {
      loading.value = false
    }
  } else if (mode === 'paste') {
    // Paste Mode
    const text = route.query.text as string | undefined
    if (text) {
      slides.value = [{
        id: 'paste',
        ref: 'Pasted Text',
        text: text
      }]
    } else {
      slides.value = [{
        id: 'paste',
        ref: 'Pasted Text',
        text: 'No text pasted. Go to Verse Explorer to paste text.'
      }]
    }
  } else if (mode === 'tree') {
    // Tree Mode
    const root = route.query.root as string | undefined
    const { activeNodeId } = useStudyTree()
    
    const targetId = root || activeNodeId.value
    if (!targetId) {
      error.value = 'No study tree path active. Start studying to present a path!'
      return
    }

    const pathNodes = ancestorsOf(targetId)
    if (!pathNodes.length) {
      error.value = 'No path found for the selected node.'
      return
    }

    slides.value = pathNodes.map(node => {
      let ref = node.label
      let blurb = `[${node.kind}] Study Branch`
      let text = ''
      let church_url = ''

      if (node.kind === 'word' && node.payload.kind === 'word') {
        ref = node.payload.word
        blurb = '1828 Webster Word Definition'
      } else if (node.kind === 'verse' && node.payload.kind === 'verse') {
        ref = node.payload.humanRef
        text = node.payload.text || ''
        const [book] = node.payload.abbrRef.split('/')
        const foundVol = CANON.find(v => v.books.some(bk => bk.abbr === book))
        const foundBook = foundVol?.books.find(bk => bk.abbr === book)
        if (foundVol && foundBook) {
          church_url = buildChurchUrl(
            foundVol.urlVolume, 
            foundBook.urlPath, 
            parseInt(node.payload.abbrRef.split('/')[1] || '1', 10), 
            node.payload.verse, 
            node.payload.verse
          )
        }
      } else if (node.kind === 'chapter' && node.payload.kind === 'chapter') {
        ref = node.payload.humanRef
        text = `Chapter containing ${node.payload.verseCount} verses`
      } else if (node.kind === 'render' && node.payload.kind === 'render') {
        ref = `Modern Translation (${node.payload.model})`
        text = node.payload.modernized
      } else if (node.kind === 'note' && node.payload.kind === 'note') {
        ref = 'Study Note'
        text = node.payload.body
      }

      return {
        id: node.id,
        ref,
        blurb,
        text,
        church_url,
        kind: node.kind,
        payload: node.payload
      }
    })
  }
}

function syncIdxFromRoute() {
  const mode = route.query.mode as string | undefined
  if (!mode || mode === 'demo') {
    const v = route.query.v as string | undefined
    if (v && slides.value.length > 0) {
      const found = slides.value.findIndex(d => d.id === v)
      if (found >= 0) verseIdx.value = found
    }
  } else {
    const slide = route.query.slide as string | undefined
    if (slide) {
      const idx = parseInt(slide, 10)
      if (!isNaN(idx) && idx >= 0 && idx < slides.value.length) {
        verseIdx.value = idx
      }
    }
  }
}

onMounted(async () => {
  await loadSlides()
  syncIdxFromRoute()
})

watch(() => [route.query.mode, route.query.b, route.query.c, route.query.r, route.query.text, route.query.root, route.query.v], async () => {
  await loadSlides()
  syncIdxFromRoute()
})

watch(() => route.query.slide, () => {
  syncIdxFromRoute()
})

const verse = computed<Slide>(() => (slides.value[verseIdx.value] ?? { id: '', ref: '', text: '' }) as Slide)
const segments = computed(() => tokenize(verse.value.text ?? ''))

// Selected word (overlay details inside presentation mode)
const selected = ref<string | null>(null)
const selected1828 = ref<Def1828Entry[]>([])
const selectedStemMatched = ref<string | null>(null)
const selectedModern = ref<ModernEntry[]>([])
const selectedModernNoEntry = ref(false)
const selectedTier = computed(() => selected.value ? data.findWord(selected.value) : undefined)

watch(selected, async (w) => {
  if (!w) {
    selected1828.value = []
    selectedStemMatched.value = null
    selectedModern.value = []
    selectedModernNoEntry.value = false
    return
  }
  const [r1828, rModern] = await Promise.all([data.get1828(w), data.getModern(w)])
  selected1828.value = r1828.entries
  selectedStemMatched.value = r1828.stem_matched
  selectedModern.value = rModern.entries ?? []
  selectedModernNoEntry.value = !rModern.found && rModern.source === 'none'
})

// Reactive data for word node slides
const wordData1828 = ref<Def1828Entry[]>([])
const wordModern = ref<ModernEntry[]>([])
const wordStemMatched = ref<string | null>(null)

watch(() => verse.value, async (v) => {
  if (v.kind === 'word' && v.payload?.kind === 'word') {
    const w = v.payload.word
    const [r1828, rModern] = await Promise.all([data.get1828(w), data.getModern(w)])
    wordData1828.value = r1828.entries
    wordStemMatched.value = r1828.stem_matched
    wordModern.value = rModern.entries ?? []
  }
})

function openWord(word: string) { selected.value = word }
// Also make openWord support clicking inside custom definition overlays
function handleDefinitionWordClick(e: MouseEvent) {
  const target = e.target as HTMLElement
  if (target.classList.contains('word-link')) {
    const word = target.dataset.word
    if (word) {
      openWord(word)
    }
  }
}
function closeWord() { selected.value = null }

// Navigation
function next() {
  if (slides.value.length === 0) return
  closeWord()
  verseIdx.value = (verseIdx.value + 1) % slides.value.length
  updateQuery()
}
function prev() {
  if (slides.value.length === 0) return
  closeWord()
  verseIdx.value = (verseIdx.value - 1 + slides.value.length) % slides.value.length
  updateQuery()
}
function updateQuery() {
  const currentQuery = { ...route.query }
  const currentSlide = slides.value[verseIdx.value]
  if (currentSlide) {
    if (!route.query.mode || route.query.mode === 'demo') {
      currentQuery.v = currentSlide.id
    } else {
      currentQuery.slide = String(verseIdx.value)
    }
  }
  router.replace({ name: 'present', query: currentQuery })
}

// Keyboard nav
function onKey(e: KeyboardEvent) {
  if (e.key === 'ArrowRight' || e.key === ' ') { next(); e.preventDefault() }
  else if (e.key === 'ArrowLeft') { prev(); e.preventDefault() }
  else if (e.key === 'Escape') { closeWord() }
}
onMounted(() => window.addEventListener('keydown', onKey))
onBeforeUnmount(() => window.removeEventListener('keydown', onKey))
</script>

<template>
  <div class="fixed inset-0 bg-[var(--paper)] z-40 overflow-auto" style="font-feature-settings: 'liga' 1, 'kern' 1;">
    <!-- Top minimal bar -->
    <div class="absolute top-0 right-0 left-0 px-6 py-3 flex items-center justify-between bg-[var(--paper-2)]/80 backdrop-blur-sm border-b border-stone-300 z-10">
      <RouterLink to="/verse" class="text-sm text-stone-500 hover:text-stone-900">← Exit presentation</RouterLink>
      <div v-if="slides.length" class="text-xs text-stone-500">
        {{ verseIdx + 1 }} / {{ slides.length }} · ← → to navigate · Esc to close card
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="min-h-screen flex items-center justify-center text-stone-500 italic">
      Loading presentation slides…
    </div>

    <!-- Error State -->
    <div v-else-if="error" class="min-h-screen flex items-center justify-center p-6">
      <div class="def-card p-6 max-w-md w-full text-center border-red-200 bg-red-50 text-red-700">
        <h2 class="font-serif text-xl font-semibold mb-2">Error</h2>
        <p class="text-sm mb-4">{{ error }}</p>
        <RouterLink to="/verse" class="px-4 py-2 bg-stone-800 text-white rounded text-sm hover:bg-stone-700 transition">Back to Verse Explorer</RouterLink>
      </div>
    </div>

    <!-- Slide Viewport -->
    <div v-else class="min-h-screen flex flex-col justify-center pt-20 pb-12 px-6">
      <div class="max-w-4xl mx-auto w-full">
        <!-- Text style slides (verses/chapters/renders) -->
        <div v-if="!verse.kind || verse.kind === 'verse' || verse.kind === 'chapter' || verse.kind === 'render'">
          <header class="text-center mb-10">
            <h1 class="text-3xl md:text-5xl font-serif tracking-tight">{{ verse.ref }}</h1>
            <p class="text-stone-600 mt-3 text-base md:text-lg">{{ verse.blurb }}</p>
          </header>

          <article class="font-serif text-2xl md:text-4xl leading-relaxed text-stone-900 text-center md:text-left">
            <template v-for="(seg, i) in segments" :key="i">
              <span
                v-if="seg.word"
                class="highlight cursor-pointer"
                :class="seg.tier === 'A++' || seg.tier === 'A+' ? 'highlight-tier-A' : ''"
                role="button"
                tabindex="0"
                @click="openWord(seg.word)"
                @keydown.enter="openWord(seg.word)"
              >{{ seg.text }}</span>
              <template v-else>{{ seg.text }}</template>
            </template>
          </article>
        </div>

        <!-- Word style slide -->
        <div v-else-if="verse.kind === 'word'" class="space-y-6">
          <header class="text-center mb-6">
            <h1 class="text-4xl md:text-6xl font-serif tracking-tight text-amber-900">{{ verse.ref }}</h1>
            <p class="text-stone-600 mt-3 text-base md:text-lg">{{ verse.blurb }}</p>
          </header>
          
          <div class="grid md:grid-cols-2 gap-8 mt-10" @click="handleDefinitionWordClick">
            <div class="bg-white border border-stone-300 rounded-lg p-6 md:p-8 shadow-sm max-h-[50vh] overflow-y-auto">
              <h3 class="text-sm uppercase tracking-wider text-stone-500 mb-4 font-sans font-semibold border-b pb-2">Webster 1828</h3>
              <div v-if="wordData1828.length">
                <div v-for="(entry, idx) in wordData1828" :key="idx" class="mb-6 last:mb-0">
                  <div class="text-xs italic text-stone-400 mb-1">{{ entry.pos }}</div>
                  <ol class="list-decimal list-outside ml-6 text-lg leading-relaxed text-stone-800 space-y-3">
                    <li v-for="(def, di) in entry.definitions.slice(0, 6)" :key="di">
                      <LinkedDefinition :text="def" />
                    </li>
                  </ol>
                </div>
              </div>
              <div v-else class="text-stone-400 italic">No 1828 entry found.</div>
            </div>

            <div class="bg-white border border-stone-300 rounded-lg p-6 md:p-8 shadow-sm max-h-[50vh] overflow-y-auto">
              <h3 class="text-sm uppercase tracking-wider text-stone-500 mb-4 font-sans font-semibold border-b pb-2">Modern Sense</h3>
              <div v-if="wordModern.length">
                <div v-for="(entry, idx) in wordModern" :key="idx" class="mb-6 last:mb-0">
                  <div class="text-xs italic text-stone-400 mb-1">{{ entry.pos }}</div>
                  <ol class="list-decimal list-outside ml-6 text-lg leading-relaxed text-stone-800 space-y-3">
                    <li v-for="(def, di) in entry.definitions.slice(0, 4)" :key="di">
                      <LinkedDefinition :text="def" />
                    </li>
                  </ol>
                </div>
              </div>
              <div v-else class="text-stone-400 italic">No modern definition found.</div>
            </div>
          </div>
        </div>

        <!-- Note style slide -->
        <div v-else-if="verse.kind === 'note'" class="space-y-6 max-w-2xl mx-auto">
          <header class="text-center mb-6">
            <h1 class="text-3xl font-serif tracking-tight text-stone-700">Study Note</h1>
            <p class="text-stone-500 mt-2 text-sm">{{ verse.blurb }}</p>
          </header>
          
          <div class="bg-amber-50/40 border border-amber-200 rounded-xl p-8 md:p-12 shadow-sm font-serif text-2xl md:text-3xl text-stone-800 leading-relaxed text-center italic">
            "{{ verse.text }}"
          </div>
        </div>

        <footer v-if="verse.church_url" class="mt-12 text-center text-sm text-stone-500">
          <a :href="verse.church_url" target="_blank" rel="noopener" class="underline text-amber-700">Full passage at churchofjesuschrist.org ↗</a>
        </footer>
      </div>
    </div>

    <!-- Big navigation chevrons — easy to thumb on a tablet -->
    <button
      v-if="slides.length > 1"
      @click="prev"
      class="fixed left-2 md:left-6 top-1/2 -translate-y-1/2 w-12 h-12 md:w-16 md:h-16 rounded-full bg-white border border-stone-300 text-stone-600 hover:bg-amber-50 hover:border-amber-400 text-2xl flex items-center justify-center shadow-md transition z-20"
      aria-label="Previous verse"
    >‹</button>
    <button
      v-if="slides.length > 1"
      @click="next"
      class="fixed right-2 md:right-6 top-1/2 -translate-y-1/2 w-12 h-12 md:w-16 md:h-16 rounded-full bg-white border border-stone-300 text-stone-600 hover:bg-amber-50 hover:border-amber-400 text-2xl flex items-center justify-center shadow-md transition z-20"
      aria-label="Next verse"
    >›</button>

    <!-- Fullscreen word card overlay -->
    <div
      v-if="selected"
      class="fixed inset-0 bg-black/60 z-50 flex items-center justify-center p-4 md:p-12"
      @click.self="closeWord"
    >
      <article class="def-card bg-white max-w-3xl w-full max-h-full overflow-auto p-6 md:p-10 space-y-6">
        <header class="flex items-baseline justify-between border-b border-stone-200 pb-4">
          <h2 class="text-3xl md:text-5xl font-serif">{{ selected }}</h2>
          <button @click="closeWord" class="text-stone-400 hover:text-stone-900 text-3xl leading-none" aria-label="Close">×</button>
        </header>

        <section v-if="selected1828.length">
          <h3 class="text-sm uppercase tracking-wider text-stone-500 mb-3 font-sans">
            Webster 1828
            <span v-if="selectedStemMatched" class="ml-2 text-xs normal-case text-stone-500 italic">
              (showing entry for <code class="bg-stone-100 px-1 rounded">{{ selectedStemMatched }}</code>)
            </span>
          </h3>
          <div v-for="(entry, idx) in selected1828" :key="idx" class="mb-4 last:mb-0">
            <div class="text-sm italic text-stone-500 mb-1.5">{{ entry.pos }}</div>
            <ol class="list-decimal list-outside ml-6 text-base md:text-lg leading-relaxed text-stone-800 space-y-2">
              <li v-for="(def, di) in entry.definitions.slice(0, 8)" :key="di">
                <LinkedDefinition :text="def" />
              </li>
            </ol>
          </div>
        </section>

        <section v-if="selectedModern.length">
          <h3 class="text-sm uppercase tracking-wider text-stone-500 mb-3 font-sans">Modern</h3>
          <div v-for="(entry, idx) in selectedModern" :key="idx" class="mb-4 last:mb-0">
            <div class="text-sm italic text-stone-500 mb-1.5">{{ entry.pos }}</div>
            <ol class="list-decimal list-outside ml-6 text-base md:text-lg leading-relaxed text-stone-800 space-y-2">
              <li v-for="(def, di) in entry.definitions.slice(0, 4)" :key="di">
                <LinkedDefinition :text="def" />
              </li>
            </ol>
          </div>
        </section>
        <section v-else-if="selectedModernNoEntry" class="text-sm text-stone-500 italic">
          No modern dictionary entry — the word is sufficiently archaic that mainstream modern dictionaries don't cover it.
        </section>

        <section v-if="selectedTier && selectedTier.studies.length">
          <h3 class="text-sm uppercase tracking-wider text-stone-500 mb-3 font-sans">Lensed in our studies</h3>
          <ul class="text-sm space-y-1.5 text-stone-700 font-mono">
            <li v-for="s in selectedTier.studies" :key="s">{{ s }}</li>
          </ul>
        </section>
      </article>
    </div>
  </div>
</template>
