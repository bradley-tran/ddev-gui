import { bench, describe } from 'vitest'
import { parseSearch } from '../addonFilter'

describe('parseSearch', () => {
  const searchStr = Array.from({ length: 100 }).map((_, i) => `token*${i}?`).join(' ')

  bench('parse many tokens', () => {
    parseSearch(searchStr)
  })
})
