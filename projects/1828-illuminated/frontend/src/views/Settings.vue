<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  llmSettings,
  applyPreset,
  canMintSession,
  dismissMigrationBanner,
  migrationNeeded,
  type ProviderID,
} from '@/composables/useLLMSettings'
import { startSession, endSession, refreshSession, sessionActive } from '@/composables/useLLMSession'

interface PresetMeta {
  id: ProviderID
  label: string
  detail: string
  needsKey: boolean
  baseUrlEditable: boolean
}

const presets: PresetMeta[] = [
  { id: 'opencode-go', label: 'OpenCode Go', detail: 'Bring your own opencode-go gateway URL + key', needsKey: true, baseUrlEditable: true },
  { id: 'opencode-zen', label: 'OpenCode Zen', detail: 'Hosted opencode variant — bring your gateway URL + key', needsKey: true, baseUrlEditable: true },
  { id: 'openai', label: 'OpenAI', detail: 'api.openai.com — bring your sk-… key', needsKey: true, baseUrlEditable: false },
  { id: 'openrouter', label: 'OpenRouter', detail: 'openrouter.ai — bring your key', needsKey: true, baseUrlEditable: false },
  { id: 'custom', label: 'Custom', detail: 'Any OpenAI-compatible /v1/chat/completions endpoint', needsKey: true, baseUrlEditable: true },
]

const apiKeyInput = ref<string>('')   // NEVER reactive-bound to llmSettings; lives in this component only
const mintError = ref<string>('')
const minting = ref(false)
const signingOut = ref(false)
const checkingSession = ref(true)

const activePreset = computed<PresetMeta | undefined>(() => presets.find(p => p.id === llmSettings.provider))
const expiresHuman = computed(() => {
  if (!llmSettings.expires_at) return ''
  try {
    return new Date(llmSettings.expires_at).toLocaleString()
  } catch {
    return llmSettings.expires_at
  }
})

onMounted(async () => {
  // Authoritative session check — server may have evicted while the tab was idle.
  try { await refreshSession() } finally { checkingSession.value = false }
})

function onPresetChange(id: ProviderID) {
  applyPreset(id)
  mintError.value = ''
}

async function onStartSession() {
  if (!canMintSession()) {
    mintError.value = 'Fill in base URL and model first.'
    return
  }
  if (!apiKeyInput.value.trim()) {
    mintError.value = 'API key is required.'
    return
  }
  minting.value = true
  mintError.value = ''
  const result = await startSession({
    provider: llmSettings.provider,
    base_url: llmSettings.baseUrl,
    api_key: apiKeyInput.value.trim(),
    model: llmSettings.model.trim(),
  })
  minting.value = false
  // Always wipe the input field — the key has either been minted into a
  // session or rejected; either way we don't keep it around.
  apiKeyInput.value = ''
  if ('error' in result) {
    mintError.value = result.message
  }
}

async function onSignOut() {
  signingOut.value = true
  await endSession()
  signingOut.value = false
}

function onDismissMigration() {
  dismissMigrationBanner()
}
</script>

