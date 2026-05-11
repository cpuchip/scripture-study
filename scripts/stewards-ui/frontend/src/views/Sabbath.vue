<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import { api, type SabbathRow } from '@/api'

const reflections = ref<SabbathRow[]>([])
const loading = ref(true)
const error = ref('')

async function load() {
  loading.value = true
  error.value = ''
  try {
    const r = await api.sabbathList()
    reflections.value = r.items
  } catch (e) {
    error.value = String(e)
  } finally {
    loading.value = false
  }
}

function fmtDate(s?: string) {
  if (!s) return ''
  return new Date(s).toLocaleString()
}

onMounted(load)
</script>

<template>
  <div class="space-y-6 max-w-4xl">
    <header class="border-b border-zinc-800 pb-3">
      <h2 class="text-2xl font-semibold tracking-tight">Sabbath log</h2>
      <p class="text-sm text-zinc-400 mt-1">
        Reflections recorded when work_items reach verified maturity. The point isn't notification — it's that endings are recorded, not skipped.
      </p>
    </header>

    <p v-if="loading" class="text-sm text-zinc-400">loading…</p>
    <p v-else-if="error" class="text-sm text-red-400">{{ error }}</p>

    <p v-else-if="reflections.length === 0" class="text-sm text-zinc-500">
      No sabbath reflections yet. They land here when a work_item completes
      its verified maturity on a sabbath-enabled pipeline.
    </p>

    <ul v-else class="space-y-4">
      <li
        v-for="r in reflections"
        :key="r.id"
        class="rounded-md border border-zinc-800 bg-zinc-900/50 p-4 space-y-3"
      >
        <div class="flex items-baseline justify-between gap-3">
          <div>
            <RouterLink
              :to="`/work-items/${r.work_item_id}`"
              class="text-base font-semibold hover:text-emerald-300"
            >{{ r.work_item_slug || r.work_item_id }}</RouterLink>
            <span class="text-xs text-zinc-500 font-mono ml-2">{{ r.pipeline_family }}</span>
          </div>
          <span class="text-xs text-zinc-500 tabular-nums">{{ fmtDate(r.at) }}</span>
        </div>

        <div>
          <div class="text-xs uppercase tracking-wide text-zinc-500 mb-1">Reflection</div>
          <p class="text-sm text-zinc-200 leading-relaxed">{{ r.reflection }}</p>
        </div>

        <div v-if="r.carry_forward" class="border-l-2 border-emerald-700/50 pl-3">
          <div class="text-xs uppercase tracking-wide text-emerald-500 mb-1">Carry forward</div>
          <p class="text-sm text-zinc-200">{{ r.carry_forward }}</p>
        </div>

        <div v-if="r.surprise" class="border-l-2 border-amber-700/50 pl-3">
          <div class="text-xs uppercase tracking-wide text-amber-500 mb-1">Surprise</div>
          <p class="text-sm text-zinc-200">{{ r.surprise }}</p>
        </div>
      </li>
    </ul>
  </div>
</template>
