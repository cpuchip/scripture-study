// LLM render — given a verse + the tier words present in it, ask the
// backend's /api/llm/render to render the verse in modern English while
// preserving each tier word's 1828 sense.
//
// Phase-5 cutover: the browser no longer talks directly to the provider.
// Instead it POSTs to the i1828 backend, which holds the reader's BYOK
// session key in-memory and round-trips to the provider server-side.
// This solves LM Studio CORS for readers, keeps reader keys out of the
// 1828.ibeco.me bundle entirely, and lets us attribute throttling
// honestly (rate_limited_by_1828 vs upstream_provider_error).

import { ref } from 'vue'

import { apiUrl } from './useApiBase'
import { llmSettings } from './useLLMSettings'
import { refreshSession, sessionActive } from './useLLMSession'
import { tokenize, useWordData } from './useWordData'

export interface RenderResult {
  modernized: string
  promptUsed: string
  durationMs: number
  model: string
  provider: string
  usage: Record<string, number>
}

export type RenderErrorKind =
  | 'reauth'                  // 401 — session expired/missing
  | 'rate_limited_by_1828'    // 429 — our cap, not the provider's
  | 'upstream_provider_error' // 502 — provider returned an error (passed through)
  | 'feature_disabled'        // 503 — proxy off
  | 'network'                 // fetch failed before we got a response
  | 'unknown'

export interface RenderError {
  kind: RenderErrorKind
  message: string
  /** Set on 429 — seconds to wait before retrying. */
  retryAfterSeconds?: number
}

export interface RenderState {
  loading: boolean
  result: RenderResult | null
  error: RenderError | null
}

export function useLLMRender() {
  const state = ref<RenderState>({ loading: false, result: null, error: null })
  const data = useWordData()

  async function render(verseText: string): Promise<void> {
    state.value = { loading: true, result: null, error: null }

    // Collect tier words present in the verse + their 1828 first-sense.
    // We hit the backend's 1828 endpoint per-unique-tier-word; the LRU
    // cache in useWordData makes repeat verses cheap. The prompt build
    // itself happens server-side now — we only need to send the verse
    // text + the {word, sense} pairs.
    const present = new Map<string, string>()
    const uniqueTierWords: string[] = []
    for (const seg of tokenize(verseText)) {
      if (seg.word && !present.has(seg.word)) {
        present.set(seg.word, '')
        uniqueTierWords.push(seg.word)
      }
    }
    if (uniqueTierWords.length > 0) {
      const results = await Promise.all(uniqueTierWords.map(w => data.get1828(w)))
      uniqueTierWords.forEach((w, i) => {
        const r = results[i]
        const firstSense = r?.entries[0]?.definitions[0] ?? ''
        if (firstSense) present.set(w, firstSense)
        else present.delete(w)
      })
    }

    const tierWords = Array.from(present.entries()).map(([word, sense]) => ({ word, sense }))

    const startMs = performance.now()
    try {
      const resp = await fetch(apiUrl('/llm/render'), {
        method: 'POST',
        credentials: 'include',  // carry the i1828_session cookie
        headers: {
          'Content-Type': 'application/json',
          // Bearer mirror — defensive against ad-blockers that strip cookies.
          ...(llmSettings.session_id ? { Authorization: `Bearer ${llmSettings.session_id}` } : {}),
        },
        body: JSON.stringify({
          verseText,
          tierWords,
          options: {
            maxTokens: llmSettings.maxTokens,
            temperature: llmSettings.temperature,
            stream: false,
          },
        }),
      })

      if (resp.status === 401) {
        // Local mirror was stale or the server janitor evicted the
        // session — sync with the server and surface a clear re-auth state.
        await refreshSession()
        state.value = {
          loading: false,
          result: null,
          error: {
            kind: 'reauth',
            message: 'Session expired or missing. Re-authenticate in Settings to render.',
          },
        }
        return
      }

      if (resp.status === 429) {
        const json = await resp.json().catch(() => ({} as any))
        const retry = Number(json?.retry_after_seconds) || undefined
        state.value = {
          loading: false,
          result: null,
          error: {
            kind: 'rate_limited_by_1828',
            // Honor D-BE-AUTH attribution: this is OUR throttle, not the
            // provider's. The body the backend returns already says so,
            // but we lift the message to the UI verbatim.
            message: (json?.message as string) || 'Throttled by 1828.ibeco.me (not your provider).',
            retryAfterSeconds: retry,
          },
        }
        return
      }

      if (resp.status === 502) {
        const json = await resp.json().catch(() => ({} as any))
        state.value = {
          loading: false,
          result: null,
          error: {
            kind: 'upstream_provider_error',
            message: (json?.upstream_message as string) || (json?.message as string) || 'Upstream provider returned an error.',
          },
        }
        return
      }

      if (resp.status === 503) {
        const json = await resp.json().catch(() => ({} as any))
        state.value = {
          loading: false,
          result: null,
          error: {
            kind: 'feature_disabled',
            message: (json?.message as string) || 'LLM render is disabled on this deploy.',
          },
        }
        return
      }

      if (!resp.ok) {
        const text = await resp.text().catch(() => '')
        state.value = {
          loading: false,
          result: null,
          error: {
            kind: 'unknown',
            message: `HTTP ${resp.status} ${resp.statusText} — ${text.slice(0, 200)}`,
          },
        }
        return
      }

      const json = await resp.json()
      const content: string = json?.modernized ?? ''
      if (!content) {
        state.value = {
          loading: false,
          result: null,
          error: { kind: 'unknown', message: 'Empty response from backend' },
        }
        return
      }

      const modernized = content.trim()
      state.value = {
        loading: false,
        result: {
          modernized,
          promptUsed: json?.promptUsed ?? '',
          durationMs: typeof json?.durationMs === 'number' ? json.durationMs : performance.now() - startMs,
          model: json?.model ?? '',
          provider: json?.provider ?? '',
          usage: json?.usage ?? {},
        },
        error: null,
      }

      // Capture the render as a study-tree node. Cross-domain payoff: this
      // attaches the modernized passage as a child of whatever was active
      // (typically the chapter or verse the reader just rendered), and
      // subsequent word-clicks inside the modernized text branch from here.
      const { visit } = await import('./useStudyTree')
      visit({
        kind: 'render',
        sourceText: verseText,
        modernized,
        model: json?.model ?? 'unknown',
        provider: json?.provider ?? 'unknown',
      })
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : String(e)
      state.value = {
        loading: false,
        result: null,
        error: { kind: 'network', message: msg },
      }
    }
  }

  function reset() {
    state.value = { loading: false, result: null, error: null }
  }

  return { state, render, reset, sessionActive }
}
