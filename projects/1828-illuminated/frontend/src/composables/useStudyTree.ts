// useStudyTree — branching study-tree composable.
//
// "Five-dimensional time-travel chess." A reader walks word → verse → word →
// branch back two steps → different word → ..., and the path is preserved
// as a graph they can navigate and toggle between branches.
//
// Design choices (full rationale in
//   projects/1828-illuminated/.spec/proposals/study-tree-and-ux-polish.md):
//   - One node per visit, parent edges only (children computed)
//   - Multiple roots allowed (a tab may hold several unrelated chains)
//   - Idempotency: same (parentId, kind, identity) under active node = reuse
//   - Persistence: localStorage `study-tree-v1`
//   - Surface-level emit: each clickable surface calls tree.visit(...)
//
// Cross-domain nodes (word/verse/chapter/render) are first-class — the tree
// captures the boundary-crossing path that makes the tool a real study aid.

import { reactive, ref, computed, watch, type ComputedRef } from 'vue'

// ─── Types ────────────────────────────────────────────────────────────

export type NodeKind = 'word' | 'verse' | 'chapter' | 'render' | 'note'

export type NodePayload =
  | { kind: 'word'; word: string; stemMatched?: string | null }
  | { kind: 'verse'; abbrRef: string; humanRef: string; verse: number; text?: string }
  | { kind: 'chapter'; abbrRef: string; humanRef: string; range?: string; verseCount: number }
  | { kind: 'render'; sourceText: string; modernized: string; model: string; provider?: string }
  | { kind: 'note'; body: string }

export interface StudyNode {
  id: string
  kind: NodeKind
  parentId: string | null
  createdAt: number
  label: string
  payload: NodePayload
}

// ─── State (module-scoped — all components share one tree) ─────────────

const LS_KEY = 'study-tree-v1'

interface PersistedTree {
  version: 1
  nodes: StudyNode[]
  activeNodeId: string | null
}

const nodes = reactive<Map<string, StudyNode>>(new Map())
const activeNodeId = ref<string | null>(null)

/** Whether the side panel is currently shown. Module-scoped so toggling
 *  from one component is visible everywhere. */
export const panelOpen = ref(false)

/** When pinned, the panel anchors to the right column under the header
 *  (rather than overlaying as a slide-out drawer). Main content reflows
 *  to make room. Persisted to localStorage so the reader's preference
 *  survives reload. */
const PIN_KEY = 'study-tree-pinned-v1'
export const panelPinned = ref<boolean>(
  typeof localStorage !== 'undefined' && localStorage.getItem(PIN_KEY) === '1',
)
watch(panelPinned, (v) => {
  try { localStorage.setItem(PIN_KEY, v ? '1' : '0') } catch { /* storage off */ }
})

// ─── ID generator (small, no crypto.randomUUID dependency for SSR-safety) ─

let _idCounter = 0
function nextId(): string {
  _idCounter += 1
  // 8 random base36 chars + counter for guaranteed uniqueness per session.
  const r = Math.random().toString(36).slice(2, 10)
  return `n${Date.now().toString(36)}${r}${_idCounter.toString(36)}`
}

// ─── Identity (used for idempotency) ──────────────────────────────────

/** Canonical identity string for a payload — two nodes are "the same"
 *  when their identity matches under the same parent. */
function identityOf(p: NodePayload): string {
  switch (p.kind) {
    case 'word':
      return `word:${p.word.toLowerCase()}`
    case 'verse':
      return `verse:${p.abbrRef}:${p.verse}`
    case 'chapter':
      return `chapter:${p.abbrRef}${p.range ? `:${p.range}` : ''}`
    case 'render': {
      // Hash-light: model + first 80 chars of source text. Two renders of the
      // same source by the same model are the same node; re-rendering with a
      // different model creates a sibling.
      const head = p.sourceText.slice(0, 80).replace(/\s+/g, ' ').trim()
      return `render:${p.model}:${head}`
    }
    case 'note':
      return `note:${p.body.slice(0, 80)}`
  }
}

