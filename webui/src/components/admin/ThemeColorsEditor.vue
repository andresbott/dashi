<script setup lang="ts">
// ThemeColorsEditor edits one colour mode's palette: a picker and a hex field
// per role. It emits a new palette on each edit (v-model:palette) and never
// mutates the one it was given.
import InputText from 'primevue/inputtext'
import ColorPicker from 'primevue/colorpicker'

import { isValidRoleColor, PALETTE_ROLES } from '@/lib/themeCss'
import type { Palette, PaletteRole } from '@/types/theme'

const props = withDefaults(
    defineProps<{
        // Scopes the field ids (color-<mode>-<role>), so a light and a dark
        // editor can be mounted side by side.
        mode: 'light' | 'dark'
        // Read-only: every control is disabled and nothing is emitted.
        disabled?: boolean
    }>(),
    { disabled: false },
)

const palette = defineModel<Palette>('palette', { required: true })

const roleLabels: Record<PaletteRole, string> = {
    bg: 'Page background',
    surface: 'Card surface',
    border: 'Border',
    text: 'Text',
    textMuted: 'Muted text',
    primary: 'Primary',
    onPrimary: 'Text on primary',
    secondary: 'Secondary',
    success: 'Success',
    warning: 'Warning',
    danger: 'Danger',
}

const roles = PALETTE_ROLES.map(r => r.role)

const fieldId = (role: PaletteRole) => `color-${props.mode}-${role}`

const setRole = (role: PaletteRole, value: string) => {
    if (props.disabled) return
    palette.value = { ...palette.value, [role]: value }
}

// ColorPicker understands only six-digit hex: expand #rgb and drop the alpha
// of #rrggbbaa for display, and keep that alpha when the picker writes back —
// except for bg, which must stay opaque.
const pickerValue = (v: string | undefined): string => {
    const h = (v ?? '').replace('#', '')
    if (/^[0-9a-fA-F]{3}$/.test(h)) return h.split('').map(c => c + c).join('')
    if (/^[0-9a-fA-F]{6}(?:[0-9a-fA-F]{2})?$/.test(h)) return h.slice(0, 6)
    return '000000'
}

const setFromPicker = (role: PaletteRole, v: string | undefined) => {
    if (v === undefined) return
    const prev = palette.value[role] ?? ''
    const alpha = role !== 'bg' && /^#[0-9a-fA-F]{8}$/.test(prev) ? prev.slice(7) : ''
    setRole(role, `#${v}${alpha}`)
}
</script>

<template>
    <div class="flex flex-column gap-2">
        <div class="color-rows">
            <div v-for="role in roles" :key="role" class="color-row">
                <label class="text-sm" :for="fieldId(role)">{{ roleLabels[role] }}</label>
                <div class="flex align-items-center gap-2">
                    <ColorPicker
                        :modelValue="pickerValue(palette[role])"
                        :disabled="disabled"
                        @update:modelValue="(v: string | undefined) => setFromPicker(role, v)"
                    />
                    <InputText
                        :id="fieldId(role)"
                        :modelValue="palette[role] ?? ''"
                        :invalid="!isValidRoleColor(role, palette[role])"
                        :disabled="disabled"
                        class="flex-grow-1 min-w-0"
                        @update:modelValue="(v: string | undefined) => setRole(role, v ?? '')"
                    />
                </div>
            </div>
        </div>
        <small class="colors-hint">
            Hex colours: #rgb, #rrggbb or #rrggbbaa. The page background
            must be opaque.
        </small>
    </div>
</template>

<style scoped lang="scss">
.color-rows {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
    gap: 0.5rem 1rem;
}

.color-row {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
}

.colors-hint {
    color: var(--p-text-muted-color);
}

/* The colour swatch is unbordered by default, so a white or near-white value
   is invisible against the card and there is nothing to click. */
:deep(.p-colorpicker-preview) {
    border: 1px solid var(--p-content-border-color);
    border-radius: 4px;
}
</style>
