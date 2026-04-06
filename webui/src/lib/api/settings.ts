import { apiClient } from '@/lib/api/client'

export interface AppSettings {
    readOnly: boolean
}

export const getSettings = async (): Promise<AppSettings> => {
    const { data } = await apiClient.get<AppSettings>('/settings')
    return data
}
