import { apiClient } from '@/lib/api/client'

export interface MarkdownResponse {
    html: string
}

export const getMarkdownHtml = async (dashboardId: string, filename: string): Promise<string> => {
    const { data } = await apiClient.get<MarkdownResponse>(
        `/dashboards/${dashboardId}/markdown/${encodeURIComponent(filename)}`
    )
    return data.html
}

export const getMarkdownRaw = async (dashboardId: string, filename: string): Promise<string> => {
    const { data } = await apiClient.get<string>(
        `/dashboards/${dashboardId}/assets/md/${encodeURIComponent(filename)}`,
        { responseType: 'text', transformResponse: [(d: string) => d] }
    )
    return data
}

export const saveMarkdown = async (dashboardId: string, filename: string, content: string): Promise<void> => {
    await apiClient.post(
        `/dashboards/${dashboardId}/assets/md/${encodeURIComponent(filename)}`,
        new TextEncoder().encode(content),
        { headers: { 'Content-Type': 'application/octet-stream' } }
    )
}

export interface MarkdownFilesResponse {
    files: string[] | null
}

export const listMarkdownFiles = async (dashboardId: string): Promise<string[]> => {
    const { data } = await apiClient.get<MarkdownFilesResponse>(
        `/dashboards/${dashboardId}/markdown`
    )
    return data.files ?? []
}
