// Pure helpers for the theme editor (views/admin/AdminThemeEdit.vue and its
// per-mode components). No Vue, no I/O: everything here takes data and
// returns new data.

import { plainCopy } from './plainCopy'
import { PALETTE_ROLES } from './themeCss'
import type { ModeBackground, Palette, Theme } from '@/types/theme'

// uniqueAssetName gives an upload a name no earlier upload has:
// <base>-<now in base 36>.<ext>. Light and dark may each upload a file with
// the same original name; without this the second upload would silently
// overwrite the other mode's image, and a browser could keep showing the
// cached old file.
export function uniqueAssetName(name: string, now: number): string {
    const stamp = now.toString(36)
    const dot = name.lastIndexOf('.')
    if (dot <= 0) return `${name}-${stamp}`
    return `${name.slice(0, dot)}-${stamp}${name.slice(dot)}`
}

// emptyPalette is a palette with every role unset: the dark palette of a
// theme that has never had dark mode on.
export function emptyPalette(): Palette {
    return Object.fromEntries(PALETTE_ROLES.map(({ role }) => [role, ''])) as unknown as Palette
}

const hasBackground = (bg: ModeBackground | undefined): boolean => !!(bg?.gradient || bg?.image)

// isDarkEmpty reports whether the theme carries no dark data at all: no dark
// role set and no dark gradient or image.
export function isDarkEmpty(t: Theme): boolean {
    const dark = t.colors.dark ?? {}
    return PALETTE_ROLES.every(({ role }) => !dark[role]) && !hasBackground(t.background?.dark)
}

// seedDark returns a copy of the theme with the dark mode filled from light,
// for the moment dark mode is switched on. Each unset dark role takes the
// light value; the dark background takes a copy of the light gradient only
// when it has neither gradient nor image. The light image is never copied:
// an image needs a ref, and sharing one would let removing the dark image
// touch the light one. Dark data that is already there is kept, so seeding
// twice changes nothing.
export function seedDark(t: Theme): Theme {
    const out = plainCopy(t)
    const dark = { ...emptyPalette(), ...out.colors.dark }
    for (const { role } of PALETTE_ROLES) {
        if (!dark[role]) dark[role] = out.colors.light[role]
    }
    out.colors.dark = dark

    out.background ??= { light: {}, dark: {} }
    if (!hasBackground(out.background.dark)) {
        const gradient = out.background.light?.gradient
        out.background.dark = gradient ? { gradient: plainCopy(gradient) } : {}
    }
    return out
}
