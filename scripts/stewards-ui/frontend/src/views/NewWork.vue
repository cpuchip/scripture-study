<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter, RouterLink } from 'vue-router'
import { api, type ProviderRow, type IntentRow, type PipelineRow } from '@/api'

const router = useRouter()

const pipeline = ref('study-write')
const pipelines = ref<PipelineRow[]>([])
const pipelinesError = ref('')
// Batch G.4.4 — file destination
const writeFile = ref(false)
const fileDestination = ref('')
// Track the last value we auto-rendered so we can distinguish "still
// matches our auto-render" (safe to update on slug change) from "user
// has manually edited" (leave alone). Without this, the watcher's
// `.includes('<slug>')` guard stopped firing after the first slug
// substitution, so subsequent slug edits never updated the path.
// Bug surfaced 2026-05-11 (physics-news work_item materialized as
// research/P.md because slug was 'P' when fileDestination first
// rendered; later slug edits didn't propagate).
const lastAutoRendered = ref('')
const slug = ref('')
const bindingQuestion = ref('')
const actor = ref('michael')
const tokenBudget = ref<number | null>(null)
const dispatch = ref(true)
// Phase 5a: destination maturity = human's ceiling for the maturity
// ladder. Empty string = NULL = full Ammon-loop to verified. Pick a
// lower rung to have the substrate surface for human review before
// continuing past that rung.
const destinationMaturity = ref<string>('')
// Phase 5d (C.8): intent_id required at creation per D-C3.
const intents = ref<IntentRow[]>([])
const intentId = ref<string>('')
const intentsError = ref('')

// Projects (Batch I.1)
import { type ProjectRow } from '@/api'
const projects = ref<ProjectRow[]>([])
const projectAssociation = ref<string>('')
const projectsError = ref('')

// Inline create-new-intent modal state
const showCreateIntent = ref(false)
const newIntentSlug = ref('')
const newIntentPurpose = ref('')
const newIntentBeneficiary = ref('')
const newIntentScripture = ref('')
const creatingIntent = ref(false)
const createIntentError = ref('')

async function loadIntents() {
  try {
    const r = await api.intentsList()
    intents.value = r.items
    if (intents.value.length > 0 && !intentId.value) {
      // Default to scripture-study (the project-level intent), if present
      const defaultIntent = intents.value.find(i => i.slug === 'scripture-study')
      const fallback = intents.value[0]
      if (defaultIntent) {
        intentId.value = defaultIntent.id
      } else if (fallback) {
        intentId.value = fallback.id
      }
    }
  } catch (e) {
    intentsError.value = String(e)
  }
}

async function createIntent() {
  if (!newIntentSlug.value || !newIntentPurpose.value) {
    createIntentError.value = 'slug + purpose required'
    return
  }
  creatingIntent.value = true
  createIntentError.value = ''
  try {
    const r = await api.intentCreate({
      slug: newIntentSlug.value,
      purpose: newIntentPurpose.value,
      beneficiary: newIntentBeneficiary.value || undefined,
      scripture_anchor: newIntentScripture.value || undefined,
    })
    await loadIntents()
    intentId.value = r.id
    showCreateIntent.value = false
    newIntentSlug.value = ''
    newIntentPurpose.value = ''
    newIntentBeneficiary.value = ''
    newIntentScripture.value = ''
  } catch (e) {
    createIntentError.value = String(e)
  } finally {
    creatingIntent.value = false
  }
}

// Phase 2 of NewWork: model picker. Pipeline-level model overrides
// land later; for now this is informational — substrate routes to
// the agent's configured (provider, model) by default. Showing the
// catalog helps Michael see what the substrate can reach.
const providers = ref<ProviderRow[]>([])
const providersError = ref('')
async function loadPipelines() {
  try {
    const r = await api.pipelinesList()
    pipelines.value = r.items
  } catch (e) {
    pipelinesError.value = String(e)
  }
}

// When pipeline changes, prefill file destination from the pipeline's
// template (D-G1: suggestion only — human can change or unset).
function renderTemplate(template: string): string {
  // <slug> placeholder; if user hasn't typed a slug, leave the literal
  return template.replace(/<slug>/g, slug.value || '<slug>')
                 .replace(/<id>/g, '<id>')
}

watch([pipeline, slug, pipelines], () => {
  const p = pipelines.value.find(pp => pp.family === pipeline.value)
  if (p?.file_destination_template) {
    const rendered = renderTemplate(p.file_destination_template)
    if (!writeFile.value && !fileDestination.value) {
      // First time we're rendering: enable writeFile and set both refs.
      writeFile.value = true
      fileDestination.value = rendered
      lastAutoRendered.value = rendered
    } else if (writeFile.value && fileDestination.value === lastAutoRendered.value) {
      // Field still shows our prior auto-render → safe to re-render on
      // slug/pipeline change. Once the user edits the input, the two
      // refs diverge and this branch stops firing — manual edits are
      // preserved.
      fileDestination.value = rendered
      lastAutoRendered.value = rendered
    }
  } else {
    // Pipeline has no template → DB-only by default
    if (!fileDestination.value) {
      writeFile.value = false
    }
  }
}, { immediate: false })

