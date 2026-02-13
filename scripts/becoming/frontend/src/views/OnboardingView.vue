<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { api, type Pillar } from '../api'

const router = useRouter()
const step = ref(1) // 1=Welcome, 2=Pillars, 3=Starter Practices, 4=Done
const loading = ref(true)

// Step 2: Pillar selection
const suggestions = ref<Pillar[]>([])
const selectedSuggestions = ref<Set<number>>(new Set())
const createdPillars = ref<Pillar[]>([])

// Step 3: Starter practice suggestions
const starterPractices = ref([
  { key: 'scripture', name: 'Scripture Study', description: 'Daily scripture reading', type: 'habit' as const, category: 'spiritual', icon: '📖', selected: true },
  { key: 'prayer', name: 'Morning Prayer', description: 'Start the day with prayer', type: 'habit' as const, category: 'spiritual', icon: '🙏', selected: true },
  { key: 'memorize', name: 'Scripture Memorization', description: 'Memorize key verses', type: 'memorize' as const, category: 'scripture', icon: '🧠', selected: false },
  { key: 'exercise', name: 'Exercise', description: 'Physical activity or workout', type: 'tracker' as const, category: 'fitness', icon: '🏃', selected: false },
  { key: 'journal', name: 'Journal / Reflect', description: 'Write daily reflections', type: 'habit' as const, category: 'spiritual', icon: '✍️', selected: false },
  { key: 'service', name: 'Act of Service', description: 'Serve someone each day', type: 'tracker' as const, category: 'social', icon: '🤝', selected: false },
])

const selectedPracticeCount = computed(() => starterPractices.value.filter(p => p.selected).length)

// Pillar → category mapping for auto-linking
const pillarCategoryMap: Record<string, string[]> = {
  'Spiritual': ['spiritual', 'scripture'],
  'Social': ['social'],
  'Intellectual': ['study'],
  'Physical': ['fitness', 'health'],
}

onMounted(async () => {
  try {
    // Check if user already has pillars → skip to practices or redirect
    const has = await api.hasPillars()
    if (has.has_pillars) {
      const practices = await api.listPractices()
      if (practices.length > 0) {
        // Already set up — go to daily
        localStorage.setItem('onboarding_complete', 'true')
        router.replace('/')
        return
      }
      // Has pillars but no practices — skip to step 3
      createdPillars.value = await api.listPillarsFlat()
      step.value = 3
    }

    const s = await api.getPillarSuggestions()
    suggestions.value = s
    selectedSuggestions.value = new Set(s.map((_, i) => i))
  } finally {
    loading.value = false
  }
})

async function createPillars() {
  loading.value = true
  try {
    for (const [idx, sug] of suggestions.value.entries()) {
      if (selectedSuggestions.value.has(idx)) {
        const p = await api.createPillar({ name: sug.name, description: sug.description, icon: sug.icon })
        createdPillars.value.push(p)
      }
    }
    step.value = 3
  } finally {
    loading.value = false
  }
}

async function createStarterPractices() {
  loading.value = true
  try {
    for (const sp of starterPractices.value) {
      if (!sp.selected) continue
      const practice: Record<string, unknown> = {
        name: sp.name,
        description: sp.description,
        type: sp.type,
        category: sp.category,
      }
      if (sp.type === 'tracker') {
        practice.config = JSON.stringify({ target_sets: 1, target_reps: 1, unit: 'times' })
      } else if (sp.type === 'memorize') {
        practice.config = JSON.stringify({ target_daily_reps: 3 })
      } else {
        practice.config = '{}'
      }

      const created = await api.createPractice(practice)

      // Auto-link to matching pillars
      if (created?.id) {
        const matchingPillarIds: number[] = []
        for (const pillar of createdPillars.value) {
          const categories = pillarCategoryMap[pillar.name] || []
          if (categories.includes(sp.category)) {
            matchingPillarIds.push(pillar.id)
          }
        }
        if (matchingPillarIds.length > 0) {
          await api.setPracticePillars(created.id, matchingPillarIds)
        }
      }
    }
    step.value = 4
  } finally {
    loading.value = false
  }
}

function finishOnboarding() {
  localStorage.setItem('onboarding_complete', 'true')
  router.replace('/')
}

function skipPillars() {
  step.value = 3
}

function skipPractices() {
  step.value = 4
}
</script>

