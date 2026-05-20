// LLM-render — given a verse + the tier-words present in it, ask the user's
// configured OpenAI-compatible endpoint to render the verse in modern English
// while preserving each tier word's 1828 sense. Output is a string with
// inline [original-word] markers so the substitution is transparent.
//
// The request happens directly from the user's browser to their endpoint
// — the 1828.ibeco.me site never sees the request or the API key.

import { ref } from 'vue'
import { llmSettings, isConfigured } from './useLLMSettings'
import { useWordData, tokenize } from './useWordData'

export interface RenderResult {
  modernized: string
  promptUsed: string
  durationMs: number
}

export interface RenderState {
  loading: boolean
  result: RenderResult | null
  error: string | null
}

export function useLLMRender() {
  const state = ref<RenderState>({ loading: false, result: null, error: null })
  const data = useWordData()

  async function render(verseText: string): Promise<void> {
    if (!isConfigured()) {
      state.value = { loading: false, result: null, error: 'Settings not configured. Open /settings to add your API endpoint.' }
      return
    }
    state.value = { loading: true, result: null, error: null }

    // Collect tier words present in the verse + their 1828 first-sense
    const present = new Map<string, string>()
    for (const seg of tokenize(verseText)) {
      if (seg.word && !present.has(seg.word)) {
        const defs = data.get1828(seg.word)
        const firstSense = defs[0]?.definitions[0] ?? ''
        if (firstSense) present.set(seg.word, firstSense)
      }
    }

    const wordTable = Array.from(present.entries())
      .map(([w, def]) => `- **${w}**: ${def.replace(/\s+/g, ' ').slice(0, 200)}`)
      .join('\n')

    const userPrompt = `You are rendering a scripture passage from KJV / Restoration English into clear modern English while preserving the 1828 Webster meanings of specific words listed below.

**Original passage:**

${verseText}

**Words to preserve in their 1828 sense:**

${wordTable || '(no flagged words — render naturally)'}

**Instructions:**

1. Render the passage in clear modern English.
2. For each flagged word, replace it with a phrase that captures its 1828 sense as defined above. Don't substitute the modern dictionary meaning.
3. Mark each substituted phrase with the original word in square brackets after it, like this: "they tolerated [allowed] their fathers' deeds". This keeps the substitution transparent.
4. Do not add theological interpretation, application, or commentary. Just translate the language.
5. Preserve sentence structure where possible.

**Output the modernized passage only. No preamble, no explanation.**`

    const startMs = performance.now()
    try {
      const url = llmSettings.baseUrl.replace(/\/$/, '') + '/chat/completions'
      const body = {
        model: llmSettings.model || undefined,
        messages: [{ role: 'user', content: userPrompt }],
        temperature: llmSettings.temperature,
        max_tokens: llmSettings.maxTokens,
      }
      const headers: Record<string, string> = { 'Content-Type': 'application/json' }
      if (llmSettings.apiKey?.trim()) {
        headers['Authorization'] = `Bearer ${llmSettings.apiKey.trim()}`
      }
      const resp = await fetch(url, {
        method: 'POST',
        headers,
        body: JSON.stringify(body),
      })
      if (!resp.ok) {
        const text = await resp.text().catch(() => '')
        throw new Error(`HTTP ${resp.status} ${resp.statusText} — ${text.slice(0, 200)}`)
      }
      const json = await resp.json()
      const content: string = json?.choices?.[0]?.message?.content ?? ''
      if (!content) throw new Error('Empty response from model')

      state.value = {
        loading: false,
        result: {
          modernized: content.trim(),
          promptUsed: userPrompt,
          durationMs: performance.now() - startMs,
        },
        error: null,
      }
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : String(e)
      state.value = { loading: false, result: null, error: msg }
    }
  }

  function reset() {
    state.value = { loading: false, result: null, error: null }
  }

  return { state, render, reset }
}
