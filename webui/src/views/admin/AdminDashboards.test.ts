import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { ref } from 'vue'
import PrimeVue from 'primevue/config'
import AdminDashboards from './AdminDashboards.vue'

const setDefaultDashboard = vi.fn()
const duplicateDashboard = vi.fn()
const toastAdd = vi.fn()
const push = vi.fn()
const dashboards = ref<{ id: string; name: string; type?: string; default?: boolean }[]>([])

vi.mock('@/composables/useDashboards', () => ({
    useListDashboards: () => ({
        dashboards,
        isLoading: ref(false),
        createDashboard: vi.fn(),
        deleteDashboard: vi.fn(),
        duplicateDashboard,
        uploadZip: vi.fn(),
        isUploadingZip: ref(false),
        setDefaultDashboard,
    }),
}))

vi.mock('@/lib/api/dashboard', () => ({ downloadDashboard: vi.fn() }))
vi.mock('@/lib/serverInfo', () => ({ dashboardViewUrl: (id: string) => `http://viewer/${id}` }))
vi.mock('vue-router', () => ({
    useRouter: () => ({ push, resolve: () => ({ href: '/somewhere' }) }),
}))
vi.mock('primevue/usetoast', () => ({ useToast: () => ({ add: toastAdd }) }))

function mountView() {
    return mount(AdminDashboards, {
        global: {
            plugins: [PrimeVue],
            directives: { tooltip: {} },
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
        push.mockClear()
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

    it('opens the preview page of the clicked dashboard, marked as coming from the list', async () => {
        const wrapper = mountView()
        await previewButton(wrapper, 1).trigger('click')
        await flushPromises()
        expect(push).toHaveBeenCalledWith({
            name: 'dashboard-eink-preview',
            params: { id: 'bbb222' },
            query: { from: 'list' },
        })

        await previewButton(wrapper, 3).trigger('click')
        await flushPromises()
        expect(push).toHaveBeenLastCalledWith(expect.objectContaining({ params: { id: 'ddd444' } }))
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

describe('AdminDashboards action buttons', () => {
    beforeEach(() => {
        vi.clearAllMocks()
        dashboards.value = [
            { id: 'aaa111', name: 'Alpha', type: 'interactive' },
            { id: 'bbb222', name: 'Bravo', type: 'image' },
        ]
    })

    function rowButtonLabels(wrapper: ReturnType<typeof mountView>, rowIndex: number) {
        return wrapper
            .findAll('tbody tr')[rowIndex]
            .findAll('button, a')
            .filter((el) => el.attributes('aria-hidden') !== 'true')
            .map((el) => el.attributes('aria-label'))
    }

    it('labels every action and keeps them in a stable order, delete last', () => {
        const wrapper = mountView()
        expect(rowButtonLabels(wrapper, 0)).toEqual([
            'Make default dashboard',
            'Open dashboard in a new tab',
            'Edit layout and widgets',
            'Dashboard settings',
            'Duplicate dashboard',
            'Export as ZIP',
            'Delete dashboard',
        ])
        expect(rowButtonLabels(wrapper, 1)).toEqual([
            'Make default dashboard',
            'Open dashboard in a new tab',
            'E-ink preview (PNG)',
            'Edit layout and widgets',
            'Dashboard settings',
            'Duplicate dashboard',
            'Export as ZIP',
            'Delete dashboard',
        ])
    })

    it('keeps the same number of slots on interactive and image rows so columns line up', () => {
        const wrapper = mountView()
        const rows = wrapper.findAll('tbody tr')
        expect(rows[0].findAll('button, a').length).toBe(rows[1].findAll('button, a').length)
    })

    it('duplicates the clicked dashboard', async () => {
        duplicateDashboard.mockResolvedValue({ id: 'new001', name: 'Bravo (copy)' })
        const wrapper = mountView()
        await wrapper.findAll('tbody tr')[1].find('[data-testid="duplicate-button"]').trigger('click')
        await flushPromises()

        expect(duplicateDashboard).toHaveBeenCalledWith('bbb222')
        expect(toastAdd).toHaveBeenCalledWith(expect.objectContaining({ severity: 'success' }))
    })

    it('reports a failed duplicate', async () => {
        duplicateDashboard.mockRejectedValue(new Error('boom'))
        const wrapper = mountView()
        await wrapper.findAll('tbody tr')[0].find('[data-testid="duplicate-button"]').trigger('click')
        await flushPromises()

        expect(toastAdd).toHaveBeenCalledWith(expect.objectContaining({ severity: 'error' }))
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
