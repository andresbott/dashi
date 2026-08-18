import { describe, it, expect, vi } from 'vitest'
import { QueryClient, VueQueryPlugin } from '@tanstack/vue-query'
import { mount, flushPromises } from '@vue/test-utils'
import { defineComponent } from 'vue'
import { useDataItems } from './useDataItems'
import * as dataApi from '@/lib/api/data'

vi.mock('@/lib/api/data', () => ({
    listData: vi.fn(),
    uploadData: vi.fn(),
    deleteData: vi.fn(),
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

describe('useDataItems', () => {
    it('fetches the list for a kind', async () => {
        vi.mocked(dataApi.listData).mockResolvedValue([
            { name: 'a.png', size: 1, modTime: '2026-01-01T00:00:00Z' },
        ])
        let result: ReturnType<typeof useDataItems>
        withQueryClient(() => {
            result = useDataItems('images')
        })
        await flushPromises()
        expect(dataApi.listData).toHaveBeenCalledWith('images')
        expect(result!.items.value).toEqual([
            { name: 'a.png', size: 1, modTime: '2026-01-01T00:00:00Z' },
        ])
    })

    it('uploadItem calls uploadData with the kind', async () => {
        vi.mocked(dataApi.listData).mockResolvedValue([])
        vi.mocked(dataApi.uploadData).mockResolvedValue(undefined)
        let result: ReturnType<typeof useDataItems>
        withQueryClient(() => {
            result = useDataItems('backgrounds')
        })
        await flushPromises()
        const bytes = new ArrayBuffer(4)
        await result!.uploadItem({ name: 'bg.png', bytes })
        expect(dataApi.uploadData).toHaveBeenCalledWith('backgrounds', 'bg.png', bytes)
    })

    it('deleteItem calls deleteData with the kind', async () => {
        vi.mocked(dataApi.listData).mockResolvedValue([])
        vi.mocked(dataApi.deleteData).mockResolvedValue(undefined)
        let result: ReturnType<typeof useDataItems>
        withQueryClient(() => {
            result = useDataItems('images')
        })
        await flushPromises()
        await result!.deleteItem('x.png')
        expect(dataApi.deleteData).toHaveBeenCalledWith('images', 'x.png')
    })
})
