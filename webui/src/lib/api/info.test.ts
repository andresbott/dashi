import { describe, it, expect, vi, beforeEach } from 'vitest'
import { getServerInfo } from './info'
import { apiClient } from './client'

vi.mock('./client', () => ({
    apiClient: {
        get: vi.fn(),
    },
}))

describe('info API', () => {
    beforeEach(() => {
        vi.clearAllMocks()
    })

    it('returns server info', async () => {
        const info = { viewer: { enabled: true, url: 'http://localhost:8087' } }
        vi.mocked(apiClient.get).mockResolvedValue({ data: info })
        const result = await getServerInfo()
        expect(result).toEqual(info)
        expect(apiClient.get).toHaveBeenCalledWith('/info')
    })
})
