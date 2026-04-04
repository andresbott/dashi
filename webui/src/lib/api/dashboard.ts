import { apiClient } from '@/lib/api/client'
import type { Dashboard, DashboardMeta, CreateDashboardDTO } from '@/types/dashboard'

const DASHBOARD_PATH = '/dashboards'

export const getDashboards = async (): Promise<DashboardMeta[]> => {
    const { data } = await apiClient.get<{ items: DashboardMeta[] }>(DASHBOARD_PATH)
    return data.items ?? []
}

export const getDashboard = async (id: string): Promise<Dashboard> => {
    const { data } = await apiClient.get<Dashboard>(`${DASHBOARD_PATH}/${id}`)
    return data
}

export const createDashboard = async (payload: CreateDashboardDTO): Promise<Dashboard> => {
    const { data } = await apiClient.post<Dashboard>(DASHBOARD_PATH, payload)
    return data
}

export const updateDashboard = async (id: string, payload: Dashboard): Promise<Dashboard> => {
    const { data } = await apiClient.put<Dashboard>(`${DASHBOARD_PATH}/${id}`, payload)
    return data
}

export const deleteDashboard = async (id: string): Promise<void> => {
    await apiClient.delete(`${DASHBOARD_PATH}/${id}`)
}

export const deletePreviews = async (): Promise<{ deleted: number }> => {
    const { data } = await apiClient.delete<{ deleted: number }>(`${DASHBOARD_PATH}/previews`)
    return data
}

export const getCustomCSS = async (id: string): Promise<string> => {
    try {
        const { data } = await apiClient.get<string>(`${DASHBOARD_PATH}/${id}/assets/custom.css`, {
            responseType: 'text',
            transformResponse: [(d) => d],
        })
        return data
    } catch {
        return ''
    }
}

export const saveCustomCSS = async (id: string, css: string): Promise<void> => {
    await apiClient.post(`${DASHBOARD_PATH}/${id}/assets/custom.css`, css, {
        headers: { 'Content-Type': 'text/css' },
    })
}
