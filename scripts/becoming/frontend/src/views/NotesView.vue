<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { api, type Note, type Practice, type Task } from '../api'

const notes = ref<Note[]>([])
const practices = ref<Practice[]>([])
const tasks = ref<Task[]>([])
const loading = ref(true)
const searchQuery = ref('')
const filterMode = ref<'all' | 'pinned' | 'practices' | 'tasks' | 'free'>('all')

// Create / edit state
const showForm = ref(false)
const editingNote = ref<Note | null>(null)
const formContent = ref('')
const formPracticeId = ref<number | null>(null)
const formTaskId = ref<number | null>(null)
const formPinned = ref(false)

const filtered = computed(() => {
  let result = notes.value

  // Filter by mode
  if (filterMode.value === 'pinned') {
    result = result.filter(n => n.pinned)
  } else if (filterMode.value === 'practices') {
    result = result.filter(n => n.practice_id)
  } else if (filterMode.value === 'tasks') {
    result = result.filter(n => n.task_id)
  } else if (filterMode.value === 'free') {
    result = result.filter(n => !n.practice_id && !n.task_id && !n.pillar_id)
  }

  // Search
  if (searchQuery.value.trim()) {
    const q = searchQuery.value.toLowerCase()
    result = result.filter(n =>
      n.content.toLowerCase().includes(q) ||
      (n.practice_name && n.practice_name.toLowerCase().includes(q)) ||
      (n.task_title && n.task_title.toLowerCase().includes(q))
    )
  }

  return result
})

function formatDate(dateStr: string): string {
  const d = new Date(dateStr)
  return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
}

function openNewNote() {
  editingNote.value = null
  formContent.value = ''
  formPracticeId.value = null
  formTaskId.value = null
  formPinned.value = false
  showForm.value = true
}

function openEditNote(note: Note) {
  editingNote.value = note
  formContent.value = note.content
  formPracticeId.value = note.practice_id ?? null
  formTaskId.value = note.task_id ?? null
  formPinned.value = note.pinned
  showForm.value = true
}

function cancelForm() {
  showForm.value = false
  editingNote.value = null
}

async function saveNote() {
  const payload: Partial<Note> = {
    content: formContent.value,
    practice_id: formPracticeId.value,
    task_id: formTaskId.value,
    pinned: formPinned.value,
  }

  try {
    if (editingNote.value) {
      await api.updateNote(editingNote.value.id, payload)
    } else {
      await api.createNote(payload)
    }
    showForm.value = false
    editingNote.value = null
    await load()
  } catch (e) {
    console.error('Failed to save note:', e)
  }
}

async function deleteNote(id: number) {
  if (!confirm('Delete this note?')) return
  try {
    await api.deleteNote(id)
    await load()
  } catch (e) {
    console.error('Failed to delete note:', e)
  }
}

async function togglePin(note: Note) {
  try {
    await api.updateNote(note.id, { ...note, pinned: !note.pinned })
    await load()
  } catch (e) {
    console.error('Failed to toggle pin:', e)
  }
}

async function load() {
  loading.value = true
  try {
    const [n, p, t] = await Promise.all([
      api.listNotes(),
      api.listPractices(),
      api.listTasks(),
    ])
    notes.value = n
    practices.value = p
    tasks.value = t
  } catch (e) {
    console.error('Failed to load:', e)
  }
  loading.value = false
}

