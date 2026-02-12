// API client for the Becoming backend

const BASE = '/api'

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  })
  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error(body.error || res.statusText)
  }
  if (res.status === 204) return undefined as T
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
  created_at: string
  completed_at?: string
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
}

export interface DailyDataPoint {
  date: string
  logs: number
  sets: number
  reps: number
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

// --- Practices ---

export const api = {
  // Practices
  listPractices(type?: string, active = true) {
    const params = new URLSearchParams()
    if (type) params.set('type', type)
    if (!active) params.set('active', 'false')
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
}