onMounted(async () => {
  try {
    const r = await api.providers()
    providers.value = r.items
  } catch (e) {
    providersError.value = String(e)
  }
  await loadIntents()
  await loadPipelines()
  try {
    const r = await api.projectsList(false)
    projects.value = r.items
  } catch (e) {
    projectsError.value = String(e)
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
      destination_maturity: destinationMaturity.value || undefined,
      intent_id: intentId.value || undefined,
      file_destination: writeFile.value && fileDestination.value
        ? fileDestination.value.replace(/<slug>/g, slug.value || '<slug>')
        : undefined,
      project_association: projectAssociation.value || undefined,
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
          Intent <span class="text-red-400">*</span>
        </label>
        <div class="flex gap-2">
          <select
            v-model="intentId"
            required
            class="flex-1 px-3 py-2 rounded border border-zinc-700 bg-zinc-900 text-sm focus:border-zinc-500 focus:outline-none"
          >
            <option value="">— pick one —</option>
            <option v-for="i in intents" :key="i.id" :value="i.id">
              {{ i.slug }} — {{ i.purpose.slice(0, 80) }}{{ i.purpose.length > 80 ? '…' : '' }}
            </option>
          </select>
          <button
            type="button"
            class="px-3 py-2 text-xs rounded border border-zinc-700 hover:bg-zinc-800 text-zinc-300"
            @click="showCreateIntent = true"
          >+ new</button>
        </div>
        <p v-if="intentsError" class="text-xs text-red-400 mt-1">{{ intentsError }}</p>
        <p class="text-xs text-zinc-500 mt-1">
          Required (D-C3 — friction is the discipline). The substrate injects the intent's
          purpose + values into every dispatched chat for this work_item.
        </p>
      </div>

      <div>
        <label class="block text-xs uppercase tracking-wide text-zinc-500 mb-1">
          Pipeline
        </label>
        <select
          v-model="pipeline"
          class="w-full px-3 py-2 rounded border border-zinc-700 bg-zinc-900 text-sm focus:border-zinc-500 focus:outline-none"
        >
          <option v-for="p in pipelines" :key="p.family" :value="p.family">
            {{ p.family }}{{ p.file_destination_template ? '  (→ ' + p.file_destination_template + ')' : '' }}
          </option>
          <option v-if="pipelines.length === 0" value="study-write">study-write</option>
        </select>
        <p v-if="pipelinesError" class="text-xs text-red-400 mt-1">{{ pipelinesError }}</p>
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

      <!-- Batch I.1 — project picker -->
      <div>
        <label class="block text-xs uppercase tracking-wide text-zinc-500 mb-1">
          Project (optional)
        </label>
        <div class="flex gap-2">
          <select
            v-model="projectAssociation"
            class="flex-1 px-3 py-2 rounded border border-zinc-700 bg-zinc-900 text-sm focus:border-zinc-500 focus:outline-none"
          >
            <option value="">— (no project)</option>
            <option v-for="p in projects" :key="p.slug" :value="p.slug">
              {{ p.slug }} — {{ p.name }}
            </option>
          </select>
          <RouterLink
            to="/projects"
            class="px-3 py-2 rounded border border-zinc-700 hover:bg-zinc-800 text-xs text-zinc-300 self-center"
            title="Manage projects"
          >manage ↗</RouterLink>
        </div>
        <p v-if="projectsError" class="text-xs text-red-400 mt-1">{{ projectsError }}</p>
        <p class="text-xs text-zinc-500 mt-1">
          Groups related work_items. Show on the WorkItems list as a chip. New
          projects via the Projects page.
        </p>
      </div>

      <div>
        <label class="block text-xs uppercase tracking-wide text-zinc-500 mb-1">
          Destination maturity (optional ceiling)
        </label>
        <select
          v-model="destinationMaturity"
          class="w-full px-3 py-2 rounded border border-zinc-700 bg-zinc-900 text-sm focus:border-zinc-500 focus:outline-none"
        >
          <option value="">— full Ammon-loop (default → verified)</option>
          <option value="researched">researched (stop after research)</option>
          <option value="planned">planned (stop after outline/plan)</option>
          <option value="specced">specced (stop after spec)</option>
          <option value="executing">executing (stop after first draft)</option>
          <option value="verified">verified (explicit, same as default)</option>
        </select>
        <p class="text-xs text-zinc-500 mt-1">
          The substrate's gate ladder surfaces work for review when it reaches
          this rung. Leave blank to let it run all the way to verified.
        </p>
      </div>

      <label class="flex items-center gap-2 text-sm text-zinc-300 cursor-pointer">
        <input v-model="dispatch" type="checkbox" class="accent-emerald-500" />
        <span>Dispatch first stage immediately after create</span>
      </label>

      <!-- Batch G.4.4 — file destination (DB-only by default; opt-in) -->
      <div class="rounded-md border border-zinc-800 bg-zinc-900/30 p-3 space-y-2">
        <label class="flex items-center gap-2 text-sm text-zinc-300 cursor-pointer">
          <input v-model="writeFile" type="checkbox" class="accent-emerald-500" />
          <span>Write to file when complete</span>
        </label>
        <div v-if="writeFile">
          <label class="block text-xs uppercase tracking-wide text-zinc-500 mb-1">
            File destination
          </label>
          <input
            v-model="fileDestination"
            type="text"
            placeholder="e.g. study/substrate--<slug>.md"
            class="w-full px-3 py-2 rounded border border-zinc-700 bg-zinc-950 text-sm font-mono focus:border-zinc-500 focus:outline-none"
          />
          <p class="text-xs text-zinc-500 mt-1">
            <code class="font-mono">&lt;slug&gt;</code> renders to the work_item slug at materialization.
            Editable later from WorkItemDetail. Materialization is explicit
            (click "Materialize now" after review) — files don't write at completion.
          </p>
        </div>
        <p v-else class="text-xs text-zinc-500">
          DB-only by default. The work_item lives in the substrate; no file
          on disk. You can change this any time before clicking "Materialize now"
          on WorkItemDetail.
        </p>
      </div>

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
          :disabled="submitting || !bindingQuestion.trim() || !intentId"
          class="px-4 py-2 rounded bg-emerald-700 hover:bg-emerald-600 text-white text-sm font-medium disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {{ submitting ? 'creating…' : (dispatch ? 'create + dispatch' : 'create') }}
        </button>
        <span v-if="error" class="text-sm text-red-400">{{ error }}</span>
      </div>
    </form>

    <!-- Inline create-new-intent modal (Phase 5d C.8) -->
    <div
      v-if="showCreateIntent"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/60"
      @click.self="showCreateIntent = false"
    >
      <div class="bg-zinc-900 border border-zinc-700 rounded-lg p-5 max-w-lg w-full space-y-3">
        <h3 class="text-lg font-semibold">Create new intent</h3>
        <p class="text-xs text-zinc-500">
          Substrate-native intents created here are NOT in YAML. Use this for one-off work
          that doesn't fit any repo-tracked intent.
        </p>
        <div>
          <label class="block text-xs uppercase tracking-wide text-zinc-500 mb-1">Slug *</label>
          <input
            v-model="newIntentSlug"
            type="text"
            placeholder="e.g. spike-lightrag-eval"
            class="w-full px-3 py-2 rounded border border-zinc-700 bg-zinc-950 text-sm font-mono focus:border-zinc-500 focus:outline-none"
          />
        </div>
        <div>
          <label class="block text-xs uppercase tracking-wide text-zinc-500 mb-1">Purpose *</label>
          <textarea
            v-model="newIntentPurpose"
            rows="3"
            placeholder="The why. One paragraph."
            class="w-full px-3 py-2 rounded border border-zinc-700 bg-zinc-950 text-sm focus:border-zinc-500 focus:outline-none resize-y"
          />
        </div>
        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="block text-xs uppercase tracking-wide text-zinc-500 mb-1">Beneficiary</label>
            <input
              v-model="newIntentBeneficiary"
              type="text"
              placeholder="who benefits"
              class="w-full px-3 py-2 rounded border border-zinc-700 bg-zinc-950 text-sm focus:border-zinc-500 focus:outline-none"
            />
          </div>
          <div>
            <label class="block text-xs uppercase tracking-wide text-zinc-500 mb-1">Scripture anchor</label>
            <input
              v-model="newIntentScripture"
              type="text"
              placeholder="e.g. D&C 88:118"
              class="w-full px-3 py-2 rounded border border-zinc-700 bg-zinc-950 text-sm focus:border-zinc-500 focus:outline-none"
            />
          </div>
        </div>
        <div class="flex items-center gap-3 pt-2">
          <button
            type="button"
            class="px-4 py-2 rounded bg-emerald-700 hover:bg-emerald-600 text-white text-sm font-medium disabled:opacity-50"
            :disabled="creatingIntent || !newIntentSlug || !newIntentPurpose"
            @click="createIntent"
          >{{ creatingIntent ? 'creating…' : 'Create' }}</button>
          <button
            type="button"
            class="px-3 py-2 text-xs rounded border border-zinc-700 hover:bg-zinc-800 text-zinc-300"
            @click="showCreateIntent = false"
          >Cancel</button>
          <span v-if="createIntentError" class="text-xs text-red-400">{{ createIntentError }}</span>
        </div>
      </div>
    </div>

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
