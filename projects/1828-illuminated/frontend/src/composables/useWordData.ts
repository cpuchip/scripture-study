// Centralized word-data composable. Backend cutover (phase 5):
//
//   - tier-words.json + manual-additions.json stay statically imported.
//     They're small (~342KB combined) and drive:
//       (1) highlight tier display + tokenize() inline highlighting
//       (2) the WordSearch primary list ("Words we've curated")
//       (3) tier badges + study cross-refs on WordCard
//
//   - 1828 and modern definitions now come from the i1828 backend
//     (`/api/dict/1828/:word`, `/api/dict/modern/:word`). Lookups are
//     async; a small in-memory LRU avoids repeated round-trips per session.
//
//   - Server-side stem fallback (D-DICT-2): the backend's /api/dict/1828
//     handler does its own archaic-suffix stripping and returns
//     `stem_matched` in the response. The client no longer duplicates that
//     logic — we still keep a small tier-side stemMatch for tokenize()'s
//     highlight pass (tier matching is in-memory and synchronous), but
//     definition lookup hands off to the server.

import { computed, ref } from 'vue'

import tierWords from '../data/tier-words.json'
import manualAdditions from '../data/manual-additions.json'

import { apiUrl } from './useApiBase'

export type Tier = 'A++' | 'A+' | 'B' | 'C' | 'D'

export interface TierWord {
  word: string
  tier: Tier
  study_tier: 'A' | 'B' | 'C' | null
  studies: string[]
  study_excerpts: string[]
  p4_score: number | null
  p4_reasons: string[]
}

export interface Def1828Entry { pos: string; definitions: string[] }
export interface ModernEntry { pos: string; definitions: string[] }

/** Result of /api/dict/1828/{word}. `found:false` is a 200-with-no-entry. */
export interface Def1828Response {
  word: string
  entries: Def1828Entry[]
  found: boolean
  stem_matched: string | null
}

/** Result of /api/dict/modern/{word}. `found:false` distinguishes
 *  "cached 404" (Free Dictionary returned no entry) from "looking up". */
export interface ModernResponse {
  word: string
  entries?: ModernEntry[]
  source: 'cache' | 'fetched' | 'none' | 'rate_limited'
  found: boolean
  error?: string | null
}

/** Result of /api/dict/search?q=… — both lists for class-E reach UX. */
export interface DictSearchHit {
  word: string
  tier?: string
}
export interface DictSearchResponse {
  query: string
  tier_results: DictSearchHit[]
  all_1828_results: DictSearchHit[]
}

const autoTierWords: TierWord[] = (tierWords as any).words
const manualTierWords: TierWord[] = (manualAdditions as any).additions ?? []

// Merge: manual additions override the auto list (and are added if not present).
const tierWordList: TierWord[] = [...autoTierWords]
const tierMap = new Map<string, TierWord>()
for (const tw of autoTierWords) tierMap.set(tw.word, tw)
for (const m of manualTierWords) {
  const existingIdx = tierWordList.findIndex(t => t.word === m.word)
  if (existingIdx >= 0) tierWordList[existingIdx] = m
  else tierWordList.push(m)
  tierMap.set(m.word, m)
}

// Recompute tier counts after merge
const tierCounts: Record<Tier, number> = (() => {
  const counts: Record<string, number> = { 'A++': 0, 'A+': 0, B: 0, C: 0, D: 0 }
  for (const w of tierWordList) counts[w.tier] = (counts[w.tier] ?? 0) + 1
  return counts as Record<Tier, number>
})()

