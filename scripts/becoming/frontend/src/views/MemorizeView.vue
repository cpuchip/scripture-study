<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { api, type Practice, type MemorizeCardStatus, type MemorizeAptitude, type PillarLink } from '../api'

const route = useRoute()

function localDateStr(d: Date = new Date()): string {
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

const today = localDateStr()
const allCards = ref<MemorizeCardStatus[]>([])
const currentIndex = ref(0)
const loading = ref(true)

// Filter state — multi-select for both categories and pillars
const activeCategories = ref<Set<string>>(new Set())
const activePillarIds = ref<Set<number>>(new Set())
const hasPillars = ref(false)
const cardPillarMap = ref<Record<number, PillarLink[]>>({})

function toggleCategory(cat: string) {
  const s = new Set(activeCategories.value)
  if (s.has(cat)) s.delete(cat); else s.add(cat)
  activeCategories.value = s
}
function clearCategories() { activeCategories.value = new Set() }

function togglePillar(id: number) {
  const s = new Set(activePillarIds.value)
  if (s.has(id)) s.delete(id); else s.add(id)
  activePillarIds.value = s
}
function clearPillars() { activePillarIds.value = new Set() }
function clearAllFilters() { clearCategories(); clearPillars() }

const availableCategories = computed(() => {
  const cats = new Set<string>()
  for (const card of allCards.value) {
    if (card.practice.category) cats.add(card.practice.category.split(',')[0]!.trim())
  }
  return Array.from(cats).sort()
})

const availablePillars = computed(() => {
  const pillars = new Map<number, { id: number; name: string; icon: string }>()
  for (const card of allCards.value) {
    const links = cardPillarMap.value[card.practice.id]
    if (links && links.length > 0) {
      for (const link of links) {
        pillars.set(link.pillar_id, { id: link.pillar_id, name: link.pillar_name, icon: link.pillar_icon })
      }
    }
  }
  return Array.from(pillars.values()).sort((a, b) => a.name.localeCompare(b.name))
})

const hasActiveFilters = computed(() => activeCategories.value.size > 0 || activePillarIds.value.size > 0)

const filteredCards = computed(() => {
  if (!hasActiveFilters.value) return allCards.value
  return allCards.value.filter(card => {
    // Category filter: if any categories selected, card must match one
    if (activeCategories.value.size > 0) {
      const cat = (card.practice.category || '').split(',')[0]?.trim() || ''
      if (!activeCategories.value.has(cat)) return false
    }
    // Pillar filter: if any pillars selected, card must have at least one
    if (activePillarIds.value.size > 0) {
      const links = cardPillarMap.value[card.practice.id]
      if (!links || !links.some(l => activePillarIds.value.has(l.pillar_id))) return false
    }
    return true
  })
})

const studyLink = computed(() => {
  const params = new URLSearchParams()
  if (activeCategories.value.size > 0) {
    params.set('category', Array.from(activeCategories.value).join(','))
  }
  if (activePillarIds.value.size > 0) {
    params.set('pillar_ids', Array.from(activePillarIds.value).join(','))
  }
  const qs = params.toString()
  return qs ? `/study?${qs}` : '/study'
})

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
interface ArrangeSlot { anchor: boolean; word: OrderWord | null }
const wordBank = ref<OrderWord[]>([])
const arrangeSlots = ref<ArrangeSlot[]>([])
const correctWords = ref<string[]>([])
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
const allDone = computed(() => filteredCards.value.length > 0 && filteredCards.value.every(c => c.reviews_today >= c.target_daily_reps))
const currentCardDone = computed(() => {
  const s = currentStatus.value
  return s ? s.reviews_today >= s.target_daily_reps : false
})
const nextIncompleteIndex = computed(() => {
  // Find next incomplete card in allCards (not filteredCards) since currentIndex indexes into allCards
  for (let i = 0; i < allCards.value.length; i++) {
    const idx = (currentIndex.value + 1 + i) % allCards.value.length
    if (idx === currentIndex.value) continue
    const c = allCards.value[idx]
    if (c && c.reviews_today < c.target_daily_reps) {
      // If filtering, only consider cards that pass the filter
      if (hasActiveFilters.value) {
        if (!filteredCards.value.includes(c)) continue
      }
      return idx
    }
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

  // Load pillar mappings
  try {
    const pillarsCheck = await api.hasPillars()
    hasPillars.value = pillarsCheck.has_pillars
    if (hasPillars.value) {
      const mapping: Record<number, PillarLink[]> = {}
      await Promise.all(allCards.value.map(async (card) => {
        try {
          const links = await api.getPracticePillars(card.practice.id)
          if (links.length > 0) mapping[card.practice.id] = links
        } catch { /* noop */ }
      }))
      cardPillarMap.value = mapping
    }
  } catch { /* pillars optional */ }

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
  arrangeSlots.value = []
  orderChecked.value = false
  quizInput.value = ''
  quizChecked.value = false
  quizDiff.value = []
  suggestedQuality.value = null
  justRated.value = false
  showCardDetail.value = false
  showEndDatePicker.value = false
  showActionsMenu.value = false

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
  const allWords: OrderWord[] = words.map((w, i) => ({ word: w, originalIndex: i, placed: false }))

  // Pick ~33% of words as anchors (hints), evenly distributed across full text
  const anchorRatio = 0.33
  const anchorCount = Math.max(0, Math.floor(words.length * anchorRatio))
  const anchorIndices = new Set<number>()
  if (anchorCount > 0 && words.length > 0) {
    // Use floating-point spacing to cover the full range [0, words.length-1]
    const span = words.length - 1
    for (let k = 0; k < anchorCount; k++) {
      const idx = Math.round((k + 0.5) * span / anchorCount)
      if (idx >= 0 && idx < words.length) anchorIndices.add(idx)
    }
  }

  // Build slots: anchors are pre-filled, rest are empty
  const slots: ArrangeSlot[] = allWords.map((w, i) => {
    if (anchorIndices.has(i)) {
      return { anchor: true, word: w }
    }
    return { anchor: false, word: null }
  })

  // Word bank contains only non-anchor words, shuffled
  const bank = allWords.filter((_, i) => !anchorIndices.has(i))
  shuffle(bank)
  bank.forEach(w => w.placed = false)

  arrangeSlots.value = slots
  wordBank.value = bank
  correctWords.value = words
  orderChecked.value = false
}

function placeWord(bankIndex: number) {
  if (orderChecked.value) return
  const word = wordBank.value[bankIndex]!
  if (word.placed) return
  // Find next empty (non-anchor) slot
  const slotIdx = arrangeSlots.value.findIndex(s => !s.anchor && s.word === null)
  if (slotIdx < 0) return
  word.placed = true
  arrangeSlots.value[slotIdx]!.word = word
}

function unplaceWord(slotIndex: number) {
  if (orderChecked.value) return
  const slot = arrangeSlots.value[slotIndex]!
  if (slot.anchor || !slot.word) return
  const bankIdx = wordBank.value.findIndex(w => w === slot.word)
  if (bankIdx >= 0) wordBank.value[bankIdx]!.placed = false
  slot.word = null
}

function checkOrder() {
  let correct = 0
  let total = 0
  arrangeSlots.value.forEach((slot, i) => {
    if (slot.anchor) return // anchors don't count
    total++
    if (slot.word && slot.word.word === correctWords.value[i]) correct++
  })
  orderChecked.value = true
  orderScore.value = { correct, total }
  suggestQuality(total > 0 ? correct / total : 0)
}

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

// --- Card detail panel ---
const showCardDetail = ref(false)
const showEndDatePicker = ref(false)
const endDateInput = ref('')
const showActionsMenu = ref(false)

// Level labels for display
const levelLabels: Record<number, string> = {
  1: 'Reveal Whole',
  2: 'Reveal Words',
  3: 'Type Words',
  4: 'Arrange',
  5: 'Type Full',
}

// Mode labels for aptitude display
const modeLabels: Record<string, string> = {
  reveal_whole: 'Reveal Whole',
  reveal_words: 'Reveal Words',
  type_words: 'Type Words',
  arrange: 'Arrange',
  type_full: 'Type Full',
  reverse_full: 'Reverse Full',
  reverse_partial: 'Reverse Partial',
  reverse_fragment: 'Reverse Fragment',
}

// Ordered modes for aptitude bars
const forwardModes = ['reveal_whole', 'reveal_words', 'type_words', 'arrange', 'type_full']
const reverseModes = ['reverse_full', 'reverse_partial', 'reverse_fragment']

function toggleCardDetail() {
  showCardDetail.value = !showCardDetail.value
  showActionsMenu.value = false
}

function getAptitudeForMode(aptitudes: MemorizeAptitude[], mode: string): MemorizeAptitude | undefined {
  return aptitudes.find(a => a.mode === mode)
}

function aptitudeColor(value: number): string {
  if (value >= 0.8) return 'bg-green-500'
  if (value >= 0.6) return 'bg-yellow-400'
  if (value >= 0.4) return 'bg-orange-400'
  return 'bg-red-400'
}

function levelColor(level: number): string {
  if (level >= 5) return 'bg-green-600'
  if (level >= 4) return 'bg-green-500'
  if (level >= 3) return 'bg-yellow-500'
  if (level >= 2) return 'bg-orange-400'
  return 'bg-gray-400'
}

// Lifecycle actions
async function markMastered() {
  if (!currentCard.value) return
  if (!confirm(`Mark "${currentCard.value.name}" as memorized?`)) return
  await api.completePractice(currentCard.value.id)
  await load()
}

async function archiveCard() {
  if (!currentCard.value) return
  if (!confirm(`Archive "${currentCard.value.name}"? This removes it from rotation.`)) return
  await api.archivePractice(currentCard.value.id)
  await load()
}

async function pauseCard() {
  if (!currentCard.value) return
  await api.pausePractice(currentCard.value.id)
  await load()
}

// End date
async function setEndDate() {
  if (!currentCard.value || !endDateInput.value) return
  // Fetch full practice, update end_date, send back
  const full = await api.getPractice(currentCard.value.id)
  full.end_date = endDateInput.value
  await api.updatePractice(currentCard.value.id, full)
  showEndDatePicker.value = false
  await load()
}

async function clearEndDate() {
  if (!currentCard.value) return
  const full = await api.getPractice(currentCard.value.id)
  full.end_date = ''
  await api.updatePractice(currentCard.value.id, full)
  showEndDatePicker.value = false
  await load()
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
          {{ filteredCards.filter(c => c.reviews_today >= c.target_daily_reps).length }}/{{ filteredCards.length }} cards done
        </span>
        <router-link
          :to="studyLink"
          class="px-3 py-1.5 bg-emerald-600 text-white rounded-lg text-sm hover:bg-emerald-700"
        >📚 Study</router-link>
        <router-link
          to="/practices?type=memorize&create=1"
          class="px-3 py-1.5 bg-indigo-600 text-white rounded-lg text-sm hover:bg-indigo-700"
        >+ Add Card</router-link>
      </div>
    </div>

    <!-- Filter chips -->
    <div v-if="!loading && allCards.length > 1 && (availableCategories.length > 1 || availablePillars.length > 0)" class="mb-4 space-y-2">
      <!-- Category chips -->
      <div v-if="availableCategories.length > 0" class="flex items-center gap-1.5 flex-wrap">
        <span class="text-xs text-gray-400 mr-0.5">Category:</span>
        <button
          v-for="cat in availableCategories"
          :key="cat"
          @click="toggleCategory(cat)"
          class="px-2.5 py-1 text-xs rounded-full border transition-colors capitalize"
          :class="activeCategories.has(cat) ? 'bg-indigo-600 border-indigo-600 text-white' : 'bg-white border-gray-200 text-gray-600 hover:bg-gray-50'"
        >{{ cat }}</button>
      </div>
      <!-- Pillar chips -->
      <div v-if="hasPillars && availablePillars.length > 0" class="flex items-center gap-1.5 flex-wrap">
        <span class="text-xs text-gray-400 mr-0.5">Pillar:</span>
        <button
          v-for="p in availablePillars"
          :key="p.id"
          @click="togglePillar(p.id)"
          class="px-2.5 py-1 text-xs rounded-full border transition-colors"
          :class="activePillarIds.has(p.id) ? 'bg-purple-600 border-purple-600 text-white' : 'bg-white border-gray-200 text-gray-600 hover:bg-gray-50'"
        >{{ p.icon }} {{ p.name }}</button>
      </div>
      <!-- Clear all -->
      <div v-if="hasActiveFilters" class="flex items-center">
        <button @click="clearAllFilters" class="text-xs text-gray-400 hover:text-gray-600 underline">
          Clear filters
        </button>
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
          v-for="card in filteredCards"
          :key="card.practice.id"
          @click="selectCard(allCards.indexOf(card))"
          class="flex items-center gap-1.5 px-3 py-1.5 rounded-full text-sm whitespace-nowrap transition-colors border"
          :class="allCards.indexOf(card) === currentIndex
            ? 'bg-indigo-600 text-white border-indigo-600'
            : card.reviews_today >= card.target_daily_reps
              ? 'bg-green-50 border-green-300 text-green-700 hover:bg-green-100'
              : 'bg-white border-gray-200 text-gray-600 hover:border-indigo-300'"
        >
          <span v-if="card.is_mastered" class="text-[10px]">⭐</span>
          {{ card.practice.name }}
          <span
            class="text-[10px] px-1.5 py-0.5 rounded-full"
            :class="allCards.indexOf(card) === currentIndex
              ? 'bg-indigo-500 text-indigo-100'
              : card.reviews_today >= card.target_daily_reps
                ? 'bg-green-200 text-green-800'
                : 'bg-gray-100 text-gray-500'"
          >{{ card.reviews_today }}/{{ card.target_daily_reps }}</span>
          <span v-if="card.days_until_end != null && card.days_until_end <= 7"
            class="text-[10px] px-1 py-0.5 rounded-full"
            :class="allCards.indexOf(card) === currentIndex ? 'bg-red-400 text-white' : 'bg-red-100 text-red-600'">
            {{ card.days_until_end > 0 ? card.days_until_end + 'd' : '!' }}
          </span>
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
              <!-- Slot-based verse layout -->
              <div class="min-h-[60px] p-3 bg-gray-50 rounded-lg border-2 border-dashed border-gray-200 mb-4 flex flex-wrap gap-1.5 items-baseline">
                <template v-for="(slot, i) in arrangeSlots">
                  <!-- Anchor word (given hint, locked) -->
                  <span v-if="slot.anchor && slot.word" :key="'a'+i"
                    class="px-2 py-1 rounded text-sm cursor-default"
                    :class="orderChecked ? 'bg-green-100 text-green-700 border border-green-300' : 'bg-gray-200 text-gray-500 border border-gray-300'">
                    {{ slot.word.word }}
                  </span>
                  <!-- User-placed word -->
                  <button v-else-if="slot.word" :key="'u'+i" @click="unplaceWord(i)"
                    class="px-2 py-1 rounded text-sm transition-colors"
                    :class="orderChecked
                      ? slot.word.word === correctWords[i] ? 'bg-green-100 text-green-700 border border-green-300' : 'bg-red-100 text-red-600 border border-red-300'
                      : 'bg-indigo-100 text-indigo-700 border border-indigo-200 hover:bg-red-50 hover:text-red-600 hover:border-red-300 cursor-pointer'">
                    {{ slot.word.word }}
                  </button>
                  <!-- Empty slot -->
                  <span v-else :key="'e'+i" class="w-8 h-6 border-b-2 border-gray-300 inline-block"></span>
                </template>
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
                <span class="text-gray-400">{{ arrangeSlots.filter(s => !s.anchor && s.word !== null).length }}/{{ wordBank.length }} words placed</span>
                <button
                  @click="checkOrder"
                  :disabled="!arrangeSlots.some(s => !s.anchor && s.word !== null)"
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

        <!-- Card stats / detail panel toggle -->
        <div class="mt-4">
          <!-- Compact stats bar (always visible) -->
          <button @click="toggleCardDetail" class="w-full flex items-center justify-center gap-4 text-xs text-gray-400 hover:text-gray-600 transition-colors py-2">
            <!-- Level badge -->
            <span class="flex items-center gap-1">
              <span class="w-5 h-5 rounded-full text-white font-bold text-[10px] flex items-center justify-center" :class="levelColor(currentCard.memorize_level || 1)">
                L{{ currentCard.memorize_level || 1 }}
              </span>
              <span>{{ levelLabels[currentCard.memorize_level || 1] }}</span>
            </span>
            <!-- Overall aptitude -->
            <span v-if="currentStatus">{{ Math.round(currentStatus.overall_aptitude * 100) }}% apt</span>
            <!-- SM-2 interval -->
            <span>{{ cardStats(currentCard).interval }}d interval</span>
            <!-- End date countdown -->
            <span v-if="currentStatus?.days_until_end != null"
              :class="currentStatus.days_until_end <= 7 ? 'text-red-500 font-semibold' : currentStatus.days_until_end <= 14 ? 'text-orange-500' : ''">
              {{ currentStatus.days_until_end > 0 ? currentStatus.days_until_end + 'd left' : currentStatus.days_until_end === 0 ? 'Due today!' : Math.abs(currentStatus.days_until_end) + 'd overdue' }}
            </span>
            <!-- Mastery badge -->
            <span v-if="currentStatus?.is_mastered" class="text-green-600 font-semibold">⭐ Mastered</span>
            <!-- Toggle arrow -->
            <span class="transition-transform" :class="showCardDetail ? 'rotate-180' : ''">▾</span>
          </button>

          <!-- Expanded detail panel -->
          <transition name="slide">
            <div v-if="showCardDetail" class="bg-white rounded-xl border border-gray-200 p-4 mt-1 space-y-4">
              <!-- Mastery suggestion banner -->
              <div v-if="currentStatus?.is_mastered" class="bg-green-50 border border-green-200 rounded-lg p-3 flex items-center justify-between">
                <div>
                  <p class="text-green-800 font-semibold text-sm">⭐ This card meets mastery criteria</p>
                  <p class="text-green-600 text-xs">High aptitude across multiple modes with strong SM-2 interval</p>
                </div>
                <button @click="markMastered" class="px-3 py-1.5 bg-green-600 text-white rounded-lg text-sm hover:bg-green-700 whitespace-nowrap">
                  Mark Memorized ✓
                </button>
              </div>

              <!-- SM-2 stats row -->
              <div class="flex items-center justify-between text-sm">
                <div class="flex gap-4 text-gray-500">
                  <span>Interval: <strong class="text-gray-700">{{ cardStats(currentCard).interval }}d</strong></span>
                  <span>Reps: <strong class="text-gray-700">{{ cardStats(currentCard).reps }}</strong></span>
                  <span>Ease: <strong class="text-gray-700">{{ cardStats(currentCard).ease }}</strong></span>
                  <span>Next: <strong class="text-gray-700">{{ cardStats(currentCard).nextReview }}</strong></span>
                </div>
              </div>

              <!-- Forward mode aptitudes -->
              <div>
                <h4 class="text-xs font-semibold text-gray-500 uppercase tracking-wide mb-2">Forward Modes</h4>
                <div class="space-y-1.5">
                  <div v-for="mode in forwardModes" :key="mode" class="flex items-center gap-2">
                    <span class="text-xs text-gray-500 w-24 truncate">{{ modeLabels[mode] }}</span>
                    <div class="flex-1 bg-gray-100 rounded-full h-3 relative overflow-hidden">
                      <div
                        v-if="currentStatus && getAptitudeForMode(currentStatus.aptitudes, mode)"
                        class="h-full rounded-full transition-all"
                        :class="aptitudeColor(getAptitudeForMode(currentStatus.aptitudes, mode)!.aptitude)"
                        :style="{ width: Math.round(getAptitudeForMode(currentStatus.aptitudes, mode)!.aptitude * 100) + '%' }"
                      ></div>
                    </div>
                    <span class="text-xs text-gray-500 w-10 text-right">
                      {{ currentStatus && getAptitudeForMode(currentStatus.aptitudes, mode)
                        ? Math.round(getAptitudeForMode(currentStatus.aptitudes, mode)!.aptitude * 100) + '%'
                        : '—' }}
                    </span>
                    <!-- Current level indicator -->
                    <span v-if="(currentCard.memorize_level || 1) === forwardModes.indexOf(mode) + 1"
                      class="text-[10px] px-1.5 py-0.5 bg-indigo-100 text-indigo-600 rounded-full">current</span>
                  </div>
                </div>
              </div>

              <!-- Reverse mode aptitudes -->
              <div v-if="currentStatus && currentStatus.aptitudes.some(a => reverseModes.includes(a.mode))">
                <h4 class="text-xs font-semibold text-gray-500 uppercase tracking-wide mb-2">Reverse Modes</h4>
                <div class="space-y-1.5">
                  <div v-for="mode in reverseModes" :key="mode" class="flex items-center gap-2">
                    <span class="text-xs text-gray-500 w-24 truncate">{{ modeLabels[mode] }}</span>
                    <div class="flex-1 bg-gray-100 rounded-full h-3 relative overflow-hidden">
                      <div
                        v-if="currentStatus && getAptitudeForMode(currentStatus.aptitudes, mode)"
                        class="h-full rounded-full transition-all"
                        :class="aptitudeColor(getAptitudeForMode(currentStatus.aptitudes, mode)!.aptitude)"
                        :style="{ width: Math.round(getAptitudeForMode(currentStatus.aptitudes, mode)!.aptitude * 100) + '%' }"
                      ></div>
                    </div>
                    <span class="text-xs text-gray-500 w-10 text-right">
                      {{ currentStatus && getAptitudeForMode(currentStatus.aptitudes, mode)
                        ? Math.round(getAptitudeForMode(currentStatus.aptitudes, mode)!.aptitude * 100) + '%'
                        : '—' }}
                    </span>
                  </div>
                </div>
              </div>

              <!-- Overall aptitude -->
              <div class="flex items-center gap-3 pt-2 border-t border-gray-100">
                <span class="text-sm font-semibold text-gray-700">Overall</span>
                <div class="flex-1 bg-gray-100 rounded-full h-4 relative overflow-hidden">
                  <div
                    v-if="currentStatus"
                    class="h-full rounded-full transition-all"
                    :class="aptitudeColor(currentStatus.overall_aptitude)"
                    :style="{ width: Math.round(currentStatus.overall_aptitude * 100) + '%' }"
                  ></div>
                </div>
                <span class="text-sm font-bold text-gray-700 w-12 text-right">
                  {{ currentStatus ? Math.round(currentStatus.overall_aptitude * 100) + '%' : '—' }}
                </span>
              </div>

              <!-- End date -->
              <div class="pt-2 border-t border-gray-100">
                <div class="flex items-center justify-between">
                  <span class="text-sm text-gray-600">
                    <template v-if="currentCard.end_date">
                      📅 Memorize by: <strong>{{ currentCard.end_date.substring(0, 10) }}</strong>
                      <span v-if="currentStatus?.days_until_end != null" class="ml-1"
                        :class="currentStatus.days_until_end <= 7 ? 'text-red-500' : currentStatus.days_until_end <= 14 ? 'text-orange-500' : 'text-gray-400'">
                        ({{ currentStatus.days_until_end > 0 ? currentStatus.days_until_end + ' days left' : currentStatus.days_until_end === 0 ? 'today!' : Math.abs(currentStatus.days_until_end) + ' days overdue' }})
                      </span>
                    </template>
                    <template v-else>
                      📅 No target date set
                    </template>
                  </span>
                  <button @click="showEndDatePicker = !showEndDatePicker; endDateInput = (currentCard.end_date || '').substring(0, 10)"
                    class="text-xs text-indigo-500 hover:text-indigo-700">
                    {{ currentCard.end_date ? 'Change' : 'Set date' }}
                  </button>
                </div>
                <!-- End date picker -->
                <div v-if="showEndDatePicker" class="mt-2 flex items-center gap-2">
                  <input type="date" v-model="endDateInput"
                    class="border border-gray-300 rounded px-2 py-1 text-sm focus:border-indigo-500 focus:ring-1 focus:ring-indigo-200 outline-none" />
                  <button @click="setEndDate" :disabled="!endDateInput"
                    class="px-3 py-1 bg-indigo-600 text-white rounded text-sm hover:bg-indigo-700 disabled:opacity-40">Save</button>
                  <button v-if="currentCard.end_date" @click="clearEndDate"
                    class="px-3 py-1 text-red-500 hover:text-red-700 text-sm">Clear</button>
                  <button @click="showEndDatePicker = false"
                    class="px-3 py-1 text-gray-400 hover:text-gray-600 text-sm">Cancel</button>
                </div>
              </div>

              <!-- Actions -->
              <div class="pt-2 border-t border-gray-100 flex items-center gap-2">
                <button @click="markMastered"
                  class="px-3 py-1.5 bg-green-50 text-green-700 border border-green-200 rounded-lg text-xs hover:bg-green-100">
                  ✓ Mark Memorized
                </button>
                <button @click="pauseCard"
                  class="px-3 py-1.5 bg-yellow-50 text-yellow-700 border border-yellow-200 rounded-lg text-xs hover:bg-yellow-100">
                  ⏸ Pause
                </button>
                <button @click="archiveCard"
                  class="px-3 py-1.5 bg-red-50 text-red-700 border border-red-200 rounded-lg text-xs hover:bg-red-100">
                  🗃 Archive
                </button>
              </div>
            </div>
          </transition>
        </div>
      </div>
    </template>
  </div>
</template>

<style scoped>
.slide-enter-active, .slide-leave-active {
  transition: all 0.2s ease;
  overflow: hidden;
}
.slide-enter-from, .slide-leave-to {
  opacity: 0;
  max-height: 0;
  transform: translateY(-8px);
}
.slide-enter-to, .slide-leave-from {
  opacity: 1;
  max-height: 800px;
}
</style>