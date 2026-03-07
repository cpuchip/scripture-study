<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { api, type Task, type BrainEntry } from '../api'
import { useAuth } from '../composables/useAuth'

const { user } = useAuth()
const tasks = ref<Task[]>([])
const brainEntries = ref<BrainEntry[]>([])
const agentOnline = ref(false)
const loading = ref(true)
const showForm = ref(false)
const activeTab = ref<'tasks' | 'brain'>('brain')
const brainFilter = ref('')
const toast = ref('')

const form = ref({
  title: '',
  description: '',
  source_doc: '',
  scripture: '',
  type: 'ongoing',
})

const hasBrain = computed(() => user.value?.brain_enabled)

// Group brain entries by category
const brainCategories = computed(() => {
  const groups: Record<string, BrainEntry[]> = {}
  for (const e of brainEntries.value) {
    if (brainFilter.value && e.category !== brainFilter.value) continue
    if (!groups[e.category]) groups[e.category] = []
    groups[e.category].push(e)
  }
  // Sort categories in a sensible order
  const order = ['actions', 'projects', 'ideas', 'people', 'study', 'journal', 'inbox']
  const sorted: [string, BrainEntry[]][] = []
  for (const cat of order) {
    if (groups[cat]) sorted.push([cat, groups[cat]])
  }
  // Any remaining categories not in the order
  for (const cat of Object.keys(groups)) {
    if (!order.includes(cat)) sorted.push([cat, groups[cat]])
  }
  return sorted
})

const categoryColors: Record<string, string> = {
  actions: 'bg-green-100 text-green-800',
  projects: 'bg-blue-100 text-blue-800',
  ideas: 'bg-purple-100 text-purple-800',
  people: 'bg-teal-100 text-teal-800',
  study: 'bg-amber-100 text-amber-800',
  journal: 'bg-pink-100 text-pink-800',
  inbox: 'bg-gray-100 text-gray-700',
}

function showToast(msg: string) {
  toast.value = msg
  setTimeout(() => (toast.value = ''), 2000)
}

async function load() {
  loading.value = true
  const promises: Promise<void>[] = [
    api.listTasks().then(t => { tasks.value = t }),
  ]
  if (hasBrain.value) {
    promises.push(
      api.listBrainEntries().then(r => {
        brainEntries.value = r.entries
        agentOnline.value = r.agent_online
      })
    )
  }
  await Promise.all(promises)
  loading.value = false
}

async function submit() {
  await api.createTask({
    title: form.value.title,
    description: form.value.description,
    source_doc: form.value.source_doc,
    scripture: form.value.scripture,
    type: form.value.type,
    status: 'active',
  })
  form.value = { title: '', description: '', source_doc: '', scripture: '', type: 'ongoing' }
  showForm.value = false
  await load()
}

async function toggleStatus(task: Task) {
  const newStatus = task.status === 'active' ? 'completed' : 'active'
  await api.updateTask(task.id, { ...task, status: newStatus })
  await load()
}

async function deleteTask(task: Task) {
  if (!confirm(`Delete "${task.title}"?`)) return
  await api.deleteTask(task.id)
  await load()
}

function isActionable(entry: BrainEntry) {
  return entry.category === 'actions' || entry.category === 'projects'
}

function isDone(entry: BrainEntry) {
  return entry.action_done || entry.status === 'done'
}

async function toggleBrainDone(entry: BrainEntry) {
  const newDone = !isDone(entry)
  const updates: Record<string, unknown> = { action_done: newDone }
  if (newDone) {
    updates.status = 'done'
  } else {
    updates.status = ''
  }
  await api.updateBrainEntry(entry.id, updates)
  entry.action_done = newDone
  entry.status = newDone ? 'done' : ''
  showToast(newDone ? 'Marked done' : 'Marked active')
}

onMounted(load)
</script>

