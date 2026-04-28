<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { api } from '../lib/api.js'
import { useSocket } from '../lib/ws.js'

const props = defineProps({ id: String })
const heats = ref([])
const current = ref(null)

async function load() {
  const s = await api.state(props.id)
  heats.value = s.heats || []
  current.value = s.current?.heat_number || null
}
onMounted(load)

const sock = useSocket(msg => {
  if (msg.type === 'state' || msg.type === 'schedule_changed' || msg.type === 'cars_changed') load()
})
onUnmounted(() => sock.close())

function exportXlsx() { window.location = api.exportURL(props.id) }
</script>

<template>
  <div class="max-w-5xl mx-auto p-6 space-y-4">
    <div class="flex items-center justify-between print:hidden">
      <h1 class="text-3xl font-bold">Heat Chart</h1>
      <div class="space-x-2">
        <button class="btn-secondary" @click="window.print()">Print</button>
        <button class="btn-secondary" @click="exportXlsx">Export .xlsx</button>
      </div>
    </div>

    <table class="w-full bg-white rounded-xl shadow">
      <thead class="bg-slate-100">
        <tr>
          <th class="py-2 px-3 text-left w-20">Heat</th>
          <th class="py-2 px-3 text-left">Lane 1</th>
          <th class="py-2 px-3 text-left">Lane 2</th>
          <th class="py-2 px-3 text-left">Lane 3</th>
          <th class="py-2 px-3 text-left w-32">Status</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="h in heats" :key="h.id"
            :class="{
              'bg-yellow-100 font-semibold': h.heat_number === current,
              'opacity-60': h.status === 'complete'
            }">
          <td class="py-2 px-3 font-mono text-lg">{{ h.heat_number }}</td>
          <td v-for="lane in 3" :key="lane" class="py-2 px-3">
            <template v-for="s in h.slots" :key="s.id">
              <span v-if="s.lane === lane">
                #{{ s.car_number }}
                <span v-if="s.car_name" class="text-slate-500 text-sm">{{ s.car_name }}</span>
                <span v-if="s.place" class="ml-1 px-1.5 py-0.5 rounded bg-blue-100 text-blue-800 text-xs">
                  {{ s.place }}
                </span>
              </span>
            </template>
          </td>
          <td class="py-2 px-3 text-sm">{{ h.status }}</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<style>
@media print {
  header, .print\:hidden { display: none !important; }
}
</style>
