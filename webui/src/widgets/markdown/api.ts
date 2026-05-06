import { apiClient } from '@/lib/api/client'

export interface MarkdownResponse {
    html: string
}

export interface DataItem {
    name: string
    size: number
    modTime: string
}

export const getMarkdownHtml = async (filename: string): Promise<string> => {
    const { data } = await apiClient.get<MarkdownResponse>(
        `/data/notes/${encodeURIComponent(filename)}`
    )
    return data.html
}

export const getMarkdownRaw = async (filename: string): Promise<string> => {
    const { data } = await apiClient.get<string>(
        `/data/notes/${encodeURIComponent(filename)}/raw`,
        { responseType: 'text', transformResponse: [(d: string) => d] }
    )
    return data
}

export const saveMarkdown = async (filename: string, content: string): Promise<void> => {
    await apiClient.post(
        `/data/notes/${encodeURIComponent(filename)}`,
        new TextEncoder().encode(content),
        { headers: { 'Content-Type': 'application/octet-stream' } }
    )
}

export const listMarkdownFiles = async (): Promise<string[]> => {
    const { data } = await apiClient.get<DataItem[]>('/data/notes')
    return (data ?? []).map(item => item.name)
}
