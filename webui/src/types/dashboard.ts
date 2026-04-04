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

export interface ImageConfig {
    width: number
    height: number
}

export interface Page {
    name: string
    rows: Row[]
}

export interface Dashboard {
    id: string
    name: string
    icon: string
    type: DashboardType
    container: Container
    imageConfig?: ImageConfig
    theme?: string
    pages: Page[]
}

export interface DashboardMeta {
    id: string
    name: string
    icon: string
    type: DashboardType
}

export interface CreateDashboardDTO {
    id?: string
    name: string
    icon: string
    type: DashboardType
    container: Container
    imageConfig?: ImageConfig
    theme?: string
    pages: Page[]
}
