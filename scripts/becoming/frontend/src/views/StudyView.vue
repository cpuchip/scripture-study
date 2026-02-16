<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { api, type StudyExercise, type SessionMomentum, type StudyScoreResponse } from '../api'

const router = useRouter()

function localDateStr(d: Date = new Date()): string {
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

const today = localDateStr()
const loading = ref(true)
const exercise = ref<StudyExercise | null>(null)
const sessionDone = ref(false)
const sessionMessage = ref('')

// Session state (tracked client-side, sent to server for next-card selection)
const recentScores = ref<number[]>([])
const momentum = ref<SessionMomentum>('steady')
const exercisesDone = ref(0)
const totalScore = ref(0)
const lastCardId = ref(0)
const keepStudying = ref(false)

// Exercise UI state
const exerciseStartTime = ref(0)
const showResult = ref(false)
const currentScore = ref(0)
const lastScoreResponse = ref<StudyScoreResponse | null>(null)

// Review (reveal_whole) state
const revealed = ref(false)

// Reveal Words state
interface BlankWord { word: string; hidden: boolean; revealed: boolean }
const blanks = ref<BlankWord[]>([])
const allBlanksRevealed = ref(false)

// Type Words state
interface TypeWord { word: string; hidden: boolean; userInput: string; checked: boolean; correct: boolean }
const typeWords = ref<TypeWord[]>([])
const typeChecked = ref(false)
const typeScore = ref({ correct: 0, total: 0 })

// Arrange state
interface OrderWord { word: string; originalIndex: number; placed: boolean }
const wordBank = ref<OrderWord[]>([])
const placedWords = ref<OrderWord[]>([])
const orderChecked = ref(false)
const orderScore = ref({ correct: 0, total: 0 })

// Type Full state
const quizInput = ref('')
const quizChecked = ref(false)
const quizDiff = ref<{ word: string; typed: string; kind: 'match' | 'wrong' | 'missing' | 'extra' }[]>([])
const quizScore = ref({ correct: 0, total: 0 })

// Reverse mode state
const referenceOptions = ref<string[]>([])
const selectedReference = ref('')
const referenceChecked = ref(false)
const referenceCorrect = ref(false)

// Helpers
function stripFootnotes(text: string): string {
  return text.replace(/<sup>\[.*?\]\(.*?\)<\/sup>/g, '').replace(/\s+/g, ' ').trim()
}

function shuffle<T>(arr: T[]) {
  for (let i = arr.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1))
    const tmp = arr[i]!; arr[i] = arr[j]!; arr[j] = tmp
  }
}

const exerciseText = computed(() => {
  if (!exercise.value?.practice.description) return ''
  return stripFootnotes(exercise.value.practice.description)
})

const exerciseName = computed(() => exercise.value?.practice.name ?? '')

// Mode labels
const modeName = computed(() => {
  if (!exercise.value) return ''
  const names: Record<string, string> = {
    reveal_whole: 'Read & Absorb',
    reveal_words: 'Reveal Words',
    type_words: 'Type Missing Words',
    arrange: 'Arrange Words',
    type_full: 'Type Full Text',
    reverse_full: 'Identify Reference',
    reverse_partial: 'Identify Reference (Partial)',
    reverse_fragment: 'Identify Reference (Fragment)',
  }
  return names[exercise.value.mode] ?? exercise.value.mode
})

const levelLabel = computed(() => {
  if (!exercise.value) return ''
  if (exercise.value.is_reverse) return `R${exercise.value.level}`
  return `Level ${exercise.value.level}`
})

const momentumEmoji = computed(() => {
  switch (momentum.value) {
    case 'struggling': return '💪'
    case 'cruising': return '🚀'
    default: return '⚡'
  }
})

const cardTypeLabel = computed(() => {
  if (!exercise.value) return ''
  const labels: Record<string, string> = {
    confidence: 'Confidence boost',
    stretch: 'Stretch challenge',
    goldilocks: 'Just right',
    fresh: 'New card',
  }
  return labels[exercise.value.card_type] ?? ''
})

// --- Session lifecycle ---

async function startSession() {
  loading.value = true
  recentScores.value = []
  momentum.value = 'steady'
  exercisesDone.value = 0
  totalScore.value = 0
  lastCardId.value = 0
  sessionDone.value = false
  keepStudying.value = false

  // Seed aptitudes from SM-2 history (idempotent)
  try { await api.studySeed() } catch { /* ok if fails */ }

  await fetchNextExercise()
  loading.value = false
}

