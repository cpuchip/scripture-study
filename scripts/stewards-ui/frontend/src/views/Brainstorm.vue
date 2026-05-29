<script setup lang="ts">
// Brainstorm dispatch form — wraps stewards.start_brainstorm() via the
// /api/brainstorm endpoints landed today (2026-05-29). Backs the J.8 + J.9
// SQL work: 4-layer dispatch fallback chain, per-lens model overrides,
// subset lens selection from the 12 available lenses.
//
// Form layout (per ratification Q3 + Q4, 2026-05-29):
//   - Existing 4 lenses pre-checked; 8 new under collapsible "More lenses"
//   - Per-lens model picker (text input; falls back to lens.suggested_model)
//   - Standard fields: binding_question, slug, destination, project,
//     cost cap
//   - Submit calls POST /api/brainstorm/start; on success routes to
//     /work-items/<parent_id>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { api, type BrainstormLensRow, type ProjectRow } from '@/api'

const router = useRouter()

const lenses = ref<BrainstormLensRow[]>([])
const lensesLoading = ref(true)
const lensesError = ref('')

const bindingQuestion = ref('')
const slug = ref('')
const destination = ref('')
const projectAssociation = ref('')
const costCapDollars = ref(0.20) // micro-dollar default 200000 = $0.20
const actor = ref('michael')

// selectedLenses keyed by short_name; true = include
const selectedLenses = ref<Record<string, boolean>>({})
// modelOverrides keyed by short_name; empty string = use lens default
const modelOverrides = ref<Record<string, string>>({})

const showMoreLenses = ref(false)
const submitting = ref(false)
const submitError = ref('')

// Projects (for project_association picker)
const projects = ref<ProjectRow[]>([])

const originalLenses = computed(() => lenses.value.filter(l => l.is_original))
const newLenses = computed(() => lenses.value.filter(l => !l.is_original))

const selectedCount = computed(() =>
  Object.values(selectedLenses.value).filter(v => v).length
)

const placeholderDestination = computed(() => {
  const s = slug.value || 'brainstorm-{timestamp}'
  return `study/.scratch/${s}.md`
})

async function loadLenses() {
  lensesLoading.value = true
  lensesError.value = ''
  try {
    const r = await api.brainstormLenses()
    lenses.value = r.items
    // Pre-check originals per Q4 ratification
    for (const lens of r.items) {
      selectedLenses.value[lens.short_name] = lens.is_original
      modelOverrides.value[lens.short_name] = ''
    }
  } catch (e) {
    lensesError.value = String(e)
  } finally {
    lensesLoading.value = false
  }
}

async function loadProjects() {
  try {
    const r = await api.projectsList()
    projects.value = r.items
  } catch (e) {
    // Non-fatal — project picker just won't populate.
    console.warn('projects load failed:', e)
  }
}

async function onSubmit() {
  submitError.value = ''
  if (!bindingQuestion.value.trim()) {
    submitError.value = 'Binding question is required.'
    return
  }
  if (selectedCount.value === 0) {
    submitError.value = 'Pick at least one lens.'
    return
  }

  const selected: string[] = []
  const models: Record<string, string> = {}
  for (const lens of lenses.value) {
    if (selectedLenses.value[lens.short_name]) {
      selected.push(lens.short_name)
      const override = modelOverrides.value[lens.short_name]?.trim()
      if (override) {
        models[lens.short_name] = override
      }
    }
  }

  submitting.value = true
  try {
    const r = await api.brainstormStart({
      binding_question: bindingQuestion.value.trim(),
      destination: destination.value.trim() || undefined,
      slug: slug.value.trim() || undefined,
      lenses: selected,
      models: Object.keys(models).length > 0 ? models : undefined,
      project_association: projectAssociation.value || undefined,
      actor: actor.value || undefined,
      cost_cap_per_lens_micro: Math.round(costCapDollars.value * 1_000_000),
    })
    router.push(`/work-items/${r.parent_id}`)
  } catch (e) {
    submitError.value = String(e)
  } finally {
    submitting.value = false
  }
}

function checkAll() {
  for (const lens of lenses.value) {
    selectedLenses.value[lens.short_name] = true
  }
}

