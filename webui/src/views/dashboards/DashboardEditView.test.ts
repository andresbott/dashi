import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { ref } from 'vue'
import PrimeVue from 'primevue/config'
import DashboardEditView from './DashboardEditView.vue'
import type { Dashboard } from '@/types/dashboard'

const draft = ref<Dashboard | null>(null)
const flush = vi.fn(async () => {})
const push = vi.fn()

vi.mock('@/composables/useDashboardDraft', () => ({
    useDashboardDraft: () => ({
        draft,
        isLoading: ref(false),
        isError: ref(false),
        saveStatus: ref('idle'),
        hasUnsaved: ref(false),
        flush,
    }),
}))
vi.mock('vue-router', () => ({
    useRoute: () => ({ params: { id: 'd1' } }),
    useRouter: () => ({ push }),
}))

const mountView = () =>
    mount(DashboardEditView, {
        global: { plugins: [PrimeVue], stubs: { DashboardRow: true } },
    })

const previewButton = (wrapper: ReturnType<typeof mountView>) => wrapper.find('[data-testid="eink-preview-button"]')

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
        expect(previewButton(wrapper).exists()).toBe(false)
    })

    it('saves pending edits, then opens the preview page on the page being edited', async () => {
        const wrapper = mountView()
        await flushPromises()
        await wrapper.findAll('.page-tab')[1].trigger('click')
        await previewButton(wrapper).trigger('click')
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
