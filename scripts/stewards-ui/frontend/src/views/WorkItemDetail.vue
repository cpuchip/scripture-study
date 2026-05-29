<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import { api, type WorkItemDetail, type CostEventsResp, type StewardActionsResp, type GateDecisionsResp } from '@/api'

const route = useRoute()
const wi = ref<WorkItemDetail | null>(null)
const cost = ref<CostEventsResp | null>(null)
const actions = ref<StewardActionsResp | null>(null)
const gateDecisions = ref<GateDecisionsResp | null>(null)
const error = ref<string>('')
const loading = ref(false)

// H.3-followup-B: proposal action state
const actionBusy = ref(false)
const actionMsg = ref<string>('')
const actionErr = ref<string>('')

async function ratifyProposal() {
  if (!wi.value) return
  actionBusy.value = true
  actionMsg.value = ''
  actionErr.value = ''
  try {
    const r = await api.workItemRatify(wi.value.id)
    actionMsg.value = r.message || 'ratified'
    await load(idFromRoute.value)
  } catch (e) {
    actionErr.value = String(e)
  } finally {
    actionBusy.value = false
  }
}

async function dispatchProposal() {
  if (!wi.value) return
  actionBusy.value = true
  actionMsg.value = ''
  actionErr.value = ''
  try {
    const r = await api.workItemDispatch(wi.value.id)
    actionMsg.value = r.message
      ? `${r.message} (work_queue id=${r.work_queue_id})`
      : `dispatched (work_queue id=${r.work_queue_id})`
    await load(idFromRoute.value)
  } catch (e) {
    actionErr.value = String(e)
  } finally {
    actionBusy.value = false
  }
}

async function cancelProposal() {
  if (!wi.value) return
  if (!confirm(`Cancel proposal "${wi.value.slug}"? It stays in the DB but won't surface in active queues.`)) return
  actionBusy.value = true
  actionMsg.value = ''
  actionErr.value = ''
  try {
    const r = await api.workItemCancelProposal(wi.value.id, 'cancelled via UI')
    actionMsg.value = r.message || 'cancelled'
    await load(idFromRoute.value)
  } catch (e) {
    actionErr.value = String(e)
  } finally {
    actionBusy.value = false
  }
}

const isProposal = computed(() =>
  wi.value?.origin === 'agent_planning' && wi.value?.status !== 'cancelled',
)

// ---------------------------------------------------------------------
// Edit + AI-revise (third proposal mode)
// ---------------------------------------------------------------------

import { type PendingRevisionRow, type RevisionJSON } from '@/api'

// Direct edit state
const editing = ref(false)
const editForm = ref({
  binding_question: '',
  slug: '',
  pipeline_family_hint: '',
  project_association: '',
  rationale: '',
})
const editBusy = ref(false)

function startEdit() {
  if (!wi.value) return
  const input = wi.value.input as any
  editForm.value = {
    binding_question: input?.binding_question ?? '',
    slug: wi.value.slug ?? '',
    pipeline_family_hint: wi.value.pipeline ?? '',
    project_association: wi.value.project_association ?? '',
    rationale: input?.rationale_from_planning ?? '',
  }
  editing.value = true
  actionMsg.value = ''
  actionErr.value = ''
}

function cancelEdit() {
  editing.value = false
}

async function submitEdit() {
  if (!wi.value) return
  editBusy.value = true
  actionMsg.value = ''
  actionErr.value = ''
  try {
    const input = wi.value.input as any
    const req: any = { id: wi.value.id }
    if (editForm.value.binding_question !== (input?.binding_question ?? '')) {
      req.binding_question = editForm.value.binding_question
    }
    if (editForm.value.slug !== wi.value.slug) {
      req.slug = editForm.value.slug
    }
    if (editForm.value.pipeline_family_hint !== wi.value.pipeline) {
      req.pipeline_family_hint = editForm.value.pipeline_family_hint
    }
    if (editForm.value.project_association !== (wi.value.project_association ?? '')) {
      req.project_association = editForm.value.project_association
    }
    if (editForm.value.rationale !== (input?.rationale_from_planning ?? '')) {
      req.rationale = editForm.value.rationale
    }
    const r = await api.workItemEditProposal(req)
    actionMsg.value = r.message || 'edited'
    editing.value = false
    await load(idFromRoute.value)
  } catch (e) {
    actionErr.value = String(e)
  } finally {
    editBusy.value = false
  }
}

// AI revise state
const reviseOpen = ref(false)
const reviseFeedback = ref('')
const reviseBusy = ref(false)
const pendingRevisions = ref<PendingRevisionRow[]>([])
let revisionsPollTimer: number | undefined

