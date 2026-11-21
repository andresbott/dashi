import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises, type VueWrapper } from '@vue/test-utils'
import PrimeVue from 'primevue/config'
import Select from 'primevue/select'
import InputNumber from 'primevue/inputnumber'
import EinkPreviewDialog from './EinkPreviewDialog.vue'

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
const flush = vi.fn(async () => {})
let objectUrls = 0

function mountDialog(props: Record<string, unknown> = {}) {
    return mount(EinkPreviewDialog, {
        props: {
            visible: true,
            dashboardId: 'd1',
            pageNames: ['Main', '', 'Weather'],
            initialPage: 1,
            flush,
            ...props,
        },
        global: {
            plugins: [PrimeVue],
            // The real Dialog teleports to <body>; the controls are what matter.
            stubs: { Dialog: { template: '<div><slot /></div>' } },
        },
    })
}

const byTestId = (wrapper: VueWrapper, id: string) =>
    wrapper.findAllComponents(Select).find(s => s.attributes('data-testid') === id)!

const requestedUrls = () => fetchMock.mock.calls.map(c => c[0])

// Lets the debounced reload fire and the fetch chain settle.
async function settle() {
    await flushPromises() // watchers run, scheduling the reload
    vi.advanceTimersByTime(1000)
    await flushPromises()
}

describe('EinkPreviewDialog', () => {
    beforeEach(() => {
        vi.useFakeTimers({ toFake: ['setTimeout', 'clearTimeout'] })
        fetchMock.mockReset()
        fetchMock.mockImplementation(async () => pngResponse())
        flush.mockClear()
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

    it('loads the first device preset on the editor page when opened', async () => {
        const wrapper = mountDialog()
        await settle()

        expect(requestedUrls()).toEqual(['/d1?format=png-bw&width=800&height=480&rotation=0&page=1'])
        expect(fetchMock.mock.calls[0][1]).toMatchObject({ cache: 'no-store' })
        // Pending edits are saved before the server renders.
        expect(flush).toHaveBeenCalledTimes(1)
        expect(flush.mock.invocationCallOrder[0]).toBeLessThan(fetchMock.mock.invocationCallOrder[0])
        expect(wrapper.find('[data-testid="eink-image"]').attributes('src')).toBe('blob:preview-1')
    })

    it('labels unnamed pages by number', async () => {
        const wrapper = mountDialog()
        await settle()
        const labels = (byTestId(wrapper, 'eink-page').props('options') as { label: string }[]).map(o => o.label)
        expect(labels).toEqual(['Main', 'Page 2', 'Weather'])
    })

    it('builds the URL from a device preset, rotation, page and format', async () => {
        const wrapper = mountDialog()
        await settle()

        byTestId(wrapper, 'eink-device').vm.$emit('update:modelValue', 's6-13in3')
        byTestId(wrapper, 'eink-rotation').vm.$emit('update:modelValue', 270)
        byTestId(wrapper, 'eink-page').vm.$emit('update:modelValue', 2)
        await settle()
        // One render for the whole burst of changes.
        expect(requestedUrls()).toHaveLength(2)
        expect(requestedUrls()[1]).toBe('/d1?format=png-spectra6&width=1600&height=1200&rotation=270&page=2')

        byTestId(wrapper, 'eink-format').vm.$emit('update:modelValue', 'png')
        await settle()
        expect(requestedUrls()[2]).toBe('/d1?format=png&width=1600&height=1200&rotation=270&page=2')
    })

    it('uses a clamped custom size', async () => {
        const wrapper = mountDialog()
        await settle()

        byTestId(wrapper, 'eink-device').vm.$emit('update:modelValue', 'custom')
        await flushPromises()
        const [w, h] = wrapper.findAllComponents(InputNumber)
        w.vm.$emit('update:modelValue', 1024)
        h.vm.$emit('update:modelValue', 99999)
        await settle()

        expect(requestedUrls().at(-1)).toBe('/d1?format=png-bw&width=1024&height=4096&rotation=0&page=1')
    })

    it('re-fetches the same render on Refresh', async () => {
        const wrapper = mountDialog()
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
        const wrapper = mountDialog()
        await settle()
        expect(wrapper.find('[data-testid="eink-error"]').text()).toContain('invalid X-Display-Width: 0')
        expect(wrapper.find('[data-testid="eink-image"]').exists()).toBe(false)
    })

    it('explains a non-image answer', async () => {
        fetchMock.mockImplementation(async () => ({
            ...pngResponse(),
            headers: { get: () => 'text/html; charset=utf-8' },
        }))
        const wrapper = mountDialog()
        await settle()
        expect(wrapper.find('[data-testid="eink-error"]').text()).toContain('Only image dashboards')
    })

    // The admin list opens the dialog with no draft to flush and no page.
    it('renders the first page without a flush, as the admin list opens it', async () => {
        const wrapper = mountDialog({ flush: undefined, initialPage: undefined, pageNames: [''] })
        await settle()
        expect(requestedUrls()).toEqual(['/d1?format=png-bw&width=800&height=480&rotation=0&page=0'])
        expect(flush).not.toHaveBeenCalled()
        expect(wrapper.find('[data-testid="eink-image"]').exists()).toBe(true)
    })

    it('does not fetch while closed', async () => {
        mountDialog({ visible: false })
        await settle()
        expect(fetchMock).not.toHaveBeenCalled()
    })
})
