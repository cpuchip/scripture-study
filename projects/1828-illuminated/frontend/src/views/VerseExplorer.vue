<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { RouterLink } from 'vue-router'
import HighlightedText from '@/components/HighlightedText.vue'
import WordCard from '@/components/WordCard.vue'
import { selectWord, selectedWord, useWordData, tokenize } from '@/composables/useWordData'
import { useLLMRender } from '@/composables/useLLMRender'
import { sessionActive, refreshSession } from '@/composables/useLLMSession'
import { apiUrl } from '@/composables/useApiBase'
import demoData from '@/data/demo-verses.json'
import { CANON, buildChurchUrl, type CanonVolume, type CanonBook } from '@/data/canon-books'

const data = useWordData()
type DemoVerse = (typeof demoData.verses)[number]
const demoVerses: DemoVerse[] = demoData.verses

const mode = ref<'demo' | 'canon' | 'paste'>('demo')
const selectedVerseId = ref<string>(demoVerses[0]?.id ?? '')
const pasteText = ref<string>('')

// ─── Canon-browse mode state ───────────────────────────────────────────
// Volume → book → chapter. The selectors are populated from CANON
// (frontend-side; stable + audit-friendly) and the chapter render comes
// from /api/scripture/chapter/{abbr}/{chapter}.
const canonVolumeId = ref<CanonVolume['id']>('bofm')
const canonBookAbbr = ref<string>('1-ne')
const canonChapter = ref<number>(3)
const canonChapterText = ref<string>('')
const canonChapterLoading = ref(false)
const canonChapterError = ref<string>('')
const canonChapterRef = ref<string>('')      // human ref returned by backend
const canonChapterAbbrRef = ref<string>('')  // abbr ref for the URL

const canonVolume = computed<CanonVolume | undefined>(() =>
  CANON.find(v => v.id === canonVolumeId.value),
)
const canonBook = computed<CanonBook | undefined>(() =>
  canonVolume.value?.books.find(b => b.abbr === canonBookAbbr.value),
)
const canonChapterMax = computed(() => canonBook.value?.chapters ?? 1)
const canonChapterChoices = computed(() => {
  const max = canonChapterMax.value
  return Array.from({ length: max }, (_, i) => i + 1)
})
const canonChurchUrl = computed(() => {
  const vol = canonVolume.value
  const book = canonBook.value
  if (!vol || !book) return ''
  return buildChurchUrl(vol.urlVolume, book.urlPath, canonChapter.value)
})

async function fetchCanonChapter() {
  const vol = canonVolume.value
  const book = canonBook.value
  if (!vol || !book) return
  canonChapterLoading.value = true
  canonChapterError.value = ''
  canonChapterText.value = ''
  try {
    const url = apiUrl(`/scripture/chapter/${encodeURIComponent(book.abbr)}/${canonChapter.value}`)
    const resp = await fetch(url)
    if (!resp.ok) {
      canonChapterError.value = `HTTP ${resp.status} — chapter not found.`
      return
    }
    const json = await resp.json()
    canonChapterRef.value = json.ref ?? `${book.name} ${canonChapter.value}`
    canonChapterAbbrRef.value = json.abbr_ref ?? `${book.abbr}/${canonChapter.value}`
    // Join verses into one rendered block. Tokenize handles the inline
    // highlights for tier words — same code path as demo + paste.
    const verses: Array<{ verse: number; text: string }> = json.verses ?? []
    canonChapterText.value = verses.map(v => `${v.verse} ${v.text}`).join(' ')
  } catch (e: unknown) {
    canonChapterError.value = e instanceof Error ? e.message : String(e)
  } finally {
    canonChapterLoading.value = false
  }
}

// When the volume changes, default to its first book.
watch(canonVolumeId, () => {
  const firstBook = canonVolume.value?.books[0]
  if (firstBook) {
    canonBookAbbr.value = firstBook.abbr
    canonChapter.value = 1
  }
})
// When the book changes, clamp chapter to the book's range.
watch(canonBookAbbr, () => {
  if (canonChapter.value > canonChapterMax.value) {
    canonChapter.value = 1
  }
})

const selectedVerse = computed<DemoVerse | undefined>(() =>
  demoVerses.find(v => v.id === selectedVerseId.value)
)

const activeText = computed(() => {
  if (mode.value === 'demo') return selectedVerse.value?.text ?? ''
  if (mode.value === 'canon') return canonChapterText.value
  return pasteText.value
})

