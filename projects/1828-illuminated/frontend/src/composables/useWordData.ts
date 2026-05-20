// Centralized data composable. The JSON bundles are imported statically so
// they're code-split per route by Vite.
import { computed, ref } from 'vue'

import tierWords from '../data/tier-words.json'
import defs1828 from '../data/definitions-1828.json'
// definitions-modern.json may not exist yet during early dev (pre-fetch).
// Use a try/import-fallback pattern via dynamic require — but in Vite, a
// missing static import errors. Solution: ship an empty stub at build time
// and let fetch_modern_defs.py overwrite it. See scripts/build_data.py.
import defsModern from '../data/definitions-modern.json'

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
export interface ModernRecord { entries?: ModernEntry[]; error?: string }

const tierWordList: TierWord[] = (tierWords as any).words
const tierCounts: Record<Tier, number> = (tierWords as any).tier_counts

const tierMap = new Map<string, TierWord>()
for (const tw of tierWordList) tierMap.set(tw.word, tw)

const def1828Map: Record<string, Def1828Entry[]> = (defs1828 as any).definitions ?? {}
const defModernMap: Record<string, ModernRecord | null> = (defsModern as any).definitions ?? {}

export function useWordData() {
  return {
    tierWords: tierWordList,
    tierCounts,
    findWord(word: string): TierWord | undefined {
      return tierMap.get(word.toLowerCase())
    },
    get1828(word: string): Def1828Entry[] {
      return def1828Map[word.toLowerCase()] ?? []
    },
    getModern(word: string): ModernRecord | null {
      const v = defModernMap[word.toLowerCase()]
      return v ?? null
    },
    hasModern(word: string): boolean {
      return defModernMap[word.toLowerCase()] != null
    },
    /** Return all tier words sorted by tier (A++ first), then alphabetical. */
    allByTier(): TierWord[] {
      const order: Record<Tier, number> = { 'A++': 0, 'A+': 1, B: 2, C: 3, D: 4 }
      return [...tierWordList].sort((a, b) => order[a.tier] - order[b.tier] || a.word.localeCompare(b.word))
    },
    /** Fuzzy-ish prefix match for the search box. */
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
// Given a chunk of text (a pasted verse, a demo verse), tokenize into
// words and return a list of {text, word?, tier?} segments so a Vue
// template can render them with v-for + highlight class.

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
    const tw = tierMap.get(norm)
    if (tw) {
      segs.push({ text: raw, word: norm, tier: tw.tier })
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
}
