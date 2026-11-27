import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { browserCss, cssVars, isValidBackground, pageBgValue, previewStyle } from '@/lib/themeCss'
import type { ModeBackground, Palette, Theme, ThemeBackground, ThemeImage } from '@/types/theme'

interface FixtureCase {
    name: string
    theme: Theme
    browserCss: string
}

// The Go generator is the contract. internal/themes/css_test.go
// (TestBrowserCSSMatchesFixture) asserts the same file; if these two
// disagree, this side is wrong.
const fixture: FixtureCase[] = JSON.parse(
    readFileSync(
        resolve(__dirname, '../../../internal/themes/testdata/theme_css_fixture.json'),
        'utf-8',
    ),
)

const light: Palette = {
    bg: '#ffffff', surface: '#f5f5f7', border: '#e0e0e5', text: '#333333', textMuted: '#666666',
    primary: '#0a84ff', onPrimary: '#ffffff', secondary: '#8b5cf6', success: '#16a34a',
    warning: '#d97706', danger: '#dc2626',
}
const dark: Palette = {
    bg: '#1a1a2e', surface: '#23233d', border: '#33334d', text: '#e0e0e0', textMuted: '#999999',
    primary: '#4a90d9', onPrimary: '#ffffff', secondary: '#a78bfa', success: '#22c55e',
    warning: '#f59e0b', danger: '#ef4444',
}

// bg builds a per-mode background; an omitted mode is plain.
const bg = (light: ModeBackground = {}, dark: ModeBackground = {}): ThemeBackground => ({ light, dark })

const grad = (direction: string, stops: string[]): ModeBackground => ({ gradient: { direction, stops } })

const theme = (over: Partial<Theme> = {}): Theme => ({
    id: 'a1', name: 'X', darkMode: true, background: bg(), colors: { light: { ...light }, dark: { ...dark } }, ...over,
})

const withImage = (image: ThemeImage): Theme => theme({ id: 'test', background: bg({ image }) })

describe('browserCss parity with Go', () => {
    it('loads a non-empty fixture', () => {
        expect(fixture.length).toBeGreaterThan(0)
    })

    for (const c of fixture) {
        it(c.name, () => {
            expect(browserCss(c.theme)).toBe(c.browserCss)
        })
    }
})

describe('browserCss', () => {
    it('emits the palette and no page-bg for a plain theme', () => {
        const css = browserCss(theme())
        expect(css).toContain('--dashi-bg:')
        expect(css).not.toContain('--dashi-page-bg')
    })

    it('emits a gradient page-bg', () => {
        const css = browserCss(theme({
            background: bg(grad('135deg', ['#000000', '#ffffff'])),
        }))
        expect(css).toContain('--dashi-page-bg:linear-gradient(')
    })

    it('falls back to the default for a malformed role, like Go withDefaults', () => {
        const t = theme()
        t.colors.light.primary = '#fff;}</style>'
        t.colors.dark.bg = '#00000080' // bg must be opaque
        const css = browserCss(t)
        expect(css).not.toContain('</style')
        expect(css).toContain(':root{--dashi-bg:#ffffff;')
        expect(css).toContain('--dashi-primary:#0a84ff;')
        expect(css).toContain(':root[data-color-mode="dark"]{--dashi-bg:#1a1a2e;')
    })

    it('drops a gradient carrying unsafe characters', () => {
        const css = browserCss(theme({
            background: bg(grad('to right;}</style>', ['#000', '#fff'])),
        }))
        expect(css).not.toContain('</style')
        expect(css).not.toContain('--dashi-page-bg')
    })

    it('drops an image whose enum fields are not recognised', () => {
        const css = browserCss(withImage({
            ref: 'asset:a.png', fit: 'cover', position: 'center;}' as ThemeImage['position'], repeat: 'no-repeat',
        }))
        expect(css).not.toContain('--dashi-page-bg')
    })

    it('renders the light palette and background for dark when darkMode is off', () => {
        const css = browserCss(theme({
            darkMode: false,
            background: bg(grad('90deg', ['#000000', '#ffffff']), grad('90deg', ['#111111', '#222222'])),
        }))
        const [lightRule, darkRule] = css.trimEnd().split('\n')
        expect(darkRule.replace(':root[data-color-mode="dark"]', '')).toBe(lightRule.replace(':root', ''))
        expect(darkRule).toContain('--dashi-bg:#ffffff;')
        expect(darkRule).toContain('linear-gradient(90deg,#000000,#ffffff)')
    })

    it('renders an empty dark background as the plain dark palette, with no fallback to light', () => {
        const css = browserCss(theme({ background: bg(grad('90deg', ['#000000', '#ffffff'])) }))
        const [lightRule, darkRule] = css.trimEnd().split('\n')
        expect(lightRule).toContain('--dashi-page-bg:linear-gradient(90deg,#000000,#ffffff);')
        expect(darkRule).not.toContain('--dashi-page-bg')
    })

    it('uses each mode\'s own image fit, position and repeat', () => {
        const image = (fit: ThemeImage['fit'], position: ThemeImage['position'], repeat: ThemeImage['repeat']): ModeBackground =>
            ({ image: { ref: 'asset:a.png', fit, position, repeat } })
        const css = browserCss(theme({ background: bg(image('cover', 'center', 'no-repeat'), image('contain', 'top', 'repeat')) }))
        const [lightRule, darkRule] = css.trimEnd().split('\n')
        expect(lightRule).toContain('a.png\') center/cover no-repeat;')
        expect(darkRule).toContain('a.png\') top/contain repeat;')
    })

    it('percent-encodes a single quote in a reference', () => {
        const css = browserCss(withImage({ ref: "asset:it's.png", fit: 'cover', position: 'center', repeat: 'no-repeat' }))
        expect(css).toContain('it%27s.png')
    })
})

