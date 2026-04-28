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

// Most recent completed heat — sits above On Deck so parents can verify what was recorded.
const justScored = computed(() => {
  const heats = state.value?.heats || []
  const cur = current.value
  const startIdx = cur
    ? heats.findIndex(h => h.id === cur.id) - 1
    : heats.length - 1
  for (let i = startIdx; i >= 0; i--) {
    if (heats[i].status === 'complete') return heats[i]
  }
  return null
})
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

      <!-- Just scored + On-deck + leaderboard -->
      <div class="space-y-6">
        <div v-if="justScored">
          <div class="text-2xl text-emerald-400 uppercase tracking-wider">Just Scored</div>
          <div class="text-4xl font-bold mb-2">Heat {{ justScored.heat_number }}</div>
          <div class="grid grid-cols-3 gap-3">
            <div v-for="lane in 3" :key="lane"
                 class="bg-emerald-900/40 border border-emerald-800 rounded-xl p-3 text-center">
              <div class="text-emerald-300 text-sm">L{{ lane }}</div>
              <template v-for="s in justScored.slots" :key="s.id">
                <div v-if="s.lane === lane">
                  <div class="text-2xl font-mono font-bold">#{{ s.car_number }}</div>
                  <div class="text-xs text-emerald-200 truncate">{{ s.car_name }}</div>
                  <div v-if="s.place" class="text-5xl font-bold text-yellow-400 mt-1">{{ s.place }}</div>
                  <div v-else class="text-xl text-slate-500 mt-1">—</div>
                </div>
              </template>
            </div>
          </div>
        </div>

        <div v-if="onDeck">
          <div class="text-2xl text-slate-400 uppercase tracking-wider">On Deck</div>
          <div class="text-4xl font-bold mb-2">Heat {{ onDeck.heat_number }}</div>
          <div class="grid grid-cols-3 gap-3">
            <div v-for="lane in 3" :key="lane" class="bg-slate-900 rounded-xl p-3 text-center">
              <div class="text-slate-400 text-sm">L{{ lane }}</div>
              <template v-for="s in onDeck.slots" :key="s.id">
                <div v-if="s.lane === lane">
                  <div class="text-2xl font-mono font-bold">#{{ s.car_number }}</div>
                  <div class="text-xs text-slate-400 truncate">{{ s.car_name }}</div>
                </div>
              </template>
            </div>
          </div>
        </div>

        <div>
          <div class="text-2xl text-slate-400 uppercase tracking-wider mb-3">Top 5 (lowest = best)</div>
          <table class="w-full text-2xl">
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
