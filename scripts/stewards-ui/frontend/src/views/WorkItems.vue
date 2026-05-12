<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRoute, useRouter, RouterLink } from 'vue-router'
import { api, type WorkItemsListResp } from '@/api'

const route = useRoute()
const router = useRouter()
const data = ref<WorkItemsListResp | null>(null)
const error = ref<string>('')
const loading = ref(false)

const pipeline = ref(String(route.query.pipeline ?? ''))
const status = ref(String(route.query.status ?? ''))
const origin = ref(String(route.query.origin ?? ''))

async function load() {
  loading.value = true
  error.value = ''
  try {
    data.value = await api.workItemsList({
      pipeline: pipeline.value || undefined,
      status: status.value || undefined,
      origin: origin.value || undefined,
      limit: 100,
    })
  } catch (e) {
    error.value = String(e)
  } finally {
    loading.value = false
  }
}
function submit() {
  const q: Record<string, string> = {}
  if (pipeline.value) q.pipeline = pipeline.value
  if (status.value) q.status = status.value
  if (origin.value) q.origin = origin.value
  router.replace({ path: '/work-items', query: q })
  load()
}

// H.3: visual style for the origin badge. agent_planning is the
// substrate-proposed work that awaits human ratification — make it
// visibly distinct from human-originated work.
function originBadgeClass(o?: string): string {
  switch (o) {
    case 'agent_planning': return 'bg-purple-900/40 text-purple-300 border border-purple-800/60'
    case 'scheduled':      return 'bg-cyan-900/40 text-cyan-300'
    case 'watchman':       return 'bg-teal-900/40 text-teal-300'
    case 'steward':        return 'bg-amber-900/40 text-amber-300'
    case 'council':        return 'bg-indigo-900/40 text-indigo-300'
    case 'human':
    default:               return ''   // no badge for default human-originated
  }
}

onMounted(load)
watch(() => route.query, load)

function fmtRelative(s?: string) {
  if (!s) return ''
  const d = new Date(s)
  if (isNaN(d.getTime())) return s
  const sec = Math.floor((Date.now() - d.getTime()) / 1000)
  if (sec < 60) return `${sec}s ago`
  if (sec < 3600) return `${Math.floor(sec / 60)}m ago`
  if (sec < 86400) return `${Math.floor(sec / 3600)}h ago`
  return `${Math.floor(sec / 86400)}d ago`
}
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-baseline justify-between">
      <h2 class="text-2xl font-semibold tracking-tight">Work items</h2>
      <span v-if="data" class="text-xs text-zinc-500">
        {{ data.total.toLocaleString() }} total
      </span>
    </div>

    <form class="flex gap-2" @submit.prevent="submit">
      <input
        v-model="pipeline"
        type="text"
        placeholder="pipeline filter (e.g. study-write)…"
        class="flex-1 px-3 py-2 rounded border border-zinc-700 bg-zinc-900 text-sm focus:border-zinc-500 focus:outline-none"
      />
      <select
        v-model="status"
        class="px-3 py-2 rounded border border-zinc-700 bg-zinc-900 text-sm"
      >
        <option value="">all status</option>
        <option value="pending">pending</option>
        <option value="dispatched">dispatched</option>
        <option value="in_progress">in_progress</option>
        <option value="completed">completed</option>
        <option value="failed">failed</option>
        <option value="cancelled">cancelled</option>
      </select>
      <select
        v-model="origin"
        class="px-3 py-2 rounded border border-zinc-700 bg-zinc-900 text-sm"
      >
        <option value="">all origins</option>
        <option value="human">human</option>
        <option value="agent_planning">agent_planning</option>
        <option value="scheduled">scheduled</option>
        <option value="watchman">watchman</option>
        <option value="steward">steward</option>
        <option value="council">council</option>
      </select>
      <button type="submit" class="px-4 py-2 rounded border border-zinc-700 hover:bg-zinc-800 text-sm">
        go
      </button>
    </form>

    <p v-if="loading" class="text-sm text-zinc-400">loading…</p>
    <p v-else-if="error" class="text-sm text-red-400">{{ error }}</p>

    <div
      v-else-if="data"
      class="rounded-md border border-zinc-800 bg-zinc-900/50 overflow-hidden"
    >
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
            v-for="w in data.items"
            :key="w.id"
            class="border-t border-zinc-800/50 hover:bg-zinc-900"
          >
            <td class="px-4 py-2">
              <RouterLink
                :to="`/work-items/${w.id}`"
                class="text-zinc-100 hover:text-white font-mono text-xs"
              >
                {{ w.slug }}
              </RouterLink>
              <span
                v-if="w.origin && w.origin !== 'human'"
                class="ml-2 inline-block px-1.5 py-0.5 rounded text-xs"
                :class="originBadgeClass(w.origin)"
                :title="`origin: ${w.origin}`"
              >
                {{ w.origin === 'agent_planning' ? '✨ proposed' : w.origin }}
              </span>
              <span
                v-if="w.project_association"
                class="ml-2 inline-block px-1.5 py-0.5 rounded text-xs bg-zinc-800/50 text-zinc-400 font-mono"
                :title="`project: ${w.project_association}`"
              >
                {{ w.project_association }}
              </span>
            </td>
            <td class="px-4 py-2 text-zinc-300">{{ w.pipeline }}</td>
            <td class="px-4 py-2 text-zinc-300">{{ w.current_stage }}</td>
            <td class="px-4 py-2">
              <span
                class="inline-block px-2 py-0.5 rounded text-xs"
                :class="{
                  'bg-emerald-900/40 text-emerald-300': w.status === 'in_progress',
                  'bg-zinc-800 text-zinc-300': w.status === 'pending',
                  'bg-amber-900/40 text-amber-300': w.status === 'dispatched',
                  'bg-blue-900/40 text-blue-300': w.status === 'completed',
                  'bg-red-900/40 text-red-300': w.status === 'failed',
                  'bg-zinc-800/40 text-zinc-500': w.status === 'cancelled',
                }"
              >{{ w.status }}</span>
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
    </div>
  </div>
</template>
