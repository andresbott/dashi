import { describe, it, expect, vi } from 'vitest'
import { useListDashboards, useGetDashboard } from './useDashboards'
import { QueryClient, VueQueryPlugin } from '@tanstack/vue-query'
import { mount, flushPromises } from '@vue/test-utils'
import { defineComponent } from 'vue'
import * as dashboardApi from '@/lib/api/dashboard'

vi.mock('@/lib/api/dashboard', () => ({
    getDashboards: vi.fn(),
    getDashboard: vi.fn(),
    createDashboard: vi.fn(),
    updateDashboard: vi.fn(),
    deleteDashboard: vi.fn(),
    deletePreviews: vi.fn(),
    getBackgrounds: vi.fn(),
    getDashboardAssets: vi.fn(),
}))

function withQueryClient(setup: () => unknown) {
    const queryClient = new QueryClient({
        defaultOptions: { queries: { retry: false } },
    })
    const Wrapper = defineComponent({
        setup,
        template: '<div />',
    })
    return mount(Wrapper, {
        global: { plugins: [[VueQueryPlugin, { queryClient }]] },
    })
}

describe('useListDashboards', () => {
    it('fetches dashboard list', async () => {
        const items = [{ id: '1', name: 'Test' }]
        vi.mocked(dashboardApi.getDashboards).mockResolvedValue(items as any)

        let result: ReturnType<typeof useListDashboards>
        withQueryClient(() => {
            result = useListDashboards()
        })

        await flushPromises()
        expect(result!.dashboards.value).toEqual(items)
    })
})

describe('useGetDashboard', () => {
    it('fetches a single dashboard by ID', async () => {
        const dash = { id: '1', name: 'Test' }
        vi.mocked(dashboardApi.getDashboard).mockResolvedValue(dash as any)

        let result: ReturnType<typeof useGetDashboard>
        withQueryClient(() => {
            result = useGetDashboard(() => '1')
        })

        await flushPromises()
        expect(result!.data.value).toEqual(dash)
        expect(dashboardApi.getDashboard).toHaveBeenCalledWith('1')
    })
})