/** Default human-readable label for a payload. */
function defaultLabel(p: NodePayload): string {
  switch (p.kind) {
    case 'word':
      return p.word
    case 'verse':
      return `${p.humanRef}`
    case 'chapter':
      return p.range ? `${p.humanRef.replace(/:.*/, '')} v.${p.range}` : p.humanRef
    case 'render':
      return `↻ render (${p.model})`
    case 'note':
      return p.body.slice(0, 30) + (p.body.length > 30 ? '…' : '')
  }
}

// ─── Persistence ──────────────────────────────────────────────────────

function loadFromStorage(): void {
  try {
    const raw = localStorage.getItem(LS_KEY)
    if (!raw) return
    const parsed = JSON.parse(raw) as PersistedTree
    if (parsed.version !== 1 || !Array.isArray(parsed.nodes)) return
    nodes.clear()
    for (const n of parsed.nodes) {
      if (n && typeof n.id === 'string') nodes.set(n.id, n)
    }
    activeNodeId.value = parsed.activeNodeId ?? null
  } catch {
    // localStorage may be unavailable (incognito, etc.) — start empty.
  }
}

function saveToStorage(): void {
  try {
    const payload: PersistedTree = {
      version: 1,
      nodes: Array.from(nodes.values()),
      activeNodeId: activeNodeId.value,
    }
    localStorage.setItem(LS_KEY, JSON.stringify(payload))
  } catch {
    // Quota exceeded or storage disabled — silent. The in-memory tree is
    // still authoritative for this session.
  }
}

// Save on any change, but debounced so a fast click chain doesn't write
// localStorage twenty times in 200ms.
let _saveTimer: ReturnType<typeof setTimeout> | null = null
function scheduleSave(): void {
  if (_saveTimer) clearTimeout(_saveTimer)
  _saveTimer = setTimeout(() => {
    _saveTimer = null
    saveToStorage()
  }, 250)
}

// One-time boot.
let _loaded = false
function ensureLoaded(): void {
  if (_loaded) return
  _loaded = true
  loadFromStorage()
}

// Auto-save on any mutation.
watch([nodes, activeNodeId], scheduleSave, { deep: true })

// ─── Operations ───────────────────────────────────────────────────────

/** Visit a node. Returns the resulting node (new or reused).
 *
 *  Behavior:
 *   - If `activeNodeId` is set AND it has a child with matching identity:
 *     navigate to that existing child (no duplicate).
 *   - Otherwise create a new child under `activeNodeId` (or a new root if
 *     activeNodeId is null), set it active.
 */
export function visit(payload: NodePayload, label?: string): StudyNode {
  ensureLoaded()
  const parentId = activeNodeId.value
  const identity = identityOf(payload)

  // Idempotency: if the active node IS this identity, no-op. This catches
  // the tree-panel-navigate case (panel calls navigateTo(X) and then routes
  // to that surface, which fires its own visit watcher with X's payload —
  // without this check we'd create X as a child of itself every time).
  if (parentId) {
    const active = nodes.get(parentId)
    if (active && identityOf(active.payload) === identity) {
      return active
    }
  }

  // Idempotency: do we already have a child of activeNode with this identity?
  for (const n of nodes.values()) {
    if (n.parentId === parentId && identityOf(n.payload) === identity) {
      activeNodeId.value = n.id
      return n
    }
  }

  // New node.
  const node: StudyNode = {
    id: nextId(),
    kind: payload.kind,
    parentId,
    createdAt: Date.now(),
    label: label ?? defaultLabel(payload),
    payload,
  }
  nodes.set(node.id, node)
  activeNodeId.value = node.id
  return node
}

