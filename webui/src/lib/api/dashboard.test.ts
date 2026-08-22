import { describe, it, expect, vi, beforeEach } from 'vitest'
import {
    getDashboards,
    getDashboard,
    createDashboard,
    updateDashboard,
    deleteDashboard,
    setDefaultDashboard,
    getDashboardAssets,
} from './dashboard'
import { apiClient } from './client'

vi.mock('./client', () => ({
    apiClient: {
        get: vi.fn(),
        post: vi.fn(),
        put: vi.fn(),
        delete: vi.fn(),
    },
}))

describe('dashboard API', () => {
    beforeEach(() => {
        vi.clearAllMocks()
    })

    describe('getDashboards', () => {
        it('returns dashboard list', async () => {
            const items = [{ id: '1', name: 'test' }]
            vi.mocked(apiClient.get).mockResolvedValue({ data: { items } })
            const result = await getDashboards()
            expect(result).toEqual(items)
            expect(apiClient.get).toHaveBeenCalledWith('/dashboards')
        })

        it('returns empty array when items is null', async () => {
            vi.mocked(apiClient.get).mockResolvedValue({ data: { items: null } })
            const result = await getDashboards()
            expect(result).toEqual([])
        })
    })

    describe('getDashboard', () => {
        it('returns a single dashboard', async () => {
            const dash = { id: '1', name: 'test' }
            vi.mocked(apiClient.get).mockResolvedValue({ data: dash })
            const result = await getDashboard('1')
            expect(result).toEqual(dash)
            expect(apiClient.get).toHaveBeenCalledWith('/dashboards/1')
        })
    })

    describe('createDashboard', () => {
        it('posts and returns created dashboard', async () => {
            const payload = { name: 'new', type: 'html' }
            const created = { id: '2', ...payload }
            vi.mocked(apiClient.post).mockResolvedValue({ data: created })
            const result = await createDashboard(payload as any)
            expect(result).toEqual(created)
            expect(apiClient.post).toHaveBeenCalledWith('/dashboards', payload)
        })
    })

    describe('updateDashboard', () => {
        it('puts and returns updated dashboard', async () => {
            const payload = { id: '1', name: 'updated' }
            vi.mocked(apiClient.put).mockResolvedValue({ data: payload })
            const result = await updateDashboard('1', payload as any)
            expect(result).toEqual(payload)
            expect(apiClient.put).toHaveBeenCalledWith('/dashboards/1', payload)
        })
    })

    describe('deleteDashboard', () => {
        it('deletes a dashboard', async () => {
            vi.mocked(apiClient.delete).mockResolvedValue({})
            await deleteDashboard('1')
            expect(apiClient.delete).toHaveBeenCalledWith('/dashboards/1')
        })
    })

    describe('setDefaultDashboard', () => {
        it('reads the dashboard and puts it back marked as default', async () => {
            const dash = { id: '1', name: 'test', default: false }
            vi.mocked(apiClient.get).mockResolvedValue({ data: dash })
            vi.mocked(apiClient.put).mockResolvedValue({ data: { ...dash, default: true } })

            const result = await setDefaultDashboard('1')

            expect(apiClient.get).toHaveBeenCalledWith('/dashboards/1')
            expect(apiClient.put).toHaveBeenCalledWith('/dashboards/1', { ...dash, default: true })
            expect(result.default).toBe(true)
        })
    })

    describe('getDashboardAssets', () => {
        it('returns asset list', async () => {
            vi.mocked(apiClient.get).mockResolvedValue({ data: { items: ['a.png', 'b.svg'] } })
            const result = await getDashboardAssets('1')
            expect(result).toEqual(['a.png', 'b.svg'])
        })

        it('returns empty array when items is null', async () => {
            vi.mocked(apiClient.get).mockResolvedValue({ data: { items: null } })
            const result = await getDashboardAssets('1')
            expect(result).toEqual([])
        })
    })
})
