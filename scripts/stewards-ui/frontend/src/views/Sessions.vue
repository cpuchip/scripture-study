<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import { api, type SessionDetail } from '@/api'

const route = useRoute()
const session = ref<SessionDetail | null>(null)
const error = ref<string>('')
const loading = ref(false)

async function load(sid: string) {
  loading.value = true
  error.value = ''
  session.value = null
  try {
    session.value = await api.sessionGet(sid)
  } catch (e) {
    error.value = String(e)
  } finally {
    loading.value = false
  }
}

const sidFromRoute = computed(() => String(route.params.sid ?? ''))
onMounted(() => sidFromRoute.value && load(sidFromRoute.value))
watch(sidFromRoute, (v) => v && load(v))

function roleClass(role: string) {
  switch (role) {
    case 'system':    return 'bg-zinc-900 border-zinc-700'
    case 'user':      return 'bg-blue-950/30 border-blue-900/40'
    case 'assistant': return 'bg-emerald-950/20 border-emerald-900/30'
    case 'tool':      return 'bg-amber-950/20 border-amber-900/30'
    default:          return 'bg-zinc-900 border-zinc-700'
  }
}
</script>

<template>
  <div class="space-y-6">
    <p v-if="!sidFromRoute" class="text-sm text-zinc-500">
      Open a work item to drill into its sessions.
      <RouterLink to="/work-items" class="text-zinc-300 hover:text-white">
        Browse work items →
      </RouterLink>
    </p>

    <template v-else>
      <header class="border-b border-zinc-800 pb-4">
        <h2 class="text-2xl font-semibold tracking-tight">Session</h2>
        <div class="text-xs text-zinc-500 mt-2 font-mono">{{ sidFromRoute }}</div>
        <div v-if="session" class="text-xs text-zinc-400 mt-1">
          {{ session.messages.length }} messages ·
          {{ session.tokens_in.toLocaleString() }} in /
          {{ session.tokens_out.toLocaleString() }} out
        </div>
      </header>

      <p v-if="loading" class="text-sm text-zinc-400">loading…</p>
      <p v-else-if="error" class="text-sm text-red-400">{{ error }}</p>

      <ul v-if="session" class="space-y-3">
        <li
          v-for="m in session.messages"
          :key="m.id"
          class="rounded-md border p-4"
          :class="roleClass(m.role)"
        >
          <div class="flex items-baseline gap-3 mb-2">
            <span class="font-mono text-xs text-zinc-500">#{{ m.id }}</span>
            <span class="text-xs uppercase tracking-wide font-semibold">{{ m.role }}</span>
            <span v-if="m.model" class="text-xs text-zinc-500 font-mono">{{ m.model }}</span>
            <span v-if="m.finish_reason" class="text-xs text-zinc-500">
              finish: {{ m.finish_reason }}
            </span>
            <span class="ml-auto text-xs text-zinc-500 tabular-nums">
              <template v-if="m.tokens_in || m.tokens_out">
                {{ (m.tokens_in ?? 0).toLocaleString() }} / {{ (m.tokens_out ?? 0).toLocaleString() }}
                <span v-if="m.reasoning_tokens">+{{ m.reasoning_tokens.toLocaleString() }} reasoning</span>
              </template>
            </span>
          </div>
          <pre class="whitespace-pre-wrap text-sm text-zinc-200 font-sans leading-relaxed">{{ m.content }}</pre>
          <details v-if="m.tool_calls" class="mt-2">
            <summary class="text-xs text-zinc-500 cursor-pointer hover:text-zinc-300">
              tool_calls
            </summary>
            <pre class="text-xs font-mono text-zinc-400 mt-1 whitespace-pre-wrap">{{ JSON.stringify(m.tool_calls, null, 2) }}</pre>
          </details>
        </li>
      </ul>
    </template>
  </div>
</template>
