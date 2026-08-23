<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import ColorPicker from 'primevue/colorpicker'

import { useGetDashboard, useUpdateDashboard, useDashboardAuth } from '@/composables/useDashboards'
import { useListBackgrounds } from '@/composables/useBackgrounds'
import { useAutosave } from '@/composables/useAutosave'
import { useThemes } from '@/composables/useThemes'
import { pageBgValue } from '@/lib/backgroundCss'
import { useToast } from 'primevue/usetoast'
import type { Dashboard } from '@/types/dashboard'
import dashiIcon from '@/assets/icon-64.png'

const route = useRoute()
const router = useRouter()
const toast = useToast()
const id = computed(() => route.params.id as string)

const { data: serverDashboard, isLoading, isError } = useGetDashboard(() => id.value)
const { updateDashboard } = useUpdateDashboard()
const { backgrounds: backgroundList } = useListBackgrounds()
const { auth: dashAuth, isLoadingAuth, setAuth, isSettingAuth, deleteAuth, isDeletingAuth } = useDashboardAuth(() => id.value)

const authUsername = ref('')
const authPassword = ref('')

const { data: themesData } = useThemes()
const themeOptions = computed(() => {
    if (!themesData.value) return []
    return themesData.value.map(t => ({ label: t.name, value: t.name }))
})

const localDashboard = ref<Dashboard | null>(null)

// A dashboard no longer describes its own background; it points at a background
// entity, which is edited in the Backgrounds admin section.
const backgroundOptions = computed(() => [
    { label: 'None (theme background)', value: '' },
    ...(backgroundList.value ?? []).map(b => ({ label: b.name, value: b.id })),
])

const backgroundId = computed({
    get: () => localDashboard.value?.backgroundId ?? '',
    set: (v: string) => {
        if (localDashboard.value) localDashboard.value.backgroundId = v
    },
})

const selectedBackground = computed(
    () => (backgroundList.value ?? []).find(b => b.id === backgroundId.value) ?? null,
)

const swatchValue = computed(() =>
    selectedBackground.value ? pageBgValue(selectedBackground.value.previewCss) : 'transparent',
)

watch(serverDashboard, (val) => {
    if (val && !localDashboard.value) {
        localDashboard.value = JSON.parse(JSON.stringify(val))
    }
}, { immediate: true })

const { status: saveStatus, flush } = useAutosave<Dashboard>({
    source: () => localDashboard.value,
    save: (d) => updateDashboard({ id: id.value, payload: d }),
})

watch(saveStatus, (s) => {
    if (s === 'error') {
        toast.add({ severity: 'error', summary: 'Error', detail: 'Failed to save settings', life: 5000 })
    }
})

const goBack = async () => {
    await flush()
    if (window.history.length > 1) {
        router.back()
    } else {
        router.push({ name: 'dashboard-edit', params: { id: id.value } })
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
        <span class="app-topbar-title" @click="router.push('/admin')">Dashi</span>
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
                <span class="text-xl font-bold text-color flex-grow-1">{{ localDashboard.name }} — Settings</span>
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
                        <label class="font-semibold text-sm">Background</label>
                        <Select
                            :modelValue="backgroundId"
                            @update:modelValue="(v: string | undefined) => { if (v !== undefined) backgroundId = v }"
                            :options="backgroundOptions"
                            optionLabel="label"
                            optionValue="value"
                            placeholder="None (theme background)"
                            class="w-full"
                        />
                        <small class="settings-hint">
                            Backgrounds are shared between dashboards and edited in the
                            Backgrounds admin section.
                        </small>
                    </div>
                    <div v-if="selectedBackground" class="flex align-items-center gap-2">
                        <div class="bg-swatch" :style="{ background: swatchValue }" />
                        <Button
                            label="Edit background"
                            icon="ti ti-external-link"
                            text
                            @click="router.push({ name: 'admin-background-edit', params: { id: selectedBackground.id } })"
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
    max-width: 1200px;
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