async function fetchNextExercise() {
  const result = await api.studyNext({
    date: today,
    lastCardId: lastCardId.value,
    momentum: momentum.value,
    recentScores: recentScores.value,
    mode: 'all', // study mode always uses all active cards — daily review is in MemorizeView
  })

  if (result.done) {
    sessionDone.value = true
    sessionMessage.value = result.message || 'All done for today!'
    exercise.value = null
    return
  }

  exercise.value = result
  resetExerciseState()
  exerciseStartTime.value = Date.now()
}

function resetExerciseState() {
  showResult.value = false
  revealed.value = false
  blanks.value = []
  allBlanksRevealed.value = false
  typeWords.value = []
  typeChecked.value = false
  typeScore.value = { correct: 0, total: 0 }
  wordBank.value = []
  placedWords.value = []
  orderChecked.value = false
  orderScore.value = { correct: 0, total: 0 }
  quizInput.value = ''
  quizChecked.value = false
  quizDiff.value = []
  quizScore.value = { correct: 0, total: 0 }
  referenceOptions.value = []
  selectedReference.value = ''
  referenceChecked.value = false
  referenceCorrect.value = false
  currentScore.value = 0
  lastScoreResponse.value = null

  // Initialize based on mode
  if (!exercise.value) return
  const mode = exercise.value.mode

  if (mode === 'reveal_words') initBlanks()
  else if (mode === 'type_words') initTypeWords()
  else if (mode === 'arrange') initOrderWords()
  else if (mode.startsWith('reverse_')) initReverseOptions()
}

// --- Mode initializers ---

