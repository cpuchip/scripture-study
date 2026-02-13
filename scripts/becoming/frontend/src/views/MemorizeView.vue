<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { api, type Practice, type MemorizeCardStatus } from '../api'

const route = useRoute()

function localDateStr(d: Date = new Date()): string {
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

const today = localDateStr()
const allCards = ref<MemorizeCardStatus[]>([])
const currentIndex = ref(0)
const loading = ref(true)

// Mode: review (tap-to-flip), practice (fill blanks), quiz (type it out)
type Mode = 'review' | 'practice' | 'quiz'
const mode = ref<Mode>('review')

// Practice sub-level
type PracticeLevel = 'reveal' | 'type' | 'order'
const practiceLevel = ref<PracticeLevel>('reveal')

// After-rate inline state
const justRated = ref(false)
const lastQuality = ref(0)

// --- Review mode state ---
const flipped = ref(false)

// --- Practice mode: Reveal state ---
interface BlankWord {
  word: string
  hidden: boolean
  revealed: boolean
}
const blanks = ref<BlankWord[]>([])
const practiceRevealed = ref(false)

// --- Practice mode: Type state ---
interface TypeWord {
  word: string
  hidden: boolean
  userInput: string
  checked: boolean
  correct: boolean
}
const typeWords = ref<TypeWord[]>([])
const typeChecked = ref(false)
const typeScore = ref({ correct: 0, total: 0 })

// --- Practice mode: Order state ---
interface OrderWord {
  word: string
  originalIndex: number
  placed: boolean
}
const wordBank = ref<OrderWord[]>([])
const placedWords = ref<OrderWord[]>([])
const orderChecked = ref(false)
const orderScore = ref({ correct: 0, total: 0 })

// --- Quiz mode state ---
const quizInput = ref('')
const quizChecked = ref(false)
const quizDiff = ref<{ word: string; typed: string; kind: 'match' | 'wrong' | 'missing' | 'extra' }[]>([])
const quizScore = ref({ correct: 0, total: 0 })

// Suggested quality based on practice accuracy
const suggestedQuality = ref<number | null>(null)

// Parse SM-2 config
function parseConfig(config: string) {
  try { return JSON.parse(config) } catch { return {} }
}

const currentStatus = computed(() => allCards.value[currentIndex.value] || null)
const currentCard = computed(() => currentStatus.value?.practice || null)

// Completion helpers
const allDone = computed(() => allCards.value.length > 0 && allCards.value.every(c => c.reviews_today >= c.target_daily_reps))
const currentCardDone = computed(() => {
  const s = currentStatus.value
  return s ? s.reviews_today >= s.target_daily_reps : false
})
const nextIncompleteIndex = computed(() => {
  for (let i = 0; i < allCards.value.length; i++) {
    const idx = (currentIndex.value + 1 + i) % allCards.value.length
    if (idx === currentIndex.value) continue
    const c = allCards.value[idx]
    if (c && c.reviews_today < c.target_daily_reps) return idx
  }
  return -1
})

// Strip HTML footnote tags from verse text
function stripFootnotes(text: string): string {
  return text.replace(/<sup>\[.*?\]\(.*?\)<\/sup>/g, '').replace(/\s+/g, ' ').trim()
}

async function load() {
  loading.value = true

  if (route.query.mode && ['review', 'practice', 'quiz'].includes(route.query.mode as string)) {
    mode.value = route.query.mode as Mode
  }

  allCards.value = await api.getMemorizeCards(today)

  const cardId = route.query.id ? Number(route.query.id) : null
  if (cardId) {
    const idx = allCards.value.findIndex(c => c.practice.id === cardId)
    currentIndex.value = idx >= 0 ? idx : 0
  } else {
    const idx = allCards.value.findIndex(c => c.reviews_today < c.target_daily_reps)
    currentIndex.value = idx >= 0 ? idx : 0
  }

  resetModeState()
  justRated.value = false
  loading.value = false
}

function selectCard(index: number) {
  currentIndex.value = index
  resetModeState()
  justRated.value = false
}

function resetModeState() {
  flipped.value = false
  practiceRevealed.value = false
  blanks.value = []
  typeWords.value = []
  typeChecked.value = false
  wordBank.value = []
  placedWords.value = []
  orderChecked.value = false
  quizInput.value = ''
  quizChecked.value = false
  quizDiff.value = []
  suggestedQuality.value = null
  justRated.value = false

  if (currentCard.value && mode.value === 'practice') {
    if (practiceLevel.value === 'reveal') initBlanks()
    else if (practiceLevel.value === 'type') initTypeWords()
    else if (practiceLevel.value === 'order') initOrderWords()
  }
}

watch(mode, () => resetModeState())
watch(practiceLevel, () => resetModeState())

// --- Review mode ---
function flip() {
  flipped.value = !flipped.value
}

// --- Utility ---
function shuffle<T>(arr: T[]) {
  for (let i = arr.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1))
    const tmp = arr[i]!
    arr[i] = arr[j]!
    arr[j] = tmp
  }
}

