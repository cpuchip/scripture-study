<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import { useRouter } from 'vue-router'
import { api, type CouncilRow, type CouncilSuggestion, type IntentRow } from '@/api'

const router = useRouter()

const councils = ref<CouncilRow[]>([])
const suggestions = ref<CouncilSuggestion[]>([])
const intents = ref<IntentRow[]>([])
const loading = ref(true)
const error = ref('')

const showConvene = ref(false)
const newIntentId = ref('')
const newBinding = ref('')
const newBishop = ref('human:michael')
const newMembers = ref([
  { agent_family: 'plan', role: 'proposer' as const, model: 'kimi-k2.6' },
  { agent_family: 'plan', role: 'critic' as const, model: 'qwen3.6-plus' },
])
const convening = ref(false)
const conveneError = ref('')

async function load() {
  loading.value = true
  error.value = ''
  try {
    const [c, s, i] = await Promise.allSettled([
      api.councilsList(50),
      api.councilSuggestions(),
      api.intentsList(),
    ])
    if (c.status === 'fulfilled') councils.value = c.value.items
    if (s.status === 'fulfilled') suggestions.value = s.value.items
    if (i.status === 'fulfilled') intents.value = i.value.items
    if (intents.value.length > 0 && !newIntentId.value) {
      const def = intents.value.find(x => x.slug === 'scripture-study')
      const fallback = intents.value[0]
      if (def) {
        newIntentId.value = def.id
      } else if (fallback) {
        newIntentId.value = fallback.id
      }
    }
  } catch (e) {
    error.value = String(e)
  } finally {
    loading.value = false
  }
}

const activeCouncil = computed(() =>
  councils.value.find(c => ['deliberating', 'synthesizing', 'awaiting_bishop'].includes(c.status))
)

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

function fmtDate(s?: string) {
  if (!s) return ''
  return new Date(s).toLocaleString()
}

function addMember() {
  if (newMembers.value.length >= 5) return
  newMembers.value.push({ agent_family: 'plan', role: 'proposer', model: 'kimi-k2.6' })
}

function removeMember(i: number) {
  if (newMembers.value.length <= 2) return
  newMembers.value.splice(i, 1)
}

async function convene() {
  if (!newIntentId.value || !newBinding.value || !newBishop.value) {
    conveneError.value = 'intent, binding question, and bishop required'
    return
  }
  convening.value = true
  conveneError.value = ''
  try {
    const r = await api.councilConvene({
      intent_id: newIntentId.value,
      binding_question: newBinding.value,
      members: newMembers.value,
      bishop: newBishop.value,
      convened_by: 'human',
    })
    showConvene.value = false
    router.push(`/councils/${r.id}`)
  } catch (e) {
    conveneError.value = String(e)
  } finally {
    convening.value = false
  }
}

function conveneFromSuggestion(s: CouncilSuggestion) {
  newBinding.value = `Should the ${s.pipeline_family} ${s.current_stage} stage be revised? (${s.lesson_count} ratified lessons accumulated)`
  showConvene.value = true
}

onMounted(load)
</script>

