import { describe, it, expect, vi, beforeEach } from 'vitest'
import { apiClient } from '@/lib/api/client'
import {
    listBackgrounds, getBackground, createBackground, updateBackground,
    deleteBackground, listBackgroundAssets, uploadBackgroundAsset,
    deleteBackgroundAsset, backgroundAssetUrl,
} from '@/lib/api/backgrounds'

vi.mock('@/lib/api/client', () => ({
    apiClient: { get: vi.fn(), post: vi.fn(), put: vi.fn(), delete: vi.fn() },
}))

describe('backgrounds api', () => {
    beforeEach(() => vi.clearAllMocks())

    it('unwraps the items envelope when listing', async () => {
        vi.mocked(apiClient.get).mockResolvedValue({ data: { items: [{ id: 'a1', name: 'A', usedBy: 0, previewCss: '' }] } })
        const got = await listBackgrounds()
        expect(apiClient.get).toHaveBeenCalledWith('/backgrounds')
        expect(got).toHaveLength(1)
        expect(got[0].id).toBe('a1')
    })

    it('returns an empty list when the server sends no items', async () => {
        vi.mocked(apiClient.get).mockResolvedValue({ data: {} })
        expect(await listBackgrounds()).toEqual([])
    })

    it('gets one background', async () => {
        vi.mocked(apiClient.get).mockResolvedValue({ data: { id: 'a1', name: 'A' } })
        const got = await getBackground('a1')
        expect(apiClient.get).toHaveBeenCalledWith('/backgrounds/a1')
        expect(got.name).toBe('A')
    })

    it('creates and updates', async () => {
        vi.mocked(apiClient.post).mockResolvedValue({ data: { id: 'a1', name: 'A' } })
        await createBackground({ name: 'A' })
        expect(apiClient.post).toHaveBeenCalledWith('/backgrounds', { name: 'A' })

        vi.mocked(apiClient.put).mockResolvedValue({ data: { id: 'a1', name: 'B' } })
        await updateBackground('a1', { id: 'a1', name: 'B' })
        expect(apiClient.put).toHaveBeenCalledWith('/backgrounds/a1', { id: 'a1', name: 'B' })
    })

    it('deletes', async () => {
        vi.mocked(apiClient.delete).mockResolvedValue({ data: null })
        await deleteBackground('a1')
        expect(apiClient.delete).toHaveBeenCalledWith('/backgrounds/a1')
    })

    it('lists, uploads and deletes assets with encoded names', async () => {
        vi.mocked(apiClient.get).mockResolvedValue({ data: { items: ['a.png'] } })
        expect(await listBackgroundAssets('a1')).toEqual(['a.png'])

        vi.mocked(apiClient.post).mockResolvedValue({ data: null })
        const bytes = new ArrayBuffer(4)
        await uploadBackgroundAsset('a1', 'my pic.png', bytes)
        expect(apiClient.post).toHaveBeenCalledWith(
            '/backgrounds/a1/assets/my%20pic.png',
            bytes,
            { headers: { 'Content-Type': 'application/octet-stream' } },
        )

        vi.mocked(apiClient.delete).mockResolvedValue({ data: null })
        await deleteBackgroundAsset('a1', 'my pic.png')
        expect(apiClient.delete).toHaveBeenCalledWith('/backgrounds/a1/assets/my%20pic.png')
    })

    it('builds an absolute asset url', () => {
        expect(backgroundAssetUrl('a1', 'my pic.png')).toBe('/api/v0/backgrounds/a1/assets/my%20pic.png')
    })

    it('preserves slashes in nested asset paths while encoding each segment', async () => {
        vi.mocked(apiClient.post).mockResolvedValue({ data: null })
        const bytes = new ArrayBuffer(4)
        await uploadBackgroundAsset('a1', 'folder/my pic.png', bytes)
        expect(apiClient.post).toHaveBeenCalledWith(
            '/backgrounds/a1/assets/folder/my%20pic.png',
            bytes,
            { headers: { 'Content-Type': 'application/octet-stream' } },
        )

        vi.mocked(apiClient.delete).mockResolvedValue({ data: null })
        await deleteBackgroundAsset('a1', 'folder/my pic.png')
        expect(apiClient.delete).toHaveBeenCalledWith('/backgrounds/a1/assets/folder/my%20pic.png')

        expect(backgroundAssetUrl('a1', 'folder/my pic.png')).toBe('/api/v0/backgrounds/a1/assets/folder/my%20pic.png')
    })
})
