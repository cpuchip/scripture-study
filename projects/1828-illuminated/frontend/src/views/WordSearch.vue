<script setup lang="ts">
import { ref, computed } from 'vue'
import { RouterLink } from 'vue-router'
import { useWordData } from '@/composables/useWordData'
import WordCard from '@/components/WordCard.vue'

const data = useWordData()
const query = ref('')
const selectedTiers = ref<Set<string>>(new Set(['A++', 'A+', 'B', 'C']))

const matches = computed(() => {
  const q = query.value.trim().toLowerCase()
  if (!q) return data.allByTier().filter(w => selectedTiers.value.has(w.tier))
  return data.searchPrefix(q, 60).filter(w => selectedTiers.value.has(w.tier))
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
      <p class="text-stone-600">Look up any tier word. 1828 and modern definitions side by side, plus the studies that lensed it.</p>
    </header>

    <div class="mb-8 space-y-4">
      <input
        v-model="query"
        type="search"
        placeholder="Start typing a word… (e.g. obtain, charity, intelligence)"
        class="w-full px-4 py-3 rounded-lg border border-stone-300 focus:border-amber-500 focus:outline-none bg-white text-lg font-serif"
        autofocus
      />
      <div class="flex flex-wrap gap-2 items-center text-sm">
        <span class="text-stone-500 mr-2">Tiers:</span>
        <button
          v-for="t in ['A++','A+','B','C','D']"
          :key="t"
          @click="toggleTier(t)"
          class="px-3 py-1 rounded-full border text-xs font-medium transition"
          :class="selectedTiers.has(t)
            ? 'bg-amber-100 border-amber-400 text-amber-900'
            : 'bg-white border-stone-300 text-stone-500 hover:border-stone-400'"
        >
          {{ t }}
          <span class="ml-1 text-stone-400">{{ data.tierCounts[t as keyof typeof data.tierCounts] }}</span>
        </button>
      </div>
    </div>

    <div v-if="!matches.length" class="text-stone-500 italic">No matches. Try a different prefix or enable more tiers.</div>

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
  </div>
</template>
