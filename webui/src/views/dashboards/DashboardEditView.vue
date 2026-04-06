<script setup lang="ts">
import { ref, computed, watch, provide } from 'vue'
import { DASHBOARD_THEME, DASHBOARD_ID, ACTIVE_PAGE, TOTAL_PAGES, EDITING_MODE } from '@/lib/injectionKeys'
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import DashboardRow from '@/components/dashboards/DashboardRow.vue'

import { useGetDashboard, useUpdateDashboard, usePreviewDashboard, useBackgrounds } from '@/composables/useDashboards'
import { useThemes } from '@/composables/useThemes'
import { useToast } from 'primevue/usetoast'
import type { Dashboard, Row, Page, Background } from '@/types/dashboard'
import { onBeforeUnmount } from 'vue'
import { v4 as uuidv4 } from 'uuid'
import Dialog from 'primevue/dialog'
import dashiIcon from '@/assets/icon-64.png'
import Select from 'primevue/select'
import ColorPicker from 'primevue/colorpicker'
import Checkbox from 'primevue/checkbox'
import type { BackgroundOption } from '@/lib/api/dashboard'

const route = useRoute()
const router = useRouter()
const toast = useToast()
const id = computed(() => route.params.id as string)

const { data: serverDashboard, isLoading, isError } = useGetDashboard(() => id.value)
const { updateDashboard, isUpdating } = useUpdateDashboard()
const { createPreview, updatePreview, deletePreview } = usePreviewDashboard()
const { data: backgroundsData } = useBackgrounds(() => id.value)

const { data: themesData } = useThemes()
const themeOptions = computed(() => {
    if (!themesData.value) return []
    return themesData.value.map(t => ({ label: t.name, value: t.name }))
})

const backgroundImageOptions = computed(() => {
    const groups: { label: string; items: { label: string; value: string }[] }[] = []
    if (backgroundOptions.value.theme.length > 0) {
        groups.push({
            label: 'Theme',
            items: backgroundOptions.value.theme.map(o => ({ label: o.name, value: o.value })),
        })
    }
    if (backgroundOptions.value.dashboard.length > 0) {
        groups.push({
            label: 'Dashboard',
            items: backgroundOptions.value.dashboard.map(o => ({ label: o.name, value: o.value })),
        })
    }
    return groups
})

const ensureBackground = () => {
    if (!localDashboard.value) return
    if (!localDashboard.value.background) {
        localDashboard.value.background = { type: 'none', value: '' }
    }
}

const backgroundType = computed({
    get: (): Background['type'] => localDashboard.value?.background?.type ?? 'none',
    set: (v: Background['type']) => {
        ensureBackground()
        localDashboard.value!.background!.type = v
        if (v === 'gradient') {
            localDashboard.value!.background!.value = gradientCSS.value
        } else {
            localDashboard.value!.background!.value = ''
        }
    }
})

const backgroundValue = computed({
    get: () => localDashboard.value?.background?.value ?? '',
    set: (v: string) => {
        ensureBackground()
        localDashboard.value!.background!.value = v
    }
})

// Gradient editor state
const gradientDirection = ref('to right')
const gradientCustomAngle = ref('135')
const gradientColor1 = ref('#667eea')
const gradientColor2 = ref('#764ba2')

const gradientDirectionOptions = [
    { label: 'To Right', value: 'to right' },
    { label: 'To Left', value: 'to left' },
    { label: 'To Bottom', value: 'to bottom' },
    { label: 'To Top', value: 'to top' },
    { label: 'To Bottom Right', value: 'to bottom right' },
    { label: 'To Top Right', value: 'to top right' },
    { label: 'Custom Angle', value: 'custom' },
]

const gradientDirectionCSS = computed(() => {
    return gradientDirection.value === 'custom'
        ? gradientCustomAngle.value + 'deg'
        : gradientDirection.value
})

const gradientCSS = computed(() => {
    return `linear-gradient(${gradientDirectionCSS.value}, ${gradientColor1.value}, ${gradientColor2.value})`
})

