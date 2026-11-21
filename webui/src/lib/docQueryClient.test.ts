import { describe, it, expect, vi } from 'vitest'
import { createDocQueryClient } from './docQueryClient'

describe('createDocQueryClient', () => {
    it('answers a mocked key prefix from the fixture, never the real queryFn', async () => {
        const real = vi.fn(() => Promise.resolve('real'))
        const client = createDocQueryClient({ weather: (key) => ({ lat: key[1] }) })

        const data = await client.fetchQuery({ queryKey: ['weather', 47, 8], queryFn: real })

        expect(data).toEqual({ lat: 47 })
        expect(real).not.toHaveBeenCalled()
    })

    it('disables refetching for mocked queries', () => {
        const client = createDocQueryClient({ sysinfo: () => ({}) })
        const options = client.defaultQueryOptions({ queryKey: ['sysinfo'], refetchInterval: 1000, staleTime: 0 })
        expect(options.refetchInterval).toBe(false)
        expect(options.staleTime).toBe(Infinity)
    })

    it('runs the real queryFn for keys without a mock', async () => {
        const real = vi.fn(() => Promise.resolve(['default']))
        const client = createDocQueryClient({ weather: () => ({}) })

        expect(await client.fetchQuery({ queryKey: ['widget-themes'], queryFn: real })).toEqual(['default'])
        expect(real).toHaveBeenCalledOnce()
    })

    it('fails the query when the fixture has no data for the key', async () => {
        const client = createDocQueryClient({ 'data-notes-html': () => undefined })
        await expect(
            client.fetchQuery({ queryKey: ['data-notes-html', 'missing.md'], queryFn: () => 'real' }),
        ).rejects.toThrow(/no doc fixture/)
    })
})
