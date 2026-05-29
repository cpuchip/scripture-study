<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter, RouterLink } from 'vue-router'
import { api, type WorkItemsListResp, type WorkItemRow } from '@/api'

const route = useRoute()
const router = useRouter()
const data = ref<WorkItemsListResp | null>(null)
const error = ref<string>('')
const loading = ref(false)

const pipeline = ref(String(route.query.pipeline ?? ''))
// J.1: default to 'open' so the page lands on what's in flight, not the
// historical pile. User can switch to 'done' or 'all statuses' via the
// status-group dropdown.
const status = ref(String(route.query.status ?? 'open'))
const origin = ref(String(route.query.origin ?? ''))

// J.1: per-parent expand/collapse state for the tree. Default expanded;
// click the chevron to collapse. Keyed by parent.id.
const collapsed = ref<Record<string, boolean>>({})
function toggle(parentId: string) {
  collapsed.value[parentId] = !collapsed.value[parentId]
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    data.value = await api.workItemsList({
      pipeline: pipeline.value || undefined,
      status: status.value || undefined,
      origin: origin.value || undefined,
      limit: 200,
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

function originBadgeClass(o?: string): string {
  switch (o) {
    case 'agent_planning': return 'bg-purple-900/40 text-purple-300 border border-purple-800/60'
    case 'agent_proposal': return 'bg-emerald-900/40 text-emerald-300 border border-emerald-800/60'
    case 'scheduled':      return 'bg-cyan-900/40 text-cyan-300'
    case 'watchman':       return 'bg-teal-900/40 text-teal-300'
    case 'steward':        return 'bg-amber-900/40 text-amber-300'
    case 'council':        return 'bg-indigo-900/40 text-indigo-300'
    case 'human':
    default:               return ''
  }
}

// J.1: parent-link badge. If the parent is in the current result set,
// show its slug; otherwise fall back to a short uuid for traceability.
function parentLabel(parentId: string): string {
  if (!parentId) return ''
  const parent = data.value?.items.find(i => i.id === parentId)
  return parent ? parent.slug : parentId.slice(0, 8)
}

// J.1: tree organization. We group by parent_work_item_id but only when
// the parent is in the same result set; otherwise the orphaned child
// renders top-level (and shows its parent-link badge for traceability).
type TreeRow = { item: WorkItemRow; depth: number }
const treeRows = computed<TreeRow[]>(() => {
  if (!data.value) return []
  const items = data.value.items
  const byParent = new Map<string, WorkItemRow[]>()
  const topLevel: WorkItemRow[] = []
  const idsInView = new Set(items.map(i => i.id))

  for (const it of items) {
    const pid = it.parent_work_item_id || ''
    if (pid && idsInView.has(pid)) {
      const arr = byParent.get(pid) ?? []
      arr.push(it)
      byParent.set(pid, arr)
    } else {
      topLevel.push(it)
    }
  }

  const out: TreeRow[] = []
  function walk(row: WorkItemRow, depth: number) {
    out.push({ item: row, depth })
    if (collapsed.value[row.id]) return
    const kids = byParent.get(row.id) ?? []
    for (const k of kids) walk(k, depth + 1)
  }
  for (const t of topLevel) walk(t, 0)
  return out
})

function hasChildrenInView(id: string): boolean {
  if (!data.value) return false
  return data.value.items.some(i => i.parent_work_item_id === id)
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

    <form class="flex gap-2 flex-wrap" @submit.prevent="submit">
      <input
        v-model="pipeline"
        type="text"
        placeholder="pipeline filter (e.g. study-write)…"
        class="flex-1 min-w-[12rem] px-3 py-2 rounded border border-zinc-700 bg-zinc-900 text-sm focus:border-zinc-500 focus:outline-none"
      />
      <!-- J.1: status-group dropdown. Two virtual groups at the top
           (open = pending/dispatched/in_progress; done =
           completed/cancelled/failed) cover the common queries;
           individual statuses follow underneath. -->
      <select
        v-model="status"
        class="px-3 py-2 rounded border border-zinc-700 bg-zinc-900 text-sm"
      >
        <optgroup label="Groups">
          <option value="open">open (not done)</option>
          <option value="done">done</option>
          <option value="">all statuses</option>
        </optgroup>
        <optgroup label="Status">
          <option value="pending">pending</option>
          <option value="dispatched">dispatched</option>
          <option value="in_progress">in_progress</option>
          <option value="completed">completed</option>
          <option value="failed">failed</option>
          <option value="cancelled">cancelled</option>
        </optgroup>
      </select>
      <select
        v-model="origin"
        class="px-3 py-2 rounded border border-zinc-700 bg-zinc-900 text-sm"
      >
        <option value="">all origins</option>
        <option value="human">human</option>
        <option value="agent_planning">agent_planning</option>
        <option value="agent_proposal">agent_proposal</option>
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
            v-for="{ item: w, depth } in treeRows"
            :key="w.id"
            class="border-t border-zinc-800/50 hover:bg-zinc-900"
          >
            <td class="px-4 py-2">
              <!-- J.1: tree indent. Each level adds 16px; the leaf rows
                   carry a unicode tree-branch marker. -->
              <span
                v-if="depth > 0"
                class="inline-block text-zinc-600 mr-1 select-none"
                :style="{ paddingLeft: ((depth - 1) * 16) + 'px' }"
              >└─</span>
              <!-- Parents with children in view get an expand/collapse
                   chevron. Leaf rows at depth 0 get a spacer so the slug
                   column stays aligned across rows. -->
              <button
                v-if="hasChildrenInView(w.id)"
                @click="toggle(w.id)"
                type="button"
                class="inline-block w-4 text-zinc-500 hover:text-zinc-200 select-none"
                :title="collapsed[w.id] ? 'expand' : 'collapse'"
              >{{ collapsed[w.id] ? '▶' : '▼' }}</button>
              <span v-else-if="depth === 0" class="inline-block w-4"></span>

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
                {{
                  w.origin === 'agent_planning' ? '✨ proposed'
                  : w.origin === 'agent_proposal' ? '🤖 agent write-back'
                  : w.origin
                }}
              </span>
              <span
                v-if="w.project_association"
                class="ml-2 inline-block px-1.5 py-0.5 rounded text-xs bg-zinc-800/50 text-zinc-400 font-mono"
                :title="`project: ${w.project_association}`"
              >
                {{ w.project_association }}
              </span>
              <!-- J.1: parent-link badge. Always visible when a parent
                   exists, so the relationship survives filters that hide
                   the parent from the result set. -->
              <RouterLink
                v-if="w.parent_work_item_id"
                :to="`/work-items/${w.parent_work_item_id}`"
                class="ml-2 inline-block px-1.5 py-0.5 rounded text-xs bg-sky-900/40 text-sky-300 border border-sky-800/60 font-mono hover:bg-sky-900/60"
                :title="`child of ${parentLabel(w.parent_work_item_id)}`"
              >
                ↪ {{ parentLabel(w.parent_work_item_id) }}
              </RouterLink>
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
              <!-- J.12: budget/cap failures get a distinct amber badge -->
              <span
                v-if="w.error_category === 'provider_budget' || w.error_category === 'spend_cap_reached'"
                class="ml-1 inline-block px-2 py-0.5 rounded text-xs bg-amber-900/40 text-amber-300"
                :title="w.error_category === 'spend_cap_reached' ? 'Refused: provider spend cap reached' : 'Provider budget / quota exhausted — refill needed'"
              >💸 budget</span>
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
