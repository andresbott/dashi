<script setup lang="ts">
// The Background section of the theme editor. A background belongs to its
// theme (there is no separate background entity): a base — the palette `bg`
// or a linear gradient — with an optional image painted over it, each with an
// optional dark-mode variant. Edits mutate the bound background in place, so
// the parent's autosave picks them up like any other theme field.
import { computed, ref } from 'vue'
import Button from 'primevue/button'
import ColorPicker from 'primevue/colorpicker'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import SelectButton from 'primevue/selectbutton'
import Slider from 'primevue/slider'
import ToggleSwitch from 'primevue/toggleswitch'
import { useToast } from 'primevue/usetoast'

import { useThemeAssets } from '@/composables/useThemes'
import { useDataItems } from '@/composables/useDataItems'
import { dataUrl } from '@/lib/api/data'
import { themeAssetUrl } from '@/lib/api/themes'
import { isValidStopColor, previewStyle } from '@/lib/themeCss'
import type {
    ImageFit,
    ImagePosition,
    ImageRepeat,
    PaletteRole,
    ThemeBackground,
    ThemeColors,
} from '@/types/theme'

const props = withDefaults(
    defineProps<{
        themeId: string
        // The theme's palettes, for the preview only.
        colors: ThemeColors
        // Read-only: every control is disabled and nothing is written.
        disabled?: boolean
    }>(),
    { disabled: false },
)

const background = defineModel<ThemeBackground>('background', { required: true })

const toast = useToast()
const { assets, uploadAsset, isUploadingAsset } = useThemeAssets(() => props.themeId)
const { items: poolItems } = useDataItems('backgrounds')

type Mode = 'light' | 'dark'

// ---- base: palette bg or gradient ----

type BaseKind = 'none' | 'gradient'

const baseKindOptions = [
    { label: 'Palette colour', value: 'none' },
    { label: 'Gradient', value: 'gradient' },
]

const baseKind = computed<BaseKind>(() => (background.value.gradient ? 'gradient' : 'none'))

const setBaseKind = (kind: BaseKind) => {
    if (props.disabled) return
    if (kind === 'gradient') {
        background.value.gradient ??= { direction: '135deg', light: ['#667eea', '#764ba2'] }
    } else {
        delete background.value.gradient
    }
}

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

const angle = computed(() => {
    const dir = background.value.gradient?.direction ?? ''
    if (dir in keywordAngles) return keywordAngles[dir]
    return parseInt(dir.replace('deg', ''), 10) || 0
})

// Moving the slider converts the direction to degrees. A stored keyword is
// left untouched until then, so simply opening the editor never rewrites it.
const setAngle = (v: number | number[]) => {
    const gradient = background.value.gradient
    if (!gradient || props.disabled) return
    const n = Math.min(360, Math.max(0, Math.round((Array.isArray(v) ? v[0] : v) || 0)))
    gradient.direction = `${n}deg`
}

// ---- gradient stops ----
// An absent `dark` list means "use the light stops". The toggle deletes the
// key rather than leaving an empty list behind.

const gradientDarkEnabled = computed(() => background.value.gradient?.dark !== undefined)

const setGradientDark = (on: boolean) => {
    const gradient = background.value.gradient
    if (!gradient || props.disabled) return
    if (on) gradient.dark = [...gradient.light]
    else delete gradient.dark
}

const stopLists = computed(() => {
    const g = background.value.gradient
    if (!g) return []
    const lists: { mode: Mode; label: string; stops: string[] }[] = [
        { mode: 'light', label: g.dark ? 'Light stops' : 'Colour stops', stops: g.light },
    ]
    if (g.dark) lists.push({ mode: 'dark', label: 'Dark stops', stops: g.dark })
    return lists
})

const stopsOf = (mode: Mode): string[] | undefined => background.value.gradient?.[mode]

// A new stop repeats the last colour, so adding one never jumps the gradient.
const addStop = (mode: Mode) => {
    const stops = stopsOf(mode)
    if (!stops || props.disabled) return
    stops.push(stops[stops.length - 1] ?? '#ffffff')
}

// A gradient needs at least two stops, so the last two cannot be removed.
const removeStop = (mode: Mode, index: number) => {
    const stops = stopsOf(mode)
    if (!stops || props.disabled || stops.length <= 2) return
    stops.splice(index, 1)
}

const moveStop = (mode: Mode, index: number, delta: -1 | 1) => {
    const stops = stopsOf(mode)
    const target = index + delta
    if (!stops || props.disabled || target < 0 || target >= stops.length) return
    ;[stops[index], stops[target]] = [stops[target], stops[index]]
}

