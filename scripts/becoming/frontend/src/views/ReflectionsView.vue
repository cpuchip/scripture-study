<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { api, type Reflection } from '../api'

const reflections = ref<Reflection[]>([])
const loading = ref(true)
const searchQuery = ref('')
const filterMood = ref<number | null>(null)

const moods = [
  { value: 1, emoji: '😟', label: 'Struggling' },
  { value: 2, emoji: '😐', label: 'Meh' },
  { value: 3, emoji: '🙂', label: 'Okay' },
  { value: 4, emoji: '😊', label: 'Good' },
  { value: 5, emoji: '😄', label: 'Great' },
]

const filtered = computed(() => {
  let result = reflections.value

  if (filterMood.value) {
    result = result.filter(r => r.mood === filterMood.value)
  }

  if (searchQuery.value.trim()) {
    const q = searchQuery.value.toLowerCase()
    result = result.filter(r =>
      r.content.toLowerCase().includes(q) ||
      (r.prompt_text && r.prompt_text.toLowerCase().includes(q))
    )
  }

  return result
})

// Extract YYYY-MM-DD from various date formats (handles both "2026-02-12" and "2026-02-12T00:00:00Z")
function dateOnly(dateStr: string): string {
  return dateStr.substring(0, 10)
}

// Group reflections by month
const groupedByMonth = computed(() => {
  const groups: Record<string, Reflection[]> = {}
  for (const r of filtered.value) {
    const d = new Date(dateOnly(r.date) + 'T12:00:00')
    const key = d.toLocaleDateString('en-US', { year: 'numeric', month: 'long' })
    if (!groups[key]) groups[key] = []
    groups[key]!.push(r)
  }
  return groups
})

function formatDate(dateStr: string): string {
  const d = new Date(dateOnly(dateStr) + 'T12:00:00')
  return d.toLocaleDateString('en-US', { weekday: 'short', month: 'short', day: 'numeric' })
}

function moodEmoji(mood?: number | null): string {
  if (!mood) return ''
  return moods.find(m => m.value === mood)?.emoji || ''
}

// Mood distribution for the summary bar
const moodDistribution = computed(() => {
  const counts: Record<number, number> = {}
  let total = 0
  for (const r of reflections.value) {
    if (r.mood) {
      counts[r.mood] = (counts[r.mood] || 0) + 1
      total++
    }
  }
  return moods.map(m => ({
    ...m,
    count: counts[m.value] || 0,
    pct: total > 0 ? Math.round(((counts[m.value] || 0) / total) * 100) : 0,
  }))
})

async function load() {
  loading.value = true
  try {
    reflections.value = await api.listReflections()
  } catch (e) {
    console.error('Failed to load reflections:', e)
  }
  loading.value = false
}

onMounted(load)
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold">Reflections</h1>
      <router-link
        to="/today?reflect=1"
        class="px-4 py-2 bg-indigo-600 text-white rounded-lg text-sm hover:bg-indigo-700"
      >
        ✏️ Write Today's Reflection
      </router-link>
    </div>

    <!-- Search -->
    <div class="mb-4">
      <input
        v-model="searchQuery"
        type="text"
        placeholder="Search reflections..."
        class="w-full border border-gray-200 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-300"
      />
    </div>

    <!-- Mood filter -->
    <div class="flex gap-1.5 flex-wrap items-center mb-4">
      <span class="text-xs text-gray-400">Mood</span>
      <button
        @click="filterMood = null"
        class="px-2.5 py-1 text-xs rounded-full border transition-colors"
        :class="filterMood === null
          ? 'bg-indigo-100 border-indigo-300 text-indigo-700'
          : 'bg-gray-50 border-gray-200 text-gray-500 hover:bg-gray-100'"
      >All</button>
      <button
        v-for="m in moods"
        :key="m.value"
        @click="filterMood = filterMood === m.value ? null : m.value"
        class="px-2.5 py-1 text-xs rounded-full border transition-colors"
        :class="filterMood === m.value
          ? 'bg-indigo-100 border-indigo-300 text-indigo-700'
          : 'bg-gray-50 border-gray-200 text-gray-500 hover:bg-gray-100'"
      >{{ m.emoji }} {{ m.label }}</button>
    </div>

    <div v-if="loading" class="text-center py-8 text-gray-400">Loading...</div>

    <template v-else>
      <!-- Mood summary bar -->
      <div v-if="reflections.length > 0 && moodDistribution.some(m => m.count > 0)" class="bg-white rounded-lg shadow p-4 mb-6">
        <div class="text-xs text-gray-500 mb-2">Mood Distribution ({{ reflections.length }} reflections)</div>
        <div class="flex rounded-full overflow-hidden h-3 bg-gray-100">
          <div
            v-for="m in moodDistribution.filter(m => m.count > 0)"
            :key="m.value"
            class="transition-all"
            :style="{ width: m.pct + '%' }"
            :class="{
              'bg-red-300': m.value === 1,
              'bg-yellow-300': m.value === 2,
              'bg-blue-300': m.value === 3,
              'bg-green-300': m.value === 4,
              'bg-emerald-400': m.value === 5,
            }"
            :title="`${m.emoji} ${m.label}: ${m.count} (${m.pct}%)`"
          ></div>
        </div>
        <div class="flex justify-between mt-1 text-[10px] text-gray-400">
          <div v-for="m in moodDistribution.filter(m => m.count > 0)" :key="m.value">
            {{ m.emoji }} {{ m.count }}
          </div>
        </div>
      </div>

      <!-- Empty state -->
      <div v-if="filtered.length === 0" class="text-center py-12 text-gray-400">
        {{ searchQuery || filterMood ? 'No matching reflections.' : 'No reflections yet. Write one from the Today page.' }}
      </div>

      <!-- Grouped by month -->
      <div v-for="(group, month) in groupedByMonth" :key="month" class="mb-6">
        <h2 class="text-sm font-semibold text-gray-500 mb-2">{{ month }}</h2>
        <div class="space-y-2">
          <div
            v-for="r in group"
            :key="r.id"
            class="bg-white rounded-lg shadow px-4 py-3"
          >
            <div class="flex items-start justify-between gap-3">
              <div class="flex-1 min-w-0">
                <!-- Date and mood -->
                <div class="flex items-center gap-2 mb-1">
                  <span class="text-xs font-semibold text-gray-600">{{ formatDate(r.date) }}</span>
                  <span v-if="r.mood" class="text-sm">{{ moodEmoji(r.mood) }}</span>
                </div>

                <!-- Prompt -->
                <div v-if="r.prompt_text" class="text-xs text-indigo-500 italic mb-1">
                  {{ r.prompt_text }}
                </div>

                <!-- Content -->
                <div class="text-sm text-gray-800 whitespace-pre-wrap">{{ r.content }}</div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>
