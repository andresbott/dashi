import { describe, it, expect, vi, beforeEach } from 'vitest'
import { listMarkdownFiles, getMarkdownHtml, getMarkdownRaw, saveMarkdown, deleteMarkdown } from './api'
import { apiClient } from '@/lib/api/client'

vi.mock('@/lib/api/client', () => ({
    apiClient: {
        get: vi.fn(),
        post: vi.fn(),
        put: vi.fn(),
        delete: vi.fn(),
    },
}))

describe('markdown API', () => {
    beforeEach(() => {
        vi.clearAllMocks()
    })

    describe('listMarkdownFiles', () => {
        it('returns list of note names', async () => {
            vi.mocked(apiClient.get).mockResolvedValue({
                data: [
                    { name: 'notes.md', size: 10, modTime: '2026-01-01T00:00:00Z' },
                    { name: 'welcome.md', size: 20, modTime: '2026-01-02T00:00:00Z' },
                ],
            })
            const result = await listMarkdownFiles()
            expect(result).toEqual(['notes.md', 'welcome.md'])
            expect(apiClient.get).toHaveBeenCalledWith('/data/notes')
        })

        it('returns empty array when data is null', async () => {
            vi.mocked(apiClient.get).mockResolvedValue({ data: null })
            const result = await listMarkdownFiles()
            expect(result).toEqual([])
        })

        it('returns empty array when data is empty', async () => {
            vi.mocked(apiClient.get).mockResolvedValue({ data: [] })
            const result = await listMarkdownFiles()
            expect(result).toEqual([])
        })
    })

    describe('getMarkdownHtml', () => {
        it('fetches rendered HTML', async () => {
            vi.mocked(apiClient.get).mockResolvedValue({ data: { html: '<h1>Hi</h1>' } })
            const result = await getMarkdownHtml('a.md')
            expect(result).toBe('<h1>Hi</h1>')
            expect(apiClient.get).toHaveBeenCalledWith('/data/notes/a.md')
        })
    })

    describe('getMarkdownRaw', () => {
        it('fetches raw markdown', async () => {
            vi.mocked(apiClient.get).mockResolvedValue({ data: '# Hi' })
            const result = await getMarkdownRaw('a.md')
            expect(result).toBe('# Hi')
            const call = vi.mocked(apiClient.get).mock.calls[0]
            expect(call[0]).toBe('/data/notes/a.md/raw')
        })
    })

    describe('saveMarkdown', () => {
        it('posts encoded content', async () => {
            vi.mocked(apiClient.post).mockResolvedValue({ data: undefined })
            await saveMarkdown('a.md', '# Hi')
            const call = vi.mocked(apiClient.post).mock.calls[0]
            expect(call[0]).toBe('/data/notes/a.md')
            expect(call[2]).toEqual({ headers: { 'Content-Type': 'application/octet-stream' } })
        })
    })

    describe('deleteMarkdown', () => {
        it('calls DELETE on the note path', async () => {
            vi.mocked(apiClient.delete).mockResolvedValue({ data: undefined })
            await deleteMarkdown('a.md')
            expect(apiClient.delete).toHaveBeenCalledWith('/data/notes/a.md')
        })

        it('URL-encodes the name', async () => {
            vi.mocked(apiClient.delete).mockResolvedValue({ data: undefined })
            await deleteMarkdown('a b.md')
            expect(apiClient.delete).toHaveBeenCalledWith('/data/notes/a%20b.md')
        })
    })
})
