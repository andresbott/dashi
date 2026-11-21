import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { isSafeBookmarkUrl, safeBookmarkHref } from './url'

// The case table shared with internal/widgets/bookmark/bookmark_test.go, so
// the editor and the server enforce the same rule.
const { cases } = JSON.parse(
    readFileSync(
        resolve(__dirname, '../../../../internal/widgets/bookmark/testdata/url_cases.json'),
        'utf-8',
    ),
) as { cases: { url: string; valid: boolean }[] }

describe('isSafeBookmarkUrl', () => {
    it('has cases to check', () => {
        expect(cases.length).toBeGreaterThan(0)
    })

    it.each(cases)('$url -> $valid', ({ url, valid }) => {
        expect(isSafeBookmarkUrl(url)).toBe(valid)
    })
})

describe('safeBookmarkHref', () => {
    it('keeps an allowed URL', () => {
        expect(safeBookmarkHref('https://example.com')).toBe('https://example.com')
        expect(safeBookmarkHref('/abc123')).toBe('/abc123')
    })

    it('drops empty, missing and rejected URLs', () => {
        expect(safeBookmarkHref('')).toBeUndefined()
        expect(safeBookmarkHref(undefined)).toBeUndefined()
        expect(safeBookmarkHref('javascript:alert(1)')).toBeUndefined()
        expect(safeBookmarkHref('//evil.com')).toBeUndefined()
    })
})
