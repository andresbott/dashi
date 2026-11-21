<script setup lang="ts">
import { ref, computed, watch, onBeforeUnmount } from 'vue'
import Dialog from 'primevue/dialog'
import Button from 'primevue/button'
import Select from 'primevue/select'
import InputNumber from 'primevue/inputnumber'
import Checkbox from 'primevue/checkbox'
import { apiErrorMessage } from '@/lib/api/client'
import {
    CUSTOM_DEVICE,
    DEVICE_PRESETS,
    MAX_PREVIEW_SIDE,
    PREVIEW_FORMATS,
    PREVIEW_ROTATIONS,
    clampSide,
    einkPreviewUrl,
    type PreviewFormat,
    type PreviewRotation,
} from '@/lib/einkPreview'

// Shows the server-rendered PNG an e-ink panel would receive for this
// dashboard. Every control is local to the dialog: nothing is written to the
// dashboard.
const props = defineProps<{
    dashboardId: string
    pageNames: string[]
    // Page selected when the dialog opens (the editor's active page).
    initialPage?: number
    // Persists pending edits first, so the render reflects them.
    flush?: () => Promise<void>
}>()

const visible = defineModel<boolean>('visible', { default: false })

const firstPreset = DEVICE_PRESETS[0]
const device = ref<string>(firstPreset.key)
const width = ref<number>(firstPreset.width)
const height = ref<number>(firstPreset.height)
const format = ref<PreviewFormat>(firstPreset.format)
const rotation = ref<PreviewRotation>(0)
const page = ref(0)
const actualSize = ref(false)

const deviceOptions = [
    ...DEVICE_PRESETS.map(p => ({ label: p.label, value: p.key })),
    { label: 'Custom size', value: CUSTOM_DEVICE },
]
const rotationOptions = PREVIEW_ROTATIONS.map(r => ({ label: `${r}°`, value: r }))
const pageOptions = computed(() =>
    props.pageNames.map((name, i) => ({ label: name || `Page ${i + 1}`, value: i })),
)
const isCustom = computed(() => device.value === CUSTOM_DEVICE)

const url = computed(() =>
    einkPreviewUrl({
        dashboardId: props.dashboardId,
        width: clampSide(width.value),
        height: clampSide(height.value),
        format: format.value,
        rotation: rotation.value,
        page: page.value,
    }),
)

const imageSrc = ref<string | null>(null)
const loading = ref(false)
const error = ref<string | null>(null)
let requestSeq = 0
// The URL of the newest request, so a control change that lands on the same
// render (e.g. the page reset when the dialog opens) does not fetch it twice.
let requestedUrl = ''

const setImage = (src: string | null) => {
    if (imageSrc.value) URL.revokeObjectURL(imageSrc.value)
    imageSrc.value = src
}

const errorText = async (res: Response): Promise<string> => {
    const body = (await res.text()).trim()
    try {
        const msg = apiErrorMessage({ response: { data: JSON.parse(body) } })
        if (msg) return msg
    } catch {
        // plain-text body
    }
    return body || `HTTP ${res.status}`
}

// Fetched rather than set as <img src> so a server error (bad size, a page
// that no longer exists) can be shown, and so the render is never cached.
const load = async () => {
    const seq = ++requestSeq
    const target = url.value
    requestedUrl = target
    loading.value = true
    error.value = null
    try {
        if (props.flush) await props.flush()
        const res = await fetch(target, { cache: 'no-store' })
        if (seq !== requestSeq) return
        if (!res.ok) {
            error.value = await errorText(res)
            setImage(null)
            return
        }
        if (!(res.headers.get('Content-Type') ?? '').startsWith('image/')) {
            error.value = 'The server did not return an image. Only image dashboards render for e-ink.'
            setImage(null)
            return
        }
        const blob = await res.blob()
        if (seq !== requestSeq) return
        setImage(URL.createObjectURL(blob))
    } catch (e) {
        if (seq !== requestSeq) return
        error.value = e instanceof Error ? e.message : 'Failed to load the preview'
        setImage(null)
    } finally {
        if (seq === requestSeq) loading.value = false
    }
}

const applyDevice = (key: string) => {
    device.value = key
    const preset = DEVICE_PRESETS.find(p => p.key === key)
    if (!preset) return
    width.value = preset.width
    height.value = preset.height
    format.value = preset.format
}

// Re-render whenever the request changes while the dialog is open, after a
// short pause so typing a custom size does not render once per keystroke.
const RELOAD_DELAY_MS = 350
let reloadTimer: ReturnType<typeof setTimeout> | undefined
const cancelReload = () => {
    if (reloadTimer !== undefined) clearTimeout(reloadTimer)
    reloadTimer = undefined
}
watch(url, (next) => {
    if (!visible.value || next === requestedUrl) return
    cancelReload()
    reloadTimer = setTimeout(() => {
        reloadTimer = undefined
        load()
    }, RELOAD_DELAY_MS)
})

const refresh = () => {
    cancelReload()
    load()
}

watch(visible, (open) => {
    cancelReload()
    if (open) {
        const max = Math.max(0, props.pageNames.length - 1)
        page.value = Math.min(Math.max(0, props.initialPage ?? 0), max)
        load()
    } else {
        requestSeq++
        loading.value = false
        setImage(null)
    }
}, { immediate: true })

