<script setup lang="ts">
// Route params are READ ONCE at setup unless wrapped in computed —
// vue-router reuses the same component instance when only the param changes
// (e.g. /word/as → /word/abide), so a plain const captures the initial value
// and never updates. Fix: derive `word` and `tier` as computed refs.
import { computed, watch } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import WordCard from '@/components/WordCard.vue'
import { useWordData } from '@/composables/useWordData'
import { visit as studyVisit } from '@/composables/useStudyTree'

const route = useRoute()
const data = useWordData()

const word = computed<string>(() => {
  const raw = Array.isArray(route.params.word) ? route.params.word[0] : route.params.word
  return (raw ?? '').toLowerCase()
})
const tier = computed(() => data.findWord(word.value))

// Add to the study tree on every word visit. Idempotency in useStudyTree
// dedupes if this exact child already exists under the active node — so
// clicking back-and-forth between two words just toggles which is active,
// not a runaway tree of duplicates.
watch(
  word,
  (w) => {
    if (w) studyVisit({ kind: 'word', word: w })
  },
  { immediate: true },
)
</script>

<template>
  <div class="max-w-3xl mx-auto px-6 py-10">
    <RouterLink to="/word" class="text-sm text-stone-500 hover:text-stone-900">← Back to word search</RouterLink>
    <div class="mt-6">
      <WordCard :word="word" />
      <div v-if="!tier" class="mt-6 text-sm text-stone-500 italic">
        "{{ word }}" isn't in our tiered highlight list. It may still have a 1828 entry — check the card above. If it appears in scripture but doesn't have a 1828 entry, that itself is signal (proper noun, Restoration-specific coinage, etc.).
      </div>
    </div>
  </div>
</template>
