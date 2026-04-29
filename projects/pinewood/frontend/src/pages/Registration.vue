<script setup>
import { ref, onMounted, computed, nextTick, watch } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '../lib/api.js'

const props = defineProps({ id: String })
const router = useRouter()
const race = ref(null)
const cars = ref([])
const num = ref('')
const name = ref('')
const numInput = ref(null)
const nameInput = ref(null)

// Inline edit state for car numbers in the list.
const editingId = ref(null)
const editingNumber = ref('')
const editError = ref('')

// Sort state.
const sortKey = ref('number')   // 'number' | 'name'
const sortDir = ref('desc')     // 'asc' | 'desc' — default: most recent number first

async function load() {
  race.value = await api.getRace(props.id)
  cars.value = await api.listCars(props.id) || []
  if (!num.value) num.value = String(nextAvailableNumber())
}
onMounted(load)

// Lowest unused positive integer (fills gaps).
function nextAvailableNumber() {
  const used = new Set(cars.value.map(c => c.number))
  let n = 1
  while (used.has(n)) n++
  return n
}

const projectedHeats = computed(() => cars.value.length >= 3 ? cars.value.length * 2 : 0)

const sortedCars = computed(() => {
  const arr = [...cars.value]
  const dir = sortDir.value === 'asc' ? 1 : -1
  arr.sort((a, b) => {
    if (sortKey.value === 'number') return (a.number - b.number) * dir
    const an = (a.name || '').toLowerCase()
    const bn = (b.name || '').toLowerCase()
    if (an < bn) return -1 * dir
    if (an > bn) return 1 * dir
    return (a.number - b.number) * dir
  })
  return arr
})

function setSort(key) {
  if (sortKey.value === key) {
    sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortKey.value = key
    sortDir.value = key === 'number' ? 'desc' : 'asc'
  }
}

function arrow(key) {
  if (sortKey.value !== key) return ''
  return sortDir.value === 'asc' ? ' ▲' : ' ▼'
}

async function add() {
  const n = parseInt(num.value, 10)
  if (!n || n < 1) {
    alert('Enter a positive car number.')
    return
  }
  if (cars.value.find(c => c.number === n)) {
    alert(`Car #${n} already added.`)
    return
  }
  try {
    await api.addCar(props.id, n, name.value || '')
  } catch (e) {
    alert('Add failed: ' + e.message)
    return
  }
  num.value = ''
  name.value = ''
  await load()
  // After first add, focus stays on Name so officials can rapid-fire car names
  // for the auto-populated next number.
  await nextTick()
  nameInput.value?.focus()
}

async function remove(c) {
  if (!confirm(`Remove car #${c.number} (${c.name || 'unnamed'})?`)) return
  await api.deleteCar(props.id, c.id)
  await load()
}

async function rename(c) {
  const nm = prompt(`Rename car #${c.number}:`, c.name || '')
  if (nm === null) return
  try {
    await api.updateCar(props.id, c.id, c.number, nm)
  } catch (e) {
    alert('Rename failed: ' + e.message)
    return
  }
  await load()
}

function startEditNumber(c) {
  editingId.value = c.id
  editingNumber.value = String(c.number)
  editError.value = ''
  nextTick(() => {
    const el = document.getElementById(`edit-num-${c.id}`)
    el?.focus()
    el?.select?.()
  })
}

function cancelEditNumber() {
  editingId.value = null
  editingNumber.value = ''
  editError.value = ''
}

async function saveEditNumber(c) {
  const n = parseInt(editingNumber.value, 10)
  if (!n || n < 1) {
    editError.value = 'Number must be positive.'
    return
  }
  if (n === c.number) {
    cancelEditNumber()
    return
  }
  if (cars.value.find(other => other.id !== c.id && other.number === n)) {
    editError.value = `Car #${n} already exists.`
    return
  }
  try {
    await api.updateCar(props.id, c.id, n, c.name || '')
  } catch (e) {
    editError.value = 'Save failed: ' + e.message
    return
  }
  cancelEditNumber()
  await load()
}