<template>
  <div class="min-h-[80vh] flex items-center justify-center px-4">
    <div class="max-w-lg w-full">
      <!-- Loading -->
      <div v-if="loading" class="text-center py-12 text-gray-400">Setting things up…</div>

      <template v-else>
        <!-- Step 1: Welcome -->
        <div v-if="step === 1" class="text-center">
          <div class="text-5xl mb-4">🌱</div>
          <h1 class="text-3xl font-bold mb-3 text-gray-800">Welcome to Become</h1>
          <p class="text-gray-500 mb-2 max-w-md mx-auto">
            A place to track the practices that help you grow — spiritually, socially, intellectually, and physically.
          </p>
          <p class="text-sm text-gray-400 italic mb-8 max-w-md mx-auto">
            "Jesus increased in wisdom and stature, and in favour with God and man" — Luke 2:52
          </p>
          <button
            @click="step = 2"
            class="px-8 py-3 bg-indigo-600 text-white rounded-xl hover:bg-indigo-700 text-lg font-semibold shadow-md"
          >
            Get Started
          </button>
        </div>

        <!-- Step 2: Pick Pillars -->
        <div v-else-if="step === 2" class="text-center">
          <div class="text-4xl mb-3">🏛️</div>
          <h2 class="text-2xl font-bold mb-2">Choose Your Pillars</h2>
          <p class="text-gray-500 text-sm mb-6 max-w-md mx-auto">
            These are the areas of growth you want to focus on. Based on the Savior's pattern in Luke 2:52, we suggest four — but you can customize later.
          </p>

          <div class="grid grid-cols-2 gap-4 mb-8">
            <button
              v-for="(sug, idx) in suggestions"
              :key="idx"
              @click="selectedSuggestions.has(idx) ? selectedSuggestions.delete(idx) : selectedSuggestions.add(idx)"
              class="p-4 rounded-xl border-2 text-left transition-all"
              :class="selectedSuggestions.has(idx)
                ? 'border-indigo-400 bg-indigo-50 shadow-md'
                : 'border-gray-200 bg-white hover:border-gray-300'"
            >
              <span class="text-3xl block mb-2">{{ sug.icon }}</span>
              <span class="font-semibold">{{ sug.name }}</span>
              <p class="text-xs text-gray-500 mt-1">{{ sug.description }}</p>
            </button>
          </div>

          <div class="flex justify-center gap-3">
            <button @click="skipPillars" class="px-4 py-2 text-gray-500 hover:text-gray-700 text-sm">
              Skip for now
            </button>
            <button
              @click="createPillars"
              :disabled="selectedSuggestions.size === 0"
              class="px-6 py-2.5 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 font-semibold disabled:opacity-40"
            >
              Create {{ selectedSuggestions.size }} Pillar{{ selectedSuggestions.size !== 1 ? 's' : '' }}
            </button>
          </div>
        </div>

        <!-- Step 3: Starter Practices -->
        <div v-else-if="step === 3" class="text-center">
          <div class="text-4xl mb-3">⚡</div>
          <h2 class="text-2xl font-bold mb-2">Start Some Practices</h2>
          <p class="text-gray-500 text-sm mb-6 max-w-md mx-auto">
            Pick a few practices to begin tracking. You can add, edit, or remove these anytime.
          </p>

          <div class="space-y-3 mb-8 text-left">
            <button
              v-for="sp in starterPractices"
              :key="sp.key"
              @click="sp.selected = !sp.selected"
              class="w-full flex items-center gap-3 p-3 rounded-xl border-2 transition-all"
              :class="sp.selected
                ? 'border-indigo-400 bg-indigo-50'
                : 'border-gray-200 bg-white hover:border-gray-300'"
            >
              <span class="text-2xl w-10 text-center flex-shrink-0">{{ sp.icon }}</span>
              <div class="flex-1 min-w-0">
                <div class="font-semibold text-gray-800">{{ sp.name }}</div>
                <div class="text-xs text-gray-500">{{ sp.description }}</div>
              </div>
              <div class="w-6 h-6 rounded-full border-2 flex items-center justify-center flex-shrink-0"
                   :class="sp.selected ? 'border-indigo-500 bg-indigo-500' : 'border-gray-300'">
                <svg v-if="sp.selected" class="w-3.5 h-3.5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7" />
                </svg>
              </div>
            </button>
          </div>

          <div class="flex justify-center gap-3">
            <button @click="skipPractices" class="px-4 py-2 text-gray-500 hover:text-gray-700 text-sm">
              Skip for now
            </button>
            <button
              @click="createStarterPractices"
              :disabled="selectedPracticeCount === 0"
              class="px-6 py-2.5 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 font-semibold disabled:opacity-40"
            >
              Create {{ selectedPracticeCount }} Practice{{ selectedPracticeCount !== 1 ? 's' : '' }}
            </button>
          </div>
        </div>

        <!-- Step 4: All Done -->
        <div v-else-if="step === 4" class="text-center">
          <div class="text-5xl mb-4">🎉</div>
          <h2 class="text-2xl font-bold mb-3">You're All Set!</h2>
          <p class="text-gray-500 mb-2 max-w-md mx-auto">
            Your pillars and practices are ready. Each day, come back to log your progress and reflect on your growth.
          </p>
          <p class="text-sm text-gray-400 italic mb-8 max-w-md mx-auto">
            "Whatever principle of intelligence we attain unto in this life, it will rise with us in the resurrection." — D&C 130:18
          </p>
          <button
            @click="finishOnboarding"
            class="px-8 py-3 bg-indigo-600 text-white rounded-xl hover:bg-indigo-700 text-lg font-semibold shadow-md"
          >
            Go to Today
          </button>
        </div>
      </template>
    </div>
  </div>
</template>
