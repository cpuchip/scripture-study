<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { api, type Practice } from '../api'

const router = useRouter()
const practices = ref<Practice[]>([])
const loading = ref(true)
const showForm = ref(false)

// Form state
const form = ref({
  name: '',
  description: '',
  type: 'habit' as Practice['type'],
  category: '',
  config: '{}',
})

// Config helpers for exercises
const exerciseConfig = ref({
  target_sets: 2,
  target_reps: 15,
  unit: 'reps',
})

const presetCategories = ['spiritual', 'scripture', 'pt', 'fitness', 'study', 'health']

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

  if (form.value.type === 'exercise') {
    p.config = JSON.stringify(exerciseConfig.value)
  } else {
    p.config = '{}'
  }

  await api.createPractice(p)
  resetForm()
  await load()
}

function resetForm() {
  form.value = { name: '', description: '', type: 'habit', category: '', config: '{}' }
  exerciseConfig.value = { target_sets: 2, target_reps: 15, unit: 'reps' }
  showForm.value = false
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
        @click="showForm = !showForm"
        class="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 text-sm"
      >
        {{ showForm ? 'Cancel' : '+ Add Practice' }}
      </button>
    </div>

    <!-- Add form -->
    <div v-if="showForm" class="bg-white rounded-lg shadow p-4 mb-6">
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
              <option value="exercise">Exercise (sets/reps)</option>
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

        <!-- Memorize hint -->
        <div v-if="form.type === 'memorize'" class="bg-indigo-50 rounded p-3 text-sm text-indigo-700">
          <strong>Tip:</strong> Put the scripture reference (e.g., "D&amp;C 93:29") as the Name — it becomes the flashcard front.
          Put the full verse text in the Description — it becomes the flashcard back.
          SM-2 scheduling will be set up automatically.
        </div>

        <!-- Exercise config -->
        <div v-if="form.type === 'exercise'" class="bg-gray-50 rounded p-3 space-y-3">
          <h3 class="text-sm font-medium text-gray-700">Exercise Settings</h3>
          <div class="grid grid-cols-3 gap-3">
            <div>
              <label class="block text-xs text-gray-500">Target Sets</label>
              <input
                v-model.number="exerciseConfig.target_sets"
                type="number"
                min="1"
                class="w-full border rounded px-2 py-1 text-sm"
              />
            </div>
            <div>
              <label class="block text-xs text-gray-500">Target Reps</label>
              <input
                v-model.number="exerciseConfig.target_reps"
                type="number"
                min="1"
                class="w-full border rounded px-2 py-1 text-sm"
              />
            </div>
            <div>
              <label class="block text-xs text-gray-500">Unit</label>
              <select v-model="exerciseConfig.unit" class="w-full border rounded px-2 py-1 text-sm">
                <option value="reps">reps</option>
                <option value="seconds">seconds</option>
                <option value="minutes">minutes</option>
              </select>
            </div>
          </div>
        </div>

        <button
          type="submit"
          class="px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 text-sm"
        >
          Add Practice
        </button>
      </form>
    </div>

    <!-- Practice list -->
    <div v-if="loading" class="text-center py-8 text-gray-400">Loading...</div>

    <div v-else-if="practices.length === 0" class="text-center py-12 text-gray-500">
      No practices yet. Add one above!
    </div>

    <div v-else class="bg-white rounded-lg shadow divide-y divide-gray-100">
      <div
        v-for="p in practices"
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
  </div>
</template>
