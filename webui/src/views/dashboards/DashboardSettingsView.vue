<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import ColorPicker from 'primevue/colorpicker'
import Checkbox from 'primevue/checkbox'

import { useGetDashboard, useUpdateDashboard, useBackgrounds, useDashboardAuth } from '@/composables/useDashboards'
import { useThemes } from '@/composables/useThemes'
import { useToast } from 'primevue/usetoast'
import type { Dashboard, Background } from '@/types/dashboard'
import type { BackgroundOption } from '@/lib/api/dashboard'
import dashiIcon from '@/assets/icon-64.png'

const route = useRoute()
const router = useRouter()
const toast = useToast()
const id = computed(() => route.params.id as string)

const { data: serverDashboard, isLoading, isError } = useGetDashboard(() => id.value)
const { updateDashboard, isUpdating } = useUpdateDashboard()
const { data: backgroundsData } = useBackgrounds(() => id.value)
const { auth: dashAuth, isLoadingAuth, setAuth, isSettingAuth, deleteAuth, isDeletingAuth } = useDashboardAuth(() => id.value)

const authUsername = ref('')
const authPassword = ref('')

const { data: themesData } = useThemes()
const themeOptions = computed(() => {
    if (!themesData.value) return []
    return themesData.value.map(t => ({ label: t.name, value: t.name }))
})

const backgroundOptions = computed(() => backgroundsData.value ?? { theme: [] as BackgroundOption[], dashboard: [] as BackgroundOption[] })

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

const localDashboard = ref<Dashboard | null>(null)

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

watch(serverDashboard, (val) => {
    if (val && !localDashboard.value) {
        localDashboard.value = JSON.parse(JSON.stringify(val))
        if (val.background?.type === 'gradient' && val.background.value) {
            parseGradientValue(val.background.value)
        }
    }
}, { immediate: true })

const goBack = () => {
    if (window.history.length > 1) {
        router.back()
    } else {
        router.push({ name: 'dashboard-edit', params: { id: id.value } })
    }
}

const save = async () => {
    if (!localDashboard.value) return
    try {
        await updateDashboard({ id: id.value, payload: localDashboard.value })
        toast.add({ severity: 'success', summary: 'Saved', detail: 'Settings saved successfully', life: 3000 })
        goBack()
    } catch (err) {
        toast.add({ severity: 'error', summary: 'Error', detail: 'Failed to save settings', life: 5000 })
    }
}

const sections = computed(() => {
    const items = [
        { key: 'general', label: 'General', icon: 'ti ti-settings' },
        { key: 'appearance', label: 'Appearance', icon: 'ti ti-palette' },
    ]
    items.push({ key: 'protection', label: 'Protection', icon: 'ti ti-lock' })
    return items
})

const activeSection = ref('general')
</script>

