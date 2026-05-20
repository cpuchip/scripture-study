<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { scheduledApi, type ScheduledRow, type ScheduledCreateReq, type ScheduledUpdateReq } from '@/api'

const rows = ref<ScheduledRow[]>([])
const loading = ref(true)
const error = ref('')

// Edit modal state
const editing = ref<ScheduledRow | null>(null)
const editCron = ref('')
const editInputTemplateText = ref('')
const editMissedWindow = ref(24)
const editNotes = ref('')
const editError = ref('')

// Create modal state
const creating = ref(false)
const newSlug = ref('')
const newPipelineFamily = ref('research-summary')
const newIntentSlug = ref('general-research')
const newCron = ref('0 13 * * 1-5')
const newInputTemplateText = ref('{}')
const newMissedWindow = ref(24)
const newNotes = ref('')
const createError = ref('')

const now = ref(new Date())
setInterval(() => { now.value = new Date() }, 1000)

async function load() {
  loading.value = true
  error.value = ''
  try {
    const r = await scheduledApi.list()
    rows.value = r.items
  } catch (e) {
    error.value = String(e)
  } finally {
    loading.value = false
  }
}

async function onToggle(row: ScheduledRow) {
  try {
    const r = await scheduledApi.toggle(row.id)
    row.enabled = r.enabled
  } catch (e) {
    error.value = String(e)
  }
}

async function onDelete(row: ScheduledRow) {
  if (!confirm(`Delete schedule "${row.slug}"? Existing work_items it spawned are preserved.`)) return
  try {
    await scheduledApi.remove(row.id)
    rows.value = rows.value.filter(r => r.id !== row.id)
  } catch (e) {
    error.value = String(e)
  }
}

function openEdit(row: ScheduledRow) {
  editing.value = row
  editCron.value = row.cron_pattern
  editInputTemplateText.value = JSON.stringify(row.input_template, null, 2)
  editMissedWindow.value = row.missed_window_hours
  editNotes.value = row.notes ?? ''
  editError.value = ''
}

function closeEdit() {
  editing.value = null
  editError.value = ''
}

async function saveEdit() {
  if (!editing.value) return
  editError.value = ''
  let parsedInput: Record<string, unknown>
  try {
    parsedInput = JSON.parse(editInputTemplateText.value)
  } catch (e) {
    editError.value = `input_template is not valid JSON: ${e}`
    return
  }
  const req: ScheduledUpdateReq = {
    cron_pattern: editCron.value.trim(),
    input_template: parsedInput,
    missed_window_hours: editMissedWindow.value,
    notes: editNotes.value,
  }
  try {
    const updated = await scheduledApi.update(editing.value.id, req)
    const idx = rows.value.findIndex(r => r.id === updated.id)
    if (idx >= 0) rows.value[idx] = updated
    closeEdit()
  } catch (e) {
    editError.value = String(e)
  }
}

function openCreate() {
  creating.value = true
  newSlug.value = ''
  newPipelineFamily.value = 'research-summary'
  newIntentSlug.value = 'general-research'
  newCron.value = '0 13 * * 1-5'
  newInputTemplateText.value = JSON.stringify({
    binding_question: 'What shipped in AI today that I should know about?',
    sources_spec: {
      queries: ['AI news today', 'claude release', 'openai update'],
      max_per_query: 10,
      since: '24h',
    },
    output_kind: 'daily-digest',
  }, null, 2)
  newMissedWindow.value = 24
  newNotes.value = ''
  createError.value = ''
}

function closeCreate() {
  creating.value = false
  createError.value = ''
}

async function saveCreate() {
  createError.value = ''
  let parsedInput: Record<string, unknown>
  try {
    parsedInput = JSON.parse(newInputTemplateText.value)
  } catch (e) {
    createError.value = `input_template is not valid JSON: ${e}`
    return
  }
  if (!newSlug.value.trim()) { createError.value = 'slug required'; return }
  const req: ScheduledCreateReq = {
    slug: newSlug.value.trim(),
    pipeline_family: newPipelineFamily.value,
    intent_slug: newIntentSlug.value,
    cron_pattern: newCron.value.trim(),
    input_template: parsedInput,
    missed_window_hours: newMissedWindow.value,
    notes: newNotes.value,
  }
  try {
    await scheduledApi.create(req)
    closeCreate()
    await load()
  } catch (e) {
    createError.value = String(e)
  }
}

function fmtTime(t?: string) {
  if (!t) return '—'
  const d = new Date(t)
  if (isNaN(d.getTime())) return t
  return d.toLocaleString()
}

function timeUntil(t?: string): string {
  if (!t) return '—'
  const target = new Date(t).getTime()
  const diff = target - now.value.getTime()
  if (diff <= 0) return 'due now'
  const m = Math.floor(diff / 60000)
  if (m < 60) return `in ${m}m`
  const h = Math.floor(m / 60)
  if (h < 24) return `in ${h}h${m % 60 ? ` ${m % 60}m` : ''}`
  const d = Math.floor(h / 24)
  return `in ${d}d${h % 24 ? ` ${h % 24}h` : ''}`
}

