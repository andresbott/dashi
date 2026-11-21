<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import Card from 'primevue/card'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import Slider from 'primevue/slider'
import SelectButton from 'primevue/selectbutton'
import ColorPicker from 'primevue/colorpicker'
import ToggleSwitch from 'primevue/toggleswitch'
import { useToast } from 'primevue/usetoast'

import { useGetTheme, useUpdateTheme, useThemeAssets } from '@/composables/useThemes'
import { useDataItems } from '@/composables/useDataItems'
import { useAutosave } from '@/composables/useAutosave'
import { browserCss, cssVars, isValidRoleColor, isValidStopColor, PALETTE_ROLES } from '@/lib/themeCss'
import type {
    Theme,
    ImageFit,
    ImagePosition,
    ImageRepeat,
    PaletteRole,
} from '@/types/theme'

const route = useRoute()
const router = useRouter()
const toast = useToast()
const id = computed(() => route.params.id as string)

const { data: serverTheme, isLoading, isError } = useGetTheme(() => id.value)

// Builtins are served read-only (the server answers 403 to writes); the list
// hides their Edit button, but the URL can still be typed.
const isBuiltin = computed(() => serverTheme.value?.builtin === true)
const { updateTheme } = useUpdateTheme()
const { assets, uploadAsset, isUploadingAsset } = useThemeAssets(() => id.value)
const { items: sharedImages } = useDataItems('backgrounds')

// Editing happens on a local clone; autosave persists it after each change.
const local = ref<Theme | null>(null)

watch(serverTheme, (val) => {
    if (val && !local.value) {
        local.value = JSON.parse(JSON.stringify(val))
        // An empty background may arrive as null from hand-edited data.
        if (local.value && !local.value.background) local.value.background = {}
    }
}, { immediate: true })

// One colour mode drives both the colour rows and the preview, so editing the
// dark palette shows the dark preview.
type Mode = 'light' | 'dark'
const mode = ref<Mode>('light')
const modeOptions = [
    { label: 'Light', value: 'light' },
    { label: 'Dark', value: 'dark' },
]

// ---- colours ----

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

// ColorPicker understands only six-digit hex: expand #rgb and drop the alpha
// of #rrggbbaa for display, and keep that alpha when the picker writes back —
// except for bg, which must stay opaque.
const pickerValue = (v: string): string => {
    const h = (v ?? '').replace('#', '')
    if (/^[0-9a-fA-F]{3}$/.test(h)) return h.split('').map(c => c + c).join('')
    if (/^[0-9a-fA-F]{6}(?:[0-9a-fA-F]{2})?$/.test(h)) return h.slice(0, 6)
    return '000000'
}

const setFromPicker = (role: PaletteRole, v: string | undefined) => {
    if (v === undefined || !local.value) return
    const palette = local.value.colors[mode.value]
    const prev = palette[role] ?? ''
    const alpha = role !== 'bg' && /^#[0-9a-fA-F]{8}$/.test(prev) ? prev.slice(7) : ''
    palette[role] = `#${v}${alpha}`
}

// ---- background: palette bg or gradient ----

type BaseKind = 'none' | 'gradient'

const baseKindOptions = [
    { label: 'Palette background', value: 'none' },
    { label: 'Gradient', value: 'gradient' },
]

const baseKind = computed<BaseKind>(() => (local.value?.background.gradient ? 'gradient' : 'none'))

const setBaseKind = (kind: BaseKind) => {
    if (!local.value) return
    if (kind === 'gradient') {
        local.value.background.gradient ??= { direction: 'to bottom right', light: ['#667eea', '#764ba2'] }
    } else {
        delete local.value.background.gradient
    }
}

// ---- dark overrides ----
// An absent `dark` key means "fall back to light". The toggles therefore delete
// the key rather than writing an empty string, which would fail validation.