<template>
  <div class="max-w-3xl mx-auto px-6 py-10">
    <header class="mb-8">
      <h1 class="text-3xl font-serif mb-2">Settings</h1>
      <p class="text-stone-600">
        Bring your own LLM provider key. The 1828 backend mints a short-lived session,
        holds your key in memory for the session's lifetime, and proxies render calls
        upstream so your key never leaves the backend or your browser's open form.
      </p>
    </header>

    <!-- v1→v2 migration banner -->
    <div v-if="migrationNeeded" class="def-card mb-6 p-4 bg-amber-50 border-amber-300">
      <div class="flex items-start justify-between gap-3">
        <div>
          <h2 class="font-serif text-base text-amber-900">Settings reshaped</h2>
          <p class="text-sm text-stone-700 mt-1">
            The render flow now goes through the 1828 backend with a session-key model.
            Your old API key was cleared from this browser; re-enter it below and click
            <em>Start session</em> to begin rendering again. Provider, model, and
            preferences were preserved.
          </p>
        </div>
        <button
          @click="onDismissMigration"
          class="text-xs text-amber-700 hover:text-amber-900 underline whitespace-nowrap"
        >Dismiss</button>
      </div>
    </div>

    <!-- Session state -->
    <div class="def-card mb-6 p-5" :class="sessionActive ? 'bg-emerald-50/40 border-emerald-300' : ''">
      <h2 class="font-serif text-lg mb-2">Session</h2>
      <div v-if="checkingSession" class="text-sm text-stone-400 italic">Checking server…</div>
      <div v-else-if="sessionActive">
        <p class="text-sm text-emerald-800">
          Session active until <strong>{{ expiresHuman }}</strong>.
          Verse Explorer's "Render in modern English" is enabled.
        </p>
        <p class="text-xs text-stone-500 mt-1">
          Provider: <code class="bg-white px-1 rounded">{{ llmSettings.provider }}</code>
          · Model: <code class="bg-white px-1 rounded">{{ llmSettings.model }}</code>
        </p>
        <button
          @click="onSignOut"
          :disabled="signingOut"
          class="mt-3 px-4 py-2 text-sm font-medium rounded-lg border border-stone-300 text-stone-700 bg-white hover:border-stone-500 hover:bg-stone-50 transition disabled:opacity-50 disabled:cursor-not-allowed"
        >
          <span v-if="signingOut">Signing out…</span>
          <span v-else>Sign out (drop the held key)</span>
        </button>
      </div>
      <div v-else class="text-sm text-stone-600">
        No active session. Configure the provider below, then click <em>Start session</em>.
      </div>
    </div>

    <div class="def-card p-6 space-y-6">
      <!-- Provider preset -->
      <section>
        <label class="block text-sm font-medium text-stone-700 mb-3">Provider</label>
        <div class="grid sm:grid-cols-2 gap-2">
          <button
            v-for="p in presets"
            :key="p.id"
            @click="onPresetChange(p.id)"
            class="text-left p-3 rounded-lg border transition"
            :class="llmSettings.provider === p.id
              ? 'border-amber-500 bg-amber-50 text-amber-900'
              : 'border-stone-300 hover:border-stone-400 bg-white'"
          >
            <div class="font-medium">{{ p.label }}</div>
            <div class="text-xs text-stone-500 mt-1">{{ p.detail }}</div>
          </button>
        </div>
      </section>

      <!-- Base URL -->
      <section v-if="activePreset?.baseUrlEditable">
        <label class="block text-sm font-medium text-stone-700 mb-1">Base URL</label>
        <input
          v-model="llmSettings.baseUrl"
          type="text"
          placeholder="https://opencode.example/v1"
          class="w-full px-3 py-2 border border-stone-300 rounded-lg font-mono text-sm focus:border-amber-500 focus:outline-none"
        />
        <p class="text-xs text-stone-500 mt-1">
          Should end with <code class="bg-stone-100 px-1 rounded">/v1</code> — the OpenAI-compatible path.
          Don't include the trailing <code class="bg-stone-100 px-1 rounded">/chat/completions</code>.
        </p>
      </section>
      <section v-else>
        <label class="block text-sm font-medium text-stone-700 mb-1">Base URL <span class="text-stone-400 font-normal">(fixed for this provider)</span></label>
        <code class="block w-full px-3 py-2 bg-stone-50 border border-stone-200 rounded-lg font-mono text-sm text-stone-600">{{ llmSettings.baseUrl }}</code>
      </section>

      <!-- API key (transient — never persisted) -->
      <section>
        <label class="block text-sm font-medium text-stone-700 mb-1">
          API key
          <span class="ml-2 text-xs font-normal text-stone-500">— used once, sent straight to the backend, never stored in this browser</span>
        </label>
        <input
          v-model="apiKeyInput"
          type="password"
          autocomplete="off"
          spellcheck="false"
          placeholder="sk-… (cleared from the field after Start session)"
          class="w-full px-3 py-2 border border-stone-300 rounded-lg font-mono text-sm focus:border-amber-500 focus:outline-none"
        />
        <p class="text-xs text-stone-500 mt-1">
          The 1828 backend probes the key, mints a session, and drops the key from server memory when the session ends or expires.
          Keys are never written to disk or DB. Backend restart drops all sessions; you'll re-authenticate.
        </p>
      </section>

      <!-- Model -->
      <section>
        <label class="block text-sm font-medium text-stone-700 mb-1">Model</label>
        <input
          v-model="llmSettings.model"
          type="text"
          placeholder="kimi-k2.6 / claude-sonnet-4 / gpt-4o-mini"
          class="w-full px-3 py-2 border border-stone-300 rounded-lg font-mono text-sm focus:border-amber-500 focus:outline-none"
        />
      </section>

      <!-- Advanced -->
      <details class="border-t border-stone-200 pt-4">
        <summary class="cursor-pointer text-sm font-medium text-stone-700">Advanced</summary>
        <div class="mt-4 grid sm:grid-cols-2 gap-4">
          <div>
            <label class="block text-xs text-stone-500 mb-1">Temperature</label>
            <input
              v-model.number="llmSettings.temperature"
              type="number"
              step="0.1"
              min="0"
              max="2"
              class="w-full px-3 py-2 border border-stone-300 rounded-lg font-mono text-sm focus:border-amber-500 focus:outline-none"
            />
          </div>
          <div>
            <label class="block text-xs text-stone-500 mb-1">Max tokens</label>
            <input
              v-model.number="llmSettings.maxTokens"
              type="number"
              step="100"
              min="100"
              max="8000"
              class="w-full px-3 py-2 border border-stone-300 rounded-lg font-mono text-sm focus:border-amber-500 focus:outline-none"
            />
          </div>
        </div>
      </details>

      <!-- Start session button -->
      <div class="border-t border-stone-200 pt-4">
        <button
          @click="onStartSession"
          :disabled="minting || !canMintSession() || !apiKeyInput.trim()"
          class="px-5 py-2.5 rounded-lg text-sm font-medium transition"
          :class="!minting && canMintSession() && apiKeyInput.trim()
            ? 'bg-amber-600 text-white hover:bg-amber-700'
            : 'bg-stone-200 text-stone-500 cursor-not-allowed'"
        >
          <span v-if="minting">Probing key + minting session…</span>
          <span v-else-if="sessionActive">Replace current session</span>
          <span v-else>Start session</span>
        </button>
        <p v-if="mintError" class="text-sm text-red-700 mt-3">{{ mintError }}</p>
      </div>
    </div>

    <section class="mt-8 def-card p-6 bg-stone-50/50">
      <h2 class="font-serif text-lg mb-3">How rendering works (new flow)</h2>
      <ol class="text-sm text-stone-700 space-y-2 list-decimal list-inside">
        <li>You paste your provider API key above and click <em>Start session</em>.</li>
        <li>The 1828 backend probes the key (a cheap <code class="bg-stone-100 px-1 rounded">/v1/models</code> call), mints a session ID, and stores your key in memory for the session lifetime.</li>
        <li>Verse Explorer's "Render in modern English" sends the verse + tier-words to <code class="bg-stone-100 px-1 rounded">/api/llm/render</code> with the session cookie. The backend uses your held key to call your provider.</li>
        <li>Sessions live ~24 hours with a sliding window — active use extends them; idle sessions expire. Backend restart drops all sessions.</li>
        <li>Token costs land on YOUR provider account, not 1828.ibeco.me's.</li>
      </ol>
    </section>

    <section class="mt-6 def-card p-6 bg-stone-50/40 border-stone-300">
      <h2 class="font-serif text-lg mb-3">CORS — now handled server-side</h2>
      <p class="text-sm text-stone-700 leading-relaxed">
        The old browser → LM Studio direct call required enabling CORS at the endpoint.
        That's no longer needed — the 1828 backend talks to your provider server-side,
        and the browser only talks to <code class="bg-stone-100 px-1 rounded">1828.ibeco.me/api/*</code> (same-origin).
      </p>
      <p class="text-xs text-stone-500 mt-2">
        LM Studio is intentionally not in the BYOK provider list — D-LP-1 reserves it for embeddings.
        Use OpenCode Go / Zen, OpenAI, OpenRouter, or a custom OpenAI-compatible endpoint.
      </p>
    </section>
  </div>
</template>
