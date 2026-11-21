// A theme is the colour entity a dashboard references by `themeId`: one
// palette per colour mode plus an optional page background. Fonts and icons
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
    light: string[]
    dark?: string[]
}

export interface ThemeImage {
    light: string
    dark?: string
    fit: ImageFit
    position: ImagePosition
    repeat: ImageRepeat
}

// There is no solid colour: the base is the palette's `bg`. An image, when
// present, is painted over the gradient (or over `bg`).
export interface ThemeBackground {
    gradient?: ThemeGradient
    image?: ThemeImage
}

export interface Theme {
    id: string
    name: string
    // Set by the server on read-only builtins; never sent back.
    builtin?: boolean
    background: ThemeBackground
    colors: ThemeColors
}

export interface ThemeMeta {
    id: string
    name: string
    builtin?: boolean
    usedBy: number
    previewCss: string
    colors: ThemeColors
}

export type CreateThemeDTO = Omit<Theme, 'id'>
