<script setup lang="ts">
import { ref, watch, onMounted, computed } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import MarkdownIt from 'markdown-it'
import { api, type StudyDetail } from '@/api'

const route = useRoute()
const md = new MarkdownIt({ html: false, linkify: true, typographer: false })

const study = ref<StudyDetail | null>(null)
const error = ref<string>('')
const loading = ref(false)

async function load(slug: string) {
  loading.value = true
  error.value = ''
  study.value = null
  try {
    study.value = await api.studyGet(slug)
  } catch (e) {
    error.value = String(e)
  } finally {
    loading.value = false
  }
}

const slugFromRoute = computed(() => String(route.params.slug ?? ''))

onMounted(() => load(slugFromRoute.value))
watch(slugFromRoute, (s) => {
  if (s) load(s)
})

const renderedBody = computed(() => {
  if (!study.value?.body) return ''
  return md.render(study.value.body)
})

function fmtDate(s?: string) {
  if (!s) return ''
  return new Date(s).toLocaleString()
}
</script>

<template>
  <div class="space-y-6">
    <div>
      <RouterLink to="/studies" class="text-xs text-zinc-500 hover:text-zinc-300">
        ← all studies
      </RouterLink>
    </div>

    <p v-if="loading" class="text-sm text-zinc-400">loading…</p>
    <p v-else-if="error" class="text-sm text-red-400">{{ error }}</p>

    <template v-if="study">
      <header class="border-b border-zinc-800 pb-4">
        <h2 class="text-2xl font-semibold tracking-tight">{{ study.title || study.slug }}</h2>
        <div class="text-xs text-zinc-500 mt-2 flex gap-3 font-mono">
          <span>kind: {{ study.kind }}</span>
          <span>slug: {{ study.slug }}</span>
          <span v-if="study.updated_at">updated: {{ fmtDate(study.updated_at) }}</span>
        </div>
      </header>

      <article
        class="prose prose-invert prose-zinc max-w-none
               prose-headings:font-semibold prose-headings:tracking-tight
               prose-h1:text-2xl prose-h2:text-xl prose-h3:text-lg
               prose-p:text-zinc-200 prose-p:leading-relaxed
               prose-blockquote:border-l-zinc-700 prose-blockquote:text-zinc-300
               prose-code:text-zinc-300 prose-code:bg-zinc-900 prose-code:px-1 prose-code:rounded
               prose-pre:bg-zinc-900 prose-pre:border prose-pre:border-zinc-800
               prose-a:text-zinc-100 prose-a:underline prose-a:decoration-zinc-600 hover:prose-a:decoration-zinc-300
               prose-strong:text-zinc-100"
        v-html="renderedBody"
      ></article>

      <!-- Citations -->
      <section
        v-if="study.citations.length"
        class="rounded-md border border-zinc-800 bg-zinc-900/50 overflow-hidden"
      >
        <div class="px-4 py-3 border-b border-zinc-800">
          <h3 class="text-sm font-semibold">Citations ({{ study.citations.length }})</h3>
        </div>
        <ul class="divide-y divide-zinc-800/50">
          <li
            v-for="(c, i) in study.citations"
            :key="i"
            class="px-4 py-2 text-sm flex items-baseline gap-3"
          >
            <span class="font-mono text-zinc-300">{{ c.ref }}</span>
            <span v-if="c.count" class="text-xs text-zinc-500 ml-auto">
              cited {{ c.count }}×
            </span>
          </li>
        </ul>
      </section>

      <!-- Similar -->
      <section
        v-if="study.similar.length"
        class="rounded-md border border-zinc-800 bg-zinc-900/50 overflow-hidden"
      >
        <div class="px-4 py-3 border-b border-zinc-800">
          <h3 class="text-sm font-semibold">Similar studies</h3>
        </div>
        <ul class="divide-y divide-zinc-800/50">
          <li v-for="h in study.similar" :key="h.slug" class="px-4 py-2 text-sm">
            <RouterLink
              :to="`/studies/${encodeURIComponent(h.slug)}`"
              class="flex items-baseline gap-3"
            >
              <span class="text-zinc-200">{{ h.title || h.slug }}</span>
              <span v-if="h.distance" class="ml-auto text-xs text-zinc-500 tabular-nums">
                {{ h.distance.toFixed(3) }}
              </span>
            </RouterLink>
          </li>
        </ul>
      </section>
    </template>
  </div>
</template>
