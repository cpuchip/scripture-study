<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api, type PassRow } from '@/api'

const passes = ref<PassRow[]>([])
const error = ref<string>('')
const loading = ref(false)

async function load() {
  loading.value = true
  try {
    const r = await api.watchmanPasses(50)
    passes.value = r.items
  } catch (e) {
    error.value = String(e)
  } finally {
    loading.value = false
  }
}
onMounted(load)

function fmtDate(s?: string) {
  if (!s) return ''
  return new Date(s).toLocaleString()
}
function durSec(p: PassRow) {
  if (!p.started_at || !p.finished_at) return ''
  const s = new Date(p.started_at).getTime()
  const f = new Date(p.finished_at).getTime()
  return `${Math.round((f - s) / 1000)}s`
}
function verdictBadges(vc?: Record<string, number>) {
  if (!vc) return []
  return Object.entries(vc)
    .filter(([, n]) => n > 0)
    .sort((a, b) => b[1] - a[1])
}
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-baseline justify-between">
      <h2 class="text-2xl font-semibold tracking-tight">Watchman passes</h2>
      <button
        class="text-xs px-2 py-1 rounded border border-zinc-700 hover:bg-zinc-800"
        @click="load"
      >refresh</button>
    </div>

    <p v-if="loading" class="text-sm text-zinc-400">loading…</p>
    <p v-else-if="error" class="text-sm text-red-400">{{ error }}</p>

    <ul v-else class="space-y-2">
      <li
        v-for="p in passes"
        :key="p.pass_id"
        class="rounded-md border border-zinc-800 bg-zinc-900/50 p-3"
      >
        <div class="flex items-baseline gap-3">
          <span class="font-mono text-xs text-zinc-300">{{ p.pass_id }}</span>
          <span
            class="text-xs px-2 py-0.5 rounded"
            :class="{
              'bg-emerald-900/40 text-emerald-300': p.status === 'completed',
              'bg-amber-900/40 text-amber-300': p.status === 'in_progress',
              'bg-red-900/40 text-red-300': p.status === 'failed',
              'bg-zinc-800 text-zinc-400': p.status === 'queued',
            }"
          >{{ p.status }}</span>
          <span v-if="p.trigger" class="text-xs text-zinc-500">{{ p.trigger }}</span>
          <span class="ml-auto text-xs text-zinc-500">{{ fmtDate(p.started_at) }}</span>
        </div>
        <div class="text-xs text-zinc-400 mt-2 flex gap-3 flex-wrap">
          <span>{{ p.doc_count_done }}/{{ p.doc_count_planned }} docs</span>
          <span>{{ p.tokens_in.toLocaleString() }} in / {{ p.tokens_out.toLocaleString() }} out</span>
          <span v-if="durSec(p)">{{ durSec(p) }}</span>
          <span v-if="p.budget_stopped" class="text-amber-400">budget-stopped</span>
        </div>
        <div v-if="verdictBadges(p.verdict_counts).length" class="mt-2 flex gap-2 flex-wrap">
          <span
            v-for="[k, n] in verdictBadges(p.verdict_counts)"
            :key="k"
            class="text-xs px-2 py-0.5 rounded bg-zinc-800 text-zinc-300"
          >
            {{ k }}: {{ n }}
          </span>
        </div>
      </li>
    </ul>
  </div>
</template>
