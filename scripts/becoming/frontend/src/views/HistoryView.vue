<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { api, type Practice, type PracticeLog } from '../api'

const route = useRoute()
const practiceId = Number(route.params.id)

const practice = ref<Practice | null>(null)
const logs = ref<PracticeLog[]>([])
const loading = ref(true)

// Date range for chart (last 30 days)
const days = 30

function localDateStr(d: Date = new Date()): string {
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

function getDateRange() {
  const end = new Date()
  const start = new Date()
  start.setDate(start.getDate() - days)
  return {
    start: localDateStr(start),
    end: localDateStr(end),
  }
}

// Build chart data: one bar per day
const chartData = computed(() => {
  const range = getDateRange()
  const result: { date: string; label: string; value: number; target: number }[] = []

  const config = practice.value ? parseConfig(practice.value.config) : {}
  const targetSets = config.target_sets || 1

  // Create date -> log count map
  const logsByDate: Record<string, number> = {}
  for (const log of logs.value) {
    logsByDate[log.date] = (logsByDate[log.date] || 0) + (log.sets || 1)
  }

  // Fill in all days
  const d = new Date(range.start + 'T12:00:00')
  const end = new Date(range.end + 'T12:00:00')
  while (d <= end) {
    const dateStr = localDateStr(d)
    result.push({
      date: dateStr,
      label: d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' }),
      value: logsByDate[dateStr] || 0,
      target: practice.value?.type === 'tracker' ? targetSets : 1,
    })
    d.setDate(d.getDate() + 1)
  }

  return result
})

const maxValue = computed(() => {
  return Math.max(...chartData.value.map(d => Math.max(d.value, d.target)), 1)
})

// Streak: consecutive days with at least one log
const streak = computed(() => {
  let count = 0
  const today = new Date()
  for (let i = 0; i < 365; i++) {
    const d = new Date(today)
    d.setDate(d.getDate() - i)
    const dateStr = localDateStr(d)
    const hasLog = logs.value.some(l => l.date === dateStr)
    if (hasLog) {
      count++
    } else if (i > 0) {
      break // skip today if not done yet
    }
  }
  return count
})

const totalLogs = computed(() => logs.value.length)

function parseConfig(config: string) {
  try { return JSON.parse(config) } catch { return {} }
}

function formatLogDate(date: string): string {
  const d = new Date(date + 'T12:00:00')
  return d.toLocaleDateString('en-US', { weekday: 'short', month: 'short', day: 'numeric' })
}

async function load() {
  loading.value = true
  const range = getDateRange()
  const [p, l] = await Promise.all([
    api.getPractice(practiceId),
    api.listPracticeLogsRange(practiceId, range.start, range.end),
  ])
  practice.value = p
  logs.value = l
  loading.value = false
}

onMounted(load)
</script>

<template>
  <div>
    <router-link to="/" class="text-sm text-gray-400 hover:text-indigo-500 mb-4 inline-block">
      ← Back to today
    </router-link>

    <div v-if="loading" class="text-center py-8 text-gray-400">Loading...</div>

    <div v-else-if="practice">
      <div class="flex items-center gap-3 mb-6">
        <h1 class="text-2xl font-bold">{{ practice.name }}</h1>
        <span class="text-xs px-2 py-0.5 rounded-full bg-gray-100 text-gray-600">{{ practice.type }}</span>
        <span v-if="practice.category" class="text-xs px-2 py-0.5 rounded-full bg-indigo-50 text-indigo-600">{{ practice.category }}</span>
      </div>

      <p v-if="practice.description" class="text-sm text-gray-600 mb-6">{{ practice.description }}</p>

      <!-- Stats -->
      <div class="grid grid-cols-3 gap-4 mb-6">
        <div class="bg-white rounded-lg shadow p-4 text-center">
          <div class="text-2xl font-bold text-indigo-600">{{ streak }}</div>
          <div class="text-xs text-gray-500">day streak</div>
        </div>
        <div class="bg-white rounded-lg shadow p-4 text-center">
          <div class="text-2xl font-bold text-green-600">{{ totalLogs }}</div>
          <div class="text-xs text-gray-500">total logs ({{ days }}d)</div>
        </div>
        <div class="bg-white rounded-lg shadow p-4 text-center">
          <div class="text-2xl font-bold text-amber-600">
            {{ chartData.filter(d => d.value >= d.target).length }}
          </div>
          <div class="text-xs text-gray-500">days completed</div>
        </div>
      </div>

      <!-- Chart -->
      <div class="bg-white rounded-lg shadow p-4 mb-6">
        <h2 class="text-sm font-semibold text-gray-700 mb-3">Last {{ days }} Days</h2>
        <div class="flex items-end gap-px h-32">
          <div
            v-for="day in chartData"
            :key="day.date"
            class="flex-1 flex flex-col items-center justify-end"
            :title="`${day.label}: ${day.value}/${day.target}`"
          >
            <div
              class="w-full rounded-t transition-all"
              :class="day.value >= day.target ? 'bg-green-400' : day.value > 0 ? 'bg-amber-300' : 'bg-gray-100'"
              :style="{ height: `${(day.value / maxValue) * 100}%`, minHeight: day.value > 0 ? '4px' : '2px' }"
            ></div>
          </div>
        </div>
        <div class="flex justify-between mt-1 text-[10px] text-gray-400">
          <span>{{ chartData[0]?.label }}</span>
          <span>{{ chartData[chartData.length - 1]?.label }}</span>
        </div>
      </div>

      <!-- Recent logs -->
      <div class="bg-white rounded-lg shadow">
        <h2 class="text-sm font-semibold text-gray-700 px-4 pt-4 pb-2">Recent Activity</h2>
        <div class="divide-y divide-gray-100">
          <div
            v-for="log in [...logs].reverse().slice(0, 20)"
            :key="log.id"
            class="px-4 py-2 flex items-center justify-between text-sm"
          >
            <span class="text-gray-600">{{ formatLogDate(log.date) }}</span>
            <div class="flex items-center gap-3 text-gray-500 text-xs">
              <span v-if="log.sets">{{ log.sets }} sets</span>
              <span v-if="log.reps">{{ log.reps }} reps</span>
              <span v-if="log.value">{{ log.value }}</span>
              <span v-if="log.notes" class="text-gray-400 italic truncate max-w-48">{{ log.notes }}</span>
            </div>
          </div>
          <div v-if="logs.length === 0" class="px-4 py-6 text-center text-gray-400 text-sm">
            No activity yet in the last {{ days }} days.
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
