import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { browserCss, pageBgValue } from '@/lib/backgroundCss'
import type { Background } from '@/types/background'

interface FixtureCase {
    name: string
    background: Background
    browserCss: string
}

// The Go generator is the contract. internal/backgrounds/css_test.go asserts
// the same file; if these two disagree, this side is wrong.
const fixture: FixtureCase[] = JSON.parse(
    readFileSync(
        resolve(__dirname, '../../../internal/backgrounds/testdata/css_fixture.json'),
        'utf-8',
    ),
)

describe('browserCss parity with Go', () => {
    it('loads a non-empty fixture', () => {
        expect(fixture.length).toBeGreaterThan(0)
    })

    for (const c of fixture) {
        it(c.name, () => {
            expect(browserCss(c.background)).toBe(c.browserCss)
        })
    }
})

// pageBgValue parses what browserCss (and Go's BrowserCSS) produce. Driving it
// from the same fixture means a change to the rule format breaks the generator
// tests and this parser together, instead of leaving swatches silently blank.
describe('pageBgValue round-trips every fixture case', () => {
    for (const c of fixture) {
        it(c.name, () => {
            const light = pageBgValue(c.browserCss, 'light')
            const dark = pageBgValue(c.browserCss, 'dark')

            expect(light).not.toBe('transparent')
            expect(dark).not.toBe('transparent')

            // The extracted values must be exactly what the stylesheet declares.
            expect(c.browserCss).toContain(`:root{--dashi-page-bg:${light};}`)
            expect(c.browserCss).toContain(`:root[data-color-mode="dark"]{--dashi-page-bg:${dark};}`)
        })
    }
})

describe('pageBgValue edge cases', () => {
    it('defaults to light mode', () => {
        const css = ':root{--dashi-page-bg:#ffffff;}\n:root[data-color-mode="dark"]{--dashi-page-bg:#000000;}\n'
        expect(pageBgValue(css)).toBe('#ffffff')
    })

    it('does not mistake the dark rule for the light one', () => {
        // Both rules start with ":root", so a lazy pattern could match the dark
        // rule's prefix and return the wrong colour.
        const css = ':root{--dashi-page-bg:#ffffff;}\n:root[data-color-mode="dark"]{--dashi-page-bg:#000000;}\n'
        expect(pageBgValue(css, 'light')).toBe('#ffffff')
        expect(pageBgValue(css, 'dark')).toBe('#000000')
    })

    it('returns transparent for an empty stylesheet', () => {
        expect(pageBgValue('')).toBe('transparent')
        expect(pageBgValue('', 'dark')).toBe('transparent')
    })

    it('keeps a multi-layer value intact', () => {
        const layered =
            "url('/api/v0/data/backgrounds/a.png') center/cover no-repeat,linear-gradient(to right,#000000,#ffffff)"
        const css = `:root{--dashi-page-bg:${layered};}\n:root[data-color-mode="dark"]{--dashi-page-bg:${layered};}\n`
        expect(pageBgValue(css)).toBe(layered)
    })
})

describe('browserCss edge cases', () => {
    it('returns an empty string when nothing is configured', () => {
        expect(browserCss({ id: 'a1', name: 'Empty' })).toBe('')
    })

    it('drops a base value that is not a hex colour or a linear-gradient', () => {
        const bg = { id: 'a1', name: 'X', color: { light: '#fff;}</style>' } } as Background
        expect(browserCss(bg)).not.toContain('</style')
    })

    it('percent-encodes a single quote in a reference', () => {
        const bg: Background = {
            id: 'a1', name: 'X',
            image: { light: "shared:it's.png", fit: 'cover', position: 'center', repeat: 'no-repeat' },
        }
        expect(browserCss(bg)).toContain('it%27s.png')
    })
})

