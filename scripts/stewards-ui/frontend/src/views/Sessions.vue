<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, computed } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import { api, type SessionDetail, type MessageRow, type ChatDispatch, type SessionListItem } from '@/api'

const route = useRoute()
const session = ref<SessionDetail | null>(null)
const error = ref<string>('')
const loading = ref(false)
const tab = ref<'timeline' | 'dispatches'>('timeline')

// List-mode state. Populated when no :sid is in the route — shows
// active sessions the user can click into. Refreshes every 5s
// (cheap query; same cadence as the dashboard).
const activeSessions = ref<SessionListItem[]>([])
const listError = ref<string>('')
const listLoading = ref(false)
let listTimer: number | undefined

async function loadList() {
  listLoading.value = true
  listError.value = ''
  try {
    const r = await api.sessionsList()
    activeSessions.value = r.sessions
  } catch (e) {
    listError.value = String(e)
  } finally {
    listLoading.value = false
  }
}

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
onMounted(() => {
  if (sidFromRoute.value) {
    load(sidFromRoute.value)
  } else {
    loadList()
    listTimer = window.setInterval(loadList, 5000)
  }
})
onUnmounted(() => {
  if (listTimer) window.clearInterval(listTimer)
})
watch(sidFromRoute, (v) => {
  if (v) {
    if (listTimer) { window.clearInterval(listTimer); listTimer = undefined }
    load(v)
  } else {
    loadList()
    if (!listTimer) listTimer = window.setInterval(loadList, 5000)
  }
})

function fmtRelative(s?: string) {
  if (!s) return ''
  const d = new Date(s)
  if (isNaN(d.getTime())) return s
  const diffMs = Date.now() - d.getTime()
  const sec = Math.floor(diffMs / 1000)
  if (sec < 60) return `${sec}s ago`
  const min = Math.floor(sec / 60)
  if (min < 60) return `${min}m ago`
  const hr = Math.floor(min / 60)
  if (hr < 24) return `${hr}h ago`
  return `${Math.floor(hr / 24)}d ago`
}

function roleClass(role: string) {
  switch (role) {
    case 'system':    return 'bg-zinc-900 border-zinc-700'
    case 'user':      return 'bg-blue-950/30 border-blue-900/40'
    case 'assistant': return 'bg-emerald-950/20 border-emerald-900/30'
    case 'tool':      return 'bg-amber-950/20 border-amber-900/30'
    default:          return 'bg-zinc-900 border-zinc-700'
  }
}

// First dispatch's system prompt + tools — surfaces the otherwise-
// invisible "what the model saw" at the start of a session.
const firstDispatch = computed<ChatDispatch | null>(() =>
  session.value?.dispatches?.[0] ?? null,
)

// Parse tool_calls JSON safely
function parseToolCalls(raw: unknown): Array<{ id: string; name: string; args: unknown }> {
  if (!raw) return []
  const arr = Array.isArray(raw) ? raw : []
  return arr.map((tc: any) => {
    let parsedArgs: unknown = tc?.function?.arguments
    if (typeof parsedArgs === 'string') {
      try { parsedArgs = JSON.parse(parsedArgs) } catch { /* keep string */ }
    }
    return {
      id: tc?.id ?? '',
      name: tc?.function?.name ?? tc?.name ?? '?',
      args: parsedArgs,
    }
  })
}

// Try to parse a tool reply's content as JSON for prettier rendering
function parseToolContent(content: string): unknown {
  if (!content) return content
  const t = content.trim()
  if (!(t.startsWith('{') || t.startsWith('['))) return content
  try { return JSON.parse(t) } catch { return content }
}

function isJSON(v: unknown): boolean {
  return typeof v === 'object' && v !== null
}

function summarizeMessage(m: MessageRow): string {
  if (m.role === 'tool') {
    const parsed = parseToolContent(m.content)
    if (isJSON(parsed)) {
      const keys = Object.keys(parsed as object).slice(0, 4)
      return keys.length ? `{ ${keys.join(', ')} }` : '(empty)'
    }
    return m.content.slice(0, 80) + (m.content.length > 80 ? '…' : '')
  }
  return m.content.slice(0, 120) + (m.content.length > 120 ? '…' : '')
}

