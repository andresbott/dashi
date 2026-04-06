import { describe, it, expect, vi, beforeEach } from 'vitest'
import { getSettings } from './settings'
import { apiClient } from './client'

vi.mock('./client', () => ({
    apiClient: {
        get: vi.fn(),
    },
}))

describe('getSettings', () => {
    beforeEach(() => {
        vi.clearAllMocks()
    })

    it('returns settings from API', async () => {
        vi.mocked(apiClient.get).mockResolvedValue({ data: { readOnly: false } })
        const result = await getSettings()
        expect(result).toEqual({ readOnly: false })
        expect(apiClient.get).toHaveBeenCalledWith('/settings')
    })

    it('returns readOnly true when API says so', async () => {
        vi.mocked(apiClient.get).mockResolvedValue({ data: { readOnly: true } })
        const result = await getSettings()
        expect(result.readOnly).toBe(true)
    })

    it('propagates API errors', async () => {
        vi.mocked(apiClient.get).mockRejectedValue(new Error('Network error'))
        await expect(getSettings()).rejects.toThrow('Network error')
    })
})
