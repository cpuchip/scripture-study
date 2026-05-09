<script setup lang="ts">
import { ref, onMounted } from 'vue'

const health = ref<'pending' | 'ok' | 'fail'>('pending')
const detail = ref<string>('')

async function check() {
  try {
    const r = await fetch('/healthz')
    if (r.ok) {
      health.value = 'ok'
      detail.value = await r.text()
    } else {
      health.value = 'fail'
      detail.value = `HTTP ${r.status}: ${await r.text()}`
    }
  } catch (e) {
    health.value = 'fail'
    detail.value = String(e)
  }
}

onMounted(check)
</script>

<template>
  <div class="space-y-6">
    <h2 class="text-2xl font-semibold tracking-tight">Dashboard</h2>

    <section class="rounded-md border border-zinc-800 bg-zinc-900/50 p-6">
      <h3 class="text-sm font-medium uppercase tracking-wide text-zinc-400 mb-3">
        Backend health
      </h3>
      <div class="flex items-center gap-3">
        <span
          class="inline-block w-2 h-2 rounded-full"
          :class="{
            'bg-zinc-500 animate-pulse': health === 'pending',
            'bg-emerald-500': health === 'ok',
            'bg-red-500': health === 'fail',
          }"
        ></span>
        <span class="font-mono text-sm">
          {{ health === 'pending' ? 'checking...' : detail }}
        </span>
        <button
          class="ml-auto text-xs px-2 py-1 rounded border border-zinc-700 hover:bg-zinc-800"
          @click="check"
        >
          refresh
        </button>
      </div>
    </section>

    <section class="rounded-md border border-zinc-800 bg-zinc-900/50 p-6">
      <h3 class="text-sm font-medium uppercase tracking-wide text-zinc-400 mb-3">
        Phase 1 scope
      </h3>
      <p class="text-sm text-zinc-300 leading-relaxed max-w-2xl">
        Foundation scaffold: Vue 3 + Vite + Tailwind 4 + vue-router, served
        by a Go binary with embed.FS. Phase 2 adds the real
        <code class="font-mono text-zinc-400">/api/dashboard</code> endpoint
        with substrate state (soak status, in-flight work_items, recent
        errors). Phases 3–7 add studies browse, work_items, sessions,
        watchman, bridge state, graph view, and new-work form.
      </p>
    </section>
  </div>
</template>
