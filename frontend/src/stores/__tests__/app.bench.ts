import { describe, bench } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useAppStore } from '../app'

describe('app store getters', () => {
  setActivePinia(createPinia())
  const store = useAppStore()

  // Generate 1000 mock projects
  const mockProjects = Array.from({ length: 1000 }, (_, i) => ({
    name: `project-${i}`,
    status: 'running',
    approot: `/path/to/project-${i}`,
    httpurl: `http://project-${i}.ddev.site`,
    httpsurl: `https://project-${i}.ddev.site`,
    primary_url: `https://project-${i}.ddev.site`,
    type: 'php',
    shortroot: `~/project-${i}`,
    docroot: '',
    php_version: '8.1',
    router_status: 'starting'
  }))

  const mockJSON = JSON.stringify(mockProjects)

  let counter = 0
  bench('both getters calculation', () => {
    // Modify slightly each iteration to trigger reactivity
    store.setProjectsJSON(mockJSON + ' '.repeat(counter % 10))
    // Access both getters
    store.projects
    store.projectsMap
    counter++
  })
})