// --- Practice: Reveal ---
function initBlanks() {
  if (!currentCard.value?.description) return
  const text = stripFootnotes(currentCard.value.description)
  const words = text.split(/\s+/)
  const ratio = 0.35
  const toBlank = Math.max(1, Math.floor(words.length * ratio))
  const result: BlankWord[] = words.map(w => ({ word: w, hidden: false, revealed: false }))
  const candidates = result
    .map((w, i) => ({ i, len: w.word.replace(/[.,;:!?'"()]/g, '').length }))
    .filter(c => c.len > 2)
  shuffle(candidates)
  for (let k = 0; k < Math.min(toBlank, candidates.length); k++) {
    result[candidates[k]!.i]!.hidden = true
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

// --- Practice: Type ---
function initTypeWords() {
  if (!currentCard.value?.description) return
  const text = stripFootnotes(currentCard.value.description)
  const words = text.split(/\s+/)
  const ratio = 0.35
  const toBlank = Math.max(1, Math.floor(words.length * ratio))
  const result: TypeWord[] = words.map(w => ({ word: w, hidden: false, userInput: '', checked: false, correct: false }))
  const candidates = result
    .map((w, i) => ({ i, len: w.word.replace(/[.,;:!?'"()]/g, '').length }))
    .filter(c => c.len > 2)
  shuffle(candidates)
  for (let k = 0; k < Math.min(toBlank, candidates.length); k++) {
    result[candidates[k]!.i]!.hidden = true
  }
  typeWords.value = result
}

function checkTypeWords() {
  let correct = 0
  let total = 0
  typeWords.value.forEach(tw => {
    if (tw.hidden) {
      total++
      const cleanOrig = tw.word.replace(/[.,;:!?'"()]/g, '').toLowerCase()
      const cleanTyped = tw.userInput.trim().replace(/[.,;:!?'"()]/g, '').toLowerCase()
      tw.correct = cleanOrig === cleanTyped
      tw.checked = true
      if (tw.correct) correct++
    }
  })
  typeChecked.value = true
  typeScore.value = { correct, total }
  suggestQuality(total > 0 ? correct / total : 0)
}

// --- Practice: Order ---
function initOrderWords() {
  if (!currentCard.value?.description) return
  const text = stripFootnotes(currentCard.value.description)
  const words = text.split(/\s+/)
  const bank: OrderWord[] = words.map((w, i) => ({ word: w, originalIndex: i, placed: false }))
  const shuffled = [...bank]
  shuffle(shuffled)
  wordBank.value = shuffled
  placedWords.value = []
  orderChecked.value = false
}

function placeWord(bankIndex: number) {
  if (orderChecked.value) return
  const w = wordBank.value[bankIndex]!
  if (w.placed) return
  w.placed = true
  placedWords.value.push({ word: w.word, originalIndex: w.originalIndex, placed: true })
}

function unplaceWord(placedIndex: number) {
  if (orderChecked.value) return
  const w = placedWords.value[placedIndex]!
  const bankItem = wordBank.value.find(b => b.originalIndex === w.originalIndex && b.placed)
  if (bankItem) bankItem.placed = false
  placedWords.value.splice(placedIndex, 1)
}

function getOriginalWords(): string[] {
  if (!currentCard.value?.description) return []
  return stripFootnotes(currentCard.value.description).split(/\s+/)
}

function isOrderCorrect(placedIndex: number): boolean {
  const original = getOriginalWords()
  return placedWords.value[placedIndex]?.word === original[placedIndex]
}

function checkOrder() {
  const original = getOriginalWords()
  let correct = 0
  const total = placedWords.value.length
  placedWords.value.forEach((w, i) => {
    // Compare word TEXT, not originalIndex — allows duplicate words to be swapped freely
    if (w.word === original[i]) correct++
  })
  orderChecked.value = true
  orderScore.value = { correct, total }
  suggestQuality(total > 0 ? correct / total : 0)
}

const orderComplete = computed(() => wordBank.value.length > 0 && wordBank.value.every(w => w.placed))

// --- Quiz mode ---

// Clean a word for comparison: strip punctuation, lowercase
function cleanWord(w: string): string {
  return w.replace(/[.,;:!?'"()—\-–]/g, '').toLowerCase()
}

// Compute LCS (Longest Common Subsequence) table for word alignment
function lcsTable(a: string[], b: string[]): number[][] {
  const m = a.length, n = b.length
  const dp: number[][] = Array.from({ length: m + 1 }, () => new Array<number>(n + 1).fill(0))
  for (let i = 1; i <= m; i++) {
    for (let j = 1; j <= n; j++) {
      if (cleanWord(a[i - 1]!) === cleanWord(b[j - 1]!)) {
        dp[i]![j] = dp[i - 1]![j - 1]! + 1
      } else {
        dp[i]![j] = Math.max(dp[i - 1]![j]!, dp[i]![j - 1]!)
      }
    }
  }
  return dp
}

// Backtrack LCS table to produce a word-level diff
function wordDiff(origWords: string[], typedWords: string[]): { word: string; typed: string; kind: 'match' | 'wrong' | 'missing' | 'extra' }[] {
  const dp = lcsTable(origWords, typedWords)
  const result: { word: string; typed: string; kind: 'match' | 'wrong' | 'missing' | 'extra' }[] = []

  let i = origWords.length, j = typedWords.length
  const ops: { type: 'match' | 'delete' | 'insert'; origIdx: number; typedIdx: number }[] = []

  while (i > 0 && j > 0) {
    if (cleanWord(origWords[i - 1]!) === cleanWord(typedWords[j - 1]!)) {
      ops.push({ type: 'match', origIdx: i - 1, typedIdx: j - 1 })
      i--; j--
    } else if (dp[i - 1]![j]! >= dp[i]![j - 1]!) {
      ops.push({ type: 'delete', origIdx: i - 1, typedIdx: -1 })
      i--
    } else {
      ops.push({ type: 'insert', origIdx: -1, typedIdx: j - 1 })
      j--
    }
  }
  while (i > 0) {
    ops.push({ type: 'delete', origIdx: i - 1, typedIdx: -1 })
    i--
  }
  while (j > 0) {
    ops.push({ type: 'insert', origIdx: -1, typedIdx: j - 1 })
    j--
  }

  ops.reverse()

  // Merge consecutive delete+insert pairs into 'wrong' (substitution)
  let k = 0
  while (k < ops.length) {
    const op = ops[k]!
    if (op.type === 'match') {
      result.push({ word: origWords[op.origIdx]!, typed: typedWords[op.typedIdx]!, kind: 'match' })
    } else if (op.type === 'delete' && k + 1 < ops.length && ops[k + 1]!.type === 'insert') {
      // Substitution: wrong word typed instead of the expected one
      const next = ops[k + 1]!
      result.push({ word: origWords[op.origIdx]!, typed: typedWords[next.typedIdx]!, kind: 'wrong' })
      k++ // skip the insert
    } else if (op.type === 'delete') {
      result.push({ word: origWords[op.origIdx]!, typed: '', kind: 'missing' })
    } else if (op.type === 'insert') {
      result.push({ word: '', typed: typedWords[op.typedIdx]!, kind: 'extra' })
    }
    k++
  }

  return result
}

function checkQuiz() {
  if (!currentCard.value?.description) return
  const original = stripFootnotes(currentCard.value.description)
  const origWords = original.split(/\s+/).filter(Boolean)
  const typedWords = quizInput.value.split(/\s+/).filter(Boolean)

  const diff = wordDiff(origWords, typedWords)
  const correct = diff.filter(d => d.kind === 'match').length

  quizDiff.value = diff
  quizScore.value = { correct, total: origWords.length }
  quizChecked.value = true
  suggestQuality(origWords.length > 0 ? correct / origWords.length : 0)
}

// --- Quality suggestion ---
function suggestQuality(accuracy: number) {
  if (accuracy >= 1.0) suggestedQuality.value = 5
  else if (accuracy >= 0.9) suggestedQuality.value = 4
  else if (accuracy >= 0.7) suggestedQuality.value = 3
  else if (accuracy >= 0.5) suggestedQuality.value = 2
  else if (accuracy > 0) suggestedQuality.value = 1
  else suggestedQuality.value = 0
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
  if (justRated.value) return false
  if (mode.value === 'review') return flipped.value
  if (mode.value === 'practice') {
    if (practiceLevel.value === 'reveal') return practiceRevealed.value
    if (practiceLevel.value === 'type') return typeChecked.value
    if (practiceLevel.value === 'order') return orderChecked.value
  }
  if (mode.value === 'quiz') return quizChecked.value
  return false
})

async function rate(quality: number) {
  if (!currentCard.value || !currentStatus.value) return
  const updated = await api.reviewCard(currentCard.value.id, quality, today)

  // Update local state
  currentStatus.value.reviews_today++
  currentStatus.value.today_qualities.push(quality)
  currentStatus.value.practice = updated

  lastQuality.value = quality
  justRated.value = true
}

function continueCard() {
  justRated.value = false
  resetModeState()
}

function goToNextCard() {
  if (nextIncompleteIndex.value >= 0) {
    selectCard(nextIncompleteIndex.value)
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

// Quality color helper
function qualityColor(q: number): string {
  if (q <= 1) return 'bg-red-400'
  if (q === 2) return 'bg-orange-400'
  if (q === 3) return 'bg-yellow-400'
  if (q === 4) return 'bg-green-400'
  return 'bg-green-500'
}

onMounted(load)
</script>

<template>
  <div>
    <!-- Header -->
    <div class="flex items-center justify-between mb-4">
      <h1 class="text-2xl font-bold">Memorize</h1>
      <div v-if="!loading && allCards.length > 0" class="flex items-center gap-3">
        <span class="text-sm text-gray-500">
          {{ allCards.filter(c => c.reviews_today >= c.target_daily_reps).length }}/{{ allCards.length }} cards done
        </span>
        <router-link
          to="/practices?type=memorize&create=1"
          class="px-3 py-1.5 bg-indigo-600 text-white rounded-lg text-sm hover:bg-indigo-700"
        >+ Add Card</router-link>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="text-center py-12 text-gray-400">Loading cards...</div>

    <!-- No cards -->
    <div v-else-if="allCards.length === 0" class="text-center py-12">
      <p class="text-gray-500 mb-4">No memorize cards yet.</p>
      <router-link to="/practices?type=memorize&create=1" class="text-indigo-600 hover:underline">Add your first card →</router-link>
    </div>

    <!-- Main content -->
    <template v-else>
      <!-- Card picker pills -->
      <div class="flex gap-2 mb-4 overflow-x-auto pb-1">
        <button
          v-for="(card, i) in allCards"
          :key="card.practice.id"
          @click="selectCard(i)"
          class="flex items-center gap-1.5 px-3 py-1.5 rounded-full text-sm whitespace-nowrap transition-colors border"
          :class="i === currentIndex
            ? 'bg-indigo-600 text-white border-indigo-600'
            : card.reviews_today >= card.target_daily_reps
              ? 'bg-green-50 border-green-300 text-green-700 hover:bg-green-100'
              : 'bg-white border-gray-200 text-gray-600 hover:border-indigo-300'"
        >
          {{ card.practice.name }}
          <span
            class="text-[10px] px-1.5 py-0.5 rounded-full"
            :class="i === currentIndex
              ? 'bg-indigo-500 text-indigo-100'
              : card.reviews_today >= card.target_daily_reps
                ? 'bg-green-200 text-green-800'
                : 'bg-gray-100 text-gray-500'"
          >{{ card.reviews_today }}/{{ card.target_daily_reps }}</span>
        </button>
      </div>

      <!-- Reps row for current card -->
      <div v-if="currentStatus" class="flex items-center gap-1.5 mb-4">
        <template v-for="i in currentStatus.target_daily_reps" :key="i">
          <div
            class="w-5 h-5 rounded-full flex items-center justify-center text-[9px] font-bold"
            :class="i <= currentStatus.reviews_today
              ? qualityColor(currentStatus.today_qualities[i - 1] ?? 0) + ' text-white'
              : 'bg-gray-200 text-gray-400'"
          >
            {{ i <= currentStatus.reviews_today ? currentStatus.today_qualities[i - 1] : '' }}
          </div>
        </template>
        <span class="text-xs text-gray-500 ml-1">
          {{ currentStatus.reviews_today }}/{{ currentStatus.target_daily_reps }}
        </span>
        <span v-if="currentStatus.is_due" class="text-[10px] px-1.5 py-0.5 bg-indigo-100 text-indigo-600 rounded-full ml-1">due</span>
      </div>

      <!-- Mode selector -->
      <div class="flex gap-1 mb-4 bg-gray-100 rounded-lg p-1">
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

      <!-- Practice level selector -->
      <div v-if="mode === 'practice'" class="flex gap-1 mb-4 bg-gray-50 rounded-lg p-1">
        <button
          v-for="l in (['reveal', 'type', 'order'] as PracticeLevel[])"
          :key="l"
          @click="practiceLevel = l"
          class="flex-1 px-2 py-1.5 text-xs rounded-md transition-colors"
          :class="practiceLevel === l ? 'bg-white shadow text-indigo-700 font-semibold' : 'text-gray-400 hover:text-gray-600'"
        >
          {{ l === 'reveal' ? '👆 Reveal' : l === 'type' ? '⌨️ Type' : '🧩 Order' }}
        </button>
      </div>

      <!-- Just rated: inline result -->
      <div v-if="justRated && currentCard" class="mb-6">
        <div class="bg-white rounded-2xl shadow-lg border border-gray-100 p-6 text-center">
          <div class="text-3xl mb-2">{{ lastQuality >= 3 ? '✅' : '🔄' }}</div>
          <p class="font-semibold text-gray-800 mb-1">
            Rep {{ currentStatus!.reviews_today }}/{{ currentStatus!.target_daily_reps }} complete
          </p>
          <p class="text-sm text-gray-500 mb-4">{{ currentCard.name }}</p>

          <div class="flex gap-2 justify-center flex-wrap">
            <button
              v-if="!currentCardDone"
              @click="continueCard"
              class="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 text-sm"
            >
              Practice again
            </button>
            <button
              v-if="nextIncompleteIndex >= 0"
              @click="goToNextCard"
              class="px-4 py-2 text-sm rounded-lg"
              :class="currentCardDone
                ? 'bg-indigo-600 text-white hover:bg-indigo-700'
                : 'bg-gray-100 text-gray-700 hover:bg-gray-200'"
            >
              Next: {{ allCards[nextIncompleteIndex]?.practice.name }} →
            </button>
            <div v-if="allDone" class="w-full mt-2">
              <p class="text-green-600 font-semibold">🎉 All cards done for today!</p>
            </div>
          </div>
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

        <!-- ===== PRACTICE: REVEAL ===== -->
        <template v-if="mode === 'practice' && practiceLevel === 'reveal'">
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

        <!-- ===== PRACTICE: TYPE ===== -->
        <template v-if="mode === 'practice' && practiceLevel === 'type'">
          <div class="mb-6">
            <div class="text-center mb-4">
              <div class="text-xs uppercase tracking-wide text-gray-400 mb-1">{{ currentCard.category || 'scripture' }}</div>
              <div class="text-xl font-bold text-gray-800">{{ currentCard.name }}</div>
            </div>

            <div class="bg-white rounded-2xl shadow-lg border border-gray-100 p-6 min-h-[200px]">
              <div class="text-lg leading-relaxed text-gray-700 flex flex-wrap gap-1 items-baseline">
                <template v-for="(tw, i) in typeWords" :key="i">
                  <span v-if="!tw.hidden" class="inline-block">{{ tw.word }}</span>
                  <template v-else>
                    <span v-if="tw.checked" class="inline-block px-1 rounded font-semibold"
                      :class="tw.correct ? 'text-green-600 bg-green-50' : 'text-red-600 bg-red-50'"
                    >
                      {{ tw.correct ? tw.word : tw.userInput || '___' }}
                      <span v-if="!tw.correct" class="text-green-600 text-sm ml-0.5">({{ tw.word }})</span>
                    </span>
                    <input
                      v-else
                      v-model="tw.userInput"
                      type="text"
                      :placeholder="'_'.repeat(Math.min(tw.word.length, 8))"
                      class="inline-block border-b-2 border-gray-300 focus:border-indigo-500 outline-none px-1 py-0.5 text-center bg-transparent min-w-[3rem] max-w-[8rem] text-lg"
                      :style="{ width: Math.max(3, tw.word.length * 0.7) + 'rem' }"
                      @keydown.enter="checkTypeWords"
                    />
                  </template>
                </template>
              </div>
              <div v-if="!typeChecked" class="mt-4 flex items-center justify-between text-sm">
                <span class="text-gray-400">Type the missing words, then check</span>
                <button
                  @click="checkTypeWords"
                  class="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 text-sm"
                >Check</button>
              </div>
              <div v-else class="mt-4 flex items-center justify-between text-sm">
                <span class="text-gray-600">
                  Score: {{ typeScore.correct }}/{{ typeScore.total }}
                  ({{ Math.round((typeScore.correct / Math.max(typeScore.total, 1)) * 100) }}%)
                </span>
                <span
                  class="text-xs px-2 py-1 rounded-full"
                  :class="typeScore.correct === typeScore.total ? 'bg-green-100 text-green-700' : typeScore.correct > typeScore.total * 0.7 ? 'bg-yellow-100 text-yellow-700' : 'bg-red-100 text-red-700'"
                >{{ typeScore.correct === typeScore.total ? 'Perfect!' : typeScore.correct > typeScore.total * 0.7 ? 'Almost!' : 'Keep practicing' }}</span>
              </div>
            </div>
          </div>
        </template>

        <!-- ===== PRACTICE: ORDER ===== -->
        <template v-if="mode === 'practice' && practiceLevel === 'order'">
          <div class="mb-6">
            <div class="text-center mb-4">
              <div class="text-xs uppercase tracking-wide text-gray-400 mb-1">{{ currentCard.category || 'scripture' }}</div>
              <div class="text-xl font-bold text-gray-800">{{ currentCard.name }}</div>
            </div>

            <div class="bg-white rounded-2xl shadow-lg border border-gray-100 p-6">
              <!-- Placed words (sentence being built) -->
              <div class="min-h-[80px] border-2 border-dashed border-gray-200 rounded-lg p-3 mb-4 flex flex-wrap gap-1.5 items-start">
                <button
                  v-for="(w, i) in placedWords"
                  :key="'placed-' + i"
                  @click="unplaceWord(i)"
                  class="px-2.5 py-1 rounded text-sm transition-colors"
                  :class="orderChecked
                    ? isOrderCorrect(i)
                      ? 'bg-green-100 text-green-700 border border-green-300'
                      : 'bg-red-100 text-red-700 border border-red-300'
                    : 'bg-indigo-100 text-indigo-700 border border-indigo-300 hover:bg-red-50 hover:text-red-600 hover:border-red-300 cursor-pointer'"
                >{{ w.word }}</button>
                <span v-if="placedWords.length === 0" class="text-sm text-gray-400 py-1">
                  Click words below to build the verse...
                </span>
              </div>

              <!-- Word bank -->
              <div class="flex flex-wrap gap-1.5">
                <button
                  v-for="(w, i) in wordBank"
                  :key="'bank-' + i"
                  @click="placeWord(i)"
                  :disabled="w.placed || orderChecked"
                  class="px-2.5 py-1 rounded text-sm border transition-colors"
                  :class="w.placed
                    ? 'bg-gray-50 text-gray-300 border-gray-100 cursor-default'
                    : 'bg-white text-gray-700 border-gray-300 hover:bg-indigo-50 hover:border-indigo-300 cursor-pointer'"
                >{{ w.word }}</button>
              </div>

              <!-- Check button + score -->
              <div v-if="!orderChecked" class="mt-4 flex items-center justify-between text-sm">
                <span class="text-gray-400">{{ placedWords.length }}/{{ wordBank.length }} words placed</span>
                <button
                  @click="checkOrder"
                  :disabled="!orderComplete"
                  class="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 text-sm disabled:opacity-40 disabled:cursor-not-allowed"
                >Check</button>
              </div>
              <div v-else class="mt-4 flex items-center justify-between text-sm">
                <span class="text-gray-600">
                  Score: {{ orderScore.correct }}/{{ orderScore.total }}
                  ({{ Math.round((orderScore.correct / Math.max(orderScore.total, 1)) * 100) }}%)
                </span>
                <span
                  class="text-xs px-2 py-1 rounded-full"
                  :class="orderScore.correct === orderScore.total ? 'bg-green-100 text-green-700' : orderScore.correct > orderScore.total * 0.7 ? 'bg-yellow-100 text-yellow-700' : 'bg-red-100 text-red-700'"
                >{{ orderScore.correct === orderScore.total ? 'Perfect!' : orderScore.correct > orderScore.total * 0.7 ? 'Almost!' : 'Keep practicing' }}</span>
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
                  :class="{
                    'text-green-700': d.kind === 'match',
                    'text-red-600 bg-red-50': d.kind === 'wrong',
                    'text-red-600 bg-red-50 border-b-2 border-dashed border-red-300': d.kind === 'missing',
                    'text-orange-500 bg-orange-50 italic': d.kind === 'extra',
                  }"
                  :title="d.kind === 'wrong' ? `You typed: ${d.typed}` : d.kind === 'missing' ? 'Missing word' : d.kind === 'extra' ? `Extra word: ${d.typed}` : ''"
                ><template v-if="d.kind === 'extra'">+{{ d.typed }}</template><template v-else>{{ d.word }}</template></span>
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
              class="flex flex-col items-center p-3 rounded-lg border transition-colors text-sm"
              :class="suggestedQuality === opt.value
                ? 'border-indigo-400 bg-indigo-50 ring-2 ring-indigo-200'
                : 'border-gray-200 hover:border-gray-300'"
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
    </template>
  </div>
</template>