/** Set an existing node as active without creating anything. */
export function navigateTo(id: string): StudyNode | null {
  const n = nodes.get(id)
  if (!n) return null
  activeNodeId.value = id
  return n
}

/** Remove a node and ALL its descendants. */
export function removeSubtree(id: string): void {
  const toRemove = new Set<string>([id])
  let changed = true
  while (changed) {
    changed = false
    for (const n of nodes.values()) {
      if (n.parentId && toRemove.has(n.parentId) && !toRemove.has(n.id)) {
        toRemove.add(n.id)
        changed = true
      }
    }
  }
  for (const rid of toRemove) nodes.delete(rid)
  if (activeNodeId.value && toRemove.has(activeNodeId.value)) {
    activeNodeId.value = null
  }
}

/** Wipe everything. Called from "Start fresh" after confirmation. */
export function clearAll(): void {
  nodes.clear()
  activeNodeId.value = null
}

/** Direct children of a node id (or roots when id === null), sorted by createdAt. */
export function childrenOf(id: string | null): StudyNode[] {
  const out: StudyNode[] = []
  for (const n of nodes.values()) {
    if (n.parentId === id) out.push(n)
  }
  out.sort((a, b) => a.createdAt - b.createdAt)
  return out
}

/** Ancestor chain (root → ... → node), inclusive of node itself. */
export function ancestorsOf(id: string): StudyNode[] {
  const chain: StudyNode[] = []
  let cur = nodes.get(id)
  while (cur) {
    chain.unshift(cur)
    cur = cur.parentId ? nodes.get(cur.parentId) : undefined
  }
  return chain
}

// ─── Public reactive surface ──────────────────────────────────────────

export function useStudyTree() {
  ensureLoaded()

  const allNodes: ComputedRef<StudyNode[]> = computed(() => Array.from(nodes.values()))
  const roots: ComputedRef<StudyNode[]> = computed(() => childrenOf(null))
  const activeNode: ComputedRef<StudyNode | null> = computed(() =>
    activeNodeId.value ? nodes.get(activeNodeId.value) ?? null : null,
  )
  const activePath: ComputedRef<StudyNode[]> = computed(() =>
    activeNodeId.value ? ancestorsOf(activeNodeId.value) : [],
  )
  const nodeCount: ComputedRef<number> = computed(() => nodes.size)

  return {
    visit,
    navigateTo,
    removeSubtree,
    clearAll,
    childrenOf,
    ancestorsOf,
    allNodes,
    roots,
    activeNode,
    activeNodeId,
    activePath,
    nodeCount,
    panelOpen,
  }
}

// ─── Markdown export ──────────────────────────────────────────────────

/** Render the whole tree (or one root subtree) as markdown for journaling. */
export function exportTreeMarkdown(rootId?: string): string {
  const lines: string[] = ['# Study tree', '']
  const rootList = rootId ? [nodes.get(rootId)].filter(Boolean) as StudyNode[] : childrenOf(null)
  for (const root of rootList) {
    appendNodeMd(root, 0, lines)
    lines.push('')
  }
  return lines.join('\n')
}

function appendNodeMd(node: StudyNode, depth: number, out: string[]): void {
  const indent = '  '.repeat(depth)
  const meta = `[${node.kind}]`
  out.push(`${indent}- ${meta} **${node.label}**`)
  // Optional payload excerpt
  if (node.kind === 'verse' && node.payload.kind === 'verse' && node.payload.text) {
    out.push(`${indent}  > ${node.payload.text}`)
  } else if (node.kind === 'render' && node.payload.kind === 'render') {
    out.push(`${indent}  > ${node.payload.modernized.slice(0, 200)}${node.payload.modernized.length > 200 ? '…' : ''}`)
  } else if (node.kind === 'note' && node.payload.kind === 'note') {
    out.push(`${indent}  > ${node.payload.body}`)
  }
  for (const c of childrenOf(node.id)) {
    appendNodeMd(c, depth + 1, out)
  }
}
