import { describe, it, expect, vi, beforeEach } from 'vitest'
import { listThemes, getTheme, copyTheme } from '@/lib/api/themes'
import { useListThemes, useGetTheme } from '@/composables/useThemes'
import { QueryClient, VueQueryPlugin } from '@tanstack/vue-query'
import { mount, flushPromises } from '@vue/test-utils'
import { defineComponent } from 'vue'
import type { Palette, Theme } from '@/types/theme'

vi.mock('@/lib/api/themes')

const palette: Palette = {
    bg: '#ffffff', surface: '#f5f5f7', border: '#e0e0e5', text: '#333333', textMuted: '#666666',
    primary: '#0a84ff', onPrimary: '#ffffff', secondary: '#8b5cf6', success: '#16a34a',
    warning: '#d97706', danger: '#dc2626',
}

const theme = (over: Partial<Theme> = {}): Theme => ({
    id: 'a1', name: 'A', darkMode: true, background: { light: {}, dark: {} }, colors: { light: palette, dark: palette }, ...over,
})

function withQueryClient(setup: () => unknown) {
    const queryClient = new QueryClient({
        defaultOptions: { queries: { retry: false } },
    })
    const Wrapper = defineComponent({
        setup,
        template: '<div />',
    })
    mount(Wrapper, {
        global: { plugins: [[VueQueryPlugin, { queryClient }]] },
    })
    return queryClient
}

describe('useThemes', () => {
    beforeEach(() => vi.clearAllMocks())

    it('exposes the theme list', async () => {
        vi.mocked(listThemes).mockResolvedValue([
            { id: 'a1', name: 'A', darkMode: true, usedBy: 2, previewCss: ':root{}', colors: { light: palette, dark: palette } },
        ])

        let result: ReturnType<typeof useListThemes>
        withQueryClient(() => {
            result = useListThemes()
        })

        await flushPromises()
        expect(result!.themes.value?.[0].usedBy).toBe(2)
    })

    it('fetches one theme reactively by id', async () => {
        vi.mocked(getTheme).mockResolvedValue(theme())

        let result: ReturnType<typeof useGetTheme>
        withQueryClient(() => {
            result = useGetTheme(() => 'a1')
        })

        await flushPromises()
        expect(getTheme).toHaveBeenCalledWith('a1')
        expect(result!.data.value?.name).toBe('A')
    })

    it('copyTheme returns the copy and invalidates the theme list', async () => {
        vi.mocked(listThemes).mockResolvedValue([])
        const copy = theme({ id: 'def456', name: 'Copy of A' })
        vi.mocked(copyTheme).mockResolvedValue(copy)

        let result: ReturnType<typeof useListThemes>
        const queryClient = withQueryClient(() => {
            result = useListThemes()
        })
        await flushPromises()
        const spy = vi.spyOn(queryClient, 'invalidateQueries')

        const got = await result!.copyTheme('abc123')
        await flushPromises()

        expect(copyTheme).toHaveBeenCalledWith('abc123')
        expect(got).toEqual(copy)
        expect(spy).toHaveBeenCalledWith(expect.objectContaining({ queryKey: ['themes'] }))
    })
})
