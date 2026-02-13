<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { api, type ReportEntry } from '../api'
import TrendLine from '../components/TrendLine.vue'
import SparkLine from '../components/SparkLine.vue'

function localDateStr(d: Date = new Date()): string {
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

const loading = ref(true)
const entries = ref<ReportEntry[]>([])

// Date range
const rangePreset = ref<'7' | '30' | '90' | 'custom'>('30')
const customStart = ref('')
const customEnd = ref('')

const startDate = computed(() => {
  if (rangePreset.value === 'custom') return customStart.value
  const d = new Date()
  d.setDate(d.getDate() - Number(rangePreset.value))
  return localDateStr(d)
})

const endDate = computed(() => {
  if (rangePreset.value === 'custom') return customEnd.value
  return localDateStr()
})

// Filters
const filterType = ref('all')
const filterCategory = ref('all')

const availableTypes = computed(() => {
  const types = new Set(entries.value.map(e => e.practice_type))
  return Array.from(types).sort()
})

const availableCategories = computed(() => {
  const cats = new Set<string>()
  for (const e of entries.value) {
    if (e.category) {
      for (const c of e.category.split(',')) {
        const t = c.trim()
        if (t) cats.add(t)
      }
    }
  }
  return Array.from(cats).sort()
})

const filtered = computed(() => {
  return entries.value.filter(e => {
    if (filterType.value !== 'all' && e.practice_type !== filterType.value) return false
    if (filterCategory.value !== 'all') {
      const cats = (e.category || '').split(',').map(c => c.trim())
      if (!cats.includes(filterCategory.value)) return false
    }
    return true
  })
})

// Summary stats across filtered entries
const summaryStats = computed(() => {
  const f = filtered.value
  const totalPractices = f.length
  const totalLogs = f.reduce((s, e) => s + e.total_logs, 0)
  const totalSets = f.reduce((s, e) => s + e.total_sets, 0)
  const totalReps = f.reduce((s, e) => s + e.total_reps, 0)
  const avgCompletion = f.length > 0
    ? f.reduce((s, e) => s + e.completion_rate, 0) / f.length
    : 0
  const bestStreak = f.reduce((max, e) => Math.max(max, e.current_streak), 0)
  return { totalPractices, totalLogs, totalSets, totalReps, avgCompletion, bestStreak }
})

// --- Overall completion trend ---
// For each date in the range, calculate (practices with activity / total practices)
const overallTrendData = computed(() => {
  const f = filtered.value
  if (f.length === 0) return { dates: [] as string[], values: [] as number[] }

  // Build date range
  const dates: string[] = []
  const d = new Date(startDate.value + 'T12:00:00')
  const end = new Date(endDate.value + 'T12:00:00')
  while (d <= end) {
    dates.push(localDateStr(d))
    d.setDate(d.getDate() + 1)
  }

  // For each practice, build a set of active dates
  const practiceDateSets = f.map(entry => {
    const s = new Set<string>()
    for (const dp of (entry.daily_data || [])) {
      if (dp.logs > 0 || dp.sets > 0) s.add(dp.date)
    }
    return s
  })

  const totalPractices = f.length
  const values = dates.map(date => {
    const active = practiceDateSets.filter(s => s.has(date)).length
    return totalPractices > 0 ? (active / totalPractices) * 100 : 0
  })

  return { dates, values }
})

// Overall trend direction
const overallTrendDirection = computed(() => {
  const vals = overallTrendData.value.values
  if (vals.length < 4) return { arrow: '→', color: 'text-gray-400', delta: 0 }
  const mid = Math.floor(vals.length / 2)
  const first = vals.slice(0, mid)
  const second = vals.slice(mid)
  const avg1 = first.reduce((s, v) => s + v, 0) / first.length
  const avg2 = second.reduce((s, v) => s + v, 0) / second.length
  const delta = Math.round(avg2 - avg1)
  if (delta > 2) return { arrow: '↑', color: 'text-green-600', delta }
  if (delta < -2) return { arrow: '↓', color: 'text-red-500', delta }
  return { arrow: '→', color: 'text-gray-400', delta }
})

// Date labels for the trend chart x-axis
const trendDateLabels = computed(() => {
  const dates = overallTrendData.value.dates
  if (dates.length === 0) return { start: '', mid: '', end: '' }
  const fmt = (ds: string) => new Date(ds + 'T12:00:00').toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
  return {
    start: fmt(dates[0]!),
    mid: fmt(dates[Math.floor(dates.length / 2)]!),
    end: fmt(dates[dates.length - 1]!),
  }
})

// Per-practice sparkline data: daily activity values (filled for the full date range)
function sparklineData(entry: ReportEntry): number[] {
  const map: Record<string, number> = {}
  for (const dp of (entry.daily_data || [])) {
    map[dp.date] = entry.practice_type === 'tracker' ? dp.sets : dp.logs
  }
  const result: number[] = []
  const d = new Date(startDate.value + 'T12:00:00')
  const end = new Date(endDate.value + 'T12:00:00')
  while (d <= end) {
    result.push(map[localDateStr(d)] || 0)
    d.setDate(d.getDate() + 1)
  }
  return result
}

// Chart helpers
function maxDailyValue(entry: ReportEntry): number {
  if (entry.practice_type === 'tracker') {
    return Math.max(...(entry.daily_data || []).map(d => d.sets), 1)
  }
  return Math.max(...(entry.daily_data || []).map(d => d.logs), 1)
}

function barValue(entry: ReportEntry, dp: { logs: number; sets: number; reps: number }): number {
  if (entry.practice_type === 'tracker') return dp.sets
  return dp.logs
}

function barTarget(entry: ReportEntry): number {
  try {
    const cfg = JSON.parse(entry.config)
    if (entry.practice_type === 'tracker') return cfg.target_sets || 1
    if (entry.practice_type === 'memorize') return cfg.target_daily_reps || 1
    return 1
  } catch { return 1 }
}

function barColor(entry: ReportEntry, dp: { logs: number; sets: number }): string {
  const v = barValue(entry, dp as any)
  const t = barTarget(entry)
  if (v >= t) return 'bg-green-400'
  if (v > 0) return 'bg-amber-300'
  return 'bg-gray-100'
}

function typeLabel(entry: ReportEntry): string {
  if (entry.practice_type === 'tracker') {
    try {
      const cfg = JSON.parse(entry.config)
      return `${cfg.target_sets || 1}×${cfg.target_reps || '?'} ${cfg.unit || 'reps'}`
    } catch { return 'tracker' }
  }
  return entry.practice_type
}

function volumeLabel(entry: ReportEntry): string {
  if (entry.practice_type === 'tracker' && entry.total_reps > 0) {
    return `${entry.total_sets} sets · ${entry.total_reps} reps`
  }
  if (entry.total_logs > 0) return `${entry.total_logs} logs`
  return 'no activity'
}

// Fill daily data for all dates in range (so chart has no gaps)
function filledDailyData(entry: ReportEntry): { date: string; logs: number; sets: number; reps: number; label: string }[] {
  const map: Record<string, { logs: number; sets: number; reps: number }> = {}
  for (const dp of (entry.daily_data || [])) {
    map[dp.date] = dp
  }

  const result: { date: string; logs: number; sets: number; reps: number; label: string }[] = []
  const d = new Date(startDate.value + 'T12:00:00')
  const end = new Date(endDate.value + 'T12:00:00')
  while (d <= end) {
    const ds = localDateStr(d)
    const dp = map[ds] || { logs: 0, sets: 0, reps: 0 }
    result.push({
      ...dp,
      date: ds,
      label: d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' }),
    })
    d.setDate(d.getDate() + 1)
  }
  return result
}

async function load() {
  loading.value = true
  try {
    entries.value = await api.getReport(startDate.value, endDate.value)
  } catch (e) {
    console.error('Failed to load report:', e)
  }
  loading.value = false
}

function changePreset(p: '7' | '30' | '90' | 'custom') {
  rangePreset.value = p
  if (p !== 'custom') load()
}

onMounted(load)
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold mb-6">Reports</h1>

    <!-- Date range -->
    <div class="bg-white rounded-lg shadow p-4 mb-6">
      <div class="flex items-center gap-3 flex-wrap">
        <span class="text-sm text-gray-500">Range:</span>
        <button
          v-for="p in [
            { value: '7', label: '7 days' },
            { value: '30', label: '30 days' },
            { value: '90', label: '90 days' },
            { value: 'custom', label: 'Custom' },
          ] as const"
          :key="p.value"
          @click="changePreset(p.value)"
          class="px-3 py-1 text-xs rounded-full border transition-colors"
          :class="rangePreset === p.value
            ? 'bg-indigo-100 border-indigo-300 text-indigo-700'
            : 'bg-gray-50 border-gray-200 text-gray-500 hover:bg-gray-100'"
        >{{ p.label }}</button>

        <template v-if="rangePreset === 'custom'">
          <input v-model="customStart" type="date" class="border rounded px-2 py-1 text-sm" />
          <span class="text-gray-400">→</span>
          <input v-model="customEnd" type="date" class="border rounded px-2 py-1 text-sm" />
          <button @click="load" class="px-3 py-1 text-xs bg-indigo-600 text-white rounded hover:bg-indigo-700">
            Go
          </button>
        </template>
      </div>
    </div>

    <!-- Filters -->
    <div class="mb-4 space-y-2">
      <div class="flex gap-1.5 flex-wrap items-center">
        <span class="text-xs text-gray-400 w-10">Type</span>
        <button
          @click="filterType = 'all'"
          class="px-2.5 py-1 text-xs rounded-full border"
          :class="filterType === 'all' ? 'bg-indigo-100 border-indigo-300 text-indigo-700' : 'bg-gray-50 border-gray-200 text-gray-500 hover:bg-gray-100'"
        >all</button>
        <button
          v-for="t in availableTypes"
          :key="t"
          @click="filterType = t"
          class="px-2.5 py-1 text-xs rounded-full border"
          :class="filterType === t ? 'bg-indigo-100 border-indigo-300 text-indigo-700' : 'bg-gray-50 border-gray-200 text-gray-500 hover:bg-gray-100'"
        >{{ t }}</button>
      </div>
      <div v-if="availableCategories.length > 0" class="flex gap-1.5 flex-wrap items-center">
        <span class="text-xs text-gray-400 w-10">Cat</span>
        <button
          @click="filterCategory = 'all'"
          class="px-2.5 py-1 text-xs rounded-full border"
          :class="filterCategory === 'all' ? 'bg-indigo-100 border-indigo-300 text-indigo-700' : 'bg-gray-50 border-gray-200 text-gray-500 hover:bg-gray-100'"
        >all</button>
        <button
          v-for="c in availableCategories"
          :key="c"
          @click="filterCategory = c"
          class="px-2.5 py-1 text-xs rounded-full border"
          :class="filterCategory === c ? 'bg-indigo-100 border-indigo-300 text-indigo-700' : 'bg-gray-50 border-gray-200 text-gray-500 hover:bg-gray-100'"
        >{{ c }}</button>
      </div>
    </div>

    <div v-if="loading" class="text-center py-8 text-gray-400">Loading...</div>

    <template v-else>
      <!-- Summary cards -->
      <div class="grid grid-cols-3 sm:grid-cols-6 gap-3 mb-6">
        <div class="bg-white rounded-lg shadow p-3 text-center">
          <div class="text-xl font-bold text-indigo-600">{{ summaryStats.totalPractices }}</div>
          <div class="text-[10px] text-gray-500 uppercase">practices</div>
        </div>
        <div class="bg-white rounded-lg shadow p-3 text-center">
          <div class="text-xl font-bold text-green-600">{{ Math.round(summaryStats.avgCompletion * 100) }}%</div>
          <div class="text-[10px] text-gray-500 uppercase">avg completion</div>
        </div>
        <div class="bg-white rounded-lg shadow p-3 text-center">
          <div class="text-xl font-bold text-amber-600">{{ summaryStats.bestStreak }}</div>
          <div class="text-[10px] text-gray-500 uppercase">best streak</div>
        </div>
        <div class="bg-white rounded-lg shadow p-3 text-center">
          <div class="text-xl font-bold text-blue-600">{{ summaryStats.totalLogs }}</div>
          <div class="text-[10px] text-gray-500 uppercase">total logs</div>
        </div>
        <div class="bg-white rounded-lg shadow p-3 text-center">
          <div class="text-xl font-bold text-purple-600">{{ summaryStats.totalSets }}</div>
          <div class="text-[10px] text-gray-500 uppercase">total sets</div>
        </div>
        <div class="bg-white rounded-lg shadow p-3 text-center">
          <div class="text-xl font-bold text-rose-600">{{ summaryStats.totalReps }}</div>
          <div class="text-[10px] text-gray-500 uppercase">total reps</div>
        </div>
      </div>

      <!-- Overall Completion Trend -->
      <div v-if="overallTrendData.values.length > 1" class="bg-white rounded-lg shadow p-4 mb-6">
        <div class="flex items-center justify-between mb-2">
          <h2 class="text-sm font-semibold text-gray-700">Completion Trend</h2>
          <div class="flex items-center gap-2">
            <span class="text-xs text-gray-400">7-day avg</span>
            <span class="text-lg font-bold" :class="overallTrendDirection.color">
              {{ overallTrendDirection.arrow }}
              <span class="text-xs font-normal">{{ overallTrendDirection.delta > 0 ? '+' : '' }}{{ overallTrendDirection.delta }}%</span>
            </span>
          </div>
        </div>
        <TrendLine
          :data="overallTrendData.values"
          :width="800"
          :height="100"
          color="#6366f1"
          :fill="true"
          :rolling-avg="7"
          :stroke-width="1.5"
          class="w-full"
        />
        <div class="flex justify-between mt-1 text-[10px] text-gray-400">
          <span>{{ trendDateLabels.start }}</span>
          <span>{{ trendDateLabels.mid }}</span>
          <span>{{ trendDateLabels.end }}</span>
        </div>
      </div>

      <!-- No data -->
      <div v-if="filtered.length === 0" class="text-center py-12 text-gray-400">
        No practices match filters.
      </div>

      <!-- Practice cards -->
      <div v-else class="space-y-4">
        <div
          v-for="entry in filtered"
          :key="entry.practice_id"
          class="bg-white rounded-lg shadow overflow-hidden"
        >
          <!-- Header -->
          <div class="px-4 py-3 border-b border-gray-100">
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-2">
                <span class="font-semibold">{{ entry.practice_name }}</span>
                <span class="text-xs px-2 py-0.5 rounded-full bg-gray-100 text-gray-600">{{ entry.practice_type }}</span>
                <span
                  v-for="cat in entry.category.split(',').map(c => c.trim()).filter(Boolean)"
                  :key="cat"
                  class="text-xs px-2 py-0.5 rounded-full bg-indigo-50 text-indigo-600"
                >{{ cat }}</span>
              </div>
              <span class="text-xs text-gray-400">{{ typeLabel(entry) }}</span>
            </div>
          </div>

          <!-- Stats row -->
          <div class="px-4 py-2 flex items-center gap-6 text-xs border-b border-gray-50">
            <div>
              <span class="font-semibold text-green-600">{{ Math.round(entry.completion_rate * 100) }}%</span>
              <span class="text-gray-400 ml-1">completion</span>
            </div>
            <div>
              <span class="font-semibold text-indigo-600">{{ entry.days_active }}</span>
              <span class="text-gray-400 ml-1">/ {{ entry.days_in_range }} days active</span>
            </div>
            <div>
              <span class="font-semibold text-amber-600">{{ entry.current_streak }}</span>
              <span class="text-gray-400 ml-1">streak</span>
            </div>
            <div>
              <span class="text-gray-500">{{ volumeLabel(entry) }}</span>
            </div>
            <div class="ml-auto">
              <SparkLine
                :data="sparklineData(entry)"
                :color="entry.completion_rate >= 0.8 ? '#22c55e' : entry.completion_rate >= 0.5 ? '#f59e0b' : '#ef4444'"
                :width="100"
                :height="20"
              />
            </div>
          </div>

          <!-- Chart -->
          <div class="px-4 py-3">
            <div class="flex items-end gap-px h-20">
              <div
                v-for="dp in filledDailyData(entry)"
                :key="dp.date"
                class="flex-1"
                :title="`${dp.label}: ${barValue(entry, dp)}/${barTarget(entry)}`"
              >
                <div
                  class="w-full rounded-t transition-all"
                  :class="barColor(entry, dp)"
                  :style="{
                    height: `${(barValue(entry, dp) / maxDailyValue(entry)) * 100}%`,
                    minHeight: barValue(entry, dp) > 0 ? '3px' : '1px',
                  }"
                ></div>
              </div>
            </div>
            <div class="flex justify-between mt-1 text-[10px] text-gray-400">
              <span>{{ startDate }}</span>
              <span class="text-[9px]">
                <span class="inline-block w-2 h-2 bg-green-400 rounded-sm mr-0.5"></span>target met
                <span class="inline-block w-2 h-2 bg-amber-300 rounded-sm ml-2 mr-0.5"></span>partial
              </span>
              <span>{{ endDate }}</span>
            </div>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>