// Tier-side stem matcher for tokenize()'s highlight pass. Keeps tokenize
// synchronous so HighlightedText renders without await. Definition lookup
// (1828/modern) is async + server-stem-fallback (the source of truth).
const ARCHAIC_SUFFIXES = ['eth', 'edst', 'est', 'ing', 'ed', 's']
function tierStemMatch(raw: string): { canonical: string; tw: TierWord } | null {
  const direct = tierMap.get(raw)
  if (direct) return { canonical: raw, tw: direct }
  for (const suf of ARCHAIC_SUFFIXES) {
    if (raw.length > suf.length + 2 && raw.endsWith(suf)) {
      const stem = raw.slice(0, -suf.length)
      const tw = tierMap.get(stem)
      if (tw) return { canonical: stem, tw }
      // Some -ing / -ed forms double the final consonant ("running" → "run")
      if ((suf === 'ing' || suf === 'ed') && stem.length >= 3) {
        const stem2 = stem.slice(0, -1)
        const tw2 = tierMap.get(stem2)
        if (tw2) return { canonical: stem2, tw: tw2 }
      }
      // "-eth" forms also try the -e-less stem ("loveth" → "love")
      if (suf === 'eth') {
        const stem3 = stem + 'e'
        const tw3 = tierMap.get(stem3)
        if (tw3) return { canonical: stem3, tw: tw3 }
      }
    }
  }
  return null
}

// ─── Per-session LRU caches for backend lookups ────────────────────────
//
// 1828 entries are nearly immutable; modern entries can flip from
// "not yet fetched" to "fetched" within a session, but once cached they're
// stable. A 200-entry cap keeps memory bounded.

const CACHE_CAP = 200

function lruGet<V>(cache: Map<string, V>, key: string): V | undefined {
  const v = cache.get(key)
  if (v !== undefined) {
    // Refresh recency: re-insert.
    cache.delete(key)
    cache.set(key, v)
  }
  return v
}

function lruSet<V>(cache: Map<string, V>, key: string, value: V) {
  if (cache.has(key)) cache.delete(key)
  cache.set(key, value)
  while (cache.size > CACHE_CAP) {
    // Evict oldest.
    const firstKey = cache.keys().next().value
    if (firstKey === undefined) break
    cache.delete(firstKey)
  }
}

const cache1828 = new Map<string, Def1828Response>()
const cacheModern = new Map<string, ModernResponse>()

// ─── Public API ───────────────────────────────────────────────────────

export function useWordData() {
  return {
    tierWords: tierWordList,
    tierCounts,
    findWord(word: string): TierWord | undefined {
      return tierMap.get(word.toLowerCase())
    },

    /**
     * Async 1828 lookup. Returns the backend's `Def1828Response` shape so
     * callers can render `stem_matched` ("Showing entry for `obtain`").
     * Network errors return `{found:false}` with a synthetic error string —
     * callers should not throw on a missing word; they should render the
     * "no entry" state.
     */
    async get1828(word: string): Promise<Def1828Response> {
      const w = word.toLowerCase()
      const cached = lruGet(cache1828, w)
      if (cached) return cached
      try {
        const resp = await fetch(apiUrl(`/dict/1828/${encodeURIComponent(w)}`))
        if (!resp.ok) {
          return { word: w, entries: [], found: false, stem_matched: null }
        }
        const json = (await resp.json()) as Def1828Response
        lruSet(cache1828, w, json)
        return json
      } catch {
        return { word: w, entries: [], found: false, stem_matched: null }
      }
    },

    /**
     * Async modern lookup. The backend handles lazy fetch + write-back
     * against the Free Dictionary API at 1 req/sec global.
     */
    async getModern(word: string): Promise<ModernResponse> {
      const w = word.toLowerCase()
      const cached = lruGet(cacheModern, w)
      if (cached) return cached
      try {
        const resp = await fetch(apiUrl(`/dict/modern/${encodeURIComponent(w)}`))
        if (!resp.ok) {
          return { word: w, source: 'none', found: false, error: `HTTP ${resp.status}` }
        }
        const json = (await resp.json()) as ModernResponse
        lruSet(cacheModern, w, json)
        return json
      } catch (e: unknown) {
        const msg = e instanceof Error ? e.message : String(e)
        return { word: w, source: 'none', found: false, error: msg }
      }
    },

    /**
     * Class-E reach search: returns tier matches AND all-1828-corpus
     * matches in one round-trip. The frontend renders them as separate
     * sections so the headline backend win ("any word in the 1828 is
     * reachable") is visible at the UX layer.
     */
    async searchDict(query: string, limit = 40): Promise<DictSearchResponse> {
      const q = query.trim().toLowerCase()
      const empty: DictSearchResponse = { query: q, tier_results: [], all_1828_results: [] }
      if (!q) return empty
      try {
        const resp = await fetch(apiUrl(`/dict/search?q=${encodeURIComponent(q)}&limit=${limit}`))
        if (!resp.ok) return empty
        return (await resp.json()) as DictSearchResponse
      } catch {
        return empty
      }
    },

    /** Return all tier words sorted by tier (A++ first), then alphabetical. */
    allByTier(): TierWord[] {
      const order: Record<Tier, number> = { 'A++': 0, 'A+': 1, B: 2, C: 3, D: 4 }
      return [...tierWordList].sort((a, b) => order[a.tier] - order[b.tier] || a.word.localeCompare(b.word))
    },

    /**
     * Synchronous prefix match against the tier list — drives the
     * primary "Words we've curated" section in WordSearch.vue. The
     * class-E reach (all 98k 1828 headwords) is surfaced separately via
     * `searchDict()` above.
     */
    searchPrefix(query: string, limit = 40): TierWord[] {
      const q = query.trim().toLowerCase()
      if (!q) return []
      const starts: TierWord[] = []
      const contains: TierWord[] = []
      for (const tw of tierWordList) {
        if (tw.word === q) {
          starts.unshift(tw)
        } else if (tw.word.startsWith(q)) {
          starts.push(tw)
        } else if (tw.word.includes(q)) {
          contains.push(tw)
        }
      }
      return [...starts, ...contains].slice(0, limit)
    },
  }
}

