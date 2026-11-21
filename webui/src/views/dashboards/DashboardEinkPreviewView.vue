<script setup lang="ts">
import { ref, computed, watch, onBeforeUnmount } from 'vue'
import { useRoute, useRouter, type LocationQuery } from 'vue-router'
import Button from 'primevue/button'
import Select from 'primevue/select'
import InputNumber from 'primevue/inputnumber'
import Checkbox from 'primevue/checkbox'
import { apiErrorMessage } from '@/lib/api/client'
import { useGetDashboard } from '@/composables/useDashboards'
import {
    CUSTOM_DEVICE,
    DEVICE_PRESETS,
    MAX_PREVIEW_SIDE,
    PREVIEW_FORMATS,
    PREVIEW_ROTATIONS,
    clampSide,
    einkPreviewUrl,
    parsePreviewQuery,
    previewQuery,
    uprightView,
    type PreviewFormat,
    type PreviewRotation,
} from '@/lib/einkPreview'
import dashiIcon from '@/assets/icon-64.png'

// Shows the server-rendered PNG an e-ink panel would receive for this
// dashboard. Every control is local to the page and mirrored in the query
// string (router.replace, so Back is unaffected); nothing is written to the
// dashboard. Opened from the editor toolbar (which saves pending edits first)
// and from the admin dashboards list.
const route = useRoute()
const router = useRouter()
const id = computed(() => route.params.id as string)

const { data: dashboard, isLoading, isError } = useGetDashboard(() => id.value)

const initial = parsePreviewQuery(route.query)
const device = ref<string>(initial.device)
const width = ref<number>(initial.width)
const height = ref<number>(initial.height)
const format = ref<PreviewFormat>(initial.format)
const rotation = ref<PreviewRotation>(initial.rotation)
const page = ref(initial.page)
const actualSize = ref(initial.actualSize)

// Where the page was opened from, so Back has somewhere to go when there is
// no in-app history (a reload keeps it, a bookmark or new tab does not).
const fromQuery = route.query.from
const from = fromQuery === 'edit' || fromQuery === 'list' ? fromQuery : undefined

const isImage = computed(() => dashboard.value?.type === 'image')
const pageNames = computed(() => (dashboard.value?.pages ?? []).map(p => p.name))

const deviceOptions = [
    ...DEVICE_PRESETS.map(p => ({ label: p.label, value: p.key })),
    { label: 'Custom size', value: CUSTOM_DEVICE },
]
const rotationOptions = PREVIEW_ROTATIONS.map(r => ({ label: `${r}°`, value: r }))
const pageOptions = computed(() =>
    pageNames.value.map((name, i) => ({ label: name || `Page ${i + 1}`, value: i })),
)
const isCustom = computed(() => device.value === CUSTOM_DEVICE)

const url = computed(() =>
    einkPreviewUrl({
        dashboardId: id.value,
        width: clampSide(width.value),
        height: clampSide(height.value),
        format: format.value,
        rotation: rotation.value,
        page: page.value,
    }),
)

// --- Query string ----------------------------------------------------------

const settingsQuery = computed(() => {
    const q = previewQuery({
        device: device.value,
        width: width.value,
        height: height.value,
        format: format.value,
        rotation: rotation.value,
        page: page.value,
        actualSize: actualSize.value,
    })
    if (from) q.from = from
    return q
})

const sameQuery = (a: Record<string, string>, b: LocationQuery) => {
    const keys = Object.keys(a)
    return keys.length === Object.keys(b).length && keys.every(k => b[k] === a[k])
}

// Also runs once on load, so an invalid or partial query is rewritten to
// what is actually shown.
watch(settingsQuery, (q) => {
    if (sameQuery(q, route.query)) return
    router.replace({ name: 'dashboard-eink-preview', params: { id: id.value }, query: q })
}, { immediate: true })

// --- Rendering -------------------------------------------------------------

const imageSrc = ref<string | null>(null)
// Size and rotation of the shown image, kept with it so the frame does not
// change shape while the next render loads.
const rendered = ref<{ width: number; height: number; rotation: PreviewRotation } | null>(null)
const loading = ref(false)
const error = ref<string | null>(null)
let requestSeq = 0
// The URL of the newest request, so a control change that lands on the same
// render does not fetch it twice.
let requestedUrl = ''

