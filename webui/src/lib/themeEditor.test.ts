import { describe, it, expect } from 'vitest'
import { emptyPalette, isDarkEmpty, seedDark, uniqueAssetName } from './themeEditor'
import { PALETTE_ROLES } from './themeCss'
import type { Palette, Theme } from '@/types/theme'

const light: Palette = {
    bg: '#ffffff', surface: '#f5f5f7', border: '#e0e0e5', text: '#333333', textMuted: '#666666',
    primary: '#0a84ff', onPrimary: '#ffffff', secondary: '#8b5cf6', success: '#16a34a',
    warning: '#d97706', danger: '#dc2626',
}

function makeTheme(over: Partial<Theme> = {}): Theme {
    return {
        id: 't1',
        name: 'Mine',
        darkMode: false,
        colors: { light: { ...light }, dark: emptyPalette() },
        background: {
            light: {
                gradient: { direction: 'to bottom', stops: ['#112233', '#445566'] },
                image: { ref: 'asset:sky.png', fit: 'cover', position: 'center', repeat: 'no-repeat' },
            },
            dark: {},
        },
        ...over,
    }
}

describe('uniqueAssetName', () => {
    const now = 1_700_000_000_000
    const stamp = now.toString(36)

    it('inserts a base-36 timestamp before the extension', () => {
        expect(uniqueAssetName('sky.png', now)).toBe(`sky-${stamp}.png`)
    })

    it('keeps everything before the last dot as the base', () => {
        expect(uniqueAssetName('my.photo.JPG', now)).toBe(`my.photo-${stamp}.JPG`)
    })

    it('appends the timestamp when there is no extension', () => {
        expect(uniqueAssetName('sky', now)).toBe(`sky-${stamp}`)
        expect(uniqueAssetName('.png', now)).toBe(`.png-${stamp}`)
    })

    it('gives a different name for a different time', () => {
        expect(uniqueAssetName('sky.png', now)).not.toBe(uniqueAssetName('sky.png', now + 1))
    })
})

describe('emptyPalette', () => {
    it('has every role set to the empty string', () => {
        const p = emptyPalette()
        expect(Object.keys(p).sort()).toEqual(PALETTE_ROLES.map(r => r.role).sort())
        expect(Object.values(p).every(v => v === '')).toBe(true)
    })

    it('returns a fresh object each call', () => {
        const a = emptyPalette()
        a.bg = '#000000'
        expect(emptyPalette().bg).toBe('')
    })
})

describe('isDarkEmpty', () => {
    it('is true for an empty dark palette and background', () => {
        expect(isDarkEmpty(makeTheme())).toBe(true)
    })

    it('is true for a partial palette with no values and a missing background', () => {
        const t = makeTheme()
        t.colors.dark = {} as Palette
        ;(t.background as { dark?: unknown }).dark = undefined
        expect(isDarkEmpty(t)).toBe(true)
    })

    it('is false once any dark role is set', () => {
        const t = makeTheme()
        t.colors.dark.primary = '#123456'
        expect(isDarkEmpty(t)).toBe(false)
    })

    it('is false with a dark gradient or image', () => {
        const g = makeTheme()
        g.background.dark = { gradient: { direction: 'to top', stops: ['#000000', '#111111'] } }
        expect(isDarkEmpty(g)).toBe(false)

        const i = makeTheme()
        i.background.dark = { image: { ref: 'asset:night.png', fit: 'cover', position: 'center', repeat: 'no-repeat' } }
        expect(isDarkEmpty(i)).toBe(false)
    })
})

describe('seedDark', () => {
    it('copies every empty dark role from light', () => {
        const seeded = seedDark(makeTheme())
        expect(seeded.colors.dark).toEqual(light)
    })

    it('copies only the light gradient, never the image', () => {
        const seeded = seedDark(makeTheme())
        expect(seeded.background.dark).toEqual({
            gradient: { direction: 'to bottom', stops: ['#112233', '#445566'] },
        })
        expect('image' in seeded.background.dark).toBe(false)
    })

    it('leaves the dark background empty when light has no gradient', () => {
        const t = makeTheme()
        t.background.light = { image: { ref: 'asset:sky.png', fit: 'cover', position: 'center', repeat: 'no-repeat' } }
        const seeded = seedDark(t)
        expect(seeded.background.dark).toEqual({})
    })

    it('does not alias the light data', () => {
        const seeded = seedDark(makeTheme())
        seeded.background.dark.gradient!.stops[0] = '#999999'
        seeded.colors.dark.bg = '#000000'
        expect(seeded.background.light.gradient!.stops[0]).toBe('#112233')
        expect(seeded.colors.light.bg).toBe('#ffffff')
    })

    it('does not mutate its input', () => {
        const t = makeTheme()
        const before = JSON.stringify(t)
        seedDark(t)
        expect(JSON.stringify(t)).toBe(before)
    })

    it('keeps dark roles that are already set', () => {
        const t = makeTheme()
        t.colors.dark.primary = '#123456'
        const seeded = seedDark(t)
        expect(seeded.colors.dark.primary).toBe('#123456')
        expect(seeded.colors.dark.bg).toBe(light.bg)
    })

    it('fills roles missing from a partial dark palette', () => {
        const t = makeTheme()
        t.colors.dark = { primary: '#123456' } as Palette
        const seeded = seedDark(t)
        expect(seeded.colors.dark).toEqual({ ...light, primary: '#123456' })
    })

    it('keeps an existing dark background', () => {
        const t = makeTheme()
        const night = { ref: 'asset:night.png', fit: 'contain' as const, position: 'top' as const, repeat: 'repeat' as const }
        t.background.dark = { image: night }
        const seeded = seedDark(t)
        expect(seeded.background.dark).toEqual({ image: night })
    })

    it('leaves an already seeded theme unchanged on a re-seed', () => {
        const once = seedDark(makeTheme())
        once.colors.dark.text = '#eeeeee'
        once.background.dark.gradient!.direction = 'to right'
        expect(seedDark(once)).toEqual(once)
    })

    it('copes with a missing dark background', () => {
        const t = makeTheme()
        ;(t.background as { dark?: unknown }).dark = undefined
        const seeded = seedDark(t)
        expect(seeded.background.dark).toEqual({
            gradient: { direction: 'to bottom', stops: ['#112233', '#445566'] },
        })
    })
})
