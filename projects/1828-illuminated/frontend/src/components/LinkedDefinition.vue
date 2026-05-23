<script setup lang="ts">
// Renders definition / body text with tier words as inline RouterLinks
// to /word/<word>. Lets the reader chain through the dictionary —
// click "intelligence" inside obtain's definition and jump there.
//
// Different from HighlightedText.vue: that one pops a CARD in the
// current view; this one navigates to a NEW route. The two have
// different semantics — HighlightedText for verse exploration where
// you stay in context; LinkedDefinition for definitions where you
// want to follow a chain.
import { computed } from 'vue'
import { tokenize, clickMode } from '@/composables/useWordData'

const props = defineProps<{
  text: string
}>()

const segments = computed(() => tokenize(props.text))

// Click mode controls routing: definition (default) → /word/X (1828 entry);
// scripture → /word-study/X (occurrences across the canon).
function linkFor(word: string): string {
  return clickMode.value === 'scripture' ? `/word-study/${word}` : `/word/${word}`
}
</script>

<template>
  <span>
    <template v-for="(seg, i) in segments" :key="i">
      <RouterLink
        v-if="seg.word"
        :to="linkFor(seg.word)"
        :class="['def-link', seg.tier === 'E' ? 'def-link-e' : '']"
        :title="clickMode === 'scripture'
          ? `Find ${seg.word} in scripture (Tier ${seg.tier})`
          : `See ${seg.word} (Tier ${seg.tier})`"
      >{{ seg.text }}</RouterLink>
      <template v-else>{{ seg.text }}</template>
    </template>
  </span>
</template>

<style scoped>
.def-link {
  color: inherit;
  border-bottom: 1px dotted var(--amber);
  text-decoration: none;
  transition: background-color 0.12s ease, color 0.12s ease;
  padding: 0 1px;
  border-radius: 2px;
}
.def-link:hover {
  background-color: var(--amber-soft);
  color: var(--ink);
}
/* Class-E words: any 1828 headword not in the curated tier list.
 * Lighter dotted underline (stone-300, not amber) so the page doesn't
 * vibrate when every other word lights up. Hover still warms to amber
 * so the affordance stays discoverable. */
.def-link-e {
  border-bottom-color: rgb(214 211 209);
  border-bottom-style: dotted;
}
.def-link-e:hover {
  border-bottom-color: var(--amber);
  background-color: var(--amber-soft);
}
</style>
