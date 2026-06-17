import { describe, bench } from 'vitest'

const generateLargeRawData = (size = 10000) => {
  const raw: Record<string, any> = {}
  for (let i = 0; i < size; i++) {
    raw[`key${i}`] = { some: 'data', index: i }
  }
  // Put the array at the end
  raw['target'] = [{ snapshot: 'data' }]
  return raw
}

describe('Snapshot Parsing Optimization', () => {
  const raw = generateLargeRawData()

  bench('Original - Object.values().find()', () => {
    const list = Object.values(raw).find(Array.isArray)
    const result = Array.isArray(list) ? list : []
  })

  bench('Optimized - for...in loop', () => {
    let result = []
    for (const key in raw) {
      const val = raw[key]
      if (Array.isArray(val)) {
        result = val
        break
      }
    }
  })
})