onMounted(load)
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold">Notes</h1>
      <button
        @click="openNewNote"
        class="px-4 py-2 bg-indigo-600 text-white rounded-lg text-sm font-medium hover:bg-indigo-700 transition-colors"
      >+ New</button>
    </div>

    <!-- Search -->
    <div class="mb-4">
      <input
        v-model="searchQuery"
        type="text"
        placeholder="Search notes..."
        class="w-full border border-gray-200 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-300"
      />
    </div>

    <!-- Filter pills -->
    <div class="flex gap-1.5 flex-wrap mb-4">
      <button
        v-for="f in ([
          { value: 'all', label: 'All' },
          { value: 'pinned', label: '📌 Pinned' },
          { value: 'practices', label: 'Practices' },
          { value: 'tasks', label: 'Tasks' },
          { value: 'free', label: 'Free' },
        ] as const)"
        :key="f.value"
        @click="filterMode = f.value"
        class="px-2.5 py-1 text-xs rounded-full border transition-colors"
        :class="filterMode === f.value
          ? 'bg-indigo-100 border-indigo-300 text-indigo-700'
          : 'bg-gray-50 border-gray-200 text-gray-500 hover:bg-gray-100'"
      >{{ f.label }}</button>
    </div>

    <!-- Create / Edit form -->
    <div v-if="showForm" class="bg-white rounded-lg shadow p-4 mb-4 border border-indigo-200">
      <h3 class="text-sm font-semibold text-gray-700 mb-3">
        {{ editingNote ? 'Edit Note' : 'New Note' }}
      </h3>

      <textarea
        v-model="formContent"
        rows="3"
        placeholder="Write your note..."
        class="w-full border border-gray-200 rounded-lg px-3 py-2 text-sm mb-3 focus:outline-none focus:ring-2 focus:ring-indigo-300 resize-y"
        autofocus
      ></textarea>

      <div class="grid grid-cols-2 gap-3 mb-3">
        <div>
          <label class="text-xs text-gray-500">Link to practice</label>
          <select
            v-model="formPracticeId"
            class="w-full border border-gray-200 rounded px-2 py-1.5 text-sm"
            @change="formTaskId = null"
          >
            <option :value="null">None</option>
            <option v-for="p in practices" :key="p.id" :value="p.id">{{ p.name }}</option>
          </select>
        </div>
        <div>
          <label class="text-xs text-gray-500">Link to task</label>
          <select
            v-model="formTaskId"
            class="w-full border border-gray-200 rounded px-2 py-1.5 text-sm"
            @change="formPracticeId = null"
          >
            <option :value="null">None</option>
            <option v-for="t in tasks" :key="t.id" :value="t.id">{{ t.title }}</option>
          </select>
        </div>
      </div>

      <div class="flex items-center justify-between">
        <label class="flex items-center gap-2 text-sm text-gray-600 cursor-pointer">
          <input type="checkbox" v-model="formPinned" class="rounded" />
          📌 Pin to top
        </label>
        <div class="flex gap-2">
          <button
            @click="cancelForm"
            class="px-3 py-1.5 text-xs text-gray-500 hover:text-gray-700"
          >Cancel</button>
          <button
            @click="saveNote"
            :disabled="!formContent.trim()"
            class="px-4 py-1.5 bg-indigo-600 text-white rounded text-xs font-medium hover:bg-indigo-700 disabled:opacity-40 transition-colors"
          >{{ editingNote ? 'Save' : 'Create' }}</button>
        </div>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="text-center py-8 text-gray-400">Loading...</div>

    <!-- Empty state -->
    <div v-else-if="filtered.length === 0" class="text-center py-12">
      <div class="text-gray-400 mb-2">{{ searchQuery || filterMode !== 'all' ? 'No matching notes.' : 'No notes yet.' }}</div>
      <button
        v-if="!showForm"
        @click="openNewNote"
        class="text-sm text-indigo-600 hover:text-indigo-700"
      >Create your first note</button>
    </div>

    <!-- Notes list -->
    <div v-else class="space-y-2">
      <div
        v-for="note in filtered"
        :key="note.id"
        class="bg-white rounded-lg shadow px-4 py-3 group hover:shadow-md transition-shadow"
      >
        <div class="flex items-start justify-between gap-3">
          <div class="flex-1 min-w-0">
            <!-- Pin indicator -->
            <span v-if="note.pinned" class="text-xs mr-1">📌</span>

            <!-- Content -->
            <span class="text-sm text-gray-800 whitespace-pre-wrap">{{ note.content }}</span>
          </div>

          <!-- Actions (visible on hover) -->
          <div class="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity shrink-0">
            <button
              @click="togglePin(note)"
              class="p-1 text-gray-400 hover:text-amber-500 text-xs"
              :title="note.pinned ? 'Unpin' : 'Pin'"
            >📌</button>
            <button
              @click="openEditNote(note)"
              class="p-1 text-gray-400 hover:text-indigo-600 text-xs"
              title="Edit"
            >✏️</button>
            <button
              @click="deleteNote(note.id)"
              class="p-1 text-gray-400 hover:text-red-500 text-xs"
              title="Delete"
            >🗑️</button>
          </div>
        </div>

        <!-- Meta row -->
        <div class="mt-1.5 flex items-center gap-3 text-[11px] text-gray-400">
          <span v-if="note.practice_name" class="bg-indigo-50 text-indigo-600 px-1.5 py-0.5 rounded">
            {{ note.practice_name }}
          </span>
          <span v-if="note.task_title" class="bg-amber-50 text-amber-600 px-1.5 py-0.5 rounded">
            {{ note.task_title }}
          </span>
          <span>{{ formatDate(note.created_at) }}</span>
          <span v-if="note.updated_at !== note.created_at" class="italic">edited</span>
        </div>
      </div>
    </div>
  </div>
</template>
