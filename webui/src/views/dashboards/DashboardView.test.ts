import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { ref } from 'vue'
import PrimeVue from 'primevue/config'
import DashboardView from './DashboardView.vue'
import type { Container, Dashboard } from '@/types/dashboard'

const dashboard = ref<Dashboard | undefined>(undefined)

vi.mock('@/composables/useDashboards', () => ({
    useGetDashboard: () => ({ data: dashboard, isLoading: ref(false), isError: ref(false) }),
    useListDashboards: () => ({ dashboards: ref([]) }),
}))
vi.mock('vue-router', () => ({
    useRoute: () => ({ params: { id: 'd1' }, query: {} }),
    useRouter: () => ({ push: vi.fn(), replace: vi.fn() }),
}))

// Plain strings: a hand-edited dashboard.json can hold anything.
const mountView = async (horizontalAlign: string, verticalAlign = 'top') => {
    dashboard.value = {
        id: 'd1', name: 'Kitchen', icon: '', type: 'interactive',
        container: {
            maxWidth: '1200px',
            verticalAlign: verticalAlign as Container['verticalAlign'],
            horizontalAlign: horizontalAlign as Container['horizontalAlign'],
        },
        pages: [{ name: 'Main', rows: [] }],
    } as Dashboard
    const wrapper = mount(DashboardView, { global: { plugins: [PrimeVue] } })
    await flushPromises()
    return wrapper
}

const style = (wrapper: Awaited<ReturnType<typeof mountView>>, selector: string) =>
    (wrapper.find(selector).element as HTMLElement).style

// Same placement as the server renders (internal/dashboard/alignment.go
// PlaceContainer) and the alignment picker.
describe('DashboardView container placement', () => {
    beforeEach(() => {
        dashboard.value = undefined
    })

    it.each(['', 'sideways', 'center'])('centres the container for horizontal align %j', async align => {
        const s = style(await mountView(align), '.dashboard-container')
        expect(s.marginLeft).toBe('auto')
        expect(s.marginRight).toBe('auto')
    })

    it.each([
        ['left', '0px', 'auto'],
        ['right', 'auto', '0px'],
    ])('pushes the container %s', async (align, left, right) => {
        const s = style(await mountView(align), '.dashboard-container')
        expect(s.marginLeft).toBe(left)
        expect(s.marginRight).toBe(right)
    })

    it('aligns rows vertically in a full-viewport column', async () => {
        const s = style(await mountView('center', 'bottom'), '.dashboard-rows')
        expect(s.minHeight).toBe('100vh')
        expect(s.justifyContent).toBe('flex-end')
    })
})
