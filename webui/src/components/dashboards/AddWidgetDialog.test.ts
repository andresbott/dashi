import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises, type VueWrapper } from '@vue/test-utils'
import PrimeVue from 'primevue/config'
import { QueryClient, VueQueryPlugin } from '@tanstack/vue-query'
import AddWidgetDialog from './AddWidgetDialog.vue'
import { EDITING_MODE } from '@/lib/injectionKeys'
import { getWidgetTypeOptions } from '@/lib/widgetRegistry'

// Queries without a doc mock (widget themes, icon lookups) fall through to the
// API client; stub it so nothing leaves the test.
const { get } = vi.hoisted(() => ({ get: vi.fn() }))
vi.mock('@/lib/api/client', () => ({ apiClient: { get } }))

let appClient: QueryClient
const mounted: VueWrapper[] = []

function mountDialog(props: { exclude?: string[] } = {}) {
    appClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
    const wrapper = mount(AddWidgetDialog, {
        props: { visible: true, ...props },
        global: {
            plugins: [PrimeVue, [VueQueryPlugin, { queryClient: appClient }]],
            // The dashboard editor provides editing mode to everything below it.
            provide: { [EDITING_MODE as symbol]: true },
            directives: { tooltip: {} },
            // The real Dialog teleports to <body>; the stub renders its slots
            // in place and, like the real one, only while visible.
            stubs: {
                Dialog: {
                    props: ['visible'],
                    template: '<div v-if="visible" class="dialog-stub"><div class="stub-content"><slot /></div>'
                        + '<div class="stub-footer"><slot name="footer" /></div></div>',
                },
            },
        },
    })
    mounted.push(wrapper)
    return wrapper
}

// Widgets are async components and their data arrives through (mocked)
// queries: wait for both.
async function settle() {
    for (let i = 0; i < 5; i++) {
        await vi.dynamicImportSettled()
        await flushPromises()
    }
}

const types = (w: VueWrapper) => w.findAll('.widget-type-option').map(o => o.attributes('data-type'))
const option = (w: VueWrapper, type: string) => w.find(`.widget-type-option[data-type="${type}"]`)
const button = (w: VueWrapper, label: string) => w.findAll('.stub-footer button').find(b => b.text() === label)!
const preview = (w: VueWrapper) => w.find('[data-testid="widget-preview"]')

beforeEach(() => {
    vi.stubGlobal('ResizeObserver', class {
        observe() {}
        unobserve() {}
        disconnect() {}
    })
    vi.stubGlobal('matchMedia', () => ({ matches: false, addEventListener: vi.fn(), removeEventListener: vi.fn() }))
    get.mockReset()
    get.mockResolvedValue({ data: [] })
})

afterEach(() => {
    while (mounted.length) mounted.pop()!.unmount()
    vi.unstubAllGlobals()
})