function parseGradientValue(val: string) {
    const match = val.match(/^linear-gradient\((.+?),\s*(.+?),\s*(.+?)\)$/)
    if (!match) return
    const dir = match[1].trim()
    gradientColor1.value = match[2].trim()
    gradientColor2.value = match[3].trim()
    if (dir.endsWith('deg')) {
        gradientDirection.value = 'custom'
        gradientCustomAngle.value = dir.replace('deg', '')
    } else {
        gradientDirection.value = dir
    }
}

function syncGradientToBackground() {
    backgroundValue.value = gradientCSS.value
}

const localDashboard = ref<Dashboard | null>(null)
const backgroundOptions = computed(() => backgroundsData.value ?? { theme: [] as BackgroundOption[], dashboard: [] as BackgroundOption[] })
const containerDialogVisible = ref(false)
const activePageIndex = ref(0)

const dashboardTheme = computed(() => localDashboard.value?.theme || 'default')
provide(DASHBOARD_THEME, dashboardTheme)
provide(DASHBOARD_ID, id)
const editTotalPages = computed(() => localDashboard.value ? localDashboard.value.pages.length : 1)
provide(ACTIVE_PAGE, activePageIndex)
provide(TOTAL_PAGES, editTotalPages)
provide(EDITING_MODE, true)
const renamePageDialogVisible = ref(false)
const renamePageIndex = ref(0)
const renamePageName = ref('')

watch(serverDashboard, (val) => {
    if (val && !localDashboard.value) {
        localDashboard.value = JSON.parse(JSON.stringify(val))
        if (val.background?.type === 'gradient' && val.background.value) {
            parseGradientValue(val.background.value)
        }
    }
}, { immediate: true })

watch(() => localDashboard.value?.type, (newType) => {
    if (!localDashboard.value) return
    if (newType === 'image' && !localDashboard.value.imageConfig) {
        localDashboard.value.imageConfig = { width: 1024, height: 0 }
    }
})

const pages = computed(() => localDashboard.value?.pages ?? [])
const activePage = computed(() => pages.value[activePageIndex.value])
const rows = computed(() => activePage.value?.rows ?? [])

const addPage = () => {
    if (!localDashboard.value) return
    localDashboard.value.pages.push({
        name: '',
        rows: []
    })
    activePageIndex.value = localDashboard.value.pages.length - 1
}

const deletePage = (index: number) => {
    if (!localDashboard.value) return
    const page = localDashboard.value.pages[index]
    if (page.rows.length > 0) {
        if (!confirm('This page has rows. Are you sure you want to delete it?')) {
            return
        }
    }
    localDashboard.value.pages.splice(index, 1)
    if (activePageIndex.value >= localDashboard.value.pages.length) {
        activePageIndex.value = Math.max(0, localDashboard.value.pages.length - 1)
    }
}

const movePageUp = (index: number) => {
    if (!localDashboard.value || index <= 0) return
    const pages = localDashboard.value.pages
    ;[pages[index - 1], pages[index]] = [pages[index], pages[index - 1]]
    if (activePageIndex.value === index) {
        activePageIndex.value = index - 1
    } else if (activePageIndex.value === index - 1) {
        activePageIndex.value = index
    }
}

const movePageDown = (index: number) => {
    if (!localDashboard.value) return
    const pages = localDashboard.value.pages
    if (index >= pages.length - 1) return
    ;[pages[index], pages[index + 1]] = [pages[index + 1], pages[index]]
    if (activePageIndex.value === index) {
        activePageIndex.value = index + 1
    } else if (activePageIndex.value === index + 1) {
        activePageIndex.value = index
    }
}

const openRenamePage = (index: number) => {
    renamePageIndex.value = index
    renamePageName.value = localDashboard.value?.pages[index]?.name ?? ''
    renamePageDialogVisible.value = true
}

const confirmRenamePage = () => {
    if (!localDashboard.value) return
    localDashboard.value.pages[renamePageIndex.value].name = renamePageName.value
    renamePageDialogVisible.value = false
}

const addRow = () => {
    if (!activePage.value) return
    activePage.value.rows.push({
        id: uuidv4(),
        height: 'auto',
        width: '100%',
        widgets: []
    })
}

const updateRow = (index: number, row: Row) => {
    if (!activePage.value) return
    activePage.value.rows[index] = row
}

const deleteRow = (index: number) => {
    if (!activePage.value) return
    activePage.value.rows.splice(index, 1)
}

