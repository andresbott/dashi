import { describe, it, expect, vi, beforeEach } from 'vitest'
import { apiClient } from '@/lib/api/client'
import {
    getWidgetThemes, getFontUrl, getIconUrl, uploadWidgetTheme, deleteWidgetTheme, widgetThemeDownloadUrl,
} from '@/lib/api/widgetThemes'

vi.mock('@/lib/api/client', () => ({
    apiClient: { get: vi.fn(), post: vi.fn(), put: vi.fn(), delete: vi.fn() },
}))

describe('widget themes api', () => {
    beforeEach(() => vi.clearAllMocks())

    it('lists widget themes as a plain array', async () => {
        vi.mocked(apiClient.get).mockResolvedValue({
            data: [{ name: 'default', description: '', fonts: [{ name: 'Inter' }], hasIcons: true, builtin: true }],
        })
        const got = await getWidgetThemes()
        expect(apiClient.get).toHaveBeenCalledWith('/widget-themes')
        expect(got[0].fonts[0].name).toBe('Inter')
    })

    it('encodes the theme and font name in font urls', () => {
        expect(getFontUrl('my theme', 'Inter')).toMatch(/\/api\/v0\/widget-themes\/my%20theme\/fonts\/Inter$/)
    })

    it('builds icon urls', () => {
        expect(getIconUrl('default', 'sunny')).toBe('/api/v0/widget-themes/default/icons/sunny')
    })

    it('uploads a zip and deletes by name', async () => {
        vi.mocked(apiClient.post).mockResolvedValue({ data: { name: 'x' } })
        const bytes = new ArrayBuffer(4)
        await uploadWidgetTheme(bytes)
        expect(apiClient.post).toHaveBeenCalledWith(
            '/widget-themes/upload',
            bytes,
            { headers: { 'Content-Type': 'application/zip' } },
        )

        vi.mocked(apiClient.delete).mockResolvedValue({ data: null })
        await deleteWidgetTheme('my theme')
        expect(apiClient.delete).toHaveBeenCalledWith('/widget-themes/my%20theme')
    })

    it('builds the download url', () => {
        expect(widgetThemeDownloadUrl('my theme')).toBe('/api/v0/widget-themes/my%20theme/download')
    })
})
