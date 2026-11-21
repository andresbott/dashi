// A widget theme supplies fonts and icons; a dashboard references one by
// name in `widgetTheme`. Colours come from the theme (types/theme.ts).
// Mirrors internal/widgetthemes.Info.

export interface FontInfo {
    name: string
}

export interface WidgetThemeInfo {
    name: string
    description: string
    fonts: FontInfo[]
    hasIcons: boolean
    iconType?: 'font' | 'image'
    builtin?: boolean
}

export interface FontIconResponse {
    class: string
}
