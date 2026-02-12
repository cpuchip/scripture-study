<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(defineProps<{
  data: number[]
  color?: string
  width?: number
  height?: number
}>(), {
  color: '#6366f1',
  width: 120,
  height: 24,
})

const points = computed(() => {
  const vals = props.data
  if (vals.length === 0) return ''
  const max = Math.max(...vals, 0.01)
  const min = Math.min(...vals, 0)
  const range = max - min || 1
  const stepX = props.width / Math.max(vals.length - 1, 1)
  const pad = 2

  return vals
    .map((v, i) => {
      const x = i * stepX
      const y = pad + (props.height - 2 * pad) - ((v - min) / range) * (props.height - 2 * pad)
      return `${x},${y}`
    })
    .join(' ')
})

/** Trend direction: compare first half avg to second half avg */
const trend = computed(() => {
  const vals = props.data
  if (vals.length < 4) return 'flat'
  const mid = Math.floor(vals.length / 2)
  const firstHalf = vals.slice(0, mid)
  const secondHalf = vals.slice(mid)
  const avg1 = firstHalf.reduce((s, v) => s + v, 0) / firstHalf.length
  const avg2 = secondHalf.reduce((s, v) => s + v, 0) / secondHalf.length
  const diff = avg2 - avg1
  const threshold = Math.max(avg1, avg2) * 0.05 || 0.01
  if (diff > threshold) return 'up'
  if (diff < -threshold) return 'down'
  return 'flat'
})

const trendArrow = computed(() => {
  if (trend.value === 'up') return '↑'
  if (trend.value === 'down') return '↓'
  return '→'
})

const trendColor = computed(() => {
  if (trend.value === 'up') return 'text-green-600'
  if (trend.value === 'down') return 'text-red-500'
  return 'text-gray-400'
})
</script>

<template>
  <span class="inline-flex items-center gap-1.5">
    <svg
      :viewBox="`0 0 ${width} ${height}`"
      :width="width"
      :height="height"
      class="overflow-visible"
    >
      <polyline
        v-if="points"
        :points="points"
        fill="none"
        :stroke="color"
        stroke-width="1.5"
        stroke-linejoin="round"
        stroke-linecap="round"
      />
    </svg>
    <span class="text-xs font-bold" :class="trendColor">{{ trendArrow }}</span>
  </span>
</template>
