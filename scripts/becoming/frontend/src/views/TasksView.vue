<script setup lang="ts">
import { ref, computed, onMounted, nextTick } from 'vue'
import { api, type Task, type BrainEntry } from '../api'
import { useAuth } from '../composables/useAuth'

const { user } = useAuth()
const tasks = ref<Task[]>([])
const brainEntries = ref<BrainEntry[]>([])
const agentOnline = ref(false)
const loading = ref(true)
const activeTab = ref<'tasks' | 'brain'>('brain')
const brainFilter = ref('')
const toast = ref('')

// Task form
const showTaskForm = ref(false)
const taskForm = ref({
  title: '',
  description: '',
  source_doc: '',
  scripture: '',
  type: 'ongoing',
})

// Brain create form
const showBrainForm = ref(false)
const brainForm = ref({
  title: '',
  category: 'inbox',
  body: '',
  due_date: '',
  tags: '',
})

// Edit dialog
const editEntry = ref<BrainEntry | null>(null)
const editDialogRef = ref<HTMLDialogElement | null>(null)
const editForm = ref({
  title: '',
  category: '',
  body: '',
  status: '',
  due_date: '',
  next_action: '',
  tags: '',
})

// Delete confirm dialog
const deleteTarget = ref<BrainEntry | null>(null)
const deleteDialogRef = ref<HTMLDialogElement | null>(null)

const hasBrain = computed(() => user.value?.brain_enabled)

