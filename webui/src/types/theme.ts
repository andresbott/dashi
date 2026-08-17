export interface FontInfo {
    name: string
}

export type ThemeKind = 'theme' | 'icon' | 'style'

export interface ThemeInfo {
    name: string
    type: ThemeKind
    description: string
    fonts: FontInfo[]
    hasIcons: boolean
    iconType?: 'font' | 'image'
    builtin?: boolean
}

export interface FontIconResponse {
    class: string
}
