<script setup>
import { ref, computed, onMounted, onUnmounted, nextTick, watch } from 'vue'
import { api } from '../lib/api.js'
import { useSocket } from '../lib/ws.js'

const props = defineProps({ id: String, heat: String })
const heats = ref([])
const heatIdx = ref(0)
const places = ref([null, null, null])
const inputEls = []          // populated by :ref function below
const error = ref('')

async function load() {
  heats.value = await api.heats(props.id) || []
  if (props.heat) {
    const idx = heats.value.findIndex(h => h.heat_number === parseInt(props.heat, 10))
    if (idx >= 0) heatIdx.value = idx
  } else {
    const idx = heats.value.findIndex(h => h.status !== 'complete')
    if (idx >= 0) heatIdx.value = idx
  }
  syncPlaces()
}
onMounted(load)

const sock = useSocket(msg => {
  if (msg.type === 'state' || msg.type === 'schedule_changed') load()
})
onUnmounted(() => sock.close())

const heat = computed(() => heats.value[heatIdx.value] || null)

// ±2 surrounding heats for verification context.
const contextHeats = computed(() => {
  const out = []
  for (let d = -2; d <= 2; d++) {
    const i = heatIdx.value + d
    if (i >= 0 && i < heats.value.length) {
      out.push({ ...heats.value[i], _offset: d })
    }
  }
  return out
})

function syncPlaces() {
  if (!heat.value) return
  places.value = [null, null, null]
  for (const s of heat.value.slots) {
    places.value[s.lane - 1] = s.place ?? null
  }
  nextTick(focusFirstEmpty)
}

watch(heatIdx, syncPlaces)

function focusLane(lane) {
  const el = inputEls[lane - 1]
  if (el && !el.disabled) {
    el.focus()
    el.select?.()
    return true
  }
  return false
}

function focusFirstEmpty() {
  for (let i = 0; i < 3; i++) {
    if (places.value[i] == null && hasCar(i + 1)) {
      if (focusLane(i + 1)) return
    }
  }
}

function hasCar(lane) {
  return !!heat.value?.slots.find(s => s.lane === lane && s.car_id)
}

function carInLane(lane) {
  return heat.value?.slots.find(s => s.lane === lane)
}

async function setPlace(lane, val) {
  error.value = ''
  let p = null
  if (val === '' || val == null) {
    p = null
  } else {
    const n = parseInt(val, 10)
    if (![1, 2, 3].includes(n)) {
      error.value = `Lane ${lane}: place must be 1, 2 or 3.`
      return false
    }
    const dup = places.value.findIndex((v, i) => v === n && (i + 1) !== lane)
    if (dup >= 0) {
      if (!confirm(`Lane ${dup + 1} already has place ${n}. Continue?`)) return false
    }
    p = n
  }
  places.value[lane - 1] = p
  try {
    await api.score(props.id, heat.value.heat_number, lane, p)
  } catch (e) {
    error.value = 'Save failed: ' + e.message
    return false
  }
  return true
}

// Save current input, then move focus to the next lane that needs a score.
// If all lanes in this heat are scored, advance to the next heat.
async function saveAndAdvance(lane, val) {
  const ok = await setPlace(lane, val)
  if (!ok) return
  // Look for next lane after `lane` that still needs a place.
  for (let nextLane = lane + 1; nextLane <= 3; nextLane++) {
    if (places.value[nextLane - 1] == null && hasCar(nextLane)) {
      await nextTick()
      if (focusLane(nextLane)) return
    }
  }
  // Then check earlier lanes (in case officials skipped one).
  for (let i = 1; i < lane; i++) {
    if (places.value[i - 1] == null && hasCar(i)) {
      await nextTick()
      if (focusLane(i)) return
    }
  }
  // All lanes scored → next heat.
  if (heatIdx.value < heats.value.length - 1) {
    heatIdx.value++
  }
}

function jump(delta) {
  const next = heatIdx.value + delta
  if (next >= 0 && next < heats.value.length) heatIdx.value = next
}

function jumpToHeat(num) {
  const idx = heats.value.findIndex(h => h.heat_number === num)
  if (idx >= 0) heatIdx.value = idx
}

function jumpToPrompt() {
  const n = prompt(`Jump to heat (1–${heats.value.length}):`)
  if (!n) return
  jumpToHeat(parseInt(n, 10))
}

