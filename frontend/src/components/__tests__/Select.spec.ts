import { describe, expect, it, afterEach } from 'vitest'
import { mount } from '@vue/test-utils'
import Select from '../Select.vue'

describe('Select', () => {
  const options = [
    { value: 'option1', label: 'Option 1' },
    { value: 'option2', label: 'Option 2' },
    { value: 'option3', label: 'Option 3' },
  ]

  afterEach(() => {
    // Clean up teleported elements
    document.body.innerHTML = ''
  })

  it('renders the initial selected label', () => {
    const wrapper = mount(Select, {
      props: {
        modelValue: 'option1',
        options,
      },
    })
    expect(wrapper.find('.custom-select-value').text()).toBe('Option 1')
  })

  it('renders the modelValue if no label is found', () => {
    const wrapper = mount(Select, {
      props: {
        modelValue: 'unknown',
        options,
      },
    })
    expect(wrapper.find('.custom-select-value').text()).toBe('unknown')
  })

  it('opens the menu when clicked', async () => {
    const wrapper = mount(Select, {
      props: {
        modelValue: 'option1',
        options,
      },
    })

    await wrapper.find('.custom-select-trigger').trigger('click')

    const menu = document.body.querySelector('.custom-select-menu')
    expect(menu).not.toBeNull()
    const optionElements = document.body.querySelectorAll('.custom-select-option')
    expect(optionElements).toHaveLength(3)
    expect(optionElements[0]!.textContent?.trim()).toBe('Option 1')
    expect(optionElements[1]!.textContent?.trim()).toBe('Option 2')
    expect(optionElements[2]!.textContent?.trim()).toBe('Option 3')
  })

  it('closes the menu when clicking the trigger again', async () => {
    const wrapper = mount(Select, {
      props: {
        modelValue: 'option1',
        options,
      },
    })

    const trigger = wrapper.find('.custom-select-trigger')
    await trigger.trigger('click') // Open
    expect(document.body.querySelector('.custom-select-menu')).not.toBeNull()

    await trigger.trigger('click') // Close
    expect(document.body.querySelector('.custom-select-menu')).toBeNull()
  })

  it('emits update:modelValue when an option is clicked and closes menu', async () => {
    const wrapper = mount(Select, {
      props: {
        modelValue: 'option1',
        options,
      },
    })

    await wrapper.find('.custom-select-trigger').trigger('click')

    const optionElements = document.body.querySelectorAll('.custom-select-option')
    await (optionElements[1] as HTMLElement).click()

    expect(wrapper.emitted('update:modelValue')).toBeTruthy()
    expect(wrapper.emitted('update:modelValue')?.[0]).toEqual(['option2'])

    // Menu should be closed
    expect(document.body.querySelector('.custom-select-menu')).toBeNull()
  })

  it('does not open when disabled', async () => {
    const wrapper = mount(Select, {
      props: {
        modelValue: 'option1',
        options,
        disabled: true,
      },
    })

    const trigger = wrapper.find('.custom-select-trigger')
    expect(trigger.attributes('disabled')).toBeDefined()

    await trigger.trigger('click')
    expect(document.body.querySelector('.custom-select-menu')).toBeNull()
  })

  it('sets focusedIndex to selected index when opening', async () => {
    const wrapper = mount(Select, {
      props: {
        modelValue: 'option2',
        options,
      },
    })

    await wrapper.find('.custom-select-trigger').trigger('click')

    const optionElements = document.body.querySelectorAll('.custom-select-option')
    expect(optionElements[1]!.classList.contains('focused')).toBe(true)
  })

  describe('Keyboard Navigation', () => {
    it('opens the menu on ArrowDown', async () => {
      const wrapper = mount(Select, {
        props: { modelValue: 'option1', options },
      })
      await wrapper.find('.custom-select-trigger').trigger('keydown', { key: 'ArrowDown' })
      expect(document.body.querySelector('.custom-select-menu')).not.toBeNull()
    })

    it('navigates with ArrowDown/ArrowUp when open', async () => {
      const wrapper = mount(Select, {
        props: { modelValue: 'option1', options },
      })
      const trigger = wrapper.find('.custom-select-trigger')
      await trigger.trigger('keydown', { key: 'ArrowDown' }) // Open

      await trigger.trigger('keydown', { key: 'ArrowDown' }) // Focus next (index 1)
      let optionsEls = document.body.querySelectorAll('.custom-select-option')
      expect(optionsEls[1]!.classList.contains('focused')).toBe(true)

      await trigger.trigger('keydown', { key: 'ArrowUp' }) // Focus previous (index 0)
      optionsEls = document.body.querySelectorAll('.custom-select-option')
      expect(optionsEls[0]!.classList.contains('focused')).toBe(true)

      // Wrap around ArrowUp
      await trigger.trigger('keydown', { key: 'ArrowUp' }) // index 2
      optionsEls = document.body.querySelectorAll('.custom-select-option')
      expect(optionsEls[2]!.classList.contains('focused')).toBe(true)

      // Wrap around ArrowDown
      await trigger.trigger('keydown', { key: 'ArrowDown' }) // index 0
      optionsEls = document.body.querySelectorAll('.custom-select-option')
      expect(optionsEls[0]!.classList.contains('focused')).toBe(true)
    })

    it('selects option with Enter when open', async () => {
      const wrapper = mount(Select, {
        props: { modelValue: 'option1', options },
      })
      const trigger = wrapper.find('.custom-select-trigger')
      await trigger.trigger('keydown', { key: 'ArrowDown' }) // Open
      await trigger.trigger('keydown', { key: 'ArrowDown' }) // Focus index 1

      await trigger.trigger('keydown', { key: 'Enter' })
      expect(wrapper.emitted('update:modelValue')?.[0]).toEqual(['option2'])
      expect(document.body.querySelector('.custom-select-menu')).toBeNull()
    })

    it('selects option with Space when open', async () => {
      const wrapper = mount(Select, {
        props: { modelValue: 'option1', options },
      })
      const trigger = wrapper.find('.custom-select-trigger')
      await trigger.trigger('keydown', { key: ' ' }) // Open
      await trigger.trigger('keydown', { key: 'ArrowDown' }) // Focus index 1

      await trigger.trigger('keydown', { key: ' ' })
      expect(wrapper.emitted('update:modelValue')?.[0]).toEqual(['option2'])
      expect(document.body.querySelector('.custom-select-menu')).toBeNull()
    })

    it('closes on Escape', async () => {
      const wrapper = mount(Select, {
        props: { modelValue: 'option1', options },
      })
      const trigger = wrapper.find('.custom-select-trigger')
      await trigger.trigger('click')
      expect(document.body.querySelector('.custom-select-menu')).not.toBeNull()

      await trigger.trigger('keydown', { key: 'Escape' })
      expect(document.body.querySelector('.custom-select-menu')).toBeNull()
    })

    it('jumps to Home and End', async () => {
      const wrapper = mount(Select, {
        props: { modelValue: 'option1', options },
      })
      const trigger = wrapper.find('.custom-select-trigger')
      await trigger.trigger('click')

      await trigger.trigger('keydown', { key: 'End' })
      let optionsEls = document.body.querySelectorAll('.custom-select-option')
      expect(optionsEls[2]!.classList.contains('focused')).toBe(true)

      await trigger.trigger('keydown', { key: 'Home' })
      optionsEls = document.body.querySelectorAll('.custom-select-option')
      expect(optionsEls[0]!.classList.contains('focused')).toBe(true)
    })
  })

  describe('Type-ahead search', () => {
    it('jumps to option starting with typed character when open', async () => {
      const wrapper = mount(Select, {
        props: {
          modelValue: 'option1',
          options: [
            { value: 'a', label: 'Apple' },
            { value: 'b', label: 'Banana' },
            { value: 'c', label: 'Cherry' },
          ]
        },
      })
      const trigger = wrapper.find('.custom-select-trigger')
      await trigger.trigger('click')

      await trigger.trigger('keydown', { key: 'c' })
      const optionsEls = document.body.querySelectorAll('.custom-select-option')
      expect(optionsEls[2]!.classList.contains('focused')).toBe(true)
    })

    it('selects option starting with typed character when closed', async () => {
      const wrapper = mount(Select, {
        props: {
          modelValue: 'a',
          options: [
            { value: 'a', label: 'Apple' },
            { value: 'b', label: 'Banana' },
            { value: 'c', label: 'Cherry' },
          ]
        },
      })
      const trigger = wrapper.find('.custom-select-trigger')
      await trigger.trigger('keydown', { key: 'b' })

      expect(wrapper.emitted('update:modelValue')?.[0]).toEqual(['b'])
    })

    it('accumulates characters for search', async () => {
      const wrapper = mount(Select, {
        props: {
          modelValue: 'a',
          options: [
            { value: 'ba', label: 'Banana' },
            { value: 'be', label: 'Berry' },
          ]
        },
      })
      const trigger = wrapper.find('.custom-select-trigger')

      await trigger.trigger('keydown', { key: 'b' })
      await trigger.trigger('keydown', { key: 'e' })

      const emissions = wrapper.emitted('update:modelValue')
      expect(emissions).toBeTruthy()
      expect(emissions?.length).toBeGreaterThanOrEqual(1)
      expect(emissions?.[emissions!.length - 1]).toEqual(['be'])
    })
  })

  describe('Outside interactions', () => {
    it('closes menu when clicking outside', async () => {
      const wrapper = mount(Select, {
        props: { modelValue: 'option1', options },
        global: {
          stubs: {
            Teleport: true,
          },
        },
      })

      const trigger = wrapper.find('.custom-select-trigger')
      await trigger.trigger('click')

      // Since Teleport is stubbed, the menu should be inside the wrapper
      expect(wrapper.find('.custom-select-menu').exists()).toBe(true)

      // Simulate click outside
      const mouseDownEvent = new MouseEvent('mousedown', {
        bubbles: true,
        cancelable: true,
      })
      document.dispatchEvent(mouseDownEvent)

      await wrapper.vm.$nextTick()
      expect(wrapper.find('.custom-select-menu').exists()).toBe(false)
    })
  })
})
