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

describe('apiErrorMessage', () => {
  const failure = (data: unknown) => ({ response: { data } })

  it.each([
    ['a plain error body', failure({ error: 'bad thing' }), 'bad thing'],
    ['a middleware-wrapped body', failure({ error: '{"error":"bad thing"}\n' }), 'bad thing'],
    ['a wrapped body without error', failure({ error: '{"other":1}' }), '{"other":1}'],
    ['no response', new Error('network'), undefined],
    ['an empty error', failure({ error: '' }), undefined],
    ['a non-string error', failure({ error: 5 }), undefined],
    ['nothing', undefined, undefined],
  ])('handles %s', async (_name, err, want) => {
    const { apiErrorMessage } = await import('@/lib/api/client')
    expect(apiErrorMessage(err)).toBe(want)
  })
})
