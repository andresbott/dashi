import { describe, it, expect, vi, beforeEach } from 'vitest'
import { listBackgrounds, getBackground } from '@/lib/api/backgrounds'
import { useListBackgrounds, useGetBackground } from '@/composables/useBackgrounds'
import { QueryClient, VueQueryPlugin } from '@tanstack/vue-query'
import { mount, flushPromises } from '@vue/test-utils'
import { defineComponent } from 'vue'

vi.mock('@/lib/api/backgrounds')

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

describe('useBackgrounds', () => {
    beforeEach(() => vi.clearAllMocks())

    it('exposes the background list', async () => {
        vi.mocked(listBackgrounds).mockResolvedValue([
            { id: 'a1', name: 'A', usedBy: 2, previewCss: ':root{}' },
        ])

        let result: ReturnType<typeof useListBackgrounds>
        withQueryClient(() => {
            result = useListBackgrounds()
        })

        await flushPromises()
        expect(result!.backgrounds.value?.[0].usedBy).toBe(2)
    })

    it('fetches one background reactively by id', async () => {
        vi.mocked(getBackground).mockResolvedValue({ id: 'a1', name: 'A' })

        let result: ReturnType<typeof useGetBackground>
        withQueryClient(() => {
            result = useGetBackground(() => 'a1')
        })

        await flushPromises()
        expect(getBackground).toHaveBeenCalledWith('a1')
        expect(result!.data.value?.name).toBe('A')
    })
})
