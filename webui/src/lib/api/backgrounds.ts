import { apiClient } from '@/lib/api/client'
import type { Background, BackgroundMeta, CreateBackgroundDTO } from '@/types/background'
import { withBase } from '@/lib/base'

export const listBackgrounds = async (): Promise<BackgroundMeta[]> => {
    const { data } = await apiClient.get<{ items?: BackgroundMeta[] }>('/backgrounds')
    return data.items ?? []
}

export const getBackground = async (id: string): Promise<Background> => {
    const { data } = await apiClient.get<Background>(`/backgrounds/${encodeURIComponent(id)}`)
    return data
}

export const createBackground = async (payload: CreateBackgroundDTO): Promise<Background> => {
    const { data } = await apiClient.post<Background>('/backgrounds', payload)
    return data
}

export const updateBackground = async (id: string, payload: Background): Promise<Background> => {
    const { data } = await apiClient.put<Background>(`/backgrounds/${encodeURIComponent(id)}`, payload)
    return data
}

export const deleteBackground = async (id: string): Promise<void> => {
    await apiClient.delete(`/backgrounds/${encodeURIComponent(id)}`)
}

export const listBackgroundAssets = async (id: string): Promise<string[]> => {
    const { data } = await apiClient.get<{ items?: string[] }>(`/backgrounds/${encodeURIComponent(id)}/assets`)
    return data.items ?? []
}

// assetPath may contain slashes; each segment is encoded separately so the
// path structure survives.
const encodeAssetPath = (assetPath: string): string =>
    assetPath.split('/').map(encodeURIComponent).join('/')

export const uploadBackgroundAsset = async (id: string, assetPath: string, bytes: ArrayBuffer): Promise<void> => {
    await apiClient.post(
        `/backgrounds/${encodeURIComponent(id)}/assets/${encodeAssetPath(assetPath)}`,
        bytes,
        { headers: { 'Content-Type': 'application/octet-stream' } },
    )
}

export const deleteBackgroundAsset = async (id: string, assetPath: string): Promise<void> => {
    await apiClient.delete(`/backgrounds/${encodeURIComponent(id)}/assets/${encodeAssetPath(assetPath)}`)
}

// backgroundAssetUrl is the browser-facing URL of a background's own image.
export const backgroundAssetUrl = (id: string, assetPath: string): string =>
    withBase(`/api/v0/backgrounds/${encodeURIComponent(id)}/assets/${encodeAssetPath(assetPath)}`)
