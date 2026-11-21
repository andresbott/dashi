import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises, enableAutoUnmount, type VueWrapper } from '@vue/test-utils'
import { ref } from 'vue'
import PrimeVue from 'primevue/config'
import Select from 'primevue/select'
import InputNumber from 'primevue/inputnumber'
import Checkbox from 'primevue/checkbox'
import DashboardEinkPreviewView from './DashboardEinkPreviewView.vue'

interface FakeResponse {
    ok: boolean
    status: number
    headers: { get: (name: string) => string | null }
    blob: () => Promise<Blob>
    text: () => Promise<string>
}

const pngResponse = (): FakeResponse => ({
    ok: true,
    status: 200,
    headers: { get: () => 'image/png' },
    blob: async () => new Blob(['png']),
    text: async () => '',
})

const fetchMock = vi.fn<(url: string, init?: RequestInit) => Promise<FakeResponse>>()
let objectUrls = 0

const dashboard = ref<{ id: string; name: string; type: string; pages: { name: string }[] } | undefined>()
const isLoading = ref(false)
const isError = ref(false)

vi.mock('@/composables/useDashboards', () => ({
    useGetDashboard: () => ({ data: dashboard, isLoading, isError }),
}))

const route = { params: { id: 'd1' }, query: {} as Record<string, string> }
const historyState: { back: string | null } = { back: null }
const replace = vi.fn()
const push = vi.fn()
const back = vi.fn()

vi.mock('vue-router', () => ({
    useRoute: () => route,
    useRouter: () => ({ replace, push, back, options: { history: { state: historyState } } }),
}))

function mountPage(query: Record<string, string> = {}) {
    route.query = query
    return mount(DashboardEinkPreviewView, { global: { plugins: [PrimeVue] } })
}

const byTestId = (wrapper: VueWrapper, id: string) =>
    wrapper.findAllComponents(Select).find(s => s.attributes('data-testid') === id)!

const requestedUrls = () => fetchMock.mock.calls.map(c => c[0])
const lastReplacedQuery = () => replace.mock.calls.at(-1)?.[0].query

// Lets the debounced reload fire and the fetch chain settle.
async function settle() {
    await flushPromises() // watchers run, scheduling the reload
    vi.advanceTimersByTime(1000)
    await flushPromises()
}

// Every test shares the dashboard ref: a page left mounted would react to the
// next test's changes.
enableAutoUnmount(afterEach)

