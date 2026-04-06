import { describe, it, expect, vi } from 'vitest'
import { useSettings } from './useSettings'
import { QueryClient, VueQueryPlugin } from '@tanstack/vue-query'
import { mount, flushPromises } from '@vue/test-utils'
import { defineComponent } from 'vue'
import * as settingsApi from '@/lib/api/settings'

vi.mock('@/lib/api/settings', () => ({
    getSettings: vi.fn(),
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

describe('useSettings', () => {
    it('fetches settings via TanStack Query', async () => {
        vi.mocked(settingsApi.getSettings).mockResolvedValue({ readOnly: false })

        let result: ReturnType<typeof useSettings>
        withQueryClient(() => {
            result = useSettings()
        })

        await flushPromises()
        expect(result!.data.value).toEqual({ readOnly: false })
        expect(settingsApi.getSettings).toHaveBeenCalledOnce()
    })
})
