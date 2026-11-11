import type { ImageFit, Palette, PaletteRole, Theme, ThemeBackground } from '@/types/theme'
import { withBase } from '@/lib/base'

// This file mirrors internal/themes/css.go and palette.go byte for byte so the
// theme editor can preview without a round-trip. Go is the contract; the
// shared fixture at internal/themes/testdata/theme_css_fixture.json is
// asserted by both sides, so drift fails CI rather than shipping.

const HEX = /^#(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{6}|[0-9a-fA-F]{8})$/
const OPAQUE_HEX = /^#(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{6})$/

// PALETTE_ROLES is in CSS variable output order (Go: paletteRoles).
export const PALETTE_ROLES: readonly { role: PaletteRole; cssVar: string }[] = [
    { role: 'bg', cssVar: '--dashi-bg' },
    { role: 'surface', cssVar: '--dashi-surface' },
    { role: 'border', cssVar: '--dashi-border' },
    { role: 'text', cssVar: '--dashi-text' },
    { role: 'textMuted', cssVar: '--dashi-text-muted' },
    { role: 'primary', cssVar: '--dashi-primary' },
    { role: 'onPrimary', cssVar: '--dashi-on-primary' },
    { role: 'secondary', cssVar: '--dashi-secondary' },
    { role: 'success', cssVar: '--dashi-success' },
    { role: 'warning', cssVar: '--dashi-warning' },
    { role: 'danger', cssVar: '--dashi-danger' },
]

// Go: DefaultPalette. Used only to replace a malformed role, as Go's
// withDefaults does, so a half-typed hex in the editor previews the way the
// server would render it.
const DEFAULT_PALETTES: Record<'light' | 'dark', Palette> = {
    light: {
        bg: '#ffffff', surface: '#f5f5f7', border: '#e0e0e5',
        text: '#333333', textMuted: '#666666',
        primary: '#0a84ff', onPrimary: '#ffffff',
        secondary: '#8b5cf6', success: '#16a34a', warning: '#d97706', danger: '#dc2626',
    },
    dark: {
        bg: '#1a1a2e', surface: '#23233d', border: '#33334d',
        text: '#e0e0e0', textMuted: '#999999',
        primary: '#4a90d9', onPrimary: '#ffffff',
        secondary: '#a78bfa', success: '#22c55e', warning: '#f59e0b', danger: '#ef4444',
    },
}

// isValidRoleColor reports whether a value is acceptable for a role: any hex
// colour, and for `bg` only an opaque one.
export const isValidRoleColor = (role: PaletteRole, value: string): boolean =>
    (role === 'bg' ? OPAQUE_HEX : HEX).test(value)

const FIT_VALUES = new Set(['cover', 'contain', 'stretch', 'original'])
const POSITION_VALUES = new Set([
    'center', 'top', 'bottom', 'left', 'right',
    'top left', 'top right', 'bottom left', 'bottom right',
])
const REPEAT_VALUES = new Set(['no-repeat', 'repeat', 'repeat-x', 'repeat-y'])