onBeforeUnmount(() => {
    cancelReload()
    requestSeq++
    setImage(null)
})
</script>

<template>
    <Dialog
        v-model:visible="visible"
        modal
        :closable="true"
        :draggable="false"
        header="E-ink preview"
        :style="{ width: 'min(1100px, 95vw)' }"
    >
        <div class="eink-controls">
            <div class="flex flex-column gap-1">
                <label class="font-semibold text-sm">Device</label>
                <Select
                    :modelValue="device"
                    @update:modelValue="(v: string | undefined) => { if (v !== undefined) applyDevice(v) }"
                    :options="deviceOptions"
                    optionLabel="label"
                    optionValue="value"
                    data-testid="eink-device"
                />
            </div>
            <div v-if="isCustom" class="flex flex-column gap-1">
                <label class="font-semibold text-sm">Size (W × H px)</label>
                <div class="flex align-items-center gap-1">
                    <InputNumber
                        v-model="width"
                        :min="1"
                        :max="MAX_PREVIEW_SIDE"
                        :useGrouping="false"
                        inputClass="eink-size"
                        data-testid="eink-width"
                    />
                    <span>×</span>
                    <InputNumber
                        v-model="height"
                        :min="1"
                        :max="MAX_PREVIEW_SIDE"
                        :useGrouping="false"
                        inputClass="eink-size"
                        data-testid="eink-height"
                    />
                </div>
            </div>
            <div class="flex flex-column gap-1">
                <label class="font-semibold text-sm">Format</label>
                <Select
                    v-model="format"
                    :options="PREVIEW_FORMATS"
                    optionLabel="label"
                    optionValue="value"
                    data-testid="eink-format"
                />
            </div>
            <div class="flex flex-column gap-1">
                <label class="font-semibold text-sm">Rotation</label>
                <Select
                    v-model="rotation"
                    :options="rotationOptions"
                    optionLabel="label"
                    optionValue="value"
                    data-testid="eink-rotation"
                />
            </div>
            <div class="flex flex-column gap-1">
                <label class="font-semibold text-sm">Page</label>
                <Select
                    v-model="page"
                    :options="pageOptions"
                    optionLabel="label"
                    optionValue="value"
                    data-testid="eink-page"
                />
            </div>
            <div class="flex align-items-end gap-3">
                <div class="flex align-items-center gap-2 eink-actual">
                    <Checkbox v-model="actualSize" inputId="eink-actual" binary />
                    <label for="eink-actual" class="text-sm">Actual pixels</label>
                </div>
                <Button
                    icon="ti ti-refresh"
                    label="Refresh"
                    severity="secondary"
                    :loading="loading"
                    data-testid="eink-refresh"
                    @click="refresh"
                />
            </div>
        </div>

        <small class="eink-hint">
            Rendered by the server exactly as a panel requests it, dithered with the dashboard's
            Dither setting. With 90° or 270° the page is laid out in portrait and rotated onto the
            panel, so the image shows the panel's native orientation. Nothing here is saved.
        </small>

        <div class="eink-stage" :class="{ 'eink-stage-actual': actualSize }">
            <div v-if="error" class="eink-error" role="alert" data-testid="eink-error">
                <i class="ti ti-alert-circle" /> {{ error }}
            </div>
            <div v-else-if="!imageSrc && loading" class="eink-placeholder">
                <i class="ti ti-loader-2 eink-spin" /> Rendering…
            </div>
            <img
                v-if="imageSrc && !error"
                :src="imageSrc"
                :class="{ 'eink-loading': loading }"
                alt="E-ink render of the dashboard"
                data-testid="eink-image"
            />
        </div>
    </Dialog>
</template>

<style scoped>
.eink-controls {
    display: flex;
    flex-wrap: wrap;
    gap: 1rem;
    align-items: flex-end;
}

.eink-controls :deep(.eink-size) {
    width: 5.5rem;
}

.eink-actual {
    height: 2.5rem;
}

.eink-hint {
    display: block;
    margin: 0.75rem 0;
    color: var(--p-text-muted-color);
}

.eink-stage {
    display: flex;
    justify-content: center;
    align-items: flex-start;
    min-height: 200px;
    max-height: 70vh;
    overflow: auto;
    padding: 0.75rem;
    border: 1px solid var(--p-surface-border);
    border-radius: 6px;
    background: repeating-conic-gradient(var(--p-surface-100) 0% 25%, var(--p-surface-0) 0% 50%) 50% / 16px 16px;
}

.eink-stage img {
    max-width: 100%;
    height: auto;
    box-shadow: 0 1px 6px rgba(0, 0, 0, 0.25);
}

.eink-stage-actual {
    justify-content: flex-start;
}

.eink-stage-actual img {
    max-width: none;
    image-rendering: pixelated;
}

.eink-loading {
    opacity: 0.5;
}

.eink-placeholder,
.eink-error {
    align-self: center;
    display: inline-flex;
    gap: 0.5rem;
    align-items: center;
    color: var(--p-text-muted-color);
}

.eink-error {
    color: var(--p-red-500);
}

.eink-spin {
    animation: eink-spin 1s linear infinite;
}

@keyframes eink-spin {
    to {
        transform: rotate(360deg);
    }
}
</style>
