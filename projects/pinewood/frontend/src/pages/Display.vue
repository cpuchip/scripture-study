<script setup>
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { api } from '../lib/api.js'
import { useSocket } from '../lib/ws.js'

const props = defineProps({ id: String })
const state = ref(null)

async function load() { state.value = await api.state(props.id) }
onMounted(load)

const sock = useSocket(msg => {
  if (msg.type === 'state' || msg.type === 'schedule_changed') load()
})
onUnmounted(() => sock.close())

const current = computed(() => state.value?.current || null)
const onDeck = computed(() => state.value?.on_deck || null)
const top5 = computed(() => (state.value?.standings || []).slice(0, 5))
</script>

<template>
  <div class="min-h-screen bg-black text-white p-8" v-if="state">
    <div class="grid grid-cols-2 gap-8 h-full">
      <!-- Current heat -->
      <div>
        <div class="text-3xl text-slate-400 uppercase tracking-wider">Now Racing</div>
        <div class="text-9xl font-bold mb-6">Heat {{ current?.heat_number ?? '—' }}</div>
        <div v-if="current" class="space-y-4">
          <template v-for="lane in 3" :key="lane">
            <div class="bg-slate-800 rounded-2xl p-6 flex items-center gap-6">
              <div class="text-5xl text-slate-400 font-mono">L{{ lane }}</div>
              <template v-for="s in current.slots" :key="s.id">
                <div v-if="s.lane === lane" class="flex-1">
                  <div class="text-3xl">{{ s.car_name || ('Car #' + s.car_number) }}</div>
                  <div class="text-7xl font-bold font-mono">#{{ s.car_number }}</div>
                </div>
                <div v-if="s.lane === lane && s.place" class="text-9xl font-bold text-yellow-400">
                  {{ s.place }}
                </div>
              </template>
            </div>
          </template>
        </div>
      </div>

      <!-- On-deck + leaderboard -->
      <div class="space-y-8">
        <div v-if="onDeck">
          <div class="text-2xl text-slate-400 uppercase tracking-wider">On Deck</div>
          <div class="text-6xl font-bold mb-3">Heat {{ onDeck.heat_number }}</div>
          <div class="grid grid-cols-3 gap-3">
            <div v-for="lane in 3" :key="lane" class="bg-slate-900 rounded-xl p-3 text-center">
              <div class="text-slate-400 text-sm">L{{ lane }}</div>
              <template v-for="s in onDeck.slots" :key="s.id">
                <div v-if="s.lane === lane">
                  <div class="text-3xl font-mono font-bold">#{{ s.car_number }}</div>
                  <div class="text-sm text-slate-400">{{ s.car_name }}</div>
                </div>
              </template>
            </div>
          </div>
        </div>

        <div>
          <div class="text-2xl text-slate-400 uppercase tracking-wider mb-3">Top 5 (lowest = best)</div>
          <table class="w-full text-3xl">
            <tr v-for="s in top5" :key="s.car_id" class="border-b border-slate-700">
              <td class="py-2 text-slate-400 w-16">#{{ s.rank }}</td>
              <td class="py-2 font-mono">#{{ s.car_number }}</td>
              <td class="py-2 text-slate-300">{{ s.car_name }}</td>
              <td class="py-2 text-right font-bold">{{ s.total }}</td>
            </tr>
          </table>
        </div>
      </div>
    </div>
  </div>
</template>
