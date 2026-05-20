<script setup lang="ts">
import { RouterLink } from 'vue-router'
import { useWordData } from '@/composables/useWordData'

const data = useWordData()
const tierA = data.allByTier().filter(w => w.tier === 'A++' || w.tier === 'A+')
</script>

<template>
  <div class="max-w-5xl mx-auto px-6 py-12">
    <div class="mb-12 max-w-3xl">
      <h1 class="text-4xl font-serif tracking-tight mb-4">Scripture in its Restoration-era language frame.</h1>
      <p class="text-lg leading-relaxed text-stone-700 mb-3">
        Words drift. The 1828 Webster dictionary catches the meanings the early Saints heard when the scriptures were translated and revealed. This tool surfaces the words where 1828 and modern English diverge — so you can read with the meaning the original audience had.
      </p>
      <p class="text-stone-600">
        1828 illuminates — it does not decode. The goal is opening the conversation, not encoding a hidden meaning.
      </p>
    </div>

    <div class="grid sm:grid-cols-3 gap-6 mb-12">
      <RouterLink to="/word" class="def-card p-6 hover:border-amber-400 transition group">
        <h2 class="font-serif text-xl mb-2 group-hover:text-amber-700">Word search</h2>
        <p class="text-sm text-stone-600">
          Look up any of the {{ data.tierWords.length }} curated words. See 1828 and modern definitions side by side, with the studies that lensed each word.
        </p>
      </RouterLink>
      <RouterLink to="/verse" class="def-card p-6 hover:border-amber-400 transition group">
        <h2 class="font-serif text-xl mb-2 group-hover:text-amber-700">Verse explorer</h2>
        <p class="text-sm text-stone-600">
          Pick a demo verse or paste your own. Words with meaning-shift are highlighted; click any to see both senses.
        </p>
      </RouterLink>
      <RouterLink to="/present" class="def-card p-6 hover:border-amber-400 transition group">
        <h2 class="font-serif text-xl mb-2 group-hover:text-amber-700">Presentation mode <span class="text-xs text-stone-400">tablet-friendly</span></h2>
        <p class="text-sm text-stone-600">
          Fullscreen, large-type rendering of the demo verses for teaching. Arrow keys to navigate; tap any word for the 1828 definition.
        </p>
      </RouterLink>
    </div>

    <section class="mb-12">
      <h2 class="font-serif text-2xl mb-3">The high-confidence vocabulary</h2>
      <p class="text-sm text-stone-600 mb-4">
        Tier A++ and A+ words have two signals converging: our own substrate studies have lensed them, AND the 1828 entry itself carries meaning-shift markers. Click to explore.
      </p>
      <div class="flex flex-wrap gap-2">
        <RouterLink
          v-for="w in tierA"
          :key="w.word"
          :to="`/word/${w.word}`"
          class="px-3 py-1.5 bg-white border border-stone-300 rounded-full text-sm hover:border-amber-500 hover:bg-amber-50 hover:text-amber-900 transition"
        >
          <span class="font-serif">{{ w.word }}</span>
          <span class="text-xs text-stone-500 ml-1.5">{{ w.tier }}</span>
        </RouterLink>
      </div>
    </section>

    <section class="def-card p-6 bg-stone-50/50">
      <h2 class="font-serif text-xl mb-3">Project state</h2>
      <ul class="text-sm space-y-1.5 text-stone-700">
        <li><span class="text-stone-500 w-32 inline-block">Tier counts:</span>
          A++ = {{ data.tierCounts['A++'] }} · A+ = {{ data.tierCounts['A+'] }} · B = {{ data.tierCounts['B'] }} · C = {{ data.tierCounts['C'] }} · D = {{ data.tierCounts['D'] }}</li>
        <li><span class="text-stone-500 w-32 inline-block">1828 source:</span> webster1828.json (98,828 headwords), local copy, never hits the LM Studio embeddings pipeline</li>
        <li><span class="text-stone-500 w-32 inline-block">Modern source:</span> Free Dictionary API (pre-fetched at build time, no runtime calls)</li>
        <li><span class="text-stone-500 w-32 inline-block">Scripture text:</span> via churchofjesuschrist.org (not bundled)</li>
      </ul>
    </section>
  </div>
</template>