function initBlanks() {
  const text = exerciseText.value
  if (!text) return
  const words = text.split(/\s+/)
  const ratio = 0.35
  const toBlank = Math.max(1, Math.floor(words.length * ratio))
  const result: BlankWord[] = words.map(w => ({ word: w, hidden: false, revealed: false }))
  const candidates = result.map((w, i) => ({ i, len: w.word.replace(/[.,;:!?'"()]/g, '').length })).filter(c => c.len > 2)
  shuffle(candidates)
  for (let k = 0; k < Math.min(toBlank, candidates.length); k++) {
    result[candidates[k]!.i]!.hidden = true
  }
  blanks.value = result
}

function revealWord(index: number) {
  blanks.value[index]!.revealed = true
  if (blanks.value.every(b => !b.hidden || b.revealed)) {
    allBlanksRevealed.value = true
  }
}

function revealAllBlanks() {
  blanks.value.forEach(b => { if (b.hidden) b.revealed = true })
  allBlanksRevealed.value = true
}

function initTypeWords() {
  const text = exerciseText.value
  if (!text) return
  const words = text.split(/\s+/)
  const ratio = 0.35
  const toBlank = Math.max(1, Math.floor(words.length * ratio))
  const result: TypeWord[] = words.map(w => ({ word: w, hidden: false, userInput: '', checked: false, correct: false }))
  const candidates = result.map((w, i) => ({ i, len: w.word.replace(/[.,;:!?'"()]/g, '').length })).filter(c => c.len > 2)
  shuffle(candidates)
  for (let k = 0; k < Math.min(toBlank, candidates.length); k++) {
    result[candidates[k]!.i]!.hidden = true
  }
  typeWords.value = result
}

function checkTypeWords() {
  let correct = 0, total = 0
  typeWords.value.forEach(tw => {
    if (!tw.hidden) return
    total++
    const expected = tw.word.replace(/[.,;:!?'"()]/g, '').toLowerCase()
    const actual = tw.userInput.trim().toLowerCase()
    tw.correct = actual === expected
    tw.checked = true
    if (tw.correct) correct++
  })
  typeChecked.value = true
  typeScore.value = { correct, total }
}

function initOrderWords() {
  const text = exerciseText.value
  if (!text) return
  const words = text.split(/\s+/)
  const ordered: OrderWord[] = words.map((w, i) => ({ word: w, originalIndex: i, placed: false }))
  const shuffled = [...ordered]
  shuffle(shuffled)
  shuffled.forEach(w => w.placed = false)
  wordBank.value = shuffled
  placedWords.value = []
}

function placeWord(index: number) {
  const word = wordBank.value[index]!
  word.placed = true
  placedWords.value.push(word)
}

function unplaceWord(index: number) {
  const word = placedWords.value.splice(index, 1)[0]!
  const bankIdx = wordBank.value.findIndex(w => w === word)
  if (bankIdx >= 0) wordBank.value[bankIdx]!.placed = false
}

function checkOrder() {
  let correct = 0
  placedWords.value.forEach((w, i) => {
    if (w.originalIndex === i) correct++
  })
  orderChecked.value = true
  orderScore.value = { correct, total: placedWords.value.length }
}

// --- Quiz (Type Full) ---
function checkQuiz() {
  const expected = exerciseText.value.split(/\s+/)
  const typed = quizInput.value.trim().split(/\s+/).filter(Boolean)
  const diff = wordDiff(expected, typed)
  quizDiff.value = diff
  const matches = diff.filter(d => d.kind === 'match').length
  quizChecked.value = true
  quizScore.value = { correct: matches, total: expected.length }
}

function wordDiff(expected: string[], typed: string[]) {
  const normalize = (w: string) => w.replace(/[.,;:!?'"()]/g, '').toLowerCase()
  const lcs = lcsTable(expected.map(normalize), typed.map(normalize))

  const result: typeof quizDiff.value = []
  let i = expected.length, j = typed.length
  const items: typeof result = []

  while (i > 0 && j > 0) {
    if (normalize(expected[i - 1]!) === normalize(typed[j - 1]!)) {
      items.push({ word: expected[i - 1]!, typed: typed[j - 1]!, kind: 'match' })
      i--; j--
    } else if (lcs[i - 1]![j]! >= lcs[i]![j - 1]!) {
      items.push({ word: expected[i - 1]!, typed: '', kind: 'missing' })
      i--
    } else {
      items.push({ word: '', typed: typed[j - 1]!, kind: 'extra' })
      j--
    }
  }
  while (i > 0) { items.push({ word: expected[--i]!, typed: '', kind: 'missing' }) }
  while (j > 0) { items.push({ word: '', typed: typed[--j]!, kind: 'extra' }) }

  return items.reverse()
}

function lcsTable(a: string[], b: string[]) {
  const m = a.length, n = b.length
  const dp: number[][] = Array.from({ length: m + 1 }, () => Array(n + 1).fill(0))
  for (let i = 1; i <= m; i++) {
    for (let j = 1; j <= n; j++) {
      dp[i]![j] = a[i - 1] === b[j - 1] ? dp[i - 1]![j - 1]! + 1 : Math.max(dp[i - 1]![j]!, dp[i]![j - 1]!)
    }
  }
  return dp
}

// --- Reverse mode ---
function initReverseOptions() {
  if (!exercise.value) return
  // We'll need more cards for options — for now use the exercise name + some fake options
  // In a real implementation, we'd pass available card names from the backend
  // For now: correct answer + 3 distractors
  const correct = exercise.value.practice.name
  referenceOptions.value = [correct]
  // TODO: get real distractors from other cards. For now, generate plausible fakes.
  const fakes = generateDistractors(correct)
  referenceOptions.value.push(...fakes)
  shuffle(referenceOptions.value)
}

function generateDistractors(correct: string): string[] {
  // Simple distractor generation — swap numbers, change book names
  const distractors: string[] = []
  // If it looks like a scripture reference, generate variations
  const match = correct.match(/^(.+?)\s+(\d+):(\d+)/)
  if (match) {
    const [, book, chapter, verse] = match
    distractors.push(`${book} ${Number(chapter) + 1}:${verse}`)
    distractors.push(`${book} ${chapter}:${Number(verse) + 2}`)
    distractors.push(`${book} ${Number(chapter) - 1 || 1}:${Number(verse) + 1}`)
  } else {
    distractors.push(correct + ' (alt)')
    distractors.push('Similar passage')
    distractors.push('Another reference')
  }
  return distractors.slice(0, 3)
}

function selectReference(ref: string) {
  if (referenceChecked.value) return
  selectedReference.value = ref
}

function checkReference() {
  if (!exercise.value || !selectedReference.value) return
  referenceChecked.value = true
  referenceCorrect.value = selectedReference.value === exercise.value.practice.name
}

// --- Partial text for reverse modes ---
const reverseDisplayText = computed(() => {
  if (!exercise.value) return ''
  const text = exerciseText.value
  const mode = exercise.value.mode

  if (mode === 'reverse_full') return text

  if (mode === 'reverse_partial') {
    const words = text.split(/\s+/)
    const ratio = 0.35
    const toBlank = Math.max(1, Math.floor(words.length * ratio))
    const indices = words.map((_, i) => i).filter(i => words[i]!.replace(/[.,;:!?'"()]/g, '').length > 2)
    shuffle(indices)
    const blanked = new Set(indices.slice(0, toBlank))
    return words.map((w, i) => blanked.has(i) ? '___' : w).join(' ')
  }

  if (mode === 'reverse_fragment') {
    const words = text.split(/\s+/).filter(w => w.replace(/[.,;:!?'"()]/g, '').length > 3)
    shuffle(words)
    return words.slice(0, Math.min(5, Math.max(3, Math.floor(words.length * 0.15)))).join(' ... ')
  }

  return text
})

// --- Score computation ---
function computeScore(): number {
  if (!exercise.value) return 0
  const mode = exercise.value.mode

  switch (mode) {
    case 'reveal_whole':
      return revealed.value ? 1.0 : 0
    case 'reveal_words':
      if (blanks.value.length === 0) return 0
      return allBlanksRevealed.value ? 0.85 : 0 // Reveal mode is low-effort, cap at 0.85
    case 'type_words':
      return typeScore.value.total > 0 ? typeScore.value.correct / typeScore.value.total : 0
    case 'arrange':
      return orderScore.value.total > 0 ? orderScore.value.correct / orderScore.value.total : 0
    case 'type_full':
      return quizScore.value.total > 0 ? quizScore.value.correct / quizScore.value.total : 0
    case 'reverse_full':
    case 'reverse_partial':
    case 'reverse_fragment':
      return referenceCorrect.value ? 1.0 : 0
    default:
      return 0
  }
}

function suggestQuality(accuracy: number): number {
  if (accuracy >= 1.0) return 5
  if (accuracy >= 0.9) return 4
  if (accuracy >= 0.7) return 3
  if (accuracy >= 0.5) return 2
  if (accuracy > 0) return 1
  return 0
}

// --- Submit score and advance ---
async function submitScore() {
  if (!exercise.value) return
  const score = computeScore()
  currentScore.value = score
  const durationS = Math.round((Date.now() - exerciseStartTime.value) / 1000)

  // Only send SM-2 quality for exercises that demonstrate active recall (level 3+)
  // Level 1-2 are exposure exercises — they shouldn't advance SM-2 scheduling
  const level = exercise.value.level
  const quality = level >= 3 ? suggestQuality(score) : undefined

  try {
    lastScoreResponse.value = await api.studyScore({
      practice_id: exercise.value.practice.id,
      mode: exercise.value.mode,
      score,
      quality,
      duration_s: durationS,
      date: today,
    })
  } catch (e) {
    console.error('Failed to record score:', e)
  }

  // Update session state
  recentScores.value.push(score)
  if (recentScores.value.length > 5) recentScores.value = recentScores.value.slice(-5)
  exercisesDone.value++
  totalScore.value += score
  lastCardId.value = exercise.value.practice.id

  // Update momentum
  updateMomentum(score)

  showResult.value = true
}

function updateMomentum(_score: number) {
  const recent = recentScores.value
  if (recent.length >= 2) {
    let poor = 0, good = 0
    for (let i = recent.length - 1; i >= Math.max(0, recent.length - 3); i--) {
      if (recent[i]! < 0.6) poor++
      else if (recent[i]! >= 0.8) good++
    }
    if (poor >= 2) momentum.value = 'struggling'
    else if (good >= 3) momentum.value = 'cruising'
    else momentum.value = 'steady'
  }
}

async function nextExercise() {
  showResult.value = false
  await fetchNextExercise()
}

async function continueStudying() {
  keepStudying.value = true
  sessionDone.value = false
  await fetchNextExercise()
}

function exitSession() {
  router.push('/memorize')
}

const scorePercent = computed(() => Math.round(currentScore.value * 100))

const averageScore = computed(() => {
  if (exercisesDone.value === 0) return 0
  return Math.round((totalScore.value / exercisesDone.value) * 100)
})

onMounted(startSession)
</script>

<template>
  <div>
    <!-- Header -->
    <div class="flex items-center justify-between mb-4">
      <div class="flex items-center gap-2">
        <button @click="exitSession" class="text-gray-400 hover:text-gray-600 text-xl">←</button>
        <h1 class="text-2xl font-bold">Study Mode</h1>
      </div>
      <div class="flex items-center gap-3 text-sm">
        <span class="text-gray-500">{{ momentumEmoji }} {{ exercisesDone }} done</span>
        <span v-if="exercisesDone > 0" class="text-gray-400">{{ averageScore }}% avg</span>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="text-center py-12 text-gray-400">Preparing study session...</div>

    <!-- Session Complete -->
    <div v-else-if="sessionDone" class="text-center py-12">
      <div class="text-5xl mb-4">🎉</div>
      <h2 class="text-xl font-bold text-gray-800 mb-2">Session Complete!</h2>
      <p class="text-gray-500 mb-2">{{ sessionMessage }}</p>
      <div v-if="exercisesDone > 0" class="text-sm text-gray-400 mb-6">
        {{ exercisesDone }} exercises · {{ averageScore }}% average accuracy
      </div>
      <div class="flex justify-center gap-3">
        <button @click="continueStudying" class="px-4 py-2 bg-emerald-600 text-white rounded-lg hover:bg-emerald-700">
          Keep Studying
        </button>
        <button @click="exitSession" class="px-4 py-2 bg-gray-200 text-gray-700 rounded-lg hover:bg-gray-300">
          Done
        </button>
      </div>
    </div>

    <!-- Active Exercise -->
    <template v-else-if="exercise">
      <!-- Exercise info bar -->
      <div class="flex items-center gap-2 mb-4 text-sm">
        <span class="px-2 py-1 bg-indigo-100 text-indigo-700 rounded-full text-xs font-medium">{{ levelLabel }}</span>
        <span class="text-gray-500">{{ modeName }}</span>
        <span class="ml-auto px-2 py-0.5 rounded-full text-xs"
          :class="{
            'bg-green-100 text-green-700': exercise.card_type === 'confidence',
            'bg-orange-100 text-orange-700': exercise.card_type === 'stretch',
            'bg-blue-100 text-blue-700': exercise.card_type === 'goldilocks',
            'bg-purple-100 text-purple-700': exercise.card_type === 'fresh',
          }"
        >{{ cardTypeLabel }}</span>
      </div>

      <!-- Card name -->
      <div class="text-sm font-medium text-gray-500 mb-2">
        <span v-if="!exercise.is_reverse">{{ exerciseName }}</span>
        <span v-else class="italic">Which scripture is this from?</span>
      </div>

      <!-- Result overlay -->
      <div v-if="showResult" class="bg-white rounded-2xl shadow-lg border border-gray-100 p-6 mb-6">
        <div class="text-center">
          <div class="text-4xl mb-2">{{ currentScore >= 0.8 ? '✅' : currentScore >= 0.5 ? '👍' : '🔄' }}</div>
          <p class="text-2xl font-bold" :class="currentScore >= 0.8 ? 'text-green-600' : currentScore >= 0.5 ? 'text-yellow-600' : 'text-red-500'">
            {{ scorePercent }}%
          </p>
          <p class="text-sm text-gray-500 mt-1">{{ exerciseName }}</p>
          <div v-if="lastScoreResponse" class="mt-3 text-xs text-gray-400">
            Overall aptitude: {{ Math.round(lastScoreResponse.overall * 100) }}%
          </div>
        </div>
        <div class="flex justify-center gap-3 mt-4">
          <button @click="nextExercise" class="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700">
            Next →
          </button>
          <button @click="exitSession" class="px-4 py-2 bg-gray-200 text-gray-700 rounded-lg hover:bg-gray-300">
            End Session
          </button>
        </div>
      </div>

      <!-- Exercise content (hidden when showing result) -->
      <div v-else>
        <!-- FORWARD: reveal_whole (Level 1 — Read & Absorb) -->
        <div v-if="exercise.mode === 'reveal_whole'" class="bg-white rounded-2xl shadow-lg border border-gray-100 p-6">
          <div class="text-lg leading-relaxed text-gray-800 whitespace-pre-line">{{ exerciseText }}</div>
          <div class="mt-4 text-center">
            <button v-if="!revealed" @click="revealed = true" class="px-4 py-2 bg-indigo-100 text-indigo-700 rounded-lg hover:bg-indigo-200">
              I've read it ✓
            </button>
            <button v-else @click="submitScore" class="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700">
              Continue →
            </button>
          </div>
        </div>

        <!-- FORWARD: reveal_words (Level 2 — Tap to Reveal) -->
        <div v-else-if="exercise.mode === 'reveal_words'" class="bg-white rounded-2xl shadow-lg border border-gray-100 p-6">
          <div class="text-lg leading-relaxed flex flex-wrap gap-1">
            <span v-for="(b, i) in blanks" :key="i">
              <span v-if="!b.hidden" class="text-gray-800">{{ b.word }}</span>
              <button v-else-if="!b.revealed" @click="revealWord(i)"
                class="px-2 py-0.5 bg-indigo-100 text-indigo-400 rounded border border-indigo-200 hover:bg-indigo-200 min-w-[3em] text-center">
                ···
              </button>
              <span v-else class="px-1 text-indigo-600 font-semibold bg-indigo-50 rounded">{{ b.word }}</span>
            </span>
          </div>
          <div class="mt-4 flex justify-between items-center">
            <button v-if="!allBlanksRevealed" @click="revealAllBlanks" class="text-sm text-gray-400 hover:text-gray-600">
              Reveal all
            </button>
            <span v-if="!allBlanksRevealed" class="text-xs text-gray-400">
              {{ blanks.filter(b => b.hidden && !b.revealed).length }} remaining
            </span>
            <button v-if="allBlanksRevealed" @click="submitScore" class="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 ml-auto">
              Continue →
            </button>
          </div>
        </div>

        <!-- FORWARD: type_words (Level 3 — Type Missing Words) -->
        <div v-else-if="exercise.mode === 'type_words'" class="bg-white rounded-2xl shadow-lg border border-gray-100 p-6">
          <div class="text-lg leading-relaxed flex flex-wrap gap-1 items-baseline">
            <template v-for="(tw, i) in typeWords" :key="i">
              <span v-if="!tw.hidden" class="text-gray-800">{{ tw.word }}</span>
              <span v-else-if="!typeChecked">
                <input v-model="tw.userInput" type="text"
                  class="border-b-2 border-indigo-300 outline-none text-center bg-transparent w-20 text-indigo-700 font-medium"
                  :placeholder="'_'.repeat(Math.min(8, tw.word.length))"
                  @keyup.enter="checkTypeWords" />
              </span>
              <span v-else
                class="px-1 rounded font-medium"
                :class="tw.correct ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-600 line-through'">
                {{ tw.correct ? tw.userInput : tw.word }}
                <span v-if="!tw.correct && tw.userInput" class="text-[10px] text-red-400 ml-0.5">({{ tw.userInput }})</span>
              </span>
            </template>
          </div>
          <div class="mt-4 flex justify-end">
            <button v-if="!typeChecked" @click="checkTypeWords" class="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700">
              Check
            </button>
            <div v-else class="flex items-center gap-3">
              <span class="text-sm text-gray-500">{{ typeScore.correct }}/{{ typeScore.total }} correct</span>
              <button @click="submitScore" class="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700">
                Continue →
              </button>
            </div>
          </div>
        </div>

        <!-- FORWARD: arrange (Level 3 — Arrange Words) -->
        <div v-else-if="exercise.mode === 'arrange'" class="bg-white rounded-2xl shadow-lg border border-gray-100 p-6">
          <!-- Placed words area -->
          <div class="min-h-[60px] p-3 bg-gray-50 rounded-lg border-2 border-dashed border-gray-200 mb-4 flex flex-wrap gap-1.5">
            <button v-for="(w, i) in placedWords" :key="'p'+i" @click="unplaceWord(i)"
              class="px-2 py-1 rounded text-sm transition-colors"
              :class="orderChecked
                ? w.originalIndex === i ? 'bg-green-100 text-green-700 border border-green-300' : 'bg-red-100 text-red-600 border border-red-300'
                : 'bg-indigo-100 text-indigo-700 border border-indigo-200 hover:bg-indigo-200'">
              {{ w.word }}
            </button>
            <span v-if="placedWords.length === 0" class="text-gray-300 text-sm italic">Tap words below to build the verse...</span>
          </div>
          <!-- Word bank -->
          <div class="flex flex-wrap gap-1.5 mb-4">
            <button v-for="(w, i) in wordBank" :key="'b'+i" @click="placeWord(i)"
              :disabled="w.placed"
              class="px-2 py-1 rounded text-sm border transition-colors"
              :class="w.placed ? 'bg-gray-100 text-gray-300 border-gray-200' : 'bg-white text-gray-700 border-gray-300 hover:border-indigo-400 hover:text-indigo-600'">
              {{ w.word }}
            </button>
          </div>
          <div class="flex justify-end">
            <button v-if="!orderChecked && placedWords.length === wordBank.length" @click="checkOrder"
              class="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700">
              Check
            </button>
            <div v-if="orderChecked" class="flex items-center gap-3">
              <span class="text-sm text-gray-500">{{ orderScore.correct }}/{{ orderScore.total }} in order</span>
              <button @click="submitScore" class="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700">
                Continue →
              </button>
            </div>
          </div>
        </div>

        <!-- FORWARD: type_full (Level 4 — Type Entire Text) -->
        <div v-else-if="exercise.mode === 'type_full'" class="bg-white rounded-2xl shadow-lg border border-gray-100 p-6">
          <div v-if="!quizChecked">
            <textarea v-model="quizInput" rows="6"
              class="w-full border rounded-lg p-3 text-gray-800 focus:border-indigo-400 focus:ring-1 focus:ring-indigo-200 outline-none resize-none"
              placeholder="Type the full text from memory..."></textarea>
            <div class="flex justify-end mt-3">
              <button @click="checkQuiz" :disabled="!quizInput.trim()"
                class="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 disabled:opacity-40">
                Check
              </button>
            </div>
          </div>
          <div v-else>
            <div class="flex flex-wrap gap-1 text-lg leading-relaxed mb-4">
              <span v-for="(d, i) in quizDiff" :key="i"
                class="px-0.5 rounded"
                :class="{
                  'text-green-700 bg-green-50': d.kind === 'match',
                  'text-red-500 bg-red-50 line-through': d.kind === 'wrong' || d.kind === 'extra',
                  'text-yellow-600 bg-yellow-50 underline': d.kind === 'missing',
                }">
                {{ d.kind === 'extra' ? d.typed : d.word }}
              </span>
            </div>
            <div class="flex items-center justify-between">
              <span class="text-sm text-gray-500">{{ quizScore.correct }}/{{ quizScore.total }} words</span>
              <button @click="submitScore" class="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700">
                Continue →
              </button>
            </div>
          </div>
        </div>

        <!-- REVERSE: all reverse modes -->
        <div v-else-if="exercise.mode.startsWith('reverse_')" class="bg-white rounded-2xl shadow-lg border border-gray-100 p-6">
          <div class="text-lg leading-relaxed text-gray-800 mb-6 whitespace-pre-line"
            :class="{ 'italic text-gray-500': exercise.mode === 'reverse_fragment' }">
            {{ reverseDisplayText }}
          </div>
          <div class="grid grid-cols-2 gap-2 mb-4">
            <button v-for="opt in referenceOptions" :key="opt" @click="selectReference(opt)"
              class="px-3 py-3 rounded-lg border text-sm text-left transition-colors"
              :class="{
                'border-indigo-500 bg-indigo-50 text-indigo-700': selectedReference === opt && !referenceChecked,
                'border-green-500 bg-green-50 text-green-700': referenceChecked && opt === exercise.practice.name,
                'border-red-300 bg-red-50 text-red-500': referenceChecked && selectedReference === opt && opt !== exercise.practice.name,
                'border-gray-200 text-gray-600 hover:border-gray-400': selectedReference !== opt && !referenceChecked,
                'border-gray-200 text-gray-400': referenceChecked && opt !== exercise.practice.name && selectedReference !== opt,
              }">
              {{ opt }}
            </button>
          </div>
          <div class="flex justify-end">
            <button v-if="!referenceChecked && selectedReference" @click="checkReference"
              class="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700">
              Check
            </button>
            <button v-if="referenceChecked" @click="submitScore"
              class="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700">
              Continue →
            </button>
          </div>
        </div>
      </div>
    </template>

    <!-- Momentum indicator (floating) -->
    <div v-if="exercise && !loading && !sessionDone && exercisesDone > 0"
      class="fixed bottom-4 left-1/2 -translate-x-1/2 px-4 py-2 rounded-full text-sm shadow-lg border"
      :class="{
        'bg-red-50 border-red-200 text-red-600': momentum === 'struggling',
        'bg-blue-50 border-blue-200 text-blue-600': momentum === 'steady',
        'bg-green-50 border-green-200 text-green-600': momentum === 'cruising',
      }">
      {{ momentumEmoji }}
      {{ momentum === 'struggling' ? 'Building confidence...' : momentum === 'cruising' ? 'On a roll!' : 'In the zone' }}
    </div>
  </div>
</template>