// ─── Verse tokenization + highlighting ─────────────────────────────────────
//
// Given a chunk of text, tokenize into segments so a Vue template can render
// it with v-for + highlight class. Stays synchronous — tier matching is in
// memory; the dictionary round-trip happens later when the user clicks a
// highlighted word.

export interface TextSegment {
  text: string
  word?: string  // canonical (lowercase) form if this segment matches a tier word
  tier?: Tier
}

const wordRe = /[A-Za-z][A-Za-z'-]+/g

export function tokenize(text: string): TextSegment[] {
  const segs: TextSegment[] = []
  let lastIdx = 0
  for (const m of text.matchAll(wordRe)) {
    const idx = m.index!
    if (idx > lastIdx) {
      segs.push({ text: text.slice(lastIdx, idx) })
    }
    const raw = m[0]
    const norm = raw.toLowerCase().replace(/[''-]+$/, '').replace(/^[''-]+/, '')
    const match = tierStemMatch(norm)
    if (match) {
      // Use the canonical form for the word card so we don't pop different
      // cards for "suffer" vs "suffereth" vs "suffered".
      segs.push({ text: raw, word: match.canonical, tier: match.tw.tier })
    } else {
      segs.push({ text: raw })
    }
    lastIdx = idx + raw.length
  }
  if (lastIdx < text.length) segs.push({ text: text.slice(lastIdx) })
  return segs
}

// Reactive "currently selected word" — shared so any component can hover/click to update.
export const selectedWord = ref<string | null>(null)
export const selectedTier = computed(() => {
  if (!selectedWord.value) return null
  return tierMap.get(selectedWord.value)?.tier ?? null
})
export function selectWord(word: string | null) {
  selectedWord.value = word
  // Capture as a study-tree node so clicks within any rendered passage
  // (canon, demo, paste) branch off the currently-active node. The tree's
  // own idempotency dedupes if the active node is already this word.
  if (word) {
    // Lazy import to avoid circular module init — useStudyTree imports nothing
    // from this file, but bundlers vary.
    void import('./useStudyTree').then(({ visit }) => {
      visit({ kind: 'word', word })
    })
  }
}
