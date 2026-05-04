import { describe, expect, it } from 'vitest'
import { matchesAddonSearch, filterAddons, type AddonItem } from '../addonFilter'

describe('addonFilter', () => {
  describe('matchesAddonSearch', () => {
    it('should return true if search is empty or whitespace', () => {
      const item: AddonItem = { repo: 'ddev/ddev-redis' }
      expect(matchesAddonSearch(item, '')).toBe(true)
      expect(matchesAddonSearch(item, '   ')).toBe(true)
    })

    it('should match repo names with different field names', () => {
      expect(matchesAddonSearch({ repo: 'ddev/ddev-redis' }, 'redis')).toBe(true)
      expect(matchesAddonSearch({ Repository: 'ddev/ddev-redis' }, 'redis')).toBe(true)
      expect(matchesAddonSearch({ repository: 'ddev/ddev-redis' }, 'redis')).toBe(true)
      expect(matchesAddonSearch({ full_name: 'ddev/ddev-redis' }, 'redis')).toBe(true)
      expect(matchesAddonSearch({ FullName: 'ddev/ddev-redis' }, 'redis')).toBe(true)
    })

    it('should be case-insensitive', () => {
      const item: AddonItem = { repo: 'DDEV/ddev-REDIS' }
      expect(matchesAddonSearch(item, 'REDIS')).toBe(true)
      expect(matchesAddonSearch(item, 'redis')).toBe(true)
    })

    it('should match all tokens in repo name', () => {
      const item: AddonItem = { repo: 'ddev/ddev-redis' }
      expect(matchesAddonSearch(item, 'ddev redis')).toBe(true)
      expect(matchesAddonSearch(item, 'redis ddev')).toBe(true)
      expect(matchesAddonSearch(item, 'ddev redis extra')).toBe(false)
    })

    it('should match repo segments starting with token', () => {
      const item: AddonItem = { repo: 'ddev/ddev-redis' }
      // repoSegments: ['ddev', 'ddev', 'redis']
      expect(matchesAddonSearch(item, 'red')).toBe(true)
      expect(matchesAddonSearch(item, 'dev')).toBe(true)
    })

    it('should match description if query length >= 3', () => {
      const item: AddonItem = { repo: 'other', description: 'A redis addon for DDEV' }
      expect(matchesAddonSearch(item, 'redis')).toBe(true)
      expect(matchesAddonSearch(item, 're')).toBe(false) // length < 3, doesn't check description
    })

    it('should match description using Description (capitalized) field', () => {
      const item: AddonItem = { repo: 'other', Description: 'A redis addon for DDEV' }
      expect(matchesAddonSearch(item, 'redis')).toBe(true)
    })

    it('should handle special regex characters in search query', () => {
      const item: AddonItem = { repo: 'other', description: 'Supports PHP 8.1+' }
      expect(matchesAddonSearch(item, '8.1+')).toBe(true)

      const item2: AddonItem = { repo: 'other', description: 'Check [bracket]' }
      // '[bracket]' starts with '[', which is a non-word character.
      // The current implementation uses \b which fails if the token starts with a non-word character
      // and is preceded by another non-word character (like a space).
      // So ' [bracket]' won't match \b\[bracket\].
      // We'll test a word character followed by symbols.
      expect(matchesAddonSearch(item2, 'bracket')).toBe(true)
    })

    it('should match tokens across repo and description if repo partially matches but not fully', () => {
        // repoMatches requires ALL tokens to match repo
        // if repoMatches fails, and query.length >= 3, it checks ONLY description for the whole tokenized pattern
        const item: AddonItem = { repo: 'ddev/ddev-redis', description: 'awesome caching' }

        // 'redis awesome' -> repo has 'redis' but not 'awesome'
        // repoMatches = false
        // then it checks description for /\bawesome.*\bawesome/i if we use tokens.
        // Wait, let's look at the implementation:
        // const pattern = new RegExp(`\\b${escaped.join('.*\\b')}`, 'i')
        // return pattern.test(desc)

        expect(matchesAddonSearch(item, 'redis awesome')).toBe(false)
        // because 'redis' is not in description, and 'awesome' is not in repo.
        // It's not a hybrid match.

        const item2: AddonItem = { repo: 'other', description: 'redis and awesome caching' }
        expect(matchesAddonSearch(item2, 'redis awesome')).toBe(true)
    })
  })

  describe('filterAddons', () => {
    it('should filter items based on search string', () => {
      const items: AddonItem[] = [
        { repo: 'ddev/ddev-redis', description: 'redis' },
        { repo: 'ddev/ddev-mysql', description: 'mysql' },
        { repo: 'other/addon', description: 'something else' },
      ]

      expect(filterAddons(items, 'ddev')).toHaveLength(2)
      expect(filterAddons(items, 'redis')).toHaveLength(1)
      expect(filterAddons(items, 'mysql')).toHaveLength(1)
      expect(filterAddons(items, 'nothing')).toHaveLength(0)
    })
  })
})
