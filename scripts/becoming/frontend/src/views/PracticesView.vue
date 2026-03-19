<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { api, type Practice, type PillarLink } from '../api'
import { useNotifications } from '../composables/useNotifications'

const { subscribed: notifSubscribed } = useNotifications()

function localDateStr(d: Date = new Date()): string {
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

const router = useRouter()
const route = useRoute()
const practices = ref<Practice[]>([])
const loading = ref(true)
const showForm = ref(false)
const editingId = ref<number | null>(null)

// Pillars
const allPillars = ref<{ id: number; name: string; icon: string }[]>([])
const formPillarIds = ref<number[]>([])

// Filters
const filterStatus = ref<string>('active')
const filterTime = ref<'all' | 'current' | 'upcoming' | 'past'>('all')
const timeFilterOptions = ['all', 'current', 'upcoming', 'past'] as const

// Tri-state filter maps: value → 'positive' | 'negative' | null
// positive = include only matching, negative = exclude matching
type TriState = 'positive' | 'negative' | null
const filterTypeState = ref<Map<string, TriState>>(new Map())
const filterCatState = ref<Map<string, TriState>>(new Map())
const filterPillarState = ref<Map<string, TriState>>(new Map())

const filterRefs = {
  type: filterTypeState,
  cat: filterCatState,
  pillar: filterPillarState,
} as const

function cycleFilter(which: 'type' | 'cat' | 'pillar', key: string, isAll = false) {
  const stateMap = filterRefs[which]
  const current = stateMap.value.get(key) || null
  const next: TriState = current === null ? 'positive' : current === 'positive' ? 'negative' : null

  if (isAll) {
    if (next === 'positive') {
      stateMap.value = new Map()
    } else {
      stateMap.value = new Map([['all', next]])
    }
  } else {
    if (stateMap.value.get('all') === 'negative' && next === 'positive') {
      stateMap.value.delete('all')
    }
    if (next === null) {
      stateMap.value.delete(key)
    } else {
      stateMap.value.set(key, next)
    }
  }
  stateMap.value = new Map(stateMap.value)
}

function filterState(stateMap: Map<string, TriState>, key: string): TriState {
  return stateMap.get(key) || null
}

function filterChipClass(state: TriState): string {
  if (state === 'positive') return 'bg-indigo-100 border-indigo-300 text-indigo-700'
  if (state === 'negative') return 'bg-red-50 border-red-300 text-red-400 line-through'
  return 'bg-gray-50 border-gray-200 text-gray-500 hover:bg-gray-100'
}

function allChipClass(stateMap: Map<string, TriState>): string {
  const allState = stateMap.get('all') || null
  if (allState === 'negative') return 'bg-red-50 border-red-300 text-red-400 line-through'
  if (stateMap.size === 0) return 'bg-indigo-100 border-indigo-300 text-indigo-700'
  return 'bg-gray-50 border-gray-200 text-gray-500 hover:bg-gray-100'
}

// Practice → pillar mappings (key: practice_id, value: { id, icon }[])
const practicePillarMap = ref<Map<number, { id: number; icon: string; name: string }[]>>(new Map())

// Form state
const form = ref({
  name: '',
  description: '',
  type: 'habit' as Practice['type'],
  category: '',
  config: '{}',
  end_date: '' as string,
  start_date: '' as string,
})

// Config helpers for tracker (was exercise)
const trackerConfig = ref({
  target_sets: 2,
  target_reps: 15,
  unit: 'reps',
})

// Config helpers for memorize daily reps
const memorizeConfig = ref({
  target_daily_reps: 1,
})

// Config helpers for scheduled practices
const scheduleConfig = ref({
  type: 'interval' as 'interval' | 'daily_slots' | 'weekly' | 'monthly' | 'once',
  interval_days: 2,
  shift_on_early: true,
  slots: ['morning', 'lunch', 'night'] as string[],
  days: [] as string[],
  day_of_month: 1,
  due_date: '',
  newSlot: '',
})

const presetCategories = ['spiritual', 'scripture', 'pt', 'fitness', 'study', 'health']

// Derived filter values from data
const availableTypes = computed(() => {
  const types = new Set(practices.value.map(p => p.type))
  return Array.from(types).sort()
})
const availableCategories = computed(() => {
  const cats = new Set<string>()
  for (const p of practices.value) {
    if (p.category) {
      for (const c of p.category.split(',')) {
        const trimmed = c.trim()
        if (trimmed) cats.add(trimmed)
      }
    }
  }
  return Array.from(cats).sort()
})

const filteredPractices = computed(() => {
  const todayStr = localDateStr()

  // Collect positive and negative sets for each filter dimension
  const typePos = new Set<string>()
  const typeNeg = new Set<string>()
  const catPos = new Set<string>()
  const catNeg = new Set<string>()
  const pillarPos = new Set<number>()
  const pillarNeg = new Set<number>()
  const allTypeNeg = filterTypeState.value.get('all') === 'negative'
  const allCatNeg = filterCatState.value.get('all') === 'negative'
  const allPillarNeg = filterPillarState.value.get('all') === 'negative'

  for (const [k, v] of filterTypeState.value) {
    if (k === 'all') continue
    if (v === 'positive') typePos.add(k)
    if (v === 'negative') typeNeg.add(k)
  }
  for (const [k, v] of filterCatState.value) {
    if (k === 'all') continue
    if (v === 'positive') catPos.add(k)
    if (v === 'negative') catNeg.add(k)
  }
  for (const [k, v] of filterPillarState.value) {
    if (k === 'all') continue
    if (v === 'positive') pillarPos.add(Number(k))
    if (v === 'negative') pillarNeg.add(Number(k))
  }

  return practices.value.filter(p => {
    // Type filter
    if (allTypeNeg) {
      // "all" negative on type: show only practices whose type is NOT in the known types
      // (this would show none since every practice has a type — effectively just excludes everything unless a positive overrides)
      if (typePos.size > 0) {
        if (!typePos.has(p.type)) return false
      } else {
        return false // "all" negative with no positives = show nothing
      }
    } else {
      if (typePos.size > 0 && !typePos.has(p.type)) return false
      if (typeNeg.has(p.type)) return false
    }

    // Category filter
    const cats = (p.category || '').split(',').map(c => c.trim()).filter(Boolean)
    if (allCatNeg) {
      // "all" negative on category: show only practices with NO category
      if (catPos.size > 0) {
        if (!cats.some(c => catPos.has(c))) return false
      } else {
        if (cats.length > 0) return false
      }
    } else {
      if (catPos.size > 0 && !cats.some(c => catPos.has(c))) return false
      if (cats.some(c => catNeg.has(c))) return false
    }

    // Pillar filter
    const pids = (practicePillarMap.value.get(p.id) || []).map(x => x.id)
    if (allPillarNeg) {
      // "all" negative on pillar: show only practices with NO pillar
      if (pillarPos.size > 0) {
        if (!pids.some(id => pillarPos.has(id))) return false
      } else {
        if (pids.length > 0) return false
      }
    } else {
      if (pillarPos.size > 0 && !pids.some(id => pillarPos.has(id))) return false
      if (pids.some(id => pillarNeg.has(id))) return false
    }

    // Time filter
    if (filterTime.value !== 'all') {
      const startDate = p.start_date ? p.start_date.slice(0, 10) : p.created_at?.slice(0, 10) || ''
      const endDate = p.end_date ? p.end_date.slice(0, 10) : ''
      if (filterTime.value === 'upcoming') {
        if (!startDate || startDate <= todayStr) return false
      } else if (filterTime.value === 'current') {
        if (startDate > todayStr) return false
        if (endDate && endDate < todayStr) return false
      } else if (filterTime.value === 'past') {
        if (!endDate || endDate >= todayStr) return false
      }
    }
    return true
  })
})

// Multi-category helpers
function categoryList(): string[] {
  return form.value.category.split(',').map(c => c.trim()).filter(Boolean)
}

function toggleCategory(cat: string) {
  const cats = categoryList()
  const idx = cats.indexOf(cat)
  if (idx >= 0) {
    cats.splice(idx, 1)
  } else {
    cats.push(cat)
  }
  form.value.category = cats.join(',')
}

// End date helpers
function parseEndDate(endDate: string): Date {
  // Backend may return "2026-02-20" or "2026-02-20T00:00:00Z"
  const dateStr = endDate.slice(0, 10)
  return new Date(dateStr + 'T00:00:00')
}

function endDateLabel(endDate: string): string {
  const end = parseEndDate(endDate)
  const now = new Date()
  now.setHours(0, 0, 0, 0)
  const diff = Math.ceil((end.getTime() - now.getTime()) / (1000 * 60 * 60 * 24))
  if (diff < 0) return `${Math.abs(diff)}d overdue`
  if (diff === 0) return 'due today'
  if (diff === 1) return '1 day left'
  return `${diff} days left`
}

function endDateClass(endDate: string): string {
  const end = parseEndDate(endDate)
  const now = new Date()
  now.setHours(0, 0, 0, 0)
  const diff = Math.ceil((end.getTime() - now.getTime()) / (1000 * 60 * 60 * 24))
  if (diff < 0) return 'bg-red-100 text-red-700'
  if (diff <= 3) return 'bg-amber-100 text-amber-700'
  return 'bg-green-50 text-green-700'
}

function endDateTooltip(endDate: string): string {
  const d = parseEndDate(endDate)
  return d.toLocaleDateString(undefined, { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric' })
}

// Start date helpers
function startDateLabel(startDate: string): string {
  const start = parseEndDate(startDate) // reuse date parser
  const now = new Date()
  now.setHours(0, 0, 0, 0)
  const diff = Math.ceil((start.getTime() - now.getTime()) / (1000 * 60 * 60 * 24))
  if (diff <= 0) return '' // already started
  if (diff === 1) return 'starts tomorrow'
  return `starts in ${diff}d`
}

function startDateClass(): string {
  return 'bg-blue-50 text-blue-600'
}

async function load() {
  loading.value = true
  const [practicesData, pillarsData, pillarLinks] = await Promise.all([
    api.listPractices(undefined, true, filterStatus.value !== 'all' ? filterStatus.value : undefined),
    api.listPillarsFlat(),
    api.getAllPracticePillarLinks(),
  ])
  practices.value = practicesData
  allPillars.value = pillarsData.map(p => ({ id: p.id, name: p.name, icon: p.icon || '' }))

  // Build practice → pillar info map (includes id, icon, name)
  const map = new Map<number, { id: number; icon: string; name: string }[]>()
  for (const link of pillarLinks) {
    const list = map.get(link.practice_id) || []
    list.push({ id: link.pillar_id, icon: link.pillar_icon || '', name: link.pillar_name || '' })
    map.set(link.practice_id, list)
  }
  practicePillarMap.value = map

  loading.value = false
}

async function submit() {
  try {
    const p: Partial<Practice> = {
    name: form.value.name,
    description: form.value.description,
    type: form.value.type,
    category: form.value.category,
    end_date: form.value.end_date || undefined,
    start_date: form.value.start_date || undefined,
  }

  if (form.value.type === 'tracker') {
    p.config = JSON.stringify(trackerConfig.value)
  } else if (form.value.type === 'scheduled') {
    const sc: any = { schedule: { type: scheduleConfig.value.type } }
    const s = sc.schedule
    if (scheduleConfig.value.type === 'interval') {
      s.interval_days = scheduleConfig.value.interval_days
      s.shift_on_early = scheduleConfig.value.shift_on_early
      // Anchor to today for new practices
      if (!editingId.value) {
        const now = new Date()
        s.anchor_date = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-${String(now.getDate()).padStart(2, '0')}`
      }
    } else if (scheduleConfig.value.type === 'daily_slots') {
      s.slots = scheduleConfig.value.slots.filter((sl: string) => sl.trim())
    } else if (scheduleConfig.value.type === 'weekly') {
      s.days = scheduleConfig.value.days
    } else if (scheduleConfig.value.type === 'monthly') {
      s.day_of_month = scheduleConfig.value.day_of_month
    } else if (scheduleConfig.value.type === 'once') {
      s.due_date = scheduleConfig.value.due_date
    }
    p.config = JSON.stringify(sc)
  } else if (form.value.type === 'memorize') {
    if (editingId.value !== null) {
      const existing = practices.value.find(pr => pr.id === editingId.value)
      if (existing) {
        try {
          const cfg = JSON.parse(existing.config)
          cfg.target_daily_reps = memorizeConfig.value.target_daily_reps
          p.config = JSON.stringify(cfg)
        } catch {
          p.config = existing.config
        }
      }
    }
    // For new cards, config will be set by backend with DefaultSM2Config
    // but we pass target_daily_reps hint
    if (!editingId.value) {
      p.config = JSON.stringify({ target_daily_reps: memorizeConfig.value.target_daily_reps })
    }
  } else {
    p.config = '{}'
  }

  if (editingId.value !== null) {
    // Merge form changes onto the existing practice to preserve all fields (active, sort_order, etc.)
    const existing = practices.value.find(pr => pr.id === editingId.value)
    if (existing) {
      const merged = { ...existing, ...p }
      // Config was already set correctly above for each type
      await api.updatePractice(editingId.value, merged as Practice)
      // Save pillar associations
      await api.setPracticePillars(editingId.value, formPillarIds.value)
    }
  } else {
    const created = await api.createPractice(p)
    // Save pillar associations for new practice
    if (created && created.id && formPillarIds.value.length > 0) {
      await api.setPracticePillars(created.id, formPillarIds.value)
    }
  }
  resetForm()
  await load()
  } catch (err) {
    console.error('Failed to save practice:', err)
    alert('Failed to save practice. Please try again.')
  }
}

function resetForm() {
  form.value = { name: '', description: '', type: 'habit', category: '', config: '{}', end_date: '', start_date: '' }
  trackerConfig.value = { target_sets: 2, target_reps: 15, unit: 'reps' }
  memorizeConfig.value = { target_daily_reps: 1 }
  scheduleConfig.value = { type: 'interval', interval_days: 2, shift_on_early: true, slots: ['morning', 'lunch', 'night'], days: [], day_of_month: 1, due_date: '', newSlot: '' }
  formPillarIds.value = []
  editingId.value = null
  showForm.value = false
}

function editPractice(p: Practice) {
  editingId.value = p.id
  form.value.name = p.name
  form.value.description = p.description
  form.value.type = p.type
  form.value.category = p.category
  form.value.config = p.config
  form.value.end_date = p.end_date ? p.end_date.slice(0, 10) : ''
  form.value.start_date = p.start_date ? p.start_date.slice(0, 10) : ''

  // Load pillar associations
  api.getPracticePillars(p.id).then(links => {
    formPillarIds.value = links.map((l: PillarLink) => l.pillar_id)
  }).catch(() => {
    formPillarIds.value = []
  })

  // Populate tracker config if editing a tracker
  if (p.type === 'tracker' && p.config) {
    try {
      const cfg = JSON.parse(p.config)
      trackerConfig.value = {
        target_sets: cfg.target_sets ?? 2,
        target_reps: cfg.target_reps ?? 15,
        unit: cfg.unit ?? 'reps',
      }
    } catch {
      // keep defaults
    }
  }

  // Populate memorize daily reps if editing a memorize card
  if (p.type === 'memorize' && p.config) {
    try {
      const cfg = JSON.parse(p.config)
      memorizeConfig.value = {
        target_daily_reps: cfg.target_daily_reps ?? 1,
      }
    } catch {
      // keep defaults
    }
  }

  // Populate schedule config if editing a scheduled practice
  if (p.type === 'scheduled' && p.config) {
    try {
      const cfg = JSON.parse(p.config)
      const s = cfg.schedule || {}
      scheduleConfig.value = {
        type: s.type || 'interval',
        interval_days: s.interval_days ?? 2,
        shift_on_early: s.shift_on_early ?? true,
        slots: s.slots || ['morning', 'lunch', 'night'],
        days: s.days || [],
        day_of_month: s.day_of_month ?? 1,
        due_date: s.due_date || '',
        newSlot: '',
      }
    } catch {
      // keep defaults
    }
  }

  showForm.value = true
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

async function toggleActive(p: Practice) {
  if (p.status === 'active' || p.active) {
    await api.pausePractice(p.id)
  } else {
    await api.restorePractice(p.id)
  }
  await load()
}

async function completePractice(p: Practice) {
  if (!confirm(`Mark "${p.name}" as completed?`)) return
  await api.completePractice(p.id)
  await load()
}

async function archivePractice(p: Practice) {
  if (!confirm(`Archive "${p.name}"?`)) return
  await api.archivePractice(p.id)
  await load()
}

async function restorePractice(p: Practice) {
  await api.restorePractice(p.id)
  await load()
}

async function deletePractice(p: Practice) {
  if (!confirm(`Delete "${p.name}" and all its logs?`)) return
  await api.deletePractice(p.id)
  await load()
}

function goToPractice(p: Practice) {
  if (p.type === 'memorize') {
    router.push(`/memorize?id=${p.id}`)
  } else {
    router.push(`/practices/${p.id}/history`)
  }
}

function practiceNotify(p: Practice): boolean {
  if (!p.config) return false
  try {
    const cfg = JSON.parse(p.config)
    return !!cfg.notify
  } catch {
    return false
  }
}

async function togglePracticeNotify(p: Practice) {
  const current = practiceNotify(p)
  let cfg: Record<string, any> = {}
  try { cfg = JSON.parse(p.config || '{}') } catch { /* use empty */ }
  cfg.notify = !current
  const newConfig = JSON.stringify(cfg)
  // Send the FULL practice with updated config — partial updates corrupt data
  const updated = await api.updatePractice(p.id, { ...p, config: newConfig })
  Object.assign(p, updated)
}

// Scripture lookup
const lookingUp = ref(false)
const lookupError = ref('')

async function lookupScripture() {
  if (!form.value.name.trim()) return
  lookingUp.value = true
  lookupError.value = ''
  try {
    const result = await api.lookupScripture(form.value.name.trim())
    if (result.verses && result.verses.length > 0) {
      form.value.description = result.verses.map(v => v.text).join(' ')
      // Normalize the name to the canonical reference
      if (result.verses.length === 1) {
        form.value.name = result.verses[0]!.reference
      }
    }
  } catch (e: any) {
    lookupError.value = e.message || 'Not found'
  }
  lookingUp.value = false
}

onMounted(async () => {
  await load()
  // Auto-open create form if redirected from Memorize tab (e.g., ?type=memorize&create=1)
  if (route.query.create === '1') {
    const qType = route.query.type as string
    if (qType && ['memorize', 'habit', 'tracker', 'scheduled'].includes(qType)) {
      form.value.type = qType as Practice['type']
      if (qType === 'memorize' && route.query.reps) {
        memorizeConfig.value.target_daily_reps = parseInt(route.query.reps as string) || 1
      }
    }
    showForm.value = true
    // Clean up query params without triggering navigation
    router.replace({ path: '/practices' })
  }
})
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold">Practices</h1>
      <button
        @click="showForm ? resetForm() : (showForm = true)"
        class="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 text-sm"
      >
        {{ showForm ? 'Cancel' : '+ Add Practice' }}
      </button>
    </div>

    <!-- Status tabs -->
    <div class="flex gap-1 mb-4 border-b border-gray-200">
      <button
        v-for="tab in [
          { value: 'active', label: 'Active' },
          { value: 'paused', label: 'Paused' },
          { value: 'completed', label: 'Completed' },
          { value: 'archived', label: 'Archived' },
        ]"
        :key="tab.value"
        @click="filterStatus = tab.value; load()"
        class="px-4 py-2 text-sm font-medium border-b-2 -mb-px transition-colors"
        :class="filterStatus === tab.value
          ? 'border-indigo-500 text-indigo-600'
          : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'"
      >
        {{ tab.label }}
      </button>
    </div>

    <!-- Add/Edit form -->
    <div v-if="showForm" class="bg-white rounded-lg shadow p-4 mb-6">
      <h2 class="text-lg font-semibold mb-3">{{ editingId ? 'Edit Practice' : 'New Practice' }}</h2>
      <form @submit.prevent="submit" class="space-y-4">
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Name</label>
            <div class="flex gap-2">
              <input
                v-model="form.name"
                required
                class="flex-1 border rounded px-3 py-2 text-sm"
                placeholder="Clamshell, D&C 93:29, Morning prayer..."
              />
              <button
                v-if="form.type === 'memorize'"
                type="button"
                @click="lookupScripture"
                :disabled="lookingUp || !form.name.trim()"
                class="px-3 py-2 bg-indigo-100 text-indigo-700 rounded text-sm hover:bg-indigo-200 disabled:opacity-40 disabled:cursor-not-allowed whitespace-nowrap"
              >
                {{ lookingUp ? '...' : '📖 Lookup' }}
              </button>
            </div>
            <p v-if="lookupError" class="text-xs text-red-500 mt-1">{{ lookupError }}</p>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Type</label>
            <select v-model="form.type" class="w-full border rounded px-3 py-2 text-sm">
              <option value="habit">Habit (daily check)</option>
              <option value="tracker">Tracker (sets/reps)</option>
              <option value="memorize">Memorize (spaced repetition)</option>
              <option value="scheduled">Scheduled (recurring)</option>
              <option value="task">Task (one-time)</option>
            </select>
          </div>
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Categories <span class="font-normal text-gray-400">(multi-select)</span></label>
          <div class="flex gap-2 flex-wrap">
            <button
              v-for="cat in presetCategories"
              :key="cat"
              type="button"
              @click="toggleCategory(cat)"
              class="px-3 py-1 text-xs rounded-full border"
              :class="categoryList().includes(cat)
                ? 'bg-indigo-100 border-indigo-300 text-indigo-700'
                : 'bg-gray-50 border-gray-200 text-gray-600 hover:bg-gray-100'"
            >
              {{ cat }}
            </button>
            <input
              v-model="form.category"
              class="px-3 py-1 text-xs border rounded-full w-36"
              placeholder="custom tags..."
            />
          </div>
          <p class="text-xs text-gray-400 mt-1">Comma-separated for multiple: pt,fitness</p>
        </div>

        <!-- Pillar selector -->
        <div v-if="allPillars.length > 0">
          <label class="block text-sm font-medium text-gray-700 mb-1">Pillars <span class="font-normal text-gray-400">(multi-select)</span></label>
          <div class="flex gap-2 flex-wrap">
            <button
              v-for="pillar in allPillars"
              :key="pillar.id"
              type="button"
              @click="formPillarIds.includes(pillar.id) ? formPillarIds = formPillarIds.filter(id => id !== pillar.id) : formPillarIds.push(pillar.id)"
              class="px-3 py-1 text-xs rounded-full border"
              :class="formPillarIds.includes(pillar.id)
                ? 'bg-indigo-100 border-indigo-300 text-indigo-700'
                : 'bg-gray-50 border-gray-200 text-gray-600 hover:bg-gray-100'"
            >
              {{ pillar.icon }} {{ pillar.name }}
            </button>
          </div>
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">
            {{ form.type === 'memorize' ? 'Verse / Quote Text' : 'Description' }}
          </label>
          <textarea
            v-model="form.description"
            :rows="form.type === 'memorize' ? 4 : 2"
            class="w-full border rounded px-3 py-2 text-sm"
            :placeholder="form.type === 'memorize'
              ? 'Man was also in the beginning with God. Intelligence, or the light of truth, was not created or made, neither indeed can be.'
              : 'Full verse text, exercise instructions, etc.'"
          ></textarea>
        </div>

        <!-- Memorize hint + daily reps -->
        <div v-if="form.type === 'memorize'" class="bg-indigo-50 rounded p-3 space-y-3">
          <p class="text-sm text-indigo-700">
            <strong>Tip:</strong> Put the scripture reference as the Name (flashcard front).
            Put the full verse text in the Description (flashcard back).
          </p>
          <div>
            <label class="block text-xs text-indigo-600 mb-1">Daily practice goal</label>
            <div class="flex items-center gap-2">
              <input
                v-model.number="memorizeConfig.target_daily_reps"
                type="number"
                min="1"
                max="20"
                class="w-20 border rounded px-2 py-1 text-sm"
              />
              <span class="text-sm text-indigo-600">reviews per day</span>
            </div>
          </div>
        </div>

        <!-- Tracker config (was exercise) -->
        <div v-if="form.type === 'tracker'" class="bg-gray-50 rounded p-3 space-y-3">
          <h3 class="text-sm font-medium text-gray-700">Tracker Settings</h3>
          <div class="grid grid-cols-3 gap-3">
            <div>
              <label class="block text-xs text-gray-500">Target Sets</label>
              <input
                v-model.number="trackerConfig.target_sets"
                type="number"
                min="1"
                class="w-full border rounded px-2 py-1 text-sm"
              />
            </div>
            <div>
              <label class="block text-xs text-gray-500">Target Reps</label>
              <input
                v-model.number="trackerConfig.target_reps"
                type="number"
                min="1"
                class="w-full border rounded px-2 py-1 text-sm"
              />
            </div>
            <div>
              <label class="block text-xs text-gray-500">Unit</label>
              <select v-model="trackerConfig.unit" class="w-full border rounded px-2 py-1 text-sm">
                <option value="reps">reps</option>
                <option value="bottles">bottles</option>
                <option value="glasses">glasses</option>
                <option value="seconds">seconds</option>
                <option value="minutes">minutes</option>
              </select>
            </div>
          </div>
        </div>

        <!-- Schedule config -->
        <div v-if="form.type === 'scheduled'" class="bg-amber-50 rounded p-3 space-y-3">
          <h3 class="text-sm font-medium text-gray-700">Schedule</h3>

          <!-- Schedule type radio -->
          <div class="flex gap-2 flex-wrap">
            <label
              v-for="opt in [
                { value: 'interval', label: 'Every N days' },
                { value: 'daily_slots', label: 'Multiple/day' },
                { value: 'weekly', label: 'Weekly' },
                { value: 'monthly', label: 'Monthly' },
                { value: 'once', label: 'One-time' },
              ]"
              :key="opt.value"
              class="flex items-center gap-1 px-2.5 py-1 text-xs rounded-full border cursor-pointer"
              :class="scheduleConfig.type === opt.value
                ? 'bg-amber-200 border-amber-400 text-amber-800'
                : 'bg-white border-gray-200 text-gray-600 hover:bg-gray-50'"
            >
              <input
                type="radio"
                v-model="scheduleConfig.type"
                :value="opt.value"
                class="sr-only"
              />
              {{ opt.label }}
            </label>
          </div>

          <!-- Interval config -->
          <div v-if="scheduleConfig.type === 'interval'" class="space-y-2">
            <div class="flex items-center gap-2">
              <label class="text-xs text-gray-600">Every</label>
              <input
                v-model.number="scheduleConfig.interval_days"
                type="number"
                min="1"
                max="365"
                class="w-16 border rounded px-2 py-1 text-sm"
              />
              <span class="text-xs text-gray-600">days</span>
            </div>
            <label class="flex items-center gap-2 text-xs text-gray-600 cursor-pointer">
              <input
                type="checkbox"
                v-model="scheduleConfig.shift_on_early"
                class="rounded border-gray-300"
              />
              Shift schedule if done early
            </label>
          </div>

          <!-- Daily slots config -->
          <div v-if="scheduleConfig.type === 'daily_slots'" class="space-y-2">
            <label class="text-xs text-gray-600">Time slots</label>
            <div class="flex gap-1.5 flex-wrap">
              <span
                v-for="(slot, idx) in scheduleConfig.slots"
                :key="idx"
                class="flex items-center gap-1 px-2 py-0.5 bg-amber-100 border border-amber-300 rounded-full text-xs text-amber-800"
              >
                {{ slot }}
                <button
                  type="button"
                  @click="scheduleConfig.slots.splice(idx, 1)"
                  class="text-amber-500 hover:text-red-500 ml-0.5"
                >&times;</button>
              </span>
              <form
                @submit.prevent="if (scheduleConfig.newSlot.trim()) { scheduleConfig.slots.push(scheduleConfig.newSlot.trim()); scheduleConfig.newSlot = '' }"
                class="flex"
              >
                <input
                  v-model="scheduleConfig.newSlot"
                  class="w-24 border rounded-l px-2 py-0.5 text-xs"
                  placeholder="add slot..."
                />
                <button
                  type="submit"
                  class="px-2 py-0.5 bg-amber-200 border border-l-0 border-amber-300 rounded-r text-xs text-amber-700 hover:bg-amber-300"
                >+</button>
              </form>
            </div>
          </div>

          <!-- Weekly config -->
          <div v-if="scheduleConfig.type === 'weekly'" class="space-y-2">
            <label class="text-xs text-gray-600">Days of the week</label>
            <div class="flex gap-1.5">
              <label
                v-for="day in ['sun','mon','tue','wed','thu','fri','sat']"
                :key="day"
                class="flex items-center justify-center w-10 h-8 text-xs rounded border cursor-pointer select-none"
                :class="scheduleConfig.days.includes(day)
                  ? 'bg-amber-200 border-amber-400 text-amber-800 font-medium'
                  : 'bg-white border-gray-200 text-gray-500 hover:bg-gray-50'"
              >
                <input
                  type="checkbox"
                  :value="day"
                  v-model="scheduleConfig.days"
                  class="sr-only"
                />
                {{ day.charAt(0).toUpperCase() + day.slice(1) }}
              </label>
            </div>
          </div>

          <!-- Monthly config -->
          <div v-if="scheduleConfig.type === 'monthly'" class="flex items-center gap-2">
            <label class="text-xs text-gray-600">Day of month:</label>
            <input
              v-model.number="scheduleConfig.day_of_month"
              type="number"
              min="1"
              max="31"
              class="w-16 border rounded px-2 py-1 text-sm"
            />
          </div>

          <!-- Once config -->
          <div v-if="scheduleConfig.type === 'once'" class="flex items-center gap-2">
            <label class="text-xs text-gray-600">Due date:</label>
            <input
              v-model="scheduleConfig.due_date"
              type="date"
              class="border rounded px-2 py-1 text-sm"
            />
          </div>
        </div>

        <!-- Start / End dates -->
        <div class="flex gap-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Start Date <span class="font-normal text-gray-400">(defaults to today)</span></label>
            <input
              v-model="form.start_date"
              type="date"
              class="border rounded px-3 py-2 text-sm"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Target End Date <span class="font-normal text-gray-400">(optional)</span></label>
            <input
              v-model="form.end_date"
              type="date"
              class="border rounded px-3 py-2 text-sm"
            />
          </div>
        </div>

        <div class="flex gap-2">
          <button
            type="submit"
            class="px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 text-sm"
          >
            {{ editingId ? 'Save Changes' : 'Add Practice' }}
          </button>
          <button
            v-if="editingId"
            type="button"
            @click="resetForm"
            class="px-4 py-2 bg-gray-200 text-gray-700 rounded-lg hover:bg-gray-300 text-sm"
          >
            Cancel
          </button>
        </div>
      </form>
    </div>

    <!-- Practice list -->
    <div v-if="loading" class="text-center py-8 text-gray-400">Loading...</div>

    <template v-else-if="practices.length > 0">
      <!-- Filters -->
      <div class="mb-4 space-y-2">
        <div class="flex gap-1.5 flex-wrap items-center">
          <span class="text-xs text-gray-400 w-10">Type</span>
          <button
            @click="cycleFilter('type', 'all', true)"
            class="px-2.5 py-1 text-xs rounded-full border transition-colors"
            :class="allChipClass(filterTypeState)"
          >all</button>
          <button
            v-for="t in availableTypes"
            :key="t"
            @click="cycleFilter('type', t)"
            class="px-2.5 py-1 text-xs rounded-full border transition-colors"
            :class="filterChipClass(filterState(filterTypeState, t))"
          >{{ t }}</button>
        </div>
        <div v-if="availableCategories.length > 1" class="flex gap-1.5 flex-wrap items-center">
          <span class="text-xs text-gray-400 w-10">Cat</span>
          <button
            @click="cycleFilter('cat', 'all', true)"
            class="px-2.5 py-1 text-xs rounded-full border transition-colors"
            :class="allChipClass(filterCatState)"
          >all</button>
          <button
            v-for="c in availableCategories"
            :key="c"
            @click="cycleFilter('cat', c)"
            class="px-2.5 py-1 text-xs rounded-full border transition-colors"
            :class="filterChipClass(filterState(filterCatState, c))"
          >{{ c }}</button>
        </div>
        <div v-if="allPillars.length > 0" class="flex gap-1.5 flex-wrap items-center">
          <span class="text-xs text-gray-400 w-10">Pillar</span>
          <button
            @click="cycleFilter('pillar', 'all', true)"
            class="px-2.5 py-1 text-xs rounded-full border transition-colors"
            :class="allChipClass(filterPillarState)"
          >all</button>
          <button
            v-for="pl in allPillars"
            :key="pl.id"
            @click="cycleFilter('pillar', String(pl.id))"
            class="px-2.5 py-1 text-xs rounded-full border transition-colors"
            :class="filterChipClass(filterState(filterPillarState, String(pl.id)))"
          >{{ pl.icon }} {{ pl.name }}</button>
        </div>
        <div class="flex gap-1.5 flex-wrap items-center">
          <span class="text-xs text-gray-400 w-10">Time</span>
          <button
            v-for="tf in timeFilterOptions"
            :key="tf"
            @click="filterTime = tf"
            class="px-2.5 py-1 text-xs rounded-full border"
            :class="filterTime === tf ? 'bg-indigo-100 border-indigo-300 text-indigo-700' : 'bg-gray-50 border-gray-200 text-gray-500 hover:bg-gray-100'"
          >{{ tf }}</button>
        </div>
      </div>

      <div v-if="filteredPractices.length === 0" class="text-center py-8 text-gray-400">
        No practices match filters.
      </div>

      <div v-else class="bg-white rounded-lg shadow divide-y divide-gray-100">
        <div
          v-for="p in filteredPractices"
          :key="p.id"
        class="flex items-center justify-between px-4 py-3"
        :class="{ 'opacity-50': p.status !== 'active' }"
      >
        <div class="min-w-0 flex-1 cursor-pointer" @click="goToPractice(p)">
          <div class="flex items-center gap-2">
            <span class="font-medium hover:text-indigo-600 transition-colors">{{ p.name }}</span>
            <span class="text-xs px-2 py-0.5 rounded-full bg-gray-100 text-gray-600">{{ p.type }}</span>
            <span v-if="p.category" class="text-xs px-2 py-0.5 rounded-full bg-indigo-50 text-indigo-600">{{ p.category }}</span>
            <span
              v-for="pl in (practicePillarMap.get(p.id) || [])"
              :key="pl.id"
              class="text-xs px-1.5 py-0.5 rounded-full bg-purple-50 text-purple-600"
              :title="pl.name"
            >{{ pl.icon || '◆' }}</span>
            <span v-if="p.start_date && startDateLabel(p.start_date)" class="text-xs px-2 py-0.5 rounded-full cursor-default" :class="startDateClass()" :title="endDateTooltip(p.start_date)">{{ startDateLabel(p.start_date) }}</span>
            <span v-if="p.end_date" class="text-xs px-2 py-0.5 rounded-full cursor-default" :class="endDateClass(p.end_date)" :title="endDateTooltip(p.end_date)">{{ endDateLabel(p.end_date) }}</span>
          </div>
          <div v-if="p.description" class="text-xs text-gray-400 truncate mt-0.5">{{ p.description }}</div>
        </div>

        <div class="flex items-center gap-2 ml-4">
          <!-- Per-practice notification toggle (scheduled practices only) -->
          <button
            v-if="notifSubscribed && p.type === 'scheduled' && p.status === 'active'"
            @click.stop="togglePracticeNotify(p)"
            class="text-xs px-1 py-1 rounded"
            :class="practiceNotify(p) ? 'text-indigo-500 hover:text-indigo-700' : 'text-gray-300 hover:text-gray-500'"
            :aria-label="practiceNotify(p) ? 'Disable notifications for ' + p.name : 'Enable notifications for ' + p.name"
            :title="practiceNotify(p) ? 'Notifications on' : 'Notifications off'"
          >
            <svg v-if="practiceNotify(p)" xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
              <path d="M10 2a6 6 0 00-6 6v3.586l-.707.707A1 1 0 004 14h12a1 1 0 00.707-1.707L16 11.586V8a6 6 0 00-6-6zM10 18a3 3 0 01-3-3h6a3 3 0 01-3 3z" />
            </svg>
            <svg v-else xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
              <path fill-rule="evenodd" d="M3.707 2.293a1 1 0 00-1.414 1.414l14 14a1 1 0 001.414-1.414l-1.473-1.473A1 1 0 0016.5 14h-13a1 1 0 01-.5-1.866V8a6 6 0 014.27-5.748A2 2 0 018 2.5V2a2 2 0 014 0v.5a2 2 0 01.73.252L3.707 2.293zM10 18a3 3 0 01-3-3h6a3 3 0 01-3 3z" clip-rule="evenodd" />
            </svg>
          </button>

          <router-link
            :to="`/practices/${p.id}/history`"
            class="text-xs text-gray-400 hover:text-indigo-500"
          >
            history
          </router-link>

          <!-- Active state actions -->
          <template v-if="p.status === 'active'">
            <button
              @click.stop="editPractice(p)"
              class="text-xs text-indigo-500 hover:text-indigo-700 px-2 py-1"
            >edit</button>
            <button
              @click.stop="toggleActive(p)"
              class="text-xs text-yellow-600 hover:bg-yellow-50 px-2 py-1 rounded"
            >pause</button>
            <button
              @click.stop="completePractice(p)"
              class="text-xs text-green-600 hover:bg-green-50 px-2 py-1 rounded"
            >✓ complete</button>
            <button
              @click.stop="archivePractice(p)"
              class="text-xs text-gray-400 hover:text-gray-600 px-2 py-1 rounded"
            >archive</button>
          </template>

          <!-- Paused state actions -->
          <template v-else-if="p.status === 'paused'">
            <button
              @click.stop="restorePractice(p)"
              class="text-xs text-green-600 hover:bg-green-50 px-2 py-1 rounded"
            >resume</button>
            <button
              @click.stop="archivePractice(p)"
              class="text-xs text-gray-400 hover:text-gray-600 px-2 py-1 rounded"
            >archive</button>
          </template>

          <!-- Completed/Archived state actions -->
          <template v-else>
            <button
              @click.stop="restorePractice(p)"
              class="text-xs text-green-600 hover:bg-green-50 px-2 py-1 rounded"
            >restore</button>
            <button
              @click.stop="deletePractice(p)"
              class="text-xs text-red-400 hover:text-red-600 px-2 py-1"
            >delete</button>
          </template>
        </div>
      </div>
    </div>
    </template>

    <div v-else class="text-center py-12 text-gray-500">
      No practices yet. Add one above!
    </div>
  </div>
</template>
