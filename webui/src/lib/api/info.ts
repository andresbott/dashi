import { apiClient } from '@/lib/api/client'

export interface ViewerInfo {
    enabled: boolean
    url: string
}

export interface ServerInfo {
    viewer: ViewerInfo
}

export const getServerInfo = async (): Promise<ServerInfo> => {
    const { data } = await apiClient.get<ServerInfo>('/info')
    return data
}