describe('AddWidgetDialog', () => {
    it('lists every widget type except the excluded ones', () => {
        const wrapper = mountDialog({ exclude: ['stack'] })
        const all = getWidgetTypeOptions().map(o => o.value)
        expect(types(wrapper)).toEqual(all.filter(t => t !== 'stack'))
    })

    it('opens wide', () => {
        const wrapper = mountDialog()
        expect(wrapper.find('.dialog-stub').attributes('style')).toContain('width: 48rem')
    })

    it.each([
        ['label', 'departures', 'transport'],
        ['description', 'xkcd comic', 'xkcd'],
        ['type', 'WEATHER-COMPACT', 'weather-compact'],
    ])('filters by %s', async (_field, query, expected) => {
        const wrapper = mountDialog()
        await wrapper.find('input[type="text"]').setValue(query)
        expect(types(wrapper)).toContain(expected)
        expect(types(wrapper).length).toBeLessThan(getWidgetTypeOptions().length)
    })

    it('says so when no widget matches the filter', async () => {
        const wrapper = mountDialog()
        await wrapper.find('input[type="text"]').setValue('no-such-widget')
        expect(types(wrapper)).toEqual([])
        expect(wrapper.find('.stub-content').text()).toContain('No widgets match')
    })

    it('starts with nothing selected: Insert is disabled and the preview shows a hint', () => {
        const wrapper = mountDialog()
        expect(button(wrapper, 'Insert').attributes('disabled')).toBeDefined()
        expect(preview(wrapper).exists()).toBe(false)
        expect(wrapper.find('.add-widget-preview-hint').exists()).toBe(true)
    })

    it('selects a clicked widget and previews it with sample data, without inserting', async () => {
        const wrapper = mountDialog()
        await option(wrapper, 'clock').trigger('click')
        await settle()

        expect(option(wrapper, 'clock').attributes('aria-pressed')).toBe('true')
        expect(preview(wrapper).find('.clock-time').exists()).toBe(true)
        expect(wrapper.find('.stub-content').text()).toContain('Sample data')
        expect(button(wrapper, 'Insert').attributes('disabled')).toBeUndefined()
        expect(wrapper.emitted('select')).toBeUndefined()
    })

    it('previews mocked widget data without calling its API', async () => {
        const wrapper = mountDialog()
        await option(wrapper, 'market').trigger('click')
        await settle()

        expect(preview(wrapper).find('.market-symbol').text()).toBe('AAPL')
        expect(get.mock.calls.map(c => String(c[0])).filter(u => u.startsWith('/widgets/'))).toEqual([])
    })

    it('previews a stack in view mode, not with its drag editor', async () => {
        const wrapper = mountDialog()
        await option(wrapper, 'stack').trigger('click')
        await settle()

        expect(preview(wrapper).find('.widget-stack').exists()).toBe(true)
        expect(preview(wrapper).find('.widget-stack-editor').exists()).toBe(false)
    })

    it('inserts the selected widget and closes', async () => {
        const wrapper = mountDialog()
        await option(wrapper, 'clock').trigger('click')
        await button(wrapper, 'Insert').trigger('click')

        expect(wrapper.emitted('select')).toEqual([['clock']])
        expect(wrapper.emitted('update:visible')).toEqual([[false]])
    })

    it('inserts a widget directly on double-click', async () => {
        const wrapper = mountDialog()
        await option(wrapper, 'battery').trigger('dblclick')

        expect(wrapper.emitted('select')).toEqual([['battery']])
        expect(wrapper.emitted('update:visible')).toEqual([[false]])
    })

    it('closes on Cancel without inserting', async () => {
        const wrapper = mountDialog()
        await option(wrapper, 'clock').trigger('click')
        await button(wrapper, 'Cancel').trigger('click')

        expect(wrapper.emitted('select')).toBeUndefined()
        expect(wrapper.emitted('update:visible')).toEqual([[false]])
    })

    it('resets the filter and the selection when reopened', async () => {
        const wrapper = mountDialog()
        await wrapper.find('input[type="text"]').setValue('clock')
        await option(wrapper, 'clock').trigger('click')

        await wrapper.setProps({ visible: false })
        await wrapper.setProps({ visible: true })

        expect((wrapper.find('input[type="text"]').element as HTMLInputElement).value).toBe('')
        expect(types(wrapper).length).toBe(getWidgetTypeOptions().length)
        expect(button(wrapper, 'Insert').attributes('disabled')).toBeDefined()
        expect(preview(wrapper).exists()).toBe(false)
    })

    it('leaves the editor query client intact when the preview goes away', async () => {
        const wrapper = mountDialog()
        appClient.setQueryData(['dashboard', 'd1'], { id: 'd1' })
        await option(wrapper, 'market').trigger('click')
        await settle()
        expect(preview(wrapper).find('.market-symbol').exists()).toBe(true)

        await wrapper.setProps({ visible: false })
        await settle()

        expect(appClient.getQueryData(['dashboard', 'd1'])).toEqual({ id: 'd1' })
        expect(appClient.getQueryCache().find({ queryKey: ['market'] })).toBeUndefined()
    })
})
