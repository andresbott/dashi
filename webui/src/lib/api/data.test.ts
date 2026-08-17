import { describe, it, expect, vi, beforeEach } from 'vitest'
import { listData, uploadData, deleteData, dataUrl } from './data'
import { apiClient } from './client'

vi.mock('./client', () => ({
    apiClient: {
        get: vi.fn(),
        post: vi.fn(),
        put: vi.fn(),
        delete: vi.fn(),
    },
}))

describe('data API', () => {
    beforeEach(() => {
        vi.clearAllMocks()
    })

    describe('listData', () => {
        it.each(['notes', 'images', 'backgrounds'] as const)('lists %s', async (kind) => {
            vi.mocked(apiClient.get).mockResolvedValue({
                data: [{ name: 'a', size: 1, modTime: '2026-01-01T00:00:00Z' }],
            })
            const result = await listData(kind)
            expect(apiClient.get).toHaveBeenCalledWith(`/data/${kind}`)
            expect(result).toEqual([{ name: 'a', size: 1, modTime: '2026-01-01T00:00:00Z' }])
        })

        it('returns empty array when data is null', async () => {
            vi.mocked(apiClient.get).mockResolvedValue({ data: null })
            expect(await listData('images')).toEqual([])
        })
    })

    describe('uploadData', () => {
        it('posts octet-stream body', async () => {
            vi.mocked(apiClient.post).mockResolvedValue({ data: undefined })
            const bytes = new ArrayBuffer(4)
            await uploadData('backgrounds', 'bg.png', bytes)
            const call = vi.mocked(apiClient.post).mock.calls[0]
            expect(call[0]).toBe('/data/backgrounds/bg.png')
            expect(call[1]).toBe(bytes)
            expect(call[2]).toEqual({ headers: { 'Content-Type': 'application/octet-stream' } })
        })

        it('URL-encodes the name', async () => {
            vi.mocked(apiClient.post).mockResolvedValue({ data: undefined })
            await uploadData('images', 'a b.png', new ArrayBuffer(0))
            expect(vi.mocked(apiClient.post).mock.calls[0][0]).toBe('/data/images/a%20b.png')
        })
    })

    describe('deleteData', () => {
        it('calls DELETE on the kind + name', async () => {
            vi.mocked(apiClient.delete).mockResolvedValue({ data: undefined })
            await deleteData('images', 'x.png')
            expect(apiClient.delete).toHaveBeenCalledWith('/data/images/x.png')
        })
    })

    describe('dataUrl', () => {
        it('builds an api v0 url', () => {
            expect(dataUrl('backgrounds', 'bg.jpg')).toBe('/api/v0/data/backgrounds/bg.jpg')
        })

        it('URL-encodes the name', () => {
            expect(dataUrl('images', 'a b.png')).toBe('/api/v0/data/images/a%20b.png')
        })
    })
})