const wordsInText = computed(() => {
  const seen = new Set<string>()
  const out: { word: string; tier: string }[] = []
  for (const seg of tokenize(activeText.value)) {
    const w = seg.word
    const t = seg.tier
    if (w && t && !seen.has(w)) {
      seen.add(w)
      out.push({ word: w, tier: t })
    }
  }
  const order: Record<string, number> = { 'A++': 0, 'A+': 1, B: 2, C: 3, D: 4 }
  return out.sort((a, b) => (order[a.tier] ?? 99) - (order[b.tier] ?? 99) || a.word.localeCompare(b.word))
})

const samplePaste = `Yea, blessed are they whose feet stand upon the land of Zion, who have obeyed my gospel; for they shall receive for their reward the good things of the earth, and it shall bring forth in its strength.`

// LLM-render — stretch goal #1. Gated behind an active BYOK session.
const { state: renderState, render: renderModern, reset: resetRender } = useLLMRender()
const llmConfigured = computed(() => sessionActive.value)

// Confirm the local session_id mirror against the server when this view
// mounts — catches the case where the backend's janitor evicted the
// session while the tab was idle.
refreshSession().catch(() => { /* network errors leave the mirror as-is */ })

function onRenderClick() {
  if (activeText.value) {
    renderModern(activeText.value)
  }
}
function onVerseChange() {
  resetRender()
}

// Reset rendered output any time the source text changes (mode switch, new
// demo selection, paste edits) so we don't leave stale renders visible.
watch(activeText, () => {
  resetRender()
})
</script>

