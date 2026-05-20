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
import { tokenize } from '@/composables/useWordData'

const props = defineProps<{
  text: string
}>()

const segments = computed(() => tokenize(props.text))
</script>

<template>
  <span>
    <template v-for="(seg, i) in segments" :key="i">
      <RouterLink
        v-if="seg.word"
        :to="`/word/${seg.word}`"
        class="def-link"
        :title="`See ${seg.word} (Tier ${seg.tier})`"
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
</style>
