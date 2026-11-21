import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import PrimeVue from 'primevue/config'
import { QueryClient, VueQueryPlugin } from '@tanstack/vue-query'
import WidgetSettingsDialog from './WidgetSettingsDialog.vue'
import type { Widget } from '@/types/dashboard'

function mountDialog(type: string) {
    const widget = { id: 'w1', type, title: 'My widget', width: 4, config: {} } as unknown as Widget
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
    return mount(WidgetSettingsDialog, {
        props: { widget, visible: true },
        global: {
            plugins: [PrimeVue, [VueQueryPlugin, { queryClient }]],
            // The real Dialog teleports to <body>; the stub keeps the width
            // (fallthrough style) and shows which slot each part lands in.
            stubs: {
                Dialog: {
                    template: '<div class="dialog-stub"><div class="stub-content"><slot /></div>'
                        + '<div class="stub-footer"><slot name="footer" /></div></div>',
                },
            },
        },
    })
}

describe('WidgetSettingsDialog', () => {
    it('opens wide for a widget that asks for it (weather)', () => {
        const wrapper = mountDialog('weather')
        expect(wrapper.find('.dialog-stub').attributes('style')).toContain('width: 60rem')
    })

    it.each(['weather-compact', 'battery', 'clock'])('keeps the default width for %s', (type) => {
        const wrapper = mountDialog(type)
        expect(wrapper.find('.dialog-stub').attributes('style')).toContain('width: 28rem')
    })

    it('puts the actions in the footer, outside the scrolling content', () => {
        const wrapper = mountDialog('battery')
        const footer = wrapper.find('.stub-footer').text()
        expect(footer).toContain('Save')
        expect(footer).toContain('Cancel')
        expect(footer).toContain('Delete')
        expect(wrapper.find('.stub-content').text()).not.toContain('Save')
    })
})
