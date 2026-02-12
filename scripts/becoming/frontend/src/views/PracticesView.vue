<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { api, type Practice } from '../api'

const router = useRouter()
const practices = ref<Practice[]>([])
const loading = ref(true)
const showForm = ref(false)
const editingId = ref<number | null>(null)

// Filters
const filterType = ref<string>('all')
const filterCategory = ref<string>('all')

// Form state
const form = ref({
  name: '',
  description: '',
  type: 'habit' as Practice['type'],
  category: '',
  config: '{}',
})

// Config helpers for tracker (was exercise)
const trackerConfig = ref({
  target_sets: 2,
  target_reps: 15,
  unit: 'reps',
})

// Config helpers for memorize daily reps
const memorizeConfig = ref({
  target_daily_reps: 1,
})

const presetCategories = ['spiritual', 'scripture', 'pt', 'fitness', 'study', 'health']

// Derived filter values from data
const availableTypes = computed(() => {
  const types = new Set(practices.value.map(p => p.type))
  return Array.from(types).sort()
})
const availableCategories = computed(() => {
  const cats = new Set(practices.value.filter(p => p.category).map(p => p.category))
  return Array.from(cats).sort()
})

const filteredPractices = computed(() => {
  return practices.value.filter(p => {
    if (filterType.value !== 'all' && p.type !== filterType.value) return false
    if (filterCategory.value !== 'all' && p.category !== filterCategory.value) return false
    return true
  })
})

async function load() {
  loading.value = true
  practices.value = await api.listPractices(undefined, false)
  loading.value = false
}

async function submit() {
  const p: Partial<Practice> = {
    name: form.value.name,
    description: form.value.description,
    type: form.value.type,
    category: form.value.category,
  }

  if (form.value.type === 'tracker') {
    p.config = JSON.stringify(trackerConfig.value)
  } else if (form.value.type === 'memorize') {
    // For new memorize cards, set target_daily_reps; for edits, merge into existing config
    if (editingId.value !== null) {
      const existing = practices.value.find(pr => pr.id === editingId.value)
      if (existing) {
        try {
          const cfg = JSON.parse(existing.config)
          cfg.target_daily_reps = memorizeConfig.value.target_daily_reps
          p.config = JSON.stringify(cfg)
        } catch {
          p.config = existing.config
        }
      }
    }
    // For new cards, config will be set by backend with DefaultSM2Config
    // but we pass target_daily_reps hint
    if (!editingId.value) {
      p.config = JSON.stringify({ target_daily_reps: memorizeConfig.value.target_daily_reps })
    }
  } else {
    p.config = '{}'
  }

  if (editingId.value !== null) {
    // Merge form changes onto the existing practice to preserve all fields (active, sort_order, etc.)
    const existing = practices.value.find(pr => pr.id === editingId.value)
    if (existing) {
      const merged = { ...existing, ...p }
      // Config was already set correctly above for each type
      await api.updatePractice(editingId.value, merged as Practice)
    }
  } else {
    await api.createPractice(p)
  }
  resetForm()
  await load()
}

function resetForm() {
  form.value = { name: '', description: '', type: 'habit', category: '', config: '{}' }
  trackerConfig.value = { target_sets: 2, target_reps: 15, unit: 'reps' }
  memorizeConfig.value = { target_daily_reps: 1 }
  editingId.value = null
  showForm.value = false
}

