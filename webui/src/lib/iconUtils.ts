export type IconType = 'tabler' | 'selfhst' | 'dashboard'

export interface ParsedIcon {
    type: IconType
    value: string
}

const IMAGE_EXTENSIONS = /\.(png|webp|svg|jpg|jpeg|ico|gif)$/i

export function parseIcon(icon: string): ParsedIcon {
    if (icon.startsWith('selfhst:')) {
        return { type: 'selfhst', value: icon.slice(8) }
    }
    if (icon.startsWith('dashboard:')) {
        return { type: 'dashboard', value: icon.slice(10) }
    }
    return { type: 'tabler', value: icon }
}

export function getSelfhstIconUrl(filename: string): string {
    const ext = filename.match(/\.(\w+)$/)?.[1] || 'webp'
    return `https://cdn.jsdelivr.net/gh/selfhst/icons@main/${ext}/${filename}`
}

export function getDashboardIconUrl(dashboardId: string, filename: string): string {
    const assetDashId = dashboardId.endsWith('-prev') ? dashboardId.slice(0, -5) : dashboardId
    return `/api/v0/dashboards/${assetDashId}/assets/${encodeURIComponent(filename)}`
}

export function isImageIcon(icon: string): boolean {
    return icon.startsWith('selfhst:') || icon.startsWith('dashboard:')
}
