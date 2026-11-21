import { apiClient } from '@/lib/api/client'
import type { Theme, ThemeMeta, CreateThemeDTO } from '@/types/theme'
import { withBase } from '@/lib/base'

export const listThemes = async (): Promise<ThemeMeta[]> => {
    const { data } = await apiClient.get<{ items?: ThemeMeta[] }>('/themes')
    return data.items ?? []
}

export const getTheme = async (id: string): Promise<Theme> => {
    const { data } = await apiClient.get<Theme>(`/themes/${encodeURIComponent(id)}`)
    return data
}

export const createTheme = async (payload: CreateThemeDTO): Promise<Theme> => {
    const { data } = await apiClient.post<Theme>('/themes', payload)
    return data
}

export const updateTheme = async (id: string, payload: Theme): Promise<Theme> => {
    const { data } = await apiClient.put<Theme>(`/themes/${encodeURIComponent(id)}`, payload)
    return data
}

export const deleteTheme = async (id: string): Promise<void> => {
    await apiClient.delete(`/themes/${encodeURIComponent(id)}`)
}

// copyTheme duplicates a theme, uploaded assets included, under a new ID. The
// server names the copy "Copy of <name>".
export const copyTheme = async (id: string): Promise<Theme> => {
    const { data } = await apiClient.post<Theme>(`/themes/${encodeURIComponent(id)}/copy`)
    return data
}

export const listThemeAssets = async (id: string): Promise<string[]> => {
    const { data } = await apiClient.get<{ items?: string[] }>(`/themes/${encodeURIComponent(id)}/assets`)
    return data.items ?? []
}

// assetPath may contain slashes; each segment is encoded separately so the
// path structure survives.
const encodeAssetPath = (assetPath: string): string =>
    assetPath.split('/').map(encodeURIComponent).join('/')

export const uploadThemeAsset = async (id: string, assetPath: string, bytes: ArrayBuffer): Promise<void> => {
    await apiClient.post(
        `/themes/${encodeURIComponent(id)}/assets/${encodeAssetPath(assetPath)}`,
        bytes,
        { headers: { 'Content-Type': 'application/octet-stream' } },
    )
}

export const deleteThemeAsset = async (id: string, assetPath: string): Promise<void> => {
    await apiClient.delete(`/themes/${encodeURIComponent(id)}/assets/${encodeAssetPath(assetPath)}`)
}

// themeAssetUrl is the browser-facing URL of a theme's own image.
export const themeAssetUrl = (id: string, assetPath: string): string =>
    withBase(`/api/v0/themes/${encodeURIComponent(id)}/assets/${encodeAssetPath(assetPath)}`)
