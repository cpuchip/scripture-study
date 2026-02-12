<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { api, type Practice } from '../api'

function localDateStr(d: Date = new Date()): string {
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

const today = localDateStr()
const dueCards = ref<Practice[]>([])
const currentIndex = ref(0)
const flipped = ref(false)
const loading = ref(true)
const sessionComplete = ref(false)
const reviewed = ref(0)

// Parse SM-2 config
function parseConfig(config: string) {
  try { return JSON.parse(config) } catch { return {} }
}

const currentCard = computed(() => dueCards.value[currentIndex.value] || null)

const progress = computed(() => ({
  current: currentIndex.value + 1,
  total: dueCards.value.length,
}))

async function load() {
  loading.value = true
  dueCards.value = await api.getDueCards(today)
  currentIndex.value = 0
  flipped.value = false
  sessionComplete.value = dueCards.value.length === 0
  reviewed.value = 0
  loading.value = false
}

function flip() {
  flipped.value = !flipped.value
}

// Quality labels for SM-2
const qualityOptions = [
  { value: 0, label: 'Blackout', color: 'bg-red-500', desc: 'No memory at all' },
  { value: 1, label: 'Wrong', color: 'bg-red-400', desc: 'Incorrect, recognized answer' },
  { value: 2, label: 'Hard', color: 'bg-orange-400', desc: 'Incorrect, seemed familiar' },
  { value: 3, label: 'Okay', color: 'bg-yellow-400', desc: 'Correct, but difficult' },
  { value: 4, label: 'Good', color: 'bg-green-400', desc: 'Correct after hesitation' },
  { value: 5, label: 'Easy', color: 'bg-green-500', desc: 'Perfect recall' },
]

async function rate(quality: number) {
  if (!currentCard.value) return

  await api.reviewCard(currentCard.value.id, quality, today)
  reviewed.value++
  flipped.value = false

  if (currentIndex.value < dueCards.value.length - 1) {
    currentIndex.value++
  } else {
    sessionComplete.value = true
  }
}

// SM-2 state display
function cardStats(card: Practice) {
  const cfg = parseConfig(card.config)
  return {
    interval: cfg.interval || 0,
    reps: cfg.repetitions || 0,
    ease: (cfg.ease_factor || 2.5).toFixed(2),
    nextReview: cfg.next_review || 'now',
  }
}

onMounted(load)
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold">Memorize</h1>
      <div v-if="!loading && !sessionComplete" class="text-sm text-gray-500">
        {{ progress.current }} / {{ progress.total }} cards
      </div>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="text-center py-12 text-gray-400">Loading cards...</div>

    <!-- Session complete -->
    <div v-else-if="sessionComplete" class="text-center py-12">
      <div class="text-5xl mb-4">🎉</div>
      <h2 class="text-xl font-bold text-gray-800 mb-2">
        {{ reviewed > 0 ? 'Session Complete!' : 'All caught up!' }}
      </h2>
      <p class="text-gray-500 mb-6">
        {{ reviewed > 0
          ? `You reviewed ${reviewed} card${reviewed > 1 ? 's' : ''} today.`
          : 'No cards due for review right now.' }}
      </p>
      <div class="flex gap-3 justify-center">
        <button @click="load" class="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 text-sm">
          Refresh
        </button>
        <router-link to="/practices" class="px-4 py-2 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 text-sm">
          Add cards
        </router-link>
      </div>
    </div>

    <!-- Flashcard -->
    <div v-else-if="currentCard" class="max-w-xl mx-auto">
      <!-- Card -->
      <div
        @click="flip"
        class="cursor-pointer select-none mb-6"
      >
        <div
          class="bg-white rounded-2xl shadow-lg border border-gray-100 p-8 min-h-[280px] flex flex-col items-center justify-center transition-all duration-200"
          :class="flipped ? 'ring-2 ring-indigo-200' : 'hover:shadow-xl'"
        >
          <!-- Front: reference / name -->
          <div v-if="!flipped" class="text-center">
            <div class="text-xs uppercase tracking-wide text-gray-400 mb-3">
              {{ currentCard.category || 'scripture' }}
            </div>
            <div class="text-2xl font-bold text-gray-800 mb-2">
              {{ currentCard.name }}
            </div>
            <div class="text-sm text-gray-400 mt-4">
              tap to reveal
            </div>
          </div>

          <!-- Back: full text -->
          <div v-else class="text-center">
            <div class="text-xs uppercase tracking-wide text-indigo-400 mb-3">
              {{ currentCard.name }}
            </div>
            <div class="text-lg leading-relaxed text-gray-700 whitespace-pre-line">
              {{ currentCard.description || '(no text — add via Practices)' }}
            </div>
          </div>
        </div>
      </div>

      <!-- Rating buttons (visible after flip) -->
      <div v-if="flipped" class="space-y-3">
        <p class="text-center text-sm text-gray-500 mb-2">How well did you remember?</p>
        <div class="grid grid-cols-3 gap-2">
          <button
            v-for="opt in qualityOptions"
            :key="opt.value"
            @click="rate(opt.value)"
            class="flex flex-col items-center p-3 rounded-lg border border-gray-200 hover:border-gray-300 transition-colors text-sm"
          >
            <span
              class="w-8 h-8 rounded-full flex items-center justify-center text-white font-bold text-sm mb-1"
              :class="opt.color"
            >
              {{ opt.value }}
            </span>
            <span class="font-medium text-gray-700">{{ opt.label }}</span>
            <span class="text-[10px] text-gray-400">{{ opt.desc }}</span>
          </button>
        </div>
      </div>

      <!-- Card stats (small, below) -->
      <div class="mt-6 flex justify-center gap-6 text-xs text-gray-400">
        <span>interval: {{ cardStats(currentCard).interval }}d</span>
        <span>reps: {{ cardStats(currentCard).reps }}</span>
        <span>ease: {{ cardStats(currentCard).ease }}</span>
      </div>
    </div>
  </div>
</template>
