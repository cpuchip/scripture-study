<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(defineProps<{
  /** Array of numeric values to plot */
  data: number[]
  /** SVG width */
  width?: number
  /** SVG height */
  height?: number
  /** Stroke color */
  color?: string
  /** Fill area under the line */
  fill?: boolean
  /** Show 7-day rolling average overlay */
  rollingAvg?: number
  /** Show dots at data points */
  showDots?: boolean
  /** Stroke width */
  strokeWidth?: number
}>(), {
  width: 400,
  height: 80,
  color: '#6366f1',
  fill: true,
  rollingAvg: 0,
  showDots: false,
  strokeWidth: 2,
})

const padding = { top: 4, right: 4, bottom: 4, left: 4 }

const chartWidth = computed(() => props.width - padding.left - padding.right)
const chartHeight = computed(() => props.height - padding.top - padding.bottom)

function clampMin(arr: number[]): { min: number; max: number } {
  const max = Math.max(...arr, 0.01)
  const min = Math.min(...arr, 0)
  return { min, max }
}

function toPoints(values: number[]): string {
  if (values.length === 0) return ''
  const { min, max } = clampMin(values)
  const range = max - min || 1
  const stepX = chartWidth.value / Math.max(values.length - 1, 1)

  return values
    .map((v, i) => {
      const x = padding.left + i * stepX
      const y = padding.top + chartHeight.value - ((v - min) / range) * chartHeight.value
      return `${x},${y}`
    })
    .join(' ')
}

function toFillPath(values: number[]): string {
  if (values.length === 0) return ''
  const { min, max } = clampMin(values)
  const range = max - min || 1
  const stepX = chartWidth.value / Math.max(values.length - 1, 1)

  const points = values.map((v, i) => {
    const x = padding.left + i * stepX
    const y = padding.top + chartHeight.value - ((v - min) / range) * chartHeight.value
    return { x, y }
  })

  const bottom = padding.top + chartHeight.value
  const first = points[0]!
  const last = points[points.length - 1]!
  let d = `M ${first.x},${bottom}`
  for (const p of points) d += ` L ${p.x},${p.y}`
  d += ` L ${last.x},${bottom} Z`
  return d
}

function rollingAverage(values: number[], window: number): number[] {
  return values.map((_, i) => {
    const start = Math.max(0, i - window + 1)
    const slice = values.slice(start, i + 1)
    return slice.reduce((s, v) => s + v, 0) / slice.length
  })
}

const rawPoints = computed(() => toPoints(props.data))
const rawFill = computed(() => toFillPath(props.data))

const avgData = computed(() => {
  if (props.rollingAvg <= 0) return []
  return rollingAverage(props.data, props.rollingAvg)
})
const avgPoints = computed(() => toPoints(avgData.value))

const dotPositions = computed(() => {
  if (!props.showDots || props.data.length === 0) return []
  const { min, max } = clampMin(props.data)
  const range = max - min || 1
  const stepX = chartWidth.value / Math.max(props.data.length - 1, 1)
  return props.data.map((v, i) => ({
    x: padding.left + i * stepX,
    y: padding.top + chartHeight.value - ((v - min) / range) * chartHeight.value,
  }))
})
</script>

<template>
  <svg
    :viewBox="`0 0 ${width} ${height}`"
    :width="width"
    :height="height"
    class="overflow-visible"
    preserveAspectRatio="none"
  >
    <!-- Fill area -->
    <path
      v-if="fill && rawFill"
      :d="rawFill"
      :fill="color"
      fill-opacity="0.1"
    />

    <!-- Raw data line (faded if rolling avg is shown) -->
    <polyline
      v-if="rawPoints"
      :points="rawPoints"
      fill="none"
      :stroke="color"
      :stroke-opacity="rollingAvg > 0 ? 0.25 : 1"
      :stroke-width="strokeWidth"
      stroke-linejoin="round"
      stroke-linecap="round"
    />

    <!-- Rolling average line -->
    <polyline
      v-if="rollingAvg > 0 && avgPoints"
      :points="avgPoints"
      fill="none"
      :stroke="color"
      :stroke-width="strokeWidth + 0.5"
      stroke-linejoin="round"
      stroke-linecap="round"
    />

    <!-- Dots -->
    <circle
      v-for="(dot, i) in dotPositions"
      :key="i"
      :cx="dot.x"
      :cy="dot.y"
      :r="strokeWidth"
      :fill="color"
    />
  </svg>
</template>
