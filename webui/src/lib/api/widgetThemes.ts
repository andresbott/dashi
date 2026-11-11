import { apiClient } from '@/lib/api/client'
import type { WidgetThemeInfo, FontIconResponse } from '@/types/widgetTheme'
import { withBase } from '@/lib/base'

const WIDGET_THEMES_PATH = '/widget-themes'

export const getWidgetThemes = async (): Promise<WidgetThemeInfo[]> => {
    const { data } = await apiClient.get<WidgetThemeInfo[]>(WIDGET_THEMES_PATH)
    return data
}

export const getFontIcon = async (themeName: string, iconName: string): Promise<FontIconResponse> => {
    const { data } = await apiClient.get<FontIconResponse>(
        `${WIDGET_THEMES_PATH}/${encodeURIComponent(themeName)}/icons/${encodeURIComponent(iconName)}`,
    )
    return data
}

export const getIconUrl = (themeName: string, iconName: string): string => {
    return withBase(`/api/v0${WIDGET_THEMES_PATH}/${encodeURIComponent(themeName)}/icons/${encodeURIComponent(iconName)}`)
}

export const getFontUrl = (themeName: string, fontName: string): string => {
    return withBase(`/api/v0${WIDGET_THEMES_PATH}/${encodeURIComponent(themeName)}/fonts/${encodeURIComponent(fontName)}`)
}

// ---------- admin CRUD ----------

export const uploadWidgetTheme = async (zipBytes: ArrayBuffer): Promise<WidgetThemeInfo> => {
    const { data } = await apiClient.post<WidgetThemeInfo>(
        `${WIDGET_THEMES_PATH}/upload`,
        zipBytes,
        { headers: { 'Content-Type': 'application/zip' } },
    )
    return data
}

export const deleteWidgetTheme = async (name: string): Promise<void> => {
    await apiClient.delete(`${WIDGET_THEMES_PATH}/${encodeURIComponent(name)}`)
}

export const widgetThemeDownloadUrl = (name: string): string => {
    return withBase(`/api/v0${WIDGET_THEMES_PATH}/${encodeURIComponent(name)}/download`)
}
