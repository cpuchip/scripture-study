<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { api, type DailySummary, type PracticeLog, type Practice } from '../api'

function localDateStr(d: Date = new Date()): string {
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

const today = ref(localDateStr())
const summary = ref<DailySummary[]>([])
const dueCards = ref<Practice[]>([])
const loading = ref(true)

// Parse exercise config
function parseConfig(config: string) {
  try { return JSON.parse(config) } catch { return {} }
}

// Group by category
const grouped = computed(() => {
  const groups: Record<string, DailySummary[]> = {}
  for (const item of summary.value) {
    const cat = item.category || item.practice_type || 'other'
    if (!groups[cat]) groups[cat] = []
    groups[cat].push(item)
  }
  return groups
})

async function load() {
  loading.value = true
  try {
    const [s, d] = await Promise.all([
      api.getDailySummary(today.value),
      api.getDueCards(today.value),
    ])
    summary.value = s
    dueCards.value = d
  } catch (e) {
    console.error('Failed to load daily summary:', e)
  }
  loading.value = false
}

// Quick log: just mark it done (habit) or log a set (exercise)
async function quickLog(item: DailySummary) {
  const config = parseConfig(item.config)
  const log: Partial<PracticeLog> = {
    practice_id: item.practice_id,
    date: today.value,
  }

  if (item.practice_type === 'tracker') {
    log.sets = 1
    log.reps = config.target_reps || undefined
  }

  await api.createLog(log)
  await load()
}

// Undo: remove last log for today
async function undoLog(item: DailySummary) {
  // Get logs for this practice today and delete the most recent
  const logs = await api.listPracticeLogs(item.practice_id, 50)
  const todayLog = logs.find(l => l.date === today.value)
  if (todayLog) {
    await api.deleteLog(todayLog.id)
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

// Completion stats
const completionStats = computed(() => {
  const total = summary.value.length
  const done = summary.value.filter(s => s.log_count > 0).length
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
  return item.log_count > 0
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

onMounted(() => {
  load()
  document.addEventListener('visibilitychange', onVisibilityChange)
})

onUnmounted(() => {
  document.removeEventListener('visibilitychange', onVisibilityChange)
})
</script>

<template>
  <div>
    <!-- Date header -->
    <div class="flex items-center justify-between mb-6">
      <button @click="prevDay" class="p-2 hover:bg-gray-200 rounded">←</button>
      <div class="text-center">
        <h1 class="text-2xl font-bold">{{ formatDate(today) }}</h1>
        <p class="text-sm text-gray-500">
          {{ completionStats.done }}/{{ completionStats.total }} completed
        </p>
      </div>
      <button @click="nextDay" class="p-2 hover:bg-gray-200 rounded">→</button>
    </div>

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

    <!-- Practice groups -->
    <div v-else class="space-y-6">
      <div v-for="(items, category) in grouped" :key="category">
        <h2 class="text-sm font-semibold text-gray-500 uppercase tracking-wide mb-2">
          {{ category }}
        </h2>
        <div class="bg-white rounded-lg shadow divide-y divide-gray-100">
          <div
            v-for="item in items"
            :key="item.practice_id"
            class="px-4 py-2"
            :class="{ 'opacity-60': isComplete(item) }"
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
                    class="flex items-center gap-1 px-2 py-1 rounded border text-xs transition-colors"
                    :class="setNum <= trackerCompletedSets(item)
                      ? 'bg-green-50 border-green-300 text-green-700'
                      : 'bg-white border-gray-200 text-gray-500 hover:border-indigo-300 hover:bg-indigo-50'"
                  >
                    <span
                      class="w-3.5 h-3.5 rounded border flex items-center justify-center text-[9px]"
                      :class="setNum <= trackerCompletedSets(item)
                        ? 'bg-green-500 border-green-500 text-white'
                        : 'border-gray-300'"
                    >
                      <span v-if="setNum <= trackerCompletedSets(item)">✓</span>
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
                    class="w-5 h-5 rounded-full border-2 flex items-center justify-center flex-shrink-0 transition-colors"
                    :class="isComplete(item)
                      ? 'bg-green-500 border-green-500 text-white'
                      : 'border-gray-300 hover:border-indigo-400'"
                  >
                    <span v-if="isComplete(item)" class="text-[10px]">✓</span>
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
            <template v-else>
              <div class="flex items-center justify-between">
                <div class="flex items-center gap-3 flex-1 min-w-0">
                  <button
                    @click="isComplete(item) ? undoLog(item) : quickLog(item)"
                    class="w-5 h-5 rounded-full border-2 flex items-center justify-center flex-shrink-0 transition-colors"
                    :class="isComplete(item)
                      ? 'bg-green-500 border-green-500 text-white'
                      : 'border-gray-300 hover:border-indigo-400'"
                  >
                    <span v-if="isComplete(item)" class="text-[10px]">✓</span>
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
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