function fmtDate(s?: string) {
  if (!s) return ''
  return new Date(s).toLocaleString()
}
</script>

<template>
  <div class="space-y-6">
    <!-- List mode: no session ID in the route -->
    <template v-if="!sidFromRoute">
      <header class="flex items-baseline justify-between border-b border-zinc-800 pb-4">
        <h2 class="text-2xl font-semibold tracking-tight">Active sessions</h2>
        <div class="text-xs text-zinc-500 flex items-center gap-3">
          <span v-if="listLoading" class="text-zinc-400">refreshing…</span>
          <span v-else-if="listError" class="text-red-400">{{ listError }}</span>
          <span v-else>{{ activeSessions.length }} session(s) · auto-refresh 5s</span>
          <button
            class="text-xs px-2 py-1 rounded border border-zinc-700 hover:bg-zinc-800"
            @click="loadList"
          >
            refresh
          </button>
        </div>
      </header>

      <p class="text-sm text-zinc-500">
        Sessions linked to in-progress work_items, plus any sessions active in the last hour.
        Click into one for the full message timeline + dispatch payloads.
      </p>

      <section
        v-if="activeSessions.length > 0"
        class="rounded-md border border-zinc-800 bg-zinc-900/50 overflow-hidden"
      >
        <table class="w-full text-sm">
          <thead class="text-zinc-500 text-xs uppercase tracking-wide">
            <tr>
              <th class="text-left px-4 py-2 font-medium">Session</th>
              <th class="text-left px-4 py-2 font-medium">Work item</th>
              <th class="text-left px-4 py-2 font-medium">Stage</th>
              <th class="text-left px-4 py-2 font-medium">Status</th>
              <th class="text-right px-4 py-2 font-medium">Messages</th>
              <th class="text-right px-4 py-2 font-medium">Cost</th>
              <th class="text-right px-4 py-2 font-medium">Active</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="s in activeSessions"
              :key="s.session_id"
              class="border-t border-zinc-800/50 hover:bg-zinc-900 cursor-pointer"
              @click="$router.push(`/sessions/${encodeURIComponent(s.session_id)}`)"
            >
              <td class="px-4 py-2 font-mono text-xs">
                <RouterLink
                  :to="`/sessions/${encodeURIComponent(s.session_id)}`"
                  class="text-zinc-200 hover:text-white"
                  @click.stop
                >
                  {{ s.session_id }}
                </RouterLink>
              </td>
              <td class="px-4 py-2">
                <RouterLink
                  v-if="s.work_item_id"
                  :to="`/work-items/${s.work_item_id}`"
                  class="text-zinc-300 hover:text-white font-mono text-xs"
                  @click.stop
                >
                  {{ s.work_item_slug || s.work_item_id.slice(0, 8) }}
                </RouterLink>
                <span v-else class="text-zinc-600 text-xs">—</span>
              </td>
              <td class="px-4 py-2 text-zinc-300">{{ s.current_stage || '—' }}</td>
              <td class="px-4 py-2">
                <span
                  v-if="s.work_item_status"
                  class="inline-block px-2 py-0.5 rounded text-xs"
                  :class="{
                    'bg-emerald-900/40 text-emerald-300': s.work_item_active,
                    'bg-zinc-800 text-zinc-400': !s.work_item_active,
                  }"
                >
                  {{ s.work_item_status }}
                </span>
                <span v-else class="text-zinc-600 text-xs">—</span>
              </td>
              <td class="px-4 py-2 text-right tabular-nums text-zinc-400">
                {{ s.message_count.toLocaleString() }}
                <span class="text-zinc-600">·</span>
                <span class="text-zinc-500">{{ s.assistant_count }} asst</span>
              </td>
              <td class="px-4 py-2 text-right tabular-nums text-zinc-400">
                {{ s.cost_total > 0 ? `$${s.cost_total.toFixed(4)}` : '—' }}
              </td>
              <td class="px-4 py-2 text-right text-zinc-500 text-xs">
                {{ fmtRelative(s.last_active_at) }}
              </td>
            </tr>
          </tbody>
        </table>
      </section>

      <div v-else-if="!listLoading" class="text-sm text-zinc-500">
        No active sessions. The substrate is quiet.
        <RouterLink to="/work-items" class="text-zinc-300 hover:text-white">
          Browse all work items →
        </RouterLink>
      </div>
    </template>

    <template v-else>
      <header class="border-b border-zinc-800 pb-4">
        <h2 class="text-2xl font-semibold tracking-tight">Session</h2>
        <div class="text-xs text-zinc-500 mt-2 font-mono">{{ sidFromRoute }}</div>
        <div v-if="session" class="text-xs text-zinc-400 mt-1 flex gap-4 flex-wrap">
          <span>{{ session.messages.length }} messages</span>
          <span>{{ session.dispatches.length }} dispatches</span>
          <span>{{ session.tokens_in.toLocaleString() }} in / {{ session.tokens_out.toLocaleString() }} out</span>
        </div>
      </header>

      <p v-if="loading" class="text-sm text-zinc-400">loading…</p>
      <p v-else-if="error" class="text-sm text-red-400">{{ error }}</p>

      <template v-if="session">
        <!-- First-dispatch system prompt + tools — the invisible context -->
        <section
          v-if="firstDispatch && firstDispatch.system_prompt"
          class="rounded-md border border-zinc-700 bg-zinc-900 p-4"
        >
          <div class="text-xs uppercase tracking-wide text-zinc-500 mb-2 flex items-baseline gap-3">
            <span>System prompt</span>
            <span class="font-mono text-zinc-600">
              {{ firstDispatch.agent_family }} · {{ firstDispatch.model }} · {{ firstDispatch.provider }}
            </span>
          </div>
          <details>
            <summary class="cursor-pointer text-sm text-zinc-400 hover:text-zinc-200">
              {{ firstDispatch.system_prompt.length.toLocaleString() }} chars — expand
            </summary>
            <pre class="text-xs text-zinc-300 whitespace-pre-wrap mt-2 max-h-96 overflow-auto font-mono">{{ firstDispatch.system_prompt }}</pre>
          </details>
          <details v-if="firstDispatch.tools" class="mt-2">
            <summary class="cursor-pointer text-sm text-zinc-400 hover:text-zinc-200">
              tools available
              <span v-if="Array.isArray(firstDispatch.tools)" class="text-zinc-600">
                ({{ firstDispatch.tools.length }})
              </span>
            </summary>
            <ul v-if="Array.isArray(firstDispatch.tools)" class="text-xs mt-2 space-y-1">
              <li
                v-for="t in firstDispatch.tools as any[]"
                :key="t.function?.name"
                class="font-mono text-zinc-400"
              >
                {{ t.function?.name }}
                <span class="text-zinc-600">— {{ (t.function?.description ?? '').slice(0, 100) }}</span>
              </li>
            </ul>
          </details>
        </section>

        <!-- Tab switcher -->
        <div class="flex gap-2 border-b border-zinc-800">
          <button
            class="px-3 py-2 text-sm border-b-2 -mb-px"
            :class="tab === 'timeline'
              ? 'border-emerald-500 text-zinc-100'
              : 'border-transparent text-zinc-400 hover:text-zinc-200'"
            @click="tab = 'timeline'"
          >
            Timeline ({{ session.messages.length }})
          </button>
          <button
            class="px-3 py-2 text-sm border-b-2 -mb-px"
            :class="tab === 'dispatches'
              ? 'border-emerald-500 text-zinc-100'
              : 'border-transparent text-zinc-400 hover:text-zinc-200'"
            @click="tab = 'dispatches'"
          >
            Dispatches ({{ session.dispatches.length }})
          </button>
        </div>

        <!-- Timeline -->
        <ul v-if="tab === 'timeline'" class="space-y-3">
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
              <span v-if="m.tool_call_id" class="text-xs text-amber-400 font-mono">
                ↳ {{ m.tool_call_id }}
              </span>
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

            <!-- Tool reply: render parsed JSON if possible -->
            <template v-if="m.role === 'tool'">
              <div class="text-xs text-zinc-500 mb-1">{{ summarizeMessage(m) }}</div>
              <details>
                <summary class="cursor-pointer text-xs text-zinc-500 hover:text-zinc-300">
                  full reply ({{ m.content.length.toLocaleString() }} chars)
                </summary>
                <pre
                  class="text-xs text-zinc-300 whitespace-pre-wrap mt-2 max-h-96 overflow-auto font-mono"
                >{{ isJSON(parseToolContent(m.content)) ? JSON.stringify(parseToolContent(m.content), null, 2) : m.content }}</pre>
              </details>
            </template>

            <!-- Assistant / user / system: prose -->
            <pre
              v-else
              class="whitespace-pre-wrap text-sm text-zinc-200 font-sans leading-relaxed"
            >{{ m.content }}</pre>

            <!-- tool_calls drill-down: parsed JSON + per-call cards -->
            <div v-if="m.tool_calls" class="mt-3 space-y-2">
              <div class="text-xs uppercase tracking-wide text-zinc-500">
                tool calls ({{ parseToolCalls(m.tool_calls).length }})
              </div>
              <details
                v-for="tc in parseToolCalls(m.tool_calls)"
                :key="tc.id"
                class="rounded border border-amber-900/40 bg-amber-950/20 p-2"
              >
                <summary class="cursor-pointer text-sm flex items-baseline gap-2">
                  <span class="font-mono text-amber-300">{{ tc.name }}</span>
                  <span class="font-mono text-xs text-zinc-500">{{ tc.id }}</span>
                </summary>
                <pre class="text-xs text-zinc-300 mt-2 whitespace-pre-wrap font-mono">{{ isJSON(tc.args) ? JSON.stringify(tc.args, null, 2) : tc.args }}</pre>
              </details>
            </div>
          </li>
          <li
            v-if="session.messages.length === 0"
            class="text-sm text-zinc-500 text-center py-8"
          >
            No persisted messages. The Dispatches tab may still show what was sent.
          </li>
        </ul>

        <!-- Dispatches: each chat work_queue row that touched this session -->
        <ul v-else-if="tab === 'dispatches'" class="space-y-3">
          <li
            v-for="d in session.dispatches"
            :key="d.work_id"
            class="rounded-md border border-zinc-700 bg-zinc-900 p-4"
          >
            <div class="flex items-baseline gap-3 mb-2 flex-wrap">
              <span class="font-mono text-xs text-zinc-500">#{{ d.work_id }}</span>
              <span class="text-xs uppercase tracking-wide font-semibold">dispatch</span>
              <span class="font-mono text-xs text-zinc-400">{{ d.model }}</span>
              <span class="text-xs text-zinc-500">via {{ d.provider }}</span>
              <span
                class="text-xs px-2 py-0.5 rounded"
                :class="d.status === 'done' ? 'bg-emerald-900/30 text-emerald-300' : 'bg-zinc-800 text-zinc-300'"
              >{{ d.status }}</span>
              <span class="ml-auto text-xs text-zinc-500">{{ fmtDate(d.created_at) }}</span>
            </div>
            <div class="text-xs text-zinc-400">
              {{ d.messages_count }} messages sent (system + history + current)
            </div>
            <details class="mt-2">
              <summary class="cursor-pointer text-xs text-zinc-500 hover:text-zinc-300">
                full request body
              </summary>
              <pre
                class="text-xs text-zinc-300 mt-2 whitespace-pre-wrap font-mono max-h-[600px] overflow-auto bg-zinc-950 p-2 rounded"
              >{{ JSON.stringify(d.body_messages, null, 2) }}</pre>
            </details>
          </li>
          <li
            v-if="session.dispatches.length === 0"
            class="text-sm text-zinc-500 text-center py-8"
          >
            No chat dispatches recorded for this session.
          </li>
        </ul>
      </template>
    </template>
  </div>
</template>
