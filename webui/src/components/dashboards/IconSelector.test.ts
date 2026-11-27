import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import { defineComponent, h, nextTick, ref } from 'vue'
import PrimeVue from 'primevue/config'
import IconSelector from './IconSelector.vue'

// The real Popover teleports to <body> once positioned; this stand-in keeps
// its contract (toggle/hide, show/hide events) and renders the content inline.
const PopoverStub = defineComponent({
    emits: ['show', 'hide'],
    setup(_, { slots, emit, expose }) {
        const visible = ref(false)
        const set = (v: boolean) => {
            visible.value = v
            emit(v ? 'show' : 'hide')
        }
        expose({ toggle: () => set(!visible.value), hide: () => set(false) })
        return () => (visible.value ? h('div', { 'data-testid': 'popover' }, slots.default?.()) : null)
    },
})

const mountSelector = (modelValue: string, options: { attachTo?: HTMLElement } = {}) =>
    mount(IconSelector, {
        props: { modelValue, labelledBy: 'icon-field-label' },
        global: { plugins: [PrimeVue], directives: { tooltip: {} }, stubs: { Popover: PopoverStub } },
        ...options,
    })

const trigger = (wrapper: ReturnType<typeof mountSelector>) => wrapper.find('[data-testid="icon-selector-trigger"]')

describe('IconSelector trigger', () => {
    it('is a button showing the chosen icon, its name and a dropdown chevron', () => {
        const wrapper = mountSelector('ti-bookmark')
        const t = trigger(wrapper)
        expect(t.element.tagName).toBe('BUTTON')
        expect(t.attributes('type')).toBe('button')
        expect(t.attributes('aria-haspopup')).toBe('dialog')
        expect(t.attributes('aria-expanded')).toBe('false')
        expect(t.find('i.ti.ti-bookmark').exists()).toBe(true)
        expect(t.find('.icon-selector-label').text()).toBe('bookmark')
        expect(t.find('i.ti.ti-chevron-down').exists()).toBe(true)
    })

    it('shows a muted placeholder when no icon is set', () => {
        const wrapper = mountSelector('')
        const t = trigger(wrapper)
        const label = t.find('.icon-selector-label')
        expect(label.text()).toBe('Select icon')
        expect(label.classes()).toContain('icon-selector-placeholder')
        expect(t.findAll('i').map(i => i.classes())).toEqual([['ti', 'ti-chevron-down', 'icon-selector-chevron']])
    })

    it('is named by the field label and the current value', () => {
        const wrapper = mountSelector('ti-bookmark')
        const t = trigger(wrapper)
        expect(t.attributes('aria-label')).toBeUndefined()
        const [labelId, valueId] = t.attributes('aria-labelledby')!.split(' ')
        expect(labelId).toBe('icon-field-label')
        expect(valueId).toBeTruthy()
        expect(t.find('.icon-selector-label').attributes('id')).toBe(valueId)
    })

    it('focuses the search field when the picker opens', async () => {
        const wrapper = mountSelector('ti-bookmark', { attachTo: document.body })
        await trigger(wrapper).trigger('click')
        await nextTick()
        expect(document.activeElement).toBe(wrapper.find('[data-testid="popover"] input').element)
        wrapper.unmount()
    })

    it('opens the icon picker and emits the picked icon', async () => {
        const wrapper = mountSelector('ti-bookmark')
        await trigger(wrapper).trigger('click')
        expect(wrapper.find('[data-testid="popover"] input').exists()).toBe(true)
        expect(trigger(wrapper).attributes('aria-expanded')).toBe('true')

        const option = wrapper.find('[data-testid="icon-option"]')
        const picked = option.find('.ti').classes().find(c => c.startsWith('ti-'))
        await option.trigger('click')
        expect(wrapper.emitted('update:modelValue')).toEqual([[picked]])
        expect(trigger(wrapper).attributes('aria-expanded')).toBe('false')
    })
})