describe('urlSegment encoding parity with Go url.PathEscape', () => {
    // urlSegment must match Go's url.PathEscape exactly. JS encodeURIComponent
    // differs on ! * ( ) + ~ characters; we reconcile these in the implementation.
    // Each test verifies an asset: reference encodes to the Go output.

    it('escapes parentheses like Go does', () => {
        const bg: Background = {
            id: 'test', name: 'Photo',
            image: { light: 'asset:my (photo).png', fit: 'cover', position: 'center', repeat: 'no-repeat' },
        }
        // Go encodes () to %28 and %29
        expect(browserCss(bg)).toContain('my%20%28photo%29.png')
    })

    it('escapes exclamation like Go does', () => {
        const bg: Background = {
            id: 'test', name: 'Bang',
            image: { light: 'asset:a!b.png', fit: 'cover', position: 'center', repeat: 'no-repeat' },
        }
        // Go encodes ! to %21
        expect(browserCss(bg)).toContain('a%21b.png')
    })

    it('escapes asterisk like Go does', () => {
        const bg: Background = {
            id: 'test', name: 'Star',
            image: { light: 'asset:x*y.png', fit: 'cover', position: 'center', repeat: 'no-repeat' },
        }
        // Go encodes * to %2A
        expect(browserCss(bg)).toContain('x%2Ay.png')
    })

    it('does not escape plus like JS would', () => {
        const bg: Background = {
            id: 'test', name: 'Plus',
            image: { light: 'asset:p+q.png', fit: 'cover', position: 'center', repeat: 'no-repeat' },
        }
        // Go leaves + literal; JS would encode to %2B, but we undo it
        expect(browserCss(bg)).toContain('p+q.png')
    })

    it('does not escape tilde', () => {
        const bg: Background = {
            id: 'test', name: 'Tilde',
            image: { light: 'asset:t~u.png', fit: 'cover', position: 'center', repeat: 'no-repeat' },
        }
        // Both Go and JS leave ~ literal
        expect(browserCss(bg)).toContain('t~u.png')
    })

    it('matches Go output for composite filename with multiple special chars', () => {
        const bg: Background = {
            id: 'test', name: 'Complex',
            image: { light: 'asset:photo (2024-08-22)!.png', fit: 'cover', position: 'center', repeat: 'no-repeat' },
        }
        // ( ) ! all get encoded; spaces get %20
        expect(browserCss(bg)).toContain('photo%20%282024-08-22%29%21.png')
    })
})

describe('urlSegment exhaustive character encoding', () => {
    // Table-driven test: all 11 divergent characters between Go's url.PathEscape
    // and JS encodeURIComponent, plus characters that must remain unchanged.
    // Expected values come from Go's url.PathEscape output.
    const testCases: Array<{ char: string; expected: string; desc: string }> = [
        // JS UNDER-escapes (Go escapes, JS leaves literal)
        { char: '!', expected: '%21', desc: 'exclamation' },
        { char: "'", expected: '%27', desc: 'single quote' },
        { char: '(', expected: '%28', desc: 'open paren' },
        { char: ')', expected: '%29', desc: 'close paren' },
        { char: '*', expected: '%2A', desc: 'asterisk' },
        // JS OVER-escapes (Go leaves literal, JS escapes)
        { char: '$', expected: '$', desc: 'dollar' },
        { char: '&', expected: '&', desc: 'ampersand' },
        { char: '+', expected: '+', desc: 'plus' },
        { char: ':', expected: ':', desc: 'colon' },
        { char: '=', expected: '=', desc: 'equals' },
        { char: '@', expected: '@', desc: 'at sign' },
        // Characters that both agree on (must remain unchanged)
        { char: ',', expected: '%2C', desc: 'comma (both escape)' },
        { char: ';', expected: '%3B', desc: 'semicolon (both escape)' },
    ]

    testCases.forEach(({ char, expected, desc }) => {
        it(`encodes ${desc} (${char}) to ${expected}`, () => {
            const bg: Background = {
                id: 'test', name: `Test ${desc}`,
                image: { light: `asset:file${char}name.png`, fit: 'cover', position: 'center', repeat: 'no-repeat' },
            }
            expect(browserCss(bg)).toContain(`file${expected}name.png`)
        })
    })

    it('handles percent sign as literal (becomes %25) without corruption', () => {
        const bg: Background = {
            id: 'test', name: 'Percent',
            image: { light: 'asset:file%name.png', fit: 'cover', position: 'center', repeat: 'no-repeat' },
        }
        // Literal % becomes %25; the test ensures replace() doesn't corrupt it
        expect(browserCss(bg)).toContain('file%25name.png')
    })
})
