<script setup lang="ts">
// WordStudy — every verse in the canon that contains the word (or a
// stemmed variant). Backend endpoint: /api/scripture/word-study/:word.
//
// The other side of the click-mode coin (the other being WordDetail).
// In definition mode, clicking a word routes to /word/<word> for the
// 1828 entry. In scripture mode, clicking a word routes here.

import { computed, ref, watch } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import VerseList, { type VerseRow } from '@/components/VerseList.vue'
import { apiUrl } from '@/composables/useApiBase'
import { visit as studyVisit } from '@/composables/useStudyTree'

const route = useRoute()
const word = computed<string>(() => {
  const raw = Array.isArray(route.params.word) ? route.params.word[0] : route.params.word
  return (raw ?? '').toLowerCase()
})

interface WordStudyResponse {
  word: string
  found: boolean
  tier?: string
  occurrences: Array<{
    ref: string
    abbr_ref: string
    text: string
    verse?: number
  }>
  study_cross_refs?: Array<{ study: string; excerpt?: string }>
}

const loading = ref(false)
const error = ref<string>('')
const data = ref<WordStudyResponse | null>(null)

async function fetchOccurrences(w: string) {
  if (!w) return
  loading.value = true
  error.value = ''
  data.value = null
  try {
    const resp = await fetch(apiUrl(`/scripture/word-study/${encodeURIComponent(w)}`))
    if (!resp.ok) {
      error.value = `HTTP ${resp.status} — could not load occurrences for "${w}".`
      return
    }
    data.value = await resp.json() as WordStudyResponse
    // Tree-visit the scripture-study node so the chain remembers we
    // looked the word up THIS way (and not via /word/).
    studyVisit({ kind: 'word', word: w }, `${w} (scripture)`)
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    loading.value = false
  }
}

watch(word, fetchOccurrences, { immediate: true })

// Group occurrences by abbreviated book so the list is scannable.
const grouped = computed(() => {
  const out = new Map<string, Array<{ ref: string; abbr_ref: string; text: string; verse: number }>>()
  for (const o of data.value?.occurrences ?? []) {
    // abbr_ref like "dc/93:36" — group by the segment before the colon
    const [book = '?'] = (o.abbr_ref ?? '').split(':')
    const arr = out.get(book) ?? []
    arr.push({
      ref: o.ref,
      abbr_ref: o.abbr_ref,
      text: o.text,
      verse: o.verse ?? Number(o.abbr_ref?.split(':')[1] ?? 0),
    })
    out.set(book, arr)
  }
  return [...out.entries()]
})

function versesAsRows(verses: Array<{ verse: number; text: string }>): VerseRow[] {
  return verses.map(v => ({ verse: v.verse, text: v.text }))
}
</script>

<template>
  <div class="max-w-3xl mx-auto px-6 py-10">
    <div class="flex items-baseline justify-between gap-2 flex-wrap mb-6">
      <RouterLink to="/word" class="text-sm text-stone-500 hover:text-stone-900">← Back to word search</RouterLink>
      <div class="text-xs text-stone-500 italic">
        Scripture mode — click any word in a verse to chase its occurrences.
        <RouterLink :to="`/word/${word}`" class="text-amber-700 hover:underline ml-2">See its 1828 definition instead ↗</RouterLink>
      </div>
    </div>

    <header class="mb-6">
      <h1 class="text-3xl font-serif">{{ word }}</h1>
      <p class="text-stone-600 text-sm mt-1">
        <span v-if="data?.found">
          {{ data?.occurrences.length ?? 0 }}
          occurrence{{ (data?.occurrences.length ?? 0) === 1 ? '' : 's' }}
          across the canon
          <span v-if="data?.tier" class="ml-2 px-2 py-0.5 rounded-full bg-stone-100 text-stone-700 text-xs">Tier {{ data.tier }}</span>
        </span>
        <span v-else-if="!loading && !error" class="italic text-stone-500">
          No canonical occurrences for "{{ word }}".
        </span>
      </p>
    </header>

    <div v-if="loading" class="text-stone-500 italic">Loading occurrences…</div>

    <div v-if="error" class="def-card p-4 text-sm text-red-700 bg-red-50 border-red-200">
      {{ error }}
    </div>

    <div v-if="!loading && data?.found" class="space-y-8">
      <section v-for="[book, verses] in grouped" :key="book" class="def-card p-5">
        <h2 class="text-sm uppercase tracking-wider text-stone-500 font-sans mb-3">
          {{ book }} <span class="text-stone-400 normal-case">({{ verses.length }})</span>
        </h2>
        <VerseList :verses="versesAsRows(verses)" :abbr-ref="book" />
      </section>
    </div>

    <div
      v-if="!loading && data?.study_cross_refs?.length"
      class="mt-10 def-card p-5"
    >
      <h2 class="text-sm uppercase tracking-wider text-stone-500 font-sans mb-3">
        Lensed in our studies
      </h2>
      <ul class="space-y-2 text-sm">
        <li v-for="x in data.study_cross_refs" :key="x.study">
          <a
            :href="`https://github.com/cpuchip/scripture-study/blob/main/${x.study.replace(/^[./]+/, '')}`"
            target="_blank"
            rel="noopener"
            class="text-amber-700 hover:underline font-mono"
          >{{ x.study.replace(/^[./]+/, '') }} ↗</a>
          <blockquote v-if="x.excerpt" class="mt-1 text-stone-600 italic border-l-2 border-stone-200 pl-3">{{ x.excerpt }}</blockquote>
        </li>
      </ul>
    </div>
  </div>
</template>