describe('DashboardEinkPreviewView', () => {
    beforeEach(() => {
        vi.useFakeTimers({ toFake: ['setTimeout', 'clearTimeout'] })
        fetchMock.mockReset()
        fetchMock.mockImplementation(async () => pngResponse())
        replace.mockClear()
        push.mockClear()
        back.mockClear()
        historyState.back = null
        dashboard.value = { id: 'd1', name: 'Kitchen', type: 'image', pages: [{ name: 'Main' }, { name: '' }, { name: 'Weather' }] }
        isLoading.value = false
        isError.value = false
        objectUrls = 0
        vi.stubGlobal('fetch', fetchMock)
        vi.stubGlobal('matchMedia', () => ({
            matches: false,
            addEventListener: vi.fn(),
            removeEventListener: vi.fn(),
        }))
        URL.createObjectURL = vi.fn(() => `blob:preview-${++objectUrls}`)
        URL.revokeObjectURL = vi.fn()
    })

    afterEach(() => {
        vi.useRealTimers()
        vi.unstubAllGlobals()
    })

    it('renders the first device preset on the page from the query and writes the full query back', async () => {
        const wrapper = mountPage({ page: '1', from: 'edit' })
        await settle()

        expect(requestedUrls()).toEqual(['/d1?format=png-bw&width=800&height=480&rotation=0&page=1'])
        expect(fetchMock.mock.calls[0][1]).toMatchObject({ cache: 'no-store' })
        expect(wrapper.find('[data-testid="eink-image"]').attributes('src')).toBe('blob:preview-1')
        expect(wrapper.text()).toContain('Kitchen — E-ink preview')
        expect(replace).toHaveBeenCalledTimes(1)
        expect(replace.mock.calls[0][0]).toEqual({
            name: 'dashboard-eink-preview',
            params: { id: 'd1' },
            query: { preset: 'bw-7in5', format: 'png-bw', rotation: '0', page: '1', from: 'edit' },
        })
    })

    it('restores every setting from the query, so a reload shows the same render', async () => {
        const query = { preset: 's6-13in3', format: 'png', rotation: '270', page: '2', actual: '1', from: 'list' }
        const wrapper = mountPage(query)
        await settle()

        expect(requestedUrls()).toEqual(['/d1?format=png&width=1600&height=1200&rotation=270&page=2'])
        expect(byTestId(wrapper, 'eink-device').props('modelValue')).toBe('s6-13in3')
        expect(wrapper.findComponent(Checkbox).props('modelValue')).toBe(true)
        expect(wrapper.find('.eink-stage').classes()).toContain('eink-stage-actual')
        // Already normalised: nothing to rewrite.
        expect(replace).not.toHaveBeenCalled()
    })

    it('falls back to defaults for invalid query values', async () => {
        mountPage({ preset: 'nope', format: 'gif', rotation: '45', page: '-3', actual: 'yes' })
        await settle()

        expect(requestedUrls()).toEqual(['/d1?format=png-bw&width=800&height=480&rotation=0&page=0'])
        expect(lastReplacedQuery()).toEqual({ preset: 'bw-7in5', format: 'png-bw', rotation: '0', page: '0' })
    })

    it('falls back to the first page when the dashboard has fewer pages', async () => {
        const wrapper = mountPage({ page: '7' })
        await settle()
        expect(requestedUrls()).toEqual(['/d1?format=png-bw&width=800&height=480&rotation=0&page=0'])
        expect(byTestId(wrapper, 'eink-page').props('modelValue')).toBe(0)
    })

    it('reads a clamped custom size from the query', async () => {
        mountPage({ preset: 'custom', width: '1024', height: '99999' })
        await settle()
        expect(requestedUrls()).toEqual(['/d1?format=png-bw&width=1024&height=4096&rotation=0&page=0'])
        expect(lastReplacedQuery()).toMatchObject({ preset: 'custom', width: '1024', height: '4096' })
    })

    it('labels unnamed pages by number', async () => {
        const wrapper = mountPage()
        await settle()
        const labels = (byTestId(wrapper, 'eink-page').props('options') as { label: string }[]).map(o => o.label)
        expect(labels).toEqual(['Main', 'Page 2', 'Weather'])
    })

    it('re-renders once per burst of control changes and mirrors them in the query', async () => {
        const wrapper = mountPage()
        await settle()

        byTestId(wrapper, 'eink-device').vm.$emit('update:modelValue', 's6-13in3')
        byTestId(wrapper, 'eink-rotation').vm.$emit('update:modelValue', 90)
        byTestId(wrapper, 'eink-page').vm.$emit('update:modelValue', 2)
        await settle()
        expect(requestedUrls()).toHaveLength(2)
        expect(requestedUrls()[1]).toBe('/d1?format=png-spectra6&width=1600&height=1200&rotation=90&page=2')
        expect(lastReplacedQuery()).toEqual({ preset: 's6-13in3', format: 'png-spectra6', rotation: '90', page: '2' })

        byTestId(wrapper, 'eink-format').vm.$emit('update:modelValue', 'png')
        wrapper.findComponent(Checkbox).vm.$emit('update:modelValue', true)
        await settle()
        expect(requestedUrls()[2]).toBe('/d1?format=png&width=1600&height=1200&rotation=90&page=2')
        expect(lastReplacedQuery()).toMatchObject({ format: 'png', actual: '1' })
        // Display-only: Actual pixels does not re-render.
        expect(requestedUrls()).toHaveLength(3)
    })

    it('uses a clamped custom size typed into the inputs', async () => {
        const wrapper = mountPage()
        await settle()

        byTestId(wrapper, 'eink-device').vm.$emit('update:modelValue', 'custom')
        await flushPromises()
        const [w, h] = wrapper.findAllComponents(InputNumber)
        w.vm.$emit('update:modelValue', 1024)
        h.vm.$emit('update:modelValue', 99999)
        await settle()

        expect(requestedUrls().at(-1)).toBe('/d1?format=png-bw&width=1024&height=4096&rotation=0&page=0')
        expect(lastReplacedQuery()).toMatchObject({ preset: 'custom', width: '1024', height: '4096' })
    })

    it('re-fetches the same render on Refresh', async () => {
        const wrapper = mountPage()
        await settle()
        await wrapper.find('[data-testid="eink-refresh"]').trigger('click')
        await settle()
        expect(requestedUrls()).toHaveLength(2)
        expect(requestedUrls()[1]).toBe(requestedUrls()[0])
        expect(wrapper.find('[data-testid="eink-image"]').attributes('src')).toBe('blob:preview-2')
        expect(URL.revokeObjectURL).toHaveBeenCalledWith('blob:preview-1')
    })

    it('shows the server error for a rejected request', async () => {
        fetchMock.mockImplementation(async () => ({
            ok: false,
            status: 400,
            headers: { get: () => 'application/json' },
            blob: async () => new Blob(),
            text: async () => JSON.stringify({ error: 'invalid X-Display-Width: 0' }),
        }))
        const wrapper = mountPage()
        await settle()
        expect(wrapper.find('[data-testid="eink-error"]').text()).toContain('invalid X-Display-Width: 0')
        expect(wrapper.find('[data-testid="eink-image"]').exists()).toBe(false)
    })

    it('explains a non-image answer', async () => {
        fetchMock.mockImplementation(async () => ({
            ...pngResponse(),
            headers: { get: () => 'text/html; charset=utf-8' },
        }))
        const wrapper = mountPage()
        await settle()
        expect(wrapper.find('[data-testid="eink-error"]').text()).toContain('Only image dashboards')
    })

    it('shows a message instead of rendering an interactive dashboard', async () => {
        dashboard.value = { id: 'd1', name: 'Hall', type: 'interactive', pages: [{ name: '' }] }
        const wrapper = mountPage()
        await settle()
        expect(fetchMock).not.toHaveBeenCalled()
        expect(wrapper.find('[data-testid="eink-not-image"]').exists()).toBe(true)
        expect(wrapper.find('[data-testid="eink-device"]').exists()).toBe(false)
    })

    it('waits for the dashboard before rendering', async () => {
        dashboard.value = undefined
        isLoading.value = true
        mountPage()
        await settle()
        expect(fetchMock).not.toHaveBeenCalled()

        dashboard.value = { id: 'd1', name: 'Kitchen', type: 'image', pages: [{ name: '' }] }
        isLoading.value = false
        await settle()
        expect(requestedUrls()).toEqual(['/d1?format=png-bw&width=800&height=480&rotation=0&page=0'])
    })

    describe('upright display', () => {
        const frame = (wrapper: VueWrapper) => wrapper.find('[data-testid="eink-frame"]').element as HTMLElement
        const image = (wrapper: VueWrapper) => wrapper.find('[data-testid="eink-image"]').element as HTMLElement

        it.each([
            { rotation: '0', transform: 'none', disp: ['800', '480'] },
            { rotation: '90', transform: 'rotate(-90deg) translateX(-100%)', disp: ['480', '800'] },
            { rotation: '180', transform: 'rotate(180deg) translate(-100%, -100%)', disp: ['800', '480'] },
            { rotation: '270', transform: 'rotate(90deg) translateY(-100%)', disp: ['480', '800'] },
        ])('turns a $rotation° render back and sizes the frame for it', async ({ rotation, transform, disp }) => {
            const wrapper = mountPage({ rotation })
            await settle()

            expect(image(wrapper).style.transform).toBe(transform)
            expect(frame(wrapper).style.getPropertyValue('--disp-w')).toBe(disp[0])
            expect(frame(wrapper).style.getPropertyValue('--disp-h')).toBe(disp[1])
            // The image itself is always the panel's size.
            expect(frame(wrapper).style.getPropertyValue('--img-w')).toBe('800')
            expect(frame(wrapper).style.getPropertyValue('--img-h')).toBe('480')

            const note = wrapper.find('[data-testid="eink-upright-note"]')
            if (rotation === '0') {
                expect(note.exists()).toBe(false)
            } else {
                expect(note.text()).toContain(`panel receives it rotated ${rotation}°`)
            }
        })

        it('keeps the shown render shape until the next one arrives', async () => {
            const wrapper = mountPage()
            await settle()
            byTestId(wrapper, 'eink-rotation').vm.$emit('update:modelValue', 90)
            await flushPromises()
            // Not re-rendered yet: still the 0° image.
            expect(image(wrapper).style.transform).toBe('none')
            await settle()
            expect(image(wrapper).style.transform).toBe('rotate(-90deg) translateX(-100%)')
        })
    })

    describe('Back', () => {
        it('goes back in history when the page was reached inside the app', async () => {
            historyState.back = '/dashboards/d1/edit'
            const wrapper = mountPage({ from: 'edit' })
            await wrapper.find('[data-testid="eink-back"]').trigger('click')
            expect(back).toHaveBeenCalled()
            expect(push).not.toHaveBeenCalled()
        })

        it('opens the editor without history when it came from there', async () => {
            const wrapper = mountPage({ from: 'edit' })
            await wrapper.find('[data-testid="eink-back"]').trigger('click')
            expect(push).toHaveBeenCalledWith({ name: 'dashboard-edit', params: { id: 'd1' } })
        })

        it('opens the dashboards list otherwise', async () => {
            const wrapper = mountPage()
            await wrapper.find('[data-testid="eink-back"]').trigger('click')
            expect(push).toHaveBeenCalledWith({ name: 'admin-dashboards' })
        })
    })
})
