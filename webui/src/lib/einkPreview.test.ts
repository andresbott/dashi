import { describe, it, expect, afterEach } from 'vitest'
import {
    clampSide,
    einkPreviewUrl,
    parsePreviewQuery,
    previewQuery,
    uprightView,
    DEVICE_PRESETS,
    MAX_PREVIEW_SIDE,
} from '@/lib/einkPreview'

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

describe('parsePreviewQuery', () => {
    const defaults = {
        device: 'bw-7in5',
        width: 800,
        height: 480,
        format: 'png-bw',
        rotation: 0,
        page: 0,
        actualSize: false,
    }

    it('defaults to the first preset for an empty query', () => {
        expect(parsePreviewQuery({})).toEqual(defaults)
    })

    it('reads a preset with its size, ignoring a stray width and height', () => {
        expect(parsePreviewQuery({
            preset: 's6-7in3', width: '10', height: '10', format: 'png', rotation: '180', page: '3', actual: '1',
        })).toEqual({
            device: 's6-7in3', width: 800, height: 480, format: 'png', rotation: 180, page: 3, actualSize: true,
        })
    })

    it('reads a custom size, clamped, and falls back per side', () => {
        expect(parsePreviewQuery({ preset: 'custom', width: '5000', height: '600' })).toMatchObject({
            device: 'custom', width: MAX_PREVIEW_SIDE, height: 600,
        })
        expect(parsePreviewQuery({ preset: 'custom', width: 'abc' })).toMatchObject({
            device: 'custom', width: 800, height: 480,
        })
    })

    it('falls back to defaults for invalid values', () => {
        expect(parsePreviewQuery({
            preset: 'nope', format: 'bw', rotation: '45', page: '1.5', actual: 'yes',
        })).toEqual(defaults)
        expect(parsePreviewQuery({ page: '-1', rotation: '' })).toEqual(defaults)
    })

    it('takes the first of a repeated param', () => {
        expect(parsePreviewQuery({ rotation: ['90', '180'], page: [null, '2'] })).toMatchObject({ rotation: 90, page: 0 })
    })

    it('round-trips through previewQuery', () => {
        const custom = { device: 'custom', width: 1024, height: 600, format: 'png' as const, rotation: 270 as const, page: 1, actualSize: true }
        expect(previewQuery(custom)).toEqual({
            preset: 'custom', width: '1024', height: '600', format: 'png', rotation: '270', page: '1', actual: '1',
        })
        expect(parsePreviewQuery(previewQuery(custom))).toEqual(custom)
        expect(previewQuery({ ...defaults, format: 'png-bw', rotation: 0 })).toEqual({
            preset: 'bw-7in5', format: 'png-bw', rotation: '0', page: '0',
        })
    })
})

// The server turns a 90° render clockwise and a 270° render counter-clockwise
// (RotateImage in internal/dashboard/image/renderer.go); the preview undoes it.
describe('uprightView', () => {
    it.each([
        [0, 800, 480, 0],
        [90, 480, 800, -90],
        [180, 800, 480, 180],
        [270, 480, 800, 90],
    ] as const)('rotation %i shows %ix%i turned by %i°', (rotation, width, height, angle) => {
        expect(uprightView(800, 480, rotation)).toMatchObject({ width, height, angle })
    })

    // Maps a point of the w × h image through the CSS transform (origin at its
    // top-left corner): it must land inside the upright box, with the
    // server's corner mapping undone.
    const apply = (transform: string, w: number, h: number, [x, y]: [number, number]): [number, number] => {
        const steps = transform === 'none' ? [] : transform.split(/\s(?=[a-z])/).reverse()
        let p: [number, number] = [x, y]
        for (const step of steps) {
            const [, fn, args] = step.match(/^(\w+)\((.*)\)$/)!
            const vals = args.split(',').map(s => s.trim())
            const pct = (v: string, size: number) => (parseFloat(v) / 100) * size
            if (fn === 'rotate') {
                const a = (parseFloat(vals[0]) * Math.PI) / 180
                p = [Math.round(p[0] * Math.cos(a) - p[1] * Math.sin(a)), Math.round(p[0] * Math.sin(a) + p[1] * Math.cos(a))]
            } else if (fn === 'translateX') {
                p = [p[0] + pct(vals[0], w), p[1]]
            } else if (fn === 'translateY') {
                p = [p[0], p[1] + pct(vals[0], h)]
            } else if (fn === 'translate') {
                p = [p[0] + pct(vals[0], w), p[1] + pct(vals[1], h)]
            }
        }
        return [p[0] + 0, p[1] + 0]
    }

    // Where the server's RotateImage moves the top-left pixel of the
    // portrait layout (lw × lh), as a corner of the panel image.
    it.each([
        // 90: dst(lh-1-y, x) — the layout's top-left lands top-right.
        { rotation: 90 as const, panelTopLeftCameFrom: 'bottom-left' },
        // 270: dst(y, lw-1-x) — the layout's top-left lands bottom-left.
        { rotation: 270 as const, panelTopLeftCameFrom: 'top-right' },
        { rotation: 180 as const, panelTopLeftCameFrom: 'bottom-right' },
    ])('puts the layout corners back for $rotation°', ({ rotation, panelTopLeftCameFrom }) => {
        const w = 800
        const h = 480
        const v = uprightView(w, h, rotation)
        const box = { 'top-left': [0, 0], 'top-right': [v.width, 0], 'bottom-left': [0, v.height], 'bottom-right': [v.width, v.height] }
        // The panel image's top-left corner must be shown where it was in the layout.
        expect(apply(v.transform, w, h, [0, 0])).toEqual(box[panelTopLeftCameFrom as keyof typeof box])
        // And the whole image stays inside the upright box.
        for (const corner of [[0, 0], [w, 0], [0, h], [w, h]] as [number, number][]) {
            const [x, y] = apply(v.transform, w, h, corner)
            expect(x).toBeGreaterThanOrEqual(0)
            expect(y).toBeGreaterThanOrEqual(0)
            expect(x).toBeLessThanOrEqual(v.width)
            expect(y).toBeLessThanOrEqual(v.height)
        }
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
