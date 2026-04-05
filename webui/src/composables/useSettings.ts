import { useQuery } from '@tanstack/vue-query'
import { apiClient } from '@/lib/api/client'

export interface AppSettings {
    readOnly: boolean
}

async function getSettings(): Promise<AppSettings> {
    const { data } = await apiClient.get<AppSettings>('/settings')
    return data
}

export function useSettings() {
    return useQuery({
        queryKey: ['settings'],
        queryFn: getSettings,
        staleTime: 5 * 60 * 1000,
    })
}
