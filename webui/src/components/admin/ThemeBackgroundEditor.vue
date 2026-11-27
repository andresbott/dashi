<script setup lang="ts">
// The Background section of the theme editor for ONE colour mode. A
// background belongs to its theme (there is no separate background entity): a
// base — the mode's palette `bg` or a linear gradient — with an optional image
// painted over it. Edits mutate the bound background in place, so the
// parent's autosave picks them up like any other theme field. The light and
// dark modes each get their own instance; nothing here is shared between them.
import { computed, ref } from 'vue'
import Button from 'primevue/button'
import ColorPicker from 'primevue/colorpicker'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import SelectButton from 'primevue/selectbutton'
import Slider from 'primevue/slider'
import { useToast } from 'primevue/usetoast'

import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import { useThemeAssets } from '@/composables/useThemes'
import { themeAssetUrl } from '@/lib/api/themes'
import { isValidStopColor } from '@/lib/themeCss'
import { uniqueAssetName } from '@/lib/themeEditor'
import type {
    ImageFit,
    ImagePosition,
    ImageRepeat,
    ModeBackground,
    PaletteRole,
} from '@/types/theme'

const props = withDefaults(
    defineProps<{
        themeId: string
        // Scopes the test ids and labels, so a light and a dark editor can be
        // mounted side by side.
        mode: 'light' | 'dark'
        // Persists the theme now (the parent's autosave flush). Awaited before
        // and after an upload, see handleUpload.
        flush: () => Promise<void>
        // Inline style of the preview tile: the theme scoped to this mode.
        previewStyle: Record<string, string>
        // Read-only: every control is disabled and nothing is written.
        disabled?: boolean
    }>(),
    { disabled: false },
)

const emit = defineEmits<{
    // True while a file is uploading: the parent blocks autosave (and the
    // whole editor) until it reports false again.
    uploading: [value: boolean]
}>()

const background = defineModel<ModeBackground>('background', { required: true })

const toast = useToast()
const { assets, uploadAsset, isUploadingAsset } = useThemeAssets(() => props.themeId)

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
        background.value.gradient ??= { direction: '135deg', stops: ['#667eea', '#764ba2'] }
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

const stops = computed<string[] | undefined>(() => background.value.gradient?.stops)

// A new stop repeats the last colour, so adding one never jumps the gradient.
const addStop = () => {
    const s = stops.value
    if (!s || props.disabled) return
    s.push(s[s.length - 1] ?? '#ffffff')
}

// A gradient needs at least two stops, so the last two cannot be removed.
const removeStop = (index: number) => {
    const s = stops.value
    if (!s || props.disabled || s.length <= 2) return
    s.splice(index, 1)
}

const moveStop = (index: number, delta: -1 | 1) => {
    const s = stops.value
    const target = index + delta
    if (!s || props.disabled || target < 0 || target >= s.length) return
    ;[s[index], s[target]] = [s[target], s[index]]
}

const setStop = (index: number, value: string) => {
    const s = stops.value
    if (!s || props.disabled) return
    s[index] = value
}

// ColorPicker understands only six-digit hex: expand #rgb and drop the alpha
// of #rrggbbaa for display, and keep that alpha when the picker writes back.
const pickerValue = (v: string): string => {
    const h = (v ?? '').replace('#', '')
    if (/^[0-9a-fA-F]{3}$/.test(h)) return h.split('').map(c => c + c).join('')
    if (/^[0-9a-fA-F]{6}(?:[0-9a-fA-F]{2})?$/.test(h)) return h.slice(0, 6)
    return '000000'
}

const setStopFromPicker = (index: number, v: string | undefined) => {
    const prev = stops.value?.[index] ?? ''
    if (v === undefined) return
    const alpha = /^#[0-9a-fA-F]{8}$/.test(prev) ? prev.slice(7) : ''
    setStop(index, `#${v}${alpha}`)
}

