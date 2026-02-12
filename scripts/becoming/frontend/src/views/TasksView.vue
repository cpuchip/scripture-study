<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api, type Task } from '../api'

const tasks = ref<Task[]>([])
const loading = ref(true)
const showForm = ref(false)

const form = ref({
  title: '',
  description: '',
  source_doc: '',
  scripture: '',
  type: 'ongoing',
})

async function load() {
  loading.value = true
  tasks.value = await api.listTasks()
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

onMounted(load)
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold">Tasks & Commitments</h1>
      <button
        @click="showForm = !showForm"
        class="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 text-sm"
      >
        {{ showForm ? 'Cancel' : '+ Add Task' }}
      </button>
    </div>

    <!-- Add form -->
    <div v-if="showForm" class="bg-white rounded-lg shadow p-4 mb-6">
      <form @submit.prevent="submit" class="space-y-3">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Title</label>
          <input v-model="form.title" required class="w-full border rounded px-3 py-2 text-sm" placeholder="Partake of sacrament with broken heart" />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Description</label>
          <textarea v-model="form.description" rows="2" class="w-full border rounded px-3 py-2 text-sm"></textarea>
        </div>
        <div class="grid grid-cols-3 gap-3">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Source Doc</label>
            <input v-model="form.source_doc" class="w-full border rounded px-3 py-2 text-sm" placeholder="study/truth.md" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Scripture</label>
            <input v-model="form.scripture" class="w-full border rounded px-3 py-2 text-sm" placeholder="D&C 93:29" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Type</label>
            <select v-model="form.type" class="w-full border rounded px-3 py-2 text-sm">
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

    <!-- Task list -->
    <div v-if="loading" class="text-center py-8 text-gray-400">Loading...</div>

    <div v-else-if="tasks.length === 0" class="text-center py-12 text-gray-500">
      No tasks yet.
    </div>

    <div v-else class="bg-white rounded-lg shadow divide-y divide-gray-100">
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
            <div class="font-medium">{{ task.title }}</div>
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
  </div>
</template>
