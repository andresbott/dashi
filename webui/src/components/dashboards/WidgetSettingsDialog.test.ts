import { describe, it, expect, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { reactive } from 'vue'
import PrimeVue from 'primevue/config'
import { QueryClient, VueQueryPlugin } from '@tanstack/vue-query'
import WidgetSettingsDialog from './WidgetSettingsDialog.vue'
import type { Widget } from '@/types/dashboard'

// These tests are about the dialog frame (width, footer), not the forms. The
// real settings forms are async components: left in, their imports resolve
// after the test ends and touch `window` once the environment is torn down.
// Swap each form for a plain stub, keeping every other registry field.
vi.mock('@/lib/widgetRegistry', async (importOriginal) => {
    const actual = await importOriginal<typeof import('@/lib/widgetRegistry')>()
    const { h } = await import('vue')
    const FormStub = { name: 'FormStub', render: () => h('div', { class: 'form-stub' }) }
    return {
        ...actual,
        getWidgetEntry: (type: string) => {
            const entry = actual.getWidgetEntry(type)
            return entry && { ...entry, configComponent: entry.configComponent ? FormStub : null }
        },
    }
})

type MountOptions = { config?: Record<string, unknown>; visible?: boolean; errorHandler?: () => void }

function mountDialog(type: string, { config = {}, visible = true, errorHandler }: MountOptions = {}) {
    const widget = { id: 'w1', type, title: 'My widget', width: 4, config } as unknown as Widget
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
    return mount(WidgetSettingsDialog, {
        props: { widget, visible },
        global: {
            plugins: [PrimeVue, [VueQueryPlugin, { queryClient }]],
            config: { errorHandler },
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
        expect(wrapper.find('.stub-content').text()).not.toContain('Save')
    })

    it('has no Delete action: delete lives in the widget card header', () => {
        const wrapper = mountDialog('battery')
        expect(wrapper.find('.stub-footer').text()).not.toContain('Delete')
    })

    it('renders the widget settings form inside the scrolling content', () => {
        const wrapper = mountDialog('weather')
        expect(wrapper.find('.stub-content .form-stub').exists()).toBe(true)
    })

    // A stack's config is a plain object, but setChildren leaves reactive
    // child proxies in its widgets array, out of toRaw's reach.
    it('stages a detached copy of a config that holds reactive children', async () => {
        const child = { id: 'c1', type: 'clock', title: '', width: 12, config: { format: '24h' } }
        const config = { widgets: [reactive(child)] }
        const errorHandler = vi.fn()
        const wrapper = mountDialog('stack', { config, visible: false, errorHandler })

        await wrapper.setProps({ visible: true })
        await flushPromises()
        expect(errorHandler).not.toHaveBeenCalled()

        // A stack has no settings form, so the staged config is what Save hands back.
        await wrapper.findAll('button').find(b => b.text().includes('Save'))!.trigger('click')
        const saved = (wrapper.emitted('save')![0][0] as Widget).config as Record<string, unknown>
        expect(saved).toEqual({ widgets: [child] })
        expect(saved).not.toBe(config)
        expect((saved.widgets as unknown[])[0]).not.toBe(config.widgets[0])
    })
})
