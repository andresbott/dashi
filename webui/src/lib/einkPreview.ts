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

// Custom sizes are clamped to the server's limit, so a typo yields the largest
// allowed render instead of a 400. Keep in sync with maxDisplaySide in
// app/router/middleware_static.go.
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

// Everything the preview page lets the user pick. It lives in the page's
// query string, so a reload or a bookmark shows the same render.
export interface PreviewSettings {
    // A DEVICE_PRESETS key or CUSTOM_DEVICE.
    device: string
    width: number
    height: number
    format: PreviewFormat
    rotation: PreviewRotation
    // 0-based, like the display protocol's page param.
    page: number
    actualSize: boolean
}

// What vue-router hands over as route.query values.
type QueryValue = string | null | (string | null)[] | undefined

const first = (v: QueryValue): string | undefined => {
    const s = Array.isArray(v) ? v[0] : v
    return s ?? undefined
}

const isFormat = (v: string | undefined): v is PreviewFormat =>
    PREVIEW_FORMATS.some(f => f.value === v)

const parseIntParam = (v: string | undefined): number | undefined =>
    v !== undefined && /^\d+$/.test(v) ? Number(v) : undefined

// Reads the settings from the page's query string. Anything missing or
// invalid falls back to the default: the first preset, its format, no
// rotation, page 0. Custom sizes are clamped like the inputs clamp them.
export function parsePreviewQuery(query: Record<string, QueryValue>): PreviewSettings {
    const key = first(query.preset)
    const preset = DEVICE_PRESETS.find(p => p.key === key)
    const fallback = DEVICE_PRESETS[0]
    const custom = key === CUSTOM_DEVICE

    const sized = preset ?? fallback
    const width = custom ? parseIntParam(first(query.width)) : undefined
    const height = custom ? parseIntParam(first(query.height)) : undefined

    const format = first(query.format)
    const rotation = Number(first(query.rotation))
    const actual = first(query.actual)

    return {
        device: preset ? preset.key : custom ? CUSTOM_DEVICE : fallback.key,
        width: width === undefined ? sized.width : clampSide(width),
        height: height === undefined ? sized.height : clampSide(height),
        format: isFormat(format) ? format : sized.format,
        rotation: (PREVIEW_ROTATIONS as readonly number[]).includes(rotation) ? (rotation as PreviewRotation) : 0,
        page: parseIntParam(first(query.page)) ?? 0,
        actualSize: actual === '1' || actual === 'true',
    }
}

// The inverse of parsePreviewQuery. The size is only written for a custom
// device: a preset's size follows from its key.
export function previewQuery(s: PreviewSettings): Record<string, string> {
    const q: Record<string, string> = { preset: s.device }
    if (s.device === CUSTOM_DEVICE) {
        q.width = String(clampSide(s.width))
        q.height = String(clampSide(s.height))
    }
    q.format = s.format
    q.rotation = String(s.rotation)
    q.page = String(s.page)
    if (s.actualSize) q.actual = '1'
    return q
}

export interface UprightView {
    // Size of the image as shown, after turning it upright.
    width: number
    height: number
    // Clockwise CSS degrees that undo the server's rotation.
    angle: number
    // CSS transform, applied with transform-origin at the top-left corner of
    // an image placed at the top-left of a box of width × height. The
    // translate brings the turned image back into that box, so it lands on
    // whole pixels (no -50% centring that could fall on half pixels).
    transform: string
}

// The server lays a 90°/270° page out in portrait and turns it onto the
// panel: RotateImage (internal/dashboard/image/renderer.go) turns 90°
// clockwise and 270° counter-clockwise. The preview turns it back, so it
// reads upright as on the mounted panel. width × height is the image the
// server returned, which is always the requested panel size.
export function uprightView(width: number, height: number, rotation: PreviewRotation): UprightView {
    switch (rotation) {
        case 90:
            return { width: height, height: width, angle: -90, transform: 'rotate(-90deg) translateX(-100%)' }
        case 180:
            return { width, height, angle: 180, transform: 'rotate(180deg) translate(-100%, -100%)' }
        case 270:
            return { width: height, height: width, angle: 90, transform: 'rotate(90deg) translateY(-100%)' }
        default:
            return { width, height, angle: 0, transform: 'none' }
    }
}
