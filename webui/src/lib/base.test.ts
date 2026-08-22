import { describe, it, expect, afterEach } from 'vitest'
import { appBase, withBase } from '@/lib/base'

type W = { __DASHI_BASE__?: string }

afterEach(() => {
  delete (window as unknown as W).__DASHI_BASE__
})

describe('appBase', () => {
  it('defaults to "/" when nothing is injected', () => {
    expect(appBase()).toBe('/')
  })
  it('returns the injected base', () => {
    ;(window as unknown as W).__DASHI_BASE__ = '/api/hassio_ingress/tok/'
    expect(appBase()).toBe('/api/hassio_ingress/tok/')
  })
  it('treats an empty string as "/"', () => {
    ;(window as unknown as W).__DASHI_BASE__ = ''
    expect(appBase()).toBe('/')
  })
})

describe('withBase', () => {
  it('joins against "/" without doubling the slash', () => {
    expect(withBase('/api/v0')).toBe('/api/v0')
  })
  it('prefixes with the ingress base', () => {
    ;(window as unknown as W).__DASHI_BASE__ = '/api/hassio_ingress/tok/'
    expect(withBase('/api/v0')).toBe('/api/hassio_ingress/tok/api/v0')
  })
})
