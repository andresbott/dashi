import { apiClient } from '@/lib/api/client'

export type DataKind = 'notes' | 'images' | 'backgrounds'

export interface DataItem {
    name: string
    size: number
    modTime: string
}

export const listData = async (kind: DataKind): Promise<DataItem[]> => {
    const { data } = await apiClient.get<DataItem[]>(`/data/${kind}`)
    return data ?? []
}

export const uploadData = async (kind: DataKind, name: string, bytes: ArrayBuffer): Promise<void> => {
    await apiClient.post(
        `/data/${kind}/${encodeURIComponent(name)}`,
        bytes,
        { headers: { 'Content-Type': 'application/octet-stream' } },
    )
}

export const deleteData = async (kind: DataKind, name: string): Promise<void> => {
    await apiClient.delete(`/data/${kind}/${encodeURIComponent(name)}`)
}

export const dataUrl = (kind: DataKind, name: string): string => {
    return `/api/v0/data/${kind}/${encodeURIComponent(name)}`
}
