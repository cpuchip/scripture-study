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
  // Edit + AI-revise (third proposal mode)
  workItemEditProposal: async (req: {
    id: string
    binding_question?: string
    slug?: string
    pipeline_family_hint?: string
    project_association?: string
    rationale?: string
  }): Promise<{ id: string; message: string }> => {
    const r = await fetch('/api/work-items/edit-proposal', {
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
  workItemReviseWithFeedback: async (id: string, feedback: string): Promise<ReviseResp> => {
    const r = await fetch('/api/work-items/revise-with-feedback', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ id, feedback }),
    })
    if (!r.ok) {
      let msg = `HTTP ${r.status}`
      try { const b = await r.json(); if (b.error) msg = b.error } catch {}
      throw new Error(msg)
    }
    return r.json()
  },
  workItemPendingRevisions: (id: string) =>
    getJSON<PendingRevisionsResp>(`/api/work-items/pending-revisions?id=${encodeURIComponent(id)}`),
  workItemApplyRevision: async (id: string): Promise<WorkItemActionResp> => {
    const r = await fetch('/api/work-items/apply-revision', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ id }),
    })
    if (!r.ok) {
      let msg = `HTTP ${r.status}`
      try { const b = await r.json(); if (b.error) msg = b.error } catch {}
      throw new Error(msg)
    }
    return r.json()
  },
  workItemRejectRevision: async (id: string, reason?: string): Promise<WorkItemActionResp> => {
    const r = await fetch('/api/work-items/reject-revision', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ id, reason }),
    })
    if (!r.ok) {
      let msg = `HTTP ${r.status}`
      try { const b = await r.json(); if (b.error) msg = b.error } catch {}
      throw new Error(msg)
    }
    return r.json()
  },
  workItemRatify: async (id: string): Promise<WorkItemActionResp> => {
    const r = await fetch('/api/work-items/ratify', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ id }),
    })
    if (!r.ok) {
      let msg = `HTTP ${r.status}`
      try { const b = await r.json(); if (b.error) msg = b.error } catch {}
      throw new Error(msg)
    }
    return r.json()
  },
  workItemDispatch: async (id: string): Promise<WorkItemActionResp> => {
    const r = await fetch('/api/work-items/dispatch', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ id }),
    })
    if (!r.ok) {
      let msg = `HTTP ${r.status}`
      try { const b = await r.json(); if (b.error) msg = b.error } catch {}
      throw new Error(msg)
    }
    return r.json()
  },
  workItemCancelProposal: async (id: string, reason?: string): Promise<WorkItemActionResp> => {
    const r = await fetch('/api/work-items/cancel-proposal', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ id, reason }),
    })
    if (!r.ok) {
      let msg = `HTTP ${r.status}`
      try { const b = await r.json(); if (b.error) msg = b.error } catch {}
      throw new Error(msg)
    }
    return r.json()
  },
  workItemsList: (params?: { pipeline?: string; status?: string; origin?: string; project_association?: string; limit?: number }) => {
    const q = new URLSearchParams()
    if (params?.pipeline) q.set('pipeline', params.pipeline)
    if (params?.status) q.set('status', params.status)
    if (params?.origin) q.set('origin', params.origin)
    if (params?.project_association) q.set('project_association', params.project_association)
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
  pipelinesList: () => getJSON<PipelinesListResp>('/api/pipelines/list'),
  workItemSetFileDestination: async (req: SetFileDestinationReq): Promise<SetFileDestinationResp> => {
    const r = await fetch('/api/work-items/set-file-destination', {
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
  workItemMaterializeFile: async (id: string): Promise<MaterializeFileResp> => {
    const r = await fetch('/api/work-items/materialize-file', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ id }),
    })
    if (!r.ok) {
      let msg = `HTTP ${r.status}`
      try { const b = await r.json(); if (b.error) msg = b.error } catch {}
      throw new Error(msg)
    }
    return r.json()
  },
  intentsList: () => getJSON<IntentsListResp>('/api/intents/list'),

  // Projects (Batch I.1)
  projectsList: (includeArchived = false) =>
    getJSON<ProjectsListResp>(`/api/projects/list${includeArchived ? '?include_archived=true' : ''}`),
  projectGet: (slug: string) =>
    getJSON<ProjectRow>(`/api/projects/get?slug=${encodeURIComponent(slug)}`),
  projectCreate: async (req: { slug: string; name: string; description?: string; root_directory?: string }): Promise<{ slug: string; message: string }> => {
    const r = await fetch('/api/projects/create', {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(req),
    })
    if (!r.ok) {
      let msg = `HTTP ${r.status}`
      try { const b = await r.json(); if (b.error) msg = b.error } catch {}
      throw new Error(msg)
    }
    return r.json()
  },
  projectUpdate: async (req: { slug: string; name?: string; description?: string; root_directory?: string }): Promise<{ slug: string; message: string }> => {
    const r = await fetch('/api/projects/update', {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(req),
    })
    if (!r.ok) {
      let msg = `HTTP ${r.status}`
      try { const b = await r.json(); if (b.error) msg = b.error } catch {}
      throw new Error(msg)
    }
    return r.json()
  },
  projectArchive: async (slug: string, archived: boolean): Promise<{ slug: string; archived: boolean; message: string }> => {
    const r = await fetch('/api/projects/archive', {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ slug, archived }),
    })
    if (!r.ok) {
      let msg = `HTTP ${r.status}`
      try { const b = await r.json(); if (b.error) msg = b.error } catch {}
      throw new Error(msg)
    }
    return r.json()
  },
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
  councilsList: (limit?: number) =>
    getJSON<CouncilsListResp>(`/api/councils/list${limit ? '?limit=' + limit : ''}`),
  councilGet: (id: string) =>
    getJSON<CouncilDetail>(`/api/councils/get?id=${encodeURIComponent(id)}`),
  councilSuggestions: (minLessons?: number) =>
    getJSON<CouncilSuggestionsResp>(`/api/councils/suggestions${minLessons ? '?min_lessons=' + minLessons : ''}`),
  councilConvene: async (req: CouncilConveneReq): Promise<CouncilConveneResp> => {
    const r = await fetch('/api/councils/convene', {
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
  councilResolve: async (req: CouncilResolveReq): Promise<CouncilResolveResp> => {
    const r = await fetch('/api/councils/resolve', {
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
  sessionsList: () =>
    getJSON<SessionsListResp>('/api/sessions/list'),
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

  // Models catalog (UI 2026-05-29 — backs Brainstorm datalist + /models view).
  modelsList: () => getJSON<ModelsListResp>('/api/models'),

  // Brainstorm (J.8 + J.9, MCP wrapper 0c1926c, UI batch 2026-05-29).
  brainstormLenses: () => getJSON<BrainstormLensesResp>('/api/brainstorm/lenses'),
  brainstormStart: async (req: BrainstormStartReq): Promise<BrainstormStartResp> => {
    const r = await fetch('/api/brainstorm/start', {
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

export type ModelRow = {
  provider: string
  model: string
  input_micro_per_mtok: number
  output_micro_per_mtok: number
  cache_write_micro_per_mtok?: number
  cache_read_micro_per_mtok?: number
  is_provider_default: boolean
  notes?: string
}

export type ModelsListResp = {
  items: ModelRow[]
  total: number
}

export type BrainstormLensRow = {
  short_name: string
  pipeline_family: string
  description: string
  default_model?: string
  suggested_model?: string
  default_provider?: string
  suggested_provider?: string
  is_original: boolean
}

export type BrainstormLensesResp = {
  items: BrainstormLensRow[]
  total: number
}

export type BrainstormStartReq = {
  binding_question: string
  destination?: string
  slug?: string
  lenses?: string[]
  models?: Record<string, string>
  project_association?: string
  actor?: string
  cost_cap_per_lens_micro?: number
}

export type BrainstormChildRow = {
  id: string
  slug: string
  pipeline_family: string
  model_override?: string
}

export type BrainstormStartResp = {
  parent_id: string
  slug: string
  destination: string
  lenses: string[]
  children: BrainstormChildRow[]
  aggregator_id?: string
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
  file_destination?: string
  project_association?: string
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

// Phase 5g (F.6+F.7) — council types

export type CouncilRow = {
  id: string
  intent_id: string
  intent_slug?: string
  binding_question: string
  convened_at?: string
  convened_by: string
  bishop: string
  status: 'deliberating' | 'synthesizing' | 'awaiting_bishop' | 'resolved' | 'dissolved'
  resolution_id?: string
  dissolved_reason?: string
  resolved_at?: string
}

export type CouncilsListResp = {
  items: CouncilRow[]
  total: number
}

export type CouncilMember = {
  agent_family: string
  role: 'proposer' | 'critic' | 'synthesizer'
  work_id?: number
  response?: string
  completed_at?: string
}

export type CouncilResolution = {
  id: string
  resolved_by: string
  text: string
  promoted_to?: string
  promoted_at?: string
  raw_proposal?: unknown
  resolved_at?: string
}

export type CouncilDetail = CouncilRow & {
  intent_purpose?: string
  members: CouncilMember[]
  resolution?: CouncilResolution
}

export type CouncilMemberSpec = {
  agent_family: string
  role: 'proposer' | 'critic' | 'synthesizer'
  model?: string
}

export type CouncilConveneReq = {
  intent_id: string
  binding_question: string
  members: CouncilMemberSpec[]
  bishop: string
  convened_by?: string
}

export type CouncilConveneResp = {
  id: string
}

export type CouncilResolveReq = {
  council_id: string
  action: 'accept' | 'request_revision' | 'dissolve'
  resolution_text?: string
  destination?: 'study' | 'decisions' | ''
  resolved_by?: string
  dissolved_reason?: string
}

export type CouncilResolveResp = {
  resolution_id: string
}

export type CouncilSuggestion = {
  pipeline_family: string
  current_stage: string
  lesson_count: number
  sample_content: string
}

export type CouncilSuggestionsResp = {
  items: CouncilSuggestion[]
  total: number
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
  // H.3 — origin + project + parent linkage
  origin?: string
  project_association?: string
  parent_work_item_id?: string
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
  // Batch G.4 — file destination + materialization (i3: materialized_at → file_enqueued_at)
  file_destination?: string
  file_enqueued_at?: string
  pipeline_file_template?: string
}

// Batch G.4 — pipeline + file destination types

export type PipelineRow = {
  family: string
  description: string
  sabbath_enabled: boolean
  atonement_enabled: boolean
  file_destination_template?: string
  file_content_jsonpath?: string
}

export type PipelinesListResp = {
  items: PipelineRow[]
  total: number
}

export type SetFileDestinationReq = {
  id: string
  file_destination: string // empty = DB-only
}

export type SetFileDestinationResp = {
  id: string
  file_destination: string
}

export type MaterializeFileResp = {
  pending_file_write_id?: number
  skipped: boolean
  skip_reason?: string
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

export type ProjectRow = {
  slug: string
  name: string
  description?: string
  root_directory?: string
  archived: boolean
  work_item_count: number
  created_at?: string
  updated_at?: string
}

export type ProjectsListResp = {
  items: ProjectRow[]
  total: number
}

export type WorkItemActionResp = {
  id: string
  status?: string
  maturity?: string
  work_queue_id?: number
  message?: string
}

export type ReviseResp = {
  revise_work_item_id: string
  work_queue_id: number
  message: string
}

// Revision JSON shape emitted by the revise stage (partial — only
// changed fields are present).
export type RevisionJSON = {
  binding_question?: string
  rationale?: string
  slug?: string
  pipeline_family_hint?: string
  project_association?: string
}

export type PendingRevisionRow = {
  id: string
  slug: string
  status: string
  maturity: string
  created_at?: string
  completed_at?: string
  cost_micro: number
  feedback?: string
  revision_json?: RevisionJSON
}

export type PendingRevisionsResp = {
  revisions: PendingRevisionRow[]
  count: number
}

export type SessionListItem = {
  session_id: string
  label?: string
  kind: string
  last_active_at?: string
  message_count: number
  assistant_count: number
  cost_total: number
  work_item_id?: string
  work_item_slug?: string
  pipeline_family?: string
  current_stage?: string
  work_item_status?: string
  work_item_active: boolean
}

export type SessionsListResp = {
  sessions: SessionListItem[]
  count: number
}

export type ProviderRow = {
  name: string
  base_url: string
  default_model: string
  kind: string
  has_api_key: boolean
}

export type ProvidersResp = { items: ProviderRow[] }

// PE-C: scheduled_pipelines (cron-style pipeline dispatches).
export type ScheduledRow = {
  id: string
  slug: string
  pipeline_family: string
  intent_id: string
  intent_slug?: string
  cron_pattern: string
  input_template: Record<string, unknown>
  enabled: boolean
  missed_window_hours: number
  last_dispatched_at?: string
  next_due_at?: string
  created_at?: string
  updated_at?: string
  notes?: string
}

export type ScheduledListResp = { items: ScheduledRow[]; total: number }

export type ScheduledCreateReq = {
  slug: string
  pipeline_family: string
  intent_slug: string
  cron_pattern: string
  input_template: Record<string, unknown>
  enabled?: boolean
  missed_window_hours?: number
  notes?: string
}

export type ScheduledUpdateReq = {
  cron_pattern?: string
  input_template?: Record<string, unknown>
  enabled?: boolean
  missed_window_hours?: number
  notes?: string
}

export type ScheduledRunRow = {
  work_item_id: string
  slug: string
  schedule_slug?: string
  pipeline_family: string
  status: string
  current_stage?: string
  created_at?: string
  completed_at?: string
  file_path?: string
}

export type ScheduledRunsResp = { items: ScheduledRunRow[]; total: number }

async function postJSON<T>(path: string, body: unknown): Promise<T> {
  const r = await fetch(path, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
  if (!r.ok) {
    let msg = `HTTP ${r.status}`
    try { const b = await r.json(); if (b?.error) msg = b.error } catch { /* ignore */ }
    throw new Error(msg)
  }
  return r.json() as Promise<T>
}

async function putJSON<T>(path: string, body: unknown): Promise<T> {
  const r = await fetch(path, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
  if (!r.ok) {
    let msg = `HTTP ${r.status}`
    try { const b = await r.json(); if (b?.error) msg = b.error } catch { /* ignore */ }
    throw new Error(msg)
  }
  return r.json() as Promise<T>
}

async function deleteJSON<T>(path: string): Promise<T> {
  const r = await fetch(path, { method: 'DELETE' })
  if (!r.ok) {
    let msg = `HTTP ${r.status}`
    try { const b = await r.json(); if (b?.error) msg = b.error } catch { /* ignore */ }
    throw new Error(msg)
  }
  return r.json() as Promise<T>
}

// Re-export onto api object. Defined as a const block below so the
// closure captures the helpers without polluting the existing api
// declaration order.
export const scheduledApi = {
  list: () => getJSON<ScheduledListResp>('/api/scheduled/list'),
  get: (idOrSlug: { id?: string; slug?: string }) => {
    const p = new URLSearchParams()
    if (idOrSlug.id) p.set('id', idOrSlug.id)
    if (idOrSlug.slug) p.set('slug', idOrSlug.slug)
    return getJSON<ScheduledRow>(`/api/scheduled/get?${p}`)
  },
  create: (req: ScheduledCreateReq) =>
    postJSON<{ id: string; slug: string }>('/api/scheduled/create', req),
  update: (id: string, req: ScheduledUpdateReq) =>
    putJSON<ScheduledRow>(`/api/scheduled/update?id=${encodeURIComponent(id)}`, req),
  toggle: (id: string) =>
    postJSON<{ enabled: boolean }>(`/api/scheduled/toggle?id=${encodeURIComponent(id)}`, {}),
  remove: (id: string) =>
    deleteJSON<{ deleted: string }>(`/api/scheduled/delete?id=${encodeURIComponent(id)}`),
  recentRuns: (limit = 7) =>
    getJSON<ScheduledRunsResp>(`/api/scheduled/recent-runs?limit=${limit}`),
}
