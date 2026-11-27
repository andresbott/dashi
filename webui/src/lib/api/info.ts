import { apiClient } from '@/lib/api/client'

// EndpointInfo says whether a server is enabled and where a browser reaches it;
// url is '' when the backend has no public URL for it.
export interface EndpointInfo {
    enabled: boolean
    url: string
}

export interface ServerInfo {
    viewer: EndpointInfo
    // Optional: backends older than the edit-button feature do not send it.
    editor?: EndpointInfo
}

export const getServerInfo = async (): Promise<ServerInfo> => {
    const { data } = await apiClient.get<ServerInfo>('/info')
    return data
}
