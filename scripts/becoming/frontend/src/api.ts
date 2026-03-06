// API client for the Becoming backend

const BASE = '/api'

// Global 401 handler — set by the auth composable
let onUnauthorized: (() => void) | null = null
export function setUnauthorizedHandler(handler: () => void) {
  onUnauthorized = handler
}

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    headers: { 'Content-Type': 'application/json' },
    credentials: 'same-origin',
    ...options,
  })
  if (!res.ok) {
    if (res.status === 401 && onUnauthorized) {
      onUnauthorized()
    }
    const body = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error(body.error || res.statusText)
  }
  if (res.status === 204) return undefined as T
  return res.json()
}

// Auth requests use a different base path (not under /api)
async function authRequest<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(path, {
    headers: { 'Content-Type': 'application/json' },
    credentials: 'same-origin',
    ...options,
  })
  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error(body.error || res.statusText)
  }
  return res.json()
}

// --- Types ---

export interface Practice {
  id: number
  name: string
  description: string
  type: 'memorize' | 'tracker' | 'habit' | 'task' | 'scheduled'
  category: string
  source_doc: string
  source_path: string
  config: string
  sort_order: number
  active: boolean
  status: 'active' | 'paused' | 'completed' | 'archived'
  created_at: string
  completed_at?: string
  archived_at?: string
  end_date?: string
  start_date?: string
  memorize_level?: number
}

export interface PracticeLog {
  id: number
  practice_id: number
  logged_at: string
  date: string
  quality?: number
  value?: string
  sets?: number
  reps?: number
  duration_s?: number
  notes?: string
  next_review?: string
}

export interface DailySummary {
  practice_id: number
  practice_name: string
  practice_type: string
  category: string
  config: string
  status: string
  end_date?: string
  start_date?: string
  created_at: string
  log_count: number
  total_sets?: number
  total_reps?: number
  last_value: string
  last_notes: string
  // Schedule-aware fields (populated for "scheduled" type)
  is_due?: boolean
  next_due?: string
  days_overdue?: number
  slots_due?: string[]
}

export interface Task {
  id: number
  title: string
  description: string
  source_doc: string
  source_section: string
  scripture: string
  type: string
  status: string
  brain_entry_id?: string
  created_at: string
  completed_at: string
}

export interface ScriptureVerse {
  number: number
  text: string
  reference: string
}

export interface ScriptureLookup {
  reference: string
  book: string
  chapter: number
  verses: ScriptureVerse[]
  path: string
}

export interface ScriptureBook {
  name: string
  volume: string
  slug: string
  path: string
}

export interface ScriptureVolume {
  name: string
  slug: string
  books: ScriptureBook[]
}

export interface MemorizeCardStatus {
  practice: Practice
  reviews_today: number
  today_qualities: number[]
  is_due: boolean
  target_daily_reps: number
  aptitudes: MemorizeAptitude[]
  overall_aptitude: number
  is_mastered: boolean
  days_until_end: number | null
}

// Study mode types
export type StudyMode = 'reveal_whole' | 'reveal_words' | 'type_words' | 'arrange' | 'type_full'
  | 'reverse_full' | 'reverse_partial' | 'reverse_fragment'

export type SessionMomentum = 'struggling' | 'steady' | 'cruising'

export interface StudyExercise {
  practice: Practice
  mode: StudyMode
  is_reverse: boolean
  level: number
  momentum: SessionMomentum
  card_type: 'goldilocks' | 'stretch' | 'confidence' | 'fresh'
  all_card_names?: string[]
  done?: boolean
  message?: string
}

export interface MemorizeAptitude {
  id: number
  practice_id: number
  user_id: number
  mode: string
  aptitude: number
  sample_count: number
  last_score_at?: string
}

export interface StudyScoreResponse {
  score: any
  aptitudes: MemorizeAptitude[]
  overall: number
}

export interface DailyDataPoint {
  date: string
  logs: number
  sets: number
  reps: number
}

export interface ActivityDay {
  date: string
  log_count: number
  practice_count: number
}

export interface ReportEntry {
  practice_id: number
  practice_name: string
  practice_type: string
  category: string
  config: string
  total_logs: number
  total_sets: number
  total_reps: number
  days_active: number
  days_in_range: number
  completion_rate: number
  current_streak: number
  daily_data: DailyDataPoint[]
}

