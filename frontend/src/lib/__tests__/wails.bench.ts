import { describe, bench, vi } from 'vitest'

describe('Runtime bridge execution - iteration comparison', () => {
  const listeners = new Set<Function>()
  for (let i = 0; i < 100; i++) {
    listeners.add(() => {})
  }

  bench('with Array.from()', () => {
    for (const listener of Array.from(listeners)) {
      listener('arg1', 'arg2')
    }
  })

  bench('with Set directly', () => {
    for (const listener of listeners) {
      listener('arg1', 'arg2')
    }
  })
})
