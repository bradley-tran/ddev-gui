import { describe, expect, it } from 'vitest'
import { ansiToHtml, escapeHtml } from '../ansi'

describe('ansi', () => {
  describe('escapeHtml', () => {
    it('should escape HTML characters', () => {
      expect(escapeHtml('<div> & "test"</div>')).toBe('&lt;div&gt; &amp; "test"&lt;/div&gt;')
    })

    it('should return empty string for empty input', () => {
      expect(escapeHtml('')).toBe('')
    })

    it('should handle non-string inputs by converting to string', () => {
      expect(escapeHtml(123 as any)).toBe('123')
      expect(escapeHtml(true as any)).toBe('true')
    })

    it('should not escape other characters', () => {
      expect(escapeHtml('Hello World! 123')).toBe('Hello World! 123')
    })
  })

  describe('ansiToHtml', () => {
    it('should return plain text as is', () => {
      expect(ansiToHtml('Hello World')).toBe('Hello World')
    })

    it('should escape HTML characters in plain text', () => {
      expect(ansiToHtml('<b>Bold</b>')).toBe('&lt;b&gt;Bold&lt;/b&gt;')
    })

    it('should convert single ANSI color code to span', () => {
      // 31 is red: #e06c75
      expect(ansiToHtml('\x1b[31mRed Text\x1b[0m')).toBe('<span style="color:#e06c75">Red Text</span>')
    })

    it('should handle multiple color changes', () => {
      // 31 red, 32 green
      const input = '\x1b[31mRed\x1b[32mGreen\x1b[0m'
      const expected = '<span style="color:#e06c75">Red</span><span style="color:#98c379">Green</span>'
      expect(ansiToHtml(input)).toBe(expected)
    })

    it('should handle complex ANSI sequences with multiple codes', () => {
      const input = '\x1b[31;32mText\x1b[0m'
      // 31: #e06c75, 32: #98c379
      const expected = '<span style="color:#e06c75"><span style="color:#98c379">Text</span></span>'
      expect(ansiToHtml(input)).toBe(expected)
    })

    it('should ignore unknown ANSI codes', () => {
      // 99 is unknown
      expect(ansiToHtml('\x1b[99mUnknown\x1b[0m')).toBe('Unknown')
    })

    it('should handle malformed ANSI sequences gracefully', () => {
      expect(ansiToHtml('\x1b[31Text')).toBe('\x1b[31Text')
      expect(ansiToHtml('Text\x1b[')).toBe('Text\x1b[')
    })

    it('should close open spans at the end of the string', () => {
      expect(ansiToHtml('\x1b[31mRed')).toBe('<span style="color:#e06c75">Red</span>')
    })

    it('should handle all defined colors', () => {
      const COLORS: Record<string, string> = {
        '30': '#4e4e4e',
        '31': '#e06c75',
        '32': '#98c379',
        '33': '#e5c07b',
        '34': '#61afef',
        '35': '#c678dd',
        '36': '#56b6c2',
        '37': '#dcdfe4',
        '90': '#7f8490',
        '91': '#e06c75',
        '92': '#98c379',
        '93': '#e5c07b',
        '94': '#61afef',
        '95': '#c678dd',
        '96': '#56b6c2',
        '97': '#ffffff',
      }

      for (const [code, color] of Object.entries(COLORS)) {
        expect(ansiToHtml(`\x1b[${code}mColor\x1b[0m`)).toBe(`<span style="color:${color}">Color</span>`)
      }
    })

    it('should escape HTML characters even when inside a span', () => {
      expect(ansiToHtml('\x1b[31m<script>\x1b[0m')).toBe('<span style="color:#e06c75">&lt;script&gt;</span>')
    })
  })
})
