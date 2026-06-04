import { bench, describe } from 'vitest'
import { createI18nState } from '../i18n'

describe('i18n t() interpolation benchmark', () => {
  const state = createI18nState('en')

  // Make sure to add a test key
  state.messages.value = {
    'test.bench.1': 'This is a test with {token1} and {token2} and {token3}.',
    'test.bench.no_vars': 'This is a string with no vars.',
  }

  bench('t() with no vars', () => {
    state.t('test.bench.no_vars')
  })

  bench('t() with 3 vars', () => {
    state.t('test.bench.1', {
      token1: 'replacement1',
      token2: 'replacement2',
      token3: 'replacement3',
    })
  })
})