<template>
  <div class="max-w-6xl mx-auto px-6 py-10">
    <header class="mb-8">
      <h1 class="text-3xl font-serif mb-2">Verse explorer</h1>
      <p class="text-stone-600">Highlighted words are in our tier list. Click any to see the 1828 + modern definitions and how our studies have lensed it.</p>
    </header>

    <!-- Mode switcher -->
    <div class="flex gap-3 mb-6 text-sm flex-wrap">
      <button
        class="px-4 py-1.5 rounded-full border font-medium transition"
        :class="mode === 'demo' ? 'bg-amber-100 border-amber-400 text-amber-900' : 'bg-white border-stone-300 text-stone-600 hover:border-stone-400'"
        @click="mode = 'demo'"
      >Demo verse</button>
      <button
        class="px-4 py-1.5 rounded-full border font-medium transition"
        :class="mode === 'canon' ? 'bg-amber-100 border-amber-400 text-amber-900' : 'bg-white border-stone-300 text-stone-600 hover:border-stone-400'"
        @click="mode = 'canon'"
      >Browse canon</button>
      <button
        class="px-4 py-1.5 rounded-full border font-medium transition"
        :class="mode === 'paste' ? 'bg-amber-100 border-amber-400 text-amber-900' : 'bg-white border-stone-300 text-stone-600 hover:border-stone-400'"
        @click="mode = 'paste'"
      >Paste your own</button>
    </div>

    <div class="grid lg:grid-cols-[1fr_360px] gap-8">
      <!-- Left: the verse + highlight markup -->
      <div>
        <div v-if="mode === 'demo'">
          <select
            v-model="selectedVerseId"
            @change="onVerseChange"
            class="px-3 py-2 rounded-lg border border-stone-300 bg-white text-sm mb-4 w-full max-w-md"
          >
            <option v-for="v in demoVerses" :key="v.id" :value="v.id">{{ v.ref }}</option>
          </select>
          <div v-if="selectedVerse" class="def-card p-6 space-y-4">
            <header class="border-b border-stone-200 pb-3">
              <h2 class="font-serif text-xl">{{ selectedVerse.ref }}</h2>
              <p class="text-sm text-stone-600 mt-1">{{ selectedVerse.blurb }}</p>
            </header>
            <HighlightedText :text="selectedVerse.text" />
            <footer class="text-xs text-stone-500 border-t border-stone-200 pt-3 space-y-1.5">
              <div class="flex flex-wrap items-baseline gap-3">
                <RouterLink
                  :to="{ name: 'present', query: { v: selectedVerse.id } }"
                  class="px-2.5 py-1 rounded border border-stone-400 text-stone-700 hover:bg-amber-50 hover:border-amber-500 transition no-underline"
                >📖 Present this verse (fullscreen)</RouterLink>
                <a :href="selectedVerse.church_url" target="_blank" rel="noopener" class="underline text-amber-700">Full passage at churchofjesuschrist.org ↗</a>
              </div>
              <div v-if="selectedVerse.study_link">
                Substrate study that lensed this passage:
                <a
                  :href="`https://github.com/cpuchip/scripture-study/blob/main/${selectedVerse.study_link.replace(/^[./]+/, '')}`"
                  target="_blank"
                  rel="noopener"
                  class="text-amber-700 bg-amber-50 hover:bg-amber-100 px-1.5 py-0.5 rounded text-xs font-mono inline-flex items-baseline gap-1 transition"
                >{{ selectedVerse.study_link.replace(/^[./]+/, '') }} <span>↗</span></a>
              </div>
            </footer>
          </div>
        </div>

        <div v-else-if="mode === 'canon'">
          <!-- Volume / book / chapter selectors. The render comes from
               /api/scripture/chapter/:abbr/:chapter. -->
          <div class="grid sm:grid-cols-[1fr_1fr_120px_auto] gap-2 mb-4">
            <select
              v-model="canonVolumeId"
              class="px-3 py-2 rounded-lg border border-stone-300 bg-white text-sm"
              aria-label="Volume"
            >
              <option v-for="v in CANON" :key="v.id" :value="v.id">{{ v.label }}</option>
            </select>
            <select
              v-model="canonBookAbbr"
              class="px-3 py-2 rounded-lg border border-stone-300 bg-white text-sm"
              aria-label="Book"
            >
              <option v-for="b in canonVolume?.books ?? []" :key="b.abbr" :value="b.abbr">{{ b.name }}</option>
            </select>
            <select
              v-model.number="canonChapter"
              class="px-3 py-2 rounded-lg border border-stone-300 bg-white text-sm"
              aria-label="Chapter"
            >
              <option v-for="c in canonChapterChoices" :key="c" :value="c">{{ c }}</option>
            </select>
            <button
              @click="fetchCanonChapter"
              :disabled="canonChapterLoading"
              class="px-4 py-2 rounded-lg text-sm font-medium bg-amber-600 text-white hover:bg-amber-700 transition disabled:bg-stone-300 disabled:cursor-not-allowed"
            >
              <span v-if="canonChapterLoading">Loading…</span>
              <span v-else>Open chapter</span>
            </button>
          </div>

          <div v-if="canonChapterError" class="def-card p-4 mb-4 text-sm text-red-700 bg-red-50 border-red-200">
            {{ canonChapterError }}
          </div>

          <div v-if="canonChapterText" class="def-card p-6 space-y-4">
            <header class="border-b border-stone-200 pb-3 flex items-baseline justify-between gap-3">
              <h2 class="font-serif text-xl">{{ canonChapterRef }}</h2>
              <a
                v-if="canonChurchUrl"
                :href="canonChurchUrl"
                target="_blank"
                rel="noopener"
                class="text-xs text-amber-700 hover:underline"
                :title="`Open ${canonChapterRef} at churchofjesuschrist.org for footnotes + study apparatus`"
              >Full chapter at churchofjesuschrist.org ↗</a>
            </header>
            <HighlightedText :text="canonChapterText" />
            <p class="text-xs text-stone-500 italic">
              Verse text from the bcbooks 2013 corpus; footnotes, headings, and study apparatus stripped.
              The "↗" above opens the canonical apparatus on churchofjesuschrist.org.
            </p>
          </div>
          <div v-else-if="!canonChapterLoading && !canonChapterError" class="def-card p-6 text-sm text-stone-500 italic">
            Pick a volume, book, and chapter, then click <strong>Open chapter</strong>.
          </div>
        </div>

        <div v-else>
          <textarea
            v-model="pasteText"
            class="w-full px-4 py-3 rounded-lg border border-stone-300 focus:border-amber-500 focus:outline-none bg-white font-serif text-lg leading-relaxed"
            rows="8"
            :placeholder="`Paste a verse here — for example:\n\n${samplePaste}`"
          />
          <button
            v-if="!pasteText"
            class="mt-3 text-sm text-amber-700 hover:underline"
            @click="pasteText = samplePaste"
          >Use the sample text</button>
          <div v-if="pasteText" class="def-card p-6 mt-4">
            <HighlightedText :text="pasteText" />
          </div>
        </div>

        <!-- Words in this passage -->
        <section v-if="wordsInText.length" class="mt-6">
          <h3 class="text-sm uppercase tracking-wider text-stone-500 mb-3 font-sans">Tier words in this passage ({{ wordsInText.length }})</h3>
          <div class="flex flex-wrap gap-2">
            <button
              v-for="w in wordsInText"
              :key="w.word"
              @click="selectWord(w.word)"
              class="px-3 py-1.5 bg-white border border-stone-300 rounded-full text-sm hover:border-amber-500 hover:bg-amber-50 transition"
              :class="selectedWord === w.word ? 'border-amber-500 bg-amber-50' : ''"
            >
              <span class="font-serif">{{ w.word }}</span>
              <span class="text-xs text-stone-500 ml-1.5">{{ w.tier }}</span>
            </button>
          </div>
        </section>

        <!-- LLM-render: stretch goal #1 -->
        <section v-if="activeText" class="mt-6 border-t border-stone-300 pt-6">
          <div class="flex items-center justify-between gap-3 mb-3">
            <h3 class="text-sm uppercase tracking-wider text-stone-500 font-sans">Render in modern English</h3>
            <RouterLink v-if="!llmConfigured" to="/settings" class="text-xs text-amber-700 hover:underline">
              Start a BYOK session ⚙ ↗
            </RouterLink>
          </div>
          <p class="text-xs text-stone-500 mb-3">
            Asks your configured LLM to render the passage in modern English while preserving each tier word's 1828 sense. Original words marked in <code class="bg-stone-100 px-1 rounded">[brackets]</code> for transparency. Token costs land on your account.
          </p>
          <button
            class="px-4 py-2 rounded-lg text-sm font-medium transition"
            :class="llmConfigured && !renderState.loading
              ? 'bg-amber-600 text-white hover:bg-amber-700'
              : 'bg-stone-200 text-stone-500 cursor-not-allowed'"
            :disabled="!llmConfigured || renderState.loading"
            @click="onRenderClick"
          >
            <span v-if="renderState.loading">Rendering…</span>
            <span v-else-if="renderState.result">Render again</span>
            <span v-else>Render in modern English</span>
          </button>

          <div v-if="renderState.error" class="mt-3 def-card p-4 text-sm text-red-700 bg-red-50 border-red-200">
            <div class="flex items-baseline gap-2">
              <strong>
                <template v-if="renderState.error.kind === 'reauth'">Session expired</template>
                <template v-else-if="renderState.error.kind === 'rate_limited_by_1828'">Throttled by 1828.ibeco.me</template>
                <template v-else-if="renderState.error.kind === 'upstream_provider_error'">Provider error</template>
                <template v-else-if="renderState.error.kind === 'feature_disabled'">Render disabled</template>
                <template v-else>Error</template>
              </strong>
              <span v-if="renderState.error.kind === 'rate_limited_by_1828'" class="text-xs text-red-500 italic">
                (this is our cap, not your provider's)
              </span>
            </div>
            <div class="mt-1">{{ renderState.error.message }}</div>
            <div v-if="renderState.error.retryAfterSeconds" class="text-xs mt-1 text-red-600">
              Retry after ~{{ renderState.error.retryAfterSeconds }}s.
            </div>
            <RouterLink
              v-if="renderState.error.kind === 'reauth'"
              to="/settings"
              class="inline-block mt-2 text-xs text-amber-700 hover:underline"
            >Re-authenticate in Settings ↗</RouterLink>
          </div>

          <div v-if="renderState.result" class="mt-4 def-card p-5 bg-amber-50/30 border-amber-300">
            <div class="text-xs text-stone-500 mb-2 flex justify-between">
              <span>Rendered ({{ Math.round(renderState.result.durationMs) }}ms)</span>
              <button @click="resetRender" class="text-stone-500 hover:text-stone-900">Clear ✕</button>
            </div>
            <div class="font-serif text-lg leading-relaxed text-stone-900 whitespace-pre-wrap">{{ renderState.result.modernized }}</div>
          </div>
        </section>
      </div>

      <!-- Right: selected word card -->
      <aside id="selected-word-card" class="lg:sticky lg:top-6 self-start">
        <div v-if="selectedWord">
          <WordCard :word="selectedWord" compact />
          <button
            class="mt-3 text-sm text-stone-500 hover:text-stone-900"
            @click="selectWord(null)"
          >Close ↓</button>
        </div>
        <div v-else class="def-card p-6 text-sm text-stone-500 italic">
          Click a highlighted word in the verse to see its 1828 and modern definitions here.
        </div>
      </aside>
    </div>
  </div>
</template>
