<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '../lib/api.js'

const props = defineProps({ id: String })
const router = useRouter()
const race = ref(null)
const cars = ref([])
const num = ref('')
const name = ref('')
const numInput = ref(null)

async function load() {
  race.value = await api.getRace(props.id)
  cars.value = await api.listCars(props.id) || []
}
onMounted(load)

const projectedHeats = computed(() => {
  // 3 lanes, 6 runs each → ceil(N*6/3) = 2N
  return cars.value.length >= 3 ? cars.value.length * 2 : 0
})

async function add() {
  const n = parseInt(num.value, 10)
  if (!n) return
  if (cars.value.find(c => c.number === n)) {
    alert(`Car #${n} already added.`)
    return
  }
  await api.addCar(props.id, n, name.value || '')
  num.value = ''
  name.value = ''
  await load()
  numInput.value?.focus()
}

async function remove(c) {
  if (!confirm(`Remove car #${c.number} (${c.name || 'unnamed'})?`)) return
  await api.deleteCar(props.id, c.id)
  await load()
}

async function rename(c) {
  const nm = prompt(`Rename car #${c.number}:`, c.name || '')
  if (nm === null) return
  await api.updateCar(props.id, c.id, c.number, nm)
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
          <input ref="numInput" v-model="num" type="number" inputmode="numeric"
            class="w-32 text-3xl font-mono text-center border-2 rounded-lg py-2" autofocus/>
        </div>
        <div class="flex-1">
          <label class="block text-sm">Name (optional)</label>
          <input v-model="name" type="text" class="w-full border rounded px-3 py-2"/>
        </div>
        <button class="btn-primary" type="submit">Add</button>
      </form>
    </div>

    <div class="card">
      <h2 class="text-lg font-semibold mb-3">Cars</h2>
      <table class="w-full">
        <thead>
          <tr class="text-left border-b">
            <th class="py-2 w-24">Number</th>
            <th>Name</th>
            <th class="w-32"></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="c in cars" :key="c.id" class="border-b">
            <td class="py-2 font-mono text-xl">#{{ c.number }}</td>
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
