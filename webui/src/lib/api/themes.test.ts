import { describe, it, expect, vi, beforeEach } from 'vitest'
import { apiClient } from '@/lib/api/client'
import {
    listThemes, getTheme, createTheme, updateTheme, deleteTheme, copyTheme,
    listThemeAssets, uploadThemeAsset, deleteThemeAsset, themeAssetUrl,
} from '@/lib/api/themes'
import type { CreateThemeDTO, Palette, Theme } from '@/types/theme'

vi.mock('@/lib/api/client', () => ({
    apiClient: { get: vi.fn(), post: vi.fn(), put: vi.fn(), delete: vi.fn() },
}))

const palette: Palette = {
    bg: '#ffffff', surface: '#f5f5f7', border: '#e0e0e5', text: '#333333', textMuted: '#666666',
    primary: '#0a84ff', onPrimary: '#ffffff', secondary: '#8b5cf6', success: '#16a34a',
    warning: '#d97706', danger: '#dc2626',
}

const theme = (over: Partial<Theme> = {}): Theme => ({
    id: 'a1', name: 'A', darkMode: true, background: { light: {}, dark: {} }, colors: { light: palette, dark: palette }, ...over,
})

describe('themes api', () => {
    beforeEach(() => vi.clearAllMocks())

    it('unwraps the items envelope when listing', async () => {
        vi.mocked(apiClient.get).mockResolvedValue({
            data: { items: [{ id: 'a1', name: 'A', darkMode: true, usedBy: 0, previewCss: '', colors: { light: palette, dark: palette } }] },
        })
        const got = await listThemes()
        expect(apiClient.get).toHaveBeenCalledWith('/themes')
        expect(got).toHaveLength(1)
        expect(got[0].id).toBe('a1')
    })

    it('returns an empty list when the server sends no items', async () => {
        vi.mocked(apiClient.get).mockResolvedValue({ data: {} })
        expect(await listThemes()).toEqual([])
    })

    it('gets one theme', async () => {
        vi.mocked(apiClient.get).mockResolvedValue({ data: theme() })
        const got = await getTheme('a1')
        expect(apiClient.get).toHaveBeenCalledWith('/themes/a1')
        expect(got.name).toBe('A')
    })

    it('creates and updates', async () => {
        const dto: CreateThemeDTO = { name: 'A', darkMode: true, background: { light: {}, dark: {} }, colors: { light: palette, dark: palette } }
        vi.mocked(apiClient.post).mockResolvedValue({ data: theme() })
        await createTheme(dto)
        expect(apiClient.post).toHaveBeenCalledWith('/themes', dto)

        const updated = theme({ name: 'B' })
        vi.mocked(apiClient.put).mockResolvedValue({ data: updated })
        await updateTheme('a1', updated)
        expect(apiClient.put).toHaveBeenCalledWith('/themes/a1', updated)
    })

    it('deletes', async () => {
        vi.mocked(apiClient.delete).mockResolvedValue({ data: null })
        await deleteTheme('a1')
        expect(apiClient.delete).toHaveBeenCalledWith('/themes/a1')
    })

    it('copies a theme and returns the new one', async () => {
        const copy = theme({ id: 'def456', name: 'Copy of A' })
        vi.mocked(apiClient.post).mockResolvedValue({ data: copy })
        const got = await copyTheme('abc123')
        expect(apiClient.post).toHaveBeenCalledWith('/themes/abc123/copy')
        expect(got).toEqual(copy)
    })

    it('lists, uploads and deletes assets with encoded names', async () => {
        vi.mocked(apiClient.get).mockResolvedValue({ data: { items: ['a.png'] } })
        expect(await listThemeAssets('a1')).toEqual(['a.png'])
        expect(apiClient.get).toHaveBeenCalledWith('/themes/a1/assets')

        vi.mocked(apiClient.post).mockResolvedValue({ data: null })
        const bytes = new ArrayBuffer(4)
        await uploadThemeAsset('a1', 'my pic.png', bytes)
        expect(apiClient.post).toHaveBeenCalledWith(
            '/themes/a1/assets/my%20pic.png',
            bytes,
            { headers: { 'Content-Type': 'application/octet-stream' } },
        )

        vi.mocked(apiClient.delete).mockResolvedValue({ data: null })
        await deleteThemeAsset('a1', 'my pic.png')
        expect(apiClient.delete).toHaveBeenCalledWith('/themes/a1/assets/my%20pic.png')
    })

    it('builds an absolute asset url', () => {
        expect(themeAssetUrl('a1', 'my pic.png')).toBe('/api/v0/themes/a1/assets/my%20pic.png')
    })

    it('preserves slashes in nested asset paths while encoding each segment', async () => {
        vi.mocked(apiClient.post).mockResolvedValue({ data: null })
        const bytes = new ArrayBuffer(4)
        await uploadThemeAsset('a1', 'folder/my pic.png', bytes)
        expect(apiClient.post).toHaveBeenCalledWith(
            '/themes/a1/assets/folder/my%20pic.png',
            bytes,
            { headers: { 'Content-Type': 'application/octet-stream' } },
        )

        vi.mocked(apiClient.delete).mockResolvedValue({ data: null })
        await deleteThemeAsset('a1', 'folder/my pic.png')
        expect(apiClient.delete).toHaveBeenCalledWith('/themes/a1/assets/folder/my%20pic.png')

        expect(themeAssetUrl('a1', 'folder/my pic.png')).toBe('/api/v0/themes/a1/assets/folder/my%20pic.png')
    })
})
