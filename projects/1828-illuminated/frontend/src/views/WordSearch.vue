<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { useWordData, type DictSearchHit } from '@/composables/useWordData'
import WordCard from '@/components/WordCard.vue'

const data = useWordData()
const query = ref('')
// All tiers ON by default — readers expect to find what they're looking
// for without first realizing they need to toggle filters. D + E were
// off-by-default historically; that hid the headline class-E reach feature
// from new readers.
const selectedTiers = ref<Set<string>>(new Set(['A++', 'A+', 'B', 'C', 'D', 'E']))

// Tier-list matches (synchronous, drives the curated section).
const matches = computed(() => {
  const q = query.value.trim().toLowerCase()
  if (!q) return data.allByTier().filter(w => selectedTiers.value.has(w.tier))
  return data.searchPrefix(q, 60).filter(w => selectedTiers.value.has(w.tier))
})

// Class-E reach — every 1828 headword the backend has, regardless of
// whether we curated it. Surfaced below the primary tier-words section
// so the "any word in the 1828 is reachable" property is visible.
const otherMatches = ref<DictSearchHit[]>([])
const otherLoading = ref(false)
let lastQueryRequested = ''

watch(query, async (q) => {
  const trimmed = q.trim().toLowerCase()
  lastQueryRequested = trimmed
  if (trimmed.length < 1) {
    otherMatches.value = []
    return
  }
  otherLoading.value = true
  try {
    // 60 hits — gives "starts with a" room to surface a meaningful slice
    // without flooding. Increased from 40 to handle single-letter prefixes
    // that match thousands of 1828 entries.
    const resp = await data.searchDict(trimmed, 60)
    // Discard stale results if the query has moved on while the fetch was in flight.
    if (lastQueryRequested !== trimmed) return
    // Filter out anything already in the tier section to avoid duplicating
    // the curated list — the secondary section is "everything else."
    const curatedSet = new Set(matches.value.map(m => m.word))
    otherMatches.value = resp.all_1828_results.filter(r => !curatedSet.has(r.word))
  } catch {
    otherMatches.value = []
  } finally {
    if (lastQueryRequested === trimmed) otherLoading.value = false
  }
})

function toggleTier(t: string) {
  if (selectedTiers.value.has(t)) selectedTiers.value.delete(t)
  else selectedTiers.value.add(t)
  selectedTiers.value = new Set(selectedTiers.value)
}
</script>

