import { describe, it, expect, vi, beforeEach } from 'vitest'
import {
    listBookmarkFiles,
    getBookmarkFile,
    saveBookmarkFile,
    deleteBookmarkFile,
} from './bookmarkFiles'
import { apiClient } from '@/lib/api/client'

vi.mock('@/lib/api/client', () => ({
    apiClient: { get: vi.fn(), put: vi.fn(), delete: vi.fn() },
}))

describe('bookmark files API', () => {
    beforeEach(() => vi.clearAllMocks())

    it('lists file names', async () => {
        vi.mocked(apiClient.get).mockResolvedValue({ data: [{ name: 'links.yaml', size: 1, modTime: '' }] })
        expect(await listBookmarkFiles()).toEqual(['links.yaml'])
        expect(apiClient.get).toHaveBeenCalledWith('/data/bookmarks')

        vi.mocked(apiClient.get).mockResolvedValue({ data: null })
        expect(await listBookmarkFiles()).toEqual([])
    })

    it('gets the parsed bookmarks of a file', async () => {
        const bookmarks = [{ url: 'https://a.example', icon: '', title: 'A', subtitle: '' }]
        vi.mocked(apiClient.get).mockResolvedValue({ data: { bookmarks } })
        expect(await getBookmarkFile('my links.yaml')).toEqual(bookmarks)
        expect(apiClient.get).toHaveBeenCalledWith('/data/bookmarks/my%20links.yaml')

        vi.mocked(apiClient.get).mockResolvedValue({ data: {} })
        expect(await getBookmarkFile('x.yaml')).toEqual([])
    })

    it('saves and deletes', async () => {
        const bookmarks = [{ url: '/a', icon: '', title: 'A', subtitle: '' }]
        vi.mocked(apiClient.put).mockResolvedValue({})
        await saveBookmarkFile('x.yaml', bookmarks)
        expect(apiClient.put).toHaveBeenCalledWith('/data/bookmarks/x.yaml', { bookmarks })

        vi.mocked(apiClient.delete).mockResolvedValue({})
        await deleteBookmarkFile('x.yaml')
        expect(apiClient.delete).toHaveBeenCalledWith('/data/bookmarks/x.yaml')
    })
})
