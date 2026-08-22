export type ImageFit = 'cover' | 'contain' | 'stretch' | 'original'

export type ImagePosition =
    | 'center' | 'top' | 'bottom' | 'left' | 'right'
    | 'top left' | 'top right' | 'bottom left' | 'bottom right'

export type ImageRepeat = 'no-repeat' | 'repeat' | 'repeat-x' | 'repeat-y'

export interface BackgroundColor {
    light: string
    dark?: string
}

export interface BackgroundGradient {
    direction: string
    light: string[]
    dark?: string[]
}

export interface BackgroundImage {
    light: string
    dark?: string
    fit: ImageFit
    position: ImagePosition
    repeat: ImageRepeat
}

// color and gradient are mutually exclusive; either may combine with image.
export interface Background {
    id: string
    name: string
    color?: BackgroundColor
    gradient?: BackgroundGradient
    image?: BackgroundImage
}

export interface BackgroundMeta {
    id: string
    name: string
    usedBy: number
    previewCss: string
}

export type CreateBackgroundDTO = Omit<Background, 'id'>
