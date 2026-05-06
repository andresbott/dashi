import { apiClient } from '@/lib/api/client'

export interface DataItem {
    name: string
    size: number
    modTime: string
}

export const listDataImages = async (): Promise<DataItem[]> => {
    const { data } = await apiClient.get<DataItem[]>('/data/images')
    return data ?? []
}
