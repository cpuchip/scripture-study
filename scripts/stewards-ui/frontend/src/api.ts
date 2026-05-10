// Typed fetch wrappers for /api/* endpoints. Each function maps to a
// single Go handler. Consumers in views/ import these directly so
// path strings live in exactly one place.

export type DashboardResp = {
  pg: { ok: boolean; error?: string }
  soak: {
    schedule_enabled: boolean
    last_pass_id?: string
    last_pass_status?: string
    last_pass_started_at?: string
    last_pass_finished_at?: string
    dirty_queue_depth: number
  }
  in_flight: WorkItemBrief[] | null
  recent_errors: ErrorBrief[] | null
  fetched_at_ms: number
}

export type WorkItemBrief = {
  id: string
  slug: string
  pipeline: string
  current_stage: string
  status: string
  tokens_in: number
  tokens_out: number
  updated_at?: string
}

export type ErrorBrief = {
  id: number
  kind: string
  provider: string
  error: string
  done_at?: string
}

async function getJSON<T>(path: string): Promise<T> {
  const r = await fetch(path)
  if (!r.ok) {
    let msg = `HTTP ${r.status}`
    try {
      const body = await r.json()
      if (body && typeof body.error === 'string') msg = body.error
    } catch {
      /* ignore */
    }
    throw new Error(msg)
  }
  return r.json() as Promise<T>
}

export type StudyBrief = {
  slug: string
  kind: string
  title?: string
  body_chars: number
  created_at?: string
  updated_at?: string
}

export type StudiesListResp = {
  items: StudyBrief[]
  total: number
}

export type CitationLite = {
  ref: string
  count?: number
}

export type SimilarHit = {
  slug: string
  title?: string
  distance?: number
}

export type StudyDetail = {
  slug: string
  kind: string
  title?: string
  body: string
  frontmatter?: Record<string, unknown>
  created_at?: string
  updated_at?: string
  citations: CitationLite[]
  similar: SimilarHit[]
}

export type SearchHit = {
  slug: string
  kind?: string
  title?: string
  snippet?: string
  score?: number
}

export type SearchResp = {
  query: string
  mode: string
  hits: SearchHit[]
}

export const api = {
  dashboard: () => getJSON<DashboardResp>('/api/dashboard'),
  studiesList: (params?: { kind?: string; limit?: number; offset?: number }) => {
    const q = new URLSearchParams()
    if (params?.kind) q.set('kind', params.kind)
    if (params?.limit) q.set('limit', String(params.limit))
    if (params?.offset) q.set('offset', String(params.offset))
    const qs = q.toString()
    return getJSON<StudiesListResp>(`/api/studies/list${qs ? '?' + qs : ''}`)
  },
  studyGet: (slug: string) =>
    getJSON<StudyDetail>(`/api/studies/get?slug=${encodeURIComponent(slug)}`),
  studiesSearch: (q: string, opts?: { mode?: string; limit?: number }) => {
    const p = new URLSearchParams({ q })
    if (opts?.mode) p.set('mode', opts.mode)
    if (opts?.limit) p.set('limit', String(opts.limit))
    return getJSON<SearchResp>(`/api/studies/search?${p}`)
  },
  workItemsList: (params?: { pipeline?: string; status?: string; limit?: number }) => {
    const q = new URLSearchParams()
    if (params?.pipeline) q.set('pipeline', params.pipeline)
    if (params?.status) q.set('status', params.status)
    if (params?.limit) q.set('limit', String(params.limit))
    const qs = q.toString()
    return getJSON<WorkItemsListResp>(`/api/work-items/list${qs ? '?' + qs : ''}`)
  },
  workItemGet: (idOrSlug: string) => {
    const q = idOrSlug.includes('-') && idOrSlug.length === 36
      ? `id=${idOrSlug}`
      : `slug=${encodeURIComponent(idOrSlug)}`
    return getJSON<WorkItemDetail>(`/api/work-items/get?${q}`)
  },
  sessionGet: (sid: string) =>
    getJSON<SessionDetail>(`/api/sessions/get?id=${encodeURIComponent(sid)}`),
  watchmanPasses: (limit?: number) => {
    const q = limit ? `?limit=${limit}` : ''
    return getJSON<{ items: PassRow[] }>(`/api/watchman/passes${q}`)
  },
  watchmanPass: (passId: string) =>
    getJSON<PassDetailResp>(`/api/watchman/pass?id=${encodeURIComponent(passId)}`),
  bridgeState: () => getJSON<BridgeStateResp>('/api/bridge/state'),
  providers: () => getJSON<ProvidersResp>('/api/providers'),
  graphStudiesCitations: (limit?: number) => {
    const q = limit ? `?limit=${limit}` : ''
    return getJSON<GraphResp>(`/api/graph/studies-citations${q}`)
  },
  workItemCreate: async (req: WorkItemCreateReq): Promise<WorkItemCreateResp> => {
    const r = await fetch('/api/work-items/create', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(req),
    })
    if (!r.ok) {
      let msg = `HTTP ${r.status}`
      try { const b = await r.json(); if (b.error) msg = b.error } catch {}
      throw new Error(msg)
    }
    return r.json()
  },
}

