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
    })

    it('should return false for files without extension but named as one', () => {
      expect(isMarkdownFile('md')).toBe(false)
      expect(isMarkdownFile('markdown')).toBe(false)
    })

    it('should handle empty string', () => {
      expect(isMarkdownFile('')).toBe(false)
    })
  })

  describe('isImageFile', () => {
    it('should return true for image extensions', () => {
      expect(isImageFile('test.png')).toBe(true)
      expect(isImageFile('TEST.JPG')).toBe(true)
      expect(isImageFile('logo.svg')).toBe(true)
      expect(isImageFile('photo.webp')).toBe(true)
    })

    it('should return false for non-image extensions', () => {
      expect(isImageFile('test.txt')).toBe(false)
      expect(isImageFile('image.png.zip')).toBe(false)
      expect(isImageFile('no-ext')).toBe(false)
    })

    it('should handle empty string', () => {
      expect(isImageFile('')).toBe(false)
    })

    it('should return false for files without extension but named as one', () => {
      expect(isImageFile('png')).toBe(false)
      expect(isImageFile('jpg')).toBe(false)
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

    it('should handle empty string', () => {
      expect(isBinaryFile('')).toBe(false)
    })

    it('should return false for files without extension but named as one', () => {
      expect(isBinaryFile('zip')).toBe(false)
      expect(isBinaryFile('exe')).toBe(false)
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

    it('should handle empty string', () => {
      expect(isCodeFile('')).toBe(false)
    })

    it('should return false for files without extension but named as one', () => {
      expect(isCodeFile('js')).toBe(false)
      expect(isCodeFile('ts')).toBe(false)
    })
  })

  describe('detectLanguage', () => {
    it.each([
      ['test.js', 'javascript'],
      ['test.ts', 'typescript'],
      ['TEST.PY', 'python'],
      ['main.go', 'go'],
      ['app.rs', 'rust'],
      ['script.sh', 'bash'],
      ['query.sql', 'sql'],
      ['index.html', 'xml'],
      ['styles.css', 'css'],
      ['data.json', 'json'],
      ['config.yaml', 'yaml'],
      ['component.vue', 'xml'],
      ['Dockerfile', 'dockerfile'],
      ['Vagrantfile', 'ruby'],
      ['.gitignore', 'bash'],
      ['.env', 'bash'],
      ['.env.example', 'bash'],
      ['unknown.xyz', 'plaintext'],
      ['no-extension', 'plaintext'],
      ['', 'plaintext'],
    ])('should detect language correctly for %s -> %s', (filename, expected) => {
      expect(detectLanguage(filename)).toBe(expected)
    })
  })

  describe('mimeFromExt', () => {
    it.each([
      ['test.png', 'image/png'],
      ['test.jpg', 'image/jpeg'],
      ['test.jpeg', 'image/jpeg'],
      ['test.gif', 'image/gif'],
      ['test.svg', 'image/svg+xml'],
      ['test.webp', 'image/webp'],
      ['test.ico', 'image/x-icon'],
      ['test.bmp', 'image/bmp'],
      ['test.avif', 'image/avif'],
      ['image.JPG', 'image/jpeg'],
    ])('should return correct mime type for images (%s -> %s)', (filename, expected) => {
      expect(mimeFromExt(filename)).toBe(expected)
    })

    it.each([
      ['png', 'image/png'],
      ['jpg', 'image/jpeg'],
      ['jpeg', 'image/jpeg'],
      ['gif', 'image/gif'],
      ['svg', 'image/svg+xml'],
      ['webp', 'image/webp'],
      ['ico', 'image/x-icon'],
      ['bmp', 'image/bmp'],
      ['avif', 'image/avif'],
    ])('should map extension %s to mime type %s', (ext, expected) => {
      expect(mimeFromExt(`test.${ext}`)).toBe(expected)
    })

    it.each([
      ['test.txt', 'image/png'],
      ['no-extension', 'image/png'],
      ['', 'image/png'],
      ['test.', 'image/png'],
      ['.hidden', 'image/png'],
    ])('should return default image/png for unknown, non-image, or edge case extensions (%s)', (filename, expected) => {
      expect(mimeFromExt(filename)).toBe(expected)
    })
  })
})
