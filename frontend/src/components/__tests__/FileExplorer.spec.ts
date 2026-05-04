import { describe, expect, it, type Mock } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import FileExplorer from '../FileExplorer.vue'

describe('FileExplorer', () => {
  it('loads a directory and previews a selected file', async () => {
    if (!window.go?.backend) {
      throw new Error('Wails backend mock is not available')
    }

    const ddevService = window.go.backend.DdevService as unknown as {
      ListDir: Mock
      ReadFile: Mock
    }

    ddevService.ListDir.mockImplementation(async (_project: string, relPath: string) => {
      if (relPath === '.') {
        return JSON.stringify([
          { name: 'app', isDir: true, size: '', modified: 'today' },
          { name: 'README.md', isDir: false, size: '12 B', modified: 'today' },
        ])
      }

      if (relPath === 'app') {
        return JSON.stringify([
          { name: 'index.php', isDir: false, size: '18 B', modified: 'today' },
        ])
      }

      return JSON.stringify([])
    })

    ddevService.ReadFile.mockImplementation(async (_project: string, relPath: string) => {
      if (relPath === 'README.md') {
        return '# Hello Explorer'
      }

      return '<?php echo "hi";'
    })

    const wrapper = mount(FileExplorer, {
      props: {
        projectName: 'demo',
        projectRoot: '/workspace/demo',
      },
    })

    await flushPromises()

    expect(ddevService.ListDir).toHaveBeenCalledWith('demo', '.')
    expect(wrapper.text()).toContain('README.md')
    expect(wrapper.text()).toContain('app')

    await wrapper.get('[data-entry-path="README.md"]').trigger('click')
    await flushPromises()

    expect(ddevService.ReadFile).toHaveBeenCalledWith('demo', 'README.md')
    expect(wrapper.text()).toContain('Hello Explorer')

    await wrapper.get('[data-entry-path="app"]').trigger('click')
    await flushPromises()

    expect(ddevService.ListDir).toHaveBeenCalledWith('demo', 'app')
    expect(wrapper.text()).toContain('index.php')
    expect(wrapper.find('[data-up-directory="true"]').exists()).toBe(true)
  })
})