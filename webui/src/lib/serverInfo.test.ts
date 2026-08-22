import { describe, it, expect, vi, beforeEach } from 'vitest'
import { loadServerInfo, publicViewerUrl, dashboardViewUrl, resetServerInfo } from './serverInfo'
import { getServerInfo } from '@/lib/api/info'

vi.mock('@/lib/api/info', () => ({
    getServerInfo: vi.fn(),
}))

describe('serverInfo', () => {
    beforeEach(() => {
        vi.clearAllMocks()
        resetServerInfo()
    })

    it('links to the public viewer once loaded', async () => {
        vi.mocked(getServerInfo).mockResolvedValue({ viewer: { enabled: true, url: 'http://localhost:8087' } })
        await loadServerInfo()
        expect(publicViewerUrl()).toBe('http://localhost:8087')
        expect(dashboardViewUrl('home')).toBe('http://localhost:8087/home')
    })

    it('falls back to same-origin paths when the viewer is disabled', async () => {
        vi.mocked(getServerInfo).mockResolvedValue({ viewer: { enabled: false, url: '' } })
        await loadServerInfo()
        expect(publicViewerUrl()).toBe('')
        expect(dashboardViewUrl('home')).toBe('/home')
    })

    it('falls back to same-origin paths when the request fails', async () => {
        vi.mocked(getServerInfo).mockRejectedValue(new Error('boom'))
        await loadServerInfo()
        expect(dashboardViewUrl('home')).toBe('/home')
    })

    it('encodes the dashboard id', async () => {
        vi.mocked(getServerInfo).mockResolvedValue({ viewer: { enabled: true, url: 'http://localhost:8087' } })
        await loadServerInfo()
        expect(dashboardViewUrl('a b')).toBe('http://localhost:8087/a%20b')
    })
})
