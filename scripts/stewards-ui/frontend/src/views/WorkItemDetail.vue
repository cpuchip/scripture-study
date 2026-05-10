<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import { api, type WorkItemDetail } from '@/api'

const route = useRoute()
const wi = ref<WorkItemDetail | null>(null)
const error = ref<string>('')
const loading = ref(false)

async function load(idOrSlug: string) {
  loading.value = true
  error.value = ''
  wi.value = null
  try {
    wi.value = await api.workItemGet(idOrSlug)
  } catch (e) {
    error.value = String(e)
  } finally {
    loading.value = false
  }
}

const idFromRoute = computed(() => String(route.params.id ?? ''))
onMounted(() => load(idFromRoute.value))
watch(idFromRoute, (v) => v && load(v))

function fmtJson(v: unknown) {
  return JSON.stringify(v, null, 2)
}
function fmtDate(s?: string) {
  if (!s) return ''
  return new Date(s).toLocaleString()
}
</script>

<template>
  <div class="space-y-6">
    <div>
      <RouterLink to="/work-items" class="text-xs text-zinc-500 hover:text-zinc-300">
        ← all work items
      </RouterLink>
    </div>

    <p v-if="loading" class="text-sm text-zinc-400">loading…</p>
    <p v-else-if="error" class="text-sm text-red-400">{{ error }}</p>

    <template v-if="wi">
      <header class="border-b border-zinc-800 pb-4">
        <h2 class="text-2xl font-semibold tracking-tight">{{ wi.slug }}</h2>
        <div class="text-xs text-zinc-500 mt-2 flex gap-3 font-mono flex-wrap">
          <span>pipeline: {{ wi.pipeline }}</span>
          <span>stage: {{ wi.current_stage }}</span>
          <span>status: {{ wi.status }}</span>
          <span v-if="wi.actor">actor: {{ wi.actor }}</span>
          <span>tokens: {{ wi.tokens_in.toLocaleString() }} in / {{ wi.tokens_out.toLocaleString() }} out</span>
          <span v-if="wi.token_budget">budget: {{ wi.token_budget.toLocaleString() }}</span>
          <span v-if="wi.completed_at">completed {{ fmtDate(wi.completed_at) }}</span>
        </div>
      </header>

      <section
        v-if="wi.error"
        class="rounded-md border border-red-900/40 bg-red-950/20 p-4 text-sm"
      >
        <div class="text-xs uppercase tracking-wide text-red-400 mb-1">Error</div>
        <pre class="whitespace-pre-wrap text-red-300 font-mono text-xs">{{ wi.error }}</pre>
      </section>

      <section class="rounded-md border border-zinc-800 bg-zinc-900/50 p-4">
        <div class="text-xs uppercase tracking-wide text-zinc-500 mb-2">Input</div>
        <pre class="text-xs font-mono text-zinc-300 whitespace-pre-wrap overflow-auto">{{ fmtJson(wi.input) }}</pre>
      </section>

      <section
        v-if="wi.stage_results"
        class="rounded-md border border-zinc-800 bg-zinc-900/50 p-4"
      >
        <div class="text-xs uppercase tracking-wide text-zinc-500 mb-2">Stage results</div>
        <pre class="text-xs font-mono text-zinc-300 whitespace-pre-wrap overflow-auto max-h-96">{{ fmtJson(wi.stage_results) }}</pre>
      </section>

      <section
        v-if="wi.session_ids?.length"
        class="rounded-md border border-zinc-800 bg-zinc-900/50 overflow-hidden"
      >
        <div class="px-4 py-3 border-b border-zinc-800">
          <h3 class="text-sm font-semibold">Sessions ({{ wi.session_ids.length }})</h3>
        </div>
        <ul class="divide-y divide-zinc-800/50">
          <li v-for="sid in wi.session_ids" :key="sid" class="px-4 py-2">
            <RouterLink
              :to="`/sessions/${encodeURIComponent(sid)}`"
              class="text-zinc-200 font-mono text-xs hover:text-white"
            >
              {{ sid }}
            </RouterLink>
          </li>
        </ul>
      </section>
    </template>
  </div>
</template>