const gradientDarkEnabled = computed({
    get: () => local.value?.background.gradient?.dark !== undefined,
    set: (on: boolean) => {
        const gradient = local.value?.background.gradient
        if (!gradient) return
        if (on) gradient.dark = [...gradient.light]
        else delete gradient.dark
    },
})

const imageDarkEnabled = computed({
    get: () => local.value?.background.image?.dark !== undefined,
    set: (on: boolean) => {
        const image = local.value?.background.image
        if (!image) return
        if (on) image.dark = image.light
        else delete image.dark
    },
})

// ---- gradient direction ----

// The direction is edited purely as an angle. A stored value may still be one
// of the CSS keywords the backend also accepts, so those map onto their
// equivalent angle for display: 0deg points up and the angle runs clockwise.
const keywordAngles: Record<string, number> = {
    'to top': 0,
    'to top right': 45,
    'to right': 90,
    'to bottom right': 135,
    'to bottom': 180,
    'to bottom left': 225,
    'to left': 270,
    'to top left': 315,
}

const angle = computed({
    get: () => {
        const dir = local.value?.background.gradient?.direction ?? ''
        if (dir in keywordAngles) return keywordAngles[dir]
        return parseInt(dir.replace('deg', ''), 10) || 0
    },
    // Moving the slider converts the direction to degrees. A stored keyword is
    // left untouched until then, so simply opening the editor never rewrites it.
    set: (v: number) => {
        const gradient = local.value?.background.gradient
        if (!gradient) return
        const n = Math.min(360, Math.max(0, Math.round(v || 0)))
        gradient.direction = `${n}deg`
    },
})

const addStop = (stopMode: Mode) => {
    const stops = local.value?.background.gradient?.[stopMode]
    if (stops) stops.push('#ffffff')
}

// A gradient needs at least two stops, so the last two cannot be removed.
const removeStop = (stopMode: Mode, index: number) => {
    const stops = local.value?.background.gradient?.[stopMode]
    if (stops && stops.length > 2) stops.splice(index, 1)
}

// ---- image ----

const fitOptions: { label: string; value: ImageFit }[] = [
    { label: 'Cover', value: 'cover' },
    { label: 'Contain', value: 'contain' },
    { label: 'Stretch', value: 'stretch' },
    { label: 'Original size', value: 'original' },
]

const positionOptions: { label: string; value: ImagePosition }[] = [
    { label: 'Center', value: 'center' },
    { label: 'Top', value: 'top' },
    { label: 'Bottom', value: 'bottom' },
    { label: 'Left', value: 'left' },
    { label: 'Right', value: 'right' },
    { label: 'Top left', value: 'top left' },
    { label: 'Top right', value: 'top right' },
    { label: 'Bottom left', value: 'bottom left' },
    { label: 'Bottom right', value: 'bottom right' },
]

const repeatOptions: { label: string; value: ImageRepeat }[] = [
    { label: 'No repeat', value: 'no-repeat' },
    { label: 'Tile', value: 'repeat' },
    { label: 'Tile horizontally', value: 'repeat-x' },
    { label: 'Tile vertically', value: 'repeat-y' },
]

const imageEnabled = computed({
    get: () => local.value?.background.image !== undefined,
    set: (on: boolean) => {
        if (!local.value) return
        if (on) {
            local.value.background.image = {
                light: '',
                fit: 'cover',
                position: 'center',
                repeat: 'no-repeat',
            }
        } else {
            delete local.value.background.image
        }
    },
})

// Two sources, grouped: this theme's own uploads (asset:) and the
// admin-curated shared pool (shared:).
const imageOptions = computed(() => {
    const groups: { label: string; items: { label: string; value: string }[] }[] = []
    const own = assets.value ?? []
    if (own.length) {
        groups.push({
            label: 'This theme',
            items: own.map(a => ({ label: a, value: `asset:${a}` })),
        })
    }
    const shared = sharedImages.value ?? []
    if (shared.length) {
        groups.push({
            label: 'Shared images',
            items: shared.map(i => ({ label: i.name, value: `shared:${i.name}` })),
        })
    }
    return groups
})