describe('isValidBackground', () => {
    const img = { fit: 'cover', position: 'center', repeat: 'no-repeat' } as const

    it.each<[string, ThemeBackground, boolean]>([
        ['an empty background', bg(), true],
        ['a valid gradient', bg(grad('90deg', ['#000', '#ffffff80'])), true],
        ['a half-typed stop', bg(grad('90deg', ['#000', '#ff'])), false],
        ['a single stop', bg(grad('90deg', ['#000'])), false],
        ['a valid dark-only gradient', bg({}, grad('90deg', ['#000', '#fff'])), true],
        ['an invalid dark stop', bg(grad('90deg', ['#000', '#fff']), grad('90deg', ['#000', 'x'])), false],
        ['a single dark stop', bg({}, grad('90deg', ['#000'])), false],
        ['a chosen image', bg({ image: { ref: 'asset:a.png', ...img } }), true],
        ['an image layer with no image', bg({ image: { ref: '', ...img } }), false],
        ['a dark image layer with no image', bg({}, { image: { ref: '', ...img } }), false],
    ])('%s → %s', (_name, background, want) => {
        expect(isValidBackground(background)).toBe(want)
    })
})

describe('previewStyle', () => {
    it('scopes one mode and composites the page background over the bg colour', () => {
        const t = theme({
            background: bg(grad('90deg', ['#000000', '#ffffff']), grad('90deg', ['#111111', '#222222'])),
        })
        const light = previewStyle(t, 'light')
        expect(light['--dashi-bg']).toBe('#ffffff')
        expect(light['--dashi-page-bg']).toBe('linear-gradient(90deg,#000000,#ffffff)')
        expect(light.background).toBe('var(--dashi-page-bg, none), var(--dashi-bg)')
        expect(previewStyle(t, 'dark')['--dashi-page-bg']).toBe('linear-gradient(90deg,#111111,#222222)')
    })

    it('has no page-bg for a plain theme', () => {
        expect(previewStyle(theme())['--dashi-page-bg']).toBeUndefined()
    })
})

describe('cssVars', () => {
    it('maps one mode rule to its custom properties', () => {
        const t = theme({ background: bg({}, grad('90deg', ['#111111', '#222222'])) })
        const vars = cssVars(browserCss(t), 'dark')
        expect(Object.keys(vars)).toHaveLength(12)
        expect(vars['--dashi-surface']).toBe('#23233d')
        expect(vars['--dashi-page-bg']).toBe('linear-gradient(90deg,#111111,#222222)')
    })

    it('returns an empty map for an empty stylesheet', () => {
        expect(cssVars('')).toEqual({})
    })
})