async function finalize() {
  if (cars.value.length < 3) { alert('Need at least 3 cars.'); return }
  if (!confirm(`Lock registration with ${cars.value.length} cars and generate the heat chart?`)) return
  try {
    await api.finalize(props.id)
    router.push(`/race/${props.id}/schedule`)
  } catch (e) {
    alert('Finalize failed: ' + e.message)
  }
}
</script>

<template>
  <div class="max-w-4xl mx-auto p-6 space-y-6">
    <h1 class="text-3xl font-bold">{{ race?.name }} — Registration</h1>
    <p class="text-slate-600">
      {{ cars.length }} cars registered · projected {{ projectedHeats }} heats
      <span v-if="race?.finalized_at" class="ml-2 px-2 py-0.5 bg-green-100 text-green-800 rounded text-sm">
        Finalized — late adds will regenerate the schedule
      </span>
    </p>

    <div class="card space-y-3">
      <h2 class="text-lg font-semibold">Add a car</h2>
      <form class="flex gap-2 items-end" @submit.prevent="add">
        <div>
          <label class="block text-sm">Number</label>
          <input ref="numInput" v-model="num" type="number" inputmode="numeric" min="1"
            class="w-32 text-3xl font-mono text-center border-2 rounded-lg py-2"/>
        </div>
        <div class="flex-1">
          <label class="block text-sm">Name (optional)</label>
          <input ref="nameInput" v-model="name" type="text" autofocus
            class="w-full border rounded px-3 py-2"/>
        </div>
        <button class="btn-primary" type="submit">Add</button>
      </form>
      <p class="text-xs text-slate-500">
        Number auto-fills to the next available (lowest unused). Edit it if needed; press Enter to add and start the next car.
      </p>
    </div>

    <div class="card">
      <h2 class="text-lg font-semibold mb-3">Cars</h2>
      <table class="w-full">
        <thead>
          <tr class="text-left border-b">
            <th class="py-2 w-32 cursor-pointer select-none hover:bg-slate-50"
                @click="setSort('number')">Number{{ arrow('number') }}</th>
            <th class="cursor-pointer select-none hover:bg-slate-50"
                @click="setSort('name')">Name{{ arrow('name') }}</th>
            <th class="w-32"></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="c in sortedCars" :key="c.id" class="border-b">
            <td class="py-2 font-mono text-xl">
              <template v-if="editingId === c.id">
                <input :id="`edit-num-${c.id}`" v-model="editingNumber" type="number" min="1"
                  class="w-20 border-2 rounded px-2 py-1 font-mono text-lg"
                  @keydown.enter.prevent="saveEditNumber(c)"
                  @keydown.escape.prevent="cancelEditNumber"
                  @blur="saveEditNumber(c)"/>
                <div v-if="editError" class="text-red-600 text-xs mt-1">{{ editError }}</div>
              </template>
              <button v-else class="hover:bg-slate-100 rounded px-2 py-1"
                      title="Click to edit number" @click="startEditNumber(c)">
                #{{ c.number }}
              </button>
            </td>
            <td>{{ c.name || '—' }}</td>
            <td class="text-right space-x-2">
              <button class="text-blue-700 hover:underline text-sm" @click="rename(c)">rename</button>
              <button class="text-red-600 hover:underline text-sm" @click="remove(c)">remove</button>
            </td>
          </tr>
        </tbody>
      </table>
      <p v-if="!cars.length" class="text-slate-500 py-4">No cars yet — add one above.</p>
    </div>

    <div class="flex justify-end">
      <button class="btn-primary text-lg px-6 py-3" :disabled="cars.length < 3" @click="finalize">
        {{ race?.finalized_at ? 'Re-generate schedule' : 'Finalize & generate schedule' }}
      </button>
    </div>
  </div>
</template>
