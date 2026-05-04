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

export function matchesAddonSearch(item: AddonItem, search: string): boolean {
  const repo = String(item.repo ?? item.Repository ?? item.repository ?? item.full_name ?? item.FullName ?? '').toLowerCase()
  const desc = String(item.description ?? item.Description ?? '').toLowerCase()

  const query = search.trim().toLowerCase()
  if (!query) return true

  const tokens = query.split(/\s+/).filter(Boolean)
  const repoSegments = repo.split(/[\/\-_\s]+/)

  const repoMatches = tokens.every(
    (token) => repoSegments.some((segment) => segment.startsWith(token)) || repo.includes(token),
  )
  if (repoMatches) return true

  if (query.length >= 3) {
    const escaped = tokens.map((token) => token.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'))
    const pattern = new RegExp(`\\b${escaped.join('.*\\b')}`, 'i')
    return pattern.test(desc)
  }

  return false
}

export function filterAddons(items: AddonItem[], search: string): AddonItem[] {
  return items.filter((item) => matchesAddonSearch(item, search))
}