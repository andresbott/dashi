export interface FontInfo {
    name: string
}

export interface ThemeInfo {
    name: string
    description: string
    fonts: FontInfo[]
    hasIcons: boolean
    iconType?: 'font' | 'image'
}

export interface FontIconResponse {
    class: string
}