function openRevise() {
  reviseOpen.value = true
  reviseFeedback.value = ''
}

function cancelRevise() {
  reviseOpen.value = false
  reviseFeedback.value = ''
}

async function submitRevise() {
  if (!wi.value || !reviseFeedback.value.trim()) return
  reviseBusy.value = true
  actionMsg.value = ''
  actionErr.value = ''
  try {
    const r = await api.workItemReviseWithFeedback(wi.value.id, reviseFeedback.value.trim())
    actionMsg.value = r.message
    reviseOpen.value = false
    reviseFeedback.value = ''
    // Kick off polling for the revision result.
    await loadPendingRevisions()
    startRevisionsPoll()
  } catch (e) {
    actionErr.value = String(e)
  } finally {
    reviseBusy.value = false
  }
}

async function loadPendingRevisions() {
  if (!wi.value) return
  try {
    const r = await api.workItemPendingRevisions(wi.value.id)
    pendingRevisions.value = r.revisions
    // Stop polling once all revisions are terminal (completed or in their
    // diff-card state with revision_json set).
    const stillRunning = r.revisions.some(rev => rev.status === 'in_progress' || rev.status === 'pending')
    if (!stillRunning && revisionsPollTimer) {
      window.clearInterval(revisionsPollTimer)
      revisionsPollTimer = undefined
    }
  } catch {
    /* tolerate transient errors */
  }
}

function startRevisionsPoll() {
  if (revisionsPollTimer) window.clearInterval(revisionsPollTimer)
  revisionsPollTimer = window.setInterval(loadPendingRevisions, 4000)
}

async function acceptRevision(revId: string) {
  try {
    const r = await api.workItemApplyRevision(revId)
    actionMsg.value = r.message || 'applied'
    await load(idFromRoute.value)
    await loadPendingRevisions()
  } catch (e) {
    actionErr.value = String(e)
  }
}

async function rejectRevision(revId: string) {
  try {
    const r = await api.workItemRejectRevision(revId, 'rejected via UI')
    actionMsg.value = r.message || 'rejected'
    await loadPendingRevisions()
  } catch (e) {
    actionErr.value = String(e)
  }
}

// Diff rendering helper. Returns labeled before/after rows for any
// revision field that's actually being changed (compares revision JSON
// to current work_item state).
function diffRows(rev: PendingRevisionRow): Array<{ label: string; before: string; after: string }> {
  const out: Array<{ label: string; before: string; after: string }> = []
  if (!wi.value || !rev.revision_json) return out
  const r = rev.revision_json as RevisionJSON
  const input = wi.value.input as any
  if (r.binding_question !== undefined) {
    out.push({
      label: 'binding_question',
      before: input?.binding_question ?? '',
      after: r.binding_question,
    })
  }
  if (r.rationale !== undefined) {
    out.push({
      label: 'rationale',
      before: input?.rationale_from_planning ?? '',
      after: r.rationale,
    })
  }
  if (r.slug !== undefined) {
    out.push({ label: 'slug', before: wi.value.slug ?? '', after: r.slug })
  }
  if (r.pipeline_family_hint !== undefined) {
    out.push({
      label: 'pipeline_family',
      before: wi.value.pipeline ?? '',
      after: r.pipeline_family_hint ?? '(cleared)',
    })
  }
  if (r.project_association !== undefined) {
    out.push({
      label: 'project_association',
      before: wi.value.project_association ?? '',
      after: r.project_association ?? '(cleared)',
    })
  }
  return out
}

// Load pending revisions whenever the work_item loads.
watch(wi, async (v) => {
  if (v && isProposal.value) {
    await loadPendingRevisions()
    // If any revision is still running, start polling.
    if (pendingRevisions.value.some(r => r.status === 'in_progress' || r.status === 'pending')) {
      startRevisionsPoll()
    }
  }
})

import { onUnmounted } from 'vue'
onUnmounted(() => {
  if (revisionsPollTimer) window.clearInterval(revisionsPollTimer)
})