export interface Note {
  id: number
  content: string
  practice_id?: number | null
  task_id?: number | null
  pillar_id?: number | null
  pinned: boolean
  created_at: string
  updated_at: string
  practice_name?: string
  task_title?: string
}

export interface Prompt {
  id: number
  text: string
  active: boolean
  sort_order: number
  created_at: string
}

export interface Reflection {
  id: number
  date: string
  prompt_id?: number | null
  prompt_text?: string
  content: string
  mood?: number | null
  created_at: string
  updated_at: string
}

export interface Pillar {
  id: number
  name: string
  description: string
  icon: string
  parent_id?: number | null
  children?: Pillar[]
  practice_count?: number
  task_count?: number
  created_at: string
}

export interface PillarLink {
  pillar_id: number
  pillar_name: string
  pillar_icon: string
}

export interface PracticePillarMapping {
  practice_id: number
  pillar_id: number
  pillar_name: string
  pillar_icon: string
}

// --- Practices ---

export const api = {
  // Practices
  listPractices(type?: string, active = true, status?: string) {
    const params = new URLSearchParams()
    if (type) params.set('type', type)
    if (status) params.set('status', status)
    else if (!active) params.set('active', 'false')
    return request<Practice[]>(`/practices?${params}`)
  },

  getPractice(id: number) {
    return request<Practice>(`/practices/${id}`)
  },

  createPractice(p: Partial<Practice>) {
    return request<Practice>('/practices', { method: 'POST', body: JSON.stringify(p) })
  },

  updatePractice(id: number, p: Partial<Practice>) {
    return request<Practice>(`/practices/${id}`, { method: 'PUT', body: JSON.stringify(p) })
  },

  deletePractice(id: number) {
    return request<void>(`/practices/${id}`, { method: 'DELETE' })
  },

  // Practice lifecycle
  completePractice(id: number) {
    return request<Practice>(`/practices/${id}/complete`, { method: 'POST' })
  },

  archivePractice(id: number) {
    return request<Practice>(`/practices/${id}/archive`, { method: 'POST' })
  },

  pausePractice(id: number) {
    return request<Practice>(`/practices/${id}/pause`, { method: 'POST' })
  },

  restorePractice(id: number) {
    return request<Practice>(`/practices/${id}/restore`, { method: 'POST' })
  },

  // Logs
  createLog(log: Partial<PracticeLog>) {
    return request<PracticeLog>('/logs', { method: 'POST', body: JSON.stringify(log) })
  },

  deleteLog(id: number) {
    return request<void>(`/logs/${id}`, { method: 'DELETE' })
  },

  deleteLatestLog(practiceId: number, date: string) {
    return request<void>(`/logs/latest?practice_id=${practiceId}&date=${date}`, { method: 'DELETE' })
  },

  listPracticeLogs(practiceId: number, limit = 100) {
    return request<PracticeLog[]>(`/practices/${practiceId}/logs?limit=${limit}`)
  },

  listPracticeLogsRange(practiceId: number, start: string, end: string) {
    return request<PracticeLog[]>(`/practices/${practiceId}/logs?start=${start}&end=${end}`)
  },

  // Daily
  getDailySummary(date: string) {
    return request<DailySummary[]>(`/daily/${date}`)
  },

  // Tasks
  listTasks(status?: string) {
    const params = status ? `?status=${status}` : ''
    return request<Task[]>(`/tasks${params}`)
  },

  createTask(t: Partial<Task>) {
    return request<Task>('/tasks', { method: 'POST', body: JSON.stringify(t) })
  },

  updateTask(id: number, t: Partial<Task>) {
    return request<Task>(`/tasks/${id}`, { method: 'PUT', body: JSON.stringify(t) })
  },

  deleteTask(id: number) {
    return request<void>(`/tasks/${id}`, { method: 'DELETE' })
  },

  // Memorize
  getDueCards(date: string) {
    return request<Practice[]>(`/memorize/due/${date}`)
  },

  getMemorizeCards(date: string) {
    return request<MemorizeCardStatus[]>(`/memorize/cards/${date}`)
  },

  reviewCard(practiceId: number, quality: number, date: string) {
    return request<Practice>('/memorize/review', {
      method: 'POST',
      body: JSON.stringify({ practice_id: practiceId, quality, date }),
    })
  },

  // Study mode (adaptive difficulty)
  studyNext(opts: {
    date: string
    category?: string
    pillarIds?: number[]
    lastCardId?: number
    momentum?: SessionMomentum
    recentScores?: number[]
    mode?: 'due' | 'all'
  }) {
    const params = new URLSearchParams({ date: opts.date })
    if (opts.category) params.set('category', opts.category)
    if (opts.pillarIds?.length) params.set('pillar_ids', opts.pillarIds.join(','))
    if (opts.lastCardId) params.set('last_card_id', String(opts.lastCardId))
    if (opts.momentum) params.set('momentum', opts.momentum)
    if (opts.recentScores?.length) params.set('recent_scores', opts.recentScores.join(','))
    if (opts.mode) params.set('mode', opts.mode)
    return request<StudyExercise>(`/memorize/study/next?${params}`)
  },

  studyScore(score: {
    practice_id: number
    mode: string
    score: number
    quality?: number
    duration_s?: number
    date: string
  }) {
    return request<StudyScoreResponse>('/memorize/study/score', {
      method: 'POST',
      body: JSON.stringify(score),
    })
  },

  studyAptitudes(practiceId: number) {
    return request<{ aptitudes: MemorizeAptitude[]; overall: number }>(
      `/memorize/study/aptitudes/${practiceId}`
    )
  },

  studySeed() {
    return request<{ status: string }>('/memorize/study/seed', { method: 'POST' })
  },

  // Scripture lookup
  lookupScripture(ref: string) {
    return request<ScriptureLookup>(`/scriptures/lookup?ref=${encodeURIComponent(ref)}`)
  },

  listScriptureBooks() {
    return request<ScriptureVolume[]>('/scriptures/books')
  },

  searchScriptureBooks(query: string) {
    return request<ScriptureBook[]>(`/scriptures/search?q=${encodeURIComponent(query)}`)
  },

  // Reports
  getReport(start: string, end: string) {
    return request<ReportEntry[]>(`/reports?start=${start}&end=${end}`)
  },

  getActivityHeatmap(start: string, end: string) {
    return request<ActivityDay[]>(`/reports/activity?start=${start}&end=${end}`)
  },

  // Notes
  listNotes(filters?: { practice_id?: number; task_id?: number; pillar_id?: number; pinned?: boolean }) {
    const params = new URLSearchParams()
    if (filters?.practice_id) params.set('practice_id', String(filters.practice_id))
    if (filters?.task_id) params.set('task_id', String(filters.task_id))
    if (filters?.pillar_id) params.set('pillar_id', String(filters.pillar_id))
    if (filters?.pinned) params.set('pinned', 'true')
    const qs = params.toString()
    return request<Note[]>(`/notes${qs ? '?' + qs : ''}`)
  },

  createNote(n: Partial<Note>) {
    return request<Note>('/notes', { method: 'POST', body: JSON.stringify(n) })
  },

  updateNote(id: number, n: Partial<Note>) {
    return request<Note>(`/notes/${id}`, { method: 'PUT', body: JSON.stringify(n) })
  },

  deleteNote(id: number) {
    return request<void>(`/notes/${id}`, { method: 'DELETE' })
  },

  // Prompts
  listPrompts(activeOnly = true) {
    const params = activeOnly ? '' : '?active=false'
    return request<Prompt[]>(`/prompts${params}`)
  },

  getTodayPrompt() {
    return request<Prompt>('/prompts/today')
  },

  createPrompt(p: Partial<Prompt>) {
    return request<Prompt>('/prompts', { method: 'POST', body: JSON.stringify(p) })
  },

  updatePrompt(id: number, p: Partial<Prompt>) {
    return request<Prompt>(`/prompts/${id}`, { method: 'PUT', body: JSON.stringify(p) })
  },

  deletePrompt(id: number) {
    return request<void>(`/prompts/${id}`, { method: 'DELETE' })
  },

  // Reflections
  listReflections(from?: string, to?: string) {
    const params = new URLSearchParams()
    if (from) params.set('from', from)
    if (to) params.set('to', to)
    const qs = params.toString()
    return request<Reflection[]>(`/reflections${qs ? '?' + qs : ''}`)
  },

  getReflection(date: string) {
    return request<Reflection | null>(`/reflections/${date}`)
  },

  upsertReflection(r: Partial<Reflection>) {
    return request<Reflection>('/reflections', { method: 'POST', body: JSON.stringify(r) })
  },

  deleteReflection(id: number) {
    return request<void>(`/reflections/${id}`, { method: 'DELETE' })
  },

  // Pillars
  listPillarsTree() {
    return request<Pillar[]>('/pillars')
  },

  listPillarsFlat() {
    return request<Pillar[]>('/pillars/flat')
  },

  getPillarSuggestions() {
    return request<Pillar[]>('/pillars/suggestions')
  },

  hasPillars() {
    return request<{ has_pillars: boolean }>('/pillars/has-pillars')
  },

  createPillar(p: Partial<Pillar>) {
    return request<Pillar>('/pillars', { method: 'POST', body: JSON.stringify(p) })
  },

  getPillar(id: number) {
    return request<Pillar>(`/pillars/${id}`)
  },

  updatePillar(id: number, p: Partial<Pillar>) {
    return request<Pillar>(`/pillars/${id}`, { method: 'PUT', body: JSON.stringify(p) })
  },

  deletePillar(id: number) {
    return request<void>(`/pillars/${id}`, { method: 'DELETE' })
  },

  setPracticePillars(practiceId: number, pillarIds: number[]) {
    return request<void>(`/practices/${practiceId}/pillars`, {
      method: 'PUT',
      body: JSON.stringify({ pillar_ids: pillarIds }),
    })
  },

  getPracticePillars(practiceId: number) {
    return request<PillarLink[]>(`/practices/${practiceId}/pillars`)
  },

  getAllPracticePillarLinks() {
    return request<PracticePillarMapping[]>('/practice-pillar-links')
  },

  // --- Document Sources ---
  listSources() {
    return request<DocumentSource[]>('/sources')
  },
  createSource(source: Partial<DocumentSource>) {
    return request<DocumentSource>('/sources', {
      method: 'POST',
      body: JSON.stringify(source),
    })
  },
  getSource(id: number) {
    return request<DocumentSource>(`/sources/${id}`)
  },
  updateSource(id: number, source: Partial<DocumentSource>) {
    return request<DocumentSource>(`/sources/${id}`, {
      method: 'PUT',
      body: JSON.stringify(source),
    })
  },
  deleteSource(id: number) {
    return request<void>(`/sources/${id}`, { method: 'DELETE' })
  },
  updateSourceTreeCache(id: number, treeJSON: string, etag: string) {
    return request<void>(`/sources/${id}/tree-cache`, {
      method: 'PUT',
      body: JSON.stringify({ tree_json: treeJSON, etag }),
    })
  },

  // --- Reading Progress ---
  listReadingProgress(sourceId: number) {
    return request<ReadingProgress[]>(`/reading-progress?source_id=${sourceId}`)
  },
  upsertReadingProgress(sourceId: number, filePath: string, scrollPct: number) {
    return request<void>('/reading-progress', {
      method: 'POST',
      body: JSON.stringify({ source_id: sourceId, file_path: filePath, scroll_pct: scrollPct }),
    })
  },

  // --- Bookmarks ---
  listBookmarks(sourceId?: number, filePath?: string) {
    const params = new URLSearchParams()
    if (sourceId) params.set('source_id', String(sourceId))
    if (filePath) params.set('file_path', filePath)
    const qs = params.toString()
    return request<Bookmark[]>(`/bookmarks${qs ? '?' + qs : ''}`)
  },
  createBookmark(b: { source_id: number; file_path: string; anchor?: string; excerpt?: string; note?: string }) {
    return request<Bookmark>('/bookmarks', {
      method: 'POST',
      body: JSON.stringify(b),
    })
  },
  updateBookmarkNote(id: number, note: string) {
    return request<void>(`/bookmarks/${id}`, {
      method: 'PATCH',
      body: JSON.stringify({ note }),
    })
  },
  deleteBookmark(id: number) {
    return request<void>(`/bookmarks/${id}`, {
      method: 'DELETE',
    })
  },
}

