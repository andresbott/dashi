import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { ref, computed } from 'vue'
import PrimeVue from 'primevue/config'
import AdminDashboards from './AdminDashboards.vue'
import EinkPreviewDialog from '@/components/dashboards/EinkPreviewDialog.vue'

const setDefaultDashboard = vi.fn()
const dashboards = ref<{ id: string; name: string; type?: string; default?: boolean }[]>([])
// Full dashboards by id, as GET /api/v0/dashboards/{id} would return them.
const fullDashboards = ref<Record<string, { id: string; pages: { name: string }[] }>>({})

vi.mock('@/composables/useDashboards', () => ({
    useListDashboards: () => ({
        dashboards,
        isLoading: ref(false),
        createDashboard: vi.fn(),
        deleteDashboard: vi.fn(),
        uploadZip: vi.fn(),
        isUploadingZip: ref(false),
        setDefaultDashboard,
    }),
    useGetDashboard: (id: () => string) => ({
        data: computed(() => fullDashboards.value[id()]),
    }),
}))

vi.mock('@/lib/api/dashboard', () => ({ downloadDashboard: vi.fn() }))
vi.mock('@/lib/serverInfo', () => ({ dashboardViewUrl: (id: string) => `http://viewer/${id}` }))
vi.mock('vue-router', () => ({
    useRouter: () => ({ push: vi.fn(), resolve: () => ({ href: '/somewhere' }) }),
}))
vi.mock('primevue/usetoast', () => ({ useToast: () => ({ add: vi.fn() }) }))

function mountView() {
    return mount(AdminDashboards, {
        global: {
            plugins: [PrimeVue],
            directives: { tooltip: {} },
            // The dialog has its own tests; here only what the list hands it matters.
            stubs: { EinkPreviewDialog: true },
        },
    })
}

function defaultButton(wrapper: ReturnType<typeof mountView>, rowIndex: number) {
    return wrapper.findAll('tbody tr')[rowIndex].find('button[aria-label$="efault dashboard"]')
}

function previewButton(wrapper: ReturnType<typeof mountView>, rowIndex: number) {
    return wrapper.findAll('tbody tr')[rowIndex].find('[data-testid="eink-preview-button"]')
}

describe('AdminDashboards e-ink preview', () => {
    beforeEach(() => {
        fullDashboards.value = {
            bbb222: { id: 'bbb222', pages: [{ name: 'Main' }, { name: '' }] },
            ddd444: { id: 'ddd444', pages: [{ name: 'Only' }] },
        }
        dashboards.value = [
            { id: 'aaa111', name: 'Alpha', type: 'interactive' },
            { id: 'bbb222', name: 'Bravo', type: 'image' },
            { id: 'ccc333', name: 'Charlie' },
            { id: 'ddd444', name: 'Delta', type: 'image' },
        ]
    })

    it('offers the preview only on image dashboards', () => {
        const wrapper = mountView()
        expect(previewButton(wrapper, 0).exists()).toBe(false)
        expect(previewButton(wrapper, 1).exists()).toBe(true)
        expect(previewButton(wrapper, 1).attributes('aria-label')).toBe('E-ink preview (PNG)')
        expect(previewButton(wrapper, 2).exists()).toBe(false)
        expect(previewButton(wrapper, 3).exists()).toBe(true)
    })

    it('does not mount the dialog until a preview is opened', () => {
        const wrapper = mountView()
        expect(wrapper.findComponent(EinkPreviewDialog).exists()).toBe(false)
    })

    it('opens the dialog for the clicked dashboard with its page names and no flush', async () => {
        const wrapper = mountView()
        await previewButton(wrapper, 1).trigger('click')
        await flushPromises()

        const dialog = wrapper.findComponent(EinkPreviewDialog)
        expect(dialog.props()).toMatchObject({ visible: true, dashboardId: 'bbb222', pageNames: ['Main', ''] })
        expect(dialog.props('flush')).toBeUndefined()
        expect(dialog.props('initialPage')).toBeUndefined()

        // Close, then open another dashboard: the same dialog switches to it.
        dialog.vm.$emit('update:visible', false)
        await flushPromises()
        expect(wrapper.findComponent(EinkPreviewDialog).props('visible')).toBe(false)

        await previewButton(wrapper, 3).trigger('click')
        await flushPromises()
        expect(wrapper.findComponent(EinkPreviewDialog).props()).toMatchObject({
            visible: true,
            dashboardId: 'ddd444',
            pageNames: ['Only'],
        })
    })

    it('offers page 1 while the dashboard is still loading', async () => {
        fullDashboards.value = {}
        const wrapper = mountView()
        await previewButton(wrapper, 1).trigger('click')
        await flushPromises()
        expect(wrapper.findComponent(EinkPreviewDialog).props('pageNames')).toEqual([''])
    })
})

describe('AdminDashboards type icon', () => {
    function typeIcon(wrapper: ReturnType<typeof mountView>, rowIndex: number) {
        return wrapper.findAll('tbody tr')[rowIndex].find('i.dashboard-type-icon')
    }

    it('shows a distinct icon per dashboard type, falling back to interactive', () => {
        dashboards.value = [
            { id: 'aaa111', name: 'Alpha', type: 'interactive' },
            { id: 'bbb222', name: 'Bravo', type: 'image' },
            { id: 'ccc333', name: 'Charlie' },
        ]
        const wrapper = mountView()

        expect(typeIcon(wrapper, 0).classes()).toContain('ti-device-desktop')
        expect(typeIcon(wrapper, 0).attributes('aria-label')).toBe('Interactive dashboard')
        expect(typeIcon(wrapper, 1).classes()).toContain('ti-photo')
        expect(typeIcon(wrapper, 1).attributes('aria-label')).toBe('Image dashboard')
        expect(typeIcon(wrapper, 2).classes()).toContain('ti-device-desktop')
    })
})

describe('AdminDashboards default dashboard button', () => {
    beforeEach(() => {
        vi.clearAllMocks()
        dashboards.value = [
            { id: 'aaa111', name: 'Alpha', default: true },
            { id: 'bbb222', name: 'Bravo' },
        ]
    })

    it('marks a non-default dashboard as the default', async () => {
        const wrapper = mountView()
        await defaultButton(wrapper, 1).trigger('click')
        await flushPromises()

        expect(setDefaultDashboard).toHaveBeenCalledWith('bbb222')
    })

    it('re-asserts the default when the current default is clicked, so a stale second default gets cleared', async () => {
        dashboards.value = [
            { id: 'aaa111', name: 'Alpha', default: true },
            { id: 'bbb222', name: 'Bravo', default: true },
        ]
        const wrapper = mountView()
        await defaultButton(wrapper, 0).trigger('click')
        await flushPromises()

        expect(setDefaultDashboard).toHaveBeenCalledWith('aaa111')
    })
})
