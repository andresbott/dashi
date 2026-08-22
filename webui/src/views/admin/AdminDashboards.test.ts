import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { ref } from 'vue'
import PrimeVue from 'primevue/config'
import AdminDashboards from './AdminDashboards.vue'

const setDefaultDashboard = vi.fn()
const dashboards = ref<{ id: string; name: string; type?: string; default?: boolean }[]>([])

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
        },
    })
}

// The default button is the first action in each row.
function defaultButton(wrapper: ReturnType<typeof mountView>, rowIndex: number) {
    return wrapper.findAll('tbody tr')[rowIndex].findAll('button')[0]
}

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