<template>
  <div>
    <!-- Toast -->
    <Transition enter-active-class="transition-opacity duration-200" leave-active-class="transition-opacity duration-150"
      enter-from-class="opacity-0" leave-to-class="opacity-0">
      <div v-if="toast" class="fixed top-4 right-4 z-50 bg-gray-800 text-white px-4 py-2 rounded-lg shadow-lg text-sm">
        {{ toast }}
      </div>
    </Transition>

    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold">
        {{ hasBrain ? 'Brain & Tasks' : 'Tasks & Commitments' }}
      </h1>
      <div class="flex items-center gap-2">
        <span v-if="hasBrain" class="text-xs px-2 py-1 rounded-full" :class="agentOnline ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-500'">
          {{ agentOnline ? '● Online' : '○ Offline' }}
        </span>
        <button
          @click="showForm = !showForm"
          class="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 text-sm"
        >
          {{ showForm ? 'Cancel' : '+ Add Task' }}
        </button>
      </div>
    </div>

    <!-- Tab bar (only when brain is enabled) -->
    <div v-if="hasBrain" class="flex gap-1 mb-4 bg-gray-100 rounded-lg p-1">
      <button
        @click="activeTab = 'brain'"
        class="flex-1 px-3 py-2 text-sm font-medium rounded-md transition-colors"
        :class="activeTab === 'brain' ? 'bg-white shadow text-gray-900' : 'text-gray-500 hover:text-gray-700'"
      >
        🧠 Brain ({{ brainEntries.length }})
      </button>
      <button
        @click="activeTab = 'tasks'"
        class="flex-1 px-3 py-2 text-sm font-medium rounded-md transition-colors"
        :class="activeTab === 'tasks' ? 'bg-white shadow text-gray-900' : 'text-gray-500 hover:text-gray-700'"
      >
        Tasks ({{ tasks.length }})
      </button>
    </div>

    <!-- Add form -->
    <div v-if="showForm" class="bg-white dark:bg-gray-800 rounded-lg shadow p-4 mb-6">
      <form @submit.prevent="submit" class="space-y-3">
        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Title</label>
          <input v-model="form.title" required class="w-full border rounded px-3 py-2 text-sm dark:bg-gray-700 dark:border-gray-600" placeholder="Partake of sacrament with broken heart" />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Description</label>
          <textarea v-model="form.description" rows="2" class="w-full border rounded px-3 py-2 text-sm dark:bg-gray-700 dark:border-gray-600"></textarea>
        </div>
        <div class="grid grid-cols-3 gap-3">
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Source Doc</label>
            <input v-model="form.source_doc" class="w-full border rounded px-3 py-2 text-sm dark:bg-gray-700 dark:border-gray-600" placeholder="study/truth.md" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Scripture</label>
            <input v-model="form.scripture" class="w-full border rounded px-3 py-2 text-sm dark:bg-gray-700 dark:border-gray-600" placeholder="D&C 93:29" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Type</label>
            <select v-model="form.type" class="w-full border rounded px-3 py-2 text-sm dark:bg-gray-700 dark:border-gray-600">
              <option value="once">Once</option>
              <option value="daily">Daily</option>
              <option value="weekly">Weekly</option>
              <option value="ongoing">Ongoing</option>
            </select>
          </div>
        </div>
        <button type="submit" class="px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 text-sm">
          Add Task
        </button>
      </form>
    </div>

    <div v-if="loading" class="text-center py-8 text-gray-400">Loading...</div>

    <!-- Brain entries view -->
    <template v-else-if="hasBrain && activeTab === 'brain'">
      <!-- Category filter chips -->
      <div v-if="brainEntries.length > 0" class="flex flex-wrap gap-2 mb-4">
        <button
          @click="brainFilter = ''"
          class="px-3 py-1 text-xs rounded-full transition-colors"
          :class="brainFilter === '' ? 'bg-gray-800 text-white' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'"
        >
          All
        </button>
        <button
          v-for="[cat, entries] in brainCategories"
          :key="cat"
          @click="brainFilter = brainFilter === cat ? '' : cat"
          class="px-3 py-1 text-xs rounded-full transition-colors"
          :class="brainFilter === cat ? 'bg-gray-800 text-white' : (categoryColors[cat] || 'bg-gray-100 text-gray-600')"
        >
          {{ cat }} ({{ entries.length }})
        </button>
      </div>

      <div v-if="brainEntries.length === 0" class="text-center py-12 text-gray-500">
        <p class="mb-2">No brain entries yet.</p>
        <p class="text-xs text-gray-400">Entries sync from brain.exe when the agent connects.</p>
      </div>

      <!-- Grouped entries -->
      <div v-else class="space-y-6">
        <div v-for="[cat, entries] in brainCategories" :key="cat">
          <h2 class="text-sm font-semibold text-gray-500 uppercase tracking-wider mb-2 flex items-center gap-2">
            <span class="inline-block w-2 h-2 rounded-full" :class="{
              'bg-green-500': cat === 'actions',
              'bg-blue-500': cat === 'projects',
              'bg-purple-500': cat === 'ideas',
              'bg-teal-500': cat === 'people',
              'bg-amber-500': cat === 'study',
              'bg-pink-500': cat === 'journal',
              'bg-gray-400': cat === 'inbox',
            }"></span>
            {{ cat }}
            <span class="text-gray-300 font-normal">({{ entries.length }})</span>
          </h2>
          <div class="bg-white dark:bg-gray-800 rounded-lg shadow divide-y divide-gray-100 dark:divide-gray-700">
            <div
              v-for="entry in entries"
              :key="entry.id"
              class="flex items-center justify-between px-4 py-3"
              :class="{ 'opacity-50': isDone(entry) }"
            >
              <div class="flex items-center gap-3 flex-1 min-w-0">
                <!-- Done toggle for actionable entries -->
                <button
                  v-if="isActionable(entry)"
                  @click="toggleBrainDone(entry)"
                  class="w-6 h-6 rounded-full border-2 flex items-center justify-center flex-shrink-0"
                  :class="isDone(entry)
                    ? 'bg-green-500 border-green-500 text-white'
                    : 'border-gray-300 hover:border-indigo-400'"
                >
                  <span v-if="isDone(entry)" class="text-xs">✓</span>
                </button>
                <!-- Category dot for non-actionable -->
                <span v-else class="w-2 h-2 rounded-full flex-shrink-0" :class="{
                  'bg-purple-400': cat === 'ideas',
                  'bg-teal-400': cat === 'people',
                  'bg-amber-400': cat === 'study',
                  'bg-pink-400': cat === 'journal',
                  'bg-gray-300': cat === 'inbox',
                }"></span>

                <div class="min-w-0">
                  <div class="font-medium" :class="{ 'line-through': isDone(entry) }">
                    {{ entry.title }}
                  </div>
                  <div class="text-xs text-gray-400 flex flex-wrap gap-2">
                    <span v-if="entry.status && entry.status !== 'done'" class="text-gray-500">{{ entry.status }}</span>
                    <span v-if="entry.due_date">📅 {{ entry.due_date }}</span>
                    <span v-if="entry.next_action" class="text-blue-500">→ {{ entry.next_action }}</span>
                    <span v-if="entry.tags?.length" class="text-gray-300">{{ entry.tags.join(', ') }}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </template>

    <!-- Native tasks view -->
    <template v-else>
      <div v-if="tasks.length === 0" class="text-center py-12 text-gray-500">
        No tasks yet.
      </div>

      <div v-else class="bg-white dark:bg-gray-800 rounded-lg shadow divide-y divide-gray-100 dark:divide-gray-700">
        <div
          v-for="task in tasks"
          :key="task.id"
          class="flex items-center justify-between px-4 py-3"
          :class="{ 'opacity-50': task.status === 'completed' }"
        >
          <div class="flex items-center gap-3 flex-1 min-w-0">
            <button
              @click="toggleStatus(task)"
              class="w-6 h-6 rounded-full border-2 flex items-center justify-center flex-shrink-0"
              :class="task.status === 'completed'
                ? 'bg-green-500 border-green-500 text-white'
                : 'border-gray-300 hover:border-indigo-400'"
            >
              <span v-if="task.status === 'completed'" class="text-xs">✓</span>
            </button>
            <div class="min-w-0">
              <div class="font-medium">
                {{ task.title }}
                <span v-if="user?.brain_enabled && task.brain_entry_id" class="text-xs ml-1" title="Linked to brain entry">🧠</span>
              </div>
              <div class="text-xs text-gray-400 flex gap-2">
                <span v-if="task.scripture">📖 {{ task.scripture }}</span>
                <span v-if="task.source_doc">from {{ task.source_doc }}</span>
                <span class="text-gray-300">{{ task.type }}</span>
              </div>
            </div>
          </div>
          <button @click="deleteTask(task)" class="text-xs text-red-400 hover:text-red-600 ml-2">delete</button>
        </div>
      </div>
    </template>
  </div>
</template>
