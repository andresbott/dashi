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

// setDefaultDashboard marks a dashboard as the default one. There is no
// dedicated endpoint, so the full dashboard is read and written back; the
// backend clears the flag on every other dashboard.
export const setDefaultDashboard = async (id: string): Promise<Dashboard> => {
    const dashboard = await getDashboard(id)
    return updateDashboard(id, { ...dashboard, default: true })
}

export interface BackgroundOption {
    name: string
    value: string
}

export interface BackgroundsResponse {
    theme: BackgroundOption[]
    dashboard: BackgroundOption[]
    shared: BackgroundOption[]
}

export const downloadDashboard = async (id: string): Promise<void> => {
    const response = await apiClient.get(`${DASHBOARD_PATH}/${id}/download`, {
        responseType: 'blob',
    })
    const disposition = response.headers['content-disposition'] || ''
    const match = disposition.match(/filename="(.+)"/)
    const filename = match ? match[1] : `dashboard-${id}.zip`

    const url = URL.createObjectURL(response.data)
    const a = document.createElement('a')
    a.href = url
    a.download = filename
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
}

export const getDashboardAssets = async (dashboardId: string): Promise<string[]> => {
    const { data } = await apiClient.get<{ items: string[] }>(`${DASHBOARD_PATH}/${dashboardId}/assets`)
    return data.items ?? []
}

export const uploadDashboardZip = async (data: ArrayBuffer): Promise<Dashboard> => {
    const { data: created } = await apiClient.post<Dashboard>(`${DASHBOARD_PATH}/upload`, data, {
        headers: { 'Content-Type': 'application/zip' },
    })
    return created
}

export const getBackgrounds = async (dashboardId: string): Promise<BackgroundsResponse> => {
    const { data } = await apiClient.get<BackgroundsResponse>('/backgrounds', {
        params: { dashboard: dashboardId },
    })
    return data
}

export interface DashboardAuthResponse {
    enabled: boolean
    username?: string
}

export const getDashboardAuth = async (id: string): Promise<DashboardAuthResponse> => {
    const { data } = await apiClient.get<DashboardAuthResponse>(`${DASHBOARD_PATH}/${id}/auth`)
    return data
}

export const setDashboardAuth = async (id: string, username: string, password: string): Promise<DashboardAuthResponse> => {
    const { data } = await apiClient.put<DashboardAuthResponse>(`${DASHBOARD_PATH}/${id}/auth`, { username, password })
    return data
}

export const deleteDashboardAuth = async (id: string): Promise<DashboardAuthResponse> => {
    const { data } = await apiClient.delete<DashboardAuthResponse>(`${DASHBOARD_PATH}/${id}/auth`)
    return data
}