<template>
  <div class="max-w-5xl mx-auto px-6 py-10">
    <header class="mb-8">
      <h1 class="text-3xl font-serif mb-2">Word search</h1>
      <p class="text-stone-600">
        Type any word — the full <strong>{{ '98,828' }}-headword 1828 corpus</strong> is searchable here, not just our curated tier list.
        Try <code class="bg-stone-100 px-1.5 py-0.5 rounded text-sm">gainsay</code>, <code class="bg-stone-100 px-1.5 py-0.5 rounded text-sm">peradventure</code>, or <code class="bg-stone-100 px-1.5 py-0.5 rounded text-sm">wist</code> — they're not in our tier list but the 1828 still has them.
      </p>
    </header>

    <div class="mb-8 space-y-4">
      <input
        v-model="query"
        type="search"
        placeholder="Start typing a word… (e.g. obtain, charity, gainsay, peradventure)"
        class="w-full px-4 py-3 rounded-lg border border-stone-300 focus:border-amber-500 focus:outline-none bg-white text-lg font-serif"
        autofocus
      />
      <div class="flex flex-wrap gap-2 items-center text-sm">
        <span class="text-stone-500 mr-2">Tiers:</span>
        <button
          v-for="t in ['A++','A+','B','C','D','E']"
          :key="t"
          @click="toggleTier(t)"
          class="px-3 py-1 rounded-full border text-xs font-medium transition"
          :class="selectedTiers.has(t)
            ? 'bg-amber-100 border-amber-400 text-amber-900'
            : 'bg-white border-stone-300 text-stone-500 hover:border-stone-400'"
          :title="t === 'E' ? 'Class-E: any 1828 headword not in our curated tier list' : ''"
        >
          {{ t }}
          <span class="ml-1 text-stone-400">{{ data.tierCounts[t as keyof typeof data.tierCounts] }}</span>
        </button>
      </div>
    </div>

    <!-- Primary section: words we've curated -->
    <section>
      <h2 class="text-xs uppercase tracking-wider text-stone-500 mb-3 font-sans">
        <span v-if="query.trim()">Curated tier words matching "{{ query.trim() }}"</span>
        <span v-else>Words we've curated</span>
      </h2>

      <div v-if="!matches.length" class="text-stone-500 italic">
        <template v-if="query.trim()">No tier words match. (See class-E results below if any.)</template>
        <template v-else>No matches. Try a different prefix or enable more tiers.</template>
      </div>

      <div v-else-if="!query" class="space-y-3">
        <div class="text-sm text-stone-500 mb-2">{{ matches.length }} words across selected tiers — click any to open.</div>
        <div class="flex flex-wrap gap-2">
          <RouterLink
            v-for="w in matches.slice(0, 200)"
            :key="w.word"
            :to="`/word/${w.word}`"
            class="px-3 py-1.5 bg-white border border-stone-300 rounded-full text-sm hover:border-amber-500 hover:bg-amber-50 transition"
          >
            <span class="font-serif">{{ w.word }}</span>
            <span class="text-xs text-stone-500 ml-1.5">{{ w.tier }}</span>
          </RouterLink>
        </div>
        <div v-if="matches.length > 200" class="text-xs text-stone-500 italic mt-3">
          … and {{ matches.length - 200 }} more. Type to narrow.
        </div>
      </div>

      <div v-else class="space-y-6">
        <WordCard
          v-for="w in matches.slice(0, 10)"
          :key="w.word"
          :word="w.word"
          compact
        />
        <div v-if="matches.length > 10" class="text-sm text-stone-500 italic text-center">
          {{ matches.length - 10 }} more matches — refine the search to see them.
        </div>
      </div>
    </section>

    <!-- Class-E reach: every 1828 headword. Shows as soon as the reader types
         a single letter — was previously gated at length >= 2 which hid the
         feature behind discovery friction. Section is also gated on the E
         tier toggle so readers who want curated-only can hide it. -->
    <section v-if="query.trim().length >= 1 && selectedTiers.has('E')" class="mt-10 border-t border-stone-200 pt-8">
      <h2 class="text-xs uppercase tracking-wider text-stone-500 mb-1 font-sans flex items-baseline gap-2 flex-wrap">
        <span>Full 1828 corpus matching "{{ query.trim() }}"</span>
        <span v-if="otherMatches.length" class="text-stone-400 normal-case">— {{ otherMatches.length }}{{ otherMatches.length === 60 ? '+' : '' }} headword{{ otherMatches.length === 1 ? '' : 's' }}</span>
      </h2>
      <p class="text-xs text-stone-500 mb-4 italic">
        Class-E reach — any of the 98,828 1828 headwords, not just our 859 tier-curated ones.
      </p>

      <div v-if="otherLoading" class="text-sm text-stone-400 italic">Searching the 1828 corpus…</div>
      <div v-else-if="!otherMatches.length" class="text-sm text-stone-500 italic">
        No additional 1828 entries match that prefix.
      </div>
      <div v-else class="flex flex-wrap gap-2">
        <RouterLink
          v-for="w in otherMatches"
          :key="w.word"
          :to="`/word/${w.word}`"
          class="px-3 py-1.5 bg-white border border-stone-200 rounded-full text-sm hover:border-amber-500 hover:bg-amber-50 text-stone-600 transition"
          :title="`Open the 1828 entry for ${w.word}`"
        >
          <span class="font-serif">{{ w.word }}</span>
        </RouterLink>
        <div v-if="otherMatches.length === 60" class="text-xs text-stone-500 italic w-full mt-2">
          Showing the first 60. Type a longer prefix to narrow.
        </div>
      </div>
    </section>
  </div>
</template>
