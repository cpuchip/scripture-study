<script setup lang="ts">
import { ref, onMounted, watch, computed, onUnmounted } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import { api, type CouncilDetail } from '@/api'

const route = useRoute()
const council = ref<CouncilDetail | null>(null)
const error = ref('')
const loading = ref(false)
let pollTimer: ReturnType<typeof setInterval> | null = null

const editingResolution = ref('')
const destination = ref<'study' | 'decisions' | ''>('')
const resolvingBy = ref('michael')
const resolving = ref(false)
const resolveError = ref('')
const dissolveReason = ref('')

const idFromRoute = computed(() => String(route.params.id ?? ''))

async function load() {
  if (!idFromRoute.value) return
  try {
    council.value = await api.councilGet(idFromRoute.value)
    if (council.value?.resolution && !editingResolution.value) {
      editingResolution.value = council.value.resolution.text
    }
  } catch (e) {
    error.value = String(e)
  } finally {
    loading.value = false
  }
}

function startPoll() {
  stopPoll()
  pollTimer = setInterval(() => {
    if (council.value && ['deliberating', 'synthesizing'].includes(council.value.status)) {
      load()
    }
  }, 5000)
}

function stopPoll() {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

onMounted(() => {
  loading.value = true
  load().then(startPoll)
})

onUnmounted(stopPoll)
watch(idFromRoute, () => {
  loading.value = true
  council.value = null
  editingResolution.value = ''
  load()
})

function statusTone(status: string): string {
  switch (status) {
    case 'deliberating':    return 'bg-blue-900/30 text-blue-300 border-blue-700/50'
    case 'synthesizing':    return 'bg-purple-900/30 text-purple-300 border-purple-700/50'
    case 'awaiting_bishop': return 'bg-amber-900/30 text-amber-200 border-amber-700/50'
    case 'resolved':        return 'bg-emerald-900/30 text-emerald-300 border-emerald-700/50'
    case 'dissolved':       return 'bg-zinc-800 text-zinc-500 border-zinc-700'
    default:                return 'bg-zinc-800 text-zinc-400 border-zinc-700'
  }
}

function roleTone(role: string): string {
  switch (role) {
    case 'proposer':    return 'border-emerald-700/50 text-emerald-300 bg-emerald-900/20'
    case 'critic':      return 'border-amber-700/50 text-amber-300 bg-amber-900/20'
    case 'synthesizer': return 'border-purple-700/50 text-purple-300 bg-purple-900/20'
    default:            return 'border-zinc-700 text-zinc-400 bg-zinc-800'
  }
}

function fmtDate(s?: string) {
  if (!s) return ''
  return new Date(s).toLocaleString()
}

async function accept() {
  if (!council.value) return
  if (editingResolution.value.trim().length === 0) {
    resolveError.value = 'Resolution text required'
    return
  }
  resolving.value = true
  resolveError.value = ''
  try {
    await api.councilResolve({
      council_id: council.value.id,
      action: 'accept',
      resolution_text: editingResolution.value,
      destination: destination.value,
      resolved_by: resolvingBy.value || 'human',
    })
    await load()
  } catch (e) {
    resolveError.value = String(e)
  } finally {
    resolving.value = false
  }
}

async function requestRevision() {
  if (!council.value) return
  resolving.value = true
  resolveError.value = ''
  try {
    await api.councilResolve({
      council_id: council.value.id,
      action: 'request_revision',
      resolution_text: editingResolution.value,
    })
    await load()
  } catch (e) {
    resolveError.value = String(e)
  } finally {
    resolving.value = false
  }
}

async function dissolve() {
  if (!council.value) return
  if (!confirm('Dissolve this council? It will be preserved in the audit log.')) return
  resolving.value = true
  resolveError.value = ''
  try {
    await api.councilResolve({
      council_id: council.value.id,
      action: 'dissolve',
      dissolved_reason: dissolveReason.value || 'no reason given',
    })
    await load()
  } catch (e) {
    resolveError.value = String(e)
  } finally {
    resolving.value = false
  }
}

const memberStatus = computed(() => {
  if (!council.value) return ''
  const total = council.value.members.length
  const done = council.value.members.filter(m => m.completed_at).length
  return `${done}/${total} responded`
})
</script>

<template>
  <div class="space-y-6 max-w-4xl">
    <div>
      <RouterLink to="/councils" class="text-xs text-zinc-500 hover:text-zinc-300">
        ← all councils
      </RouterLink>
    </div>

    <p v-if="loading" class="text-sm text-zinc-400">loading…</p>
    <p v-else-if="error" class="text-sm text-red-400">{{ error }}</p>

    <template v-if="council">
      <header class="border-b border-zinc-800 pb-4 space-y-2">
        <div class="flex items-baseline justify-between gap-3">
          <h2 class="text-xl font-semibold tracking-tight">{{ council.binding_question }}</h2>
          <span
            class="text-xs px-2 py-0.5 rounded border font-mono"
            :class="statusTone(council.status)"
          >{{ council.status }}</span>
        </div>
        <div class="text-xs text-zinc-500 flex gap-3 font-mono flex-wrap">
          <span>intent: {{ council.intent_slug }}</span>
          <span>bishop: {{ council.bishop }}</span>
          <span>convened by {{ council.convened_by }} @ {{ fmtDate(council.convened_at) }}</span>
          <span v-if="council.resolved_at">resolved {{ fmtDate(council.resolved_at) }}</span>
          <span class="text-emerald-500">members: {{ memberStatus }}</span>
        </div>
        <p v-if="council.intent_purpose" class="text-sm text-zinc-300 italic">
          {{ council.intent_purpose }}
        </p>
      </header>

      <!-- Members section — the room -->
      <section class="space-y-4">
        <h3 class="text-sm uppercase tracking-wide text-zinc-500">Members</h3>
        <div
          v-for="m in council.members"
          :key="m.role + '|' + m.agent_family"
          class="rounded-md border border-zinc-800 bg-zinc-900/50 p-4"
        >
          <div class="flex items-baseline gap-3 mb-2">
            <span
              class="text-xs px-2 py-0.5 rounded border font-mono uppercase"
              :class="roleTone(m.role)"
            >{{ m.role }}</span>
            <span class="text-sm font-mono text-zinc-300">{{ m.agent_family }}</span>
            <span v-if="m.completed_at" class="ml-auto text-xs text-emerald-500 tabular-nums">
              ✓ {{ fmtDate(m.completed_at) }}
            </span>
            <span v-else class="ml-auto text-xs text-zinc-500 italic">
              dispatched, waiting…
            </span>
          </div>
          <p
            v-if="m.response"
            class="text-sm text-zinc-200 whitespace-pre-wrap leading-relaxed"
          >{{ m.response }}</p>
          <p v-else class="text-sm text-zinc-500 italic">(no response yet)</p>
        </div>
      </section>

      <!-- Synthesizer / draft resolution -->
      <section
        v-if="council.resolution"
        class="rounded-md border border-purple-900/40 bg-purple-950/10 p-4 space-y-3"
      >
        <div class="flex items-baseline justify-between">
          <h3 class="text-sm uppercase tracking-wide text-purple-400">
            Synthesizer's draft resolution
          </h3>
          <span v-if="council.resolution.resolved_by !== '__draft__'" class="text-xs text-emerald-500">
            ✓ accepted by {{ council.resolution.resolved_by }}
          </span>
          <span v-else class="text-xs text-amber-400">awaiting bishop</span>
        </div>
        <p class="text-sm text-zinc-200 whitespace-pre-wrap leading-relaxed">
          {{ council.resolution.text }}
        </p>
        <div v-if="council.resolution.promoted_to" class="text-xs text-zinc-500">
          Promoted to: <span class="font-mono text-zinc-300">{{ council.resolution.promoted_to }}</span>
        </div>
      </section>

      <!-- Bishop resolution form -->
      <section
        v-if="council.status === 'awaiting_bishop'"
        class="rounded-md border border-amber-900/40 bg-amber-950/20 p-4 space-y-3"
      >
        <h3 class="text-sm uppercase tracking-wide text-amber-300">
          Bishop's resolution
        </h3>
        <p class="text-xs text-zinc-400">
          Edit freely. Pick a destination if the resolution should land in study/ or .mind/decisions.md.
        </p>
        <textarea
          v-model="editingResolution"
          rows="6"
          class="w-full px-3 py-2 rounded border border-amber-800/50 bg-zinc-950 text-sm focus:border-amber-500 focus:outline-none resize-y"
        />
        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="block text-xs uppercase tracking-wide text-zinc-500 mb-1">Destination</label>
            <select
              v-model="destination"
              class="w-full px-3 py-2 rounded border border-zinc-700 bg-zinc-950 text-sm focus:border-zinc-500 focus:outline-none"
            >
              <option value="">— resolutions only (no file write)</option>
              <option value="study">study/&lt;council-id&gt;.md</option>
              <option value="decisions">.mind/decisions.md</option>
            </select>
          </div>
          <div>
            <label class="block text-xs uppercase tracking-wide text-zinc-500 mb-1">Accepted by</label>
            <input
              v-model="resolvingBy"
              type="text"
              class="w-full px-3 py-2 rounded border border-zinc-700 bg-zinc-950 text-sm focus:border-zinc-500 focus:outline-none"
            />
          </div>
        </div>
        <div class="flex items-center gap-2">
          <button
            class="px-4 py-2 rounded bg-emerald-700 hover:bg-emerald-600 text-white text-sm font-medium disabled:opacity-50"
            :disabled="resolving || editingResolution.trim().length === 0"
            @click="accept"
          >{{ resolving ? 'submitting…' : 'Accept' }}</button>
          <button
            class="px-3 py-2 text-xs rounded border border-purple-700 hover:bg-purple-900/30 text-purple-300 disabled:opacity-50"
            :disabled="resolving"
            @click="requestRevision"
          >Request revision (re-fire synth)</button>
          <input
            v-model="dissolveReason"
            type="text"
            placeholder="reason for dissolve"
            class="flex-1 px-2 py-1 text-xs rounded border border-zinc-700 bg-zinc-950 focus:border-zinc-500 focus:outline-none"
          />
          <button
            class="px-3 py-2 text-xs rounded border border-red-800 hover:bg-red-900/30 text-red-400 disabled:opacity-50"
            :disabled="resolving"
            @click="dissolve"
          >Dissolve</button>
        </div>
        <p v-if="resolveError" class="text-xs text-red-400">{{ resolveError }}</p>
      </section>

      <p
        v-if="council.status === 'deliberating' || council.status === 'synthesizing'"
        class="text-xs text-zinc-500 italic text-center"
      >
        Auto-refreshing every 5s while the council is in flight…
      </p>
    </template>
  </div>
</template>
