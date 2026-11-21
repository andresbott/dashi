import { describe, it, expect, afterEach } from 'vitest'
import { clampSide, einkPreviewUrl, DEVICE_PRESETS, MAX_PREVIEW_SIDE } from '@/lib/einkPreview'

type W = { __DASHI_BASE__?: string }

afterEach(() => {
    delete (window as unknown as W).__DASHI_BASE__
})

describe('einkPreviewUrl', () => {
    const params = {
        dashboardId: 'abc123',
        width: 800,
        height: 480,
        format: 'png-bw' as const,
        rotation: 90 as const,
        page: 2,
    }

    it('puts every display param in the query of the dashboard path', () => {
        const url = new URL(einkPreviewUrl(params), 'http://editor')
        expect(url.pathname).toBe('/abc123')
        expect(Object.fromEntries(url.searchParams)).toEqual({
            format: 'png-bw',
            width: '800',
            height: '480',
            rotation: '90',
            page: '2',
        })
    })

    it('stays under the ingress base', () => {
        ;(window as unknown as W).__DASHI_BASE__ = '/api/hassio_ingress/tok/'
        expect(einkPreviewUrl(params)).toMatch(/^\/api\/hassio_ingress\/tok\/abc123\?/)
    })
})

describe('clampSide', () => {
    it('keeps sizes within 1..MAX_PREVIEW_SIDE and drops fractions', () => {
        expect(clampSide(800.7)).toBe(800)
        expect(clampSide(0)).toBe(1)
        expect(clampSide(-5)).toBe(1)
        expect(clampSide(null)).toBe(1)
        expect(clampSide(99999)).toBe(MAX_PREVIEW_SIDE)
    })
})

describe('DEVICE_PRESETS', () => {
    it('only uses formats the preview can show and unique keys', () => {
        const keys = new Set(DEVICE_PRESETS.map(p => p.key))
        expect(keys.size).toBe(DEVICE_PRESETS.length)
        for (const p of DEVICE_PRESETS) {
            expect(['png-bw', 'png-spectra6']).toContain(p.format)
        }
    })
})
