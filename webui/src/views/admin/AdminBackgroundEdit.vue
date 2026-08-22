<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import Card from 'primevue/card'
import InputText from 'primevue/inputtext'
import InputNumber from 'primevue/inputnumber'
import Select from 'primevue/select'
import SelectButton from 'primevue/selectbutton'
import ColorPicker from 'primevue/colorpicker'
import ToggleSwitch from 'primevue/toggleswitch'
import { useToast } from 'primevue/usetoast'

import { useGetBackground, useUpdateBackground, useBackgroundAssets } from '@/composables/useBackgrounds'
import { useDataItems } from '@/composables/useDataItems'
import { browserCss, pageBgValue } from '@/lib/backgroundCss'
import type {
    Background,
    ImageFit,
    ImagePosition,
    ImageRepeat,
} from '@/types/background'

const route = useRoute()
const router = useRouter()
const toast = useToast()
const id = computed(() => route.params.id as string)

const { data: serverBackground, isLoading, isError } = useGetBackground(() => id.value)
const { updateBackground, isUpdating } = useUpdateBackground()
const { assets, uploadAsset, isUploadingAsset } = useBackgroundAssets(() => id.value)
const { items: sharedImages } = useDataItems('backgrounds')

// Editing happens on a local clone so nothing is written until Save.
const local = ref<Background | null>(null)

watch(serverBackground, (val) => {
    if (val && !local.value) {
        local.value = JSON.parse(JSON.stringify(val))
    }
}, { immediate: true })

// ---- base: colour or gradient, never both ----

type BaseKind = 'none' | 'color' | 'gradient'

const baseKindOptions = [
    { label: 'None', value: 'none' },
    { label: 'Colour', value: 'color' },
    { label: 'Gradient', value: 'gradient' },
]

const baseKind = computed<BaseKind>(() => {
    if (local.value?.color) return 'color'
    if (local.value?.gradient) return 'gradient'
    return 'none'
})

// The server rejects a background carrying both a colour and a gradient, so
// selecting one deletes the other rather than leaving it behind.
const setBaseKind = (kind: BaseKind) => {
    if (!local.value) return
    delete local.value.color
    delete local.value.gradient
    if (kind === 'color') {
        local.value.color = { light: '#ffffff' }
    } else if (kind === 'gradient') {
        local.value.gradient = { direction: 'to bottom right', light: ['#667eea', '#764ba2'] }
    }
}

// ---- dark overrides ----
// An absent `dark` key means "fall back to light". The toggles therefore delete
// the key rather than writing an empty string, which would fail validation.

const colorDarkEnabled = computed({
    get: () => local.value?.color?.dark !== undefined,
    set: (on: boolean) => {
        if (!local.value?.color) return
        if (on) local.value.color.dark = local.value.color.light
        else delete local.value.color.dark
    },
})

const gradientDarkEnabled = computed({
    get: () => local.value?.gradient?.dark !== undefined,
    set: (on: boolean) => {
        if (!local.value?.gradient) return
        if (on) local.value.gradient.dark = [...local.value.gradient.light]
        else delete local.value.gradient.dark
    },
})

const imageDarkEnabled = computed({
    get: () => local.value?.image?.dark !== undefined,
    set: (on: boolean) => {
        if (!local.value?.image) return
        if (on) local.value.image.dark = local.value.image.light
        else delete local.value.image.dark
    },
})

// ---- gradient direction ----

const directionKeywords = [
    'to top', 'to bottom', 'to left', 'to right',
    'to top left', 'to top right', 'to bottom left', 'to bottom right',
]

const directionOptions = [
    ...directionKeywords.map(k => ({ label: k, value: k })),
    { label: 'Custom angle…', value: 'custom' },
]

// A direction is either one of the eight keywords or "{n}deg"; the Select shows
// "custom" for the latter and reveals a number input.
const directionChoice = computed({
    get: () => {
        const dir = local.value?.gradient?.direction ?? ''
        return directionKeywords.includes(dir) ? dir : 'custom'
    },
    set: (v: string) => {
        if (!local.value?.gradient) return
        local.value.gradient.direction = v === 'custom' ? '135deg' : v
    },
})

const customAngle = computed({
    get: () => parseInt((local.value?.gradient?.direction ?? '').replace('deg', ''), 10) || 0,
    set: (v: number) => {
        if (!local.value?.gradient) return
        const n = Math.min(360, Math.max(0, Math.round(v || 0)))
        local.value.gradient.direction = `${n}deg`
    },
})

const addStop = (mode: 'light' | 'dark') => {
    const stops = local.value?.gradient?.[mode]
    if (stops) stops.push('#ffffff')
}

