<script setup lang="ts">
import { computed } from 'vue'
import type { ActivityDay } from '../api'

const props = withDefaults(defineProps<{
  days: ActivityDay[]
  compact?: boolean       // compact = mini version for DailyView
  navigateToDate?: (date: string) => void
}>(), {
  compact: false,
})

// --- Color scale ---
const colorScale = [
  { min: 0, max: 0, bg: 'bg-gray-100', hex: '#f3f4f6' },
  { min: 1, max: 2, bg: 'bg-orange-100', hex: '#ffedd5' },
  { min: 3, max: 5, bg: 'bg-orange-300', hex: '#fdba74' },
  { min: 6, max: 9, bg: 'bg-orange-400', hex: '#fb923c' },
  { min: 10, max: Infinity, bg: 'bg-orange-600', hex: '#ea580c' },
]

function colorFor(count: number): string {
  for (const c of colorScale) {
    if (count >= c.min && count <= c.max) return c.hex
  }
  return colorScale[0]!.hex
}

// --- Grid layout ---
// Organize days into a week-based grid: rows = day-of-week (0=Sun), columns = weeks
interface GridCell {
  date: string
  logCount: number
  practiceCount: number
  dayOfWeek: number // 0=Sun
  weekIndex: number
  month: number
  isToday: boolean
}

const todayStr = new Date().toISOString().slice(0, 10)

const grid = computed(() => {
  if (props.days.length === 0) return { cells: [], weeks: 0, monthLabels: [] as { label: string; weekIndex: number }[] }

  // Build a lookup
  const lookup = new Map<string, ActivityDay>()
  for (const d of props.days) lookup.set(d.date, d)

  // Find the full range from first day to last day
  const firstDate = new Date(props.days[0]!.date + 'T12:00:00')
  const lastDate = new Date(props.days[props.days.length - 1]!.date + 'T12:00:00')

  // Extend back to the previous Sunday to start cleanly
  const startDate = new Date(firstDate)
  startDate.setDate(startDate.getDate() - startDate.getDay())

  const cells: GridCell[] = []
  let weekIndex = 0
  const d = new Date(startDate)

  // Track month labels
  const monthLabels: { label: string; weekIndex: number }[] = []
  let lastMonth = -1

  while (d <= lastDate) {
    const ds = d.toISOString().slice(0, 10)
    // Avoid artificial padding—only show days in the actual data range
    // but pad the first partial week with empty cells
    const dayOfWeek = d.getDay()

    if (dayOfWeek === 0 && cells.length > 0) weekIndex++

    // Month labels
    const m = d.getMonth()
    if (m !== lastMonth) {
      monthLabels.push({
        label: d.toLocaleDateString('en-US', { month: 'short' }),
        weekIndex,
      })
      lastMonth = m
    }

    const ad = lookup.get(ds)
    const inRange = d >= firstDate
    cells.push({
      date: ds,
      logCount: inRange && ad ? ad.log_count : -1, // -1 = out of range (padding)
      practiceCount: inRange && ad ? ad.practice_count : 0,
      dayOfWeek,
      weekIndex,
      month: m,
      isToday: ds === todayStr,
    })

    d.setDate(d.getDate() + 1)
  }

  return { cells, weeks: weekIndex + 1, monthLabels }
})

// Compute current streak from today backwards
const currentStreak = computed(() => {
  const lookup = new Map<string, number>()
  for (const d of props.days) lookup.set(d.date, d.log_count)

  let streak = 0
  const d = new Date(todayStr + 'T12:00:00')
  // If today has no activity, start from yesterday
  if (!lookup.get(todayStr) || lookup.get(todayStr) === 0) {
    d.setDate(d.getDate() - 1)
  }
  for (let i = 0; i < 365; i++) {
    const ds = d.toISOString().slice(0, 10)
    if ((lookup.get(ds) || 0) > 0) {
      streak++
      d.setDate(d.getDate() - 1)
    } else {
      break
    }
  }
  return streak
})

// Total active days
const totalActive = computed(() => props.days.filter(d => d.log_count > 0).length)

// Total logs
const totalLogs = computed(() => props.days.reduce((s, d) => s + d.log_count, 0))

const dayLabels = ['', 'Mon', '', 'Wed', '', 'Fri', '']

function tooltipText(cell: GridCell): string {
  if (cell.logCount < 0) return '' // padding
  const d = new Date(cell.date + 'T12:00:00')
  const dateStr = d.toLocaleDateString('en-US', { weekday: 'short', month: 'short', day: 'numeric' })
  if (cell.logCount === 0) return `${dateStr} — no activity`
  return `${dateStr} — ${cell.logCount} log${cell.logCount !== 1 ? 's' : ''} across ${cell.practiceCount} practice${cell.practiceCount !== 1 ? 's' : ''}`
}

