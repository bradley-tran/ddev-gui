import type { DdevProject } from './types'

export function getProjectName(project: DdevProject): string {
  return String(project.name ?? project.project ?? project.projectname ?? '')
}

export function getProjectStatus(project: DdevProject): string {
  return String(project.status_desc ?? project.status ?? project.state ?? '')
}

export function getProjectType(project: DdevProject): string {
  return String(pickProjectValue(project, ['type', 'projecttype']) ?? '')
}

/**
 * Picks a value from a project object based on a list of potential keys.
 * Case-insensitive lookup is performed if an exact match is not found.
 */
export function pickProjectValue(project: DdevProject | null, keys: string[]): unknown {
  if (!project) return undefined

  let lowerKeysMap: Record<string, unknown> | null = null

  for (const key of keys) {
    // 1. Fast path: exact match
    if (project[key] !== undefined && project[key] !== null) {
      return project[key]
    }

    // 2. Slow path: case-insensitive match
    if (!lowerKeysMap) {
      lowerKeysMap = Object.create(null) as Record<string, unknown>
      for (const projectKey in project) {
        const val = project[projectKey]
        if (val !== undefined && val !== null) {
          const lowerProjectKey = projectKey.toLowerCase()
          if (lowerKeysMap[lowerProjectKey] === undefined) {
            lowerKeysMap[lowerProjectKey] = val
          }
        }
      }
    }

    const lowerKey = key.toLowerCase()
    if (lowerKeysMap[lowerKey] !== undefined) {
      return lowerKeysMap[lowerKey]
    }
  }

  return undefined
}

export function getPrimaryUrl(project: DdevProject): string {
  const https = project.httpsurl ?? ''
  const primary = project.primary_url ?? ''
  const http = project.httpurl ?? project.url ?? ''

  return (
    (typeof https === 'string' && https) ||
    (typeof primary === 'string' && /^https:/i.test(primary) ? primary : '') ||
    (typeof primary === 'string' && primary) ||
    (typeof http === 'string' && http) ||
    ''
  )
}

export function getMailpitUrl(project: DdevProject): string {
  return String(project.mailpit_https_url ?? project.mailpit_url ?? '')
}

export function isProjectRunning(project: DdevProject): boolean {
  return getProjectStatus(project).toLowerCase().includes('run')
}

export function isProjectStopped(project: DdevProject): boolean {
  const status = getProjectStatus(project).toLowerCase()
  return status.includes('stop') || status.includes('pause')
}

export function parseProjectsJSON(jsonStr: string): DdevProject[] {
  try {
    const data = JSON.parse(jsonStr) as { raw?: unknown; items?: unknown }
    if (Array.isArray(data)) return data
    if (Array.isArray(data?.raw)) return data.raw as DdevProject[]
    if (Array.isArray(data?.items)) return data.items as DdevProject[]
  } catch {
    return []
  }

  return []
}

export function uid(): string {
  return `id_${Date.now()}_${Math.random().toString(36).slice(2)}`
}

export function coerceToBool(value: unknown, defaultValue = true): boolean {
  if (value === true || value === 'true') return true
  if (value === false || value === 'false') return false
  return defaultValue
}

export function getIsLinux(): boolean {
  return /linux/i.test(navigator.userAgent || navigator.platform || '')
}

export function openUrl(url: string, openInBrowser = true): void {
  if (!url) return

  if (getIsLinux()) {
    window.runtime?.EventsEmit('open:url', { url })
    return
  }

  if (openInBrowser) {
    window.runtime?.EventsEmit('open:url', { url })
    return
  }

  try {
    window.open(url, '_blank', 'width=1024,height=700')
  } catch {
    return
  }
}