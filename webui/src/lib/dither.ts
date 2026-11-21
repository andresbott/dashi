import type { DitherMode } from '@/types/dashboard'

// The UI value for "no dither field": the server's plain Floyd-Steinberg.
// The draft stores it as an absent field, never as this string.
export const DITHER_PLAIN = 'plain'

export interface DitherOption {
    value: DitherMode | typeof DITHER_PLAIN
    label: string
    hint: string
}

// Mirrors dashboard.DitherModes (internal/dashboard/dither.go), which the
// server validates on save; the presets come from lib/einkimage.
export const DITHER_OPTIONS: DitherOption[] = [
    {
        value: DITHER_PLAIN,
        label: 'Plain',
        hint: 'Floyd-Steinberg against the panel colours with no tone adjustment. The default.',
    },
    {
        value: 'auto',
        label: 'Auto',
        hint: 'Looks at each rendered image and picks the processing for it.',
    },
    {
        value: 'balanced',
        label: 'Balanced',
        hint: 'Compresses the tonal range to what the panel can show.',
    },
    {
        value: 'dynamic',
        label: 'Dynamic',
        hint: 'S-curve contrast for brighter, punchier images.',
    },
    {
        value: 'vivid',
        label: 'Vivid',
        hint: 'Boosts colour. Suits colour panels such as Spectra 6.',
    },
    {
        value: 'soft',
        label: 'Soft',
        hint: 'Lower contrast and smoother diffusion.',
    },
    {
        value: 'grayscale',
        label: 'Grayscale',
        hint: 'Drops colour and keeps edges crisp. Suits black-and-white panels.',
    },
]