// Mirror Go's url.PathEscape exactly, measured character by character.
// Go escapes ! ' ( ) * which encodeURIComponent leaves literal; Go leaves
// $ & + : = @ literal in a path segment, which encodeURIComponent escapes.
// (Note: , and ; agree between the two implementations.)
const urlSegment = (s: string): string =>
    encodeURIComponent(s)
        .replace(/[!'()*]/g, c => '%' + c.charCodeAt(0).toString(16).toUpperCase())
        .replace(/%(24|26|2B|3A|3D|40)/g, (_m, hex) => String.fromCharCode(parseInt(hex, 16)))

const refUrl = (themeId: string, ref: string): string | null => {
    const idx = ref.indexOf(':')
    if (idx < 0) return null
    const scheme = ref.slice(0, idx)
    const path = ref.slice(idx + 1)
    if (!path) return null
    if (scheme === 'asset') {
        const segments = path.split('/').map(urlSegment).join('/')
        return withBase(`/api/v0/themes/${urlSegment(themeId)}/assets/${segments}`)
    }
    if (scheme === 'shared') {
        return withBase(`/api/v0/data/backgrounds/${urlSegment(path)}`)
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

// Defence in depth for hand-edited data: the output lands in a <style> block.
const safeBaseValue = (s: string): boolean => {
    if (!s.startsWith('linear-gradient(') || !s.endsWith(')')) return false
    return !/[<>;{}"'\\]/.test(s)
}

interface Variant {
    image: string
    baseValue: string
}

const resolve = (bg: ThemeBackground, dark: boolean): Variant => {
    const v: Variant = { image: '', baseValue: '' }
    if (bg.image) {
        v.image = dark && bg.image.dark ? bg.image.dark : bg.image.light
    }
    if (bg.gradient) {
        const stops = dark && bg.gradient.dark?.length ? bg.gradient.dark : bg.gradient.light
        if (stops?.length) {
            v.baseValue = `linear-gradient(${bg.gradient.direction},${stops.join(',')})`
        }
    }
    return v
}

const pageBg = (theme: Theme, v: Variant): string => {
    const layers: string[] = []
    const img = theme.background?.image
    if (
        v.image && img &&
        FIT_VALUES.has(img.fit) && POSITION_VALUES.has(img.position) && REPEAT_VALUES.has(img.repeat)
    ) {
        const url = refUrl(theme.id, v.image)
        if (url) {
            layers.push(`url('${url}') ${img.position}/${cssSize(img.fit)} ${img.repeat}`)
        }
    }
    if (v.baseValue && safeBaseValue(v.baseValue)) {
        layers.push(v.baseValue)
    }
    return layers.join(',')
}

const rule = (selector: string, theme: Theme, mode: 'light' | 'dark'): string => {
    const palette = theme.colors?.[mode]
    const defaults = DEFAULT_PALETTES[mode]
    let body = ''
    for (const { role, cssVar } of PALETTE_ROLES) {
        const value = palette?.[role] ?? ''
        body += `${cssVar}:${isValidRoleColor(role, value) ? value : defaults[role]};`
    }
    const bg = pageBg(theme, resolve(theme.background ?? {}, mode === 'dark'))
    if (bg) body += `--dashi-page-bg:${bg};`
    return `${selector}{${body}}\n`
}

// browserCss returns the body of the page's theme <style> block: the palette
// vars, plus --dashi-page-bg when the theme has a gradient or image, for light
// on :root and for dark under [data-color-mode="dark"]. Never empty.
export function browserCss(theme: Theme): string {
    return rule(':root', theme, 'light') + rule(':root[data-color-mode="dark"]', theme, 'dark')
}

const LIGHT_RULE = /(?:^|\n):root\{([^}]*)\}/
const DARK_RULE = /:root\[data-color-mode="dark"\]\{([^}]*)\}/

// cssVars parses one colour mode's rule out of a stylesheet in the shape
// browserCss and Go's BrowserCSS produce, as a map of custom property to
// value. It lets UI scope a theme to one element through an inline style
// instead of :root. No value can contain ';' or '}': gradients carrying them
// are dropped and URLs percent-encode them.
export function cssVars(css: string, mode: 'light' | 'dark' = 'light'): Record<string, string> {
    const body = css.match(mode === 'dark' ? DARK_RULE : LIGHT_RULE)?.[1] ?? ''
    const vars: Record<string, string> = {}
    for (const decl of body.split(';')) {
        const idx = decl.indexOf(':')
        if (idx > 0) vars[decl.slice(0, idx)] = decl.slice(idx + 1)
    }
    return vars
}

// pageBgValue extracts what one colour mode's page actually shows out of a
// generated stylesheet, for UI that has only the server's `previewCss` string
// and no full Theme to regenerate from (the admin list, the settings swatch):
// the --dashi-page-bg value, or the --dashi-bg colour when there is none.
//
// The fixture-driven tests covering the stylesheet shape also cover this
// parser: if the rule format ever changes, they fail together instead of this
// silently returning 'transparent' and every swatch going blank.
export function pageBgValue(previewCss: string, mode: 'light' | 'dark' = 'light'): string {
    const vars = cssVars(previewCss, mode)
    return vars['--dashi-page-bg'] ?? vars['--dashi-bg'] ?? 'transparent'
}
