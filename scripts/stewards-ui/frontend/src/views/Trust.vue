<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { api, type TrustScoreRow, type TrustTransitionRow } from '@/api'

const scores = ref<TrustScoreRow[]>([])
const transitions = ref<TrustTransitionRow[]>([])
const loading = ref(true)
const error = ref('')

const adjustOpen = ref<TrustScoreRow | null>(null)
const adjustNewLevel = ref<'trainee' | 'journeyman' | 'master'>('journeyman')
const adjustJustification = ref('')
const adjustActor = ref('michael')
const adjusting = ref(false)
const adjustError = ref('')

async function load() {
  loading.value = true
  error.value = ''
  try {
    const [s, t] = await Promise.allSettled([
      api.trustScores(),
      api.trustTransitions({ limit: 50 }),
    ])
    if (s.status === 'fulfilled') scores.value = s.value.items
    if (t.status === 'fulfilled') transitions.value = t.value.items
  } catch (e) {
    error.value = String(e)
  } finally {
    loading.value = false
  }
}

function openAdjust(row: TrustScoreRow) {
  adjustOpen.value = row
  adjustNewLevel.value = row.trust_level === 'trainee' ? 'journeyman' : row.trust_level
  adjustJustification.value = ''
  adjustError.value = ''
}

async function submitAdjust() {
  if (!adjustOpen.value) return
  if (adjustJustification.value.trim().length < 10) {
    adjustError.value = 'Justification must be at least 10 characters (D-E2)'
    return
  }
  adjusting.value = true
  adjustError.value = ''
  try {
    await api.trustAdjust({
      agent_family: adjustOpen.value.agent_family,
      pipeline_family: adjustOpen.value.pipeline_family,
      model: adjustOpen.value.model,
      new_level: adjustNewLevel.value,
      actor: adjustActor.value || 'human',
      justification: adjustJustification.value,
    })
    adjustOpen.value = null
    await load()
  } catch (e) {
    adjustError.value = String(e)
  } finally {
    adjusting.value = false
  }
}

function fmtDate(s?: string) {
  if (!s) return ''
  return new Date(s).toLocaleString()
}

function levelTone(level: string): string {
  switch (level) {
    case 'master':     return 'bg-amber-900/30 text-amber-200 border-amber-700/50'
    case 'journeyman': return 'bg-blue-900/30 text-blue-300 border-blue-700/50'
    case 'trainee':    return 'bg-zinc-800 text-zinc-400 border-zinc-700'
    default:           return 'bg-zinc-800 text-zinc-400 border-zinc-700'
  }
}

const groupedByPipeline = computed(() => {
  const out: Record<string, TrustScoreRow[]> = {}
  for (const r of scores.value) {
    if (!out[r.pipeline_family]) out[r.pipeline_family] = []
    out[r.pipeline_family]!.push(r)
  }
  return out
})

onMounted(load)
</script>

