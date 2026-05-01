export interface Widget {
    id: string
    type: string
    title: string
    width: number
    config?: Record<string, unknown>
}

export interface Row {
    id: string
    title?: string
    height: string
    width: string
    widgets: Widget[]
}

export type DashboardType = 'interactive' | 'image'

export interface Container {
    maxWidth: string
    verticalAlign: 'top' | 'center' | 'bottom'
    horizontalAlign: 'left' | 'center' | 'right'
    showBoxes?: boolean
}

export interface Background {
    type: 'none' | 'image' | 'color' | 'gradient'
    value: string
}

export interface Page {
    name: string
    refreshInterval?: number
    rows: Row[]
}

export type ColorMode = 'auto' | 'light' | 'dark'

export interface Dashboard {
    id: string
    name: string
    icon: string
    type: DashboardType
    default?: boolean
    container: Container
    theme?: string
    colorMode?: ColorMode
    accentColor?: string
    background?: Background
    pages: Page[]
}

export interface DashboardMeta {
    id: string
    name: string
    icon: string
    type: DashboardType
    default?: boolean
}

export interface CreateDashboardDTO {
    id?: string
    name: string
    icon: string
    type: DashboardType
    container: Container
    theme?: string
    colorMode?: ColorMode
    accentColor?: string
    background?: Background
    pages: Page[]
}
