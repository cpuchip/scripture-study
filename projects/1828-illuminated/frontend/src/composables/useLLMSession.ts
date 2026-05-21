// LLM session lifecycle — BYOK + session_id flow (D-LP-2, ratified 2026-05-20).
//
// The reader pastes their provider API key into Settings; we POST it to
// /api/llm/session; the backend probes the key, mints a 32-byte session_id,
// holds the key in-memory for the session TTL, and returns the id + an
// expiry. We persist `{session_id, expires_at}` locally so the UI can show
// "Session active until …" without an extra round-trip. The API KEY ITSELF
// never reaches localStorage.

import { ref } from 'vue'

import { apiUrl } from './useApiBase'
import { llmSettings, type ProviderID } from './useLLMSettings'

/** True after a successful mint or session-inspect; flipped to false on
 *  end-session, 401, or expiry. Other components observe this so render
 *  buttons enable/disable in sync with Settings. */
export const sessionActive = ref(false)

export interface SessionInfo {
  active: boolean
  provider?: string
  model?: string
  expires_at?: string
}

export interface MintRequest {
  provider: ProviderID
  base_url?: string
  api_key: string  // passes through, never persisted client-side
  model: string
}

export interface MintError {
  error: string       // backend error code (e.g. key_probe_failed)
  message: string     // human-readable
}

/** Mint a new session. The api_key is sent once and discarded — never
 *  stored anywhere client-side. */
export async function startSession(req: MintRequest): Promise<{ session_id: string; expires_at: string } | MintError> {
  try {
    const resp = await fetch(apiUrl('/llm/session'), {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',  // accept the i1828_session cookie back
      body: JSON.stringify({
        provider: req.provider,
        base_url: req.base_url ?? '',
        api_key: req.api_key,
        model: req.model,
      }),
    })
    const json = await resp.json().catch(() => ({}))
    if (!resp.ok) {
      return {
        error: (json && (json.error || json.code)) || `http_${resp.status}`,
        message: (json && json.message) || `Session mint failed (HTTP ${resp.status}).`,
      }
    }
    llmSettings.session_id = json.session_id ?? ''
    llmSettings.expires_at = json.expires_at ?? ''
    sessionActive.value = Boolean(llmSettings.session_id)
    return { session_id: llmSettings.session_id, expires_at: llmSettings.expires_at }
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : String(e)
    return { error: 'network_error', message: msg }
  }
}

/** Drop the held key on the backend, expire the cookie, clear local
 *  mirror. Safe to call when there's no active session. */
export async function endSession(): Promise<void> {
  const id = llmSettings.session_id
  llmSettings.session_id = ''
  llmSettings.expires_at = ''
  sessionActive.value = false
  if (!id) return
  try {
    await fetch(apiUrl('/llm/session'), {
      method: 'DELETE',
      credentials: 'include',
      headers: { Authorization: `Bearer ${id}` },
    })
  } catch {
    /* best-effort; the local mirror is already cleared */
  }
}

/** Authoritative session check against the backend. Updates the local
 *  mirror and the reactive flag. Call on Settings page mount + after any
 *  401 from /api/llm/render. */
export async function refreshSession(): Promise<SessionInfo> {
  const id = llmSettings.session_id
  const headers: Record<string, string> = {}
  if (id) headers.Authorization = `Bearer ${id}`
  try {
    const resp = await fetch(apiUrl('/llm/session'), {
      method: 'GET',
      credentials: 'include',
      headers,
    })
    const json = (await resp.json().catch(() => ({}))) as SessionInfo
    if (json?.active) {
      sessionActive.value = true
      if (json.expires_at) llmSettings.expires_at = json.expires_at
      return json
    }
    // Server says no — clear the local mirror.
    sessionActive.value = false
    llmSettings.session_id = ''
    llmSettings.expires_at = ''
    return { active: false }
  } catch {
    // Network error — leave the local mirror alone; we'll discover the
    // real state on the next attempt.
    return { active: sessionActive.value }
  }
}

/** Cheap optimistic check using only localStorage. Useful for gating
 *  buttons; the authoritative check is refreshSession(). */
export function isSessionActive(): boolean {
  return sessionActive.value
}

// Initialize sessionActive from localStorage on module load. Components
// that mount Settings call refreshSession() to confirm against the server.
if (llmSettings.session_id && llmSettings.expires_at) {
  try {
    if (new Date(llmSettings.expires_at).getTime() > Date.now()) {
      sessionActive.value = true
    } else {
      llmSettings.session_id = ''
      llmSettings.expires_at = ''
    }
  } catch {
    /* ignore */
  }
}
