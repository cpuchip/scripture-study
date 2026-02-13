<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api, type DailySummary, type PracticeLog, type Practice, type Reflection, type Prompt, type PillarLink } from '../api'

function localDateStr(d: Date = new Date()): string {
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

const route = useRoute()
const router = useRouter()

const today = ref(localDateStr())
const summary = ref<DailySummary[]>([])
const dueCards = ref<Practice[]>([])
const loading = ref(true)

// Parse exercise config
function parseConfig(config: string) {
  try { return JSON.parse(config) } catch { return {} }
}

// Group by category (uses first category if comma-separated)
const groupMode = ref<'category' | 'pillar'>('category')
const activeFilter = ref<string>('all')
const practicePillarMap = ref<Record<number, PillarLink[]>>({})
const hasPillars = ref(false)

// Available filter options based on group mode
const availableCategories = computed(() => {
  const cats = new Set<string>()
  for (const item of summary.value) {
    const raw = item.category || item.practice_type || 'other'
    const cat = raw.split(',')[0]?.trim() || raw
    cats.add(cat)
  }
  return Array.from(cats).sort()
})

const availablePillars = computed(() => {
  const pillars = new Map<string, string>() // label → label (deduped)
  for (const item of summary.value) {
    const links = practicePillarMap.value[item.practice_id]
    if (links && links.length > 0) {
      for (const link of links) {
        const label = `${link.pillar_icon} ${link.pillar_name}`
        pillars.set(label, label)
      }
    } else {
      pillars.set('Uncategorized', 'Uncategorized')
    }
  }
  return Array.from(pillars.keys()).sort()
})

const filterOptions = computed(() => {
  return groupMode.value === 'pillar' && hasPillars.value
    ? availablePillars.value
    : availableCategories.value
})

// Reset filter when switching group mode
function setGroupMode(mode: 'category' | 'pillar') {
  groupMode.value = mode
  activeFilter.value = 'all'
}

// Filtered summary based on active filter
const filteredSummary = computed(() => {
  if (activeFilter.value === 'all') return summary.value
  if (groupMode.value === 'pillar' && hasPillars.value) {
    return summary.value.filter(item => {
      const links = practicePillarMap.value[item.practice_id]
      if (activeFilter.value === 'Uncategorized') {
        return !links || links.length === 0
      }
      return links?.some(l => `${l.pillar_icon} ${l.pillar_name}` === activeFilter.value)
    })
  }
  // Category filter
  return summary.value.filter(item => {
    const raw = item.category || item.practice_type || 'other'
    const cat = raw.split(',')[0]?.trim() || raw
    return cat === activeFilter.value
  })
})

const grouped = computed(() => {
  const source = filteredSummary.value
  if (groupMode.value === 'pillar' && hasPillars.value) {
    const groups: Record<string, DailySummary[]> = {}
    for (const item of source) {
      const links = practicePillarMap.value[item.practice_id]
      if (links && links.length > 0) {
        for (const link of links) {
          const label = `${link.pillar_icon} ${link.pillar_name}`
          if (activeFilter.value !== 'all' && label !== activeFilter.value) continue
          if (!groups[label]) groups[label] = []
          groups[label].push(item)
        }
      } else {
        if (!groups['Uncategorized']) groups['Uncategorized'] = []
        groups['Uncategorized'].push(item)
      }
    }
    return groups
  }
  const groups: Record<string, DailySummary[]> = {}
  for (const item of source) {
    const raw = item.category || item.practice_type || 'other'
    const cat = raw.split(',')[0]?.trim() || raw
    if (!groups[cat]) groups[cat] = []
    groups[cat].push(item)
  }
  return groups
})

// Reflection state
const reflection = ref<Reflection | null>(null)
const todayPrompt = ref<Prompt | null>(null)
const reflectionContent = ref('')
const reflectionMood = ref<number | null>(null)
const reflectionExpanded = ref(false)
const reflectionSaving = ref(false)
const reflectionTextarea = ref<HTMLTextAreaElement | null>(null)

const moods = [
  { value: 1, emoji: '😟', label: 'Struggling' },
  { value: 2, emoji: '😐', label: 'Meh' },
  { value: 3, emoji: '🙂', label: 'Okay' },
  { value: 4, emoji: '😊', label: 'Good' },
  { value: 5, emoji: '😄', label: 'Great' },
]

async function load() {
  loading.value = true
  try {
    const [s, d, ref_, prompt, pillarsCheck] = await Promise.all([
      api.getDailySummary(today.value),
      api.getDueCards(today.value),
      api.getReflection(today.value),
      api.getTodayPrompt(),
      api.hasPillars(),
    ])
    summary.value = s
    dueCards.value = d
    reflection.value = ref_
    todayPrompt.value = prompt
    hasPillars.value = pillarsCheck.has_pillars

    // Load pillar mappings for all practices
    if (hasPillars.value) {
      const mapping: Record<number, PillarLink[]> = {}
      await Promise.all(s.map(async (item) => {
        try {
          const links = await api.getPracticePillars(item.practice_id)
          if (links.length > 0) mapping[item.practice_id] = links
        } catch { /* noop */ }
      }))
      practicePillarMap.value = mapping
    }

    if (ref_) {
      reflectionContent.value = ref_.content
      reflectionMood.value = ref_.mood ?? null
      reflectionExpanded.value = true
    } else {
      reflectionContent.value = ''
      reflectionMood.value = null
      reflectionExpanded.value = false
    }
  } catch (e) {
    console.error('Failed to load daily summary:', e)
  }
  loading.value = false
}

async function saveReflection() {
  if (!reflectionContent.value.trim()) return
  reflectionSaving.value = true
  try {
    const r = await api.upsertReflection({
      date: today.value,
      prompt_id: todayPrompt.value?.id ?? undefined,
      prompt_text: todayPrompt.value?.text ?? undefined,
      content: reflectionContent.value,
      mood: reflectionMood.value,
    })
    reflection.value = r
  } catch (e) {
    console.error('Failed to save reflection:', e)
  }
  reflectionSaving.value = false
}

// Quick log: just mark it done (habit) or log a set (exercise)
async function quickLog(item: DailySummary, slotName?: string) {
  const config = parseConfig(item.config)
  const log: Partial<PracticeLog> = {
    practice_id: item.practice_id,
    date: today.value,
  }

  if (item.practice_type === 'tracker') {
    log.sets = 1
    log.reps = config.target_reps || undefined
  }

  // For daily_slots scheduled items, record which slot was completed.
  if (slotName) {
    log.value = slotName
  }

  await api.createLog(log)
  await load()
}

// Undo: remove last log for today (server-side finds the most recent)
async function undoLog(item: DailySummary) {
  try {
    await api.deleteLatestLog(item.practice_id, today.value)
    await load()
  } catch {
    // No log found to delete — just refresh
    await load()
  }
}

// Navigate date
function prevDay() {
  const d = new Date(today.value + 'T12:00:00')
  d.setDate(d.getDate() - 1)
  today.value = localDateStr(d)
  load()
}
function nextDay() {
  const d = new Date(today.value + 'T12:00:00')
  d.setDate(d.getDate() + 1)
  today.value = localDateStr(d)
  load()
}
function goToToday() {
  today.value = localDateStr()
  load()
}
const isToday = computed(() => today.value === localDateStr())

// Completion stats (respects active filter)
const completionStats = computed(() => {
  // Exclude non-due scheduled items from the total.
  const relevant = filteredSummary.value.filter(s =>
    s.practice_type !== 'scheduled' || s.is_due === true || s.log_count > 0
  )
  const total = relevant.length
  const done = relevant.filter(s => {
    if (s.practice_type === 'tracker') {
      const config = parseConfig(s.config)
      return (s.total_sets || s.log_count) >= (config.target_sets || 1)
    }
    if (s.practice_type === 'memorize') {
      const config = parseConfig(s.config)
      return s.log_count >= (config.target_daily_reps || 1)
    }
    if (s.practice_type === 'scheduled') {
      const config = parseConfig(s.config)
      if (config.schedule?.type === 'daily_slots') {
        return !s.slots_due || s.slots_due.length === 0
      }
      return s.log_count > 0
    }
    return s.log_count > 0
  }).length
  return { done, total }
})

// Display helpers
function trackerTargetSets(item: DailySummary): number {
  const config = parseConfig(item.config)
  return config.target_sets || 1
}

function trackerCompletedSets(item: DailySummary): number {
  return item.total_sets || item.log_count
}

function trackerRepsLabel(item: DailySummary): string {
  const config = parseConfig(item.config)
  return `${config.target_reps || '?'} ${config.unit || 'reps'}`
}

function memorizeTargetReps(item: DailySummary): number {
  const config = parseConfig(item.config)
  return config.target_daily_reps || 1
}

function memorizeCompletedReps(item: DailySummary): number {
  return item.log_count
}

function isComplete(item: DailySummary): boolean {
  if (item.practice_type === 'tracker') {
    const config = parseConfig(item.config)
    const targetSets = config.target_sets || 1
    const done = item.total_sets || item.log_count
    return done >= targetSets
  }
  if (item.practice_type === 'memorize') {
    const target = memorizeTargetReps(item)
    return item.log_count >= target
  }
  if (item.practice_type === 'scheduled') {
    const config = parseConfig(item.config)
    const sched = config.schedule
    if (sched?.type === 'daily_slots') {
      // Complete when all slots are done (no remaining slots_due).
      return !item.slots_due || item.slots_due.length === 0
    }
    // For other schedule types, one log = done.
    return item.log_count > 0
  }
  return item.log_count > 0
}

// Schedule display helpers.
function scheduleIsDue(item: DailySummary): boolean {
  return item.is_due === true
}

function scheduleLabel(item: DailySummary): string {
  const config = parseConfig(item.config)
  const sched = config.schedule
  if (!sched) return ''
  switch (sched.type) {
    case 'interval': return `every ${sched.interval_days} days`
    case 'daily_slots': return `${(sched.slots || []).length} slots/day`
    case 'weekly': return (sched.days || []).join(', ')
    case 'monthly': return `monthly (${sched.day_of_month}${ordSuffix(sched.day_of_month)})`
    case 'once': return `due ${sched.due_date}`
    default: return ''
  }
}

function nextDueLabel(item: DailySummary): string {
  if (!item.next_due) return ''
  const d = new Date(item.next_due + 'T12:00:00')
  const t = new Date(today.value + 'T12:00:00')
  const diff = Math.round((d.getTime() - t.getTime()) / 86400000)
  if (diff <= 0) return ''
  if (diff === 1) return 'tomorrow'
  return `in ${diff} days`
}

function ordSuffix(n: number): string {
  if (n >= 11 && n <= 13) return 'th'
  switch (n % 10) {
    case 1: return 'st'
    case 2: return 'nd'
    case 3: return 'rd'
    default: return 'th'
  }
}

function formatDate(date: string): string {
  return new Date(date + 'T12:00:00').toLocaleDateString('en-US', {
    weekday: 'long',
    month: 'long',
    day: 'numeric',
    year: 'numeric',
  })
}

// Refresh when returning to the page (e.g., after reviewing a card)
function onVisibilityChange() {
  if (!document.hidden) load()
}

onMounted(async () => {
  await load()
  document.addEventListener('visibilitychange', onVisibilityChange)

  // Auto-expand reflection when arriving from Reflections page (?reflect=1)
  if (route.query.reflect === '1') {
    reflectionExpanded.value = true
    router.replace({ path: '/today' })
    await nextTick()
    reflectionTextarea.value?.scrollIntoView({ behavior: 'smooth', block: 'center' })
    reflectionTextarea.value?.focus()
  }
})

onUnmounted(() => {
  document.removeEventListener('visibilitychange', onVisibilityChange)
})
</script>

<template>
  <div>
    <!-- Date header -->
    <div class="flex items-center justify-between mb-2">
      <button @click="prevDay" class="p-2 hover:bg-gray-200 rounded">←</button>
      <div class="text-center">
        <h1 class="text-2xl font-bold">{{ formatDate(today) }}</h1>
        <p class="text-sm text-gray-500">
          {{ completionStats.done }}/{{ completionStats.total }} completed
        </p>
      </div>
      <button @click="nextDay" class="p-2 hover:bg-gray-200 rounded">→</button>
    </div>
    <!-- Jump to today -->
    <div v-if="!isToday" class="flex justify-center mb-6">
      <button
        @click="goToToday"
        class="px-3 py-1 text-sm font-medium text-indigo-600 bg-indigo-50 hover:bg-indigo-100 rounded-full transition-colors"
      >
        ↩ Jump to Today
      </button>
    </div>
    <div v-else class="mb-6"></div>

    <!-- Memorize cards due -->
    <div v-if="!loading && dueCards.length > 0" class="mb-6">
      <div class="bg-indigo-50 border border-indigo-200 rounded-lg overflow-hidden">
        <router-link
          to="/memorize"
          class="block px-4 py-3 hover:bg-indigo-100 transition-colors border-b border-indigo-200"
        >
          <div class="flex items-center justify-between">
            <div>
              <span class="font-semibold text-indigo-700">{{ dueCards.length }} card{{ dueCards.length > 1 ? 's' : '' }} due</span>
              <span class="text-sm text-indigo-500 ml-2">for review</span>
            </div>
            <span class="text-indigo-400">Review all →</span>
          </div>
        </router-link>
        <div class="divide-y divide-indigo-100">
          <router-link
            v-for="card in dueCards"
            :key="card.id"
            :to="`/memorize?id=${card.id}`"
            class="flex items-center justify-between px-4 py-2 hover:bg-indigo-100/50 transition-colors"
          >
            <span class="text-sm font-medium text-indigo-800">{{ card.name }}</span>
            <span class="text-xs text-indigo-400">practice →</span>
          </router-link>
        </div>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="text-center py-8 text-gray-400">Loading...</div>

    <!-- Empty state -->
    <div v-else-if="summary.length === 0" class="text-center py-12">
      <p class="text-gray-500 mb-4">No practices yet.</p>
      <router-link to="/practices" class="text-indigo-600 hover:underline">
        Add your first practice →
      </router-link>
    </div>

    <!-- Filter empty state -->
    <div v-else-if="filteredSummary.length === 0" class="text-center py-8">
      <p class="text-gray-400 text-sm">No practices match this filter.</p>
      <button @click="activeFilter = 'all'" class="mt-2 text-indigo-600 text-sm hover:underline">
        Show all
      </button>
    </div>

    <!-- Practice groups -->
    <div v-else class="space-y-6">
      <!-- Filter bar -->
      <div class="space-y-2">
        <!-- Group mode toggle -->
        <div class="flex items-center gap-2">
          <span class="text-xs text-gray-400">Group by:</span>
          <button
            @click="setGroupMode('category')"
            class="px-2.5 py-1 text-xs rounded-full border transition-colors"
            :class="groupMode === 'category' ? 'bg-indigo-100 border-indigo-300 text-indigo-700' : 'bg-gray-50 border-gray-200 text-gray-500 hover:bg-gray-100'"
          >Category</button>
          <button
            v-if="hasPillars"
            @click="setGroupMode('pillar')"
            class="px-2.5 py-1 text-xs rounded-full border transition-colors"
            :class="groupMode === 'pillar' ? 'bg-indigo-100 border-indigo-300 text-indigo-700' : 'bg-gray-50 border-gray-200 text-gray-500 hover:bg-gray-100'"
          >Pillar</button>
        </div>
        <!-- Filter chips -->
        <div v-if="filterOptions.length > 1" class="flex items-center gap-1.5 flex-wrap">
          <button
            @click="activeFilter = 'all'"
            class="px-2.5 py-1 text-xs rounded-full border transition-colors"
            :class="activeFilter === 'all' ? 'bg-gray-800 border-gray-800 text-white' : 'bg-white border-gray-200 text-gray-600 hover:bg-gray-50'"
          >All</button>
          <button
            v-for="opt in filterOptions"
            :key="opt"
            @click="activeFilter = activeFilter === opt ? 'all' : opt"
            class="px-2.5 py-1 text-xs rounded-full border transition-colors capitalize"
            :class="activeFilter === opt ? 'bg-gray-800 border-gray-800 text-white' : 'bg-white border-gray-200 text-gray-600 hover:bg-gray-50'"
          >{{ opt }}</button>
        </div>
      </div>

      <div v-for="(items, category) in grouped" :key="category">
        <h2 class="text-sm font-semibold text-gray-500 uppercase tracking-wide mb-2">
          {{ category }}
        </h2>
        <div class="bg-white rounded-lg shadow divide-y divide-gray-100">
          <div
            v-for="item in items"
            :key="item.practice_id"
            class="px-4 py-2"
            :class="{ 'opacity-60': isComplete(item) && (item.practice_type !== 'scheduled' || scheduleIsDue(item)) }"
          >
            <!-- Tracker: name + reps label + inline set buttons + history -->
            <template v-if="item.practice_type === 'tracker'">
              <div class="flex items-center justify-between gap-2">
                <div class="flex items-center gap-2 min-w-0">
                  <span class="font-medium truncate">{{ item.practice_name }}</span>
                  <span class="text-xs text-gray-400 whitespace-nowrap">{{ trackerRepsLabel(item) }}</span>
                </div>
                <div class="flex items-center gap-1.5 flex-shrink-0">
                  <button
                    v-for="setNum in trackerTargetSets(item)"
                    :key="setNum"
                    @click="setNum <= trackerCompletedSets(item) ? undoLog(item) : quickLog(item)"
                    class="group flex items-center gap-1 px-2 py-1 rounded border text-xs transition-colors cursor-pointer"
                    :class="setNum <= trackerCompletedSets(item)
                      ? 'bg-green-50 border-green-300 text-green-700 hover:bg-red-50 hover:border-red-300 hover:text-red-600'
                      : 'bg-white border-gray-200 text-gray-500 hover:border-indigo-300 hover:bg-indigo-50'"
                  >
                    <span
                      class="w-3.5 h-3.5 rounded border flex items-center justify-center text-[9px] transition-colors"
                      :class="setNum <= trackerCompletedSets(item)
                        ? 'bg-green-500 border-green-500 text-white group-hover:bg-red-400 group-hover:border-red-400'
                        : 'border-gray-300'"
                    >
                      <span v-if="setNum <= trackerCompletedSets(item)" class="group-hover:hidden">✓</span>
                      <span v-if="setNum <= trackerCompletedSets(item)" class="hidden group-hover:inline">✕</span>
                    </span>
                    {{ setNum }}
                  </button>
                  <router-link
                    :to="`/practices/${item.practice_id}/history`"
                    class="text-xs text-gray-400 hover:text-indigo-500 ml-1"
                  >
                    history
                  </router-link>
                </div>
              </div>
            </template>

            <!-- Memorize: name + reps progress + link -->
            <template v-else-if="item.practice_type === 'memorize'">
              <div class="flex items-center justify-between gap-2">
                <div class="flex items-center gap-3 min-w-0">
                  <button
                    @click="isComplete(item) ? undoLog(item) : quickLog(item)"
                    class="group w-5 h-5 rounded-full border-2 flex items-center justify-center flex-shrink-0 transition-colors cursor-pointer"
                    :class="isComplete(item)
                      ? 'bg-green-500 border-green-500 text-white hover:bg-red-400 hover:border-red-400'
                      : 'border-gray-300 hover:border-indigo-400'"
                  >
                    <span v-if="isComplete(item)" class="text-[10px] group-hover:hidden">✓</span>
                    <span v-if="isComplete(item)" class="text-[10px] hidden group-hover:inline">✕</span>
                  </button>
                  <router-link
                    :to="`/memorize?id=${item.practice_id}`"
                    class="font-medium truncate hover:text-indigo-600 transition-colors"
                  >{{ item.practice_name }}</router-link>
                </div>
                <div class="flex items-center gap-2 flex-shrink-0">
                  <span
                    class="text-xs px-2 py-0.5 rounded-full"
                    :class="memorizeCompletedReps(item) >= memorizeTargetReps(item)
                      ? 'bg-green-100 text-green-700'
                      : 'bg-gray-100 text-gray-500'"
                  >
                    {{ memorizeCompletedReps(item) }}/{{ memorizeTargetReps(item) }}
                  </span>
                  <router-link
                    :to="`/practices/${item.practice_id}/history`"
                    class="text-xs text-gray-400 hover:text-indigo-500"
                  >
                    history
                  </router-link>
                </div>
              </div>
            </template>

            <!-- Habit/other: single completion circle -->
            <template v-else-if="item.practice_type !== 'scheduled'">
              <div class="flex items-center justify-between">
                <div class="flex items-center gap-3 flex-1 min-w-0">
                  <button
                    @click="isComplete(item) ? undoLog(item) : quickLog(item)"
                    class="group w-5 h-5 rounded-full border-2 flex items-center justify-center flex-shrink-0 transition-colors cursor-pointer"
                    :class="isComplete(item)
                      ? 'bg-green-500 border-green-500 text-white hover:bg-red-400 hover:border-red-400'
                      : 'border-gray-300 hover:border-indigo-400'"
                  >
                    <span v-if="isComplete(item)" class="text-[10px] group-hover:hidden">✓</span>
                    <span v-if="isComplete(item)" class="text-[10px] hidden group-hover:inline">✕</span>
                  </button>
                  <div class="min-w-0">
                    <div class="font-medium truncate">{{ item.practice_name }}</div>
                    <div v-if="item.last_value" class="text-xs text-gray-400">
                      {{ item.last_value }}
                    </div>
                  </div>
                </div>
                <router-link
                  :to="`/practices/${item.practice_id}/history`"
                  class="ml-2 text-xs text-gray-400 hover:text-indigo-500"
                >
                  history
                </router-link>
              </div>
            </template>

            <!-- Scheduled: due/not-due state, slot buttons, overdue badge -->
            <template v-else>
              <div
                class="flex items-center justify-between gap-2"
                :class="{ 'opacity-40': !scheduleIsDue(item) && !isComplete(item) }"
              >
                <div class="flex items-center gap-3 min-w-0 flex-1">
                  <!-- Daily slots: row of pill buttons -->
                  <template v-if="parseConfig(item.config).schedule?.type === 'daily_slots'">
                    <div class="flex items-center gap-1.5 flex-wrap">
                      <span class="font-medium truncate mr-1">{{ item.practice_name }}</span>
                      <button
                        v-for="slot in (parseConfig(item.config).schedule?.slots || [])"
                        :key="slot"
                        @click="
                          item.slots_due && item.slots_due.includes(slot)
                            ? quickLog(item, slot)
                            : undoLog(item)
                        "
                        class="group px-2 py-0.5 rounded-full text-xs border transition-colors cursor-pointer"
                        :class="
                          item.slots_due && item.slots_due.includes(slot)
                            ? 'bg-white border-gray-200 text-gray-600 hover:border-amber-400 hover:bg-amber-50'
                            : 'bg-green-50 border-green-300 text-green-700 hover:bg-red-50 hover:border-red-300 hover:text-red-600'
                        "
                      >
                        <span v-if="!(item.slots_due && item.slots_due.includes(slot))" class="group-hover:hidden">✓ </span>
                        <span v-if="!(item.slots_due && item.slots_due.includes(slot))" class="hidden group-hover:inline">✕ </span>
                        {{ slot }}
                      </button>
                    </div>
                  </template>

                  <!-- Other schedule types: single circle -->
                  <template v-else>
                    <button
                      v-if="scheduleIsDue(item) || isComplete(item)"
                      @click="isComplete(item) ? undoLog(item) : quickLog(item)"
                      class="group w-5 h-5 rounded-full border-2 flex items-center justify-center flex-shrink-0 transition-colors cursor-pointer"
                      :class="isComplete(item)
                        ? 'bg-green-500 border-green-500 text-white hover:bg-red-400 hover:border-red-400'
                        : 'border-amber-400 hover:border-amber-500'"
                    >
                      <span v-if="isComplete(item)" class="text-[10px] group-hover:hidden">✓</span>
                      <span v-if="isComplete(item)" class="text-[10px] hidden group-hover:inline">✕</span>
                    </button>
                    <span v-else class="w-5 h-5 rounded-full border-2 border-gray-200 flex-shrink-0"></span>
                    <div class="min-w-0">
                      <div class="font-medium truncate">{{ item.practice_name }}</div>
                      <div class="text-xs text-gray-400">
                        {{ scheduleLabel(item) }}
                        <span v-if="!scheduleIsDue(item) && nextDueLabel(item)" class="ml-1 text-gray-300">
                          · {{ nextDueLabel(item) }}
                        </span>
                      </div>
                    </div>
                  </template>
                </div>

                <div class="flex items-center gap-2 flex-shrink-0">
                  <span
                    v-if="item.days_overdue && item.days_overdue > 0"
                    class="text-xs px-2 py-0.5 rounded-full bg-red-100 text-red-600"
                  >{{ item.days_overdue }}d overdue</span>
                  <router-link
                    :to="`/practices/${item.practice_id}/history`"
                    class="text-xs text-gray-400 hover:text-indigo-500"
                  >
                    history
                  </router-link>
                </div>
              </div>
            </template>
          </div>
        </div>
      </div>
    </div>

    <!-- Daily Reflection -->
    <div v-if="!loading" class="mt-6">
      <div class="bg-white rounded-lg shadow overflow-hidden">
        <button
          @click="reflectionExpanded = !reflectionExpanded"
          class="w-full px-4 py-3 flex items-center justify-between text-left hover:bg-gray-50 transition-colors"
        >
          <div class="flex items-center gap-2">
            <span class="text-sm font-semibold text-gray-700">Daily Reflection</span>
            <span v-if="reflection" class="text-xs px-2 py-0.5 rounded-full bg-green-100 text-green-600">saved</span>
            <span v-if="reflection?.mood" class="text-sm">{{ moods.find(m => m.value === reflection!.mood)?.emoji }}</span>
          </div>
          <span class="text-gray-400 text-xs">{{ reflectionExpanded ? '▲' : '▼' }}</span>
        </button>

        <div v-if="reflectionExpanded" class="px-4 pb-4 border-t border-gray-100 pt-3">
          <!-- Prompt -->
          <div v-if="todayPrompt?.text" class="text-sm text-indigo-600 font-medium mb-2 italic">
            {{ todayPrompt.text }}
          </div>

          <!-- Content -->
          <textarea
            ref="reflectionTextarea"
            v-model="reflectionContent"
            rows="3"
            :placeholder="todayPrompt?.text || 'What\'s on your heart today?'"
            class="w-full border border-gray-200 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-300 resize-y mb-3"
          ></textarea>

          <!-- Mood -->
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-1">
              <span class="text-xs text-gray-400 mr-2">Mood</span>
              <button
                v-for="m in moods"
                :key="m.value"
                @click="reflectionMood = reflectionMood === m.value ? null : m.value"
                class="w-8 h-8 rounded-full flex items-center justify-center text-lg transition-all"
                :class="reflectionMood === m.value
                  ? 'bg-indigo-100 ring-2 ring-indigo-300 scale-110'
                  : 'hover:bg-gray-100'"
                :title="m.label"
              >{{ m.emoji }}</button>
            </div>

            <button
              @click="saveReflection"
              :disabled="!reflectionContent.trim() || reflectionSaving"
              class="px-4 py-1.5 bg-indigo-600 text-white rounded text-xs font-medium hover:bg-indigo-700 disabled:opacity-40 transition-colors"
            >{{ reflectionSaving ? 'Saving...' : (reflection ? 'Update' : 'Save') }}</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
