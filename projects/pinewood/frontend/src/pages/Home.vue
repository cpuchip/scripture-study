<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '../lib/api.js'

const router = useRouter()
const races = ref([])
const newName = ref('')
const importFile = ref(null)
const importMsg = ref('')

async function load() { races.value = await api.listRaces() || [] }
onMounted(load)

async function create() {
  const name = newName.value || `Pinewood Derby ${new Date().toLocaleDateString()}`
  const r = await api.createRace(name, 3)
  router.push(`/race/${r.id}/registration`)
}

async function doImport() {
  if (!importFile.value) return
  importMsg.value = 'Importing...'
  try {
    const res = await api.import(importFile.value)
    importMsg.value = `Imported race #${res.race.id}. ${(res.warnings || []).length} warnings.`
    await load()
    router.push(`/race/${res.race.id}/results`)
  } catch (e) {
    importMsg.value = 'Import failed: ' + e.message
  }
}

async function del(id) {
  if (!confirm(`Delete race #${id}? This cannot be undone.`)) return
  await api.deleteRace(id)
  await load()
}
</script>

<template>
  <div class="max-w-3xl mx-auto p-6 space-y-6">
    <h1 class="text-3xl font-bold">Pinewood Derby</h1>

    <div class="card space-y-3">
      <h2 class="text-xl font-semibold">Start a new race</h2>
      <div class="flex gap-2">
        <input v-model="newName" placeholder="Race name (optional)" class="flex-1 border rounded px-3 py-2"/>
        <button class="btn-primary" @click="create">Create</button>
      </div>
    </div>

    <div class="card space-y-3">
      <h2 class="text-xl font-semibold">Import .xlsx</h2>
      <input type="file" accept=".xlsx" @change="e => importFile = e.target.files[0]"/>
      <button class="btn-secondary" @click="doImport" :disabled="!importFile">Import</button>
      <p v-if="importMsg" class="text-sm text-slate-600">{{ importMsg }}</p>
    </div>

    <div class="card">
      <h2 class="text-xl font-semibold mb-3">Recent races</h2>
      <ul v-if="races.length" class="divide-y">
        <li v-for="r in races" :key="r.id" class="py-2 flex items-center justify-between">
          <RouterLink :to="`/race/${r.id}/registration`" class="text-blue-700 hover:underline">
            #{{ r.id }} — {{ r.name }} <span class="text-slate-500 text-sm">({{ r.status }})</span>
          </RouterLink>
          <button class="text-red-600 text-sm hover:underline" @click="del(r.id)">delete</button>
        </li>
      </ul>
      <p v-else class="text-slate-500">No races yet. Create one above.</p>
    </div>
  </div>
</template>
