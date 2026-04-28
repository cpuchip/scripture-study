<script setup>
import { ref, computed, onMounted, onUnmounted, nextTick, watch } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '../lib/api.js'
import { useSocket } from '../lib/ws.js'

const props = defineProps({ id: String, heat: String })
const router = useRouter()
const heats = ref([])
const heatIdx = ref(0)  // index into heats
const places = ref([null, null, null])
const inputRefs = [ref(null), ref(null), ref(null)]
const error = ref('')

async function load() {
  heats.value = await api.heats(props.id) || []
  if (props.heat) {
    const idx = heats.value.findIndex(h => h.heat_number === parseInt(props.heat, 10))
    if (idx >= 0) heatIdx.value = idx
  } else {
    // jump to first non-complete heat
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

function syncPlaces() {
  if (!heat.value) return
  places.value = [null, null, null]
  for (const s of heat.value.slots) {
    places.value[s.lane - 1] = s.place ?? null
  }
  nextTick(() => focusFirstEmpty())
}

watch(heatIdx, syncPlaces)

function focusFirstEmpty() {
  for (let i = 0; i < 3; i++) {
    if (places.value[i] == null && hasCar(i + 1)) {
      inputRefs[i].value?.focus()
      inputRefs[i].value?.select?.()
      return
    }
  }
}

function hasCar(lane) {
  return heat.value?.slots.find(s => s.lane === lane && s.car_id)
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
      return
    }
    // Detect duplicate within this heat (warn, not block — manual override may be intentional).
    const dup = places.value.findIndex((v, i) => v === n && (i + 1) !== lane)
    if (dup >= 0) {
      if (!confirm(`Lane ${dup + 1} already has place ${n}. Continue?`)) {
        return
      }
    }
    p = n
  }
  places.value[lane - 1] = p
  try {
    await api.score(props.id, heat.value.heat_number, lane, p)
  } catch (e) {
    error.value = 'Save failed: ' + e.message
    return
  }
  // auto-advance
  if (p != null) advanceFocus(lane)
}

function advanceFocus(fromLane) {
  for (let i = fromLane; i < 3; i++) {
    if (places.value[i] == null && hasCar(i + 1)) {
      inputRefs[i].value?.focus()
      inputRefs[i].value?.select?.()
      return
    }
  }
  // all filled → next heat
  if (heatIdx.value < heats.value.length - 1) {
    heatIdx.value++
  }
}

function jump(delta) {
  const next = heatIdx.value + delta
  if (next >= 0 && next < heats.value.length) heatIdx.value = next
}

function jumpTo() {
  const n = prompt(`Jump to heat (1–${heats.value.length}):`)
  if (!n) return
  const idx = heats.value.findIndex(h => h.heat_number === parseInt(n, 10))
  if (idx >= 0) heatIdx.value = idx
}

function onKey(e, lane) {
  if (e.key === 'Enter') {
    e.preventDefault()
    setPlace(lane, e.target.value)
  } else if (e.key === 'Backspace' && e.target.value === '') {
    e.preventDefault()
    setPlace(lane, null)
  } else if (e.key === 'ArrowRight') {
    inputRefs[Math.min(2, lane)].value?.focus()
  } else if (e.key === 'ArrowLeft') {
    inputRefs[Math.max(0, lane - 2)].value?.focus()
  }
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
          <button class="text-sm text-blue-700 hover:underline" @click="jumpTo">jump to…</button>
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
            :ref="inputRefs[lane - 1]"
            class="numpad-input"
            inputmode="numeric"
            type="number"
            min="1" max="3"
            :value="places[lane - 1] ?? ''"
            :disabled="!hasCar(lane)"
            @keydown="onKey($event, lane)"
            placeholder="—"
          />
          <div class="text-xs text-slate-500 mt-1">Enter 1, 2 or 3 → Enter</div>
        </div>
      </div>

      <p v-if="error" class="text-red-600 mt-4">{{ error }}</p>
    </template>
  </div>
</template>
