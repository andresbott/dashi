export interface Widget {
    id: string
    type: string
    title: string
    width: number
    // 1-based grid column (1..12) where the widget starts, so it can sit
    // anywhere in the row. Omitted / 0 means "flow after the previous widget"
    // (the pre-column behaviour), which keeps older dashboards rendering
    // unchanged. See lib/rowLayout.ts.
    column?: number
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
    // References a theme entity (colours and background, see types/theme.ts)
    // by ID. Empty or dangling renders the builtin "default".
    themeId?: string
    // References a widget theme (fonts and icons, see types/widgetTheme.ts)
    // by name. Empty or unknown renders "default".
    widgetTheme?: string
    colorMode?: ColorMode
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
    themeId?: string
    widgetTheme?: string
    colorMode?: ColorMode
    pages: Page[]
}
