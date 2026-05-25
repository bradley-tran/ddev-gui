import { describe, expect, it, vi, beforeEach } from 'vitest'
import {
  getProjectName,
  getProjectStatus,
  getProjectType,
  getPrimaryUrl,
  getMailpitUrl,
  isProjectRunning,
  isProjectStopped,
  parseProjectsJSON,
  uid,
  coerceToBool,
  getIsLinux,
  openUrl,
} from '../utils'
import type { DdevProject } from '../types'

describe('utils', () => {
  describe('getProjectName', () => {
    it('should prioritize name over project and projectname', () => {
      const p: DdevProject = {
        name: 'name-val',
        project: 'project-val',
        projectname: 'projectname-val',
      }
      expect(getProjectName(p)).toBe('name-val')
    })

    it('should prioritize project over projectname', () => {
      const p: DdevProject = {
        project: 'project-val',
        projectname: 'projectname-val',
      }
      expect(getProjectName(p)).toBe('project-val')
    })

    it('should use projectname if others are missing', () => {
      const p: DdevProject = {
        projectname: 'projectname-val',
      }
      expect(getProjectName(p)).toBe('projectname-val')
    })

    it('should return empty string if no name fields exist', () => {
      const p: DdevProject = {}
      expect(getProjectName(p)).toBe('')
    })
  })

  describe('getProjectStatus', () => {
    it('should prioritize status_desc over status and state', () => {
      const p: DdevProject = {
        status_desc: 'status_desc-val',
        status: 'status-val',
        state: 'state-val',
      }
      expect(getProjectStatus(p)).toBe('status_desc-val')
    })

    it('should prioritize status over state', () => {
      const p: DdevProject = {
        status: 'status-val',
        state: 'state-val',
      }
      expect(getProjectStatus(p)).toBe('status-val')
    })

    it('should use state if others are missing', () => {
      const p: DdevProject = {
        state: 'state-val',
      }
      expect(getProjectStatus(p)).toBe('state-val')
    })

    it('should return empty string if no status fields exist', () => {
      const p: DdevProject = {}
      expect(getProjectStatus(p)).toBe('')
    })
  })

  describe('getProjectType', () => {
    it('should return value of "type" key', () => {
      const p: DdevProject = { type: 'php' }
      expect(getProjectType(p)).toBe('php')
    })

    it('should return value of "projecttype" key', () => {
      const p: DdevProject = { projecttype: 'drupal' }
      expect(getProjectType(p)).toBe('drupal')
    })

    it('should be case-insensitive for keys', () => {
      const p = { TYPE: 'laravel' } as unknown as DdevProject
      expect(getProjectType(p)).toBe('laravel')

      const p2 = { ProjectType: 'wordpress' } as unknown as DdevProject
      expect(getProjectType(p2)).toBe('wordpress')
    })

    it('should return empty string if no type key found', () => {
      const p: DdevProject = { name: 'test' }
      expect(getProjectType(p)).toBe('')
    })
  })

  describe('getPrimaryUrl', () => {
    it('should prioritize httpsurl', () => {
      const p: DdevProject = {
        httpsurl: 'https://httpsurl.test',
        primary_url: 'https://primary.test',
        httpurl: 'http://httpurl.test',
      }
      expect(getPrimaryUrl(p)).toBe('https://httpsurl.test')
    })

    it('should prioritize HTTPS primary_url if httpsurl is missing', () => {
      const p: DdevProject = {
        primary_url: 'https://primary.test',
        httpurl: 'http://httpurl.test',
      }
      expect(getPrimaryUrl(p)).toBe('https://primary.test')
    })

    it('should fall back to non-HTTPS primary_url if httpsurl and HTTPS primary_url are missing', () => {
      const p: DdevProject = {
        primary_url: 'http://primary.test',
        httpurl: 'http://httpurl.test',
      }
      expect(getPrimaryUrl(p)).toBe('http://primary.test')
    })

    it('should fall back to httpurl if others are missing', () => {
      const p: DdevProject = {
        httpurl: 'http://httpurl.test',
      }
      expect(getPrimaryUrl(p)).toBe('http://httpurl.test')
    })

    it('should fall back to url if others are missing', () => {
      const p: DdevProject = {
        url: 'http://url.test',
      }
      expect(getPrimaryUrl(p)).toBe('http://url.test')
    })

    it('should return empty string if no URLs exist', () => {
      const p: DdevProject = {}
      expect(getPrimaryUrl(p)).toBe('')
    })
  })

  describe('getMailpitUrl', () => {
    it('should prioritize mailpit_https_url over mailpit_url', () => {
      const p: DdevProject = {
        mailpit_https_url: 'https://mailpit.test',
        mailpit_url: 'http://mailpit.test',
      }
      expect(getMailpitUrl(p)).toBe('https://mailpit.test')
    })

    it('should use mailpit_url if mailpit_https_url is missing', () => {
      const p: DdevProject = {
        mailpit_url: 'http://mailpit.test',
      }
      expect(getMailpitUrl(p)).toBe('http://mailpit.test')
    })

    it('should return empty string if no mailpit URLs exist', () => {
      const p: DdevProject = {}
      expect(getMailpitUrl(p)).toBe('')
    })
  })

  describe('isProjectRunning', () => {
    it('should return true if status contains "run"', () => {
      expect(isProjectRunning({ status: 'running' })).toBe(true)
      expect(isProjectRunning({ status_desc: 'Running' })).toBe(true)
    })

    it('should return false if status does not contain "run"', () => {
      expect(isProjectRunning({ status: 'stopped' })).toBe(false)
      expect(isProjectRunning({})).toBe(false)
    })
  })

  describe('isProjectStopped', () => {
    it('should return true if status contains "stop" or "pause"', () => {
      expect(isProjectStopped({ status: 'stopped' })).toBe(true)
      expect(isProjectStopped({ status_desc: 'Paused' })).toBe(true)
    })

    it('should return false if status does not contain "stop" or "pause"', () => {
      expect(isProjectStopped({ status: 'running' })).toBe(false)
      expect(isProjectStopped({})).toBe(false)
    })
  })

  describe('parseProjectsJSON', () => {
    it('should parse direct array', () => {
      const data = [{ name: 'p1' }, { name: 'p2' }]
      const result = parseProjectsJSON(JSON.stringify(data))
      expect(result).toHaveLength(2)
      expect(result[0]!.name).toBe('p1')
    })

    it('should parse raw property', () => {
      const data = { raw: [{ name: 'p1' }] }
      const result = parseProjectsJSON(JSON.stringify(data))
      expect(result).toHaveLength(1)
      expect(result[0]!.name).toBe('p1')
    })

    it('should parse items property', () => {
      const data = { items: [{ name: 'p1' }] }
      const result = parseProjectsJSON(JSON.stringify(data))
      expect(result).toHaveLength(1)
      expect(result[0]!.name).toBe('p1')
    })

    it('should return empty array on invalid JSON', () => {
      expect(parseProjectsJSON('invalid')).toEqual([])
    })

    it('should return empty array on JSON.parse error', () => {
      const parseSpy = vi.spyOn(JSON, 'parse').mockImplementation(() => {
        throw new Error('Test error')
      })

      expect(parseProjectsJSON('{"valid": "json"}')).toEqual([])

      parseSpy.mockRestore()
    })

    it('should return empty array if no expected keys found', () => {
      expect(parseProjectsJSON(JSON.stringify({ other: [] }))).toEqual([])
    })
  })

  describe('uid', () => {
    it('should generate a string starting with id_', () => {
      expect(uid()).toMatch(/^id_/)
    })

    it('should generate unique IDs', () => {
      const id1 = uid()
      const id2 = uid()
      expect(id1).not.toBe(id2)
    })
  })

  describe('coerceToBool', () => {
    it('should handle boolean inputs', () => {
      expect(coerceToBool(true)).toBe(true)
      expect(coerceToBool(false)).toBe(false)
    })

    it('should handle string "true" and "false"', () => {
      expect(coerceToBool('true')).toBe(true)
      expect(coerceToBool('false')).toBe(false)
    })

    it('should return defaultValue for other inputs', () => {
      expect(coerceToBool(null, true)).toBe(true)
      expect(coerceToBool(undefined, false)).toBe(false)
      expect(coerceToBool('random', true)).toBe(true)
    })
  })

  describe('getIsLinux', () => {
    it('should return true if userAgent contains "linux"', () => {
      vi.stubGlobal('navigator', { userAgent: 'Mozilla/5.0 (Linux; Android 10)' })
      expect(getIsLinux()).toBe(true)
    })

    it('should return true if platform contains "linux"', () => {
      vi.stubGlobal('navigator', { platform: 'Linux x86_64' })
      expect(getIsLinux()).toBe(true)
    })

    it('should return false if neither contain "linux"', () => {
      vi.stubGlobal('navigator', { userAgent: 'Macintosh', platform: 'MacIntel' })
      expect(getIsLinux()).toBe(false)
    })
  })

  describe('openUrl', () => {
    beforeEach(() => {
      vi.stubGlobal('window', {
        runtime: {
          EventsEmit: vi.fn(),
        },
        open: vi.fn(),
      })
    })

    it('should not do anything if url is empty', () => {
      openUrl('')
      expect(window.runtime!.EventsEmit).not.toHaveBeenCalled()
      expect(window.open).not.toHaveBeenCalled()
    })

    it('should use EventsEmit on Linux', () => {
      vi.stubGlobal('navigator', { userAgent: 'linux' })
      openUrl('https://test.com')
      expect(window.runtime!.EventsEmit).toHaveBeenCalledWith('open:url', { url: 'https://test.com' })
    })

    it('should use EventsEmit if openInBrowser is true', () => {
      vi.stubGlobal('navigator', { userAgent: 'mac' })
      openUrl('https://test.com', true)
      expect(window.runtime!.EventsEmit).toHaveBeenCalledWith('open:url', { url: 'https://test.com' })
    })

    it('should use window.open if not Linux and openInBrowser is false', () => {
      vi.stubGlobal('navigator', { userAgent: 'mac' })
      openUrl('https://test.com', false)
      expect(window.open).toHaveBeenCalledWith('https://test.com', '_blank', expect.any(String))
    })
  })
})