// A gradient needs at least two stops, so the last two cannot be removed.
const removeStop = (mode: 'light' | 'dark', index: number) => {
    const stops = local.value?.gradient?.[mode]
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
    get: () => local.value?.image !== undefined,
    set: (on: boolean) => {
        if (!local.value) return
        if (on) {
            local.value.image = {
                light: '',
                fit: 'cover',
                position: 'center',
                repeat: 'no-repeat',
            }
        } else {
            delete local.value.image
        }
    },
})

// Two sources, grouped: this background's own uploads (asset:) and the
// admin-curated shared pool (shared:).
const imageOptions = computed(() => {
    const groups: { label: string; items: { label: string; value: string }[] }[] = []
    const own = assets.value ?? []
    if (own.length) {
        groups.push({
            label: 'This background',
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
        if (local.value?.image) local.value.image.light = `asset:${file.name}`
        toast.add({ severity: 'success', summary: 'Uploaded', detail: file.name, life: 3000 })
    } catch {
        toast.add({ severity: 'error', summary: 'Error', detail: 'Failed to upload image', life: 5000 })
    } finally {
        if (uploadInput.value) uploadInput.value.value = ''
    }
}

// ---- preview ----

const previewMode = ref<'light' | 'dark'>('light')
const previewModeOptions = [
    { label: 'Light', value: 'light' },
    { label: 'Dark', value: 'dark' },
]

// Generated by the same logic the server uses, so the preview shows what the
// dashboard will actually render — no save and no round-trip required.
const previewStyle = computed(() => {
    if (!local.value) return {}
    const value = pageBgValue(browserCss(local.value), previewMode.value)
    return { background: value }
})

// ---- save ----

const handleSave = async () => {
    if (!local.value) return
    try {
        await updateBackground({ id: id.value, payload: local.value })
        toast.add({ severity: 'success', summary: 'Saved', detail: local.value.name, life: 3000 })
    } catch (err: unknown) {
        // A 400 carries the store's validation message — surface it rather than
        // a generic failure, since it names the offending field.
        const detail =
            (err as { response?: { data?: { error?: string } } })?.response?.data?.error ??
            'Failed to save background'
        toast.add({ severity: 'error', summary: 'Error', detail, life: 6000 })
    }
}
</script>

<template>
    <div class="admin-section">
        <div class="admin-section-header">
            <h2 class="admin-section-title">Edit background</h2>
            <div class="flex gap-2">
                <Button
                    label="Back"
                    icon="ti ti-arrow-left"
                    severity="secondary"
                    @click="router.push({ name: 'admin-backgrounds' })"
                />
                <Button
                    label="Save"
                    icon="ti ti-device-floppy"
                    :loading="isUpdating"
                    :disabled="!local"
                    @click="handleSave"
                />
            </div>
        </div>

        <Card v-if="isLoading">
            <template #content><div class="info-message">Loading…</div></template>
        </Card>

        <Card v-else-if="isError">
            <template #content><div class="info-message">Background not found.</div></template>
        </Card>

        <template v-else-if="local">
            <Card>
                <template #content>
                    <div class="edit-grid">
                        <div class="edit-fields">
                            <div class="flex flex-column gap-1">
                                <label class="font-semibold text-sm">Name</label>
                                <InputText v-model="local.name" class="w-full" />
                            </div>

                            <!-- Base: colour or gradient, mutually exclusive -->
                            <label class="edit-subsection">Base</label>
                            <SelectButton
                                :modelValue="baseKind"
                                :options="baseKindOptions"
                                optionLabel="label"
                                optionValue="value"
                                :allowEmpty="false"
                                @update:modelValue="(v: BaseKind) => setBaseKind(v)"
                            />

                            <template v-if="local.color">
                                <div class="flex flex-column gap-1">
                                    <label class="font-semibold text-sm">Colour</label>
                                    <div class="flex align-items-center gap-2">
                                        <ColorPicker
                                            :modelValue="local.color.light.replace('#', '')"
                                            @update:modelValue="(v: string | undefined) => { if (v !== undefined && local?.color) local.color.light = '#' + v }"
                                        />
                                        <InputText v-model="local.color.light" class="flex-grow-1" placeholder="#ffffff" />
                                    </div>
                                </div>
                                <div class="flex align-items-center gap-2">
                                    <ToggleSwitch v-model="colorDarkEnabled" />
                                    <span class="text-sm">Different colour in dark mode</span>
                                </div>
                                <div v-if="local.color.dark !== undefined" class="flex align-items-center gap-2">
                                    <ColorPicker
                                        :modelValue="local.color.dark.replace('#', '')"
                                        @update:modelValue="(v: string | undefined) => { if (v !== undefined && local?.color) local.color.dark = '#' + v }"
                                    />
                                    <InputText v-model="local.color.dark" class="flex-grow-1" placeholder="#101014" />
                                </div>
                            </template>

                            <template v-if="local.gradient">
                                <div class="flex flex-column gap-1">
                                    <label class="font-semibold text-sm">Direction</label>
                                    <Select
                                        :modelValue="directionChoice"
                                        @update:modelValue="(v: string | undefined) => { if (v !== undefined) directionChoice = v }"
                                        :options="directionOptions"
                                        optionLabel="label"
                                        optionValue="value"
                                        class="w-full"
                                    />
                                </div>
                                <div v-if="directionChoice === 'custom'" class="flex flex-column gap-1">
                                    <label class="font-semibold text-sm">Angle</label>
                                    <InputNumber
                                        :modelValue="customAngle"
                                        @update:modelValue="(v: number | null) => { customAngle = v ?? 0 }"
                                        :min="0"
                                        :max="360"
                                        suffix="°"
                                        class="w-full"
                                    />
                                </div>

                                <div class="flex flex-column gap-1">
                                    <label class="font-semibold text-sm">Colour stops</label>
                                    <div
                                        v-for="(_stop, i) in local.gradient.light"
                                        :key="'l' + i"
                                        class="flex align-items-center gap-2"
                                    >
                                        <ColorPicker
                                            :modelValue="local.gradient.light[i].replace('#', '')"
                                            @update:modelValue="(v: string | undefined) => { if (v !== undefined && local?.gradient) local.gradient.light[i] = '#' + v }"
                                        />
                                        <InputText v-model="local.gradient.light[i]" class="flex-grow-1" />
                                        <Button
                                            icon="ti ti-trash"
                                            text
                                            rounded
                                            severity="danger"
                                            :disabled="local.gradient.light.length <= 2"
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
                                <div v-if="local.gradient.dark" class="flex flex-column gap-1">
                                    <label class="font-semibold text-sm">Dark stops</label>
                                    <div
                                        v-for="(_stop, i) in local.gradient.dark"
                                        :key="'d' + i"
                                        class="flex align-items-center gap-2"
                                    >
                                        <ColorPicker
                                            :modelValue="local.gradient.dark[i].replace('#', '')"
                                            @update:modelValue="(v: string | undefined) => { if (v !== undefined && local?.gradient?.dark) local.gradient.dark[i] = '#' + v }"
                                        />
                                        <InputText v-model="local.gradient.dark[i]" class="flex-grow-1" />
                                        <Button
                                            icon="ti ti-trash"
                                            text
                                            rounded
                                            severity="danger"
                                            :disabled="local.gradient.dark.length <= 2"
                                            @click="removeStop('dark', i)"
                                        />
                                    </div>
                                    <Button label="Add stop" icon="ti ti-plus" text size="small" @click="addStop('dark')" />
                                </div>
                            </template>

                            <!-- Image layer, painted over the base -->
                            <label class="edit-subsection">Image</label>
                            <div class="flex align-items-center gap-2">
                                <ToggleSwitch v-model="imageEnabled" />
                                <span class="text-sm">Use an image</span>
                            </div>

                            <template v-if="local.image">
                                <div class="flex flex-column gap-1">
                                    <label class="font-semibold text-sm">Image</label>
                                    <div class="flex gap-2">
                                        <Select
                                            v-model="local.image.light"
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
                                <div v-if="local.image.dark !== undefined" class="flex flex-column gap-1">
                                    <label class="font-semibold text-sm">Dark image</label>
                                    <Select
                                        v-model="local.image.dark"
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
                                        <Select v-model="local.image.fit" :options="fitOptions" optionLabel="label" optionValue="value" />
                                    </div>
                                    <div class="flex flex-column gap-1 flex-grow-1">
                                        <label class="font-semibold text-sm">Position</label>
                                        <Select v-model="local.image.position" :options="positionOptions" optionLabel="label" optionValue="value" />
                                    </div>
                                    <div class="flex flex-column gap-1 flex-grow-1">
                                        <label class="font-semibold text-sm">Repeat</label>
                                        <Select v-model="local.image.repeat" :options="repeatOptions" optionLabel="label" optionValue="value" />
                                    </div>
                                </div>
                                <small class="edit-hint">
                                    Fit, position and repeat apply in browsers only. E-ink and
                                    image dashboards always cover the canvas.
                                </small>
                            </template>
                        </div>

                        <!-- Live preview -->
                        <div class="edit-preview">
                            <div class="flex align-items-center justify-content-between">
                                <label class="font-semibold text-sm">Preview</label>
                                <SelectButton
                                    v-model="previewMode"
                                    :options="previewModeOptions"
                                    optionLabel="label"
                                    optionValue="value"
                                    :allowEmpty="false"
                                    size="small"
                                />
                            </div>
                            <div class="preview-box" :style="previewStyle" />
                            <small class="edit-hint">
                                Unsaved changes are shown. Nothing is stored until you press Save.
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
