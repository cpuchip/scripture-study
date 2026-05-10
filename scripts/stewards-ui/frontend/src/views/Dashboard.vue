<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { api, type DashboardResp } from '@/api'

const data = ref<DashboardResp | null>(null)
const error = ref<string>('')
const loading = ref(false)

async function load() {
  loading.value = true
  error.value = ''
  try {
    data.value = await api.dashboard()
  } catch (e) {
    error.value = String(e)
  } finally {
    loading.value = false
  }
}

let timer: number | undefined
onMounted(() => {
  load()
  // 5s auto-refresh — cheap (single dashboard endpoint)
  timer = window.setInterval(load, 5000)
})
onUnmounted(() => {
  if (timer) window.clearInterval(timer)
})

function fmtRelative(s?: string) {
  if (!s) return ''
  const d = new Date(s)
  if (isNaN(d.getTime())) return s
  const diffMs = Date.now() - d.getTime()
  const sec = Math.floor(diffMs / 1000)
  if (sec < 60) return `${sec}s ago`
  const min = Math.floor(sec / 60)
  if (min < 60) return `${min}m ago`
  const hr = Math.floor(min / 60)
  if (hr < 24) return `${hr}h ago`
  const days = Math.floor(hr / 24)
  return `${days}d ago`
}

const inFlightCount = computed(() => data.value?.in_flight?.length ?? 0)
const errorCount = computed(() => data.value?.recent_errors?.length ?? 0)
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-baseline justify-between">
      <h2 class="text-2xl font-semibold tracking-tight">Dashboard</h2>
      <div class="text-xs text-zinc-500 flex items-center gap-3">
        <span v-if="loading" class="text-zinc-400">refreshing…</span>
        <span v-else-if="error" class="text-red-400">{{ error }}</span>
        <span v-else-if="data">updated {{ fmtRelative(new Date(data.fetched_at_ms).toISOString()) }}</span>
        <button
          class="text-xs px-2 py-1 rounded border border-zinc-700 hover:bg-zinc-800"
          @click="load"
        >
          refresh
        </button>
      </div>
    </div>

    <!-- Top row: 4 status cards -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
      <!-- pg health -->
      <div class="rounded-md border border-zinc-800 bg-zinc-900/50 p-4">
        <div class="text-xs uppercase tracking-wide text-zinc-500 mb-2">Postgres</div>
        <div class="flex items-center gap-2">
          <span
            class="inline-block w-2 h-2 rounded-full"
            :class="data?.pg.ok ? 'bg-emerald-500' : 'bg-red-500'"
          ></span>
          <span class="text-lg font-semibold">{{ data?.pg.ok ? 'healthy' : 'down' }}</span>
        </div>
        <div v-if="data?.pg.error" class="text-xs text-red-400 mt-1">
          {{ data.pg.error }}
        </div>
      </div>

      <!-- soak status -->
      <div class="rounded-md border border-zinc-800 bg-zinc-900/50 p-4">
        <div class="text-xs uppercase tracking-wide text-zinc-500 mb-2">Soak</div>
        <div class="flex items-center gap-2">
          <span
            class="inline-block w-2 h-2 rounded-full"
            :class="data?.soak.schedule_enabled ? 'bg-emerald-500' : 'bg-zinc-600'"
          ></span>
          <span class="text-lg font-semibold">
            {{ data?.soak.schedule_enabled ? 'on' : 'paused' }}
          </span>
        </div>
        <div class="text-xs text-zinc-400 mt-1">
          last: {{ fmtRelative(data?.soak.last_pass_started_at) || '—' }}
        </div>
      </div>

      <!-- dirty queue depth -->
      <div class="rounded-md border border-zinc-800 bg-zinc-900/50 p-4">
        <div class="text-xs uppercase tracking-wide text-zinc-500 mb-2">Dirty queue</div>
        <div class="text-2xl font-semibold tabular-nums">
          {{ data?.soak.dirty_queue_depth ?? '—' }}
        </div>
        <div class="text-xs text-zinc-400 mt-1">docs awaiting watchman</div>
      </div>

      <!-- in-flight -->
      <div class="rounded-md border border-zinc-800 bg-zinc-900/50 p-4">
        <div class="text-xs uppercase tracking-wide text-zinc-500 mb-2">In flight</div>
        <div class="text-2xl font-semibold tabular-nums">
          {{ inFlightCount }}
        </div>
        <div class="text-xs text-zinc-400 mt-1">active work_items</div>
      </div>
    </div>

    <!-- In-flight detail table -->
    <section
      v-if="inFlightCount > 0"
      class="rounded-md border border-zinc-800 bg-zinc-900/50 overflow-hidden"
    >
      <div class="px-4 py-3 border-b border-zinc-800">
        <h3 class="text-sm font-semibold">In-flight work items</h3>
      </div>
      <table class="w-full text-sm">
        <thead class="text-zinc-500 text-xs uppercase tracking-wide">
          <tr>
            <th class="text-left px-4 py-2 font-medium">Slug</th>
            <th class="text-left px-4 py-2 font-medium">Pipeline</th>
            <th class="text-left px-4 py-2 font-medium">Stage</th>
            <th class="text-left px-4 py-2 font-medium">Status</th>
            <th class="text-right px-4 py-2 font-medium">Tokens</th>
            <th class="text-right px-4 py-2 font-medium">Updated</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="w in data?.in_flight ?? []"
            :key="w.id"
            class="border-t border-zinc-800/50 hover:bg-zinc-900"
          >
            <td class="px-4 py-2 font-mono text-xs">{{ w.slug }}</td>
            <td class="px-4 py-2 text-zinc-300">{{ w.pipeline }}</td>
            <td class="px-4 py-2 text-zinc-300">{{ w.current_stage }}</td>
            <td class="px-4 py-2">
              <span
                class="inline-block px-2 py-0.5 rounded text-xs"
                :class="{
                  'bg-emerald-900/40 text-emerald-300': w.status === 'in_progress',
                  'bg-zinc-800 text-zinc-300': w.status === 'pending',
                  'bg-amber-900/40 text-amber-300': w.status === 'dispatched',
                }"
              >
                {{ w.status }}
              </span>
            </td>
            <td class="px-4 py-2 text-right tabular-nums text-zinc-400">
              {{ w.tokens_in.toLocaleString() }} / {{ w.tokens_out.toLocaleString() }}
            </td>
            <td class="px-4 py-2 text-right text-zinc-500 text-xs">
              {{ fmtRelative(w.updated_at) }}
            </td>
          </tr>
        </tbody>
      </table>
    </section>

    <!-- Recent errors -->
    <section
      v-if="errorCount > 0"
      class="rounded-md border border-red-900/40 bg-red-950/20 overflow-hidden"
    >
      <div class="px-4 py-3 border-b border-red-900/40 flex items-center gap-2">
        <span class="inline-block w-2 h-2 rounded-full bg-red-500"></span>
        <h3 class="text-sm font-semibold">Recent errors (24h)</h3>
        <span class="text-xs text-zinc-400">{{ errorCount }} item(s)</span>
      </div>
      <ul class="divide-y divide-red-900/30">
        <li
          v-for="e in data?.recent_errors ?? []"
          :key="e.id"
          class="px-4 py-3 text-sm"
        >
          <div class="flex items-baseline gap-3">
            <span class="font-mono text-xs text-zinc-500">#{{ e.id }}</span>
            <span class="text-zinc-300">{{ e.kind }}</span>
            <span class="text-zinc-500 text-xs">via {{ e.provider }}</span>
            <span class="ml-auto text-xs text-zinc-500">{{ fmtRelative(e.done_at) }}</span>
          </div>
          <div class="text-xs text-red-300 mt-1 font-mono whitespace-pre-wrap">
            {{ e.error }}
          </div>
        </li>
      </ul>
    </section>

    <div
      v-if="!loading && inFlightCount === 0 && errorCount === 0 && data"
      class="text-sm text-zinc-500"
    >
      Quiet substrate — no in-flight work, no recent errors.
    </div>
  </div>
</template>