<template>
  <div class="space-y-6 max-w-5xl">
    <header class="border-b border-zinc-800 pb-3">
      <h2 class="text-2xl font-semibold tracking-tight">Trust matrix</h2>
      <p class="text-sm text-zinc-400 mt-1">
        Per (agent_family × pipeline_family × model) trust state. Trainee surfaces every gate-advance for human ratification; journeyman + master proceed automatically. Demote on any human override (D-E3).
      </p>
    </header>

    <p v-if="loading" class="text-sm text-zinc-400">loading…</p>
    <p v-else-if="error" class="text-sm text-red-400">{{ error }}</p>

    <p v-else-if="scores.length === 0" class="text-sm text-zinc-500">
      No trust scores yet. Cells materialize when a work_item reaches verified maturity (success), is quarantined (failure), or a gate decision is overridden.
    </p>

    <div v-else class="space-y-6">
      <section v-for="(rows, pipeline) in groupedByPipeline" :key="pipeline">
        <h3 class="text-sm uppercase tracking-wide text-zinc-500 mb-2">{{ pipeline }}</h3>
        <table class="w-full text-sm border border-zinc-800 rounded">
          <thead>
            <tr class="text-zinc-500 border-b border-zinc-800 bg-zinc-900/50">
              <th class="text-left px-3 py-2 font-mono text-xs">agent</th>
              <th class="text-left px-3 py-2 font-mono text-xs">model</th>
              <th class="text-left px-3 py-2 font-mono text-xs">level</th>
              <th class="text-right px-3 py-2 font-mono text-xs">success</th>
              <th class="text-right px-3 py-2 font-mono text-xs">fail</th>
              <th class="text-right px-3 py-2 font-mono text-xs">override</th>
              <th class="text-left px-3 py-2 font-mono text-xs">last seen</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="r in rows"
              :key="r.agent_family + '|' + r.model"
              class="border-b border-zinc-800/50 hover:bg-zinc-800/30"
            >
              <td class="px-3 py-2 font-mono text-zinc-300">{{ r.agent_family }}</td>
              <td class="px-3 py-2 font-mono text-zinc-400">{{ r.model }}</td>
              <td class="px-3 py-2">
                <span
                  class="text-xs px-2 py-0.5 rounded border font-mono"
                  :class="levelTone(r.trust_level)"
                >{{ r.trust_level }}</span>
              </td>
              <td class="px-3 py-2 text-right tabular-nums text-emerald-400">{{ r.successful_completions }}</td>
              <td class="px-3 py-2 text-right tabular-nums text-red-400">{{ r.failed_completions }}</td>
              <td class="px-3 py-2 text-right tabular-nums text-amber-400">{{ r.human_overrides }}</td>
              <td class="px-3 py-2 text-xs text-zinc-500 tabular-nums">{{ fmtDate(r.last_completion_at || r.last_evaluated_at) }}</td>
              <td class="px-3 py-2 text-right">
                <button
                  class="text-xs px-2 py-1 rounded border border-zinc-700 hover:bg-zinc-800 text-zinc-300"
                  @click="openAdjust(r)"
                >adjust</button>
              </td>
            </tr>
          </tbody>
        </table>
      </section>

      <section
        v-if="transitions.length > 0"
        class="rounded-md border border-zinc-800 bg-zinc-900/50 overflow-hidden"
      >
        <div class="px-4 py-3 border-b border-zinc-800">
          <h3 class="text-sm font-semibold">Recent trust transitions ({{ transitions.length }})</h3>
        </div>
        <ul class="divide-y divide-zinc-800/50">
          <li v-for="t in transitions" :key="t.id" class="px-4 py-2 text-xs">
            <div class="flex items-baseline gap-3 flex-wrap">
              <span class="font-mono text-zinc-400">{{ t.agent_family }} / {{ t.pipeline_family }} / {{ t.model }}</span>
              <span class="font-mono">
                <span :class="levelTone(t.from_level).replace('bg-', 'text-').split(' ')[0]">{{ t.from_level }}</span>
                <span class="text-zinc-600 mx-1">→</span>
                <span :class="levelTone(t.to_level).replace('bg-', 'text-').split(' ')[0]">{{ t.to_level }}</span>
              </span>
              <span
                class="text-[10px] px-1.5 py-0.5 rounded border uppercase tracking-wide"
                :class="t.transition_kind === 'manual'
                  ? 'border-purple-700 text-purple-300 bg-purple-900/20'
                  : 'border-zinc-700 text-zinc-500'"
              >{{ t.transition_kind }}</span>
              <span class="text-zinc-500">by {{ t.actor }}</span>
              <span class="ml-auto text-zinc-500 tabular-nums">{{ fmtDate(t.at) }}</span>
            </div>
            <div v-if="t.justification" class="mt-1 text-zinc-400 italic">{{ t.justification }}</div>
          </li>
        </ul>
      </section>
    </div>

    <!-- Manual adjust modal -->
    <div
      v-if="adjustOpen"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/60"
      @click.self="adjustOpen = null"
    >
      <div class="bg-zinc-900 border border-zinc-700 rounded-lg p-5 max-w-lg w-full space-y-3">
        <h3 class="text-lg font-semibold">Manual trust adjustment</h3>
        <p class="text-xs text-zinc-500 font-mono">
          {{ adjustOpen.agent_family }} / {{ adjustOpen.pipeline_family }} / {{ adjustOpen.model }}
        </p>
        <p class="text-xs text-zinc-400">
          Current: <span class="font-mono">{{ adjustOpen.trust_level }}</span>
          ({{ adjustOpen.successful_completions }} ok / {{ adjustOpen.failed_completions }} fail / {{ adjustOpen.human_overrides }} override)
        </p>
        <div>
          <label class="block text-xs uppercase tracking-wide text-zinc-500 mb-1">New level</label>
          <select
            v-model="adjustNewLevel"
            class="w-full px-3 py-2 rounded border border-zinc-700 bg-zinc-950 text-sm focus:border-zinc-500 focus:outline-none"
          >
            <option value="trainee">trainee</option>
            <option value="journeyman">journeyman</option>
            <option value="master">master</option>
          </select>
        </div>
        <div>
          <label class="block text-xs uppercase tracking-wide text-zinc-500 mb-1">Actor</label>
          <input
            v-model="adjustActor"
            type="text"
            class="w-full px-3 py-2 rounded border border-zinc-700 bg-zinc-950 text-sm focus:border-zinc-500 focus:outline-none"
          />
        </div>
        <div>
          <label class="block text-xs uppercase tracking-wide text-zinc-500 mb-1">
            Justification <span class="text-red-400">*</span>
            <span class="text-zinc-600 normal-case">(D-E2: at least 10 characters)</span>
          </label>
          <textarea
            v-model="adjustJustification"
            rows="3"
            placeholder="why this adjustment? Visible to future-you in the audit log."
            class="w-full px-3 py-2 rounded border border-zinc-700 bg-zinc-950 text-sm focus:border-zinc-500 focus:outline-none resize-y"
          ></textarea>
        </div>
        <div class="flex items-center gap-3 pt-2">
          <button
            class="px-4 py-2 rounded bg-emerald-700 hover:bg-emerald-600 text-white text-sm font-medium disabled:opacity-50"
            :disabled="adjusting || adjustJustification.trim().length < 10"
            @click="submitAdjust"
          >{{ adjusting ? 'submitting…' : 'Adjust' }}</button>
          <button
            class="px-3 py-2 text-xs rounded border border-zinc-700 hover:bg-zinc-800 text-zinc-300"
            @click="adjustOpen = null"
          >Cancel</button>
          <span v-if="adjustError" class="text-xs text-red-400">{{ adjustError }}</span>
        </div>
      </div>
    </div>
  </div>
</template>
