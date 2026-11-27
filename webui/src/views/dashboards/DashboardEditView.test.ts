import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { ref } from 'vue'
import PrimeVue from 'primevue/config'
import DashboardEditView from './DashboardEditView.vue'
import type { Dashboard } from '@/types/dashboard'

const draft = ref<Dashboard | null>(null)
const saveStatus = ref('idle')
const flush = vi.fn(async () => {})
const push = vi.fn()

vi.mock('@/composables/useDashboardDraft', () => ({
    useDashboardDraft: () => ({
        draft,
        isLoading: ref(false),
        isError: ref(false),
        saveStatus,
        hasUnsaved: ref(false),
        flush,
    }),
}))
vi.mock('vue-router', () => ({
    useRoute: () => ({ params: { id: 'd1' } }),
    useRouter: () => ({ push }),
}))
vi.mock('@/lib/serverInfo', () => ({ dashboardViewUrl: (id: string) => `http://viewer/${id}` }))

const mountView = () =>
    mount(DashboardEditView, {
        global: { plugins: [PrimeVue], stubs: { DashboardRow: true } },
    })

const einkButton = (wrapper: ReturnType<typeof mountView>) => wrapper.find('[data-testid="eink-preview-button"]')
const viewerButton = (wrapper: ReturnType<typeof mountView>) => wrapper.find('[data-testid="viewer-preview-button"]')

describe('DashboardEditView e-ink preview button', () => {
    beforeEach(() => {
        flush.mockClear()
        push.mockClear()
        vi.stubGlobal('matchMedia', () => ({
            matches: false,
            addEventListener: vi.fn(),
            removeEventListener: vi.fn(),
        }))
        draft.value = {
            id: 'd1', name: 'Kitchen', icon: '', type: 'image',
            container: { maxWidth: '', verticalAlign: 'top', horizontalAlign: 'center' },
            pages: [{ name: 'Main', rows: [] }, { name: 'Weather', rows: [] }],
        } as Dashboard
    })

    afterEach(() => {
        vi.unstubAllGlobals()
    })

    it('is offered only for image dashboards', async () => {
        draft.value!.type = 'interactive'
        const wrapper = mountView()
        await flushPromises()
        expect(einkButton(wrapper).exists()).toBe(false)
    })

    it('saves pending edits, then opens the preview page on the page being edited', async () => {
        const wrapper = mountView()
        await flushPromises()
        await wrapper.findAll('.page-tab')[1].trigger('click')
        await einkButton(wrapper).trigger('click')
        await flushPromises()

        expect(flush).toHaveBeenCalledTimes(1)
        expect(push).toHaveBeenCalledWith({
            name: 'dashboard-eink-preview',
            params: { id: 'd1' },
            query: { page: '1', from: 'edit' },
        })
        expect(flush.mock.invocationCallOrder[0]).toBeLessThan(push.mock.invocationCallOrder[0])
    })
})

describe('DashboardEditView live viewer preview button', () => {
    let opened: { opener: unknown; location: { href: string } }
    let open: ReturnType<typeof vi.fn>

    beforeEach(() => {
        flush.mockReset()
        flush.mockImplementation(async () => {})
        saveStatus.value = 'idle'
        vi.stubGlobal('matchMedia', () => ({
            matches: false,
            addEventListener: vi.fn(),
            removeEventListener: vi.fn(),
        }))
        opened = { opener: {}, location: { href: '' } }
        open = vi.fn(() => opened)
        vi.stubGlobal('open', open)
        draft.value = {
            id: 'd1', name: 'Kitchen', icon: '', type: 'interactive',
            container: { maxWidth: '', verticalAlign: 'top', horizontalAlign: 'center' },
            pages: [{ name: 'Main', rows: [] }, { name: 'Weather', rows: [] }],
        } as Dashboard
    })

    afterEach(() => {
        vi.unstubAllGlobals()
    })

    it('is offered only for interactive dashboards', async () => {
        draft.value!.type = 'image'
        const wrapper = mountView()
        await flushPromises()
        expect(viewerButton(wrapper).exists()).toBe(false)
        expect(einkButton(wrapper).exists()).toBe(true)
    })

    it('opens a tab synchronously, saves, then points it at the dashboard root', async () => {
        let resolveFlush!: () => void
        flush.mockImplementation(() => new Promise<void>(r => { resolveFlush = r }))
        const wrapper = mountView()
        await flushPromises()
        await wrapper.findAll('.page-tab')[1].trigger('click')
        await viewerButton(wrapper).trigger('click')

        expect(open).toHaveBeenCalledWith('', '_blank')
        expect(opened.opener).toBeNull()
        expect(opened.location.href).toBe('')

        resolveFlush()
        await flushPromises()
        expect(flush).toHaveBeenCalledTimes(1)
        expect(opened.location.href).toBe('http://viewer/d1')
    })

    // flush() resolves after a failed save and reports it through saveStatus.
    it('still navigates the tab when the save fails', async () => {
        flush.mockImplementation(async () => { saveStatus.value = 'error' })
        const wrapper = mountView()
        await flushPromises()
        await viewerButton(wrapper).trigger('click')
        await flushPromises()
        expect(opened.location.href).toBe('http://viewer/d1')
    })

    it('still saves, without throwing, when the browser blocks the tab', async () => {
        open.mockReturnValue(null)
        const errorHandler = vi.fn()
        const wrapper = mount(DashboardEditView, {
            global: { plugins: [PrimeVue], stubs: { DashboardRow: true }, config: { errorHandler } },
        })
        await flushPromises()
        await viewerButton(wrapper).trigger('click')
        await flushPromises()
        expect(open).toHaveBeenCalledWith('', '_blank')
        expect(flush).toHaveBeenCalledTimes(1)
        expect(errorHandler).not.toHaveBeenCalled()
    })
})