const uploadInput = ref<HTMLInputElement | null>(null)

const handleUpload = async (event: Event) => {
    const input = event.target as HTMLInputElement
    const file = input.files?.[0]
    if (!file) return
    try {
        await uploadAsset({ name: file.name, bytes: await file.arrayBuffer() })
        // Select what was just uploaded — that is invariably why it was uploaded.
        if (local.value?.background.image) local.value.background.image.light = `asset:${file.name}`
        toast.add({ severity: 'success', summary: 'Uploaded', detail: file.name, life: 3000 })
    } catch {
        toast.add({ severity: 'error', summary: 'Error', detail: 'Failed to upload image', life: 5000 })
    } finally {
        if (uploadInput.value) uploadInput.value.value = ''
    }
}

// ---- preview ----

// Generated by the same logic the server uses, so the preview shows what the
// dashboard will actually render — no save and no round-trip required. The
// variables are applied inline to the preview box only, never to :root, so
// the admin UI around it keeps its own styling.
const previewStyle = computed(() => {
    if (!local.value) return {}
    return {
        ...cssVars(browserCss(local.value), mode.value),
        background: 'var(--dashi-page-bg, var(--dashi-bg))',
    }
})

const previewChips: { role: PaletteRole; label: string }[] = [
    { role: 'primary', label: 'Primary' },
    { role: 'secondary', label: 'Secondary' },
    { role: 'success', label: 'Success' },
    { role: 'warning', label: 'Warning' },
    { role: 'danger', label: 'Danger' },
]

// ---- validity ----
// Mirrors the server's Validate for the fields the editor lets you break, so
// a half-typed value is highlighted instead of autosaved into a 400.

const nameInvalid = computed(() => !local.value?.name.trim())
const imageLightInvalid = computed(() => local.value?.background.image?.light === '')

const isValid = computed(() => {
    const t = local.value
    if (!t || nameInvalid.value || imageLightInvalid.value) return false
    for (const m of ['light', 'dark'] as const) {
        if (!roles.every(role => isValidRoleColor(role, t.colors[m][role]))) return false
    }
    const g = t.background.gradient
    if (g && ![...g.light, ...(g.dark ?? [])].every(isValidStopColor)) return false
    return true
})

// ---- autosave ----

// Invalid states are neither scheduled nor sent: a save timer started by an
// earlier valid edit is skipped too. The edit that makes the theme valid
// again schedules a save of the whole theme.
const { status: saveStatus, error: saveError, flush } = useAutosave<Theme>({
    source: () => local.value,
    save: (theme) => (isValid.value ? updateTheme({ id: id.value, payload: theme }) : Promise.resolve()),
    enabled: () => !isBuiltin.value && isValid.value,
})

// A 400 carries the store's validation message (it names the offending field),
// so surface that rather than a generic failure — same as the old manual Save.
watch(saveStatus, (s) => {
    if (s !== 'error') return
    const detail =
        (saveError.value as { response?: { data?: { error?: string } } })?.response?.data?.error ??
        'Failed to save theme'
    toast.add({ severity: 'error', summary: 'Error', detail, life: 6000 })
})

const goBack = async () => {
    await flush()
    router.push({ name: 'admin-themes' })
}
</script>

