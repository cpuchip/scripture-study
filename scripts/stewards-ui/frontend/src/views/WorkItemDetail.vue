<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import { api, type WorkItemDetail, type CostEventsResp, type StewardActionsResp } from '@/api'

const route = useRoute()
const wi = ref<WorkItemDetail | null>(null)
const cost = ref<CostEventsResp | null>(null)
const actions = ref<StewardActionsResp | null>(null)
const error = ref<string>('')
const loading = ref(false)

async function load(idOrSlug: string) {
  loading.value = true
  error.value = ''
  wi.value = null
  cost.value = null
  actions.value = null
  try {
    const detail = await api.workItemGet(idOrSlug)
    wi.value = detail
    // Fire cost + actions in parallel; failures don't block the view
    const [c, a] = await Promise.allSettled([
      api.workItemCost(detail.id),
      api.workItemActions(detail.id),
    ])
    if (c.status === 'fulfilled') cost.value = c.value
    if (a.status === 'fulfilled') actions.value = a.value
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
function fmtMicro(micro?: number | null) {
  if (micro == null) return '—'
  // micro-dollars → readable USD with 4 decimals
  return '$' + (micro / 1_000_000).toFixed(4)
}
function capPercent(used: number, cap?: number | null): string {
  if (!cap || cap <= 0) return ''
  return ((used * 100) / cap).toFixed(1) + '%'
}
function escalationBadge(state: string): { label: string; cls: string } {
  switch (state) {
    case 'queued':
      return { label: 'queued for opus boost', cls: 'bg-amber-900/30 text-amber-300 border-amber-700/50' }
    case 'in_progress':
      return { label: 'escalation in progress', cls: 'bg-blue-900/30 text-blue-300 border-blue-700/50' }
    case 'resolved':
      return { label: 'escalation resolved', cls: 'bg-emerald-900/30 text-emerald-300 border-emerald-700/50' }
    case 'failed':
      return { label: 'escalation failed', cls: 'bg-red-900/30 text-red-300 border-red-700/50' }
    default:
      return { label: 'normal', cls: 'bg-zinc-800 text-zinc-400 border-zinc-700' }
  }
}
function actionTone(action: string): string {
  if (action === 'retry_dispatched') return 'text-blue-300'
  if (action === 'queue_for_opus') return 'text-amber-300'
  if (action === 'quarantine' || action === 'escalation_failed') return 'text-red-300'
  if (action === 'escalation_resolved') return 'text-emerald-300'
  if (action === 'tick_error' || action === 'dispatch_error') return 'text-orange-300'
  if (action === 'defer_breaker_open') return 'text-yellow-300'
  return 'text-zinc-300'
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

      <!-- Phase 4j: Steward status panel -->
      <section
        v-if="wi.escalation_state !== 'normal' || wi.failure_count > 0 || wi.quarantined_at || wi.model_override"
        class="rounded-md border border-zinc-800 bg-zinc-900/50 p-4"
      >
        <div class="flex items-baseline justify-between mb-3">
          <div class="text-xs uppercase tracking-wide text-zinc-500">Steward status</div>
          <span
            class="text-xs px-2 py-0.5 rounded border font-mono"
            :class="escalationBadge(wi.escalation_state).cls"
          >{{ escalationBadge(wi.escalation_state).label }}</span>
        </div>
        <div class="grid grid-cols-2 md:grid-cols-3 gap-3 text-xs">
          <div>
            <div class="text-zinc-500">Failure count</div>
            <div class="font-mono text-zinc-200">{{ wi.failure_count }} / 3</div>
          </div>
          <div v-if="wi.last_failure_diagnosis">
            <div class="text-zinc-500">Last diagnosis</div>
            <div class="font-mono text-zinc-200">{{ wi.last_failure_diagnosis }}</div>
          </div>
          <div v-if="wi.escalation_attempts > 0">
            <div class="text-zinc-500">Escalation attempts</div>
            <div class="font-mono text-zinc-200">{{ wi.escalation_attempts }}</div>
          </div>
          <div v-if="wi.model_override" class="col-span-2">
            <div class="text-zinc-500">Override</div>
            <div class="font-mono text-zinc-200">
              {{ wi.model_override }}
              <span v-if="wi.provider_override" class="text-zinc-500">via {{ wi.provider_override }}</span>
            </div>
          </div>
          <div v-if="wi.escalation_claimed_by">
            <div class="text-zinc-500">Claimed by</div>
            <div class="font-mono text-zinc-200">{{ wi.escalation_claimed_by }}</div>
          </div>
          <div v-if="wi.quarantined_at" class="col-span-2 md:col-span-3">
            <div class="text-red-400">Quarantined</div>
            <div class="font-mono text-red-300">
              {{ wi.quarantine_reason }} at {{ fmtDate(wi.quarantined_at) }}
            </div>
          </div>
        </div>
        <details v-if="wi.last_failure_reason" class="mt-3">
          <summary class="text-xs cursor-pointer text-zinc-500 hover:text-zinc-300">
            Last failure reason
          </summary>
          <pre class="text-xs font-mono text-zinc-400 whitespace-pre-wrap mt-2">{{ wi.last_failure_reason }}</pre>
        </details>
      </section>

      <!-- Phase 4j: Cost panel -->
      <section
        v-if="cost"
        class="rounded-md border border-zinc-800 bg-zinc-900/50 p-4"
      >
        <div class="flex items-baseline justify-between mb-3">
          <div class="text-xs uppercase tracking-wide text-zinc-500">Cost</div>
          <span
            v-if="cost.cost_capped_at"
            class="text-xs px-2 py-0.5 rounded border font-mono bg-red-900/30 text-red-300 border-red-700/50"
          >cap exceeded</span>
        </div>
        <div class="flex items-baseline gap-6 text-sm">
          <div>
            <div class="text-xs text-zinc-500">Cumulative</div>
            <div class="text-2xl font-mono text-zinc-100">{{ fmtMicro(cost.work_item_cost_micro) }}</div>
          </div>
          <div v-if="cost.cost_cap_micro">
            <div class="text-xs text-zinc-500">Cap</div>
            <div class="text-lg font-mono text-zinc-300">
              {{ fmtMicro(cost.cost_cap_micro) }}
              <span class="text-xs text-zinc-500">({{ capPercent(cost.work_item_cost_micro, cost.cost_cap_micro) }})</span>
            </div>
          </div>
          <div>
            <div class="text-xs text-zinc-500">Events</div>
            <div class="text-lg font-mono text-zinc-300">{{ cost.total_events }}</div>
          </div>
        </div>
        <details v-if="cost.items.length > 0" class="mt-3">
          <summary class="text-xs cursor-pointer text-zinc-500 hover:text-zinc-300">
            Per-attempt breakdown ({{ cost.total_events }} events)
          </summary>
          <table class="w-full text-xs font-mono mt-2">
            <thead>
              <tr class="text-zinc-500 border-b border-zinc-800">
                <th class="text-left py-1 pr-3">#</th>
                <th class="text-left py-1 pr-3">model</th>
                <th class="text-right py-1 pr-3">in</th>
                <th class="text-right py-1 pr-3">out</th>
                <th class="text-right py-1 pr-3">cache w/r</th>
                <th class="text-right py-1 pr-3">cost</th>
                <th class="text-left py-1">at</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="ev in cost.items"
                :key="ev.id"
                class="text-zinc-300 hover:bg-zinc-800/30"
              >
                <td class="py-1 pr-3 text-zinc-500">{{ ev.attempt_seq }}</td>
                <td class="py-1 pr-3">{{ ev.model }}</td>
                <td class="py-1 pr-3 text-right tabular-nums">{{ ev.input_tokens.toLocaleString() }}</td>
                <td class="py-1 pr-3 text-right tabular-nums">{{ ev.output_tokens.toLocaleString() }}</td>
                <td class="py-1 pr-3 text-right tabular-nums text-zinc-500">
                  {{ ev.cache_write_tokens || '·' }} / {{ ev.cache_read_tokens || '·' }}
                </td>
                <td class="py-1 pr-3 text-right tabular-nums">{{ fmtMicro(ev.micro_dollars) }}</td>
                <td class="py-1 text-zinc-500">{{ fmtDate(ev.at) }}</td>
              </tr>
            </tbody>
          </table>
        </details>
      </section>

      <!-- Phase 4j: Steward actions audit -->
      <section
        v-if="actions && actions.count > 0"
        class="rounded-md border border-zinc-800 bg-zinc-900/50 overflow-hidden"
      >
        <div class="px-4 py-3 border-b border-zinc-800 flex items-baseline justify-between">
          <h3 class="text-sm font-semibold">Steward actions ({{ actions.count }})</h3>
          <span class="text-xs text-zinc-500">most recent first</span>
        </div>
        <ul class="divide-y divide-zinc-800/50">
          <li v-for="a in actions.items" :key="a.id" class="px-4 py-2 text-xs">
            <div class="flex items-baseline gap-3 flex-wrap">
              <span
                class="font-mono uppercase tracking-wide"
                :class="actionTone(a.action)"
              >{{ a.action }}</span>
              <span v-if="a.diagnosis" class="text-zinc-500 font-mono">[{{ a.diagnosis }}]</span>
              <span v-if="a.model_used" class="text-zinc-500 font-mono">model: {{ a.model_used }}</span>
              <span class="ml-auto text-zinc-500 tabular-nums">{{ fmtDate(a.at) }}</span>
            </div>
            <div class="mt-1 text-zinc-300">{{ a.observation }}</div>
            <details v-if="a.details && Object.keys(a.details as object).length > 0" class="mt-1">
              <summary class="cursor-pointer text-zinc-600 hover:text-zinc-400">details</summary>
              <pre class="mt-1 font-mono text-xs text-zinc-400 whitespace-pre-wrap">{{ fmtJson(a.details) }}</pre>
            </details>
          </li>
        </ul>
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