// The strip above the stop list shows the stops left to right over a
// checkerboard, so translucent stops are visible as such.
const stripStyle = (list: string[]) => {
    const valid = list.filter(isValidStopColor)
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
// rejected on save, so it is not uploadable.
const UPLOAD_ACCEPT = '.png,.jpg,.jpeg,.webp,.svg'

const ASSET = 'asset:'

const imageRef = computed(() => background.value.image?.ref ?? '')

// Only asset: refs exist: a theme references nothing but its own uploads.
const assetName = computed(() => (imageRef.value.startsWith(ASSET) ? imageRef.value.slice(ASSET.length) : ''))

const refLabel = computed(() => imageRef.value.slice(imageRef.value.indexOf(':') + 1))

// A reference to a file the theme does not have (a leftover from the removed
// shared pool, or a hand-edited file): it renders nothing, so it is called out
// once the asset list has loaded.
const isMissing = computed(() => {
    if (!imageRef.value || assets.value === undefined) return false
    return !assetName.value || !assets.value.includes(assetName.value)
})

const thumbUrl = computed(() =>
    assetName.value && !isMissing.value ? themeAssetUrl(props.themeId, assetName.value) : '',
)

const uploadInput = ref<HTMLInputElement | null>(null)

const pickUpload = () => {
    if (props.disabled) return
    uploadInput.value?.click()
}

// Uploads go into this theme's own assets under a name no earlier upload has
// (see uniqueAssetName), so light and dark never overwrite each other's file.
//
// The order pairs with the server's prune-on-save and its grace window:
//  1. flush pending edits, so nothing unsaved is waiting while autosave is off;
//  2. block autosave (and the editor) while the file uploads;
//  3. reopen autosave BEFORE setting the ref, so that edit is not dropped;
//  4. set the ref (Replace keeps fit/position/repeat) and flush it right away,
//     so the new file is referenced well inside the grace window and the old
//     one is pruned.
// A failed upload sets nothing.
const handleUpload = async (event: Event) => {
    const input = event.target as HTMLInputElement
    const file = input.files?.[0]
    if (!file || props.disabled) return
    const name = uniqueAssetName(file.name, Date.now())
    try {
        await props.flush()
        emit('uploading', true)
        try {
            await uploadAsset({ name, bytes: await file.arrayBuffer() })
        } finally {
            emit('uploading', false)
        }
        // No disabled guard here: the prop still reads the uploading state
        // until the parent re-renders.
        const img = background.value.image
        if (img) img.ref = `${ASSET}${name}`
        else background.value.image = { ref: `${ASSET}${name}`, fit: 'cover', position: 'center', repeat: 'no-repeat' }
        await props.flush()
        toast.add({ severity: 'success', summary: 'Uploaded', detail: file.name, life: 3000 })
    } catch {
        toast.add({ severity: 'error', summary: 'Error', detail: 'Failed to upload image', life: 5000 })
    } finally {
        input.value = ''
    }
}

// Remove only clears this mode's reference: the other mode may use the same
// file, and the server deletes unreferenced files on save.
const removeDialogVisible = ref(false)

const askRemove = () => {
    if (props.disabled || !background.value.image) return
    removeDialogVisible.value = true
}

const confirmRemove = () => {
    removeDialogVisible.value = false
    if (props.disabled) return
    delete background.value.image
}

// ---- preview ----

const modeLabel = computed(() => (props.mode === 'light' ? 'Light' : 'Dark'))

const swatches: PaletteRole[] = ['primary', 'secondary', 'success', 'warning', 'danger']
</script>

<template>
    <div class="bg-editor">
        <!-- Live preview of this mode -->
        <div class="bg-preview-tile" :style="previewStyle" :data-testid="`bg-preview-${mode}`">
            <span class="bg-preview-mode">{{ modeLabel }}</span>
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

            <div class="flex flex-column gap-1">
                <label class="font-semibold text-sm">Colour stops</label>
                <div class="stop-strip" :style="stripStyle(background.gradient.stops)" />
                <div
                    v-for="(stop, i) in background.gradient.stops"
                    :key="i"
                    class="flex align-items-center gap-2"
                >
                    <ColorPicker
                        :modelValue="pickerValue(stop)"
                        :disabled="disabled"
                        @update:modelValue="(v: string | undefined) => setStopFromPicker(i, v)"
                    />
                    <InputText
                        :modelValue="stop"
                        :invalid="!isValidStopColor(stop)"
                        :disabled="disabled"
                        :data-testid="`stop-${mode}-${i}`"
                        class="flex-grow-1 min-w-0"
                        @update:modelValue="(v: string | undefined) => setStop(i, v ?? '')"
                    />
                    <Button
                        icon="ti ti-arrow-up"
                        text
                        rounded
                        severity="secondary"
                        aria-label="Move stop up"
                        :disabled="disabled || i === 0"
                        :data-testid="`stop-${mode}-${i}-up`"
                        @click="moveStop(i, -1)"
                    />
                    <Button
                        icon="ti ti-arrow-down"
                        text
                        rounded
                        severity="secondary"
                        aria-label="Move stop down"
                        :disabled="disabled || i === background.gradient.stops.length - 1"
                        :data-testid="`stop-${mode}-${i}-down`"
                        @click="moveStop(i, 1)"
                    />
                    <Button
                        icon="ti ti-trash"
                        text
                        rounded
                        severity="danger"
                        aria-label="Remove stop"
                        :disabled="disabled || background.gradient.stops.length <= 2"
                        :data-testid="`stop-${mode}-${i}-remove`"
                        @click="removeStop(i)"
                    />
                </div>
                <Button
                    label="Add stop"
                    icon="ti ti-plus"
                    text
                    size="small"
                    class="align-self-start"
                    :disabled="disabled"
                    :data-testid="`stop-${mode}-add`"
                    @click="addStop"
                />
            </div>
            <small class="edit-hint">
                Stops are spread evenly. Hex colours, alpha allowed: e-ink and image
                dashboards flatten translucent stops over the page background.
            </small>
        </template>

        <!-- Image layer, painted over the base: one file per mode -->
        <div class="flex flex-column gap-1 mt-2">
            <label class="font-semibold text-sm">Image</label>
            <div class="image-slot">
                <div class="image-thumb">
                    <img
                        v-if="thumbUrl"
                        :src="thumbUrl"
                        :alt="refLabel"
                        :data-testid="`image-${mode}-thumb`"
                    />
                    <span v-else class="edit-hint text-sm">No image</span>
                </div>
                <div class="flex flex-column gap-2 min-w-0">
                    <span v-if="assetName" class="image-name" :title="assetName">{{ assetName }}</span>
                    <div class="flex flex-wrap gap-2">
                        <Button
                            :label="background.image ? 'Replace' : 'Upload'"
                            icon="ti ti-upload"
                            severity="secondary"
                            size="small"
                            :loading="isUploadingAsset"
                            :disabled="disabled"
                            :data-testid="`image-${mode}-upload`"
                            @click="pickUpload"
                        />
                        <Button
                            v-if="background.image"
                            label="Remove"
                            icon="ti ti-x"
                            severity="secondary"
                            size="small"
                            text
                            :disabled="disabled"
                            :data-testid="`image-${mode}-remove`"
                            @click="askRemove"
                        />
                    </div>
                </div>
            </div>
            <small v-if="background.image && !imageRef" class="image-warning">
                Upload an image, or remove the image layer.
            </small>
            <small v-else-if="isMissing" class="image-warning">
                “{{ refLabel }}” no longer exists. Upload another image.
            </small>
            <small class="edit-hint">
                Images are stored in the theme itself and copied along with it.
            </small>
        </div>

        <ConfirmDialog
            v-model:visible="removeDialogVisible"
            :name="refLabel"
            title="Remove image"
            :message="`Remove the ${mode} background image?`"
            @confirm="confirmRemove"
        />

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

.image-slot {
    display: flex;
    align-items: center;
    gap: 0.75rem;
}

.image-thumb {
    flex: none;
    width: 120px;
    aspect-ratio: 4 / 3;
    display: flex;
    align-items: center;
    justify-content: center;
    border: 1px solid var(--p-content-border-color);
    border-radius: 6px;
    background: var(--p-surface-100);
    overflow: hidden;

    img {
        width: 100%;
        height: 100%;
        object-fit: cover;
    }
}

.image-name {
    font-size: 0.8rem;
    color: var(--p-text-muted-color);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

/* The colour swatch is unbordered by default, so a white or near-white value
   is invisible against the card and there is nothing to click. */
:deep(.p-colorpicker-preview) {
    border: 1px solid var(--p-content-border-color);
    border-radius: 4px;
}

@include bp.below-sm {
    .edit-row {
        flex-direction: column;
    }
}
</style>
