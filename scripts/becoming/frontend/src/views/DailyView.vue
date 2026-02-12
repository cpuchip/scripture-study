<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { api, type DailySummary, type PracticeLog } from '../api'

function localDateStr(d: Date = new Date()): string {
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

const today = ref(localDateStr())
const summary = ref<DailySummary[]>([])
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
    summary.value = await api.getDailySummary(today.value)
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

  if (item.practice_type === 'exercise') {
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
function exerciseProgress(item: DailySummary): string {
  const config = parseConfig(item.config)
  const targetSets = config.target_sets || 1
  const targetReps = config.target_reps
  const done = item.total_sets || item.log_count
  let text = `${done}/${targetSets} sets`
  if (targetReps) text += ` × ${targetReps} ${config.unit || 'reps'}`
  return text
}

function isComplete(item: DailySummary): boolean {
  if (item.practice_type === 'exercise') {
    const config = parseConfig(item.config)
    const targetSets = config.target_sets || 1
    const done = item.total_sets || item.log_count
    return done >= targetSets
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

onMounted(load)
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
            class="flex items-center justify-between px-4 py-3"
            :class="{ 'opacity-60': isComplete(item) }"
          >
            <div class="flex items-center gap-3 flex-1 min-w-0">
              <!-- Completion indicator -->
              <button
                @click="isComplete(item) ? undoLog(item) : quickLog(item)"
                class="w-6 h-6 rounded-full border-2 flex items-center justify-center flex-shrink-0 transition-colors"
                :class="isComplete(item)
                  ? 'bg-green-500 border-green-500 text-white'
                  : 'border-gray-300 hover:border-indigo-400'"
              >
                <span v-if="isComplete(item)" class="text-xs">✓</span>
              </button>

              <div class="min-w-0">
                <div class="font-medium truncate">{{ item.practice_name }}</div>
                <div v-if="item.practice_type === 'exercise'" class="text-xs text-gray-500">
                  {{ exerciseProgress(item) }}
                </div>
                <div v-if="item.last_value" class="text-xs text-gray-400">
                  {{ item.last_value }}
                </div>
              </div>
            </div>

            <!-- Quick add button for exercises (add another set) -->
            <button
              v-if="item.practice_type === 'exercise' && !isComplete(item)"
              @click="quickLog(item)"
              class="ml-2 px-3 py-1 text-xs bg-indigo-50 text-indigo-600 rounded hover:bg-indigo-100"
            >
              + set
            </button>

            <!-- History link -->
            <router-link
              :to="`/practices/${item.practice_id}/history`"
              class="ml-2 text-xs text-gray-400 hover:text-indigo-500"
            >
              history
            </router-link>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
