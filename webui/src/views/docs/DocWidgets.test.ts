import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises, type VueWrapper } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'
import PrimeVue from 'primevue/config'
import DocWidgets from './DocWidgets.vue'
import DocWidget from './DocWidget.vue'
import { getWidgetTypes, getWidgetEntry } from '@/lib/widgetRegistry'
import { WIDGET_CATEGORIES } from '@/widgets/categories'

// Queries without a doc mock (widget themes, icon lookups) fall through to the
// API client; stub it so nothing leaves the test, and so the tests can check
// that mocked widget data never reaches it.
const { get } = vi.hoisted(() => ({ get: vi.fn() }))
vi.mock('@/lib/api/client', () => ({ apiClient: { get } }))

async function mountAt(path: string, component: typeof DocWidgets | typeof DocWidget) {
    const router = createRouter({
        history: createMemoryHistory(),
        routes: [
            { path: '/admin/docs/widgets', name: 'doc-widgets', component: DocWidgets },
            { path: '/admin/docs/widgets/:type', name: 'doc-widget', component: DocWidget },
        ],
    })
    await router.push(path)
    await router.isReady()
    const wrapper = mount(component, {
        attachTo: document.body,
        global: { plugins: [router, PrimeVue], directives: { tooltip: {} } },
    })
    mounted.push(wrapper)
    await settle()
    return wrapper
}

// Widgets and forms are async components and their data arrives through
// (mocked) queries: wait for both.
async function settle() {
    for (let i = 0; i < 5; i++) {
        await vi.dynamicImportSettled()
        await flushPromises()
    }
}

const mounted: VueWrapper[] = []

beforeEach(() => {
    // PrimeVue's TabList watches its size and Select listens for orientation
    // changes; jsdom has neither API.
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

describe('widget docs index', () => {
    it('lists every registered widget exactly once, under its category', async () => {
        const wrapper = await mountAt('/admin/docs/widgets', DocWidgets)

        const sections = wrapper.findAll('.doc-widget-category')
        expect(sections.map(s => s.find('h3').text())).toEqual(WIDGET_CATEGORIES.map(c => c.label))

        const seen: string[] = []
        for (const section of sections) {
            for (const card of section.findAll('.doc-widget-card')) {
                const type = card.attributes('data-type')!
                expect(getWidgetEntry(type)?.category).toBe(section.attributes('data-category'))
                seen.push(type)
            }
        }
        expect(seen.sort()).toEqual([...getWidgetTypes()].sort())
    })

    it('uses the user-chosen grouping', async () => {
        const wrapper = await mountAt('/admin/docs/widgets', DocWidgets)
        const group = (id: string) =>
            wrapper.findAll(`[data-category="${id}"] .doc-widget-card`).map(c => c.attributes('data-type')).sort()

        expect(group('time')).toEqual(['clock'])
        expect(group('weather')).toEqual(['weather', 'weather-compact'])
        expect(group('information')).toEqual(['image', 'markdown', 'market', 'sysinfo', 'transport', 'xkcd'])
        expect(group('navigation')).toEqual(['battery', 'bookmark', 'page-indicator', 'search', 'stack'])
    })

    it('links each card to the widget page and shows its name under a preview', async () => {
        const wrapper = await mountAt('/admin/docs/widgets', DocWidgets)
        const card = wrapper.find('.doc-widget-card[data-type="transport"]')

        expect(card.attributes('href')).toBe('/admin/docs/widgets/transport')
        expect(card.find('.doc-widget-card-name').text()).toBe('Departures')
        expect(card.find('[data-testid="widget-preview"]').exists()).toBe(true)
    })

    it('renders data widgets from fixtures, without calling their APIs', async () => {
        const wrapper = await mountAt('/admin/docs/widgets', DocWidgets)

        expect(wrapper.find('[data-type="market"] .market-symbol').text()).toBe('AAPL')
        expect(wrapper.find('[data-type="transport"] .transport-row').exists()).toBe(true)
        expect(wrapper.find('[data-type="weather-compact"] .weather-compact-temp').text()).toBe('12°')
        const urls = get.mock.calls.map(c => String(c[0]))
        expect(urls.filter(u => u.startsWith('/widgets/') || u.startsWith('/data/'))).toEqual([])
    })
})

describe('widget docs page', () => {
    let wrapper: VueWrapper

    const tabs = () => wrapper.findAll('[role="tab"]').map(t => t.text())
    const preview = () => wrapper.find('[data-testid="widget-preview"]')

    it('shows the summary and the Overview and Details tabs', async () => {
        wrapper = await mountAt('/admin/docs/widgets/clock', DocWidget)

        expect(wrapper.find('h2').text()).toBe('Clock')
        expect(wrapper.find('.doc-widget-summary').text()).toContain('digital clock')
        expect(tabs()).toEqual(['Overview', 'Details'])
        expect(preview().find('.clock-time').exists()).toBe(true)
    })

    it('lists every parameter on the Details tab', async () => {
        wrapper = await mountAt('/admin/docs/widgets/clock', DocWidget)
        await wrapper.findAll('[role="tab"]')[1].trigger('click')
        await settle()

        const rows = wrapper.findAll('tr[data-param]').map(r => r.attributes('data-param'))
        expect(rows).toEqual(['hour12', 'showSeconds', 'showDate', 'font'])
    })

    it('updates the clock preview when a setting changes', async () => {
        wrapper = await mountAt('/admin/docs/widgets/clock', DocWidget)
        expect(preview().find('.clock-date').exists()).toBe(true)

        await wrapper.find('input#showDate').setValue(false)
        await settle()

        expect(preview().find('.clock-date').exists()).toBe(false)
    })

    it('updates the market preview from the mocked data when settings change', async () => {
        wrapper = await mountAt('/admin/docs/widgets/market', DocWidget)
        expect(preview().find('.market-symbol').text()).toBe('AAPL')
        expect(preview().find('.market-change').exists()).toBe(true)

        await wrapper.find('.doc-overview-settings input[type="text"]').setValue('tsla')
        await wrapper.find('input#marketShowChange').setValue(false)
        await settle()

        expect(preview().find('.market-symbol').text()).toBe('TSLA')
        expect(preview().find('.market-change').exists()).toBe(false)
        expect(get.mock.calls.map(c => c[0])).not.toContain('/widgets/market')
    })

    it('feeds the settings form from the fixtures too (sysinfo disks)', async () => {
        wrapper = await mountAt('/admin/docs/widgets/sysinfo', DocWidget)
        expect(preview().findAll('.sysinfo-label').map(l => l.text())).toContain('/srv/media')

        await wrapper.find('input[id="disk-/boot"]').setValue(true)
        await settle()

        expect(preview().findAll('.sysinfo-label').map(l => l.text())).toContain('/boot')
    })

    it('offers sample presets instead of the real form when it writes user data', async () => {
        wrapper = await mountAt('/admin/docs/widgets/markdown', DocWidget)
        expect(preview().text()).toContain('Welcome home')

        const option = wrapper.findAll('.p-togglebutton').find(b => b.text() === 'chores.md')!
        await option.trigger('click')
        await settle()

        expect(preview().text()).toContain('Water the plants')
    })

    it('handles an unknown widget type', async () => {
        wrapper = await mountAt('/admin/docs/widgets/nope', DocWidget)
        expect(wrapper.find('h2').text()).toBe('Unknown widget')
    })
})
