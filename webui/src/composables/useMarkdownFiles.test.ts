import { describe, it, expect, vi, beforeEach } from 'vitest'
import { ref } from 'vue'
import { useMarkdownFiles } from './useMarkdownFiles'
import { QueryClient, VueQueryPlugin } from '@tanstack/vue-query'
import { mount, flushPromises } from '@vue/test-utils'
import { defineComponent } from 'vue'
import * as markdownApi from '@/lib/api/markdown'

vi.mock('@/lib/api/markdown', () => ({
    listMarkdownFiles: vi.fn(),
    getMarkdownHtml: vi.fn(),
    getMarkdownRaw: vi.fn(),
    saveMarkdown: vi.fn(),
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

describe('useMarkdownFiles', () => {
    beforeEach(() => {
        vi.clearAllMocks()
    })

    it('fetches the file list when dashboardId is set', async () => {
        vi.mocked(markdownApi.listMarkdownFiles).mockResolvedValue(['a.md', 'b.md'])

        let result: ReturnType<typeof useMarkdownFiles>
        withQueryClient(() => {
            result = useMarkdownFiles(ref('abc123'))
        })

        await flushPromises()
        expect(result!.files.value).toEqual(['a.md', 'b.md'])
        expect(markdownApi.listMarkdownFiles).toHaveBeenCalledWith('abc123')
    })

    it('does not fetch when dashboardId is empty', async () => {
        vi.mocked(markdownApi.listMarkdownFiles).mockResolvedValue([])

        withQueryClient(() => {
            useMarkdownFiles(ref(''))
        })

        await flushPromises()
        expect(markdownApi.listMarkdownFiles).not.toHaveBeenCalled()
    })

    it('exposes invalidate that triggers a refetch', async () => {
        vi.mocked(markdownApi.listMarkdownFiles)
            .mockResolvedValueOnce(['a.md'])
            .mockResolvedValueOnce(['a.md', 'b.md'])

        let result: ReturnType<typeof useMarkdownFiles>
        withQueryClient(() => {
            result = useMarkdownFiles(ref('abc123'))
        })

        await flushPromises()
        expect(result!.files.value).toEqual(['a.md'])

        await result!.invalidate()
        await flushPromises()
        expect(result!.files.value).toEqual(['a.md', 'b.md'])
    })
})