const sortedRows = computed(() =>
  [...rows.value].sort((a, b) => {
    if (a.enabled !== b.enabled) return a.enabled ? -1 : 1
    return (a.next_due_at ?? '').localeCompare(b.next_due_at ?? '')
  })
)

onMounted(load)
</script>

<template>
  <div class="space-y-6 max-w-5xl">
    <header class="border-b border-zinc-800 pb-3 flex items-center justify-between">
      <div>
        <h2 class="text-2xl font-semibold tracking-tight">Scheduled pipelines</h2>
        <p class="text-sm text-zinc-400 mt-1">
          Cron-style dispatches. Each row fires a new <code class="font-mono text-zinc-300">work_item</code>
          of its pipeline when <code class="font-mono text-zinc-300">next_due_at</code> is reached.
          The 60s bgworker tick wraps both watchman and scheduled-pipelines schedulers.
        </p>
      </div>
      <button
        class="px-3 py-1.5 rounded border border-zinc-700 hover:bg-zinc-800 text-zinc-100 text-sm"
        @click="openCreate"
      >+ New schedule</button>
    </header>

    <p v-if="loading" class="text-sm text-zinc-400">loading…</p>
    <p v-else-if="error" class="text-sm text-red-400">{{ error }}</p>

    <ul v-else-if="sortedRows.length" class="space-y-3">
      <li
        v-for="row in sortedRows"
        :key="row.id"
        class="rounded-md border border-zinc-800 bg-zinc-900/50 p-4"
        :class="{ 'opacity-60': !row.enabled }"
      >
        <div class="flex items-baseline justify-between gap-3">
          <div class="flex-1">
            <h3 class="text-lg font-semibold">
              {{ row.slug }}
              <span class="ml-2 text-xs font-normal text-zinc-500">{{ row.pipeline_family }}</span>
            </h3>
            <p class="text-sm text-zinc-300 mt-1 font-mono">
              {{ row.cron_pattern }}
              <span class="text-zinc-500 font-sans ml-2">(intent: {{ row.intent_slug }})</span>
            </p>
          </div>
          <div class="flex items-center gap-2">
            <label class="text-xs flex items-center gap-1.5 cursor-pointer">
              <input type="checkbox" :checked="row.enabled" @change="onToggle(row)" />
              <span :class="row.enabled ? 'text-emerald-400' : 'text-zinc-500'">
                {{ row.enabled ? 'enabled' : 'disabled' }}
              </span>
            </label>
            <button
              class="px-2 py-1 rounded border border-zinc-700 hover:bg-zinc-800 text-zinc-300 text-xs"
              @click="openEdit(row)"
            >edit</button>
            <button
              class="px-2 py-1 rounded border border-red-900 hover:bg-red-950 text-red-300 text-xs"
              @click="onDelete(row)"
            >delete</button>
          </div>
        </div>

        <div class="mt-3 grid grid-cols-3 gap-3 text-xs">
          <div>
            <div class="text-zinc-500 uppercase tracking-wide">Next due</div>
            <div class="text-zinc-300 mt-0.5">
              {{ fmtTime(row.next_due_at) }}
              <span class="text-zinc-500 ml-1">({{ timeUntil(row.next_due_at) }})</span>
            </div>
          </div>
          <div>
            <div class="text-zinc-500 uppercase tracking-wide">Last dispatched</div>
            <div class="text-zinc-300 mt-0.5">{{ fmtTime(row.last_dispatched_at) }}</div>
          </div>
          <div>
            <div class="text-zinc-500 uppercase tracking-wide">Missed window</div>
            <div class="text-zinc-300 mt-0.5">{{ row.missed_window_hours }}h</div>
          </div>
        </div>

        <div v-if="row.notes" class="mt-3 text-xs text-zinc-400 italic">{{ row.notes }}</div>
      </li>
    </ul>

    <div v-else class="text-sm text-zinc-500">
      No schedules yet. Click <em>+ New schedule</em> to add one.
    </div>

    <!-- Edit modal -->
    <div
      v-if="editing"
      class="fixed inset-0 bg-black/70 flex items-start justify-center pt-12 px-4 z-50"
      @click.self="closeEdit"
    >
      <div class="bg-zinc-950 border border-zinc-800 rounded-md p-5 max-w-2xl w-full space-y-3">
        <h3 class="text-lg font-semibold">Edit schedule — {{ editing.slug }}</h3>

        <div>
          <label class="block text-xs uppercase tracking-wide text-zinc-500 mb-1">Cron pattern</label>
          <input
            v-model="editCron"
            class="w-full bg-zinc-900 border border-zinc-700 rounded px-2 py-1.5 text-sm font-mono"
            placeholder="0 13 * * 1-5"
          />
          <p class="text-xs text-zinc-500 mt-1">Standard 5-field cron (UTC). Supports ranges (1-5), lists (1,3,5), step values (*/15).</p>
        </div>

        <div>
          <label class="block text-xs uppercase tracking-wide text-zinc-500 mb-1">Input template (JSON)</label>
          <textarea
            v-model="editInputTemplateText"
            rows="10"
            class="w-full bg-zinc-900 border border-zinc-700 rounded px-2 py-1.5 text-xs font-mono"
          ></textarea>
        </div>

        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="block text-xs uppercase tracking-wide text-zinc-500 mb-1">Missed window (hours)</label>
            <input
              v-model.number="editMissedWindow"
              type="number"
              min="0"
              class="w-full bg-zinc-900 border border-zinc-700 rounded px-2 py-1.5 text-sm"
            />
          </div>
          <div>
            <label class="block text-xs uppercase tracking-wide text-zinc-500 mb-1">Notes</label>
            <input
              v-model="editNotes"
              class="w-full bg-zinc-900 border border-zinc-700 rounded px-2 py-1.5 text-sm"
            />
          </div>
        </div>

        <p v-if="editError" class="text-sm text-red-400">{{ editError }}</p>

        <div class="flex justify-end gap-2 pt-2">
          <button class="px-3 py-1.5 rounded border border-zinc-700 hover:bg-zinc-800 text-zinc-300 text-sm" @click="closeEdit">Cancel</button>
          <button class="px-3 py-1.5 rounded border border-emerald-700 hover:bg-emerald-950 text-emerald-200 text-sm" @click="saveEdit">Save</button>
        </div>
      </div>
    </div>

    <!-- Create modal -->
    <div
      v-if="creating"
      class="fixed inset-0 bg-black/70 flex items-start justify-center pt-12 px-4 z-50"
      @click.self="closeCreate"
    >
      <div class="bg-zinc-950 border border-zinc-800 rounded-md p-5 max-w-2xl w-full space-y-3">
        <h3 class="text-lg font-semibold">New schedule</h3>

        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="block text-xs uppercase tracking-wide text-zinc-500 mb-1">Slug</label>
            <input
              v-model="newSlug"
              class="w-full bg-zinc-900 border border-zinc-700 rounded px-2 py-1.5 text-sm font-mono"
              placeholder="ai-news-7am"
            />
          </div>
          <div>
            <label class="block text-xs uppercase tracking-wide text-zinc-500 mb-1">Pipeline family</label>
            <select v-model="newPipelineFamily" class="w-full bg-zinc-900 border border-zinc-700 rounded px-2 py-1.5 text-sm">
              <option value="research-summary">research-summary</option>
              <option value="research-write">research-write</option>
              <option value="yt-gospel-evaluate">yt-gospel-evaluate</option>
              <option value="yt-secular-digest">yt-secular-digest</option>
              <option value="study-write">study-write</option>
              <option value="study-write-qwen">study-write-qwen</option>
              <option value="echo-test">echo-test</option>
            </select>
          </div>
        </div>

        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="block text-xs uppercase tracking-wide text-zinc-500 mb-1">Intent slug</label>
            <select v-model="newIntentSlug" class="w-full bg-zinc-900 border border-zinc-700 rounded px-2 py-1.5 text-sm">
              <option value="general-research">general-research</option>
              <option value="scripture-study">scripture-study</option>
              <option value="planning-partner">planning-partner</option>
            </select>
          </div>
          <div>
            <label class="block text-xs uppercase tracking-wide text-zinc-500 mb-1">Cron pattern</label>
            <input
              v-model="newCron"
              class="w-full bg-zinc-900 border border-zinc-700 rounded px-2 py-1.5 text-sm font-mono"
              placeholder="0 13 * * 1-5"
            />
          </div>
        </div>

        <div>
          <label class="block text-xs uppercase tracking-wide text-zinc-500 mb-1">Input template (JSON)</label>
          <textarea
            v-model="newInputTemplateText"
            rows="10"
            class="w-full bg-zinc-900 border border-zinc-700 rounded px-2 py-1.5 text-xs font-mono"
          ></textarea>
        </div>

        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="block text-xs uppercase tracking-wide text-zinc-500 mb-1">Missed window (hours)</label>
            <input v-model.number="newMissedWindow" type="number" min="0" class="w-full bg-zinc-900 border border-zinc-700 rounded px-2 py-1.5 text-sm" />
          </div>
          <div>
            <label class="block text-xs uppercase tracking-wide text-zinc-500 mb-1">Notes</label>
            <input v-model="newNotes" class="w-full bg-zinc-900 border border-zinc-700 rounded px-2 py-1.5 text-sm" />
          </div>
        </div>

        <p v-if="createError" class="text-sm text-red-400">{{ createError }}</p>

        <div class="flex justify-end gap-2 pt-2">
          <button class="px-3 py-1.5 rounded border border-zinc-700 hover:bg-zinc-800 text-zinc-300 text-sm" @click="closeCreate">Cancel</button>
          <button class="px-3 py-1.5 rounded border border-emerald-700 hover:bg-emerald-950 text-emerald-200 text-sm" @click="saveCreate">Create</button>
        </div>
      </div>
    </div>
  </div>
</template>
