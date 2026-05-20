<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRoute, useRouter, RouterLink } from 'vue-router'
import { tokenize, useWordData } from '@/composables/useWordData'
import LinkedDefinition from '@/components/LinkedDefinition.vue'
import demoData from '@/data/demo-verses.json'

type DemoVerse = (typeof demoData.verses)[number]
const demoVerses: DemoVerse[] = demoData.verses

const route = useRoute()
const router = useRouter()
const data = useWordData()

// Verse selection via query string ?v=<id>; default to first demo.
const verseIdx = ref<number>(0)
function syncIdxFromRoute() {
  const v = route.query.v as string | undefined
  if (v) {
    const found = demoVerses.findIndex(d => d.id === v)
    if (found >= 0) verseIdx.value = found
  }
}
onMounted(syncIdxFromRoute)

const verse = computed<DemoVerse>(() => (demoVerses[verseIdx.value] ?? demoVerses[0]) as DemoVerse)
const segments = computed(() => tokenize(verse.value.text))

// Selected word — opens a fullscreen overlay card
const selected = ref<string | null>(null)
const selected1828 = computed(() => selected.value ? data.get1828(selected.value) : [])
const selectedModern = computed(() => selected.value ? data.getModern(selected.value) : null)
const selectedTier = computed(() => selected.value ? data.findWord(selected.value) : undefined)

function openWord(word: string) { selected.value = word }
function closeWord() { selected.value = null }

// Navigation
function next() {
  closeWord()
  verseIdx.value = (verseIdx.value + 1) % demoVerses.length
  updateQuery()
}
function prev() {
  closeWord()
  verseIdx.value = (verseIdx.value - 1 + demoVerses.length) % demoVerses.length
  updateQuery()
}
function updateQuery() {
  router.replace({ name: 'present', query: { v: verse.value.id } })
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
      <div class="text-xs text-stone-500">
        {{ verseIdx + 1 }} / {{ demoVerses.length }} · ← → to navigate · Esc to close card
      </div>
    </div>

    <!-- Verse body — large + tablet-friendly -->
    <div class="min-h-screen flex flex-col justify-center pt-20 pb-12 px-6">
      <div class="max-w-4xl mx-auto w-full">
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

        <footer class="mt-12 text-center text-sm text-stone-500">
          <a :href="verse.church_url" target="_blank" rel="noopener" class="underline text-amber-700">Full passage at churchofjesuschrist.org ↗</a>
        </footer>
      </div>
    </div>

    <!-- Big navigation chevrons — easy to thumb on a tablet -->
    <button
      @click="prev"
      class="fixed left-2 md:left-6 top-1/2 -translate-y-1/2 w-12 h-12 md:w-16 md:h-16 rounded-full bg-white border border-stone-300 text-stone-600 hover:bg-amber-50 hover:border-amber-400 text-2xl flex items-center justify-center shadow-md transition"
      aria-label="Previous verse"
    >‹</button>
    <button
      @click="next"
      class="fixed right-2 md:right-6 top-1/2 -translate-y-1/2 w-12 h-12 md:w-16 md:h-16 rounded-full bg-white border border-stone-300 text-stone-600 hover:bg-amber-50 hover:border-amber-400 text-2xl flex items-center justify-center shadow-md transition"
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
          <h3 class="text-sm uppercase tracking-wider text-stone-500 mb-3 font-sans">Webster 1828</h3>
          <div v-for="(entry, idx) in selected1828" :key="idx" class="mb-4 last:mb-0">
            <div class="text-sm italic text-stone-500 mb-1.5">{{ entry.pos }}</div>
            <ol class="list-decimal list-outside ml-6 text-base md:text-lg leading-relaxed text-stone-800 space-y-2">
              <li v-for="(def, di) in entry.definitions.slice(0, 8)" :key="di">
                <LinkedDefinition :text="def" />
              </li>
            </ol>
          </div>
        </section>

        <section v-if="selectedModern?.entries?.length">
          <h3 class="text-sm uppercase tracking-wider text-stone-500 mb-3 font-sans">Modern</h3>
          <div v-for="(entry, idx) in selectedModern.entries" :key="idx" class="mb-4 last:mb-0">
            <div class="text-sm italic text-stone-500 mb-1.5">{{ entry.pos }}</div>
            <ol class="list-decimal list-outside ml-6 text-base md:text-lg leading-relaxed text-stone-800 space-y-2">
              <li v-for="(def, di) in entry.definitions.slice(0, 4)" :key="di">
                <LinkedDefinition :text="def" />
              </li>
            </ol>
          </div>
        </section>
        <section v-else-if="selectedModern === null" class="text-sm text-stone-500 italic">
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
