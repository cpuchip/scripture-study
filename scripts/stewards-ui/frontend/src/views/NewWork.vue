<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { api, type ProviderRow } from '@/api'

const router = useRouter()

const pipeline = ref('study-write')
const slug = ref('')
const bindingQuestion = ref('')
const actor = ref('michael')
const tokenBudget = ref<number | null>(null)
const dispatch = ref(true)

// Phase 2 of NewWork: model picker. Pipeline-level model overrides
// land later; for now this is informational — substrate routes to
// the agent's configured (provider, model) by default. Showing the
// catalog helps Michael see what the substrate can reach.
const providers = ref<ProviderRow[]>([])
const providersError = ref('')
onMounted(async () => {
  try {
    const r = await api.providers()
    providers.value = r.items
  } catch (e) {
    providersError.value = String(e)
  }
})

const submitting = ref(false)
const error = ref('')
const result = ref<{ id: string; dispatched: boolean } | null>(null)

const inputJson = computed(() => {
  if (!bindingQuestion.value.trim()) return {}
  return { binding_question: bindingQuestion.value.trim() }
})

async function submit() {
  submitting.value = true
  error.value = ''
  result.value = null
  try {
    const r = await api.workItemCreate({
      pipeline: pipeline.value,
      slug: slug.value || undefined,
      input: inputJson.value,
      actor: actor.value || 'human',
      token_budget: tokenBudget.value || undefined,
      dispatch: dispatch.value,
    })
    result.value = { id: r.id, dispatched: r.dispatched }
  } catch (e) {
    error.value = String(e)
  } finally {
    submitting.value = false
  }
}

function goToWorkItem() {
  if (result.value) router.push(`/work-items/${result.value.id}`)
}
</script>

<template>
  <div class="space-y-6 max-w-2xl">
    <h2 class="text-2xl font-semibold tracking-tight">New work item</h2>
    <p class="text-sm text-zinc-400">
      Create + (optionally) dispatch a work item. Mirrors what
      <code class="font-mono text-zinc-300">stewards-cli work-item create</code>
      does.
    </p>

    <form class="space-y-4" @submit.prevent="submit">
      <div>
        <label class="block text-xs uppercase tracking-wide text-zinc-500 mb-1">
          Pipeline
        </label>
        <select
          v-model="pipeline"
          class="w-full px-3 py-2 rounded border border-zinc-700 bg-zinc-900 text-sm focus:border-zinc-500 focus:outline-none"
        >
          <option value="study-write">study-write</option>
          <option value="echo-test">echo-test</option>
        </select>
      </div>

      <div>
        <label class="block text-xs uppercase tracking-wide text-zinc-500 mb-1">
          Slug (optional — auto-generated if blank)
        </label>
        <input
          v-model="slug"
          type="text"
          placeholder="e.g. mysteries-of-god-text-vs-spirit"
          class="w-full px-3 py-2 rounded border border-zinc-700 bg-zinc-900 text-sm focus:border-zinc-500 focus:outline-none font-mono"
        />
      </div>

      <div>
        <label class="block text-xs uppercase tracking-wide text-zinc-500 mb-1">
          Binding question
        </label>
        <textarea
          v-model="bindingQuestion"
          rows="6"
          placeholder="What specific question should this work item answer? Be precise — the agent's whole loop hangs on this."
          class="w-full px-3 py-2 rounded border border-zinc-700 bg-zinc-900 text-sm focus:border-zinc-500 focus:outline-none resize-y"
        ></textarea>
      </div>

      <div class="grid grid-cols-2 gap-4">
        <div>
          <label class="block text-xs uppercase tracking-wide text-zinc-500 mb-1">Actor</label>
          <input
            v-model="actor"
            type="text"
            class="w-full px-3 py-2 rounded border border-zinc-700 bg-zinc-900 text-sm focus:border-zinc-500 focus:outline-none"
          />
        </div>
        <div>
          <label class="block text-xs uppercase tracking-wide text-zinc-500 mb-1">
            Token budget (optional)
          </label>
          <input
            v-model.number="tokenBudget"
            type="number"
            min="1000"
            placeholder="e.g. 2000000"
            class="w-full px-3 py-2 rounded border border-zinc-700 bg-zinc-900 text-sm focus:border-zinc-500 focus:outline-none tabular-nums"
          />
        </div>
      </div>

      <label class="flex items-center gap-2 text-sm text-zinc-300 cursor-pointer">
        <input v-model="dispatch" type="checkbox" class="accent-emerald-500" />
        <span>Dispatch first stage immediately after create</span>
      </label>

      <details class="rounded-md border border-zinc-800 bg-zinc-900/50 p-3">
        <summary class="cursor-pointer text-xs text-zinc-400 hover:text-zinc-200">
          Available models (substrate provider catalog)
          <span class="text-zinc-600">— {{ providers.length }} loaded</span>
        </summary>
        <div v-if="providersError" class="text-xs text-red-400 mt-2">
          {{ providersError }}
        </div>
        <ul v-else class="text-xs mt-2 space-y-1">
          <li v-for="p in providers" :key="p.name" class="flex items-baseline gap-3 font-mono">
            <span class="text-zinc-200">{{ p.name }}</span>
            <span class="text-zinc-500">{{ p.default_model }}</span>
            <span class="text-zinc-600">{{ p.kind }}</span>
            <span
              v-if="!p.has_api_key"
              class="text-amber-500 text-[10px] uppercase tracking-wide"
            >no api key</span>
          </li>
        </ul>
        <p class="text-xs text-zinc-500 mt-2">
          Per-pipeline model override is a v2 feature; the substrate currently
          routes to the agent's configured (provider, model) per stage.
        </p>
      </details>

      <div class="flex items-center gap-3">
        <button
          type="submit"
          :disabled="submitting || !bindingQuestion.trim()"
          class="px-4 py-2 rounded bg-emerald-700 hover:bg-emerald-600 text-white text-sm font-medium disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {{ submitting ? 'creating…' : (dispatch ? 'create + dispatch' : 'create') }}
        </button>
        <span v-if="error" class="text-sm text-red-400">{{ error }}</span>
      </div>
    </form>

    <div
      v-if="result"
      class="rounded-md border border-emerald-900/40 bg-emerald-950/20 p-4 space-y-2"
    >
      <div class="text-sm text-emerald-300">
        ✓ Work item created
        <span v-if="result.dispatched" class="text-emerald-400">— first stage dispatched</span>
      </div>
      <div class="text-xs font-mono text-zinc-400">id: {{ result.id }}</div>
      <button
        class="text-xs px-3 py-1 rounded border border-emerald-700 hover:bg-emerald-900/30 text-emerald-200"
        @click="goToWorkItem"
      >
        open detail →
      </button>
    </div>
  </div>
</template>
