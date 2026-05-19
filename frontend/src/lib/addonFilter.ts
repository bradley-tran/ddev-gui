export interface AddonItem {
  repo?: string
  description?: string
  Description?: string
  Repository?: string
  repository?: string
  full_name?: string
  FullName?: string
  [key: string]: unknown
}

export interface ParsedSearch {
  query: string
  tokens: string[]
  pattern: RegExp | null
}

export function parseSearch(search: string): ParsedSearch {
  const query = search.trim().toLowerCase()
  if (!query) return { query, tokens: [], pattern: null }

  const tokens = query.split(/\s+/).filter(Boolean)
  let pattern: RegExp | null = null

  if (query.length >= 3) {
    const escaped = tokens.map((token) => token.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'))
    pattern = new RegExp(`\\b${escaped.join('.*\\b')}`, 'i')
  }

  return { query, tokens, pattern }
}

export function matchesAddonSearch(item: AddonItem, search: string | ParsedSearch): boolean {
  const parsed = typeof search === 'string' ? parseSearch(search) : search
  const { query, tokens, pattern } = parsed

  if (!query) return true

  const repo = String(item.repo ?? item.Repository ?? item.repository ?? item.full_name ?? item.FullName ?? '').toLowerCase()
  const desc = String(item.description ?? item.Description ?? '').toLowerCase()

  const repoSegments = repo.split(/[/\-_\s]+/)

  const repoMatches = tokens.every(
    (token) => repoSegments.some((segment) => segment.startsWith(token)) || repo.includes(token),
  )
  if (repoMatches) return true

  if (pattern) {
    return pattern.test(desc)
  }

  return false
}

export function filterAddons(items: AddonItem[], search: string): AddonItem[] {
  const parsed = parseSearch(search)
  if (!parsed.query) return items

  return items.filter((item) => matchesAddonSearch(item, parsed))
}
