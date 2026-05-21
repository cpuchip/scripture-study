<script setup lang="ts">
// Renders a list of verses each as its own paragraph with a small verse
// number marker, passing each verse text through HighlightedText for the
// tier-word click affordance. Replaces the verses.join(' ') wall of text
// that the canon-browse mode used to produce.
//
// The verse-number pill on the left also doubles as a "copy ref" affordance
// — click to copy the abbr ref (e.g. dc/93:36) to the clipboard. Honest
// feedback via a tiny "copied" tooltip-replacement.

import { ref } from 'vue'
import HighlightedText from './HighlightedText.vue'

export interface VerseRow {
  verse: number
  text: string
}

const props = defineProps<{
  verses: VerseRow[]
  /** abbr ref like "dc/93" — verse number appended for per-verse copy */
  abbrRef: string
}>()

const copiedVerse = ref<number | null>(null)

async function copyVerseRef(verse: number) {
  const ref = `${props.abbrRef}:${verse}`
  try {
    await navigator.clipboard.writeText(ref)
    copiedVerse.value = verse
    setTimeout(() => {
      if (copiedVerse.value === verse) copiedVerse.value = null
    }, 1400)
  } catch {
    // Clipboard unavailable in some contexts (insecure http, etc.) — silently ignore.
  }
}
</script>

<template>
  <div class="verse-list space-y-3">
    <div
      v-for="v in verses"
      :key="v.verse"
      :id="`v-${v.verse}`"
      class="verse-row flex gap-3 items-baseline group"
    >
      <button
        @click="copyVerseRef(v.verse)"
        class="verse-num font-sans text-xs text-stone-400 hover:text-amber-700 hover:bg-amber-50 rounded px-1.5 py-0.5 transition tabular-nums min-w-[2.25rem] text-right cursor-pointer"
        :title="`Copy ${abbrRef}:${v.verse} to clipboard`"
        :aria-label="`Copy reference for verse ${v.verse}`"
      >
        <span v-if="copiedVerse === v.verse" class="text-amber-700">✓</span>
        <span v-else>{{ v.verse }}</span>
      </button>
      <div class="flex-1">
        <HighlightedText :text="v.text" />
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Verse number tabular alignment so 1-digit and 3-digit numbers line up
   on a clean right edge. */
.verse-num { font-variant-numeric: tabular-nums; }
</style>
