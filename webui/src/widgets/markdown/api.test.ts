import { describe, it, expect, vi, beforeEach } from 'vitest'
import { listMarkdownFiles } from './api'
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
        it('returns list of md file names', async () => {
            vi.mocked(apiClient.get).mockResolvedValue({
                data: { files: ['notes.md', 'welcome.md'] },
            })
            const result = await listMarkdownFiles('abc123')
            expect(result).toEqual(['notes.md', 'welcome.md'])
            expect(apiClient.get).toHaveBeenCalledWith('/dashboards/abc123/markdown')
        })

        it('returns empty array when files is null', async () => {
            vi.mocked(apiClient.get).mockResolvedValue({ data: { files: null } })
            const result = await listMarkdownFiles('abc123')
            expect(result).toEqual([])
        })

        it('returns empty array when files is missing', async () => {
            vi.mocked(apiClient.get).mockResolvedValue({ data: {} })
            const result = await listMarkdownFiles('abc123')
            expect(result).toEqual([])
        })
    })
})
