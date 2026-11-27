// A theme is the colour entity a dashboard references by `themeId`: one
// palette and one optional page background per colour mode. Fonts and icons
// live in the separate widget theme (see types/widgetTheme.ts).
// Mirrors internal/themes (palette.go, types.go).

// Palette is the set of colour roles for one colour mode. Every value is a hex
// colour (#rgb, #rrggbb or #rrggbbaa); `bg` must be opaque because it is the
// page base everything else is composited onto.
export interface Palette {
    bg: string
    surface: string
    border: string
    text: string
    textMuted: string
    primary: string
    onPrimary: string
    secondary: string
    success: string
    warning: string
    danger: string
}

export type PaletteRole = keyof Palette

export interface ThemeColors {
    light: Palette
    dark: Palette
}

export type ImageFit = 'cover' | 'contain' | 'stretch' | 'original'

export type ImagePosition =
    | 'center' | 'top' | 'bottom' | 'left' | 'right'
    | 'top left' | 'top right' | 'bottom left' | 'bottom right'

export type ImageRepeat = 'no-repeat' | 'repeat' | 'repeat-x' | 'repeat-y'

export interface ThemeGradient {
    direction: string
    stops: string[]
}

export interface ThemeImage {
    ref: string
    fit: ImageFit
    position: ImagePosition
    repeat: ImageRepeat
}

// One colour mode's page background. There is no solid colour: the base is
// that mode's palette `bg`. An image, when present, is painted over the
// gradient (or over `bg`).
export interface ModeBackground {
    gradient?: ThemeGradient
    image?: ThemeImage
}

// A background per colour mode, mirroring ThemeColors. The modes are
// independent: an empty dark background renders the plain dark palette, with
// no fallback to light.
export interface ThemeBackground {
    light: ModeBackground
    dark: ModeBackground
}

export interface Theme {
    id: string
    name: string
    // Set by the server on read-only builtins; never sent back.
    builtin?: boolean
    // Whether the theme has a dark variant. Off: every colour mode renders the
    // light palette and background; the dark data stays stored but unused.
    // The server rejects a POST/PUT body without it.
    darkMode: boolean
    background: ThemeBackground
    colors: ThemeColors
}

export interface ThemeMeta {
    id: string
    name: string
    builtin?: boolean
    // With darkMode off, `colors.dark` is the light palette (rendered view).
    darkMode: boolean
    usedBy: number
    previewCss: string
    colors: ThemeColors
}

export type CreateThemeDTO = Omit<Theme, 'id'>