describe('pageBgValue', () => {
    it('returns the --dashi-bg value when there is no page-bg', () => {
        const css = browserCss(theme())
        expect(pageBgValue(css)).toBe('#ffffff')
        expect(pageBgValue(css, 'dark')).toBe('#1a1a2e')
    })

    it('round-trips every fixture case', () => {
        for (const c of fixture) {
            for (const mode of ['light', 'dark'] as const) {
                const v = pageBgValue(c.browserCss, mode)
                expect(v).not.toBe('transparent')
                const selector = mode === 'dark' ? ':root[data-color-mode="dark"]{' : ':root{'
                const start = c.browserCss.indexOf(selector)
                const rule = c.browserCss.slice(start, c.browserCss.indexOf('\n', start))
                const declared = rule.includes('--dashi-page-bg:')
                    ? `--dashi-page-bg:${v};}`
                    : `--dashi-bg:${v};`
                expect(rule).toContain(declared)
            }
        }
    })

    it('does not mistake the dark rule for the light one', () => {
        const css = browserCss(theme({
            background: bg(grad('90deg', ['#111111', '#222222']), grad('90deg', ['#333333', '#444444'])),
        }))
        expect(pageBgValue(css, 'light')).toBe('linear-gradient(90deg,#111111,#222222)')
        expect(pageBgValue(css, 'dark')).toBe('linear-gradient(90deg,#333333,#444444)')
    })

    it('returns transparent for an empty stylesheet', () => {
        expect(pageBgValue('')).toBe('transparent')
        expect(pageBgValue('', 'dark')).toBe('transparent')
    })

    it('keeps a multi-layer value intact', () => {
        const css = browserCss(theme({
            background: bg({
                gradient: { direction: 'to right', stops: ['#000000', '#ffffff'] },
                image: { ref: 'asset:a.png', fit: 'cover', position: 'center', repeat: 'no-repeat' },
            }),
        }))
        expect(pageBgValue(css)).toBe(
            "url('/api/v0/themes/a1/assets/a.png') center/cover no-repeat,linear-gradient(to right,#000000,#ffffff)",
        )
    })
})

describe('urlSegment encoding parity with Go url.PathEscape', () => {
    // Expected values come from Go's url.PathEscape output. JS
    // encodeURIComponent under-escapes ! ' ( ) * and over-escapes $ & + : = @.
    const testCases: Array<{ char: string; expected: string; desc: string }> = [
        { char: '!', expected: '%21', desc: 'exclamation' },
        { char: "'", expected: '%27', desc: 'single quote' },
        { char: '(', expected: '%28', desc: 'open paren' },
        { char: ')', expected: '%29', desc: 'close paren' },
        { char: '*', expected: '%2A', desc: 'asterisk' },
        { char: '$', expected: '$', desc: 'dollar' },
        { char: '&', expected: '&', desc: 'ampersand' },
        { char: '+', expected: '+', desc: 'plus' },
        { char: ':', expected: ':', desc: 'colon' },
        { char: '=', expected: '=', desc: 'equals' },
        { char: '@', expected: '@', desc: 'at sign' },
        { char: ',', expected: '%2C', desc: 'comma (both escape)' },
        { char: ';', expected: '%3B', desc: 'semicolon (both escape)' },
        { char: '~', expected: '~', desc: 'tilde (neither escapes)' },
        { char: '%', expected: '%25', desc: 'percent' },
    ]

    testCases.forEach(({ char, expected, desc }) => {
        it(`encodes ${desc} (${char}) to ${expected}`, () => {
            const css = browserCss(withImage({
                ref: `asset:file${char}name.png`, fit: 'cover', position: 'center', repeat: 'no-repeat',
            }))
            expect(css).toContain(`/api/v0/themes/test/assets/file${expected}name.png`)
        })
    })

    it('matches Go output for a composite filename', () => {
        const css = browserCss(withImage({
            ref: 'asset:photo (2024-08-22)!.png', fit: 'cover', position: 'center', repeat: 'no-repeat',
        }))
        expect(css).toContain('photo%20%282024-08-22%29%21.png')
    })
})