<template>
    <div class="admin-section">
        <div class="admin-section-header">
            <h2 class="admin-section-title">Edit theme</h2>
            <div class="flex align-items-center gap-3">
                <Button
                    label="Back"
                    icon="ti ti-arrow-left"
                    severity="secondary"
                    @click="goBack"
                />
            </div>
        </div>

        <Card v-if="isLoading">
            <template #content><div class="info-message">Loading…</div></template>
        </Card>

        <Card v-else-if="isError">
            <template #content><div class="info-message">Theme not found.</div></template>
        </Card>

        <Card v-else-if="isBuiltin">
            <template #content>
                <div class="info-message">
                    Built-in themes cannot be edited. Copy it from the Themes list
                    and edit the copy.
                </div>
            </template>
        </Card>

        <template v-else-if="local">
            <Card>
                <template #content>
                    <div class="edit-grid">
                        <div class="edit-fields">
                            <div class="flex flex-column gap-1">
                                <label class="font-semibold text-sm">Name</label>
                                <InputText v-model="local.name" :invalid="nameInvalid" class="w-full" />
                            </div>

                            <!-- Colours: one palette per colour mode -->
                            <label class="edit-subsection">Colours</label>
                            <SelectButton
                                v-model="mode"
                                :options="modeOptions"
                                optionLabel="label"
                                optionValue="value"
                                :allowEmpty="false"
                            />
                            <div class="color-rows">
                                <div v-for="role in roles" :key="role" class="color-row">
                                    <label class="text-sm" :for="`color-${role}`">{{ roleLabels[role] }}</label>
                                    <div class="flex align-items-center gap-2">
                                        <ColorPicker
                                            :modelValue="pickerValue(local.colors[mode][role])"
                                            @update:modelValue="(v: string | undefined) => setFromPicker(role, v)"
                                        />
                                        <InputText
                                            :id="`color-${role}`"
                                            v-model="local.colors[mode][role]"
                                            :invalid="!isValidRoleColor(role, local.colors[mode][role])"
                                            class="flex-grow-1"
                                        />
                                    </div>
                                </div>
                            </div>
                            <small class="edit-hint">
                                Hex colours: #rgb, #rrggbb or #rrggbbaa. The page background
                                must be opaque.
                            </small>

                            <!-- Background: palette bg or gradient, optionally an image over it -->
                            <label class="edit-subsection">Background</label>
                            <SelectButton
                                :modelValue="baseKind"
                                :options="baseKindOptions"
                                optionLabel="label"
                                optionValue="value"
                                :allowEmpty="false"
                                @update:modelValue="(v: BaseKind) => setBaseKind(v)"
                            />

                            <template v-if="local.background.gradient">
                                <div class="flex flex-column gap-1">
                                    <div class="flex align-items-center justify-content-between">
                                        <label class="font-semibold text-sm">Angle</label>
                                        <span class="angle-readout">{{ angle }}°</span>
                                    </div>
                                    <Slider
                                        :modelValue="angle"
                                        @update:modelValue="(v: number | number[]) => { angle = Array.isArray(v) ? v[0] : v }"
                                        :min="0"
                                        :max="360"
                                        :step="5"
                                        class="angle-slider"
                                    />
                                    <small class="edit-hint">0° points up, increasing clockwise.</small>
                                </div>

                                <div class="flex flex-column gap-1">
                                    <label class="font-semibold text-sm">Colour stops</label>
                                    <div
                                        v-for="(_stop, i) in local.background.gradient.light"
                                        :key="'l' + i"
                                        class="flex align-items-center gap-2"
                                    >
                                        <ColorPicker
                                            :modelValue="pickerValue(local.background.gradient.light[i])"
                                            @update:modelValue="(v: string | undefined) => { if (v !== undefined && local?.background.gradient) local.background.gradient.light[i] = '#' + v }"
                                        />
                                        <InputText
                                        v-model="local.background.gradient.light[i]"
                                        :invalid="!isValidStopColor(local.background.gradient.light[i])"
                                        :data-testid="`stop-light-${i}`"
                                        class="flex-grow-1"
                                    />
                                        <Button
                                            icon="ti ti-trash"
                                            text
                                            rounded
                                            severity="danger"
                                            :disabled="local.background.gradient.light.length <= 2"
                                            @click="removeStop('light', i)"
                                        />
                                    </div>
                                    <Button
                                        label="Add stop"
                                        icon="ti ti-plus"
                                        text
                                        size="small"
                                        @click="addStop('light')"
                                    />
                                </div>

                                <div class="flex align-items-center gap-2">
                                    <ToggleSwitch v-model="gradientDarkEnabled" />
                                    <span class="text-sm">Different stops in dark mode</span>
                                </div>
                                <div v-if="local.background.gradient.dark" class="flex flex-column gap-1">
                                    <label class="font-semibold text-sm">Dark stops</label>
                                    <div
                                        v-for="(_stop, i) in local.background.gradient.dark"
                                        :key="'d' + i"
                                        class="flex align-items-center gap-2"
                                    >
                                        <ColorPicker
                                            :modelValue="pickerValue(local.background.gradient.dark[i])"
                                            @update:modelValue="(v: string | undefined) => { if (v !== undefined && local?.background.gradient?.dark) local.background.gradient.dark[i] = '#' + v }"
                                        />
                                        <InputText
                                            v-model="local.background.gradient.dark[i]"
                                            :invalid="!isValidStopColor(local.background.gradient.dark[i])"
                                            :data-testid="`stop-dark-${i}`"
                                            class="flex-grow-1"
                                        />
                                        <Button
                                            icon="ti ti-trash"
                                            text
                                            rounded
                                            severity="danger"
                                            :disabled="local.background.gradient.dark.length <= 2"
                                            @click="removeStop('dark', i)"
                                        />
                                    </div>
                                    <Button label="Add stop" icon="ti ti-plus" text size="small" @click="addStop('dark')" />
                                </div>
                            </template>

                            <!-- Image layer, painted over the base -->
                            <div class="flex align-items-center gap-2">
                                <ToggleSwitch v-model="imageEnabled" data-testid="image-toggle" />
                                <span class="text-sm">Use an image</span>
                            </div>

                            <template v-if="local.background.image">
                                <div class="flex flex-column gap-1">
                                    <label class="font-semibold text-sm">Image</label>
                                    <div class="flex gap-2">
                                        <Select
                                            v-model="local.background.image.light"
                                            :invalid="imageLightInvalid"
                                            :options="imageOptions"
                                            optionLabel="label"
                                            optionValue="value"
                                            optionGroupLabel="label"
                                            optionGroupChildren="items"
                                            placeholder="Select an image…"
                                            class="flex-grow-1"
                                        />
                                        <Button
                                            label="Upload"
                                            icon="ti ti-upload"
                                            severity="secondary"
                                            :loading="isUploadingAsset"
                                            @click="uploadInput?.click()"
                                        />
                                        <input
                                            ref="uploadInput"
                                            type="file"
                                            accept="image/*"
                                            style="display: none"
                                            @change="handleUpload"
                                        />
                                    </div>
                                </div>

                                <div class="flex align-items-center gap-2">
                                    <ToggleSwitch v-model="imageDarkEnabled" />
                                    <span class="text-sm">Different image in dark mode</span>
                                </div>
                                <div v-if="local.background.image.dark !== undefined" class="flex flex-column gap-1">
                                    <label class="font-semibold text-sm">Dark image</label>
                                    <Select
                                        v-model="local.background.image.dark"
                                        :options="imageOptions"
                                        optionLabel="label"
                                        optionValue="value"
                                        optionGroupLabel="label"
                                        optionGroupChildren="items"
                                        placeholder="Select an image…"
                                        class="w-full"
                                    />
                                </div>

                                <div class="edit-row">
                                    <div class="flex flex-column gap-1 flex-grow-1">
                                        <label class="font-semibold text-sm">Fit</label>
                                        <Select v-model="local.background.image.fit" :options="fitOptions" optionLabel="label" optionValue="value" />
                                    </div>
                                    <div class="flex flex-column gap-1 flex-grow-1">
                                        <label class="font-semibold text-sm">Position</label>
                                        <Select v-model="local.background.image.position" :options="positionOptions" optionLabel="label" optionValue="value" />
                                    </div>
                                    <div class="flex flex-column gap-1 flex-grow-1">
                                        <label class="font-semibold text-sm">Repeat</label>
                                        <Select v-model="local.background.image.repeat" :options="repeatOptions" optionLabel="label" optionValue="value" />
                                    </div>
                                </div>
                                <small class="edit-hint">
                                    Fit, position and repeat apply in browsers only. E-ink and
                                    image dashboards always cover the canvas.
                                </small>
                            </template>
                        </div>

                        <!-- Live preview, scoped to the box -->
                        <div class="edit-preview">
                            <div class="flex align-items-center justify-content-between">
                                <label class="font-semibold text-sm">Preview</label>
                                <SelectButton
                                    v-model="mode"
                                    :options="modeOptions"
                                    optionLabel="label"
                                    optionValue="value"
                                    :allowEmpty="false"
                                    size="small"
                                />
                            </div>
                            <div class="preview-box" :style="previewStyle">
                                <div class="preview-card">
                                    <div class="preview-title">Card title</div>
                                    <div class="preview-muted">Muted secondary text</div>
                                    <div class="preview-chips">
                                        <span
                                            v-for="chip in previewChips"
                                            :key="chip.role"
                                            class="preview-chip"
                                            :class="{ 'preview-chip-primary': chip.role === 'primary' }"
                                            :style="{ background: `var(--dashi-${chip.role})` }"
                                        >
                                            {{ chip.label }}
                                        </span>
                                    </div>
                                </div>
                            </div>
                            <small class="edit-hint">
                                <template v-if="isValid">Changes are saved automatically.</template>
                                <template v-else>Fix the highlighted fields to save your changes.</template>
                            </small>
                        </div>
                    </div>
                </template>
            </Card>
        </template>
    </div>
