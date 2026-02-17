// GitHubContentService — fetches file trees and markdown content from GitHub.
// For public repos, all requests go directly to GitHub from the browser (no server involved).
// Uses raw.githubusercontent.com for content (no rate limit) and the Trees API for listings.

export interface TreeEntry {
  path: string
  type: 'blob' | 'tree'
  size?: number
}

interface GitHubTreeResponse {
  sha: string
  tree: Array<{
    path: string
    mode: string
    type: 'blob' | 'tree'
    sha: string
    size?: number
    url: string
  }>
  truncated: boolean
}

// Minimatch-style glob matching (simplified but handles **, *, and basic patterns)
function matchGlob(pattern: string, path: string): boolean {
  // Convert glob to regex
  let regex = pattern
    // Escape regex special chars except * and ?
    .replace(/[.+^${}()|[\]\\]/g, '\\$&')
    // ** matches across directories
    .replace(/\*\*/g, '{{GLOBSTAR}}')
    // * matches within a single directory
    .replace(/\*/g, '[^/]*')
    // ? matches single char
    .replace(/\?/g, '[^/]')
    // Restore globstar
    .replace(/\{\{GLOBSTAR\}\}/g, '.*')

  return new RegExp(`^${regex}$`).test(path)
}

function matchesAny(patterns: string[], path: string): boolean {
  return patterns.some(p => matchGlob(p, path))
}

interface CachedTree {
  entries: TreeEntry[]
  fetchedAt: number
  etag?: string
}

const TREE_CACHE_TTL = 5 * 60 * 1000 // 5 minutes

export class GitHubContentService {
  private treeCache = new Map<string, CachedTree>()

  private treeCacheKey(repo: string, branch: string): string {
    return `${repo}@${branch}`
  }

  /**
   * Fetch the full recursive file tree for a repo, filtered by include/exclude patterns.
   * Results are cached in memory for 5 minutes.
   */
  async getTree(
    repo: string,
    branch: string,
    include: string[],
    exclude: string[]
  ): Promise<TreeEntry[]> {
    const key = this.treeCacheKey(repo, branch)
    const cached = this.treeCache.get(key)
    if (cached && Date.now() - cached.fetchedAt < TREE_CACHE_TTL) {
      return this.filterTree(cached.entries, include, exclude)
    }

    const url = `https://api.github.com/repos/${repo}/git/trees/${branch}?recursive=1`
    const headers: Record<string, string> = {
      Accept: 'application/vnd.github.v3+json',
    }
    if (cached?.etag) {
      headers['If-None-Match'] = cached.etag
    }

    const res = await fetch(url, { headers })

    if (res.status === 304 && cached) {
      // Not modified — refresh TTL
      cached.fetchedAt = Date.now()
      return this.filterTree(cached.entries, include, exclude)
    }

    if (!res.ok) {
      throw new Error(`GitHub Trees API error: ${res.status} ${res.statusText}`)
    }

    const data: GitHubTreeResponse = await res.json()
    const entries: TreeEntry[] = data.tree
      .filter(e => e.type === 'blob' && e.path.endsWith('.md'))
      .map(e => ({ path: e.path, type: e.type, size: e.size }))

    const etag = res.headers.get('etag') || undefined
    this.treeCache.set(key, { entries, fetchedAt: Date.now(), etag })

    if (data.truncated) {
      console.warn(`GitHub tree for ${repo} was truncated — some files may be missing`)
    }

    return this.filterTree(entries, include, exclude)
  }

  /**
   * Fetch raw markdown content for a file. Uses raw.githubusercontent.com (no rate limit for public repos).
   */
  async getContent(repo: string, branch: string, path: string): Promise<string> {
    const url = `https://raw.githubusercontent.com/${repo}/${branch}/${path}`
    const res = await fetch(url)
    if (!res.ok) {
      throw new Error(`Failed to fetch ${path}: ${res.status} ${res.statusText}`)
    }
    return res.text()
  }

  /**
   * Apply include/exclude filters to a list of tree entries.
   */
  private filterTree(entries: TreeEntry[], include: string[], exclude: string[]): TreeEntry[] {
    let filtered = entries

    // If include patterns are specified, only keep matching files
    if (include.length > 0) {
      filtered = filtered.filter(e => matchesAny(include, e.path))
    }

    // Remove files matching exclude patterns
    if (exclude.length > 0) {
      filtered = filtered.filter(e => !matchesAny(exclude, e.path))
    }

    return filtered
  }

  /**
   * Invalidate the cached tree for a repo.
   */
  invalidateTree(repo: string, branch: string): void {
    this.treeCache.delete(this.treeCacheKey(repo, branch))
  }

  /**
   * Build a nested directory structure from flat file paths for sidebar display.
   */
  buildFileTree(entries: TreeEntry[]): FileTreeNode[] {
    const root: FileTreeNode[] = []

    for (const entry of entries) {
      const parts = entry.path.split('/')
      let current = root

      for (let i = 0; i < parts.length; i++) {
        const name = parts[i]!
        const isFile = i === parts.length - 1

        let node = current.find(n => n.name === name)
        if (!node) {
          node = {
            name,
            path: parts.slice(0, i + 1).join('/'),
            type: isFile ? 'file' : 'directory',
            children: isFile ? undefined : [],
          }
          current.push(node)
        }
        if (!isFile && node?.children) {
          current = node.children
        }
      }
    }

    // Sort: directories first, then alphabetically
    const sortNodes = (nodes: FileTreeNode[]) => {
      nodes.sort((a, b) => {
        if (a.type !== b.type) return a.type === 'directory' ? -1 : 1
        return a.name.localeCompare(b.name)
      })
      for (const node of nodes) {
        if (node.children) sortNodes(node.children)
      }
    }
    sortNodes(root)

    return root
  }
}

export interface FileTreeNode {
  name: string
  path: string
  type: 'file' | 'directory'
  children?: FileTreeNode[]
}

// Singleton instance
export const github = new GitHubContentService()
