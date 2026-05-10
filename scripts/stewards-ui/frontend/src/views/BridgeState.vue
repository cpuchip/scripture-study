<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api, type ServerState } from '@/api'

const servers = ref<ServerState[]>([])
const error = ref<string>('')
const loading = ref(false)
const expanded = ref<Set<string>>(new Set())

async function load() {
  loading.value = true
  try {
    const r = await api.bridgeState()
    servers.value = r.servers
  } catch (e) {
    error.value = String(e)
  } finally {
    loading.value = false
  }
}
onMounted(load)

function toggle(name: string) {
  if (expanded.value.has(name)) expanded.value.delete(name)
  else expanded.value.add(name)
  // Trigger reactivity since Set mutation isn't deeply tracked
  expanded.value = new Set(expanded.value)
}

function fmtRelative(s?: string) {
  if (!s) return ''
  const d = new Date(s)
  if (isNaN(d.getTime())) return s
  const sec = Math.floor((Date.now() - d.getTime()) / 1000)
  if (sec < 60) return `${sec}s ago`
  if (sec < 3600) return `${Math.floor(sec / 60)}m ago`
  if (sec < 86400) return `${Math.floor(sec / 3600)}h ago`
  return `${Math.floor(sec / 86400)}d ago`
}
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-baseline justify-between">
      <h2 class="text-2xl font-semibold tracking-tight">Bridge state</h2>
      <button
        class="text-xs px-2 py-1 rounded border border-zinc-700 hover:bg-zinc-800"
        @click="load"
      >refresh</button>
    </div>

    <p v-if="loading" class="text-sm text-zinc-400">loading…</p>
    <p v-else-if="error" class="text-sm text-red-400">{{ error }}</p>

    <ul v-else class="space-y-2">
      <li
        v-for="s in servers"
        :key="s.server"
        class="rounded-md border border-zinc-800 bg-zinc-900/50 overflow-hidden"
      >
        <button
          class="w-full px-4 py-3 flex items-baseline gap-3 text-left hover:bg-zinc-900"
          @click="toggle(s.server)"
        >
          <span
            class="inline-block w-2 h-2 rounded-full"
            :class="{
              'bg-emerald-500': s.enabled && !s.last_error,
              'bg-amber-500': s.enabled && s.last_error,
              'bg-zinc-500': !s.enabled,
            }"
          ></span>
          <span class="font-mono font-semibold">{{ s.server }}</span>
          <span class="text-xs text-zinc-500">{{ s.transport }}</span>
          <span
            class="text-xs px-2 py-0.5 rounded"
            :class="s.enabled ? 'bg-emerald-900/30 text-emerald-300' : 'bg-zinc-800 text-zinc-500'"
          >{{ s.enabled ? 'enabled' : 'disabled' }}</span>
          <span class="text-xs text-zinc-400">{{ s.active_tools }} tools</span>
          <span v-if="s.last_tools_refresh_at" class="ml-auto text-xs text-zinc-500">
            refresh {{ fmtRelative(s.last_tools_refresh_at) }}
          </span>
        </button>

        <div v-if="s.last_error" class="px-4 py-2 text-xs text-red-300 bg-red-950/20 border-t border-red-900/30 font-mono">
          {{ s.last_error }}
        </div>

        <div
          v-if="expanded.has(s.server) && s.tools?.length"
          class="border-t border-zinc-800 bg-zinc-950/40"
        >
          <ul class="divide-y divide-zinc-800/50">
            <li v-for="t in s.tools" :key="t.name" class="px-4 py-2 text-sm">
              <div class="flex items-baseline gap-3">
                <span class="font-mono text-zinc-200">{{ t.name }}</span>
                <span
                  v-if="!t.active"
                  class="text-xs px-2 py-0.5 rounded bg-zinc-800 text-zinc-500"
                >inactive</span>
              </div>
              <p v-if="t.description" class="text-xs text-zinc-400 mt-1">
                {{ t.description }}
              </p>
            </li>
          </ul>
        </div>
      </li>
    </ul>
  </div>
</template>
