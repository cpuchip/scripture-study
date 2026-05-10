<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRoute, useRouter, RouterLink } from 'vue-router'
import { api, type StudiesListResp, type SearchResp } from '@/api'

const route = useRoute()
const router = useRouter()

const list = ref<StudiesListResp | null>(null)
const search = ref<SearchResp | null>(null)
const error = ref<string>('')
const loading = ref(false)

const query = ref<string>(String(route.query.q ?? ''))
const kind = ref<string>(String(route.query.kind ?? ''))

async function load() {
  loading.value = true
  error.value = ''
  try {
    if (query.value.trim()) {
      search.value = await api.studiesSearch(query.value.trim(), { limit: 50 })
      list.value = null
    } else {
      list.value = await api.studiesList({
        kind: kind.value || undefined,
        limit: 100,
      })
      search.value = null
    }
  } catch (e) {
    error.value = String(e)
  } finally {
    loading.value = false
  }
}

function submit() {
  router.replace({
    path: '/studies',
    query: query.value.trim()
      ? { q: query.value.trim() }
      : kind.value
      ? { kind: kind.value }
      : {},
  })
  load()
}

onMounted(load)
watch(
  () => route.query,
  (q) => {
    query.value = String(q.q ?? '')
    kind.value = String(q.kind ?? '')
    load()
  },
)

function fmtDate(s?: string) {
  if (!s) return ''
  return new Date(s).toLocaleDateString()
}
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-baseline justify-between">
      <h2 class="text-2xl font-semibold tracking-tight">Studies</h2>
      <span v-if="list" class="text-xs text-zinc-500">
        {{ list.total.toLocaleString() }} total
      </span>
      <span v-else-if="search" class="text-xs text-zinc-500">
        {{ search.hits.length }} hits for "{{ search.query }}"
      </span>
    </div>

    <form
      class="flex gap-2"
      @submit.prevent="submit"
    >
      <input
        v-model="query"
        type="text"
        placeholder="search studies (FTS)…"
        class="flex-1 px-3 py-2 rounded border border-zinc-700 bg-zinc-900 text-sm focus:border-zinc-500 focus:outline-none"
      />
      <select
        v-model="kind"
        class="px-3 py-2 rounded border border-zinc-700 bg-zinc-900 text-sm"
      >
        <option value="">all kinds</option>
        <option value="study">study</option>
        <option value="doc">doc</option>
        <option value="proposal">proposal</option>
        <option value="phase-doc">phase-doc</option>
        <option value="journal">journal</option>
      </select>
      <button
        type="submit"
        class="px-4 py-2 rounded border border-zinc-700 hover:bg-zinc-800 text-sm"
      >
        go
      </button>
    </form>

    <p v-if="loading" class="text-sm text-zinc-400">loading…</p>
    <p v-else-if="error" class="text-sm text-red-400">{{ error }}</p>

    <!-- Search results -->
    <ul
      v-if="search"
      class="rounded-md border border-zinc-800 bg-zinc-900/50 divide-y divide-zinc-800"
    >
      <li v-for="h in search.hits" :key="h.slug" class="p-4 hover:bg-zinc-900">
        <RouterLink
          :to="`/studies/${encodeURIComponent(h.slug)}`"
          class="block"
        >
          <div class="flex items-baseline gap-2">
            <span class="font-medium text-zinc-100">{{ h.title || h.slug }}</span>
            <span v-if="h.kind" class="text-xs text-zinc-500">{{ h.kind }}</span>
            <span v-if="h.score" class="ml-auto text-xs text-zinc-500 tabular-nums">
              score {{ h.score.toFixed(3) }}
            </span>
          </div>
          <div
            v-if="h.snippet"
            class="text-xs text-zinc-400 mt-1"
            v-html="h.snippet"
          ></div>
          <div class="text-xs text-zinc-500 mt-1 font-mono">{{ h.slug }}</div>
        </RouterLink>
      </li>
      <li v-if="search.hits.length === 0" class="p-6 text-sm text-zinc-500 text-center">
        no hits
      </li>
    </ul>

    <!-- List view -->
    <div
      v-else-if="list"
      class="rounded-md border border-zinc-800 bg-zinc-900/50 overflow-hidden"
    >
      <table class="w-full text-sm">
        <thead class="text-zinc-500 text-xs uppercase tracking-wide">
          <tr>
            <th class="text-left px-4 py-2 font-medium">Title</th>
            <th class="text-left px-4 py-2 font-medium">Kind</th>
            <th class="text-right px-4 py-2 font-medium">Body</th>
            <th class="text-right px-4 py-2 font-medium">Updated</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="s in list.items"
            :key="s.slug"
            class="border-t border-zinc-800/50 hover:bg-zinc-900"
          >
            <td class="px-4 py-2">
              <RouterLink
                :to="`/studies/${encodeURIComponent(s.slug)}`"
                class="text-zinc-100 hover:text-white"
              >
                {{ s.title || s.slug }}
              </RouterLink>
              <div class="text-xs text-zinc-500 font-mono">{{ s.slug }}</div>
            </td>
            <td class="px-4 py-2 text-zinc-400">{{ s.kind }}</td>
            <td class="px-4 py-2 text-right tabular-nums text-zinc-500">
              {{ s.body_chars.toLocaleString() }}
            </td>
            <td class="px-4 py-2 text-right text-zinc-500 text-xs">
              {{ fmtDate(s.updated_at) }}
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
