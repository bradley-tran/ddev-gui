interface CacheEntry<T> {
  data: T
  timestamp: number
}

export class Cache<T> {
  private readonly store = new Map<string, CacheEntry<T>>()
  private readonly ttlMs: number

  constructor(ttlMs = 5 * 60 * 1000) {
    this.ttlMs = ttlMs
  }

  get(key: string): T | undefined {
    const entry = this.store.get(key)
    if (!entry) return undefined

    if (Date.now() - entry.timestamp > this.ttlMs) {
      this.store.delete(key)
      return undefined
    }

    return entry.data
  }

  set(key: string, data: T): void {
    this.store.set(key, { data, timestamp: Date.now() })
  }

  has(key: string): boolean {
    return this.get(key) !== undefined
  }

  delete(key: string): void {
    this.store.delete(key)
  }

  clear(): void {
    this.store.clear()
  }

  get size(): number {
    return this.store.size
  }
}

interface FileEntry {
  name: string
  isDir: boolean
  size: string
  modified: string
}

export const dirCache = new Cache<FileEntry[]>(2 * 60 * 1000)
export const fileCache = new Cache<string>(5 * 60 * 1000)

export function cacheKey(projectName: string, relPath: string): string {
  return `${projectName}:${relPath}`
}