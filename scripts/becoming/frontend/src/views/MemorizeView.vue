<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { api, type Practice } from '../api'

const route = useRoute()

function localDateStr(d: Date = new Date()): string {
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

const today = localDateStr()
const cards = ref<Practice[]>([])
const currentIndex = ref(0)
const loading = ref(true)
const sessionComplete = ref(false)
const reviewed = ref(0)

// Mode: review (tap-to-flip), practice (fill blanks), quiz (type it out)
type Mode = 'review' | 'practice' | 'quiz'
const mode = ref<Mode>('review')

// --- Review mode state ---
const flipped = ref(false)

// --- Practice mode state ---
interface BlankWord {
  word: string
  hidden: boolean
  revealed: boolean
}
const blanks = ref<BlankWord[]>([])
const practiceRevealed = ref(false)

// --- Quiz mode state ---
const quizInput = ref('')
const quizChecked = ref(false)
const quizDiff = ref<{ word: string; typed: string; match: boolean }[]>([])
const quizScore = ref({ correct: 0, total: 0 })

// Parse SM-2 config
function parseConfig(config: string) {
  try { return JSON.parse(config) } catch { return {} }
}

const currentCard = computed(() => cards.value[currentIndex.value] || null)
const progress = computed(() => ({ current: currentIndex.value + 1, total: cards.value.length }))

// Strip HTML footnote tags from verse text
function stripFootnotes(text: string): string {
  return text.replace(/<sup>\[.*?\]\(.*?\)<\/sup>/g, '').replace(/\s+/g, ' ').trim()
}

async function load() {
  loading.value = true

  // Check for specific card ID from query param
  const cardId = route.query.id ? Number(route.query.id) : null
  // Check for mode from query param
  if (route.query.mode && ['review', 'practice', 'quiz'].includes(route.query.mode as string)) {
    mode.value = route.query.mode as Mode
  }

  if (cardId) {
    const card = await api.getPractice(cardId)
    cards.value = card ? [card] : []
  } else {
    cards.value = await api.getDueCards(today)
  }

  currentIndex.value = 0
  resetModeState()
  sessionComplete.value = cards.value.length === 0
  reviewed.value = 0
  loading.value = false
}

function resetModeState() {
  flipped.value = false
  practiceRevealed.value = false
  blanks.value = []
  quizInput.value = ''
  quizChecked.value = false
  quizDiff.value = []

  if (currentCard.value && mode.value === 'practice') {
    initBlanks()
  }
}

// Watch for mode changes to reset state
watch(mode, () => {
  resetModeState()
})

// --- Review mode ---
function flip() {
  flipped.value = !flipped.value
}

// --- Practice mode: create word blanks ---
function initBlanks() {
  if (!currentCard.value?.description) return
  const text = stripFootnotes(currentCard.value.description)
  const words = text.split(/\s+/)
  const ratio = 0.35
  const toBlank = Math.max(1, Math.floor(words.length * ratio))

  // Create word objects
  const result: BlankWord[] = words.map(w => ({ word: w, hidden: false, revealed: false }))

  // Pick random indices to blank (skip very short words)
  const candidates = result
    .map((w, i) => ({ i, len: w.word.replace(/[.,;:!?'"()]/g, '').length }))
    .filter(c => c.len > 2)

  // Shuffle candidates
  for (let i = candidates.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1))
    const tmp = candidates[i]!
    candidates[i] = candidates[j]!
    candidates[j] = tmp
  }

  for (let k = 0; k < Math.min(toBlank, candidates.length); k++) {
    const idx = candidates[k]!.i
    result[idx]!.hidden = true
  }

  blanks.value = result
}

function revealWord(index: number) {
  blanks.value[index]!.revealed = true
  if (blanks.value.every(b => !b.hidden || b.revealed)) {
    practiceRevealed.value = true
  }
}

function revealAllBlanks() {
  blanks.value.forEach(b => { if (b.hidden) b.revealed = true })
  practiceRevealed.value = true
}

const blanksRemaining = computed(() => blanks.value.filter(b => b.hidden && !b.revealed).length)

// --- Quiz mode: compare typed text ---
function checkQuiz() {
  if (!currentCard.value?.description) return
  const original = stripFootnotes(currentCard.value.description)
  const origWords = original.split(/\s+/).filter(Boolean)
  const typedWords = quizInput.value.split(/\s+/).filter(Boolean)

  let correct = 0
  const diff = origWords.map((w, i) => {
    const cleanOrig = w.replace(/[.,;:!?'"()]/g, '').toLowerCase()
    const typed = typedWords[i] || ''
    const cleanTyped = typed.replace(/[.,;:!?'"()]/g, '').toLowerCase()
    const match = cleanOrig === cleanTyped
    if (match) correct++
    return { word: w, typed: typed || '___', match }
  })

  quizDiff.value = diff
  quizScore.value = { correct, total: origWords.length }
  quizChecked.value = true
}

// --- Shared: SM-2 quality rating ---
const qualityOptions = [
  { value: 0, label: 'Blackout', color: 'bg-red-500', desc: 'No memory at all' },
  { value: 1, label: 'Wrong', color: 'bg-red-400', desc: 'Incorrect, recognized answer' },
  { value: 2, label: 'Hard', color: 'bg-orange-400', desc: 'Incorrect, seemed familiar' },
  { value: 3, label: 'Okay', color: 'bg-yellow-400', desc: 'Correct, but difficult' },
  { value: 4, label: 'Good', color: 'bg-green-400', desc: 'Correct after hesitation' },
  { value: 5, label: 'Easy', color: 'bg-green-500', desc: 'Perfect recall' },
]

const showRating = computed(() => {
  if (mode.value === 'review') return flipped.value
  if (mode.value === 'practice') return practiceRevealed.value
  if (mode.value === 'quiz') return quizChecked.value
  return false
})

async function rate(quality: number) {
  if (!currentCard.value) return
  await api.reviewCard(currentCard.value.id, quality, today)
  reviewed.value++

  if (currentIndex.value < cards.value.length - 1) {
    currentIndex.value++
    resetModeState()
    if (mode.value === 'practice') initBlanks()
  } else {
    sessionComplete.value = true
  }
}

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
    <div class="flex items-center justify-between mb-4">
      <h1 class="text-2xl font-bold">Memorize</h1>
      <div v-if="!loading && !sessionComplete" class="text-sm text-gray-500">
        {{ progress.current }} / {{ progress.total }} cards
      </div>
    </div>

    <!-- Mode selector -->
    <div v-if="!loading && !sessionComplete" class="flex gap-1 mb-6 bg-gray-100 rounded-lg p-1">
      <button
        v-for="m in (['review', 'practice', 'quiz'] as Mode[])"
        :key="m"
        @click="mode = m"
        class="flex-1 px-3 py-2 text-sm rounded-md transition-colors capitalize"
        :class="mode === m ? 'bg-white shadow text-indigo-700 font-semibold' : 'text-gray-500 hover:text-gray-700'"
      >
        {{ m === 'review' ? '👁 Review' : m === 'practice' ? '✏️ Practice' : '📝 Quiz' }}
      </button>
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

    <!-- Card area -->
    <div v-else-if="currentCard" class="max-w-xl mx-auto">

      <!-- ===== REVIEW MODE ===== -->
      <template v-if="mode === 'review'">
        <div @click="flip" class="cursor-pointer select-none mb-6">
          <div
            class="bg-white rounded-2xl shadow-lg border border-gray-100 p-8 min-h-[280px] flex flex-col items-center justify-center transition-all duration-200"
            :class="flipped ? 'ring-2 ring-indigo-200' : 'hover:shadow-xl'"
          >
            <div v-if="!flipped" class="text-center">
              <div class="text-xs uppercase tracking-wide text-gray-400 mb-3">
                {{ currentCard.category || 'scripture' }}
              </div>
              <div class="text-2xl font-bold text-gray-800 mb-2">{{ currentCard.name }}</div>
              <div class="text-sm text-gray-400 mt-4">tap to reveal</div>
            </div>
            <div v-else class="text-center">
              <div class="text-xs uppercase tracking-wide text-indigo-400 mb-3">{{ currentCard.name }}</div>
              <div class="text-lg leading-relaxed text-gray-700 whitespace-pre-line">
                {{ currentCard.description || '(no text — add via Practices)' }}
              </div>
            </div>
          </div>
        </div>
      </template>

      <!-- ===== PRACTICE MODE (fill in blanks) ===== -->
      <template v-if="mode === 'practice'">
        <div class="mb-6">
          <div class="text-center mb-4">
            <div class="text-xs uppercase tracking-wide text-gray-400 mb-1">{{ currentCard.category || 'scripture' }}</div>
            <div class="text-xl font-bold text-gray-800">{{ currentCard.name }}</div>
          </div>

          <div class="bg-white rounded-2xl shadow-lg border border-gray-100 p-6 min-h-[200px]">
            <div class="text-lg leading-relaxed text-gray-700 flex flex-wrap gap-1">
              <template v-for="(b, i) in blanks" :key="i">
                <span v-if="!b.hidden || b.revealed" class="inline-block">
                  <span :class="b.hidden && b.revealed ? 'text-indigo-600 font-semibold bg-indigo-50 px-1 rounded' : ''">{{ b.word }}</span>
                </span>
                <button
                  v-else
                  @click="revealWord(i)"
                  class="inline-block px-2 py-0.5 bg-gray-100 border border-gray-300 rounded text-gray-400 hover:bg-indigo-50 hover:border-indigo-300 hover:text-indigo-500 transition-colors min-w-[3rem] text-center"
                >
                  ____
                </button>
              </template>
            </div>
            <div v-if="!practiceRevealed" class="mt-4 flex items-center justify-between text-sm text-gray-400">
              <span>{{ blanksRemaining }} blank{{ blanksRemaining !== 1 ? 's' : '' }} remaining</span>
              <button @click="revealAllBlanks" class="text-indigo-500 hover:text-indigo-700">Reveal all</button>
            </div>
          </div>
        </div>
      </template>

      <!-- ===== QUIZ MODE (type it out) ===== -->
      <template v-if="mode === 'quiz'">
        <div class="mb-6">
          <div class="text-center mb-4">
            <div class="text-xs uppercase tracking-wide text-gray-400 mb-1">{{ currentCard.category || 'scripture' }}</div>
            <div class="text-xl font-bold text-gray-800">{{ currentCard.name }}</div>
          </div>

          <div v-if="!quizChecked" class="bg-white rounded-2xl shadow-lg border border-gray-100 p-6">
            <textarea
              v-model="quizInput"
              rows="6"
              class="w-full border border-gray-200 rounded-lg px-4 py-3 text-lg leading-relaxed focus:ring-2 focus:ring-indigo-200 focus:border-indigo-300 outline-none resize-none"
              placeholder="Type the verse from memory..."
              @keydown.ctrl.enter="checkQuiz"
            ></textarea>
            <div class="flex justify-between items-center mt-3">
              <span class="text-xs text-gray-400">Ctrl+Enter to check</span>
              <button
                @click="checkQuiz"
                :disabled="!quizInput.trim()"
                class="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 text-sm disabled:opacity-40 disabled:cursor-not-allowed"
              >
                Check
              </button>
            </div>
          </div>

          <div v-else class="bg-white rounded-2xl shadow-lg border border-gray-100 p-6">
            <div class="flex items-center justify-between mb-4">
              <span class="text-sm font-semibold text-gray-700">
                Score: {{ quizScore.correct }}/{{ quizScore.total }}
                ({{ Math.round((quizScore.correct / Math.max(quizScore.total, 1)) * 100) }}%)
              </span>
              <span
                class="text-xs px-2 py-1 rounded-full"
                :class="quizScore.correct === quizScore.total ? 'bg-green-100 text-green-700' : quizScore.correct > quizScore.total * 0.7 ? 'bg-yellow-100 text-yellow-700' : 'bg-red-100 text-red-700'"
              >
                {{ quizScore.correct === quizScore.total ? 'Perfect!' : quizScore.correct > quizScore.total * 0.7 ? 'Almost!' : 'Keep practicing' }}
              </span>
            </div>
            <div class="text-lg leading-relaxed flex flex-wrap gap-1">
              <span
                v-for="(d, i) in quizDiff"
                :key="i"
                class="inline-block px-0.5 rounded"
                :class="d.match ? 'text-green-700' : 'text-red-600 bg-red-50 line-through'"
              >{{ d.word }}</span>
            </div>
            <div v-if="quizScore.correct < quizScore.total" class="mt-3 text-sm text-gray-500 border-t pt-3">
              <strong>You typed:</strong>
              <div class="text-gray-600 mt-1 whitespace-pre-line">{{ quizInput }}</div>
            </div>
          </div>
        </div>
      </template>

      <!-- Rating buttons (shared across all modes) -->
      <div v-if="showRating" class="space-y-3 mb-6">
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

      <!-- Card stats -->
      <div class="mt-4 flex justify-center gap-6 text-xs text-gray-400">
        <span>interval: {{ cardStats(currentCard).interval }}d</span>
        <span>reps: {{ cardStats(currentCard).reps }}</span>
        <span>ease: {{ cardStats(currentCard).ease }}</span>
      </div>
    </div>
  </div>
</template>
