import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest'
import { Cache, dirCache, fileCache, cacheKey } from '../cache'

describe('Cache utility', () => {
  describe('Cache class', () => {
    beforeEach(() => {
      vi.useFakeTimers()
    })

    afterEach(() => {
      vi.useRealTimers()
    })

    it('should store and retrieve data', () => {
      const cache = new Cache<string>()
      cache.set('key1', 'value1')
      expect(cache.get('key1')).toBe('value1')
    })

    it('should return undefined for non-existent keys', () => {
      const cache = new Cache<string>()
      expect(cache.get('non-existent')).toBeUndefined()
    })

    it('should respect TTL (Time To Live)', () => {
      const ttl = 1000
      const cache = new Cache<string>(ttl)

      cache.set('key1', 'value1')
      expect(cache.get('key1')).toBe('value1')

      // Advance time by ttl
      vi.advanceTimersByTime(ttl)
      expect(cache.get('key1')).toBe('value1')

      // Advance time by 1ms more
      vi.advanceTimersByTime(1)

      expect(cache.get('key1')).toBeUndefined()
      expect(cache.size).toBe(0)
    })

    it('should return true for has() if key exists and is not expired', () => {
      const ttl = 1000
      const cache = new Cache<string>(ttl)

      cache.set('key1', 'value1')
      expect(cache.has('key1')).toBe(true)

      vi.advanceTimersByTime(ttl + 1)
      expect(cache.has('key1')).toBe(false)
    })

    it('should delete keys', () => {
      const cache = new Cache<string>()
      cache.set('key1', 'value1')
      expect(cache.has('key1')).toBe(true)

      cache.delete('key1')
      expect(cache.has('key1')).toBe(false)
    })

    it('should clear all keys', () => {
      const cache = new Cache<string>()
      cache.set('key1', 'value1')
      cache.set('key2', 'value2')
      expect(cache.size).toBe(2)

      cache.clear()
      expect(cache.size).toBe(0)
    })

    it('should report correct size', () => {
      const cache = new Cache<string>()
      expect(cache.size).toBe(0)

      cache.set('key1', 'value1')
      expect(cache.size).toBe(1)

      cache.set('key2', 'value2')
      expect(cache.size).toBe(2)

      cache.delete('key1')
      expect(cache.size).toBe(1)
    })
  })

  describe('cacheKey', () => {
    it('should format keys correctly', () => {
      expect(cacheKey('my-project', 'path/to/file')).toBe('my-project:path/to/file')
    })
  })

  describe('exported instances', () => {
    beforeEach(() => {
      vi.useFakeTimers()
    })

    afterEach(() => {
      vi.useRealTimers()
      dirCache.clear()
      fileCache.clear()
    })

    it('dirCache should have 2 minutes TTL', () => {
      const data = [{ name: 'test', isDir: true, size: '0', modified: '' }]
      dirCache.set('key', data)
      expect(dirCache.get('key')).toBe(data)

      vi.advanceTimersByTime(2 * 60 * 1000)
      expect(dirCache.get('key')).toBe(data)

      vi.advanceTimersByTime(1)
      expect(dirCache.get('key')).toBeUndefined()
    })

    it('fileCache should have 5 minutes TTL', () => {
      const data = 'file content'
      fileCache.set('key', data)
      expect(fileCache.get('key')).toBe(data)

      vi.advanceTimersByTime(5 * 60 * 1000)
      expect(fileCache.get('key')).toBe(data)

      vi.advanceTimersByTime(1)
      expect(fileCache.get('key')).toBeUndefined()
    })
  })
})
