<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import {
  useWordData,
  type Def1828Entry,
  type ModernEntry,
  type TierWord,
} from '@/composables/useWordData'
import LinkedDefinition from './LinkedDefinition.vue'

const props = defineProps<{
  word: string
  compact?: boolean
}>()

const data = useWordData()
// Tier metadata stays synchronous — driven by the static tier-words.json bundle.
const tier = computed<TierWord | undefined>(() => data.findWord(props.word))

// 1828 + modern are now backend-fetched per word. We expose a loading state
// per section so the empty-state message ("No 1828 entry on file") doesn't
// flash before the response lands.
const defs1828 = ref<Def1828Entry[]>([])
const stemMatched = ref<string | null>(null)
const loading1828 = ref(false)
const defsModern = ref<ModernEntry[]>([])
const modernSource = ref<'cache' | 'fetched' | 'none' | 'rate_limited' | ''>('')
const modernError = ref<string | null>(null)
const loadingModern = ref(false)

watch(
  () => props.word,
  async (w) => {
    if (!w) {
      defs1828.value = []
      stemMatched.value = null
      defsModern.value = []
      modernSource.value = ''
      modernError.value = null
      return
    }
    loading1828.value = true
    loadingModern.value = true
    // Fire both in parallel — they're independent endpoints.
    const [r1828, rModern] = await Promise.all([data.get1828(w), data.getModern(w)])
    defs1828.value = r1828.entries
    stemMatched.value = r1828.stem_matched
    loading1828.value = false
    defsModern.value = rModern.entries ?? []
    modernSource.value = rModern.source
    modernError.value = rModern.error ?? null
    loadingModern.value = false
  },
  { immediate: true },
)

const route = useRoute()
const isDefinitionView = computed(() => {
  return route.name === 'word-detail' && route.params.word?.toString().toLowerCase() === props.word.toLowerCase()
})
const cardTitleLink = computed(() => {
  return isDefinitionView.value ? `/word-study/${props.word}` : `/word/${props.word}`
})
const cardTitleTooltip = computed(() => {
  return isDefinitionView.value ? 'Click to see scripture occurrences for this word' : 'Click to see 1828 definition'
})

// Studies live in the workspace repo at github.com/cpuchip/scripture-study.
// Link out so readers can read the original lensing context.
function studyHref(path: string): string {
  return `https://github.com/cpuchip/scripture-study/blob/main/${path}`
}
</script>

<template>
  <article class="def-card p-5 space-y-4">
    <header class="flex items-baseline justify-between gap-3 border-b border-stone-200 pb-3">
      <h2 class="text-2xl font-serif">
        <RouterLink :to="cardTitleLink" :title="cardTitleTooltip" class="hover:text-amber-700">{{ word }}</RouterLink>
      </h2>
      <div v-if="tier" class="text-xs text-stone-500 flex items-baseline gap-3">
        <span class="px-2 py-0.5 rounded-full text-stone-700 bg-stone-100 border border-stone-200">
          Tier {{ tier.tier }}
        </span>
        <span v-if="tier.studies.length" class="text-stone-500">
          lensed in {{ tier.studies.length }} stud{{ tier.studies.length === 1 ? 'y' : 'ies' }}
        </span>
      </div>
    </header>

    <!-- 1828 -->
    <section v-if="loading1828" class="text-sm text-stone-400 italic">
      Loading 1828 entry…
    </section>
    <section v-else-if="defs1828.length">
      <h3 class="text-sm uppercase tracking-wider text-stone-500 mb-2 font-sans">
        Webster 1828
        <span v-if="stemMatched" class="ml-2 text-xs normal-case text-stone-500 italic">
          (showing entry for <code class="bg-stone-100 px-1 rounded">{{ stemMatched }}</code>)
        </span>
      </h3>
      <div v-for="(entry, idx) in defs1828" :key="idx" class="mb-3 last:mb-0">
        <div class="text-xs italic text-stone-500 mb-1">{{ entry.pos }}</div>
        <ol class="list-decimal list-outside ml-5 text-sm leading-relaxed text-stone-800 space-y-1">
          <li v-for="(def, di) in entry.definitions.slice(0, compact ? 2 : 12)" :key="di">
            <LinkedDefinition :text="def" />
          </li>
        </ol>
        <div v-if="compact && entry.definitions.length > 2" class="text-xs text-stone-500 mt-1 italic">
          (+{{ entry.definitions.length - 2 }} more senses — open the word card for the full entry)
        </div>
      </div>
    </section>
    <section v-else class="text-sm text-stone-500 italic">
      No 1828 entry on file for "{{ word }}".
    </section>

    <!-- Modern -->
    <section v-if="loadingModern" class="text-sm text-stone-400 italic">
      Loading modern entry…
    </section>
    <section v-else-if="defsModern.length">
      <h3 class="text-sm uppercase tracking-wider text-stone-500 mb-2 font-sans">
        Modern
        <span v-if="modernSource === 'fetched'" class="ml-2 text-xs normal-case text-stone-400 italic">
          (just fetched from the Free Dictionary API)
        </span>
      </h3>
      <div v-for="(entry, idx) in defsModern" :key="idx" class="mb-3 last:mb-0">
        <div class="text-xs italic text-stone-500 mb-1">{{ entry.pos }}</div>
        <ol class="list-decimal list-outside ml-5 text-sm leading-relaxed text-stone-800 space-y-1">
          <li v-for="(def, di) in entry.definitions.slice(0, compact ? 1 : 6)" :key="di">
            <LinkedDefinition :text="def" />
          </li>
        </ol>
      </div>
    </section>
    <section v-else class="text-sm text-stone-500 italic">
      <template v-if="modernError">
        Modern lookup failed: {{ modernError }}
      </template>
      <template v-else-if="modernSource === 'none'">
        No modern dictionary entry returned by the Free Dictionary API. This often means the word is sufficiently archaic that mainstream modern dictionaries don't cover it — a meaningful signal in itself.
      </template>
      <template v-else-if="modernSource === 'rate_limited'">
        Modern lookups are paused for the rest of the day — the Free Dictionary API daily cap was reached. Cached entries still serve normally.
      </template>
      <template v-else>
        Modern definition not available.
      </template>
    </section>

    <!-- Study cross-references -->
    <section v-if="tier && tier.studies.length">
      <h3 class="text-sm uppercase tracking-wider text-stone-500 mb-2 font-sans">Lensed in our studies</h3>
      <ul class="text-sm space-y-1">
        <li v-for="s in tier.studies" :key="s" class="text-stone-700">
          <a
            :href="studyHref(s)"
            target="_blank"
            rel="noopener"
            class="inline-flex items-baseline gap-1 text-amber-700 bg-amber-50 hover:bg-amber-100 px-1.5 py-0.5 rounded text-xs font-mono transition"
            :title="`Open ${s} on github.com/cpuchip/scripture-study`"
          >
            {{ s }}
            <span class="text-amber-600 not-italic">↗</span>
          </a>
        </li>
      </ul>
      <details v-if="tier.study_excerpts.length" class="mt-3">
        <summary class="text-xs text-stone-500 cursor-pointer hover:text-stone-700">Sample excerpts</summary>
        <blockquote
          v-for="(ex, i) in tier.study_excerpts"
          :key="i"
          class="text-xs text-stone-600 italic mt-2 border-l-2 border-stone-300 pl-3"
        >{{ ex }}</blockquote>
      </details>
    </section>
  </article>
</template>
