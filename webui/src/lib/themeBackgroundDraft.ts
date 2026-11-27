// Temporary adapter between the stored per-mode background (ThemeBackground:
// light and dark each with their own gradient and image) and the shape the
// background editor still edits: one gradient and one image whose dark
// stops/image are optional overrides of the light ones, with direction, fit,
// position and repeat shared. The editor rework that edits each mode directly
// replaces this file.
//
// The mapping is lossy for data the old shape cannot express (a dark-only
// gradient, a plain dark over a light image, a different dark direction or
// fit): toDraft shows the light side, and fromDraft writes the light values
// into dark wherever the draft has no override. The theme editor only writes
// fromDraft back after an edit, so merely opening a theme changes nothing.
import type { ImageFit, ImagePosition, ImageRepeat, ModeBackground, ThemeBackground } from '@/types/theme'

export interface EditorGradient {
    direction: string
    light: string[]
    // Absent: dark uses the light stops.
    dark?: string[]
}

export interface EditorImage {
    light: string
    // Absent: dark uses the light image.
    dark?: string
    fit: ImageFit
    position: ImagePosition
    repeat: ImageRepeat
}

export interface EditorBackground {
    gradient?: EditorGradient
    image?: EditorImage
}

const sameStops = (a: string[], b: string[]): boolean =>
    a.length === b.length && a.every((s, i) => s === b[i])

// toDraft builds the editor shape from the light background, attaching the
// dark stops or image only where they differ from the light ones.
export function toDraft(bg: ThemeBackground | undefined): EditorBackground {
    const light = bg?.light ?? {}
    const dark = bg?.dark ?? {}
    const draft: EditorBackground = {}
    if (light.gradient) {
        draft.gradient = { direction: light.gradient.direction, light: [...light.gradient.stops] }
        const darkStops = dark.gradient?.stops
        if (darkStops && !sameStops(darkStops, light.gradient.stops)) {
            draft.gradient.dark = [...darkStops]
        }
    }
    if (light.image) {
        const { ref, fit, position, repeat } = light.image
        draft.image = { light: ref, fit, position, repeat }
        const darkRef = dark.image?.ref
        if (darkRef && darkRef !== ref) draft.image.dark = darkRef
    }
    return draft
}

const modeFromDraft = (d: EditorBackground, mode: 'light' | 'dark'): ModeBackground => {
    const mb: ModeBackground = {}
    if (d.gradient) {
        const stops = (mode === 'dark' ? d.gradient.dark : undefined) ?? d.gradient.light
        mb.gradient = { direction: d.gradient.direction, stops: [...stops] }
    }
    if (d.image) {
        const ref = (mode === 'dark' ? d.image.dark : undefined) ?? d.image.light
        mb.image = { ref, fit: d.image.fit, position: d.image.position, repeat: d.image.repeat }
    }
    return mb
}

// fromDraft builds the stored background: light from the draft, dark a copy
// of light with the draft's dark overrides applied. Nothing is shared with
// the draft or between the modes.
export function fromDraft(d: EditorBackground): ThemeBackground {
    return { light: modeFromDraft(d, 'light'), dark: modeFromDraft(d, 'dark') }
}
