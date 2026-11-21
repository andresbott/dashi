import { withBase } from '@/lib/base'

// Formats the e-ink preview can show. The packed wire formats (bw, spectra6)
// are raw panel bytes, not images, so they are left out; png-bw and
// png-spectra6 dither exactly like them.
export type PreviewFormat = 'png-bw' | 'png-spectra6' | 'png'

export const PREVIEW_FORMATS: { value: PreviewFormat; label: string }[] = [
    { value: 'png-bw', label: 'Black & white (png-bw)' },
    { value: 'png-spectra6', label: 'Spectra 6 colour (png-spectra6)' },
    { value: 'png', label: 'Full colour, no dither (png)' },
]

// The display protocol accepts only these (validRotations in
// app/router/middleware_static.go).
export const PREVIEW_ROTATIONS = [0, 90, 180, 270] as const
export type PreviewRotation = (typeof PREVIEW_ROTATIONS)[number]

export interface DevicePreset {
    key: string
    label: string
    width: number
    height: number
    format: PreviewFormat
}

export const CUSTOM_DEVICE = 'custom'

// Common e-ink panels, in native (landscape) orientation. Dashboards store no
// display size; these only seed the preview controls.
export const DEVICE_PRESETS: DevicePreset[] = [
    { key: 'bw-7in5', label: '7.5" BW 800×480', width: 800, height: 480, format: 'png-bw' },
    { key: 'bw-7in5-v1', label: '7.5" BW 640×384', width: 640, height: 384, format: 'png-bw' },
    { key: 's6-7in3', label: '7.3" Spectra 6 800×480', width: 800, height: 480, format: 'png-spectra6' },
    { key: 'bw-4in2', label: '4.2" BW 400×300', width: 400, height: 300, format: 'png-bw' },
    { key: 'bw-2in9', label: '2.9" BW 296×128', width: 296, height: 128, format: 'png-bw' },
    { key: 's6-13in3', label: '13.3" Spectra 6 1600×1200', width: 1600, height: 1200, format: 'png-spectra6' },
]

// Custom sizes are clamped to this: the server renders whatever it is asked
// for, and a typo should not ask it for a gigapixel image.
export const MAX_PREVIEW_SIDE = 4096

export interface PreviewParams {
    dashboardId: string
    width: number
    height: number
    format: PreviewFormat
    rotation: PreviewRotation
    page: number
}

// The editor server renders the PNG itself: its dashboard middleware answers
// GET /{id} with display query params exactly like the viewer, but without
// the viewer's per-dashboard basic auth. withBase keeps the path under the
// Home Assistant ingress prefix.
export function einkPreviewUrl(p: PreviewParams): string {
    const q = new URLSearchParams({
        format: p.format,
        width: String(p.width),
        height: String(p.height),
        rotation: String(p.rotation),
        page: String(p.page),
    })
    return withBase(`/${encodeURIComponent(p.dashboardId)}?${q.toString()}`)
}

export function clampSide(v: number | null | undefined): number {
    const n = Math.floor(Number(v) || 0)
    return Math.min(MAX_PREVIEW_SIDE, Math.max(1, n))
}
