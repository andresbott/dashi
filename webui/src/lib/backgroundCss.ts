import type { Background, ImageFit } from '@/types/background'

// This file mirrors internal/backgrounds/css.go byte for byte so the editor
// can preview a background without a round-trip. Go is the contract; the
// shared fixture at internal/backgrounds/testdata/css_fixture.json is
// asserted by both sides, so drift fails CI rather than shipping.

const HEX = /^#(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{6}|[0-9a-fA-F]{8})$/

// Mirror Go's url.PathEscape exactly, measured character by character.
// Go escapes ! ' ( ) * which encodeURIComponent leaves literal; Go leaves
// $ & + : = @ literal in a path segment, which encodeURIComponent escapes.
// (Note: , and ; agree between the two implementations.)
const urlSegment = (s: string): string =>
    encodeURIComponent(s)
        .replace(/[!'()*]/g, c => '%' + c.charCodeAt(0).toString(16).toUpperCase())
        .replace(/%(24|26|2B|3A|3D|40)/g, (_m, hex) => String.fromCharCode(parseInt(hex, 16)))

const refUrl = (bgId: string, ref: string): string | null => {
    const idx = ref.indexOf(':')
    if (idx < 0) return null
    const scheme = ref.slice(0, idx)
    const path = ref.slice(idx + 1)
    if (!path) return null
    if (scheme === 'asset') {
        const segments = path.split('/').map(urlSegment).join('/')
        return `/api/v0/backgrounds/${urlSegment(bgId)}/assets/${segments}`
    }
    if (scheme === 'shared') {
        return `/api/v0/data/backgrounds/${urlSegment(path)}`
    }
    return null
}

const cssSize = (fit: ImageFit | undefined): string => {
    switch (fit) {
        case 'contain': return 'contain'
        case 'stretch': return '100% 100%'
        case 'original': return 'auto'
        default: return 'cover'
    }
}

const safeBaseValue = (s: string): boolean => {
    if (HEX.test(s)) return true
    if (!s.startsWith('linear-gradient(') || !s.endsWith(')')) return false
    return !/[<>;{}"'\\]/.test(s)
}

interface Variant {
    image: string
    baseValue: string
}

const resolve = (bg: Background, dark: boolean): Variant => {
    const v: Variant = { image: '', baseValue: '' }
    if (bg.image) {
        v.image = dark && bg.image.dark ? bg.image.dark : bg.image.light
    }
    if (bg.color) {
        v.baseValue = dark && bg.color.dark ? bg.color.dark : bg.color.light
    } else if (bg.gradient) {
        const stops = dark && bg.gradient.dark?.length ? bg.gradient.dark : bg.gradient.light
        if (stops?.length) {
            v.baseValue = `linear-gradient(${bg.gradient.direction},${stops.join(',')})`
        }
    }
    return v
}

const variantValue = (bg: Background, v: Variant): string => {
    const layers: string[] = []
    if (v.image && bg.image) {
        const url = refUrl(bg.id, v.image)
        if (url) {
            layers.push(`url('${url}') ${bg.image.position}/${cssSize(bg.image.fit)} ${bg.image.repeat}`)
        }
    }
    if (v.baseValue && safeBaseValue(v.baseValue)) {
        layers.push(v.baseValue)
    }
    return layers.join(',')
}

// browserCss returns the body of the page's background <style> block, or ''
// when the background defines nothing.
export function browserCss(bg: Background): string {
    const light = variantValue(bg, resolve(bg, false))
    const dark = variantValue(bg, resolve(bg, true))
    if (!light && !dark) return ''
    return (
        `:root{--dashi-page-bg:${light};}\n` +
        `:root[data-color-mode="dark"]{--dashi-page-bg:${dark};}\n`
    )
}
