import { bench, describe } from 'vitest'
import { ansiToHtml } from '../ansi'

describe('ansiToHtml', () => {
  const longStr = '\x1b[32mSuccess!\x1b[m The operation completed successfully. Here is some more output: \x1b[31mError:\x1b[m Just kidding, no error. <>&'.repeat(10000)

  bench('ansiToHtml', () => {
    ansiToHtml(longStr)
  })
})
