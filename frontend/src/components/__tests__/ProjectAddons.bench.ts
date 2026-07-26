import { describe, bench } from 'vitest'

const addons = Array.from({ length: 10000 }).map((_, i) => ({
  Name: `Addon ${i}`,
  Version: `1.0.${i}`,
  Repository: `user/addon-${i}`,
  InstalledDate: `2024-01-01`
}))

// Add some empty ones to trigger filter
for (let i = 0; i < 1000; i++) {
  addons.push({ Name: '', Version: '', Repository: '', InstalledDate: '' })
}

describe('normalizedAddons', () => {
  bench('map filter', () => {
    addons
      .map((item) => ({
        name: String(item.Name ?? (item as any).name ?? '').trim(),
        version: String(item.Version ?? (item as any).version ?? '').trim(),
        repo: String(
          item.Repository ?? (item as any).repository ?? (item as any).full_name ?? (item as any).FullName ?? (item as any).repo ?? (item as any).source ?? '',
        ).trim(),
        installed: String(
          item.InstalledDate ?? (item as any).installed_date ?? (item as any).installedDate ?? (item as any).installed ?? (item as any).date ?? '',
        ).trim(),
      }))
      .filter((item) => item.name)
  })

  bench('reduce', () => {
    addons.reduce((acc: any[], item) => {
      const name = String(item.Name ?? (item as any).name ?? '').trim()
      if (name) {
        acc.push({
          name,
          version: String(item.Version ?? (item as any).version ?? '').trim(),
          repo: String(
            item.Repository ?? (item as any).repository ?? (item as any).full_name ?? (item as any).FullName ?? (item as any).repo ?? (item as any).source ?? '',
          ).trim(),
          installed: String(
            item.InstalledDate ?? (item as any).installed_date ?? (item as any).installedDate ?? (item as any).installed ?? (item as any).date ?? '',
          ).trim(),
        })
      }
      return acc
    }, [])
  })

  bench('for loop', () => {
    const result = []
    for (let i = 0, len = addons.length; i < len; i++) {
      const item = addons[i]!
      const name = String(item.Name ?? (item as any).name ?? '').trim()
      if (name) {
        result.push({
          name,
          version: String(item.Version ?? (item as any).version ?? '').trim(),
          repo: String(
            item.Repository ?? (item as any).repository ?? (item as any).full_name ?? (item as any).FullName ?? (item as any).repo ?? (item as any).source ?? '',
          ).trim(),
          installed: String(
            item.InstalledDate ?? (item as any).installed_date ?? (item as any).installedDate ?? (item as any).installed ?? (item as any).date ?? '',
          ).trim(),
        })
      }
    }
  })
})
