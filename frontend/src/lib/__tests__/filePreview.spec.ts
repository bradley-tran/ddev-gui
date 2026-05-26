import { describe, expect, it } from 'vitest'
import {
  isMarkdownFile,
  isImageFile,
  isBinaryFile,
  isCodeFile,
  detectLanguage,
  mimeFromExt,
} from '../filePreview'

describe('filePreview', () => {
  describe('isMarkdownFile', () => {
    it('should return true for markdown extensions', () => {
      expect(isMarkdownFile('test.md')).toBe(true)
      expect(isMarkdownFile('TEST.MARKDOWN')).toBe(true)
      expect(isMarkdownFile('index.mdx')).toBe(true)
    })

    it('should return false for other extensions', () => {
      expect(isMarkdownFile('test.txt')).toBe(false)
      expect(isMarkdownFile('test.md.bak')).toBe(false)
      expect(isMarkdownFile('md')).toBe(false)
    })
  })

  describe('isImageFile', () => {
    it.each([
      ['test.png', true],
      ['TEST.JPG', true],
      ['image.jpeg', true],
      ['animation.gif', true],
      ['logo.svg', true],
      ['photo.webp', true],
      ['favicon.ico', true],
      ['image.bmp', true],
      ['picture.avif', true],
      ['test.txt', false],
      ['image.png.zip', false],
      ['no-ext', false],
    ])('should evaluate %s to %s', (filename, expected) => {
      expect(isImageFile(filename)).toBe(expected)
    })
  })

  describe('isBinaryFile', () => {
    it('should return true for binary extensions', () => {
      expect(isBinaryFile('archive.zip')).toBe(true)
      expect(isBinaryFile('SETUP.EXE')).toBe(true)
      expect(isBinaryFile('doc.pdf')).toBe(true)
      expect(isBinaryFile('video.mp4')).toBe(true)
    })

    it('should return false for non-binary extensions', () => {
      expect(isBinaryFile('test.txt')).toBe(false)
      expect(isBinaryFile('script.sh')).toBe(false)
    })
  })

  describe('isCodeFile', () => {
    it('should return true for code extensions', () => {
      expect(isCodeFile('script.js')).toBe(true)
      expect(isCodeFile('MAIN.GO')).toBe(true)
      expect(isCodeFile('styles.css')).toBe(true)
    })

    it('should return true for specific filenames', () => {
      expect(isCodeFile('Dockerfile')).toBe(true)
      expect(isCodeFile('Makefile')).toBe(true)
      expect(isCodeFile('.gitignore')).toBe(true)
      expect(isCodeFile('.env')).toBe(true)
    })

    it('should return false for non-code files', () => {
      expect(isCodeFile('image.png')).toBe(false)
      expect(isCodeFile('notes.txt')).toBe(false)
      expect(isCodeFile('README.md')).toBe(false)
    })
  })

  describe('detectLanguage', () => {
    it('should detect language based on extension', () => {
      expect(detectLanguage('test.ts')).toBe('typescript')
      expect(detectLanguage('TEST.PY')).toBe('python')
      expect(detectLanguage('index.html')).toBe('xml')
    })

    it('should detect language based on filename', () => {
      expect(detectLanguage('Dockerfile')).toBe('dockerfile')
      expect(detectLanguage('Vagrantfile')).toBe('ruby')
      expect(detectLanguage('.gitignore')).toBe('bash')
    })

    it('should return plaintext for unknown extensions', () => {
      expect(detectLanguage('unknown.xyz')).toBe('plaintext')
      expect(detectLanguage('no-extension')).toBe('plaintext')
    })
  })

  describe('mimeFromExt', () => {
    it('should return correct mime type for images', () => {
      expect(mimeFromExt('test.png')).toBe('image/png')
      expect(mimeFromExt('image.JPG')).toBe('image/jpeg')
      expect(mimeFromExt('logo.svg')).toBe('image/svg+xml')
    })

    it('should return default image/png for unknown or non-image extensions', () => {
      expect(mimeFromExt('test.txt')).toBe('image/png')
      expect(mimeFromExt('no-extension')).toBe('image/png')
    })
  })
})