// --- Document Source Types ---

export interface DocumentSource {
  id: number
  user_id: number
  name: string
  source_type: 'github_public' | 'github_private'
  repo: string
  branch: string
  include_paths: string  // JSON array
  exclude_paths: string  // JSON array
  tree_cache?: string
  tree_etag?: string
  tree_cached_at?: string
  created_at: string
}

export interface ReadingProgress {
  id: number
  user_id: number
  source_id: number
  file_path: string
  read_at: string
  scroll_pct: number
}

export interface SharedLink {
  id: number
  code: string
  user_id?: number
  source_id?: number
  provider: string
  repo: string
  branch: string
  doc_filter: string
  file_path?: string
  hits: number
  created_at: string
}

export interface Bookmark {
  id: number
  source_id: number
  file_path: string
  anchor: string
  excerpt: string
  note: string
  created_at: string
  source_name?: string
}

// --- Public API (no auth required) ---

async function publicRequest<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}/public${path}`, {
    headers: { 'Content-Type': 'application/json' },
    credentials: 'same-origin',
    ...options,
  })
  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error(body.error || res.statusText)
  }
  if (res.status === 204) return undefined as T
  return res.json()
}

export const publicApi = {
  resolveShareLink(code: string) {
    return publicRequest<SharedLink>(`/share/${code}`)
  },
  createShareLink(params: { repo: string; branch?: string; doc_filter?: string; file_path?: string; source_id?: number }) {
    return publicRequest<SharedLink>('/share', {
      method: 'POST',
      body: JSON.stringify(params),
    })
  },
}

// --- Auth Types ---

export interface User {
  id: number
  email: string
  name: string
  avatar_url: string
  provider: string
  has_password: boolean
  google_linked: boolean
  brain_enabled: boolean
  created_at: string
  last_login: string
}

export interface AuthProviders {
  email: boolean
  google: boolean
}

export interface APIToken {
  id: number
  name: string
  prefix: string
  created_at: string
  last_used?: string
  expires_at?: string
  brain_enabled: boolean
}

export interface SessionInfo {
  id: string
  user_agent: string
  ip_address: string
  created_at: string
  last_active: string
  is_current: boolean
}

// --- Auth API ---

export const authApi = {
  register(email: string, password: string, name?: string) {
    return authRequest<{ user: User }>('/auth/register', {
      method: 'POST',
      body: JSON.stringify({ email, password, name }),
    })
  },

  login(email: string, password: string) {
    return authRequest<{ user: User }>('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    })
  },

  logout() {
    return authRequest<{ status: string }>('/auth/logout', { method: 'POST' })
  },

  providers() {
    return request<AuthProviders>('/auth/providers')
  },

  me() {
    return request<User>('/me')
  },

  updateMe(name: string) {
    return request<User>('/me', {
      method: 'PUT',
      body: JSON.stringify({ name }),
    })
  },

  listTokens() {
    return request<APIToken[]>('/tokens')
  },

  createToken(name: string) {
    return request<APIToken & { token: string }>('/tokens', {
      method: 'POST',
      body: JSON.stringify({ name }),
    })
  },

  deleteToken(id: number) {
    return request<{ status: string }>(`/tokens/${id}`, { method: 'DELETE' })
  },

  updateToken(id: number, updates: { brain_enabled?: boolean }) {
    return request<{ status: string }>(`/tokens/${id}`, {
      method: 'PATCH',
      body: JSON.stringify(updates),
    })
  },

  // Password
  changePassword(currentPassword: string, newPassword: string) {
    return request<{ status: string }>('/me/password', {
      method: 'PUT',
      body: JSON.stringify({ current_password: currentPassword, new_password: newPassword }),
    })
  },

  // Google account linking
  unlinkGoogle() {
    return request<{ status: string }>('/me/google', { method: 'DELETE' })
  },

  // Sessions
  listSessions() {
    return request<SessionInfo[]>('/sessions')
  },

  revokeSession(id: string) {
    return request<{ status: string }>(`/sessions/${id}`, { method: 'DELETE' })
  },

  revokeOtherSessions() {
    return request<{ status: string }>('/sessions', { method: 'DELETE' })
  },

  // Account
  deleteAccount(password: string) {
    return request<{ status: string }>('/me', {
      method: 'DELETE',
      body: JSON.stringify({ password }),
    })
  },

  // Export
  async exportData() {
    const res = await fetch(`${BASE}/export`, { credentials: 'same-origin' })
    if (!res.ok) throw new Error('Export failed')
    const blob = await res.blob()
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `becoming-export-${new Date().toISOString().slice(0, 10)}.json`
    a.click()
    URL.revokeObjectURL(url)
  },
}
