<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { api, type Bookmark } from '../api'

const router = useRouter()
const bookmarks = ref<Bookmark[]>([])
const loading = ref(true)
const searchQuery = ref('')
const editingId = ref<number | null>(null)
const editNote = ref('')

const filtered = computed(() => {
  if (!searchQuery.value) return bookmarks.value
  const q = searchQuery.value.toLowerCase()
  return bookmarks.value.filter(b =>
    b.file_path.toLowerCase().includes(q) ||
    b.excerpt.toLowerCase().includes(q) ||
    b.note.toLowerCase().includes(q) ||
    (b.source_name || '').toLowerCase().includes(q) ||
    b.anchor.toLowerCase().includes(q)
  )
})

// Group bookmarks by source
const grouped = computed(() => {
  const groups = new Map<string, Bookmark[]>()
  for (const b of filtered.value) {
    const key = b.source_name || `Source ${b.source_id}`
    if (!groups.has(key)) groups.set(key, [])
    groups.get(key)!.push(b)
  }
  return groups
})

async function load() {
  loading.value = true
  try {
    bookmarks.value = await api.listBookmarks()
  } catch (e) {
    console.error('Failed to load bookmarks:', e)
  } finally {
    loading.value = false
  }
}

function goToBookmark(b: Bookmark) {
  const query: Record<string, string> = { f: b.file_path }
  const hash = b.anchor ? `#${b.anchor}` : undefined
  router.push({ name: 'reader', params: { id: b.source_id }, query, hash })
}

function startEdit(b: Bookmark) {
  editingId.value = b.id
  editNote.value = b.note
}

function cancelEdit() {
  editingId.value = null
  editNote.value = ''
}

async function saveNote(b: Bookmark) {
  try {
    await api.updateBookmarkNote(b.id, editNote.value)
    b.note = editNote.value
    editingId.value = null
    editNote.value = ''
  } catch (e) {
    console.error('Failed to update bookmark note:', e)
  }
}

async function remove(id: number) {
  try {
    await api.deleteBookmark(id)
    bookmarks.value = bookmarks.value.filter(b => b.id !== id)
  } catch (e) {
    console.error('Failed to delete bookmark:', e)
  }
}

function formatPath(filePath: string): string {
  // Show just the filename without extension, or last two path segments
  const parts = filePath.replace(/\.md$/, '').split('/')
  return parts.length > 2 ? parts.slice(-2).join(' / ') : parts.join(' / ')
}

function formatDate(dateStr: string): string {
  const d = new Date(dateStr)
  return d.toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' })
}

onMounted(load)
</script>

<template>
  <div>
    <!-- Header -->
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold">Bookmarks</h1>
    </div>

    <!-- Search -->
    <input
      v-model="searchQuery"
      type="text"
      placeholder="Search bookmarks..."
      class="w-full border border-gray-200 rounded-lg px-3 py-2 text-sm mb-4 focus:outline-none focus:ring-2 focus:ring-orange-400"
    />

    <!-- Loading -->
    <div v-if="loading" class="text-center py-8 text-gray-400">Loading bookmarks...</div>

    <!-- Empty -->
    <div v-else-if="bookmarks.length === 0" class="text-center py-12">
      <div class="text-4xl mb-3">🔖</div>
      <p class="text-gray-500">No bookmarks yet</p>
      <p class="text-sm text-gray-400 mt-1">Bookmark sections while reading to save them here</p>
    </div>

    <!-- No results -->
    <div v-else-if="filtered.length === 0" class="text-center py-8 text-gray-400">
      No bookmarks match "{{ searchQuery }}"
    </div>

    <!-- Grouped list -->
    <div v-else class="space-y-6">
      <div v-for="[sourceName, items] in grouped" :key="sourceName">
        <h2 class="text-xs font-semibold uppercase tracking-wider text-gray-400 mb-2">{{ sourceName }}</h2>
        <div class="space-y-2">
          <div
            v-for="b in items"
            :key="b.id"
            class="bg-white rounded-lg shadow px-4 py-3 group hover:shadow-md transition-shadow"
          >
            <div class="flex items-start justify-between gap-3">
              <div class="flex-1 min-w-0 cursor-pointer" @click="goToBookmark(b)">
                <div class="flex items-center gap-2 mb-1">
                  <span class="text-sm font-medium text-gray-800 truncate">{{ formatPath(b.file_path) }}</span>
                  <span v-if="b.anchor" class="text-xs text-gray-400">#{{ b.anchor }}</span>
                </div>
                <p v-if="b.excerpt" class="text-sm text-gray-600 line-clamp-2">{{ b.excerpt }}</p>

                <!-- Note display -->
                <div v-if="b.note && editingId !== b.id" class="mt-1.5 text-xs text-gray-500 italic border-l-2 border-orange-300 pl-2">
                  {{ b.note }}
                </div>
              </div>

              <!-- Actions -->
              <div class="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity shrink-0">
                <span class="text-xs text-gray-400 mr-2">{{ formatDate(b.created_at) }}</span>
                <button @click.stop="startEdit(b)" class="text-gray-400 hover:text-orange-500 text-sm" title="Edit note">✏️</button>
                <button @click.stop="remove(b.id)" class="text-gray-400 hover:text-red-500 text-sm" title="Delete">🗑️</button>
              </div>
            </div>

            <!-- Inline edit -->
            <div v-if="editingId === b.id" class="mt-2 flex gap-2" @click.stop>
              <input
                v-model="editNote"
                type="text"
                placeholder="Add a note..."
                class="flex-1 border border-gray-200 rounded px-2 py-1 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400"
                @keydown.enter="saveNote(b)"
                @keydown.escape="cancelEdit"
              />
              <button @click="saveNote(b)" class="text-sm px-2 py-1 bg-orange-500 text-white rounded hover:bg-orange-600">Save</button>
              <button @click="cancelEdit" class="text-sm px-2 py-1 text-gray-500 hover:text-gray-700">Cancel</button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
