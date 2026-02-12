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
  type: 'memorize' | 'tracker' | 'habit' | 'task'
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
}
