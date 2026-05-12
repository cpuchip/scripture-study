<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { RouterLink } from 'vue-router'
import { api, type ProjectRow } from '@/api'

const projects = ref<ProjectRow[]>([])
const error = ref('')
const loading = ref(false)
const includeArchived = ref(false)

// Create form state
const showCreate = ref(false)
const createSlug = ref('')
const createName = ref('')
const createDesc = ref('')
const createRoot = ref('')
const createBusy = ref(false)
const createErr = ref('')

// Edit form state (one row at a time)
const editingSlug = ref<string | null>(null)
const editName = ref('')
const editDesc = ref('')
const editRoot = ref('')
const editBusy = ref(false)

async function load() {
  loading.value = true
  error.value = ''
  try {
    const r = await api.projectsList(includeArchived.value)
    projects.value = r.items
  } catch (e) {
    error.value = String(e)
  } finally {
    loading.value = false
  }
}

onMounted(load)

const slugValid = computed(() => /^[a-z0-9-]+$/.test(createSlug.value))

async function submitCreate() {
  if (!createSlug.value || !createName.value) return
  if (!slugValid.value) {
    createErr.value = 'slug must match ^[a-z0-9-]+$'
    return
  }
  createBusy.value = true
  createErr.value = ''
  try {
    await api.projectCreate({
      slug: createSlug.value,
      name: createName.value,
      description: createDesc.value || undefined,
      root_directory: createRoot.value || undefined,
    })
    showCreate.value = false
    createSlug.value = ''
    createName.value = ''
    createDesc.value = ''
    createRoot.value = ''
    await load()
  } catch (e) {
    createErr.value = String(e)
  } finally {
    createBusy.value = false
  }
}

function startEdit(p: ProjectRow) {
  editingSlug.value = p.slug
  editName.value = p.name
  editDesc.value = p.description ?? ''
  editRoot.value = p.root_directory ?? ''
}

function cancelEdit() {
  editingSlug.value = null
}

async function submitEdit() {
  if (!editingSlug.value) return
  editBusy.value = true
  try {
    await api.projectUpdate({
      slug: editingSlug.value,
      name: editName.value,
      description: editDesc.value,
      root_directory: editRoot.value,
    })
    editingSlug.value = null
    await load()
  } catch (e) {
    error.value = String(e)
  } finally {
    editBusy.value = false
  }
}

async function toggleArchive(p: ProjectRow) {
  const msg = p.archived
    ? `Unarchive "${p.slug}"?`
    : `Archive "${p.slug}"? It stays in the DB but is hidden by default.`
  if (!confirm(msg)) return
  try {
    await api.projectArchive(p.slug, !p.archived)
    await load()
  } catch (e) {
    error.value = String(e)
  }
}
</script>