const setStop = (mode: Mode, index: number, value: string) => {
    const stops = stopsOf(mode)
    if (!stops || props.disabled) return
    stops[index] = value
}

// ColorPicker understands only six-digit hex: expand #rgb and drop the alpha
// of #rrggbbaa for display, and keep that alpha when the picker writes back.
const pickerValue = (v: string): string => {
    const h = (v ?? '').replace('#', '')
    if (/^[0-9a-fA-F]{3}$/.test(h)) return h.split('').map(c => c + c).join('')
    if (/^[0-9a-fA-F]{6}(?:[0-9a-fA-F]{2})?$/.test(h)) return h.slice(0, 6)
    return '000000'
}

const setStopFromPicker = (mode: Mode, index: number, v: string | undefined) => {
    const prev = stopsOf(mode)?.[index] ?? ''
    if (v === undefined) return
    const alpha = /^#[0-9a-fA-F]{8}$/.test(prev) ? prev.slice(7) : ''
    setStop(mode, index, `#${v}${alpha}`)
}

// The strip above each stop list shows the stops left to right over a
// checkerboard, so translucent stops are visible as such.
const stripStyle = (stops: string[]) => {
    const valid = stops.filter(isValidStopColor)
    const checker = 'repeating-conic-gradient(#d4d4d8 0 25%, #ffffff 0 50%) 0 0/12px 12px'
    if (valid.length < 2) return { background: checker }
    return { background: `linear-gradient(90deg,${valid.join(',')}), ${checker}` }
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

// The extensions internal/themes ValidateRef accepts: anything else would be
// rejected on save, so it is neither offered nor uploadable.
const IMAGE_EXT = /\.(?:png|jpe?g|webp|svg)$/i
const UPLOAD_ACCEPT = '.png,.jpg,.jpeg,.webp,.svg'

interface ImageChoice {
    ref: string
    label: string
    url: string
}

// Two sources: this theme's own uploads (asset:, copied along with the theme)
// and the shared background-images pool (shared:).
const ownImages = computed<ImageChoice[]>(() =>
    (assets.value ?? [])
        .filter(a => IMAGE_EXT.test(a))
        .map(a => ({ ref: `asset:${a}`, label: a, url: themeAssetUrl(props.themeId, a) })),
)

const poolImages = computed<ImageChoice[]>(() =>
    (poolItems.value ?? [])
        .filter(i => IMAGE_EXT.test(i.name))
        .map(i => ({ ref: `shared:${i.name}`, label: i.name, url: dataUrl('backgrounds', i.name) })),
)

const imageGroups = computed(() =>
    [
        { key: 'own', label: 'This theme', items: ownImages.value },
        { key: 'pool', label: 'Shared background images', items: poolImages.value },
    ].filter(g => g.items.length),
)

const knownRefs = computed(() => new Set([...ownImages.value, ...poolImages.value].map(c => c.ref)))

// A reference to an image that was deleted after being picked: still valid to
// save, but it renders nothing, so it is called out once both lists loaded.
const isMissing = (ref: string | undefined): boolean =>
    !!ref && assets.value !== undefined && poolItems.value !== undefined && !knownRefs.value.has(ref)

const refLabel = (ref: string): string => ref.slice(ref.indexOf(':') + 1)

const imageSlots = computed(() => {
    const img = background.value.image
    const slots: { mode: Mode; label: string; selected: string }[] = [
        { mode: 'light', label: img?.dark !== undefined ? 'Light mode' : 'Both modes', selected: img?.light ?? '' },
    ]
    if (img?.dark !== undefined) slots.push({ mode: 'dark', label: 'Dark mode', selected: img.dark })
    return slots
})

const selectImage = (mode: Mode, ref: string) => {
    if (props.disabled) return
    const img = background.value.image
    if (mode === 'dark') {
        if (img) img.dark = ref
        return
    }
    if (img) img.light = ref
    else background.value.image = { light: ref, fit: 'cover', position: 'center', repeat: 'no-repeat' }
}

const clearImage = () => {
    if (props.disabled) return
    delete background.value.image
}

const imageDarkEnabled = computed(() => background.value.image?.dark !== undefined)

const setImageDark = (on: boolean) => {
    const img = background.value.image
    if (!img || props.disabled) return
    if (on) img.dark = img.light
    else delete img.dark
}

const uploadInput = ref<HTMLInputElement | null>(null)
const uploadTarget = ref<Mode>('light')

const pickUpload = (mode: Mode) => {
    if (props.disabled) return
    uploadTarget.value = mode
    uploadInput.value?.click()
}

// Uploads go into this theme's own assets, so the theme stays self-contained
// (a copy of the theme carries them along), and select what was just uploaded.
const handleUpload = async (event: Event) => {
    const input = event.target as HTMLInputElement
    const file = input.files?.[0]
    if (!file) return
    try {
        await uploadAsset({ name: file.name, bytes: await file.arrayBuffer() })
        selectImage(uploadTarget.value, `asset:${file.name}`)
        toast.add({ severity: 'success', summary: 'Uploaded', detail: file.name, life: 3000 })
    } catch {
        toast.add({ severity: 'error', summary: 'Error', detail: 'Failed to upload image', life: 5000 })
    } finally {
        input.value = ''
    }
}

// ---- preview ----

// Both colour modes side by side, generated by the same code the viewer's
// stylesheet mirrors, so what you see is what the dashboard renders.
const previewTiles = computed(() =>
    (['light', 'dark'] as const).map(mode => ({
        mode,
        label: mode === 'light' ? 'Light' : 'Dark',
        style: previewStyle(
            { id: props.themeId, name: '', background: background.value, colors: props.colors },
            mode,
        ),
    })),
)

const swatches: PaletteRole[] = ['primary', 'secondary', 'success', 'warning', 'danger']
</script>

<template>
    <div class="bg-editor">
        <!-- Live preview: light and dark side by side -->
        <div class="bg-preview">
            <div
                v-for="tile in previewTiles"
                :key="tile.mode"
                class="bg-preview-tile"
                :style="tile.style"
                :data-testid="`bg-preview-${tile.mode}`"
            >
                <span class="bg-preview-mode">{{ tile.label }}</span>
                <div class="bg-preview-card">
                    <div class="bg-preview-text">Aa Sample text</div>
                    <div class="bg-preview-muted">Muted text</div>
                    <div class="bg-preview-swatches">
                        <span
                            v-for="role in swatches"
                            :key="role"
                            class="bg-preview-swatch"
                            :style="{ background: `var(--dashi-${role})` }"
                        />
                    </div>
                </div>
            </div>
        </div>

        <!-- Base layer -->
        <div class="flex flex-column gap-1">
            <label class="font-semibold text-sm">Base</label>
            <SelectButton
                :modelValue="baseKind"
                :options="baseKindOptions"
                optionLabel="label"
                optionValue="value"
                :allowEmpty="false"
                :disabled="disabled"
                data-testid="bg-base"
                @update:modelValue="(v: BaseKind) => setBaseKind(v)"
            />
            <small v-if="baseKind === 'none'" class="edit-hint">
                The page shows the palette's page background colour.
            </small>
        </div>

        <template v-if="background.gradient">
            <div class="flex flex-column gap-1">
                <div class="flex align-items-center justify-content-between">
                    <label class="font-semibold text-sm">Angle</label>
                    <span class="angle-readout" data-testid="bg-angle">{{ angle }}°</span>
                </div>
                <Slider
                    :modelValue="angle"
                    :min="0"
                    :max="360"
                    :step="5"
                    :disabled="disabled"
                    class="angle-slider"
                    @update:modelValue="setAngle"
                />
                <small class="edit-hint">0° points up, increasing clockwise.</small>
            </div>

            <div v-for="list in stopLists" :key="list.mode" class="flex flex-column gap-1">
                <label class="font-semibold text-sm">{{ list.label }}</label>
                <div class="stop-strip" :style="stripStyle(list.stops)" />
                <div
                    v-for="(stop, i) in list.stops"
                    :key="`${list.mode}-${i}`"
                    class="flex align-items-center gap-2"
                >
                    <ColorPicker
                        :modelValue="pickerValue(stop)"
                        :disabled="disabled"
                        @update:modelValue="(v: string | undefined) => setStopFromPicker(list.mode, i, v)"
                    />
                    <InputText
                        :modelValue="stop"
                        :invalid="!isValidStopColor(stop)"
                        :disabled="disabled"
                        :data-testid="`stop-${list.mode}-${i}`"
                        class="flex-grow-1 min-w-0"
                        @update:modelValue="(v: string | undefined) => setStop(list.mode, i, v ?? '')"
                    />
                    <Button
                        icon="ti ti-arrow-up"
                        text
                        rounded
                        severity="secondary"
                        aria-label="Move stop up"
                        :disabled="disabled || i === 0"
                        :data-testid="`stop-${list.mode}-${i}-up`"
                        @click="moveStop(list.mode, i, -1)"
                    />
                    <Button
                        icon="ti ti-arrow-down"
                        text
                        rounded
                        severity="secondary"
                        aria-label="Move stop down"
                        :disabled="disabled || i === list.stops.length - 1"
                        :data-testid="`stop-${list.mode}-${i}-down`"
                        @click="moveStop(list.mode, i, 1)"
                    />
                    <Button
                        icon="ti ti-trash"
                        text
                        rounded
                        severity="danger"
                        aria-label="Remove stop"
                        :disabled="disabled || list.stops.length <= 2"
                        :data-testid="`stop-${list.mode}-${i}-remove`"
                        @click="removeStop(list.mode, i)"
                    />
                </div>
                <Button
                    label="Add stop"
                    icon="ti ti-plus"
                    text
                    size="small"
                    class="align-self-start"
                    :disabled="disabled"
                    :data-testid="`stop-${list.mode}-add`"
                    @click="addStop(list.mode)"
                />
            </div>
            <small class="edit-hint">
                Stops are spread evenly. Hex colours, alpha allowed: e-ink and image
                dashboards flatten translucent stops over the page background.
            </small>

            <div class="flex align-items-center gap-2">
                <ToggleSwitch
                    :modelValue="gradientDarkEnabled"
                    :disabled="disabled"
                    data-testid="gradient-dark-toggle"
                    @update:modelValue="setGradientDark"
                />
                <span class="text-sm">Different stops in dark mode</span>
            </div>
        </template>

        <!-- Image layer, painted over the base -->
        <div class="flex flex-wrap align-items-center justify-content-between gap-2 mt-2">
            <label class="font-semibold text-sm">Image</label>
            <div class="flex flex-wrap align-items-center gap-2">
                <router-link :to="{ name: 'admin-background-images' }" class="pool-link text-sm">
                    Manage shared images
                </router-link>
                <Button
                    label="Remove image"
                    icon="ti ti-x"
                    severity="secondary"
                    size="small"
                    text
                    :disabled="disabled || !background.image"
                    data-testid="image-clear"
                    @click="clearImage"
                />
            </div>
        </div>

        <div v-if="background.image" class="flex align-items-center gap-2">
            <ToggleSwitch
                :modelValue="imageDarkEnabled"
                :disabled="disabled"
                data-testid="image-dark-toggle"
                @update:modelValue="setImageDark"
            />
            <span class="text-sm">Different image in dark mode</span>
        </div>

        <div v-for="slot in imageSlots" :key="slot.mode" class="flex flex-column gap-1">
            <div class="flex align-items-center justify-content-between">
                <span class="text-sm font-semibold">{{ slot.label }}</span>
                <Button
                    label="Upload"
                    icon="ti ti-upload"
                    severity="secondary"
                    size="small"
                    :loading="isUploadingAsset && uploadTarget === slot.mode"
                    :disabled="disabled"
                    :data-testid="`image-upload-${slot.mode}`"
                    @click="pickUpload(slot.mode)"
                />
            </div>
            <small v-if="background.image && slot.mode === 'light' && !slot.selected" class="image-warning">
                Pick an image, or remove the image layer.
            </small>
            <small v-else-if="isMissing(slot.selected)" class="image-warning">
                “{{ refLabel(slot.selected) }}” no longer exists. Pick another image.
            </small>
            <div v-if="!imageGroups.length" class="edit-hint text-sm">
                No images yet. Upload one, or add some to the shared background images.
            </div>
            <div v-for="group in imageGroups" :key="group.key" class="flex flex-column gap-1">
                <span class="image-group-label">{{ group.label }}</span>
                <div class="image-grid">
                    <button
                        v-for="choice in group.items"
                        :key="choice.ref"
                        type="button"
                        class="image-tile"
                        :class="{ selected: slot.selected === choice.ref }"
                        :title="choice.label"
                        :disabled="disabled"
                        :aria-pressed="slot.selected === choice.ref"
                        :data-testid="`image-${slot.mode}-${choice.ref}`"
                        @click="selectImage(slot.mode, choice.ref)"
                    >
                        <img :src="choice.url" :alt="choice.label" loading="lazy" />
                        <span class="image-tile-name">{{ choice.label }}</span>
                    </button>
                </div>
            </div>
        </div>

        <input
            ref="uploadInput"
            type="file"
            :accept="UPLOAD_ACCEPT"
            style="display: none"
            data-testid="image-upload-input"
            @change="handleUpload"
        />

        <template v-if="background.image">
            <div class="edit-row">
                <div class="flex flex-column gap-1 flex-grow-1">
                    <label class="font-semibold text-sm">Fit</label>
                    <Select v-model="background.image.fit" :options="fitOptions" optionLabel="label" optionValue="value" :disabled="disabled" />
                </div>
                <div class="flex flex-column gap-1 flex-grow-1">
                    <label class="font-semibold text-sm">Position</label>
                    <Select v-model="background.image.position" :options="positionOptions" optionLabel="label" optionValue="value" :disabled="disabled" />
                </div>
                <div class="flex flex-column gap-1 flex-grow-1">
                    <label class="font-semibold text-sm">Repeat</label>
                    <Select v-model="background.image.repeat" :options="repeatOptions" optionLabel="label" optionValue="value" :disabled="disabled" />
                </div>
            </div>
            <small class="edit-hint">
                Fit, position and repeat apply in browsers only. E-ink and image
                dashboards always cover the canvas.
            </small>
        </template>
    </div>
</template>

<style scoped lang="scss">
@use '@/assets/breakpoints' as bp;

.bg-editor {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
}

.bg-preview {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 0.75rem;
}

.bg-preview-tile {
    position: relative;
    height: 140px;
    border-radius: 6px;
    border: 1px solid var(--p-content-border-color);
    display: flex;
    align-items: flex-end;
    padding: 0.75rem;
    overflow: hidden;
}

.bg-preview-mode {
    position: absolute;
    top: 0.4rem;
    left: 0.5rem;
    font-size: 0.7rem;
    font-weight: 600;
    padding: 0.05rem 0.4rem;
    border-radius: 999px;
    background: var(--dashi-surface);
    color: var(--dashi-text-muted);
}

.bg-preview-card {
    background: var(--dashi-surface);
    border: 1px solid var(--dashi-border);
    border-radius: 6px;
    padding: 0.4rem 0.6rem;
    color: var(--dashi-text);
}

.bg-preview-text {
    font-weight: 600;
    font-size: 0.85rem;
}

.bg-preview-muted {
    color: var(--dashi-text-muted);
    font-size: 0.75rem;
}

.bg-preview-swatches {
    display: flex;
    gap: 0.2rem;
    margin-top: 0.3rem;
}

.bg-preview-swatch {
    width: 0.8rem;
    height: 0.8rem;
    border-radius: 50%;
}

.stop-strip {
    height: 1.25rem;
    border-radius: 4px;
    border: 1px solid var(--p-content-border-color);
}

.angle-readout {
    font-variant-numeric: tabular-nums;
    color: var(--p-text-muted-color);
    font-size: 0.875rem;
}

.angle-slider {
    margin: 0.5rem 0.25rem;
}

.edit-row {
    display: flex;
    gap: 0.75rem;
}

.edit-hint {
    color: var(--p-text-muted-color);
}

.image-warning {
    color: var(--p-red-500);
}

.image-group-label {
    font-size: 0.75rem;
    color: var(--p-text-muted-color);
}

.image-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(96px, 1fr));
    gap: 0.5rem;
    max-height: 260px;
    overflow-y: auto;
}

.image-tile {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
    padding: 0.25rem;
    border: 2px solid var(--p-content-border-color);
    border-radius: 6px;
    background: var(--p-surface-0);
    cursor: pointer;
    text-align: left;
}

.image-tile:hover:not(:disabled) {
    border-color: var(--p-primary-200);
}

.image-tile.selected {
    border-color: var(--p-primary-color);
}

.image-tile:disabled {
    cursor: default;
    opacity: 0.7;
}

.image-tile img {
    width: 100%;
    aspect-ratio: 4 / 3;
    object-fit: cover;
    border-radius: 4px;
    background: var(--p-surface-100);
}

.image-tile-name {
    font-size: 0.7rem;
    color: var(--p-text-muted-color);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.pool-link {
    color: var(--p-primary-color);
}

/* The colour swatch is unbordered by default, so a white or near-white value
   is invisible against the card and there is nothing to click. */
:deep(.p-colorpicker-preview) {
    border: 1px solid var(--p-content-border-color);
    border-radius: 4px;
}

@include bp.below-sm {
    .bg-preview,
    .edit-row {
        grid-template-columns: 1fr;
        flex-direction: column;
    }
}
</style>
