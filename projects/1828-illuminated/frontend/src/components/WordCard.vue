<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink } from 'vue-router'
import { useWordData, type TierWord } from '@/composables/useWordData'

const props = defineProps<{
  word: string
  compact?: boolean
}>()

const data = useWordData()
const tier = computed<TierWord | undefined>(() => data.findWord(props.word))
const defs1828 = computed(() => data.get1828(props.word))
const defModern = computed(() => data.getModern(props.word))
</script>

<template>
  <article class="def-card p-5 space-y-4">
    <header class="flex items-baseline justify-between gap-3 border-b border-stone-200 pb-3">
      <h2 class="text-2xl font-serif">
        <RouterLink :to="`/word/${word}`" class="hover:text-amber-700">{{ word }}</RouterLink>
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
    <section v-if="defs1828.length">
      <h3 class="text-sm uppercase tracking-wider text-stone-500 mb-2 font-sans">Webster 1828</h3>
      <div v-for="(entry, idx) in defs1828" :key="idx" class="mb-3 last:mb-0">
        <div class="text-xs italic text-stone-500 mb-1">{{ entry.pos }}</div>
        <ol class="list-decimal list-outside ml-5 text-sm leading-relaxed text-stone-800 space-y-1">
          <li v-for="(def, di) in entry.definitions.slice(0, compact ? 2 : 12)" :key="di">{{ def }}</li>
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
    <section v-if="defModern?.entries?.length">
      <h3 class="text-sm uppercase tracking-wider text-stone-500 mb-2 font-sans">Modern</h3>
      <div v-for="(entry, idx) in defModern.entries" :key="idx" class="mb-3 last:mb-0">
        <div class="text-xs italic text-stone-500 mb-1">{{ entry.pos }}</div>
        <ol class="list-decimal list-outside ml-5 text-sm leading-relaxed text-stone-800 space-y-1">
          <li v-for="(def, di) in entry.definitions.slice(0, compact ? 1 : 6)" :key="di">{{ def }}</li>
        </ol>
      </div>
    </section>
    <section v-else class="text-sm text-stone-500 italic">
      <template v-if="defModern === null">
        No modern dictionary entry returned by the Free Dictionary API. This often means the word is sufficiently archaic that mainstream modern dictionaries don't cover it — a meaningful signal in itself.
      </template>
      <template v-else>
        Modern definition not yet fetched.
      </template>
    </section>

    <!-- Study cross-references -->
    <section v-if="tier && tier.studies.length">
      <h3 class="text-sm uppercase tracking-wider text-stone-500 mb-2 font-sans">Lensed in our studies</h3>
      <ul class="text-sm space-y-1">
        <li v-for="s in tier.studies" :key="s" class="text-stone-700">
          <code class="text-amber-700 bg-amber-50 px-1.5 py-0.5 rounded text-xs font-mono">{{ s }}</code>
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