<template>
    <header class="app-topbar">
        <img :src="dashiIcon" alt="Dashi" class="app-topbar-icon" />
        <span class="app-topbar-title" @click="router.push('/dashboards')">Dashi</span>
    </header>
    <div class="settings-view">
        <div v-if="isLoading" class="p-4">Loading...</div>
        <div v-else-if="isError" class="p-4">Failed to load dashboard.</div>
        <template v-else-if="localDashboard">
            <div class="flex align-items-center gap-2 mb-4">
                <Button
                    icon="ti ti-arrow-left"
                    severity="secondary"
                    text
                    rounded
                    @click="goBack"
                />
                <span class="text-xl font-bold text-color">{{ localDashboard.name }} — Settings</span>
            </div>

            <div class="settings-layout">
                <nav class="settings-nav">
                    <div
                        v-for="section in sections"
                        :key="section.key"
                        class="settings-nav-item"
                        :class="{ active: activeSection === section.key }"
                        @click="activeSection = section.key"
                    >
                        <i :class="section.icon" />
                        <span>{{ section.label }}</span>
                    </div>
                </nav>

                <div class="settings-panel">

                <!-- General -->
                <div v-if="activeSection === 'general'" class="flex flex-column gap-3">
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
                    <div class="settings-actions">
                        <Button label="Save" icon="ti ti-check" :loading="isUpdating" @click="save" />
                        <Button label="Cancel" icon="ti ti-x" severity="secondary" @click="goBack" />
                    </div>
                </div>

                <!-- Appearance (theme, colors, container, background) -->
                <div v-if="activeSection === 'appearance'" class="flex flex-column gap-3">
                    <label class="settings-subsection">Theme & Colors</label>
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

                    <label class="settings-subsection">Container</label>
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

                    <label class="settings-subsection">Background</label>
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

                    <div class="settings-actions">
                        <Button label="Save" icon="ti ti-check" :loading="isUpdating" @click="save" />
                        <Button label="Cancel" icon="ti ti-x" severity="secondary" @click="goBack" />
                    </div>
                </div>

                <!-- Protection -->
                <div v-if="activeSection === 'protection'" class="flex flex-column gap-3">
                    <div v-if="isLoadingAuth" class="text-sm text-color-secondary">Loading...</div>
                    <template v-else-if="dashAuth?.enabled">
                        <div class="text-sm mb-2">
                            Protected as user: <strong>{{ dashAuth.username }}</strong>
                        </div>
                        <div class="flex flex-column gap-2">
                            <InputText v-model="authUsername" placeholder="Username" />
                            <InputText v-model="authPassword" type="password" placeholder="New password" />
                            <div class="flex gap-2">
                                <Button
                                    label="Update"
                                    icon="ti ti-check"
                                    severity="secondary"
                                    size="small"
                                    :loading="isSettingAuth"
                                    :disabled="!authUsername || !authPassword"
                                    @click="setAuth({ username: authUsername, password: authPassword }).then(() => { authUsername = ''; authPassword = '' })"
                                />
                                <Button
                                    label="Remove Protection"
                                    icon="ti ti-lock-open"
                                    severity="danger"
                                    size="small"
                                    outlined
                                    :loading="isDeletingAuth"
                                    @click="deleteAuth()"
                                />
                            </div>
                        </div>
                    </template>
                    <template v-else>
                        <div class="flex flex-column gap-2">
                            <InputText v-model="authUsername" placeholder="Username" />
                            <InputText v-model="authPassword" type="password" placeholder="Password" />
                            <Button
                                label="Set Protection"
                                icon="ti ti-lock"
                                severity="secondary"
                                size="small"
                                :loading="isSettingAuth"
                                :disabled="!authUsername || !authPassword"
                                @click="setAuth({ username: authUsername, password: authPassword }).then(() => { authUsername = ''; authPassword = '' })"
                            />
                        </div>
                    </template>
                    <div class="settings-actions">
                        <Button label="Back" icon="ti ti-arrow-left" severity="secondary" @click="goBack" />
                    </div>
                </div>

                </div>
            </div>
        </template>
    </div>
</template>

<style scoped>
.settings-view {
    max-width: 900px;
    margin: 0 auto;
    padding: 1.5rem 1rem;
}

.settings-layout {
    display: flex;
    gap: 2rem;
}

.settings-nav {
    position: sticky;
    top: 1rem;
    align-self: flex-start;
    min-width: 180px;
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
}

.settings-nav-item {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.5rem 0.75rem;
    border-radius: 6px;
    cursor: pointer;
    font-size: 0.875rem;
    color: var(--p-text-muted-color);
    transition: all 0.15s;
    white-space: nowrap;
}

.settings-nav-item:hover {
    background-color: var(--p-surface-100);
    color: var(--p-text-color);
}

.settings-nav-item.active {
    background-color: var(--p-primary-50);
    color: var(--p-primary-color);
    font-weight: 600;
}

.settings-nav-item i {
    font-size: 1.1rem;
    width: 1.25rem;
    text-align: center;
}

.settings-panel {
    flex: 1;
    min-width: 0;
}

.settings-subsection {
    font-size: 0.8rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--p-text-muted-color);
    margin-top: 0.75rem;
    padding-bottom: 0.25rem;
    border-bottom: 1px solid var(--p-surface-200);
}

.settings-actions {
    display: flex;
    gap: 0.5rem;
    margin-top: 1.5rem;
    padding-top: 1rem;
    border-top: 1px solid var(--p-surface-200);
}

.gradient-preview {
    width: 100%;
    height: 24px;
    border-radius: 4px;
    border: 1px solid var(--p-surface-border);
}
</style>