<template>
  <div class="space-y-6 max-w-5xl">
    <header class="border-b border-zinc-800 pb-3">
      <div class="flex items-baseline justify-between">
        <div>
          <h2 class="text-2xl font-semibold tracking-tight">Councils</h2>
          <p class="text-sm text-zinc-400 mt-1">
            Multi-agent deliberation on a single binding question. Bishop facilitates,
            doesn't orchestrate. One council at a time per D-F1.
          </p>
        </div>
        <button
          class="px-4 py-2 rounded bg-emerald-700 hover:bg-emerald-600 text-white text-sm font-medium disabled:opacity-50 disabled:cursor-not-allowed"
          :disabled="!!activeCouncil"
          @click="showConvene = true"
        >
          {{ activeCouncil ? 'one council active' : 'Convene…' }}
        </button>
      </div>
    </header>

    <!-- Watchman suggestions banner -->
    <section
      v-if="suggestions.length > 0"
      class="rounded-md border border-amber-900/40 bg-amber-950/20 p-4"
    >
      <div class="flex items-baseline justify-between mb-2">
        <h3 class="text-sm font-semibold text-amber-300">Watchman suggests convening:</h3>
        <span class="text-xs text-zinc-500">{{ suggestions.length }} cluster{{ suggestions.length === 1 ? '' : 's' }}</span>
      </div>
      <ul class="space-y-2">
        <li v-for="s in suggestions" :key="s.pipeline_family + '|' + s.current_stage" class="text-sm">
          <div class="flex items-baseline gap-3">
            <span class="font-mono text-amber-200">{{ s.pipeline_family }} / {{ s.current_stage }}</span>
            <span class="text-xs text-zinc-500">{{ s.lesson_count }} ratified lessons</span>
            <button
              class="ml-auto text-xs px-2 py-1 rounded border border-amber-700 hover:bg-amber-900/30 text-amber-200"
              @click="conveneFromSuggestion(s)"
            >Convene from this →</button>
          </div>
          <pre class="mt-1 font-mono text-xs text-zinc-400 whitespace-pre-wrap">{{ s.sample_content }}</pre>
        </li>
      </ul>
    </section>

    <p v-if="loading" class="text-sm text-zinc-400">loading…</p>
    <p v-else-if="error" class="text-sm text-red-400">{{ error }}</p>

    <p v-else-if="councils.length === 0" class="text-sm text-zinc-500">
      No councils yet. The substrate's culmination — convene the first when a question is too big for one agent to decide alone.
    </p>

    <ul v-else class="space-y-3">
      <li
        v-for="c in councils"
        :key="c.id"
        class="rounded-md border border-zinc-800 bg-zinc-900/50 p-4"
      >
        <div class="flex items-baseline justify-between gap-3 mb-2">
          <RouterLink
            :to="`/councils/${c.id}`"
            class="text-base font-semibold hover:text-emerald-300"
          >{{ c.binding_question }}</RouterLink>
          <span
            class="text-xs px-2 py-0.5 rounded border font-mono"
            :class="statusTone(c.status)"
          >{{ c.status }}</span>
        </div>
        <div class="flex items-baseline gap-3 text-xs text-zinc-500 font-mono">
          <span v-if="c.intent_slug">intent: {{ c.intent_slug }}</span>
          <span>bishop: {{ c.bishop }}</span>
          <span>by: {{ c.convened_by }}</span>
          <span>convened {{ fmtDate(c.convened_at) }}</span>
          <span v-if="c.resolved_at">resolved {{ fmtDate(c.resolved_at) }}</span>
        </div>
        <p v-if="c.dissolved_reason" class="text-xs text-zinc-400 mt-1 italic">
          dissolved: {{ c.dissolved_reason }}
        </p>
      </li>
    </ul>

    <!-- Convene modal -->
    <div
      v-if="showConvene"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/60"
      @click.self="showConvene = false"
    >
      <div class="bg-zinc-900 border border-zinc-700 rounded-lg p-5 max-w-2xl w-full space-y-3 max-h-[90vh] overflow-y-auto">
        <h3 class="text-lg font-semibold">Convene a council</h3>
        <p class="text-xs text-zinc-500">2–5 members. Synthesizer is auto-fired when proposer + critic responses arrive.</p>

        <div>
          <label class="block text-xs uppercase tracking-wide text-zinc-500 mb-1">Intent <span class="text-red-400">*</span></label>
          <select
            v-model="newIntentId"
            class="w-full px-3 py-2 rounded border border-zinc-700 bg-zinc-950 text-sm focus:border-zinc-500 focus:outline-none"
          >
            <option v-for="i in intents" :key="i.id" :value="i.id">
              {{ i.slug }} — {{ i.purpose.slice(0, 60) }}{{ i.purpose.length > 60 ? '…' : '' }}
            </option>
          </select>
        </div>

        <div>
          <label class="block text-xs uppercase tracking-wide text-zinc-500 mb-1">Binding question <span class="text-red-400">*</span></label>
          <textarea
            v-model="newBinding"
            rows="2"
            placeholder="The single question this council convenes on."
            class="w-full px-3 py-2 rounded border border-zinc-700 bg-zinc-950 text-sm focus:border-zinc-500 focus:outline-none resize-y"
          />
        </div>

        <div>
          <label class="block text-xs uppercase tracking-wide text-zinc-500 mb-1">
            Bishop <span class="text-red-400">*</span>
            <span class="text-zinc-600 normal-case">(human:&lt;name&gt; or agent:&lt;family&gt;:&lt;pipeline&gt;:master)</span>
          </label>
          <input
            v-model="newBishop"
            type="text"
            class="w-full px-3 py-2 rounded border border-zinc-700 bg-zinc-950 text-sm font-mono focus:border-zinc-500 focus:outline-none"
          />
        </div>

        <div>
          <div class="flex items-baseline justify-between mb-1">
            <label class="text-xs uppercase tracking-wide text-zinc-500">Members ({{ newMembers.length }})</label>
            <button
              class="text-xs px-2 py-1 rounded border border-zinc-700 hover:bg-zinc-800 text-zinc-300"
              :disabled="newMembers.length >= 5"
              @click="addMember"
            >+ add</button>
          </div>
          <ul class="space-y-2">
            <li v-for="(m, i) in newMembers" :key="i" class="flex gap-2">
              <input
                v-model="m.agent_family"
                placeholder="agent_family"
                class="w-1/3 px-2 py-1 text-sm rounded border border-zinc-700 bg-zinc-950 font-mono focus:border-zinc-500 focus:outline-none"
              />
              <select
                v-model="m.role"
                class="px-2 py-1 text-sm rounded border border-zinc-700 bg-zinc-950 font-mono focus:border-zinc-500 focus:outline-none"
              >
                <option value="proposer">proposer</option>
                <option value="critic">critic</option>
                <option value="synthesizer">synthesizer</option>
              </select>
              <input
                v-model="m.model"
                placeholder="model"
                class="flex-1 px-2 py-1 text-sm rounded border border-zinc-700 bg-zinc-950 font-mono focus:border-zinc-500 focus:outline-none"
              />
              <button
                class="text-xs px-2 py-1 rounded border border-zinc-700 hover:bg-red-900/30 text-zinc-500 disabled:opacity-30"
                :disabled="newMembers.length <= 2"
                @click="removeMember(i)"
              >×</button>
            </li>
          </ul>
        </div>

        <div class="flex items-center gap-3 pt-2">
          <button
            class="px-4 py-2 rounded bg-emerald-700 hover:bg-emerald-600 text-white text-sm font-medium disabled:opacity-50"
            :disabled="convening || !newIntentId || !newBinding || !newBishop"
            @click="convene"
          >{{ convening ? 'convening…' : 'Convene' }}</button>
          <button
            class="px-3 py-2 text-xs rounded border border-zinc-700 hover:bg-zinc-800 text-zinc-300"
            @click="showConvene = false"
          >Cancel</button>
          <span v-if="conveneError" class="text-xs text-red-400">{{ conveneError }}</span>
        </div>
      </div>
    </div>
  </div>
</template>