function checkOriginals() {
  for (const lens of lenses.value) {
    selectedLenses.value[lens.short_name] = lens.is_original
  }
}

function uncheckAll() {
  for (const lens of lenses.value) {
    selectedLenses.value[lens.short_name] = false
  }
}

onMounted(async () => {
  await Promise.all([loadLenses(), loadProjects()])
})
</script>

<template>
  <div class="brainstorm-view">
    <header class="page-header">
      <h1>Brainstorm</h1>
      <p class="subtitle">
        Dispatch a multi-lens brainstorm across the substrate. Pick lenses, optionally override per-lens models,
        and the synthesis aggregator combines them when all children verify.
      </p>
    </header>

    <form @submit.prevent="onSubmit" class="form">
      <!-- Binding question -->
      <div class="field">
        <label for="bq">Binding question <span class="req">*</span></label>
        <textarea
          id="bq"
          v-model="bindingQuestion"
          rows="3"
          placeholder="What specific question should the brainstorm answer? (1-3 sentences)"
          required
        ></textarea>
      </div>

      <!-- Slug + destination -->
      <div class="field-row">
        <div class="field">
          <label for="slug">Slug (optional)</label>
          <input
            id="slug"
            v-model="slug"
            type="text"
            placeholder="brainstorm-YYYYMMDD-HHMMSS (auto)"
          />
        </div>
        <div class="field grow">
          <label for="dest">Destination (optional)</label>
          <input
            id="dest"
            v-model="destination"
            type="text"
            :placeholder="placeholderDestination"
          />
        </div>
      </div>

      <!-- Project + cost cap -->
      <div class="field-row">
        <div class="field">
          <label for="proj">Project association</label>
          <select id="proj" v-model="projectAssociation">
            <option value="">(none)</option>
            <option v-for="p in projects" :key="p.slug" :value="p.slug">{{ p.slug }}</option>
          </select>
        </div>
        <div class="field">
          <label for="cap">Cost cap per lens (USD)</label>
          <input
            id="cap"
            v-model.number="costCapDollars"
            type="number"
            step="0.05"
            min="0.01"
          />
        </div>
        <div class="field">
          <label for="actor">Actor</label>
          <input id="actor" v-model="actor" type="text" />
        </div>
      </div>

      <!-- Lens picker -->
      <div class="lens-section">
        <div class="lens-header">
          <h2>Lenses ({{ selectedCount }} selected)</h2>
          <div class="lens-actions">
            <button type="button" @click="checkOriginals" class="ghost">Originals only</button>
            <button type="button" @click="checkAll" class="ghost">All 12</button>
            <button type="button" @click="uncheckAll" class="ghost">None</button>
          </div>
        </div>

        <div v-if="lensesLoading" class="status">Loading lenses…</div>
        <div v-else-if="lensesError" class="error">Could not load lenses: {{ lensesError }}</div>

        <div v-else class="lens-grid">
          <!-- Originals (always visible) -->
          <div class="lens-group">
            <h3>Originals (J.4)</h3>
            <div
              v-for="lens in originalLenses"
              :key="lens.short_name"
              class="lens-row"
            >
              <label class="lens-check">
                <input type="checkbox" v-model="selectedLenses[lens.short_name]" />
                <span class="lens-name">{{ lens.short_name }}</span>
              </label>
              <input
                type="text"
                v-model="modelOverrides[lens.short_name]"
                :placeholder="lens.suggested_model || lens.default_model || 'model override'"
                class="model-input"
                :disabled="!selectedLenses[lens.short_name]"
              />
              <p class="lens-desc">{{ lens.description }}</p>
            </div>
          </div>

          <!-- New lenses (J.9, collapsible) -->
          <div class="lens-group">
            <button
              type="button"
              class="more-toggle"
              @click="showMoreLenses = !showMoreLenses"
            >
              <span>{{ showMoreLenses ? '▼' : '▶' }}</span>
              More lenses (J.9, 8 added) — {{ showMoreLenses ? 'hide' : 'show' }}
            </button>
            <div v-if="showMoreLenses">
              <div
                v-for="lens in newLenses"
                :key="lens.short_name"
                class="lens-row"
              >
                <label class="lens-check">
                  <input type="checkbox" v-model="selectedLenses[lens.short_name]" />
                  <span class="lens-name">{{ lens.short_name }}</span>
                </label>
                <input
                  type="text"
                  v-model="modelOverrides[lens.short_name]"
                  :placeholder="lens.suggested_model || lens.default_model || 'model override'"
                  class="model-input"
                  :disabled="!selectedLenses[lens.short_name]"
                />
                <p class="lens-desc">{{ lens.description }}</p>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Submit -->
      <div v-if="submitError" class="error">{{ submitError }}</div>
      <div class="actions">
        <button type="submit" :disabled="submitting || selectedCount === 0" class="primary">
          {{ submitting ? 'Dispatching…' : `Dispatch brainstorm (${selectedCount} lens${selectedCount === 1 ? '' : 'es'})` }}
        </button>
      </div>
    </form>
  </div>
