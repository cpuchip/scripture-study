---
title: 1828-illuminated — LLM proxy
date: 2026-05-20
status: proposed
workstream: WS7
parent: backend-pivot.md
purpose: >
  Move the LLM-render call out of the user's browser and into the
  backend. Solves LM Studio CORS (the deployed site can't talk to a
  reader's localhost from the browser cleanly), enables server-side
  key holding, and lets us swap providers without frontend churn.
---

# LLM Proxy

## I. The problem

Today, `frontend/src/composables/useLLMRender.ts` does this:

```ts
const url = llmSettings.baseUrl.replace(/\/$/, '') + '/chat/completions'
…
const resp = await fetch(url, {
  method: 'POST',
  headers,
  body: JSON.stringify(body),
})
```

The fetch goes from the user's browser directly to whatever URL they configured. That works for **Michael's** machine when he's pointed at `http://localhost:1234/v1` because his browser and LM Studio are on the same machine. It does NOT work for:

- Any reader who configures a remote OpenAI / Anthropic / OpenRouter key. The key is now sitting in `localStorage` on a page hosted at `1828.ibeco.me` — exposed to any script running in that tab.
- LM Studio CORS. The reader installs LM Studio at home, sets baseUrl to `http://localhost:1234`, and the browser refuses the call because `1828.ibeco.me` is a different origin. LM Studio doesn't ship permissive CORS headers by default. The reader is stuck.
- Any provider that requires authentication via header signing (some hosted Anthropic endpoints, AWS Bedrock, etc.).

A backend proxy fixes all three. The frontend sends `{ verseText, tierWords }` to `/api/llm/render`; the backend reads its own env vars for provider config and key, calls the upstream, returns the result.

## II. Tradeoffs of moving the call server-side

- **Backend now holds a usable API key.** If the deployed site uses an OpenAI key, that key has cost-bearing potential. **D-BE-AUTH** in `backend-pivot.md §IX` flags this: open `/api/llm/render` to anonymous traffic is a free-LLM-proxy invitation. Per-IP rate limit + maybe a daily token cap solves the abuse path. Phase 2 hardening; phase 1 ships behind a `LLM_PROXY_ENABLED=false` default.
- **Local-LM-Studio users lose their personal setup.** If Michael's setup is "my LM Studio at home, my model, my temp," the deployed site's proxy uses the *deploy's* config, not Michael's. Acceptable: the deployed site is for *readers*. Local dev keeps the same env vars on a `.env` file pointing at his own LM Studio.
- **One more thing to monitor.** Provider outages now affect the deployed site. The frontend should gracefully degrade — "LLM rendering temporarily unavailable, try again in a moment" — instead of failing opaque.

These are net wins given the project's audience (anyone studying scripture with the 1828 lens; the LM Studio path was Michael-shaped).

## III. The endpoint

```
POST /api/llm/render
Content-Type: application/json

{
  "verseText": "And the glory of God is intelligence, or, in other words, light and truth.",
  "tierWords": [
    { "word": "intelligence", "sense": "The act or state of knowing; the perception of facts and truth..." },
    { "word": "glory", "sense": "Brightness; luster; splendor..." }
  ],
  "options": {
    "maxTokens": 800,     // optional override; backend caps at MAX_TOKENS_HARD (default 1500)
    "temperature": 0.3,   // optional override; backend caps at 0.7
    "stream": false       // streaming is phase 2
  }
}
```

Returns:

```json
{
  "modernized": "And the glory [brightness; splendor] of God is intelligence [knowing], or, in other words, light and truth.",
  "promptUsed": "...",
  "model": "kimi-k2.6",
  "provider": "lm-studio",
  "durationMs": 1834,
  "usage": { "prompt_tokens": 312, "completion_tokens": 78 }
}
```

`provider` and `model` come from the backend's env config; the frontend doesn't choose. The `usage` block is informational; useful for the audit phase.

**Streaming via SSE (`?stream=1`)** is the obvious upgrade and **out of scope for phase 1**. Document the carry-forward; the frontend's existing single-shot UX needs no streaming.

## IV. Provider abstraction

A small Go interface so swapping providers is a config change:

```go
// backend/internal/llm/provider.go

type Provider interface {
    Name() string
    Render(ctx context.Context, req RenderRequest) (*RenderResponse, error)
}

type RenderRequest struct {
    VerseText  string
    TierWords  []TierWord
    Options    Options
}

type RenderResponse struct {
    Modernized   string
    PromptUsed   string
    Model        string
    DurationMs   int64
    Usage        Usage
}

// Implementations:
//   - openai_compat.go    — works for LM Studio, OpenAI, OpenRouter, opencode-go,
//                            anything that speaks /v1/chat/completions
//   - anthropic.go        — native Anthropic Messages API (different shape; common provider)
//   - mock.go             — for dev / tests; returns a canned response
```

Backend reads `LLM_PROVIDER` env (`openai-compat`, `anthropic`, `mock`) at boot, instantiates one provider, holds it. Hot-swap is a deploy.

**Why bother with the abstraction in phase 1?** Because the frontend will be written against `/api/llm/render` once; changing providers without breaking the frontend is the whole point of the abstraction. Cost: maybe 80 lines of Go beyond a single-provider implementation.

## V. The prompt — moves from frontend to backend

The current prompt template is in `useLLMRender.ts`. It's good. We move it verbatim into the backend (with the same `{verseText}` + `{wordTable}` substitution) so the prompt isn't a moving target across versions of the frontend. The frontend ships only the input (text + tier-word list), not the prompt scaffolding.

**Stewardship fix during the move:** the current template ends with `"**Output the modernized passage only. No preamble, no explanation.**"`. Add a hard cap: `"Reply in 800 tokens or fewer. If the passage is longer, modernize until the cap and end with [...continued]."` Removes one runtime failure mode (model rambles, hits token cap mid-word, returns truncated mush). Document the change in the migration commit.

## VI. Environment configuration

```bash
# Required for production
LLM_PROVIDER=openai-compat                  # or 'anthropic' or 'mock'
LLM_BASE_URL=http://lm-studio.host:1234/v1  # ignored by 'anthropic' provider
LLM_API_KEY=sk-…                            # blank for LM Studio
LLM_MODEL=kimi-k2.6                         # or 'claude-sonnet-4-7' etc.

# Defaults (with safe fallbacks)
LLM_MAX_TOKENS_DEFAULT=800
LLM_MAX_TOKENS_HARD=1500
LLM_TEMPERATURE_DEFAULT=0.3
LLM_TEMPERATURE_HARD=0.7
LLM_TIMEOUT_SECONDS=60

# Rate limiting (D-LP-3)
LLM_RATE_PER_IP_PER_MIN=10
LLM_RATE_PER_IP_PER_DAY=200
LLM_GLOBAL_TOKEN_CAP_PER_DAY=200000        # ~$0.40 at modern rates; tune by provider

# Kill switch
LLM_PROXY_ENABLED=true
```

Backend boots; if `LLM_PROVIDER=mock` or `LLM_PROXY_ENABLED=false`, `/api/llm/render` returns 503 + a friendly "feature disabled" body. Frontend renders that state with the existing "Settings not configured" UX.

**The `useLLMSettings.ts` localStorage** today carries provider + URL + key + model + temp + maxTokens. After cutover:
- The user-configurable surface shrinks to: temperature, max tokens, and an "advanced" override of provider/URL/key/model only if the deploy enables it via `LLM_ALLOW_USER_OVERRIDE=true`. Default off in production.
- For Michael's local dev, the override path lets him test against his own LM Studio without redeploying. So the feature isn't deleted — it becomes optional and gated.

## VII. Rate limiting + abuse protection

**Three layers, each cheap:**

1. **Per-IP request rate** — leaky-bucket per source IP. Default `10/min` and `200/day`. Configured via env. Hits beyond the cap return 429.
2. **Global daily token cap** — accumulated `prompt_tokens + completion_tokens` across all requests. Hits beyond the cap return 503 "daily quota exhausted." Resets at UTC midnight. Logs to stderr at 50%, 80%, 100%.
3. **Per-request token cap** — `LLM_MAX_TOKENS_HARD`. Backend rejects (or silently clamps; D-LP-5 decides) requests asking for more.

**No per-user accounts in v1.** IP-based limiting is coarse but proportionate to the threat — an LLM proxy worth abusing needs >10 req/min per attacker, and that's a flag.

**Logging.** Every `/api/llm/render` call logs (anonymized: hashed IP, request token count, response token count, provider, model, duration, status). Goes to stdout; Dokploy collects it. Provides the observability needed to tune the caps over time.

## VIII. Decisions

| # | Decision | Default | Stakes |
|---|---|---|---|
| **D-LP-1** | Provider for v1 deploy | LM Studio at `engine.ibeco.me` (Michael's existing infra) or hosted via OpenRouter | Per-token cost vs latency + dependency posture |
| **D-LP-2** | Allow user override of provider/URL/key | Off in production; on in local dev | Threat surface vs flexibility |
| **D-LP-3** | Per-IP rate limits | 10/min, 200/day | Conservative; cheap to raise |
| **D-LP-4** | Global daily token cap | 200,000 (~$0.40-1.50/day at modern rates) | If hit, 503 — visible to users; that's intentional |
| **D-LP-5** | Behavior when `maxTokens > LLM_MAX_TOKENS_HARD` | Clamp silently and log | Reject would be more honest; clamp keeps UX smooth |
| **D-LP-6** | Streaming via SSE | Phase 2 | Single-shot is fine for verse-render UX |
| **D-LP-7** | Cache rendered results in DB | Phase 2 | Same verse rendered twice = same cost twice today; cache would be a clear win once we measure |
| **D-LP-8** | Frontend retains `useLLMSettings.ts` (with reduced scope) vs deletes | Retain (reduced) | Local dev override use case |
| **D-LP-9** | Default `LLM_PROVIDER` if env unset | `mock` | Fail safe; production must set explicitly |

## IX. Verification

After phase ships:
- `curl -X POST /api/llm/render -d '{"verseText": "the glory of God is intelligence", "tierWords": [{"word":"intelligence","sense":"…"}]}'` returns a modernized rendering.
- Setting `LLM_PROVIDER=mock` returns a canned response — useful for the frontend integration test without spending tokens.
- Setting `LLM_PROXY_ENABLED=false` returns 503; the frontend's existing error path renders cleanly.
- A reader on a different network from the LM Studio host gets a successful response (CORS solved, key safely server-side).
- 11th request in the same minute from one IP returns 429.
- The frontend's existing `Render` button still works against the new endpoint without UX regression.

## X. Risks

- **Cost runaway.** Real. Multi-layer cap (per-IP, global, per-request) is the mitigation; the kill switch (`LLM_PROXY_ENABLED=false`) is the failsafe. Worth a budget alarm in Dokploy if the provider supports usage webhooks.
- **Provider lock-in via the abstraction layer being too OpenAI-shaped.** Anthropic's Messages API has a different prompt structure (system message separate from user). The interface needs to model that, not pretend everything is OpenAI. Mitigation: build the Anthropic provider in phase 1 alongside the openai-compat one (even if it stays unused) to keep the interface honest.
- **Local dev divergence.** Michael's local-LM-Studio path now requires either running the backend locally with env overrides, or keeping the user-override path in the frontend behind a dev flag. Both are fine; pick one and document.
- **The `useLLMSettings.ts` localStorage data persists across deploys.** After cutover, users with old settings may see confusing "settings not used" hints. Migration: on first load after cutover, frontend reads localStorage, detects the v1 shape, shows a one-time banner explaining the new model. Cheap; honest.
- **Hosted LM Studio reachability.** If the deploy uses `engine.ibeco.me`'s LM Studio (Michael's existing infra), and that goes down, the proxy goes down. Acceptable for v1; documented as a known dependency.

## XI. Out of scope

- **Streaming responses** (SSE) — phase 2.
- **Render-result caching** (DB-backed memo per `{verseHash, modelHash, promptHash}`) — phase 2; clear win but adds complexity.
- **Per-user accounts / API keys** — explicitly not v1.
- **Multiple simultaneous providers (failover)** — single provider per deploy.
- **Function-calling / tool-use rendering** — out of scope; this surface is text-in, text-out.
- **Render presets** ("rewrite for 8th grade" / "scholarly tone") — possibly worth adding as a prompt-template selector later; for v1, one template.