function editPractice(p: Practice) {
  editingId.value = p.id
  form.value.name = p.name
  form.value.description = p.description
  form.value.type = p.type
  form.value.category = p.category
  form.value.config = p.config

  // Populate tracker config if editing a tracker
  if (p.type === 'tracker' && p.config) {
    try {
      const cfg = JSON.parse(p.config)
      trackerConfig.value = {
        target_sets: cfg.target_sets ?? 2,
        target_reps: cfg.target_reps ?? 15,
        unit: cfg.unit ?? 'reps',
      }
    } catch {
      // keep defaults
    }
  }

  // Populate memorize daily reps if editing a memorize card
  if (p.type === 'memorize' && p.config) {
    try {
      const cfg = JSON.parse(p.config)
      memorizeConfig.value = {
        target_daily_reps: cfg.target_daily_reps ?? 1,
      }
    } catch {
      // keep defaults
    }
  }

  showForm.value = true
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

async function toggleActive(p: Practice) {
  await api.updatePractice(p.id, { ...p, active: !p.active })
  await load()
}

async function deletePractice(p: Practice) {
  if (!confirm(`Delete "${p.name}" and all its logs?`)) return
  await api.deletePractice(p.id)
  await load()
}

function goToPractice(p: Practice) {
  if (p.type === 'memorize') {
    router.push(`/memorize?id=${p.id}`)
  } else {
    router.push(`/practices/${p.id}/history`)
  }
}

// Scripture lookup
const lookingUp = ref(false)
const lookupError = ref('')

async function lookupScripture() {
  if (!form.value.name.trim()) return
  lookingUp.value = true
  lookupError.value = ''
  try {
    const result = await api.lookupScripture(form.value.name.trim())
    if (result.verses && result.verses.length > 0) {
      form.value.description = result.verses.map(v => v.text).join(' ')
      // Normalize the name to the canonical reference
      if (result.verses.length === 1) {
        form.value.name = result.verses[0]!.reference
      }
    }
  } catch (e: any) {
    lookupError.value = e.message || 'Not found'
  }
  lookingUp.value = false
}

onMounted(load)
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold">Practices</h1>
      <button
        @click="showForm ? resetForm() : (showForm = true)"
        class="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 text-sm"
      >
        {{ showForm ? 'Cancel' : '+ Add Practice' }}
      </button>
    </div>

    <!-- Add/Edit form -->
    <div v-if="showForm" class="bg-white rounded-lg shadow p-4 mb-6">
      <h2 class="text-lg font-semibold mb-3">{{ editingId ? 'Edit Practice' : 'New Practice' }}</h2>
      <form @submit.prevent="submit" class="space-y-4">
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Name</label>
            <div class="flex gap-2">
              <input
                v-model="form.name"
                required
                class="flex-1 border rounded px-3 py-2 text-sm"
                placeholder="Clamshell, D&C 93:29, Morning prayer..."
              />
              <button
                v-if="form.type === 'memorize'"
                type="button"
                @click="lookupScripture"
                :disabled="lookingUp || !form.name.trim()"
                class="px-3 py-2 bg-indigo-100 text-indigo-700 rounded text-sm hover:bg-indigo-200 disabled:opacity-40 disabled:cursor-not-allowed whitespace-nowrap"
              >
                {{ lookingUp ? '...' : '📖 Lookup' }}
              </button>
            </div>
            <p v-if="lookupError" class="text-xs text-red-500 mt-1">{{ lookupError }}</p>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Type</label>
            <select v-model="form.type" class="w-full border rounded px-3 py-2 text-sm">
              <option value="habit">Habit (daily check)</option>
              <option value="tracker">Tracker (sets/reps)</option>
              <option value="memorize">Memorize (spaced repetition)</option>
              <option value="task">Task (one-time)</option>
            </select>
          </div>
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Category</label>
          <div class="flex gap-2 flex-wrap">
            <button
              v-for="cat in presetCategories"
              :key="cat"
              type="button"
              @click="form.category = cat"
              class="px-3 py-1 text-xs rounded-full border"
              :class="form.category === cat
                ? 'bg-indigo-100 border-indigo-300 text-indigo-700'
                : 'bg-gray-50 border-gray-200 text-gray-600 hover:bg-gray-100'"
            >
              {{ cat }}
            </button>
            <input
              v-model="form.category"
              class="px-3 py-1 text-xs border rounded-full w-28"
              placeholder="custom..."
            />
          </div>
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">
            {{ form.type === 'memorize' ? 'Verse / Quote Text' : 'Description' }}
          </label>
          <textarea
            v-model="form.description"
            :rows="form.type === 'memorize' ? 4 : 2"
            class="w-full border rounded px-3 py-2 text-sm"
            :placeholder="form.type === 'memorize'
              ? 'Man was also in the beginning with God. Intelligence, or the light of truth, was not created or made, neither indeed can be.'
              : 'Full verse text, exercise instructions, etc.'"
          ></textarea>
        </div>

        <!-- Memorize hint + daily reps -->
        <div v-if="form.type === 'memorize'" class="bg-indigo-50 rounded p-3 space-y-3">
          <p class="text-sm text-indigo-700">
            <strong>Tip:</strong> Put the scripture reference as the Name (flashcard front).
            Put the full verse text in the Description (flashcard back).
          </p>
          <div>
            <label class="block text-xs text-indigo-600 mb-1">Daily practice goal</label>
            <div class="flex items-center gap-2">
              <input
                v-model.number="memorizeConfig.target_daily_reps"
                type="number"
                min="1"
                max="20"
                class="w-20 border rounded px-2 py-1 text-sm"
              />
              <span class="text-sm text-indigo-600">reviews per day</span>
            </div>
          </div>
        </div>

        <!-- Tracker config (was exercise) -->
        <div v-if="form.type === 'tracker'" class="bg-gray-50 rounded p-3 space-y-3">
          <h3 class="text-sm font-medium text-gray-700">Tracker Settings</h3>
          <div class="grid grid-cols-3 gap-3">
            <div>
              <label class="block text-xs text-gray-500">Target Sets</label>
              <input
                v-model.number="trackerConfig.target_sets"
                type="number"
                min="1"
                class="w-full border rounded px-2 py-1 text-sm"
              />
            </div>
            <div>
              <label class="block text-xs text-gray-500">Target Reps</label>
              <input
                v-model.number="trackerConfig.target_reps"
                type="number"
                min="1"
                class="w-full border rounded px-2 py-1 text-sm"
              />
            </div>
            <div>
              <label class="block text-xs text-gray-500">Unit</label>
              <select v-model="trackerConfig.unit" class="w-full border rounded px-2 py-1 text-sm">
                <option value="reps">reps</option>
                <option value="bottles">bottles</option>
                <option value="glasses">glasses</option>
                <option value="seconds">seconds</option>
                <option value="minutes">minutes</option>
              </select>
            </div>
          </div>
        </div>

        <div class="flex gap-2">
          <button
            type="submit"
            class="px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 text-sm"
          >
            {{ editingId ? 'Save Changes' : 'Add Practice' }}
          </button>
          <button
            v-if="editingId"
            type="button"
            @click="resetForm"
            class="px-4 py-2 bg-gray-200 text-gray-700 rounded-lg hover:bg-gray-300 text-sm"
          >
            Cancel
          </button>
        </div>
      </form>
    </div>

    <!-- Practice list -->
    <div v-if="loading" class="text-center py-8 text-gray-400">Loading...</div>

    <template v-else-if="practices.length > 0">
      <!-- Filters -->
      <div class="mb-4 space-y-2">
        <div class="flex gap-1.5 flex-wrap items-center">
          <span class="text-xs text-gray-400 w-10">Type</span>
          <button
            @click="filterType = 'all'"
            class="px-2.5 py-1 text-xs rounded-full border"
            :class="filterType === 'all' ? 'bg-indigo-100 border-indigo-300 text-indigo-700' : 'bg-gray-50 border-gray-200 text-gray-500 hover:bg-gray-100'"
          >all</button>
          <button
            v-for="t in availableTypes"
            :key="t"
            @click="filterType = t"
            class="px-2.5 py-1 text-xs rounded-full border"
            :class="filterType === t ? 'bg-indigo-100 border-indigo-300 text-indigo-700' : 'bg-gray-50 border-gray-200 text-gray-500 hover:bg-gray-100'"
          >{{ t }}</button>
        </div>
        <div v-if="availableCategories.length > 1" class="flex gap-1.5 flex-wrap items-center">
          <span class="text-xs text-gray-400 w-10">Cat</span>
          <button
            @click="filterCategory = 'all'"
            class="px-2.5 py-1 text-xs rounded-full border"
            :class="filterCategory === 'all' ? 'bg-indigo-100 border-indigo-300 text-indigo-700' : 'bg-gray-50 border-gray-200 text-gray-500 hover:bg-gray-100'"
          >all</button>
          <button
            v-for="c in availableCategories"
            :key="c"
            @click="filterCategory = c"
            class="px-2.5 py-1 text-xs rounded-full border"
            :class="filterCategory === c ? 'bg-indigo-100 border-indigo-300 text-indigo-700' : 'bg-gray-50 border-gray-200 text-gray-500 hover:bg-gray-100'"
          >{{ c }}</button>
        </div>
      </div>

      <div v-if="filteredPractices.length === 0" class="text-center py-8 text-gray-400">
        No practices match filters.
      </div>

      <div v-else class="bg-white rounded-lg shadow divide-y divide-gray-100">
        <div
          v-for="p in filteredPractices"
          :key="p.id"
        class="flex items-center justify-between px-4 py-3"
        :class="{ 'opacity-50': !p.active }"
      >
        <div class="min-w-0 flex-1 cursor-pointer" @click="goToPractice(p)">
          <div class="flex items-center gap-2">
            <span class="font-medium hover:text-indigo-600 transition-colors">{{ p.name }}</span>
            <span class="text-xs px-2 py-0.5 rounded-full bg-gray-100 text-gray-600">{{ p.type }}</span>
            <span v-if="p.category" class="text-xs px-2 py-0.5 rounded-full bg-indigo-50 text-indigo-600">{{ p.category }}</span>
          </div>
          <div v-if="p.description" class="text-xs text-gray-400 truncate mt-0.5">{{ p.description }}</div>
        </div>

        <div class="flex items-center gap-2 ml-4">
          <router-link
            :to="`/practices/${p.id}/history`"
            class="text-xs text-gray-400 hover:text-indigo-500"
          >
            history
          </router-link>
          <button
            @click.stop="editPractice(p)"
            class="text-xs text-indigo-500 hover:text-indigo-700 px-2 py-1"
          >
            edit
          </button>
          <button
            @click="toggleActive(p)"
            class="text-xs px-2 py-1 rounded"
            :class="p.active ? 'text-yellow-600 hover:bg-yellow-50' : 'text-green-600 hover:bg-green-50'"
          >
            {{ p.active ? 'pause' : 'resume' }}
          </button>
          <button
            @click="deletePractice(p)"
            class="text-xs text-red-400 hover:text-red-600 px-2 py-1"
          >
            delete
          </button>
        </div>
      </div>
    </div>
    </template>

    <div v-else class="text-center py-12 text-gray-500">
      No practices yet. Add one above!
    </div>
  </div>
</template>
