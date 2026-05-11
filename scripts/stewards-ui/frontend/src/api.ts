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
  workItemCost: (id: string) =>
    getJSON<CostEventsResp>(`/api/work-items/cost?id=${encodeURIComponent(id)}`),
  workItemActions: (id: string) =>
    getJSON<StewardActionsResp>(`/api/work-items/actions?id=${encodeURIComponent(id)}`),
  workItemGateDecisions: (id: string) =>
    getJSON<GateDecisionsResp>(`/api/work-items/gate-decisions?id=${encodeURIComponent(id)}`),
  intentsList: () => getJSON<IntentsListResp>('/api/intents/list'),
  intentGet: (idOrSlug: string) => {
    const q = idOrSlug.length === 36 && idOrSlug.includes('-')
      ? `id=${encodeURIComponent(idOrSlug)}`
      : `slug=${encodeURIComponent(idOrSlug)}`
    return getJSON<IntentRow>(`/api/intents/get?${q}`)
  },
  intentCreate: async (req: IntentCreateReq): Promise<IntentCreateResp> => {
    const r = await fetch('/api/intents/create', {
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
  covenantActive: (scope?: string) =>
    getJSON<CovenantRow>(`/api/covenants/active${scope ? '?scope=' + encodeURIComponent(scope) : ''}`),
  covenantsList: () => getJSON<CovenantsListResp>('/api/covenants/list'),
  lessonsList: (params?: { kind?: string; ratified?: 'true' | 'false'; limit?: number }) => {
    const q = new URLSearchParams()
    if (params?.kind) q.set('kind', params.kind)
    if (params?.ratified) q.set('ratified', params.ratified)
    if (params?.limit) q.set('limit', String(params.limit))
    const qs = q.toString()
    return getJSON<LessonsListResp>(`/api/lessons/list${qs ? '?' + qs : ''}`)
  },
  lessonRatify: async (req: LessonRatifyReq): Promise<LessonRatifyResp> => {
    const r = await fetch('/api/lessons/ratify', {
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
  sabbathList: (pipeline?: string) =>
    getJSON<SabbathListResp>(`/api/sabbath/list${pipeline ? '?pipeline=' + encodeURIComponent(pipeline) : ''}`),
  trustScores: () => getJSON<TrustScoresResp>('/api/trust/scores'),
  trustTransitions: (params?: { agent?: string; pipeline?: string; model?: string; limit?: number }) => {
    const q = new URLSearchParams()
    if (params?.agent) q.set('agent', params.agent)
    if (params?.pipeline) q.set('pipeline', params.pipeline)
    if (params?.model) q.set('model', params.model)
    if (params?.limit) q.set('limit', String(params.limit))
    const qs = q.toString()
    return getJSON<TrustTransitionsResp>(`/api/trust/transitions${qs ? '?' + qs : ''}`)
  },
  trustAdjust: async (req: TrustAdjustReq): Promise<TrustAdjustResp> => {
    const r = await fetch('/api/trust/adjust', {
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
  gateOverrideApply: async (req: GateOverrideApplyReq): Promise<GateOverrideApplyResp> => {
    const r = await fetch('/api/gate-overrides/apply', {
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
  destination_maturity?: string
  intent_id?: string
}

// Phase 5d (C.7+C.8) — intent + covenant types

export type IntentValue = {
  key: string
  description: string
  source?: string
  kind?: string
  severity?: string
}

export type IntentRow = {
  id: string
  slug: string
  purpose: string
  beneficiary?: string
  values_hierarchy: IntentValue[]
  non_goals?: string[]
  scripture_anchor?: string
  source_file?: string
  work_item_count: number
  created_at?: string
  updated_at?: string
}

export type IntentsListResp = {
  items: IntentRow[]
  total: number
}

export type IntentCreateReq = {
  slug: string
  purpose: string
  beneficiary?: string
  non_goals?: string[]
  scripture_anchor?: string
}

export type IntentCreateResp = {
  id: string
  slug: string
}

export type CovenantCommit = {
  key: string
  description: string
  why?: string
}

export type CovenantRow = {
  id: string
  scope: string
  human_commits_to: CovenantCommit[]
  agent_commits_to: CovenantCommit[]
  when_broken?: string
  recovery?: string
  council_moment?: string
  teaching_extension?: unknown
  activated_at?: string
  deactivated_at?: string
  ratified_by: string
  source_file?: string
}

export type CovenantsListResp = {
  items: CovenantRow[]
  total: number
}

// Phase 5e (D.6+D.7) — lessons + sabbath types

export type LessonRow = {
  id: number
  work_item_id?: string
  work_item_slug?: string
  at?: string
  kind: 'principle' | 'decision' | 'lesson' | 'sabbath_reflection'
  content: string
  raw_response?: unknown
  ratified_at?: string
  ratified_by?: string
  promoted_to?: string
  work_id?: number
  pipeline_family?: string
  current_stage?: string
}

export type LessonsListResp = {
  items: LessonRow[]
  total: number
}

export type LessonRatifyReq = {
  id: number
  ratified_by: string
  promoted_to?: string
}

export type LessonRatifyResp = {
  id: number
  ratified_at: string
}

export type SabbathRow = {
  id: number
  work_item_id: string
  work_item_slug: string
  pipeline_family: string
  at?: string
  reflection: string
  carry_forward?: string
  surprise?: string
}

export type SabbathListResp = {
  items: SabbathRow[]
  total: number
}

// Phase 5f (E.6+E.7) — trust + gate override types

export type TrustScoreRow = {
  agent_family: string
  pipeline_family: string
  model: string
  successful_completions: number
  failed_completions: number
  human_overrides: number
  trust_level: 'trainee' | 'journeyman' | 'master'
  last_evaluated_at?: string
  last_completion_at?: string
}

export type TrustScoresResp = {
  items: TrustScoreRow[]
  total: number
}

export type TrustTransitionRow = {
  id: number
  at?: string
  agent_family: string
  pipeline_family: string
  model: string
  from_level: string
  to_level: string
  transition_kind: 'auto' | 'manual'
  actor: string
  justification?: string
  metrics?: unknown
}

export type TrustTransitionsResp = {
  items: TrustTransitionRow[]
  total: number
}

export type TrustAdjustReq = {
  agent_family: string
  pipeline_family: string
  model: string
  new_level: 'trainee' | 'journeyman' | 'master'
  actor: string
  justification: string
}

export type TrustAdjustResp = {
  new_level: string
}

export type GateOverrideApplyReq = {
  gate_decision_id: number
  overridden_by: string
  new_action: 'advance' | 'revise' | 'surface'
  justification: string
}

export type GateOverrideApplyResp = {
  new_maturity: string
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
  // Phase 4j — steward + cost surface
  failure_count: number
  last_failure_reason?: string
  last_failure_diagnosis?: string
  quarantined_at?: string
  quarantine_reason?: string
  model_override?: string
  provider_override?: string
  escalation_state: string
  escalation_claimed_by?: string
  escalation_attempts: number
  cost_micro_dollars: number
  cost_cap_micro?: number
  cost_capped_at?: string
  // Phase 5a — maturity ladder surface
  maturity: string
  destination_maturity?: string
  revision_count: number
  scenarios?: unknown
  spec?: string
}

export type GateDecisionRow = {
  id: number
  at?: string
  from_maturity: string
  action: string
  reasoning?: string
  feedback?: string
  work_id?: number
  revision_count: number
  raw_response?: unknown
}

export type GateDecisionsResp = {
  items: GateDecisionRow[]
  count: number
}

export type CostEvent = {
  id: number
  attempt_seq: number
  at?: string
  provider: string
  model: string
  input_tokens: number
  output_tokens: number
  cache_write_tokens: number
  cache_read_tokens: number
  micro_dollars: number
  pricing_effective_at?: string
  notes?: string
}

export type CostEventsResp = {
  items: CostEvent[]
  total_events: number
  total_micro_dollars: number
  work_item_cost_micro: number
  cost_cap_micro?: number
  cost_capped_at?: string
}

export type StewardAction = {
  id: number
  at?: string
  observation: string
  diagnosis?: string
  action: string
  details?: unknown
  model_used?: string
  cost_micro?: number
}

export type StewardActionsResp = {
  items: StewardAction[]
  count: number
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
