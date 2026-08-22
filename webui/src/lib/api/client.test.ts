import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'

type W = { __DASHI_BASE__?: string }

describe('apiClient baseURL', () => {
  beforeEach(() => vi.resetModules())
  afterEach(() => delete (window as unknown as W).__DASHI_BASE__)

  it('defaults to /api/v0', async () => {
    const { apiClient } = await import('@/lib/api/client')
    expect(apiClient.defaults.baseURL).toBe('/api/v0')
  })

  it('is prefixed with the ingress base', async () => {
    ;(window as unknown as W).__DASHI_BASE__ = '/api/hassio_ingress/tok/'
    const { apiClient } = await import('@/lib/api/client')
    expect(apiClient.defaults.baseURL).toBe('/api/hassio_ingress/tok/api/v0')
  })
})