function onKey(e, lane) {
  if (e.key === 'Enter' || e.key === 'Tab') {
    e.preventDefault()
    saveAndAdvance(lane, e.target.value)
  } else if (e.key === 'Backspace' && e.target.value === '') {
    e.preventDefault()
    setPlace(lane, null)
  } else if (e.key === 'ArrowRight') {
    e.preventDefault()
    focusLane(Math.min(3, lane + 1))
  } else if (e.key === 'ArrowLeft') {
    e.preventDefault()
    focusLane(Math.max(1, lane - 1))
  }
}

function slotForLane(h, lane) {
  return h.slots?.find(s => s.lane === lane)
}
</script>

<template>
  <div class="max-w-4xl mx-auto p-6">
    <div v-if="!heat" class="card text-center text-slate-500">
      No heats yet. Finalize registration to generate the schedule.
    </div>
    <template v-else>
      <div class="flex items-center justify-between mb-4">
        <button class="btn-secondary" :disabled="heatIdx === 0" @click="jump(-1)">← Prev</button>
        <div class="text-center">
          <div class="text-sm text-slate-500">Heat</div>
          <div class="text-5xl font-bold">{{ heat.heat_number }}</div>
          <button class="text-sm text-blue-700 hover:underline" @click="jumpToPrompt">jump to…</button>
          <span class="ml-2 text-sm text-slate-500">({{ heatIdx + 1 }} / {{ heats.length }})</span>
        </div>
        <button class="btn-secondary" :disabled="heatIdx >= heats.length - 1" @click="jump(1)">Next →</button>
      </div>

      <div class="grid grid-cols-3 gap-4">
        <div v-for="lane in 3" :key="lane" class="card text-center">
          <div class="text-sm text-slate-500 mb-1">Lane {{ lane }}</div>
          <div class="text-4xl font-mono font-bold mb-2">
            <template v-if="carInLane(lane)?.car_id">
              #{{ carInLane(lane).car_number }}
            </template>
            <template v-else>—</template>
          </div>
          <div class="text-sm text-slate-600 mb-3 h-5">{{ carInLane(lane)?.car_name || '' }}</div>
          <input
            :ref="el => inputEls[lane - 1] = el"
            class="numpad-input"
            inputmode="numeric"
            type="number"
            min="1" max="3"
            :value="places[lane - 1] ?? ''"
            :disabled="!hasCar(lane)"
            @keydown="onKey($event, lane)"
            placeholder="—"
          />
          <div class="text-xs text-slate-500 mt-1">Type 1, 2 or 3 → Enter</div>
        </div>
      </div>

      <p v-if="error" class="text-red-600 mt-4">{{ error }}</p>

      <!-- Verification strip: previous, current, upcoming heats. -->
      <div class="mt-8 card">
        <div class="text-sm font-semibold text-slate-600 mb-2 uppercase tracking-wider">
          Recently scored / upcoming
        </div>
        <table class="w-full text-sm">
          <thead>
            <tr class="text-left text-slate-500 border-b">
              <th class="py-1 w-16">Heat</th>
              <th class="py-1">Lane 1</th>
              <th class="py-1">Lane 2</th>
              <th class="py-1">Lane 3</th>
              <th class="py-1 w-20"></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="h in contextHeats" :key="h.id"
                :class="{
                  'bg-yellow-50 font-semibold': h._offset === 0,
                  'text-slate-400': h._offset > 0,
                  'cursor-pointer hover:bg-slate-50': h._offset !== 0,
                }"
                @click="h._offset !== 0 && jumpToHeat(h.heat_number)">
              <td class="py-2 font-mono">{{ h.heat_number }}</td>
              <td v-for="lane in 3" :key="lane" class="py-2">
                <template v-if="slotForLane(h, lane)?.car_id">
                  <span class="font-mono">#{{ slotForLane(h, lane).car_number }}</span>
                  <span v-if="slotForLane(h, lane).place"
                        class="ml-2 inline-block px-2 py-0.5 rounded bg-blue-600 text-white text-xs font-bold">
                    {{ slotForLane(h, lane).place }}
                  </span>
                  <span v-else-if="h._offset < 0"
                        class="ml-2 text-red-500 text-xs">unscored</span>
                </template>
                <template v-else>—</template>
              </td>
              <td class="py-2 text-right text-xs">
                <span v-if="h._offset === 0" class="text-yellow-700">scoring</span>
                <span v-else-if="h._offset < 0" class="text-slate-400">{{ h.status }}</span>
                <span v-else class="text-slate-400">upcoming</span>
              </td>
            </tr>
          </tbody>
        </table>
        <p class="text-xs text-slate-400 mt-2">Click a previous or upcoming row to jump there.</p>
      </div>
    </template>
  </div>
</template>
