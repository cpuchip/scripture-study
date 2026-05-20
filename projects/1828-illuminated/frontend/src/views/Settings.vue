<script setup lang="ts">
import { llmSettings, applyPreset, isConfigured } from '@/composables/useLLMSettings'

const presets = [
  { id: 'lm-studio' as const, label: 'LM Studio', detail: 'Local model at http://localhost:1234/v1 — no API key needed' },
  { id: 'opencode-go' as const, label: 'OpenCode Go', detail: 'Your OpenCode Go instance — supply base URL + key' },
  { id: 'custom' as const, label: 'Custom', detail: 'Any OpenAI-compatible /v1/chat/completions endpoint' },
]

function onPresetChange(id: 'lm-studio' | 'opencode-go' | 'custom') {
  applyPreset(id)
}
</script>

<template>
  <div class="max-w-3xl mx-auto px-6 py-10">
    <header class="mb-8">
      <h1 class="text-3xl font-serif mb-2">Settings</h1>
      <p class="text-stone-600">
        Configure your LLM endpoint to use the "Render in modern English" feature on Verse Explorer.
        <strong>This site never sees your API key</strong> — requests go directly from your browser to the endpoint you configure.
      </p>
    </header>

    <div class="def-card p-6 space-y-6">
      <!-- Provider preset -->
      <section>
        <label class="block text-sm font-medium text-stone-700 mb-3">Provider preset</label>
        <div class="grid sm:grid-cols-3 gap-2">
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
      <section>
        <label class="block text-sm font-medium text-stone-700 mb-1">Base URL</label>
        <input
          v-model="llmSettings.baseUrl"
          type="text"
          placeholder="http://localhost:1234/v1"
          class="w-full px-3 py-2 border border-stone-300 rounded-lg font-mono text-sm focus:border-amber-500 focus:outline-none"
        />
        <p class="text-xs text-stone-500 mt-1">
          Should end with <code class="bg-stone-100 px-1 rounded">/v1</code> — the OpenAI-compatible path. Don't include the trailing <code class="bg-stone-100 px-1 rounded">/chat/completions</code>.
        </p>
      </section>

      <!-- API key -->
      <section>
        <label class="block text-sm font-medium text-stone-700 mb-1">API key</label>
        <input
          v-model="llmSettings.apiKey"
          type="password"
          placeholder="Optional — empty for LM Studio"
          autocomplete="off"
          class="w-full px-3 py-2 border border-stone-300 rounded-lg font-mono text-sm focus:border-amber-500 focus:outline-none"
        />
        <p class="text-xs text-stone-500 mt-1">
          Stored in your browser's localStorage. Never sent anywhere except the Base URL you configured above.
        </p>
      </section>

      <!-- Model -->
      <section>
        <label class="block text-sm font-medium text-stone-700 mb-1">Model name</label>
        <input
          v-model="llmSettings.model"
          type="text"
          placeholder="kimi-k2.6 / claude-sonnet-4 / gpt-4o / (blank for LM Studio's loaded model)"
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

      <div class="border-t border-stone-200 pt-4 text-sm">
        <span v-if="isConfigured()" class="text-emerald-700">
          ✓ Configured — Verse Explorer will offer the "Render in modern English" button.
        </span>
        <span v-else class="text-amber-700">
          Base URL is empty. Fill it in to enable the render feature.
        </span>
      </div>
    </div>

    <section class="mt-8 def-card p-6 bg-stone-50/50">
      <h2 class="font-serif text-lg mb-3">How rendering works</h2>
      <ol class="text-sm text-stone-700 space-y-2 list-decimal list-inside">
        <li>You click "Render in modern English" on a verse in the Verse Explorer.</li>
        <li>The site builds a prompt containing the verse text and the 1828 definitions of every tier word in it.</li>
        <li>Your browser POSTs that prompt to the endpoint above (this site never sees the request).</li>
        <li>The model returns a modernized rendering with original words marked in <code class="bg-stone-100 px-1 rounded">[brackets]</code> for transparency.</li>
        <li>Original and rendered versions display side by side.</li>
      </ol>
      <p class="text-xs text-stone-500 mt-4">
        This is a stretch-goal feature. Token costs land on your account / your local machine — that's why it's gated behind your own endpoint rather than running for every visitor.
      </p>
    </section>

    <section class="mt-8 def-card p-6 bg-amber-50/40 border-amber-300">
      <h2 class="font-serif text-lg mb-3">⚠ CORS: enable cross-origin in your endpoint</h2>
      <p class="text-sm text-stone-700 leading-relaxed">
        Browsers block requests from this site (e.g. <code class="bg-white px-1 rounded">localhost:8080</code>) to a different origin (e.g. LM Studio on <code class="bg-white px-1 rounded">localhost:1234</code>) unless the endpoint sends a permissive <code class="bg-white px-1 rounded">Access-Control-Allow-Origin</code> header. If "Render in modern English" returns a network error in your browser DevTools console, this is almost certainly why.
      </p>
      <h3 class="text-sm font-semibold text-stone-800 mt-4 mb-1">LM Studio</h3>
      <ol class="text-sm text-stone-700 space-y-1 list-decimal list-inside">
        <li>Open LM Studio</li>
        <li>Click the <strong>Developer</strong> (or <strong>Local Server</strong>) tab in the left sidebar</li>
        <li>Toggle <strong>"Enable CORS"</strong> ON before starting the server</li>
        <li>(Re)start the server</li>
      </ol>
      <h3 class="text-sm font-semibold text-stone-800 mt-4 mb-1">OpenCode Go / custom endpoint</h3>
      <p class="text-sm text-stone-700">
        Add an <code class="bg-white px-1 rounded">Access-Control-Allow-Origin: *</code> response header (or restrict it to your specific origin) in your server config. The request is a standard OpenAI-compatible POST to <code class="bg-white px-1 rounded">/v1/chat/completions</code>; the browser will preflight with OPTIONS, so allow that method too.
      </p>
      <h3 class="text-sm font-semibold text-stone-800 mt-4 mb-1">Quick verification</h3>
      <p class="text-sm text-stone-700">
        Open the browser DevTools (F12) → Console tab → click "Render in modern English". A CORS error will explicitly say <em>"blocked by CORS policy: No 'Access-Control-Allow-Origin' header"</em> — different from a generic network error or 4xx response. If you see THAT exact message, the fix is at your endpoint, not this site.
      </p>
    </section>
  </div>
</template>