const brainCategories = computed(() => {
  const groups: Record<string, BrainEntry[]> = {}
  for (const e of brainEntries.value) {
    if (brainFilter.value && e.category !== brainFilter.value) continue
    if (!groups[e.category]) groups[e.category] = []
    groups[e.category].push(e)
  }
  const order = ['actions', 'projects', 'ideas', 'people', 'study', 'journal', 'inbox']
  const sorted: [string, BrainEntry[]][] = []
  for (const cat of order) {
    if (groups[cat]) sorted.push([cat, groups[cat]])
  }
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

const allCategories = ['inbox', 'actions', 'projects', 'ideas', 'people', 'study', 'journal']

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

// Task CRUD
async function submitTask() {
  await api.createTask({
    title: taskForm.value.title,
    description: taskForm.value.description,
    source_doc: taskForm.value.source_doc,
    scripture: taskForm.value.scripture,
    type: taskForm.value.type,
    status: 'active',
  })
  taskForm.value = { title: '', description: '', source_doc: '', scripture: '', type: 'ongoing' }
  showTaskForm.value = false
  await load()
}

async function toggleStatus(task: Task) {
  const newStatus = task.status === 'active' ? 'completed' : 'active'
  await api.updateTask(task.id, { ...task, status: newStatus })
  await load()
}

async function deleteTask(task: Task) {
  await api.deleteTask(task.id)
  await load()
}

// Brain entry CRUD
async function submitBrainEntry() {
  const tags = brainForm.value.tags
    .split(',')
    .map(t => t.trim())
    .filter(Boolean)
  await api.createBrainEntry({
    title: brainForm.value.title,
    category: brainForm.value.category,
    body: brainForm.value.body,
    due_date: brainForm.value.due_date,
    tags,
  })
  brainForm.value = { title: '', category: 'inbox', body: '', due_date: '', tags: '' }
  showBrainForm.value = false
  showToast('Entry created')
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

function openEditDialog(entry: BrainEntry) {
  editEntry.value = entry
  editForm.value = {
    title: entry.title,
    category: entry.category,
    body: entry.body || '',
    status: entry.status || '',
    due_date: entry.due_date || '',
    next_action: entry.next_action || '',
    tags: entry.tags?.join(', ') || '',
  }
  nextTick(() => editDialogRef.value?.showModal())
}

function closeEditDialog() {
  editDialogRef.value?.close()
  editEntry.value = null
}

async function saveEdit() {
  if (!editEntry.value) return
  const tags = editForm.value.tags
    .split(',')
    .map(t => t.trim())
    .filter(Boolean)
  const updates: Record<string, unknown> = {
    title: editForm.value.title,
    category: editForm.value.category,
    body: editForm.value.body,
    status: editForm.value.status,
    due_date: editForm.value.due_date,
    next_action: editForm.value.next_action,
    tags,
  }
  await api.updateBrainEntry(editEntry.value.id, updates)
  showToast('Entry updated')
  closeEditDialog()
  await load()
}

function confirmDelete(entry: BrainEntry) {
  deleteTarget.value = entry
  nextTick(() => deleteDialogRef.value?.showModal())
}

function closeDeleteDialog() {
  deleteDialogRef.value?.close()
  deleteTarget.value = null
}

async function executeDelete() {
  if (!deleteTarget.value) return
  await api.deleteBrainEntry(deleteTarget.value.id)
  showToast('Entry deleted')
  closeDeleteDialog()
  await load()
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

    <!-- Edit dialog -->
    <Teleport to="body">
      <dialog ref="editDialogRef" @close="editEntry = null" @cancel.prevent="closeEditDialog"
        class="rounded-xl border border-gray-200 bg-white p-6 shadow-xl backdrop:bg-black/50 dark:border-gray-700 dark:bg-gray-800 w-full max-w-lg">
        <form @submit.prevent="saveEdit" class="space-y-4">
          <h2 class="text-lg font-semibold">Edit Entry</h2>
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Title</label>
            <input v-model="editForm.title" required class="w-full border rounded px-3 py-2 text-sm dark:bg-gray-700 dark:border-gray-600" />
          </div>
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Category</label>
              <select v-model="editForm.category" class="w-full border rounded px-3 py-2 text-sm dark:bg-gray-700 dark:border-gray-600">
                <option v-for="cat in allCategories" :key="cat" :value="cat">{{ cat }}</option>
              </select>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Status</label>
              <input v-model="editForm.status" class="w-full border rounded px-3 py-2 text-sm dark:bg-gray-700 dark:border-gray-600" placeholder="active, done, waiting..." />
            </div>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Body</label>
            <textarea v-model="editForm.body" rows="4" class="w-full border rounded px-3 py-2 text-sm dark:bg-gray-700 dark:border-gray-600"></textarea>
          </div>
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Due Date</label>
              <input v-model="editForm.due_date" type="date" class="w-full border rounded px-3 py-2 text-sm dark:bg-gray-700 dark:border-gray-600" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Next Action</label>
              <input v-model="editForm.next_action" class="w-full border rounded px-3 py-2 text-sm dark:bg-gray-700 dark:border-gray-600" />
            </div>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Tags (comma-separated)</label>
            <input v-model="editForm.tags" class="w-full border rounded px-3 py-2 text-sm dark:bg-gray-700 dark:border-gray-600" placeholder="tag1, tag2" />
          </div>
          <div class="flex justify-end gap-2 pt-2">
            <button type="button" @click="closeEditDialog" class="px-4 py-2 text-sm text-gray-600 hover:text-gray-800">Cancel</button>
            <button type="submit" class="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 text-sm">Save</button>
          </div>
        </form>
      </dialog>
    </Teleport>

    <!-- Delete confirm dialog -->
    <Teleport to="body">
      <dialog ref="deleteDialogRef" @close="deleteTarget = null" @cancel.prevent="closeDeleteDialog"
        class="rounded-xl border border-gray-200 bg-white p-6 shadow-xl backdrop:bg-black/50 dark:border-gray-700 dark:bg-gray-800 w-full max-w-sm">
        <h2 class="text-lg font-semibold mb-2">Delete Entry</h2>
        <p class="text-sm text-gray-600 dark:text-gray-400 mb-4">
          Delete "<span class="font-medium">{{ deleteTarget?.title }}</span>"? This will remove it from both the cache and brain.exe.
        </p>
        <div class="flex justify-end gap-2">
          <button @click="closeDeleteDialog" class="px-4 py-2 text-sm text-gray-600 hover:text-gray-800">Cancel</button>
          <button @click="executeDelete" class="px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 text-sm">Delete</button>
        </div>
      </dialog>
    </Teleport>

    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold">
        {{ hasBrain ? 'Brain & Tasks' : 'Tasks & Commitments' }}
      </h1>
      <div class="flex items-center gap-2">
        <span v-if="hasBrain" class="text-xs px-2 py-1 rounded-full" :class="agentOnline ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-500'">
          {{ agentOnline ? '● Online' : '○ Offline' }}
        </span>
        <button
          v-if="hasBrain && activeTab === 'brain'"
          @click="showBrainForm = !showBrainForm"
          class="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 text-sm"
        >
          {{ showBrainForm ? 'Cancel' : '+ New Entry' }}
        </button>
        <button
          v-if="!hasBrain || activeTab === 'tasks'"
          @click="showTaskForm = !showTaskForm"
          class="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 text-sm"
        >
          {{ showTaskForm ? 'Cancel' : '+ Add Task' }}
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

    <!-- Brain create form -->
    <Transition enter-active-class="transition-all duration-200" leave-active-class="transition-all duration-150"
      enter-from-class="opacity-0 -translate-y-2" leave-to-class="opacity-0 -translate-y-2">
      <div v-if="showBrainForm" class="bg-white dark:bg-gray-800 rounded-lg shadow p-4 mb-6">
        <form @submit.prevent="submitBrainEntry" class="space-y-3">
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Title</label>
            <input v-model="brainForm.title" required class="w-full border rounded px-3 py-2 text-sm dark:bg-gray-700 dark:border-gray-600" placeholder="What's on your mind?" />
          </div>
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Category</label>
              <select v-model="brainForm.category" class="w-full border rounded px-3 py-2 text-sm dark:bg-gray-700 dark:border-gray-600">
                <option v-for="cat in allCategories" :key="cat" :value="cat">{{ cat }}</option>
              </select>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Due Date</label>
              <input v-model="brainForm.due_date" type="date" class="w-full border rounded px-3 py-2 text-sm dark:bg-gray-700 dark:border-gray-600" />
            </div>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Body</label>
            <textarea v-model="brainForm.body" rows="2" class="w-full border rounded px-3 py-2 text-sm dark:bg-gray-700 dark:border-gray-600" placeholder="Details, notes, context..."></textarea>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Tags (comma-separated)</label>
            <input v-model="brainForm.tags" class="w-full border rounded px-3 py-2 text-sm dark:bg-gray-700 dark:border-gray-600" placeholder="tag1, tag2" />
          </div>
          <button type="submit" class="px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 text-sm">
            Create Entry
          </button>
        </form>
      </div>
    </Transition>

    <!-- Task form -->
    <Transition enter-active-class="transition-all duration-200" leave-active-class="transition-all duration-150"
      enter-from-class="opacity-0 -translate-y-2" leave-to-class="opacity-0 -translate-y-2">
      <div v-if="showTaskForm && (!hasBrain || activeTab === 'tasks')" class="bg-white dark:bg-gray-800 rounded-lg shadow p-4 mb-6">
        <form @submit.prevent="submitTask" class="space-y-3">
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Title</label>
            <input v-model="taskForm.title" required class="w-full border rounded px-3 py-2 text-sm dark:bg-gray-700 dark:border-gray-600" placeholder="Partake of sacrament with broken heart" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Description</label>
            <textarea v-model="taskForm.description" rows="2" class="w-full border rounded px-3 py-2 text-sm dark:bg-gray-700 dark:border-gray-600"></textarea>
          </div>
          <div class="grid grid-cols-3 gap-3">
            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Source Doc</label>
              <input v-model="taskForm.source_doc" class="w-full border rounded px-3 py-2 text-sm dark:bg-gray-700 dark:border-gray-600" placeholder="study/truth.md" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Scripture</label>
              <input v-model="taskForm.scripture" class="w-full border rounded px-3 py-2 text-sm dark:bg-gray-700 dark:border-gray-600" placeholder="D&C 93:29" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Type</label>
              <select v-model="taskForm.type" class="w-full border rounded px-3 py-2 text-sm dark:bg-gray-700 dark:border-gray-600">
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
    </Transition>

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
        <p class="text-xs text-gray-400">Create one above, or entries sync from brain.exe when the agent connects.</p>
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
              class="flex items-center justify-between px-4 py-3 group"
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

                <div class="min-w-0 cursor-pointer" @click="openEditDialog(entry)">
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
              <div class="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                <button @click="openEditDialog(entry)" class="p-1 text-gray-400 hover:text-indigo-600" aria-label="Edit entry">
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" /></svg>
                </button>
                <button @click="confirmDelete(entry)" class="p-1 text-gray-400 hover:text-red-600" aria-label="Delete entry">
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
                </button>
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
