<script setup lang="ts">
import { computed } from 'vue'
import { tokenize, selectWord } from '@/composables/useWordData'

const props = defineProps<{
  text: string
}>()

const segments = computed(() => tokenize(props.text))

function onWordClick(word: string) {
  selectWord(word)
  // Scroll the word card into view if it's offscreen
  setTimeout(() => {
    document.getElementById('selected-word-card')?.scrollIntoView({ behavior: 'smooth', block: 'nearest' })
  }, 50)
}
</script>

<template>
  <p class="font-serif text-lg leading-relaxed text-stone-900">
    <template v-for="(seg, i) in segments" :key="i">
      <span
        v-if="seg.word"
        class="highlight"
        :class="seg.tier === 'A++' || seg.tier === 'A+' ? 'highlight-tier-A' : ''"
        :title="`Tier ${seg.tier} — click for definition`"
        role="button"
        tabindex="0"
        @click="onWordClick(seg.word)"
        @keydown.enter="onWordClick(seg.word)"
      >{{ seg.text }}</span>
      <template v-else>{{ seg.text }}</template>
    </template>
  </p>
</template>
