<script setup>
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '../lib/api.js'
import { useSocket } from '../lib/ws.js'

const props = defineProps({ id: String })
const router = useRouter()
const data = ref({ standings: [], ties: [] })
const selected = ref(new Set())

async function load() {
  data.value = await api.standings(props.id)
  // Pre-select tied top cars by default.
  selected.value = new Set((data.value.ties[0] || []))
}
onMounted(load)

const sock = useSocket(msg => {
  if (msg.type === 'state') load()
})
onUnmounted(() => sock.close())

const standings = computed(() => data.value.standings || [])
const topTies = computed(() => data.value.ties || [])

function toggle(num) {
  const s = new Set(selected.value)
  if (s.has(num)) s.delete(num); else s.add(num)
  selected.value = s
}

async function startRunoff() {
  const cars = [...selected.value]
  if (cars.length < 2) { alert('Select at least 2 cars.'); return }
  try {
    const r = await api.runoff(props.id, cars, '')
    router.push(`/race/${r.id}/score`)
  } catch (e) {
    alert('Run-off failed: ' + e.message)
  }
}

function exportXlsx() { window.location = api.exportURL(props.id) }
</script>

<template>
  <div class="max-w-4xl mx-auto p-6 space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-3xl font-bold">Results</h1>
      <button class="btn-secondary" @click="exportXlsx">Export .xlsx</button>
    </div>

    <div v-if="topTies.length" class="card border-l-4 border-yellow-400 bg-yellow-50">
      <div class="font-semibold text-yellow-900 mb-2">⚠️ Tie at the top</div>
      <div v-for="(group, idx) in topTies" :key="idx" class="text-sm text-yellow-800">
        Cars tied: <span v-for="(c, i) in group" :key="c">
          <span class="font-mono">#{{ c }}</span><span v-if="i < group.length - 1">, </span>
        </span>
      </div>
      <p class="text-sm text-yellow-700 mt-2">Pre-selected for run-off below.</p>
    </div>

    <div class="card">
      <h2 class="text-lg font-semibold mb-3">Standings (lowest score wins)</h2>
      <table class="w-full">
        <thead>
          <tr class="text-left border-b">
            <th class="w-12 py-2"></th>
            <th class="w-16">Rank</th>
            <th class="w-20">Car</th>
            <th>Name</th>
            <th class="w-16 text-right">Heats</th>
            <th class="w-20 text-right">Total</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="s in standings" :key="s.car_id" class="border-b">
            <td class="py-2 text-center">
              <input type="checkbox" :checked="selected.has(s.car_number)"
                @change="toggle(s.car_number)"/>
            </td>
            <td class="font-bold">{{ s.rank }}</td>
            <td class="font-mono">#{{ s.car_number }}</td>
            <td>{{ s.car_name || '—' }}</td>
            <td class="text-right text-slate-500">{{ s.heats }}</td>
            <td class="text-right font-bold">{{ s.total }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <div class="card flex items-center justify-between">
      <div>
        <span class="font-semibold">{{ selected.size }}</span> car<span v-if="selected.size !== 1">s</span> selected for run-off
      </div>
      <button class="btn-primary" :disabled="selected.size < 2" @click="startRunoff">
        Start run-off
      </button>
    </div>
  </div>
</template>
