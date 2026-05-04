import { describe, expect, it, beforeEach } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useAppStore, defaultModals } from '../app'

describe('app store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  describe('defaultModals', () => {
    it('should return initial state for all modals', () => {
      const modals = defaultModals()
      expect(modals).toEqual({
        newProject: false,
        envInfo: false,
        settings: false,
        about: false,
      })
    })
  })

  describe('initial state', () => {
    it('should initialize with default modal states', () => {
      const store = useAppStore()
      expect(store.modals).toEqual(defaultModals())
    })
  })

  describe('actions', () => {
    it('openModal should set modal state to true', () => {
      const store = useAppStore()
      store.openModal('settings')
      expect(store.modals.settings).toBe(true)
    })

    it('closeModal should set modal state to false', () => {
      const store = useAppStore()
      store.modals.settings = true
      store.closeModal('settings')
      expect(store.modals.settings).toBe(false)
    })

    it('closeAllModals should reset all modal states to false', () => {
      const store = useAppStore()
      store.modals.settings = true
      store.modals.about = true

      store.closeAllModals()

      expect(store.modals).toEqual(defaultModals())
      expect(store.modals.settings).toBe(false)
      expect(store.modals.about).toBe(false)
    })
  })
})
