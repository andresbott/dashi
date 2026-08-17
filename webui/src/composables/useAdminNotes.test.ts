import { describe, it, expect, vi } from 'vitest'
import { QueryClient, VueQueryPlugin } from '@tanstack/vue-query'
import { mount, flushPromises } from '@vue/test-utils'
import { defineComponent } from 'vue'
import { useAdminNotes } from './useAdminNotes'
import * as markdownApi from '@/widgets/markdown/api'

vi.mock('@/widgets/markdown/api', () => ({
    listMarkdownFiles: vi.fn(),
    deleteMarkdown: vi.fn(),
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

describe('useAdminNotes', () => {
    it('fetches the list of notes', async () => {
        vi.mocked(markdownApi.listMarkdownFiles).mockResolvedValue(['a.md', 'b.md'])
        let result: ReturnType<typeof useAdminNotes>
        withQueryClient(() => {
            result = useAdminNotes()
        })
        await flushPromises()
        expect(result!.notes.value).toEqual(['a.md', 'b.md'])
    })

    it('deleteNote calls deleteMarkdown', async () => {
        vi.mocked(markdownApi.listMarkdownFiles).mockResolvedValue([])
        vi.mocked(markdownApi.deleteMarkdown).mockResolvedValue(undefined)
        let result: ReturnType<typeof useAdminNotes>
        withQueryClient(() => {
            result = useAdminNotes()
        })
        await flushPromises()
        await result!.deleteNote('a.md')
        expect(markdownApi.deleteMarkdown).toHaveBeenCalledWith('a.md')
    })
})
