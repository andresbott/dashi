import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { ref, defineComponent, nextTick } from 'vue'
import { mount, flushPromises } from '@vue/test-utils'
import { QueryClient, VueQueryPlugin } from '@tanstack/vue-query'
import * as dashboardApi from '@/lib/api/dashboard'
import { useDashboardDraft, type UseDashboardDraftResult } from './useDashboardDraft'
import type { Dashboard } from '@/types/dashboard'

vi.mock('@/lib/api/dashboard', () => ({
    getDashboards: vi.fn(),
    getDashboard: vi.fn(),
    createDashboard: vi.fn(),
    updateDashboard: vi.fn(),
    deleteDashboard: vi.fn(),
    setDefaultDashboard: vi.fn(),
    getDashboardAssets: vi.fn(),
    uploadDashboardZip: vi.fn(),
    getDashboardAuth: vi.fn(),
    setDashboardAuth: vi.fn(),
    deleteDashboardAuth: vi.fn(),
}))
const toastAdd = vi.hoisted(() => vi.fn())
vi.mock('primevue/usetoast', () => ({ useToast: () => ({ add: toastAdd }) }))

function makeDashboard(id: string, name: string): Dashboard {
    return {
        id,
        name,
        icon: '',
        type: 'interactive',
        container: { maxWidth: '100%', verticalAlign: 'top', horizontalAlign: 'left' },
        pages: [{ name: 'Home', rows: [] }],
    }
}

// Mounts a component whose setup wires useDashboardDraft to a reactive id, and
// hands back the id ref, the composable result, the query client (to inspect
// the cache) and the wrapper (to unmount).
function setupDraft(initialId: string) {
    const id = ref(initialId)
    const queryClient = new QueryClient({
        defaultOptions: { queries: { retry: false } },
    })
    let result!: UseDashboardDraftResult
    const wrapper = mount(
        defineComponent({
            setup() {
                result = useDashboardDraft(() => id.value)
                return () => null
            },
        }),
        { global: { plugins: [[VueQueryPlugin, { queryClient }]] } },
    )
    return { id, result, queryClient, wrapper }
}

describe('useDashboardDraft', () => {
    beforeEach(() => vi.useFakeTimers())
    afterEach(() => {
        vi.useRealTimers()
        vi.restoreAllMocks()
    })

    it('seeds the draft as a deep copy of the server dashboard', async () => {
        vi.mocked(dashboardApi.getDashboard).mockResolvedValue(makeDashboard('1', 'One'))
        const { result, queryClient } = setupDraft('1')
        await flushPromises()

        expect(result.draft.value?.name).toBe('One')

        result.draft.value!.name = 'Changed'
        result.draft.value!.pages[0].name = 'Renamed'

        const cached = queryClient.getQueryData<Dashboard>(['dashboard', '1'])
        expect(cached?.name).toBe('One')
        expect(cached?.pages[0].name).toBe('Home')
        expect(result.hasUnsaved.value).toBe(true)
    })

    it('re-seeds the draft when the id changes', async () => {
        vi.mocked(dashboardApi.getDashboard).mockImplementation(async (id: string) =>
            makeDashboard(id, `Dashboard ${id}`),
        )
        vi.mocked(dashboardApi.updateDashboard).mockImplementation(async (_id, payload) => payload)
        const { id, result } = setupDraft('1')
        await flushPromises()
        expect(result.draft.value?.id).toBe('1')

        result.draft.value!.name = 'Edited One'
        id.value = '2'
        await nextTick()
        // The old copy is kept until its pending change has been flushed
        // (against its own id); it is never re-labelled with the new id.
        expect(result.draft.value === null || result.draft.value.id === '1').toBe(true)

        await flushPromises()
        expect(result.draft.value?.id).toBe('2')
        expect(result.draft.value?.name).toBe('Dashboard 2')

        const calls = vi.mocked(dashboardApi.updateDashboard).mock.calls
        expect(calls).toHaveLength(1)
        expect(calls[0][0]).toBe('1')
        expect(calls[0][1]).toMatchObject({ id: '1', name: 'Edited One' })
        expect(result.hasUnsaved.value).toBe(false)
    })

    it('shows the server reason when a save is rejected', async () => {
        vi.mocked(dashboardApi.getDashboard).mockResolvedValue(makeDashboard('1', 'One'))
        vi.mocked(dashboardApi.updateDashboard).mockRejectedValue({
            response: { status: 400, data: { error: '{"error":"invalid widget config: bookmark widget \\"b1\\""}' } },
        })
        toastAdd.mockClear()
        const { result } = setupDraft('1')
        await flushPromises()

        result.draft.value!.name = 'Changed'
        await result.flush()
        await flushPromises()

        expect(toastAdd).toHaveBeenCalledWith(expect.objectContaining({
            severity: 'error',
            detail: 'Failed to save dashboard: invalid widget config: bookmark widget "b1"',
        }))
    })

    it('registers a beforeunload guard while mounted and removes it on unmount', async () => {
        vi.mocked(dashboardApi.getDashboard).mockResolvedValue(makeDashboard('1', 'One'))
        const addSpy = vi.spyOn(window, 'addEventListener')
        const removeSpy = vi.spyOn(window, 'removeEventListener')

        const { result, wrapper } = setupDraft('1')
        await flushPromises()

        const call = addSpy.mock.calls.find(([type]) => type === 'beforeunload')
        expect(call).toBeDefined()
        const handler = call![1] as (e: Event) => void

        const clean = new Event('beforeunload', { cancelable: true })
        handler(clean)
        expect(clean.defaultPrevented).toBe(false)

        result.draft.value!.name = 'Changed'
        const unsaved = new Event('beforeunload', { cancelable: true })
        handler(unsaved)
        expect(unsaved.defaultPrevented).toBe(true)

        wrapper.unmount()
        expect(removeSpy).toHaveBeenCalledWith('beforeunload', handler)
    })
})