<template>
  <div class="space-y-6">
    <header class="flex items-baseline justify-between border-b border-zinc-800 pb-4">
      <div>
        <h2 class="text-2xl font-semibold tracking-tight">Projects</h2>
        <p class="text-xs text-zinc-500 mt-1">
          Formalize <code class="font-mono">work_items.project_association</code> into
          named projects with descriptions + optional root directories.
        </p>
      </div>
      <div class="flex items-center gap-3 text-xs text-zinc-500">
        <label class="flex items-center gap-1 cursor-pointer">
          <input v-model="includeArchived" type="checkbox" @change="load" />
          show archived
        </label>
        <button
          class="px-3 py-1.5 rounded border border-zinc-700 hover:bg-zinc-800 text-zinc-200"
          @click="showCreate = !showCreate"
        >
          {{ showCreate ? '✕ Cancel' : '+ New project' }}
        </button>
      </div>
    </header>

    <!-- Create form -->
    <section
      v-if="showCreate"
      class="rounded-md border border-emerald-800/40 bg-emerald-950/10 p-4 space-y-3"
    >
      <div class="text-xs uppercase tracking-wide text-emerald-300">Create project</div>

      <div class="grid grid-cols-2 gap-3">
        <label class="block text-xs text-zinc-400">slug <span class="text-red-400">*</span>
          <input
            v-model="createSlug"
            class="w-full mt-1 px-2 py-1 rounded bg-zinc-950 border border-zinc-700 text-sm font-mono text-zinc-200"
            placeholder="kebab-case-slug"
          />
          <span v-if="createSlug && !slugValid" class="text-xs text-red-400">
            slug must match ^[a-z0-9-]+$
          </span>
        </label>
        <label class="block text-xs text-zinc-400">name <span class="text-red-400">*</span>
          <input
            v-model="createName"
            class="w-full mt-1 px-2 py-1 rounded bg-zinc-950 border border-zinc-700 text-sm text-zinc-200"
            placeholder="Human-readable project name"
          />
        </label>
      </div>

      <label class="block text-xs text-zinc-400">description
        <textarea
          v-model="createDesc"
          rows="2"
          class="w-full mt-1 px-2 py-1 rounded bg-zinc-950 border border-zinc-700 text-sm text-zinc-200"
        />
      </label>

      <label class="block text-xs text-zinc-400">root directory (optional; for future workspace mode)
        <input
          v-model="createRoot"
          class="w-full mt-1 px-2 py-1 rounded bg-zinc-950 border border-zinc-700 text-sm font-mono text-zinc-200"
          placeholder="projects/my-thing/"
        />
      </label>

      <div v-if="createErr" class="text-xs text-red-400">{{ createErr }}</div>

      <div class="flex gap-2">
        <button
          class="px-3 py-1.5 rounded text-xs bg-emerald-900/40 text-emerald-200 hover:bg-emerald-900/60 border border-emerald-800/60 disabled:opacity-50"
          :disabled="createBusy || !createSlug || !createName || !slugValid"
          @click="submitCreate"
        >Create</button>
      </div>
    </section>

    <p v-if="loading" class="text-sm text-zinc-400">loading…</p>
    <p v-else-if="error" class="text-sm text-red-400">{{ error }}</p>

    <section
      v-else-if="projects.length"
      class="rounded-md border border-zinc-800 bg-zinc-900/50 overflow-hidden"
    >
      <table class="w-full text-sm">
        <thead class="text-zinc-500 text-xs uppercase tracking-wide">
          <tr>
            <th class="text-left px-4 py-2 font-medium">Slug</th>
            <th class="text-left px-4 py-2 font-medium">Name</th>
            <th class="text-left px-4 py-2 font-medium">Description</th>
            <th class="text-right px-4 py-2 font-medium">Work items</th>
            <th class="text-right px-4 py-2 font-medium">Actions</th>
          </tr>
        </thead>
        <tbody>
          <template v-for="p in projects" :key="p.slug">
            <tr
              class="border-t border-zinc-800/50"
              :class="p.archived ? 'opacity-60' : 'hover:bg-zinc-900'"
            >
              <td class="px-4 py-2 font-mono text-xs">
                {{ p.slug }}
                <span v-if="p.archived" class="ml-2 text-zinc-500 text-xs">(archived)</span>
              </td>
              <td class="px-4 py-2 text-zinc-300">{{ p.name }}</td>
              <td class="px-4 py-2 text-zinc-400 text-xs max-w-xl">
                {{ p.description || '—' }}
              </td>
              <td class="px-4 py-2 text-right tabular-nums">
                <RouterLink
                  :to="`/work-items?project_association=${encodeURIComponent(p.slug)}`"
                  class="text-zinc-300 hover:text-white"
                >
                  {{ p.work_item_count }}
                </RouterLink>
              </td>
              <td class="px-4 py-2 text-right space-x-1">
                <button
                  class="px-2 py-1 rounded text-xs border border-zinc-700 hover:bg-zinc-800 text-zinc-300"
                  @click="startEdit(p)"
                >✎ edit</button>
                <button
                  class="px-2 py-1 rounded text-xs border border-zinc-700 hover:bg-zinc-800 text-zinc-300"
                  @click="toggleArchive(p)"
                >{{ p.archived ? 'unarchive' : 'archive' }}</button>
              </td>
            </tr>

            <!-- Inline edit row -->
            <tr v-if="editingSlug === p.slug" class="border-t border-zinc-800/50 bg-zinc-900/70">
              <td colspan="5" class="px-4 py-3">
                <div class="space-y-2 max-w-2xl">
                  <label class="block text-xs text-zinc-400">name
                    <input v-model="editName"
                      class="w-full mt-1 px-2 py-1 rounded bg-zinc-950 border border-zinc-700 text-sm text-zinc-200" />
                  </label>
                  <label class="block text-xs text-zinc-400">description
                    <textarea v-model="editDesc" rows="2"
                      class="w-full mt-1 px-2 py-1 rounded bg-zinc-950 border border-zinc-700 text-sm text-zinc-200" />
                  </label>
                  <label class="block text-xs text-zinc-400">root_directory
                    <input v-model="editRoot"
                      class="w-full mt-1 px-2 py-1 rounded bg-zinc-950 border border-zinc-700 text-sm font-mono text-zinc-200" />
                  </label>
                  <div class="flex gap-2 pt-1">
                    <button
                      class="px-3 py-1.5 rounded text-xs bg-emerald-900/40 text-emerald-200 hover:bg-emerald-900/60 border border-emerald-800/60 disabled:opacity-50"
                      :disabled="editBusy"
                      @click="submitEdit"
                    >Save</button>
                    <button
                      class="px-3 py-1.5 rounded text-xs border border-zinc-700 text-zinc-400 hover:bg-zinc-800"
                      :disabled="editBusy"
                      @click="cancelEdit"
                    >Cancel</button>
                  </div>
                </div>
              </td>
            </tr>
          </template>
        </tbody>
      </table>
    </section>

    <p v-else class="text-sm text-zinc-500 italic">
      No projects yet. Click "+ New project" to create one, or any
      work_item with a project_association string will surface its
      project here after the next backfill.
    </p>
  </div>
</template>