</template>

<style scoped>
.brainstorm-view {
  max-width: 980px;
  margin: 0 auto;
  padding: 1rem 2rem 4rem;
}

.page-header h1 {
  margin: 0 0 0.25rem;
}
.subtitle {
  color: var(--text-muted, #888);
  margin: 0 0 1.5rem;
}

.form {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}
.field label {
  font-size: 0.875rem;
  font-weight: 500;
}
.field.grow { flex: 1; }
.req { color: #c0392b; }
.field input,
.field textarea,
.field select {
  padding: 0.5rem 0.6rem;
  border: 1px solid var(--border, #444);
  border-radius: 4px;
  background: var(--surface, #1a1a1a);
  color: inherit;
  font-family: inherit;
  font-size: 0.95rem;
}

.field-row {
  display: flex;
  gap: 1rem;
  flex-wrap: wrap;
}
.field-row .field {
  flex: 1 1 200px;
}

.lens-section {
  border: 1px solid var(--border, #444);
  border-radius: 6px;
  padding: 1rem;
}
.lens-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}
.lens-header h2 { margin: 0; font-size: 1.1rem; }
.lens-actions { display: flex; gap: 0.5rem; }
.ghost {
  background: transparent;
  border: 1px solid var(--border, #444);
  color: inherit;
  padding: 0.3rem 0.6rem;
  border-radius: 4px;
  font-size: 0.85rem;
  cursor: pointer;
}
.ghost:hover { background: var(--surface-hover, #2a2a2a); }

.lens-group h3 {
  font-size: 0.95rem;
  margin: 0.5rem 0 0.75rem;
  color: var(--text-muted, #888);
}

.lens-row {
  display: grid;
  grid-template-columns: 200px 200px 1fr;
  gap: 0.75rem;
  align-items: start;
  padding: 0.5rem 0;
  border-bottom: 1px solid var(--border-subtle, #2a2a2a);
}
.lens-row:last-child { border-bottom: none; }

.lens-check {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  cursor: pointer;
}
.lens-name { font-weight: 500; font-family: var(--font-mono, monospace); }

.model-input {
  padding: 0.3rem 0.5rem;
  border: 1px solid var(--border, #444);
  border-radius: 4px;
  background: var(--surface, #1a1a1a);
  color: inherit;
  font-family: var(--font-mono, monospace);
  font-size: 0.85rem;
}
.model-input:disabled { opacity: 0.4; }

.lens-desc {
  font-size: 0.85rem;
  color: var(--text-muted, #888);
  margin: 0;
}

.more-toggle {
  background: transparent;
  border: none;
  color: var(--accent, #4a9eff);
  padding: 0.5rem 0;
  text-align: left;
  cursor: pointer;
  font-size: 0.95rem;
  display: flex;
  align-items: center;
  gap: 0.4rem;
}

.actions {
  display: flex;
  justify-content: flex-end;
}
.primary {
  padding: 0.6rem 1.2rem;
  background: var(--accent, #4a9eff);
  border: none;
  border-radius: 4px;
  color: white;
  font-weight: 500;
  cursor: pointer;
  font-size: 0.95rem;
}
.primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.status {
  padding: 1rem;
  color: var(--text-muted, #888);
  text-align: center;
}
.error {
  padding: 0.75rem 1rem;
  background: rgba(192, 57, 43, 0.15);
  border: 1px solid #c0392b;
  border-radius: 4px;
  color: #ff6b5b;
}
</style>
