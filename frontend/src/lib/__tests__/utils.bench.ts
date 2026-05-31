import { bench, describe } from 'vitest'
import { pickProjectValue } from '../utils'
import type { DdevProject } from '../types'

describe('pickProjectValue', () => {
  const project = {
    Name: 'test',
    ProjecTType: 'php',
    STATUS: 'running',
    a: 1,
    b: 2,
    c: 3,
    d: 4,
    e: 5,
    f: 6,
    g: 7,
    h: 8,
    i: 9,
    j: 10,
    k: 11,
    l: 12,
    m: 13,
    n: 14,
    o: 15,
    p: 16,
    q: 17,
    r: 18,
    s: 19,
    t: 20,
  } as unknown as DdevProject

  bench('exact match', () => {
    pickProjectValue(project, ['Name'])
  })

  bench('case-insensitive match', () => {
    pickProjectValue(project, ['name'])
  })

  bench('case-insensitive multiple keys', () => {
    pickProjectValue(project, ['type', 'projecttype'])
  })

  bench('not found', () => {
    pickProjectValue(project, ['notfound1', 'notfound2'])
  })
})