async function load(idOrSlug: string) {
  loading.value = true
  error.value = ''
  wi.value = null
  cost.value = null
  actions.value = null
  gateDecisions.value = null
  try {
    const detail = await api.workItemGet(idOrSlug)
    wi.value = detail
    // Fire cost + actions + gate decisions in parallel; failures don't block the view
    const [c, a, g] = await Promise.allSettled([
      api.workItemCost(detail.id),
      api.workItemActions(detail.id),
      api.workItemGateDecisions(detail.id),
    ])
    if (c.status === 'fulfilled') cost.value = c.value
    if (a.status === 'fulfilled') actions.value = a.value
    if (g.status === 'fulfilled') gateDecisions.value = g.value
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
// Batch G.4.4 — file destination panel
const editingFileDestination = ref(false)
const editFileDestinationValue = ref('')
const savingFileDestination = ref(false)
const fileDestinationError = ref('')
const materializingFile = ref(false)
const materializeMsg = ref('')

function startEditFileDestination() {
  editFileDestinationValue.value = wi.value?.file_destination || ''
  editingFileDestination.value = true
  fileDestinationError.value = ''
}

function useTemplateForFileDestination() {
  if (!wi.value?.pipeline_file_template) return
  editFileDestinationValue.value = wi.value.pipeline_file_template
    .replace(/<slug>/g, wi.value.slug || '<slug>')
}

async function saveFileDestination() {
  if (!wi.value) return
  savingFileDestination.value = true
  fileDestinationError.value = ''
  try {
    await api.workItemSetFileDestination({
      id: wi.value.id,
      file_destination: editFileDestinationValue.value.trim(),
    })
    editingFileDestination.value = false
    await load(wi.value.id)
  } catch (e) {
    fileDestinationError.value = String(e)
  } finally {
    savingFileDestination.value = false
  }
}

async function materializeFile() {
  if (!wi.value) return
  materializingFile.value = true
  materializeMsg.value = ''
  try {
    const r = await api.workItemMaterializeFile(wi.value.id)
    if (r.skipped) {
      materializeMsg.value = `Skipped: ${r.skip_reason || 'no destination set'}`
    } else {
      materializeMsg.value = `Queued pending file write #${r.pending_file_write_id}. Run \`stewards-cli materialize-writes\` (or commit — pre-commit hook materializes automatically).`
    }
    await load(wi.value.id)
  } catch (e) {
    materializeMsg.value = `Error: ${e}`
  } finally {
    materializingFile.value = false
  }
}

// Phase 5a — maturity ladder helpers
const MATURITY_LADDER = ['raw', 'researched', 'planned', 'specced', 'executing', 'verified']
function maturityIndex(m: string): number {
  return MATURITY_LADDER.indexOf(m)
}
function gateActionTone(action: string): string {
  if (action === 'advance') return 'text-emerald-300 border-emerald-700/50 bg-emerald-900/30'
  if (action === 'revise') return 'text-amber-300 border-amber-700/50 bg-amber-900/30'
  if (action === 'surface') return 'text-blue-300 border-blue-700/50 bg-blue-900/30'
  return 'text-zinc-300 border-zinc-700 bg-zinc-800'
}

// Phase 5f (E.7) — gate override modal state
const overrideOpen = ref<{ id: number; current_action: string } | null>(null)
const overrideNewAction = ref<'advance' | 'revise' | 'surface'>('revise')
const overrideJustification = ref('')
const overrideBy = ref('michael')
const overriding = ref(false)
const overrideError = ref('')

function openOverride(g: { id: number; action: string }) {
  overrideOpen.value = { id: g.id, current_action: g.action }
  overrideNewAction.value = g.action === 'advance' ? 'revise'
                          : g.action === 'revise'  ? 'surface'
                          : 'advance'
  overrideJustification.value = ''
  overrideError.value = ''
}

async function submitOverride() {
  if (!overrideOpen.value) return
  if (overrideJustification.value.trim().length < 10) {
    overrideError.value = 'Justification must be at least 10 characters'
    return
  }
  if (overrideNewAction.value === overrideOpen.value.current_action) {
    overrideError.value = 'New action must differ from original'
    return
  }
  overriding.value = true
  overrideError.value = ''
  try {
    await api.gateOverrideApply({
      gate_decision_id: overrideOpen.value.id,
      overridden_by: overrideBy.value || 'human',
      new_action: overrideNewAction.value,
      justification: overrideJustification.value,
    })
    overrideOpen.value = null
    if (wi.value) await load(wi.value.id)
  } catch (e) {
    overrideError.value = String(e)
  } finally {
    overriding.value = false
  }
}
function scenariosArray(s: unknown): string[] {
  if (!Array.isArray(s)) return []
  return s.map((x) => (typeof x === 'string' ? x : JSON.stringify(x)))
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

// J.12 — human label + hint for a classified error category. Returns null
// for 'none'/'other'/unknown so the banner only shows for actionable kinds.
function errorCategoryInfo(cat?: string): { label: string; hint: string; budget: boolean } | null {
  switch (cat) {
    case 'spend_cap_reached':
      return { label: '💸 Spend cap reached', budget: true,
        hint: "The substrate refused this dispatch — the provider hit its enforced spend cap. Top up + reset: SELECT stewards.provider_cap_refill('<provider>');" }
    case 'provider_budget':
      return { label: '💸 Provider budget / quota exhausted', budget: true,
        hint: 'The provider rejected the call for quota/billing/balance (e.g. a Gemini prepaid key out of credit). Refill the provider balance, then retry.' }
    case 'rate_limited':
      return { label: '⏳ Rate limited', budget: false,
        hint: 'Transient rate limit — usually retryable shortly.' }
    case 'auth':
      return { label: '🔑 Auth / key problem', budget: false,
        hint: 'Bad or missing API key, or permission denied for this provider.' }
    case 'timeout':
      return { label: '⌛ Timed out', budget: false,
        hint: 'The request exceeded its timeout.' }
    default:
      return null
  }
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

      <!-- J.12 — classified failure banner (budget/cap stands out) -->
      <section
        v-if="errorCategoryInfo(wi.error_category)"
        :class="[
          'rounded-md border p-4 text-sm mb-3',
          errorCategoryInfo(wi.error_category)!.budget
            ? 'border-amber-600/60 bg-amber-950/30'
            : 'border-orange-800/50 bg-orange-950/20'
        ]"
      >
        <div
          :class="[
            'font-semibold mb-1',
            errorCategoryInfo(wi.error_category)!.budget ? 'text-amber-300' : 'text-orange-300'
          ]"
        >{{ errorCategoryInfo(wi.error_category)!.label }}</div>
        <div class="text-zinc-300">{{ errorCategoryInfo(wi.error_category)!.hint }}</div>
      </section>

      <section
        v-if="wi.error"
        class="rounded-md border border-red-900/40 bg-red-950/20 p-4 text-sm"
      >
        <div class="text-xs uppercase tracking-wide text-red-400 mb-1">
          Error<span v-if="wi.error_category && wi.error_category !== 'none' && wi.error_category !== 'other'" class="ml-2 normal-case text-red-300/70">[{{ wi.error_category }}]</span>
        </div>
        <pre class="whitespace-pre-wrap text-red-300 font-mono text-xs">{{ wi.error }}</pre>
      </section>

      <!-- H.3-followup-B: proposal action panel for agent_planning work_items -->
      <section
        v-if="isProposal"
        class="rounded-md border border-purple-800/60 bg-purple-950/20 p-4 space-y-3"
      >
        <div class="flex items-baseline justify-between">
          <div class="text-xs uppercase tracking-wide text-purple-300 flex items-center gap-2">
            <span>✨ Proposed work</span>
            <span class="text-zinc-500 normal-case">— from planning run</span>
          </div>
          <RouterLink
            v-if="wi.parent_work_item_id"
            :to="`/work-items/${wi.parent_work_item_id}`"
            class="text-xs text-purple-300 hover:text-purple-100 font-mono"
          >
            ← parent
          </RouterLink>
        </div>

        <div v-if="wi.input && (wi.input as any).rationale_from_planning"
             class="text-sm text-zinc-300 leading-relaxed">
          <span class="text-xs text-zinc-500 uppercase tracking-wide">Rationale:</span>
          {{ (wi.input as any).rationale_from_planning }}
        </div>

        <div class="flex flex-wrap gap-2 items-center">
          <button
            v-if="wi.maturity === 'raw'"
            class="px-3 py-2 rounded bg-emerald-900/40 text-emerald-200 hover:bg-emerald-900/60 border border-emerald-800/60 text-sm font-semibold disabled:opacity-50"
            :disabled="actionBusy"
            @click="ratifyProposal"
          >
            ✓ Ratify (raw → researched)
          </button>
          <button
            v-if="wi.maturity !== 'raw' && wi.status === 'pending'"
            class="px-3 py-2 rounded bg-blue-900/40 text-blue-200 hover:bg-blue-900/60 border border-blue-800/60 text-sm font-semibold disabled:opacity-50"
            :disabled="actionBusy"
            @click="dispatchProposal"
          >
            ▶ Dispatch (run first stage)
          </button>
          <button
            class="ml-auto px-3 py-2 rounded bg-zinc-800 text-zinc-400 hover:bg-zinc-700 hover:text-zinc-200 border border-zinc-700 text-sm disabled:opacity-50"
            :disabled="actionBusy"
            @click="cancelProposal"
          >
            ✕ Cancel proposal
          </button>
        </div>

        <div v-if="actionMsg" class="text-xs text-emerald-300">{{ actionMsg }}</div>
        <div v-if="actionErr" class="text-xs text-red-400">{{ actionErr }}</div>

        <p class="text-xs text-zinc-500">
          Ratifying advances maturity but doesn't dispatch yet — you can adjust the
          pipeline_family_hint / project / inputs first if needed, then click Dispatch.
        </p>

        <!-- Edit + revise sub-panels (collapsible) -->
        <div class="border-t border-purple-800/40 pt-3 mt-3 space-y-2">
          <div class="flex flex-wrap gap-2">
            <button
              v-if="!editing"
              class="px-3 py-1.5 rounded text-xs border border-zinc-700 hover:bg-zinc-800 text-zinc-300"
              @click="startEdit"
            >
              ✎ Edit fields directly
            </button>
            <button
              v-if="!reviseOpen"
              class="px-3 py-1.5 rounded text-xs border border-purple-800/60 hover:bg-purple-900/30 text-purple-200"
              @click="openRevise"
            >
              ↻ Revise with AI feedback
            </button>
          </div>

          <!-- Direct edit form -->
          <div
            v-if="editing"
            class="rounded border border-zinc-700 bg-zinc-900/70 p-3 space-y-2"
          >
            <div class="text-xs uppercase tracking-wide text-zinc-500 mb-1">Edit fields</div>

            <label class="block text-xs text-zinc-400">slug
              <input v-model="editForm.slug"
                class="w-full mt-1 px-2 py-1 rounded bg-zinc-950 border border-zinc-700 text-sm font-mono text-zinc-200"
                placeholder="kebab-case-slug" />
            </label>

            <label class="block text-xs text-zinc-400">binding_question
              <textarea v-model="editForm.binding_question"
                class="w-full mt-1 px-2 py-1 rounded bg-zinc-950 border border-zinc-700 text-sm text-zinc-200"
                rows="3" />
            </label>

            <label class="block text-xs text-zinc-400">rationale
              <textarea v-model="editForm.rationale"
                class="w-full mt-1 px-2 py-1 rounded bg-zinc-950 border border-zinc-700 text-sm text-zinc-200"
                rows="2" />
            </label>

            <div class="grid grid-cols-2 gap-2">
              <label class="block text-xs text-zinc-400">pipeline_family
                <input v-model="editForm.pipeline_family_hint"
                  class="w-full mt-1 px-2 py-1 rounded bg-zinc-950 border border-zinc-700 text-sm font-mono text-zinc-200"
                  placeholder="research-write | planning | …" />
              </label>
              <label class="block text-xs text-zinc-400">project_association
                <input v-model="editForm.project_association"
                  class="w-full mt-1 px-2 py-1 rounded bg-zinc-950 border border-zinc-700 text-sm font-mono text-zinc-200"
                  placeholder="space-center | pg-ai-stewards | …" />
              </label>
            </div>

            <div class="flex gap-2 pt-1">
              <button
                class="px-3 py-1.5 rounded text-xs bg-emerald-900/40 text-emerald-200 hover:bg-emerald-900/60 border border-emerald-800/60 disabled:opacity-50"
                :disabled="editBusy"
                @click="submitEdit"
              >Save changes</button>
              <button
                class="px-3 py-1.5 rounded text-xs bg-zinc-800 text-zinc-400 hover:bg-zinc-700 border border-zinc-700"
                :disabled="editBusy"
                @click="cancelEdit"
              >Cancel</button>
            </div>
          </div>

          <!-- AI revise textarea -->
          <div
            v-if="reviseOpen"
            class="rounded border border-purple-800/60 bg-purple-950/30 p-3 space-y-2"
          >
            <div class="text-xs uppercase tracking-wide text-purple-300 mb-1">
              Revise with feedback
            </div>
            <p class="text-xs text-zinc-400">
              Write what should change. The agent reads the original + parent plan + your feedback
              and proposes a revision. You'll see a diff before it applies. ~$0.02-0.05 per revise.
            </p>
            <textarea v-model="reviseFeedback"
              class="w-full mt-1 px-2 py-1 rounded bg-zinc-950 border border-purple-800/40 text-sm text-zinc-200"
              rows="4"
              placeholder="e.g. 'scope tighter — focus only on the webcam permissions test, not the full ML pipeline'" />
            <div class="flex gap-2">
              <button
                class="px-3 py-1.5 rounded text-xs bg-purple-900/50 text-purple-100 hover:bg-purple-800/60 border border-purple-700/60 disabled:opacity-50"
                :disabled="reviseBusy || !reviseFeedback.trim()"
                @click="submitRevise"
              >
                {{ reviseBusy ? 'Dispatching…' : '↻ Send to agent' }}
              </button>
              <button
                class="px-3 py-1.5 rounded text-xs bg-zinc-800 text-zinc-400 hover:bg-zinc-700 border border-zinc-700"
                :disabled="reviseBusy"
                @click="cancelRevise"
              >Cancel</button>
            </div>
          </div>
        </div>

        <!-- Pending revisions diff cards -->
        <div
          v-if="pendingRevisions.length > 0"
          class="border-t border-purple-800/40 pt-3 mt-3 space-y-3"
        >
          <div class="text-xs uppercase tracking-wide text-purple-300">
            Pending revisions ({{ pendingRevisions.length }})
          </div>

          <div
            v-for="rev in pendingRevisions"
            :key="rev.id"
            class="rounded border border-purple-800/40 bg-zinc-950/60 p-3 space-y-2"
          >
            <div class="flex items-baseline justify-between text-xs">
              <div class="text-purple-300 font-mono">{{ rev.slug }}</div>
              <div class="text-zinc-500">
                <span v-if="rev.status === 'in_progress' || rev.status === 'pending'"
                      class="text-amber-400">⏳ {{ rev.status }} ({{ rev.maturity }})</span>
                <span v-else>{{ rev.status }} · ${{ (rev.cost_micro / 1_000_000).toFixed(4) }}</span>
              </div>
            </div>

            <div v-if="rev.feedback" class="text-xs text-zinc-400 italic">
              <span class="text-zinc-500">your feedback:</span> "{{ rev.feedback }}"
            </div>

            <!-- Diff rows once revision_json is available -->
            <template v-if="rev.revision_json && rev.status === 'completed'">
              <div
                v-for="row in diffRows(rev)"
                :key="row.label"
                class="grid grid-cols-2 gap-2 text-xs"
              >
                <div class="rounded bg-red-950/30 border border-red-900/30 p-2">
                  <div class="text-red-400 uppercase tracking-wide text-[10px] mb-1">before · {{ row.label }}</div>
                  <pre class="whitespace-pre-wrap font-sans text-zinc-300">{{ row.before }}</pre>
                </div>
                <div class="rounded bg-emerald-950/30 border border-emerald-900/30 p-2">
                  <div class="text-emerald-400 uppercase tracking-wide text-[10px] mb-1">after · {{ row.label }}</div>
                  <pre class="whitespace-pre-wrap font-sans text-zinc-200">{{ row.after }}</pre>
                </div>
              </div>

              <div class="flex gap-2 pt-1">
                <button
                  class="px-3 py-1.5 rounded text-xs bg-emerald-900/40 text-emerald-200 hover:bg-emerald-900/60 border border-emerald-800/60"
                  @click="acceptRevision(rev.id)"
                >✓ Accept revision</button>
                <button
                  class="px-3 py-1.5 rounded text-xs bg-zinc-800 text-zinc-400 hover:bg-zinc-700 border border-zinc-700"
                  @click="rejectRevision(rev.id)"
                >✕ Reject</button>
              </div>
            </template>

            <div v-else class="text-xs text-zinc-500 italic">
              Revising… polling every 4s for completion.
            </div>
          </div>
        </div>
      </section>

      <!-- Phase 5a (Phase B): Maturity ladder panel -->
      <section class="rounded-md border border-zinc-800 bg-zinc-900/50 p-4">
        <div class="flex items-baseline justify-between mb-3">
          <div class="text-xs uppercase tracking-wide text-zinc-500">Maturity</div>
          <div class="text-xs text-zinc-500 font-mono">
            <span v-if="wi.destination_maturity">ceiling: {{ wi.destination_maturity }}</span>
            <span v-else>full Ammon-loop → verified</span>
            <span v-if="wi.revision_count > 0" class="ml-3 text-amber-400">
              revisions: {{ wi.revision_count }} / 2
            </span>
          </div>
        </div>
        <ol class="flex items-center gap-1 text-xs font-mono">
          <li
            v-for="(m, i) in MATURITY_LADDER"
            :key="m"
            class="flex items-center gap-1"
          >
            <span
              class="px-2 py-1 rounded border"
              :class="[
                maturityIndex(wi.maturity) === i
                  ? 'bg-emerald-900/40 text-emerald-200 border-emerald-700/60 font-semibold'
                  : maturityIndex(wi.maturity) > i
                    ? 'bg-zinc-800/50 text-zinc-500 border-zinc-700'
                    : 'bg-zinc-900 text-zinc-600 border-zinc-800',
                wi.destination_maturity === m
                  ? 'ring-1 ring-blue-500/50'
                  : '',
              ]"
            >{{ m }}</span>
            <span
              v-if="i < MATURITY_LADDER.length - 1"
              class="text-zinc-700"
            >→</span>
          </li>
        </ol>
        <p v-if="wi.destination_maturity" class="text-xs text-zinc-500 mt-2">
          Substrate will surface for review when maturity reaches
          <span class="font-mono text-blue-300">{{ wi.destination_maturity }}</span>.
        </p>
      </section>

      <!-- Batch G.4.4 — File destination panel -->
      <section class="rounded-md border border-zinc-800 bg-zinc-900/50 p-4 space-y-3">
        <div class="flex items-baseline justify-between">
          <div class="text-xs uppercase tracking-wide text-zinc-500">File destination</div>
          <span
            v-if="wi.file_enqueued_at"
            class="text-xs text-emerald-500 font-mono"
          >✓ queued {{ fmtDate(wi.file_enqueued_at) }}</span>
        </div>

        <div v-if="!editingFileDestination">
          <div v-if="wi.file_destination" class="text-sm font-mono text-zinc-200">
            {{ wi.file_destination }}
          </div>
          <div v-else class="text-sm text-zinc-500 italic">
            DB-only (no file write)
          </div>
          <div class="flex items-center gap-2 mt-2">
            <button
              class="text-xs px-3 py-1 rounded border border-zinc-700 hover:bg-zinc-800 text-zinc-300"
              @click="startEditFileDestination"
            >{{ wi.file_destination ? 'Edit' : 'Set destination' }}</button>
            <button
              v-if="wi.file_destination && !wi.file_enqueued_at"
              class="text-xs px-3 py-1 rounded bg-emerald-700 hover:bg-emerald-600 text-white disabled:opacity-50"
              :disabled="materializingFile"
              @click="materializeFile"
            >{{ materializingFile ? 'queueing…' : 'Materialize now' }}</button>
            <span v-if="materializeMsg" class="text-xs text-zinc-400">{{ materializeMsg }}</span>
          </div>
        </div>

        <div v-else class="space-y-2">
          <input
            v-model="editFileDestinationValue"
            type="text"
            placeholder="path (or blank for DB-only)"
            class="w-full px-3 py-2 rounded border border-zinc-700 bg-zinc-950 text-sm font-mono focus:border-zinc-500 focus:outline-none"
          />
          <div class="flex items-center gap-2 text-xs">
            <button
              class="px-3 py-1 rounded bg-emerald-700 hover:bg-emerald-600 text-white disabled:opacity-50"
              :disabled="savingFileDestination"
              @click="saveFileDestination"
            >{{ savingFileDestination ? 'saving…' : 'Save' }}</button>
            <button
              class="px-3 py-1 rounded border border-zinc-700 hover:bg-zinc-800 text-zinc-300"
              @click="editingFileDestination = false"
            >Cancel</button>
            <button
              v-if="wi.pipeline_file_template"
              class="px-3 py-1 rounded border border-zinc-700 hover:bg-zinc-800 text-zinc-300"
              @click="useTemplateForFileDestination"
              :title="wi.pipeline_file_template"
            >Use pipeline default</button>
            <span v-if="fileDestinationError" class="text-red-400">{{ fileDestinationError }}</span>
          </div>
          <p class="text-xs text-zinc-500">
            Empty = DB-only. <code class="font-mono">&lt;slug&gt;</code> in the path renders to the work_item slug.
          </p>
        </div>
      </section>

      <!-- Phase 5a (Phase B): Scenarios panel — only shown if any -->
      <section
        v-if="scenariosArray(wi.scenarios).length > 0"
        class="rounded-md border border-zinc-800 bg-zinc-900/50 p-4"
      >
        <div class="text-xs uppercase tracking-wide text-zinc-500 mb-2">
          Scenarios <span class="text-zinc-600">— acceptance criteria</span>
        </div>
        <ul class="text-sm text-zinc-200 space-y-1 list-disc list-inside">
          <li v-for="(s, i) in scenariosArray(wi.scenarios)" :key="i">{{ s }}</li>
        </ul>
      </section>

      <!-- Phase 5a (Phase B): Spec panel — only shown if non-empty -->
      <section
        v-if="wi.spec"
        class="rounded-md border border-zinc-800 bg-zinc-900/50 p-4"
      >
        <div class="text-xs uppercase tracking-wide text-zinc-500 mb-2">Spec</div>
        <pre class="text-xs font-mono text-zinc-300 whitespace-pre-wrap overflow-auto max-h-96">{{ wi.spec }}</pre>
      </section>

      <!-- Phase 5a (Phase B): Gate decisions audit -->
      <section
        v-if="gateDecisions && gateDecisions.count > 0"
        class="rounded-md border border-zinc-800 bg-zinc-900/50 overflow-hidden"
      >
        <div class="px-4 py-3 border-b border-zinc-800 flex items-baseline justify-between">
          <h3 class="text-sm font-semibold">Gate decisions ({{ gateDecisions.count }})</h3>
          <span class="text-xs text-zinc-500">most recent first</span>
        </div>
        <ul class="divide-y divide-zinc-800/50">
          <li v-for="g in gateDecisions.items" :key="g.id" class="px-4 py-3 text-xs">
            <div class="flex items-baseline gap-3 flex-wrap">
              <span
                class="px-2 py-0.5 rounded border font-mono uppercase tracking-wide"
                :class="gateActionTone(g.action)"
              >{{ g.action }}</span>
              <span class="text-zinc-500 font-mono">from: {{ g.from_maturity }}</span>
              <span v-if="g.revision_count > 0" class="text-amber-400 font-mono">rev #{{ g.revision_count }}</span>
              <span v-if="g.work_id" class="text-zinc-500 font-mono">work_id: {{ g.work_id }}</span>
              <span class="ml-auto text-zinc-500 tabular-nums">{{ fmtDate(g.at) }}</span>
            </div>
            <div v-if="g.reasoning" class="mt-2 text-zinc-300 leading-relaxed">{{ g.reasoning }}</div>
            <details v-if="g.feedback" class="mt-2">
              <summary class="cursor-pointer text-zinc-500 hover:text-zinc-300">feedback</summary>
              <pre class="mt-1 font-mono text-zinc-400 whitespace-pre-wrap">{{ g.feedback }}</pre>
            </details>
            <details v-if="g.raw_response && Object.keys(g.raw_response as object).length > 0" class="mt-1">
              <summary class="cursor-pointer text-zinc-600 hover:text-zinc-400">raw response</summary>
              <pre class="mt-1 font-mono text-zinc-400 whitespace-pre-wrap">{{ fmtJson(g.raw_response) }}</pre>
            </details>
            <button
              class="mt-2 text-xs px-2 py-1 rounded border border-purple-700 hover:bg-purple-900/30 text-purple-300"
              @click="openOverride(g)"
            >Override gate decision…</button>
          </li>
        </ul>
      </section>

      <!-- Phase 5f (E.7): override modal -->
      <div
        v-if="overrideOpen"
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/60"
        @click.self="overrideOpen = null"
      >
        <div class="bg-zinc-900 border border-zinc-700 rounded-lg p-5 max-w-lg w-full space-y-3">
          <h3 class="text-lg font-semibold">Override gate decision</h3>
          <p class="text-xs text-zinc-500">
            Original action: <span class="font-mono">{{ overrideOpen.current_action }}</span>.
            Override counts as a failure for trust scoring (D-E3) — the cell auto-demotes one level.
          </p>
          <div>
            <label class="block text-xs uppercase tracking-wide text-zinc-500 mb-1">New action</label>
            <select
              v-model="overrideNewAction"
              class="w-full px-3 py-2 rounded border border-zinc-700 bg-zinc-950 text-sm focus:border-zinc-500 focus:outline-none"
            >
              <option value="advance">advance</option>
              <option value="revise">revise</option>
              <option value="surface">surface</option>
            </select>
          </div>
          <div>
            <label class="block text-xs uppercase tracking-wide text-zinc-500 mb-1">Overridden by</label>
            <input
              v-model="overrideBy"
              type="text"
              class="w-full px-3 py-2 rounded border border-zinc-700 bg-zinc-950 text-sm focus:border-zinc-500 focus:outline-none"
            />
          </div>
          <div>
            <label class="block text-xs uppercase tracking-wide text-zinc-500 mb-1">
              Justification <span class="text-red-400">*</span>
              <span class="text-zinc-600 normal-case">(at least 10 chars)</span>
            </label>
            <textarea
              v-model="overrideJustification"
              rows="3"
              placeholder="why is the gate's call wrong?"
              class="w-full px-3 py-2 rounded border border-zinc-700 bg-zinc-950 text-sm focus:border-zinc-500 focus:outline-none resize-y"
            ></textarea>
          </div>
          <div class="flex items-center gap-3 pt-2">
            <button
              class="px-4 py-2 rounded bg-purple-700 hover:bg-purple-600 text-white text-sm font-medium disabled:opacity-50"
              :disabled="overriding || overrideJustification.trim().length < 10"
              @click="submitOverride"
            >{{ overriding ? 'submitting…' : 'Apply override' }}</button>
            <button
              class="px-3 py-2 text-xs rounded border border-zinc-700 hover:bg-zinc-800 text-zinc-300"
              @click="overrideOpen = null"
            >Cancel</button>
            <span v-if="overrideError" class="text-xs text-red-400">{{ overrideError }}</span>
          </div>
        </div>
      </div>

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