</template>

<style scoped>
.admin-section {
    display: flex;
    flex-direction: column;
    gap: 1rem;
}

.admin-section-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 1rem;
}

.admin-section-title {
    font-size: 1.5rem;
    font-weight: 700;
    color: var(--p-text-color);
    margin: 0;
}

.info-message {
    padding: 1rem;
    text-align: center;
    color: var(--p-text-muted-color);
}

.edit-grid {
    display: flex;
    gap: 2rem;
    align-items: flex-start;
}

.edit-fields {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    flex: 1;
    min-width: 0;
}

.edit-preview {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    width: 320px;
    position: sticky;
    top: 1rem;
}

.preview-box {
    height: 220px;
    border-radius: 6px;
    border: 1px solid var(--p-surface-border);
    background-size: cover;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 1rem;
}

.preview-card {
    width: 100%;
    background: var(--dashi-surface);
    border: 1px solid var(--dashi-border);
    border-radius: 6px;
    padding: 0.75rem;
    color: var(--dashi-text);
}

.preview-title {
    font-weight: 600;
}

.preview-muted {
    color: var(--dashi-text-muted);
    font-size: 0.85rem;
    margin-bottom: 0.5rem;
}

.preview-chips {
    display: flex;
    flex-wrap: wrap;
    gap: 0.25rem;
}

.preview-chip {
    font-size: 0.7rem;
    padding: 0.1rem 0.4rem;
    border-radius: 999px;
    color: #ffffff;
}

.preview-chip-primary {
    color: var(--dashi-on-primary);
}

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

.edit-row {
    display: flex;
    gap: 0.75rem;
}

.edit-subsection {
    margin-top: 0.5rem;
    font-size: 0.75rem;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--p-text-muted-color);
}

.edit-hint {
    color: var(--p-text-muted-color);
}

/* A slider has no visible value, so the angle is shown beside its label. */
.angle-readout {
    font-variant-numeric: tabular-nums;
    color: var(--p-text-muted-color);
    font-size: 0.875rem;
}

.angle-slider {
    margin: 0.5rem 0.25rem;
}

/* The colour swatch is unbordered by default, so a white or near-white value
   is invisible against the card and there is nothing to click. */
:deep(.p-colorpicker-preview) {
    border: 1px solid var(--p-surface-border);
    border-radius: 4px;
}

@media (max-width: 900px) {
    .edit-grid {
        flex-direction: column;
    }

    .edit-preview {
        width: 100%;
        position: static;
    }
}
</style>
