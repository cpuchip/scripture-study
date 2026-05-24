<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { tokenize, selectWord, clickMode } from '@/composables/useWordData'

const props = defineProps<{
  text: string
}>()

const router = useRouter()
const segments = computed(() => tokenize(props.text))

function onWordClick(word: string, event: MouseEvent | KeyboardEvent) {
  const toggleAction = event.shiftKey || event.ctrlKey || event.altKey || event.metaKey
  const action = toggleAction
    ? (clickMode.value === 'scripture' ? 'definition' : 'scripture')
    : clickMode.value

  if (action === 'scripture') {
    router.push({ name: 'word-study', params: { word } })
    return
  }
  // Definition mode (default) — keep the in-place WordCard behavior.
  selectWord(word)
  setTimeout(() => {
    document.getElementById('selected-word-card')?.scrollIntoView({ behavior: 'smooth', block: 'nearest' })
  }, 50)
}
</script>

<style scoped>
/* Class-E words inside verse text: subtler than tier A/B/C/D so a verse
 * full of 1828-known words doesn't visually vibrate. Stone-400 dotted
 * underline (vs amber dashed for tier words). Hover still warms to amber. */
.highlight-tier-E {
  background-color: transparent !important;
  border-bottom: 1px dotted rgb(168 162 158);
  padding-bottom: 1px;
}
.highlight-tier-E:hover {
  background-color: var(--amber-soft) !important;
  border-bottom-color: var(--amber);
}
</style>

<template>
  <p class="font-serif text-lg leading-relaxed text-stone-900">
    <template v-for="(seg, i) in segments" :key="i">
      <span
        v-if="seg.word"
        class="highlight"
        :class="[
          seg.tier === 'A++' || seg.tier === 'A+' ? 'highlight-tier-A' : '',
          seg.tier === 'E' ? 'highlight-tier-E' : '',
        ]"
        :title="clickMode === 'scripture'
          ? `Tier ${seg.tier} — click to find ${seg.word} in scripture (Shift+click for definition)`
          : `Tier ${seg.tier} — click for definition (Shift+click to find ${seg.word} in scripture)`"
        role="button"
        tabindex="0"
        @click="onWordClick(seg.word, $event)"
        @keydown.enter="onWordClick(seg.word, $event)"
      >{{ seg.text }}</span>
      <template v-else>{{ seg.text }}</template>
    </template>
  </p>
</template>
