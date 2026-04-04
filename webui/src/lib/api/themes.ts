import { apiClient } from '@/lib/api/client'
import type { ThemeInfo, FontIconResponse } from '@/types/theme'

const THEMES_PATH = '/themes'

export const getThemes = async (): Promise<ThemeInfo[]> => {
    const { data } = await apiClient.get<ThemeInfo[]>(THEMES_PATH)
    return data
}

export const getFontIcon = async (themeName: string, iconName: string): Promise<FontIconResponse> => {
    const { data } = await apiClient.get<FontIconResponse>(`${THEMES_PATH}/${themeName}/icons/${iconName}`)
    return data
}

export const getIconUrl = (themeName: string, iconName: string): string => {
    return `/api/v0${THEMES_PATH}/${themeName}/icons/${iconName}`
}

export const getFontUrl = (themeName: string, fontName: string): string => {
    return `/api/v0${THEMES_PATH}/${encodeURIComponent(themeName)}/fonts/${encodeURIComponent(fontName)}`
}