const moveRowUp = (index: number) => {
    if (!activePage.value || index <= 0) return
    const rows = activePage.value.rows
    ;[rows[index - 1], rows[index]] = [rows[index], rows[index - 1]]
}

const moveRowDown = (index: number) => {
    if (!activePage.value) return
    const rows = activePage.value.rows
    if (index >= rows.length - 1) return
    ;[rows[index], rows[index + 1]] = [rows[index + 1], rows[index]]
}

const save = async () => {
    if (!localDashboard.value) return
    try {
        await updateDashboard({ id: id.value, payload: localDashboard.value })
        toast.add({ severity: 'success', summary: 'Saved', detail: 'Dashboard saved successfully', life: 3000 })
    } catch (err) {
        toast.add({ severity: 'error', summary: 'Error', detail: 'Failed to save dashboard', life: 5000 })
    }
}

const cancel = () => {
    router.push({ name: 'dashboards' })
}

const previewId = ref<string | null>(null)

const isPreviewing = ref(false)

const preview = async () => {
    if (!localDashboard.value) return
    isPreviewing.value = true
    try {
        const prevId = id.value + '-prev'
        const previewPayload = {
            id: prevId,
            name: localDashboard.value.name + ' - preview',
            icon: localDashboard.value.icon,
            type: localDashboard.value.type,
            container: JSON.parse(JSON.stringify(localDashboard.value.container)),
            imageConfig: localDashboard.value.imageConfig
                ? JSON.parse(JSON.stringify(localDashboard.value.imageConfig))
                : undefined,
            theme: localDashboard.value.theme,
            colorMode: localDashboard.value.colorMode,
            background: localDashboard.value.background
                ? JSON.parse(JSON.stringify(localDashboard.value.background))
                : undefined,
            pages: JSON.parse(JSON.stringify(localDashboard.value.pages))
        }
        if (previewId.value) {
            await updatePreview({ id: prevId, payload: { ...previewPayload } as Dashboard })
        } else {
            await createPreview(previewPayload)
            previewId.value = prevId
        }
        const resolved = router.resolve({ name: 'dashboard-view', params: { id: previewId.value! } })
        window.open(resolved.href, '_blank')
    } catch (err) {
        console.error('Preview failed:', err)
        toast.add({ severity: 'error', summary: 'Error', detail: 'Failed to create preview', life: 5000 })
    } finally {
        isPreviewing.value = false
    }
}

const cleanupPreview = async () => {
    if (previewId.value) {
        try {
            await deletePreview(previewId.value)
        } catch {
            // ignore cleanup errors
        }
        previewId.value = null
    }
}

onBeforeUnmount(cleanupPreview)
</script>

