// LLM settings — localStorage-backed so the API key stays on the user's
// machine. Settings cover both LM Studio (localhost:1234) and opencode-go
// (configurable URL), and any other OpenAI-compatible chat-completion
// endpoint. The site itself does not hold or proxy any key.

import { reactive, watch } from 'vue'

const STORAGE_KEY = '1828-illuminated:llm-settings:v1'

export interface LLMSettings {
  provider: 'lm-studio' | 'opencode-go' | 'custom'
  baseUrl: string         // e.g. http://localhost:1234/v1
  apiKey: string          // optional for LM Studio
  model: string           // e.g. kimi-k2.6 / claude-sonnet-4 / openai/gpt-4o
  temperature: number
  maxTokens: number
}

const DEFAULTS: LLMSettings = {
  provider: 'lm-studio',
  baseUrl: 'http://localhost:1234/v1',
  apiKey: '',
  model: '',
  temperature: 0.3,
  maxTokens: 800,
}

function load(): LLMSettings {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return { ...DEFAULTS }
    const parsed = JSON.parse(raw)
    return { ...DEFAULTS, ...parsed }
  } catch {
    return { ...DEFAULTS }
  }
}

export const llmSettings = reactive<LLMSettings>(load())

watch(llmSettings, (s) => {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(s))
  } catch {
    /* localStorage might be disabled — ignore */
  }
}, { deep: true })

// Presets — wired to the provider selector in Settings.vue
export const PROVIDER_PRESETS: Record<LLMSettings['provider'], Partial<LLMSettings>> = {
  'lm-studio': {
    baseUrl: 'http://localhost:1234/v1',
    apiKey: '',
    model: '',  // LM Studio auto-uses the loaded model when model is blank
  },
  'opencode-go': {
    baseUrl: 'https://opencode.example/v1',  // user fills in real URL
    apiKey: '',
    model: 'kimi-k2.6',
  },
  'custom': {},
}

export function applyPreset(provider: LLMSettings['provider']) {
  const preset = PROVIDER_PRESETS[provider]
  llmSettings.provider = provider
  if (preset.baseUrl !== undefined) llmSettings.baseUrl = preset.baseUrl
  if (preset.apiKey !== undefined) llmSettings.apiKey = preset.apiKey
  if (preset.model !== undefined) llmSettings.model = preset.model
}

export function isConfigured(): boolean {
  return Boolean(llmSettings.baseUrl?.trim())
}