export type WorkItemCreateReq = {
  pipeline: string
  slug?: string
  input?: unknown
  user_input?: string
  actor?: string
  token_budget?: number
  dispatch?: boolean
}

export type WorkItemCreateResp = {
  id: string
  work_queue_id?: number
  dispatched: boolean
  error?: string
}

export type GraphNode = { id: string; label: string; kind?: string }
export type GraphEdge = { source: string; target: string; weight?: number }
export type GraphResp = { nodes: GraphNode[]; edges: GraphEdge[] }


export type PassRow = {
  pass_id: string
  status: string
  trigger?: string
  started_at?: string
  finished_at?: string
  provider?: string
  model?: string
  agent_family?: string
  doc_count_planned: number
  doc_count_done: number
  tokens_in: number
  tokens_out: number
  token_budget?: number
  budget_stopped: boolean
  verdict_counts?: Record<string, number>
}

export type VerdictRow = {
  study_id: string
  verdict: string
  reasoning?: string
  model?: string
  tokens?: number
  actor?: string
  created_at?: string
}

export type PassDetailResp = {
  pass: PassRow
  verdicts: VerdictRow[]
}

export type ServerState = {
  server: string
  transport: string
  enabled: boolean
  last_health_check_at?: string
  last_tools_refresh_at?: string
  active_tools: number
  last_error?: string
  tools?: ToolBrief[]
}

export type ToolBrief = {
  name: string
  description?: string
  active: boolean
  input_schema?: unknown
}

export type BridgeStateResp = {
  servers: ServerState[]
}

export type WorkItemRow = {
  id: string
  slug: string
  pipeline: string
  current_stage: string
  status: string
  actor?: string
  tokens_in: number
  tokens_out: number
  token_budget?: number
  created_at?: string
  updated_at?: string
  completed_at?: string
}

export type WorkItemsListResp = {
  items: WorkItemRow[]
  total: number
}

export type WorkItemDetail = WorkItemRow & {
  input: unknown
  stage_results: unknown
  session_ids?: string[]
  error?: string
}

export type MessageRow = {
  id: number
  role: string
  content: string
  model?: string
  tool_call_id?: string
  tool_calls?: unknown
  finish_reason?: string
  tokens_in?: number
  tokens_out?: number
  reasoning_tokens?: number
  parent_work_id?: number
  created_at?: string
}

export type ChatDispatch = {
  work_id: number
  provider: string
  model?: string
  agent_family?: string
  system_prompt?: string
  tools?: unknown
  messages_count: number
  body_messages?: unknown
  status: string
  created_at?: string
  done_at?: string
}

export type SessionDetail = {
  session_id: string
  messages: MessageRow[]
  dispatches: ChatDispatch[]
  tokens_in: number
  tokens_out: number
}

export type ProviderRow = {
  name: string
  base_url: string
  default_model: string
  kind: string
  has_api_key: boolean
}

export type ProvidersResp = { items: ProviderRow[] }