<template>
    <header class="app-topbar">
        <img :src="dashiIcon" alt="Dashi" class="app-topbar-icon" />
        <span class="app-topbar-title" @click="router.push('/dashboards')">Dashi</span>
    </header>
    <div class="dashboard-edit-view">
        <div v-if="isLoading" class="p-4">Loading...</div>
        <div v-else-if="isError" class="p-4">Failed to load dashboard.</div>
        <template v-else-if="localDashboard">
        <div class="flex align-items-center gap-2 mb-3">
            <Button
                icon="ti ti-arrow-left"
                severity="secondary"
                text
                rounded
                @click="router.push({ name: 'dashboards' })"
            />
            <span class="text-xl font-bold text-color flex-grow-1">{{ localDashboard.name }}</span>
            <Button
                icon="ti ti-settings"
                label="Settings"
                severity="secondary"
                @click="containerDialogVisible = true"
            />
            <Button
                icon="ti ti-eye"
                label="Preview"
                severity="secondary"
                :loading="isPreviewing"
                @click="preview"
            />
            <Button
                label="Save"
                icon="ti ti-check"
                :loading="isUpdating"
                @click="save"
            />
            <Button
                label="Cancel"
                icon="ti ti-x"
                severity="secondary"
                @click="cancel"
            />
        </div>

        <div class="page-tabs">
            <div
                v-for="(page, index) in pages"
                :key="index"
                class="page-tab"
                :class="{ active: index === activePageIndex }"
                @click="activePageIndex = index"
            >
                <span>{{ page.name || `Page ${index + 1}` }}</span>
                <div v-if="index === activePageIndex" class="page-tab-actions">
                    <Button
                        icon="ti ti-pencil"
                        severity="secondary"
                        text
                        size="small"
                        @click.stop="openRenamePage(index)"
                    />
                    <Button
                        icon="ti ti-arrow-left"
                        severity="secondary"
                        text
                        size="small"
                        :disabled="index === 0"
                        @click.stop="movePageUp(index)"
                    />
                    <Button
                        icon="ti ti-arrow-right"
                        severity="secondary"
                        text
                        size="small"
                        :disabled="index === pages.length - 1"
                        @click.stop="movePageDown(index)"
                    />
                    <Button
                        icon="ti ti-trash"
                        severity="danger"
                        text
                        size="small"
                        :disabled="pages.length === 1"
                        @click.stop="deletePage(index)"
                    />
                </div>
            </div>
            <Button
                icon="ti ti-plus"
                severity="secondary"
                text
                size="small"
                label="Add Page"
                class="add-page-btn"
                @click="addPage"
            />
        </div>

        <DashboardRow
            v-for="(row, index) in rows"
            :key="row.id"
            :row="row"
            :is-first="index === 0"
            :is-last="index === rows.length - 1"
            @update="updateRow(index, $event)"
            @delete="deleteRow(index)"
            @move-up="moveRowUp(index)"
            @move-down="moveRowDown(index)"
        />

        <div class="mt-2">
            <Button
                label="Add Row"
                icon="ti ti-plus"
                severity="secondary"
                @click="addRow"
            />
        </div>

        <Dialog
            v-model:visible="containerDialogVisible"
            modal
            :closable="true"
            :draggable="false"
            header="Dashboard Settings"
        >
            <div class="flex flex-column gap-3" style="min-width: 350px">
                <div class="flex flex-column gap-1">
                    <label class="font-semibold text-sm">Name</label>
                    <InputText v-model="localDashboard.name" placeholder="Dashboard name" />
                </div>
                <div class="flex align-items-center gap-2">
                    <Checkbox v-model="localDashboard.default" :binary="true" inputId="dashboardDefault" />
                    <label for="dashboardDefault" class="font-semibold text-sm">Default dashboard</label>
                </div>
                <div class="flex flex-column gap-1">
                    <label class="font-semibold text-sm">Type</label>
                    <Select
                        v-model="localDashboard.type"
                        :options="[
                            { label: 'Interactive', value: 'interactive' },
                            { label: 'Image', value: 'image' },
                        ]"
                        optionLabel="label"
                        optionValue="value"
                        class="w-full"
                    />
                </div>
                <div class="flex flex-column gap-1">
                    <label class="font-semibold text-sm">Theme</label>
                    <Select
                        :modelValue="localDashboard.theme || 'default'"
                        @update:modelValue="(v: string | undefined) => { if (localDashboard && v !== undefined) localDashboard.theme = v }"
                        :options="themeOptions"
                        optionLabel="label"
                        optionValue="value"
                        class="w-full"
                    />
                </div>
                <div class="flex flex-column gap-1">
                    <label class="font-semibold text-sm">Color Mode</label>
                    <Select
                        :modelValue="localDashboard.colorMode || 'auto'"
                        @update:modelValue="(v: string | undefined) => { if (localDashboard && v !== undefined) localDashboard.colorMode = v as any }"
                        :options="[
                            { label: 'Auto', value: 'auto' },
                            { label: 'Light', value: 'light' },
                            { label: 'Dark', value: 'dark' },
                        ]"
                        optionLabel="label"
                        optionValue="value"
                        class="w-full"
                    />
                </div>
                <div class="flex flex-column gap-1">
                    <label class="font-semibold text-sm">Accent Color</label>
                    <div class="flex align-items-center gap-2">
                        <ColorPicker
                            :modelValue="localDashboard!.accentColor?.replace('#', '') || '3B82F6'"
                            @update:modelValue="(v: string | undefined) => { if (localDashboard && v !== undefined) localDashboard.accentColor = '#' + v }"
                        />
                        <InputText
                            :modelValue="localDashboard!.accentColor || '#3B82F6'"
                            @update:modelValue="(v: string | undefined) => { if (localDashboard && v !== undefined) localDashboard.accentColor = v }"
                            class="flex-1"
                            placeholder="#3B82F6"
                        />
                    </div>
                </div>
                <div class="flex flex-column gap-1">
                    <label class="font-semibold text-sm">Max Width</label>
                    <InputText v-model="localDashboard.container.maxWidth" placeholder="e.g. 1200px, 80%, 100%" />
                </div>
                <div class="flex flex-column gap-1">
                    <label class="font-semibold text-sm">Vertical Align</label>
                    <Select
                        v-model="localDashboard.container.verticalAlign"
                        :options="[
                            { label: 'Top', value: 'top' },
                            { label: 'Center', value: 'center' },
                            { label: 'Bottom', value: 'bottom' },
                        ]"
                        optionLabel="label"
                        optionValue="value"
                        class="w-full"
                    />
                </div>
                <div class="flex flex-column gap-1">
                    <label class="font-semibold text-sm">Horizontal Align</label>
                    <Select
                        v-model="localDashboard.container.horizontalAlign"
                        :options="[
                            { label: 'Left', value: 'left' },
                            { label: 'Center', value: 'center' },
                            { label: 'Right', value: 'right' },
                        ]"
                        optionLabel="label"
                        optionValue="value"
                        class="w-full"
                    />
                </div>
                <div class="flex flex-column gap-1">
                    <label class="font-semibold text-sm">Background</label>
                    <Select
                        :modelValue="backgroundType"
                        @update:modelValue="(v: Background['type'] | undefined) => { if (v !== undefined) backgroundType = v }"
                        :options="[
                            { label: 'None', value: 'none' },
                            { label: 'Image', value: 'image' },
                            { label: 'Color', value: 'color' },
                            { label: 'Gradient', value: 'gradient' },
                        ]"
                        optionLabel="label"
                        optionValue="value"
                        class="w-full"
                    />
                </div>
                <div v-if="backgroundType === 'image'" class="flex flex-column gap-1">
                    <label class="font-semibold text-sm">Background Image</label>
                    <Select
                        :modelValue="backgroundValue"
                        @update:modelValue="(v: string | undefined) => { if (v !== undefined) backgroundValue = v }"
                        :options="backgroundImageOptions"
                        optionLabel="label"
                        optionValue="value"
                        optionGroupLabel="label"
                        optionGroupChildren="items"
                        placeholder="Select an image..."
                        class="w-full"
                    />
                </div>
                <div v-if="backgroundType === 'color'" class="flex flex-column gap-1">
                    <label class="font-semibold text-sm">Background Color</label>
                    <div class="flex align-items-center gap-2">
                        <ColorPicker
                            :modelValue="backgroundValue.replace('#', '')"
                            @update:modelValue="(v: string | undefined) => { if (v !== undefined) backgroundValue = '#' + v }"
                        />
                        <InputText
                            :modelValue="backgroundValue"
                            @update:modelValue="(v: string | undefined) => { if (v !== undefined) backgroundValue = v }"
                            placeholder="#1a1a2e"
                            class="flex-grow-1"
                        />
                    </div>
                </div>
                <template v-if="backgroundType === 'gradient'">
                    <div class="flex flex-column gap-1">
                        <label class="font-semibold text-sm">Direction</label>
                        <Select
                            v-model="gradientDirection"
                            :options="gradientDirectionOptions"
                            optionLabel="label"
                            optionValue="value"
                            class="w-full"
                            @update:modelValue="syncGradientToBackground"
                        />
                    </div>
                    <div v-if="gradientDirection === 'custom'" class="flex flex-column gap-1">
                        <label class="font-semibold text-sm">Angle (degrees)</label>
                        <InputText
                            v-model="gradientCustomAngle"
                            placeholder="135"
                            class="w-full"
                            @update:modelValue="syncGradientToBackground"
                        />
                    </div>
                    <div class="flex flex-column gap-1">
                        <label class="font-semibold text-sm">Start Color</label>
                        <div class="flex align-items-center gap-2">
                            <ColorPicker
                                :modelValue="gradientColor1.replace('#', '')"
                                @update:modelValue="(v: string | undefined) => { if (v !== undefined) { gradientColor1 = '#' + v; syncGradientToBackground() } }"
                            />
                            <InputText
                                v-model="gradientColor1"
                                placeholder="#667eea"
                                class="flex-grow-1"
                                @update:modelValue="syncGradientToBackground"
                            />
                        </div>
                    </div>
                    <div class="flex flex-column gap-1">
                        <label class="font-semibold text-sm">End Color</label>
                        <div class="flex align-items-center gap-2">
                            <ColorPicker
                                :modelValue="gradientColor2.replace('#', '')"
                                @update:modelValue="(v: string | undefined) => { if (v !== undefined) { gradientColor2 = '#' + v; syncGradientToBackground() } }"
                            />
                            <InputText
                                v-model="gradientColor2"
                                placeholder="#764ba2"
                                class="flex-grow-1"
                                @update:modelValue="syncGradientToBackground"
                            />
                        </div>
                    </div>
                    <div
                        class="gradient-preview"
                        :style="{ background: gradientCSS }"
                    />
                </template>
                <template v-if="localDashboard!.type === 'image'">
                    <div class="flex flex-column gap-1">
                        <label class="font-semibold text-sm">Image Width (px)</label>
                        <InputText
                            :modelValue="String(localDashboard!.imageConfig?.width ?? 1024)"
                            @update:modelValue="(v: string | undefined) => {
                                if (!localDashboard || v === undefined) return
                                if (!localDashboard.imageConfig) localDashboard.imageConfig = { width: 1024, height: 0 }
                                localDashboard.imageConfig.width = parseInt(v) || 1024
                            }"
                            placeholder="1024"
                        />
                    </div>
                    <div class="flex flex-column gap-1">
                        <label class="font-semibold text-sm">Image Height (px, 0 = auto)</label>
                        <InputText
                            :modelValue="String(localDashboard!.imageConfig?.height ?? 0)"
                            @update:modelValue="(v: string | undefined) => {
                                if (!localDashboard || v === undefined) return
                                if (!localDashboard.imageConfig) localDashboard.imageConfig = { width: 1024, height: 0 }
                                localDashboard.imageConfig.height = parseInt(v) || 0
                            }"
                            placeholder="0 (auto)"
                        />
                    </div>
                </template>
            </div>
            <div class="flex justify-content-end mt-4">
                <Button label="Done" icon="ti ti-check" @click="containerDialogVisible = false" />
            </div>
        </Dialog>

        <Dialog
            v-model:visible="renamePageDialogVisible"
            modal
            :closable="true"
            :draggable="false"
            header="Rename Page"
        >
            <div class="flex flex-column gap-3" style="min-width: 350px">
                <div class="flex flex-column gap-1">
                    <label class="font-semibold text-sm">Page Name</label>
                    <InputText v-model="renamePageName" placeholder="Page name" />
                </div>
            </div>
            <div class="flex justify-content-end gap-2 mt-4">
                <Button label="Cancel" severity="secondary" @click="renamePageDialogVisible = false" />
                <Button label="Confirm" icon="ti ti-check" @click="confirmRenamePage" />
            </div>
        </Dialog>
        </template>

    </div>
</template>

<style scoped>
.dashboard-edit-view {
    max-width: 1600px;
    margin: 0 auto;
    padding: 1.5rem 1rem;
}

.page-tabs {
    display: flex;
    align-items: center;
    border-bottom: 1px solid var(--p-surface-border);
    margin-bottom: 1.5rem;
    gap: 0.5rem;
}

.page-tab {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.75rem 1rem;
    cursor: pointer;
    border-bottom: 2px solid transparent;
    transition: all 0.2s;
}

.page-tab:hover {
    background-color: var(--p-surface-100);
}

.page-tab.active {
    color: var(--p-primary-color);
    border-bottom-color: var(--p-primary-color);
    font-weight: 600;
}

.page-tab-actions {
    display: inline-flex;
    gap: 0.25rem;
}

.add-page-btn {
    margin-left: auto;
}

.gradient-preview {
    width: 100%;
    height: 24px;
    border-radius: 4px;
    border: 1px solid var(--p-surface-border);
}
</style>