const setImage = (src: string | null) => {
    if (imageSrc.value) URL.revokeObjectURL(imageSrc.value)
    imageSrc.value = src
    if (!src) rendered.value = null
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
    const size = { width: clampSide(width.value), height: clampSide(height.value), rotation: rotation.value }
    requestedUrl = target
    loading.value = true
    error.value = null
    try {
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
        rendered.value = size
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

// A page from the URL that the dashboard no longer has falls back to the
// first one.
const clampPage = () => {
    const count = pageNames.value.length
    if (count > 0 && page.value > count - 1) page.value = 0
}
watch(() => pageNames.value.length, clampPage)

// Re-render whenever the request changes, after a short pause so typing a
// custom size does not render once per keystroke.
const RELOAD_DELAY_MS = 350
let reloadTimer: ReturnType<typeof setTimeout> | undefined
const cancelReload = () => {
    if (reloadTimer !== undefined) clearTimeout(reloadTimer)
    reloadTimer = undefined
}
watch(url, (next) => {
    if (!isImage.value || next === requestedUrl) return
    cancelReload()
    reloadTimer = setTimeout(() => {
        reloadTimer = undefined
        load()
    }, RELOAD_DELAY_MS)
})

// First render as soon as the dashboard is known to be an image dashboard;
// an interactive one answers a display request with HTML, so it gets a
// message instead.
watch(isImage, (image) => {
    if (!image) return
    clampPage()
    load()
}, { immediate: true })

const refresh = () => {
    cancelReload()
    load()
}

onBeforeUnmount(() => {
    cancelReload()
    requestSeq++
    setImage(null)
})

// --- Upright display -------------------------------------------------------

// The panel image is turned back by the rotation it was rendered with, so a
// 90°/270° (portrait) dashboard reads upright. The frame takes the turned
// size, so the image never overflows it; see uprightView.
const view = computed(() => {
    const r = rendered.value
    return r ? uprightView(r.width, r.height, r.rotation) : null
})

const frameStyle = computed(() => {
    const r = rendered.value
    const v = view.value
    if (!r || !v) return {}
    return {
        '--disp-w': String(v.width),
        '--disp-h': String(v.height),
        '--img-w': String(r.width),
        '--img-h': String(r.height),
    }
})

// --- Navigation ------------------------------------------------------------

// vue-router records the previous in-app location in history.state.back; it
// is null when the page was opened directly, so fall back to the entry point
// it was linked from.
const goBack = () => {
    if (router.options.history.state.back) {
        router.back()
    } else if (from === 'edit') {
        router.push({ name: 'dashboard-edit', params: { id: id.value } })
    } else {
        router.push({ name: 'admin-dashboards' })
    }
}

const goToAdmin = () => router.push({ name: 'admin-dashboards' })
</script>

<template>
    <header class="app-topbar">
        <img :src="dashiIcon" alt="Dashi" class="app-topbar-icon" />
        <button type="button" class="app-topbar-title" @click="goToAdmin">Dashi</button>
    </header>
    <div class="eink-page">
        <div class="eink-toolbar">
            <Button
                icon="ti ti-arrow-left"
                severity="secondary"
                text
                rounded
                aria-label="Back"
                data-testid="eink-back"
                @click="goBack"
            />
            <span class="eink-title text-xl font-bold text-color">
                <template v-if="dashboard">{{ dashboard.name }} — </template>E-ink preview
            </span>
        </div>

        <div v-if="isLoading" class="p-4">Loading...</div>
        <div v-else-if="isError || !dashboard" class="p-4">Failed to load dashboard.</div>
        <div v-else-if="!isImage" class="eink-message" data-testid="eink-not-image">
            <i class="ti ti-info-circle" />
            E-ink preview is only available for image dashboards. Interactive dashboards are
            served as live HTML; change the type in the dashboard settings to render it for a panel.
        </div>

        <div v-else class="eink-body">
            <section class="eink-view">
                <div v-if="rendered" class="eink-meta">
                    <span data-testid="eink-size">{{ rendered.width }} × {{ rendered.height }} px</span>
                    <span v-if="rendered.rotation" data-testid="eink-upright-note">
                        <i class="ti ti-rotate-2" />
                        Shown upright — the panel receives it rotated {{ rendered.rotation }}°
                    </span>
                </div>
                <div class="eink-stage" :class="{ 'eink-stage-actual': actualSize }">
                    <div v-if="error" class="eink-status eink-error" role="alert" data-testid="eink-error">
                        <i class="ti ti-alert-circle" /> {{ error }}
                    </div>
                    <div v-else-if="!imageSrc && loading" class="eink-status">
                        <i class="ti ti-loader-2 eink-spin" /> Rendering…
                    </div>
                    <div
                        v-if="imageSrc && view && !error"
                        class="eink-frame"
                        :class="{ 'eink-loading': loading }"
                        :style="frameStyle"
                        data-testid="eink-frame"
                    >
                        <img
                            :src="imageSrc"
                            :style="{ transform: view.transform }"
                            alt="E-ink render of the dashboard"
                            data-testid="eink-image"
                        />
                    </div>
                </div>
            </section>

            <aside class="eink-panel">
                <div class="flex flex-column gap-1">
                    <label class="font-semibold text-sm">Device</label>
                    <Select
                        :modelValue="device"
                        @update:modelValue="(v: string | undefined) => { if (v !== undefined) applyDevice(v) }"
                        :options="deviceOptions"
                        optionLabel="label"
                        optionValue="value"
                        class="w-full"
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
                        class="w-full"
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
                        class="w-full"
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
                        class="w-full"
                        data-testid="eink-page"
                    />
                </div>
                <div class="flex align-items-center gap-2">
                    <Checkbox v-model="actualSize" inputId="eink-actual" binary data-testid="eink-actual" />
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
                <small class="eink-hint">
                    Rendered by the server exactly as a panel requests it, dithered with the
                    dashboard's Dither setting. With 90° or 270° the page is laid out in portrait
                    and turned onto the panel; the preview turns it back so it reads upright.
                    Nothing here is saved.
                </small>
            </aside>
        </div>
    </div>
</template>

<style scoped lang="scss">
@use '@/assets/breakpoints' as bp;

/* Fills the window under the topbar (3rem), so the image gets all the room
   the controls leave. Below md the page scrolls instead. */
.eink-page {
    display: flex;
    flex-direction: column;
    height: calc(100dvh - 3rem);
    padding: 1rem;
    gap: 0.75rem;

    @include bp.below-md {
        height: auto;
        padding: 1rem 0.5rem;
    }
}

.eink-toolbar {
    flex: none;
    display: flex;
    align-items: center;
    gap: 0.5rem;
}

.eink-title {
    min-width: 0;
    overflow-wrap: break-word;
}

.eink-message {
    display: flex;
    gap: 0.5rem;
    align-items: flex-start;
    max-width: 40rem;
    padding: 1rem;
    color: var(--p-text-muted-color);
}

.eink-body {
    flex: 1;
    min-height: 0;
    display: flex;
    gap: 1rem;

    @include bp.below-md {
        flex-direction: column;
    }
}

.eink-view {
    flex: 1;
    min-width: 0;
    min-height: 0;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
}

.eink-meta {
    flex: none;
    display: flex;
    flex-wrap: wrap;
    gap: 0.25rem 1rem;
    font-size: 0.85rem;
    color: var(--p-text-muted-color);
}

/* A size container: the fitted frame below measures itself against it. */
.eink-stage {
    flex: 1;
    min-height: 0;
    display: flex;
    overflow: auto;
    padding: 0.75rem;
    container-type: size;
    border: 1px solid var(--p-surface-border);
    border-radius: 6px;
    background: repeating-conic-gradient(var(--p-surface-100) 0% 25%, var(--p-surface-0) 0% 50%) 50% / 16px 16px;

    /* Stacked: the stage takes the height of the image, fitted to the width. */
    @include bp.below-md {
        flex: none;
        min-height: 200px;
        container-type: inline-size;
    }
}

/* The box of the upright image. Fitted: as large as the stage allows, never
   upscaled. Auto margins centre it and still let an overflowing (actual
   pixels) frame scroll to its top-left corner. */
.eink-frame {
    position: relative;
    flex: none;
    margin: auto;
    aspect-ratio: var(--disp-w) / var(--disp-h);
    width: min(calc(var(--disp-w) * 1px), 100cqw, calc(100cqh * var(--disp-w) / var(--disp-h)));
    box-shadow: 0 1px 6px rgba(0, 0, 0, 0.25);

    @include bp.below-md {
        width: min(calc(var(--disp-w) * 1px), 100cqw);
    }
}

/* The panel image, turned upright about its top-left corner (see uprightView);
   for 90°/270° its width is the frame's height and vice versa. */
.eink-frame img {
    position: absolute;
    top: 0;
    left: 0;
    max-width: none;
    width: calc(100% * var(--img-w) / var(--disp-w));
    height: calc(100% * var(--img-h) / var(--disp-h));
    transform-origin: 0 0;
}

.eink-stage-actual .eink-frame {
    width: calc(var(--disp-w) * 1px);
    height: calc(var(--disp-h) * 1px);
}

.eink-stage-actual .eink-frame img {
    width: calc(var(--img-w) * 1px);
    height: calc(var(--img-h) * 1px);
    image-rendering: pixelated;
}

.eink-loading {
    opacity: 0.5;
}

.eink-status {
    margin: auto;
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

.eink-panel {
    flex: none;
    width: 300px;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 1rem;
    padding: 0.25rem 0.25rem 0.25rem 1rem;
    border-left: 1px solid var(--p-surface-border);

    @include bp.below-md {
        width: auto;
        overflow: visible;
        padding: 1rem 0 0;
        border-left: 0;
        border-top: 1px solid var(--p-surface-border);
    }
}

.eink-panel :deep(.eink-size) {
    width: 5.5rem;
}

.eink-hint {
    color: var(--p-text-muted-color);
}
</style>
