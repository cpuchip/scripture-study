<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api, type IntentRow } from '@/api'

const intents = ref<IntentRow[]>([])
const error = ref('')
const loading = ref(true)
const expanded = ref<Record<string, boolean>>({})

async function load() {
  loading.value = true
  error.value = ''
  try {
    const r = await api.intentsList()
    intents.value = r.items
  } catch (e) {
    error.value = String(e)
  } finally {
    loading.value = false
  }
}

function toggle(id: string) {
  expanded.value[id] = !expanded.value[id]
}

onMounted(load)
</script>

<template>
  <div class="space-y-6 max-w-4xl">
    <header class="border-b border-zinc-800 pb-3">
      <h2 class="text-2xl font-semibold tracking-tight">Intents</h2>
      <p class="text-sm text-zinc-400 mt-1">
        Why each work_item exists. Edited in <code class="font-mono text-zinc-300">intent.yaml</code>;
        seeded into the substrate via the git pre-commit hook (Phase C).
      </p>
    </header>

    <p v-if="loading" class="text-sm text-zinc-400">loading…</p>
    <p v-else-if="error" class="text-sm text-red-400">{{ error }}</p>

    <ul v-else class="space-y-3">
      <li
        v-for="intent in intents"
        :key="intent.id"
        class="rounded-md border border-zinc-800 bg-zinc-900/50 p-4"
      >
        <div class="flex items-baseline justify-between gap-3">
          <div>
            <h3 class="text-lg font-semibold">{{ intent.slug }}</h3>
            <p class="text-sm text-zinc-300 mt-1">{{ intent.purpose }}</p>
          </div>
          <div class="flex items-center gap-3 text-xs">
            <span class="text-zinc-500">{{ intent.work_item_count }} work items</span>
            <button
              class="px-2 py-1 rounded border border-zinc-700 hover:bg-zinc-800 text-zinc-300"
              @click="toggle(intent.id)"
            >{{ expanded[intent.id] ? 'collapse' : 'details' }}</button>
          </div>
        </div>

        <div v-if="expanded[intent.id]" class="mt-3 space-y-2">
          <div v-if="intent.beneficiary" class="text-xs">
            <span class="text-zinc-500">Beneficiary:</span>
            <span class="text-zinc-300 ml-2">{{ intent.beneficiary }}</span>
          </div>
          <div v-if="intent.scripture_anchor" class="text-xs">
            <span class="text-zinc-500">Scripture anchor:</span>
            <span class="text-zinc-300 ml-2 font-mono">{{ intent.scripture_anchor }}</span>
          </div>
          <div v-if="intent.source_file" class="text-xs">
            <span class="text-zinc-500">Source:</span>
            <code class="text-zinc-300 ml-2">{{ intent.source_file }}</code>
          </div>

          <div v-if="intent.values_hierarchy?.length" class="mt-3">
            <div class="text-xs uppercase tracking-wide text-zinc-500 mb-1">
              Values (in priority order)
            </div>
            <ul class="text-xs space-y-1">
              <li
                v-for="(v, i) in intent.values_hierarchy"
                :key="i"
                class="text-zinc-300"
              >
                <span class="font-mono text-zinc-200">{{ v.key }}</span>
                <span
                  v-if="v.kind === 'constraint'"
                  class="text-amber-400 text-[10px] uppercase tracking-wide ml-2"
                >constraint{{ v.severity ? ' / ' + v.severity : '' }}</span>
                <div class="text-zinc-400 mt-0.5">{{ v.description }}</div>
              </li>
            </ul>
          </div>

          <div v-if="intent.non_goals?.length" class="mt-3">
            <div class="text-xs uppercase tracking-wide text-zinc-500 mb-1">Non-goals</div>
            <ul class="text-xs text-zinc-400 list-disc list-inside">
              <li v-for="(g, i) in intent.non_goals" :key="i">{{ g }}</li>
            </ul>
          </div>
        </div>
      </li>
    </ul>

    <p v-if="!loading && intents.length === 0" class="text-sm text-zinc-500">
      No intents yet. Edit <code class="font-mono">intent.yaml</code> at the repo root and commit
      to trigger the pre-commit seed hook.
    </p>
  </div>
</template>
