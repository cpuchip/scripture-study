<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api, type CovenantRow } from '@/api'

const covenant = ref<CovenantRow | null>(null)
const error = ref('')
const loading = ref(true)
const showTeaching = ref(false)

async function load() {
  loading.value = true
  error.value = ''
  try {
    covenant.value = await api.covenantActive('global')
  } catch (e) {
    error.value = String(e)
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<template>
  <div class="space-y-6 max-w-5xl">
    <header class="border-b border-zinc-800 pb-3">
      <h2 class="text-2xl font-semibold tracking-tight">Covenant</h2>
      <p class="text-sm text-zinc-400 mt-1">
        Bilateral commitments. Edited in
        <code class="font-mono text-zinc-300">.spec/covenant.yaml</code>;
        seeded into the substrate via the git pre-commit hook (Phase C).
      </p>
    </header>

    <p v-if="loading" class="text-sm text-zinc-400">loading…</p>
    <p v-else-if="error" class="text-sm text-red-400">{{ error }}</p>

    <template v-if="covenant">
      <div class="text-xs text-zinc-500 flex gap-4 font-mono">
        <span>scope: {{ covenant.scope }}</span>
        <span>ratified by: {{ covenant.ratified_by }}</span>
        <span v-if="covenant.source_file">source: {{ covenant.source_file }}</span>
      </div>

      <div class="grid md:grid-cols-2 gap-4">
        <section class="rounded-md border border-zinc-800 bg-zinc-900/50 p-4">
          <h3 class="text-sm font-semibold uppercase tracking-wide text-zinc-400 mb-3">
            The human commits to
          </h3>
          <ul class="space-y-3 text-sm">
            <li v-for="c in covenant.human_commits_to" :key="c.key">
              <div class="font-mono text-zinc-200">{{ c.key }}</div>
              <p class="text-zinc-300 mt-1">{{ c.description }}</p>
              <details v-if="c.why" class="mt-1">
                <summary class="text-xs text-zinc-500 hover:text-zinc-300 cursor-pointer">why</summary>
                <p class="text-xs text-zinc-400 mt-1">{{ c.why }}</p>
              </details>
            </li>
          </ul>
        </section>

        <section class="rounded-md border border-zinc-800 bg-zinc-900/50 p-4">
          <h3 class="text-sm font-semibold uppercase tracking-wide text-zinc-400 mb-3">
            The agent commits to
          </h3>
          <ul class="space-y-3 text-sm">
            <li v-for="c in covenant.agent_commits_to" :key="c.key">
              <div class="font-mono text-zinc-200">{{ c.key }}</div>
              <p class="text-zinc-300 mt-1">{{ c.description }}</p>
              <details v-if="c.why" class="mt-1">
                <summary class="text-xs text-zinc-500 hover:text-zinc-300 cursor-pointer">why</summary>
                <p class="text-xs text-zinc-400 mt-1">{{ c.why }}</p>
              </details>
            </li>
          </ul>
        </section>
      </div>

      <section
        v-if="covenant.when_broken || covenant.recovery"
        class="rounded-md border border-zinc-800 bg-zinc-900/50 p-4 space-y-3"
      >
        <h3 class="text-sm font-semibold uppercase tracking-wide text-zinc-400">
          When broken
        </h3>
        <p v-if="covenant.when_broken" class="text-sm text-zinc-300">{{ covenant.when_broken }}</p>
        <div v-if="covenant.recovery">
          <div class="text-xs uppercase tracking-wide text-zinc-500 mb-1">Recovery</div>
          <p class="text-sm text-zinc-300">{{ covenant.recovery }}</p>
        </div>
      </section>

      <section
        v-if="covenant.council_moment"
        class="rounded-md border border-zinc-800 bg-zinc-900/50 p-4"
      >
        <h3 class="text-sm font-semibold uppercase tracking-wide text-zinc-400 mb-2">
          Council moment
        </h3>
        <p class="text-sm text-zinc-300">{{ covenant.council_moment }}</p>
      </section>

      <section
        v-if="covenant.teaching_extension"
        class="rounded-md border border-zinc-800 bg-zinc-900/50 p-4"
      >
        <button
          class="text-sm font-semibold uppercase tracking-wide text-zinc-400 hover:text-zinc-200 cursor-pointer"
          @click="showTeaching = !showTeaching"
        >
          Teaching extension {{ showTeaching ? '▾' : '▸' }}
        </button>
        <pre v-if="showTeaching" class="mt-3 text-xs font-mono text-zinc-300 whitespace-pre-wrap">{{ JSON.stringify(covenant.teaching_extension, null, 2) }}</pre>
      </section>
    </template>
  </div>
</template>