function handleClick(cell: GridCell) {
  if (cell.logCount < 0) return
  if (props.navigateToDate) props.navigateToDate(cell.date)
}
</script>

<template>
  <!-- Compact version for DailyView -->
  <div v-if="compact" class="bg-white rounded-lg shadow p-3">
    <div class="flex items-center justify-between mb-2">
      <span class="text-xs font-semibold text-gray-600">Activity</span>
      <div class="flex items-center gap-3 text-xs text-gray-500">
        <span>🔥 {{ currentStreak }} day streak</span>
        <span>{{ totalActive }} active days</span>
      </div>
    </div>
    <div class="overflow-x-auto">
      <div class="flex gap-px" style="min-width: max-content">
        <div
          v-for="cell in grid.cells.filter(c => c.logCount >= 0)"
          :key="cell.date"
          class="rounded-sm cursor-pointer transition-transform hover:scale-125"
          :class="[cell.isToday ? 'ring-1 ring-indigo-400' : '']"
          :style="{
            width: '10px',
            height: '10px',
            backgroundColor: colorFor(cell.logCount),
          }"
          :title="tooltipText(cell)"
          @click="handleClick(cell)"
        />
      </div>
    </div>
    <!-- Legend -->
    <div class="flex items-center justify-end gap-1 mt-1.5">
      <span class="text-[9px] text-gray-400">Less</span>
      <div v-for="c in colorScale" :key="c.min"
        class="rounded-sm"
        :style="{ width: '8px', height: '8px', backgroundColor: c.hex }" />
      <span class="text-[9px] text-gray-400">More</span>
    </div>
  </div>

  <!-- Full version for ReportsView -->
  <div v-else class="bg-white rounded-lg shadow p-4">
    <div class="flex items-center justify-between mb-3">
      <h2 class="text-sm font-semibold text-gray-700">Activity</h2>
      <div class="flex items-center gap-4 text-xs text-gray-500">
        <span>🔥 <span class="font-semibold text-orange-600">{{ currentStreak }}</span> day streak</span>
        <span><span class="font-semibold text-indigo-600">{{ totalActive }}</span> active days</span>
        <span><span class="font-semibold text-green-600">{{ totalLogs }}</span> total logs</span>
      </div>
    </div>

    <div class="overflow-x-auto">
      <div class="flex">
        <!-- Day-of-week labels -->
        <div class="flex flex-col gap-px mr-1.5 pt-4">
          <div v-for="(label, i) in dayLabels" :key="i"
            class="text-[9px] text-gray-400 leading-none flex items-center"
            :style="{ height: '13px' }">
            {{ label }}
          </div>
        </div>

        <div>
          <!-- Month labels -->
          <div class="flex mb-1" :style="{ height: '12px' }">
          </div>

          <!-- Grid -->
          <div class="relative">
            <!-- Month labels row - absolutely positioned -->
            <div class="flex gap-px mb-1" style="min-width: max-content">
              <template v-for="(ml, idx) in grid.monthLabels" :key="idx">
                <span
                  class="text-[9px] text-gray-400 absolute"
                  :style="{ left: ml.weekIndex * 15 + 'px' }"
                >{{ ml.label }}</span>
              </template>
            </div>
            <div class="pt-3.5">
              <div class="grid gap-px" :style="{
                gridTemplateRows: 'repeat(7, 13px)',
                gridAutoFlow: 'column',
                gridAutoColumns: '13px',
                width: 'max-content',
              }">
                <div
                  v-for="cell in grid.cells"
                  :key="cell.date"
                  class="rounded-sm transition-transform"
                  :class="[
                    cell.logCount >= 0 ? 'cursor-pointer hover:scale-125' : 'opacity-0',
                    cell.isToday ? 'ring-1 ring-indigo-400 ring-offset-1' : '',
                  ]"
                  :style="{
                    width: '12px',
                    height: '12px',
                    backgroundColor: cell.logCount >= 0 ? colorFor(cell.logCount) : 'transparent',
                  }"
                  :title="tooltipText(cell)"
                  @click="handleClick(cell)"
                />
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Legend -->
    <div class="flex items-center justify-between mt-3">
      <span class="text-[10px] text-gray-400">Click a day to see details</span>
      <div class="flex items-center gap-1">
        <span class="text-[10px] text-gray-400">Less</span>
        <div v-for="c in colorScale" :key="c.min"
          class="rounded-sm"
          :style="{ width: '10px', height: '10px', backgroundColor: c.hex }" />
        <span class="text-[10px] text-gray-400">More</span>
      </div>
    </div>
  </div>
</template>
