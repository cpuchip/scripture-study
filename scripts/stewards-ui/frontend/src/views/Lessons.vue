<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { RouterLink } from 'vue-router'
import { api, type LessonRow } from '@/api'

const lessons = ref<LessonRow[]>([])
const loading = ref(true)
const error = ref('')
const filterKind = ref<string>('')
const filterRatified = ref<'true' | 'false' | ''>('false')

const ratifying = ref<Record<number, boolean>>({})
const ratifiedBy = ref('michael')

async function load() {
  loading.value = true
  error.value = ''
  try {
    const r = await api.lessonsList({
      kind: filterKind.value || undefined,
      ratified: filterRatified.value || undefined,
      limit: 200,
    })
    lessons.value = r.items
  } catch (e) {
    error.value = String(e)
  } finally {
    loading.value = false
  }
}

async function ratify(l: LessonRow, promoted_to?: string) {
  ratifying.value[l.id] = true
  try {
    await api.lessonRatify({
      id: l.id,
      ratified_by: ratifiedBy.value || 'human',
      promoted_to,
    })
    await load()
  } catch (e) {
    error.value = String(e)
  } finally {
    ratifying.value[l.id] = false
  }
}

function fmtDate(s?: string) {
  if (!s) return ''
  return new Date(s).toLocaleString()
}

function kindTone(kind: string): string {
  switch (kind) {
    case 'principle':
      return 'bg-purple-900/30 text-purple-300 border-purple-700/50'
    case 'decision':
      return 'bg-blue-900/30 text-blue-300 border-blue-700/50'
    case 'lesson':
      return 'bg-emerald-900/30 text-emerald-300 border-emerald-700/50'
    case 'sabbath_reflection':
      return 'bg-amber-900/30 text-amber-300 border-amber-700/50'
    default:
      return 'bg-zinc-800 text-zinc-300 border-zinc-700'
  }
}

const groupedByKind = computed(() => {
  const out: Record<string, LessonRow[]> = {}
  for (const l of lessons.value) {
    if (!out[l.kind]) out[l.kind] = []
    out[l.kind]!.push(l)
  }
  return out
})

onMounted(load)
</script>

<template>
  <div class="space-y-6 max-w-4xl">
    <header class="border-b border-zinc-800 pb-3">
      <h2 class="text-2xl font-semibold tracking-tight">Lessons</h2>
      <p class="text-sm text-zinc-400 mt-1">
        Atonement and Sabbath outputs. Unratified rows wait for a human to either approve them
        (substrate-only) or approve and promote them to <code class="font-mono text-zinc-300">.mind/principles.md</code>
        / <code class="font-mono text-zinc-300">.mind/decisions.md</code>.
      </p>
    </header>

    <div class="flex items-center gap-3 text-sm">
      <label class="text-zinc-400">Kind:</label>
      <select
        v-model="filterKind"
        @change="load"
        class="px-3 py-1 rounded border border-zinc-700 bg-zinc-900 text-sm focus:border-zinc-500 focus:outline-none"
      >
        <option value="">all</option>
        <option value="principle">principle</option>
        <option value="decision">decision</option>
        <option value="lesson">lesson</option>
        <option value="sabbath_reflection">sabbath_reflection</option>
      </select>

      <label class="text-zinc-400 ml-3">Ratified:</label>
      <select
        v-model="filterRatified"
        @change="load"
        class="px-3 py-1 rounded border border-zinc-700 bg-zinc-900 text-sm focus:border-zinc-500 focus:outline-none"
      >
        <option value="false">unratified only</option>
        <option value="true">ratified only</option>
        <option value="">all</option>
      </select>

      <span class="text-zinc-500 text-xs ml-auto">{{ lessons.length }} entries</span>
    </div>

    <div class="flex items-center gap-3 text-sm">
      <label class="text-zinc-400">Ratify as:</label>
      <input
        v-model="ratifiedBy"
        type="text"
        class="px-3 py-1 rounded border border-zinc-700 bg-zinc-900 text-sm focus:border-zinc-500 focus:outline-none"
      />
    </div>

    <p v-if="loading" class="text-sm text-zinc-400">loading…</p>
    <p v-else-if="error" class="text-sm text-red-400">{{ error }}</p>

    <p v-else-if="lessons.length === 0" class="text-sm text-zinc-500">
      No matching lessons. Try changing the filters above.
    </p>

    <div v-else class="space-y-6">
      <section v-for="(items, kind) in groupedByKind" :key="kind">
        <h3 class="text-sm uppercase tracking-wide text-zinc-500 mb-2">
          {{ kind }} <span class="text-zinc-700">({{ items.length }})</span>
        </h3>
        <ul class="space-y-2">
          <li
            v-for="l in items"
            :key="l.id"
            class="rounded-md border border-zinc-800 bg-zinc-900/50 p-3"
          >
            <div class="flex items-baseline gap-3 flex-wrap mb-2">
              <span
                class="text-xs px-2 py-0.5 rounded border font-mono"
                :class="kindTone(l.kind)"
              >{{ l.kind }}</span>
              <RouterLink
                v-if="l.work_item_id"
                :to="`/work-items/${l.work_item_id}`"
                class="text-xs text-zinc-400 hover:text-zinc-200 font-mono"
              >{{ l.work_item_slug || l.work_item_id }}</RouterLink>
              <span v-if="l.pipeline_family" class="text-xs text-zinc-500 font-mono">{{ l.pipeline_family }} / {{ l.current_stage }}</span>
              <span v-if="l.ratified_at" class="text-xs text-emerald-500 ml-auto">
                ✓ ratified by {{ l.ratified_by }} {{ fmtDate(l.ratified_at) }}
                <span v-if="l.promoted_to" class="text-zinc-500"> → {{ l.promoted_to }}</span>
              </span>
              <span v-else class="text-xs text-zinc-500 ml-auto">{{ fmtDate(l.at) }}</span>
            </div>
            <p class="text-sm text-zinc-200 leading-relaxed whitespace-pre-wrap">{{ l.content }}</p>

            <div v-if="!l.ratified_at" class="flex items-center gap-2 mt-3">
              <button
                class="px-3 py-1 text-xs rounded bg-emerald-700 hover:bg-emerald-600 text-white disabled:opacity-50"
                :disabled="ratifying[l.id]"
                @click="ratify(l)"
              >Approve</button>
              <button
                v-if="l.kind === 'principle'"
                class="px-3 py-1 text-xs rounded border border-purple-700 hover:bg-purple-900/30 text-purple-300 disabled:opacity-50"
                :disabled="ratifying[l.id]"
                @click="ratify(l, '.mind/principles.md')"
              >Approve &amp; promote → .mind/principles.md</button>
              <button
                v-if="l.kind === 'decision'"
                class="px-3 py-1 text-xs rounded border border-blue-700 hover:bg-blue-900/30 text-blue-300 disabled:opacity-50"
                :disabled="ratifying[l.id]"
                @click="ratify(l, '.mind/decisions.md')"
              >Approve &amp; promote → .mind/decisions.md</button>
            </div>
          </li>
        </ul>
      </section>
    </div>
  </div>
</template>
