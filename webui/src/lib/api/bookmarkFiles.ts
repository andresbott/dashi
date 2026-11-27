import { apiClient } from '@/lib/api/client'
import type { DataItem } from './data'

/** One bookmark of a content file: same fields as a dashboard bookmark. */
export interface BookmarkFileItem {
    url: string
    icon: string
    title: string
    subtitle: string
}

const path = (name: string) => `/data/bookmarks/${encodeURIComponent(name)}`

export const listBookmarkFiles = async (): Promise<string[]> => {
    const { data } = await apiClient.get<DataItem[]>('/data/bookmarks')
    return (data ?? []).map(item => item.name)
}

/** The parsed bookmarks of a file, as the server's widget renders them. */
export const getBookmarkFile = async (name: string): Promise<BookmarkFileItem[]> => {
    const { data } = await apiClient.get<{ bookmarks?: BookmarkFileItem[] }>(path(name))
    return data?.bookmarks ?? []
}

/** Saves the bookmarks of a file; the server writes them as YAML. */
export const saveBookmarkFile = async (name: string, bookmarks: BookmarkFileItem[]): Promise<void> => {
    await apiClient.put(path(name), { bookmarks })
}

export const deleteBookmarkFile = async (name: string): Promise<void> => {
    await apiClient.delete(path(name))
}
