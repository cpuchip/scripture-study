// LLM settings — localStorage-backed reader preferences for the BYOK
// session flow. Phase-5 reshape (D-LP-2, ratified 2026-05-20):
//
//   - api_key NEVER persists in localStorage anymore. It passes through
//     the Settings form straight into the POST /api/llm/session body and
//     is discarded after the session mints. The backend holds it in-memory
//     for the session TTL; we hold only the session_id.
//   - localStorage now carries `{provider, baseUrl, model, temperature,
//     maxTokens, session_id, expires_at}`. session_id is mirrored from
//     the cookie so the UI can show "Session active until …" without an
//     extra round-trip.
//   - A v1→v2 migration runs on first load: any old shape with an
//     `apiKey` field gets that field wiped, the rest is preserved, and a
//     one-time banner flag is set so Settings.vue can prompt the reader.

import { reactive, ref, watch } from 'vue'

const STORAGE_KEY_V1 = '1828-illuminated:llm-settings:v1'
const STORAGE_KEY_V2 = '1828-illuminated:llm-settings:v2'
const MIGRATION_BANNER_KEY = '1828-illuminated:llm-settings:migration-needed'

export type ProviderID = 'openai' | 'openrouter' | 'opencode-go' | 'opencode-zen' | 'lm-studio' | 'custom' | 'mock'

export interface LLMSettings {
  provider: ProviderID
  baseUrl: string        // base URL used at session-mint; the deployed proxy keeps it server-side after that
  model: string          // e.g. kimi-k2.6 / claude-sonnet-4 / openai/gpt-4o
  temperature: number
  maxTokens: number
  session_id: string     // empty when no active session
  expires_at: string     // RFC3339; empty when no active session
}

const DEFAULTS: LLMSettings = {
  provider: 'opencode-go',
  baseUrl: '',
  model: '',
  temperature: 0.3,
  maxTokens: 800,
  session_id: '',
  expires_at: '',
}

// Reactive flag so Settings.vue can render the migration banner once and
// dismiss it on user acknowledgement.
export const migrationNeeded = ref(false)

function load(): LLMSettings {
  try {
    // v2 first — happy path on subsequent loads.
    const rawV2 = localStorage.getItem(STORAGE_KEY_V2)
    if (rawV2) {
      const parsed = JSON.parse(rawV2)
      return { ...DEFAULTS, ...parsed, session_id: parsed.session_id ?? '', expires_at: parsed.expires_at ?? '' }
    }
    // v1 migration: any document with `apiKey` is the old shape. Strip
    // the key, preserve the rest, mark the migration banner.
    const rawV1 = localStorage.getItem(STORAGE_KEY_V1)
    if (rawV1) {
      const parsed = JSON.parse(rawV1)
      const hadKey = typeof parsed?.apiKey === 'string' && parsed.apiKey.length > 0
      const migrated: LLMSettings = {
        ...DEFAULTS,
        provider: (parsed.provider as ProviderID) ?? DEFAULTS.provider,
        baseUrl: parsed.baseUrl ?? '',
        model: parsed.model ?? '',
        temperature: typeof parsed.temperature === 'number' ? parsed.temperature : DEFAULTS.temperature,
        maxTokens: typeof parsed.maxTokens === 'number' ? parsed.maxTokens : DEFAULTS.maxTokens,
        session_id: '',
        expires_at: '',
      }
      // Wipe v1 and prep for v2 persistence. Migration banner shown once.
      localStorage.removeItem(STORAGE_KEY_V1)
      if (hadKey) {
        localStorage.setItem(MIGRATION_BANNER_KEY, '1')
      }
      return migrated
    }
  } catch {
    /* localStorage might be disabled — fall through to defaults */
  }
  return { ...DEFAULTS }
}

export const llmSettings = reactive<LLMSettings>(load())

// Surface the migration banner once. Settings.vue dismisses it via
// dismissMigrationBanner() after the reader acknowledges.
if (typeof localStorage !== 'undefined' && localStorage.getItem(MIGRATION_BANNER_KEY) === '1') {
  migrationNeeded.value = true
}

watch(
  llmSettings,
  (s) => {
    try {
      localStorage.setItem(STORAGE_KEY_V2, JSON.stringify(s))
    } catch {
      /* localStorage might be disabled — ignore */
    }
  },
  { deep: true },
)

export function dismissMigrationBanner() {
  migrationNeeded.value = false
  try {
    localStorage.removeItem(MIGRATION_BANNER_KEY)
  } catch {
    /* ignore */
  }
}

// Presets — wired to the provider selector in Settings.vue. Each preset
// suggests a base URL + a typical model id; readers can override either.
//
// LM Studio is intentionally absent from the BYOK matrix (D-LP-1 reserves
// LM Studio for embeddings only) but kept as a preset for local dev: it
// uses provider=mock semantics on the server when no key probe is needed.
// "custom" lets readers point at any OpenAI-compatible endpoint.
export const PROVIDER_PRESETS: Record<ProviderID, Partial<LLMSettings>> = {
  'openai': {
    baseUrl: 'https://api.openai.com/v1',
    model: 'gpt-4o-mini',
  },
  'openrouter': {
    baseUrl: 'https://openrouter.ai/api/v1',
    model: 'openai/gpt-4o-mini',
  },
  'opencode-go': {
    baseUrl: '',  // reader provides their own
    model: 'kimi-k2.6',
  },
  'opencode-zen': {
    baseUrl: '',
    model: 'kimi-k2.6',
  },
  'lm-studio': {
    baseUrl: 'http://localhost:1234/v1',
    model: '',
  },
  'custom': {},
  'mock': {
    baseUrl: '',
    model: 'mock-model',
  },
}

export function applyPreset(provider: ProviderID) {
  const preset = PROVIDER_PRESETS[provider]
  llmSettings.provider = provider
  if (preset.baseUrl !== undefined) llmSettings.baseUrl = preset.baseUrl
  if (preset.model !== undefined) llmSettings.model = preset.model
}

/** True iff there's a session_id that hasn't expired locally. The
 *  authoritative check is GET /api/llm/session — see useLLMSession. */
export function isSessionLikelyActive(): boolean {
  if (!llmSettings.session_id) return false
  if (!llmSettings.expires_at) return false
  try {
    return new Date(llmSettings.expires_at).getTime() > Date.now()
  } catch {
    return false
  }
}

/** True iff Settings has enough to attempt a session-mint. */
export function canMintSession(): boolean {
  if (!llmSettings.provider) return false
  if (!llmSettings.model.trim()) return false
  // opencode-go / opencode-zen / custom need a base URL; openai /
  // openrouter / lm-studio / mock have defaults the backend supplies.
  if (['opencode-go', 'opencode-zen', 'custom'].includes(llmSettings.provider)) {
    return Boolean(llmSettings.baseUrl.trim())
  }
  return true
}
