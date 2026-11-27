<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import AlignmentPicker from '@/components/dashboards/AlignmentPicker.vue'

import { useDashboardAuth } from '@/composables/useDashboards'
import { useDashboardDraft } from '@/composables/useDashboardDraft'
import { useListThemes } from '@/composables/useThemes'
import { useWidgetThemes } from '@/composables/useWidgetThemes'
import { pageBgValue } from '@/lib/themeCss'
import { DITHER_OPTIONS, DITHER_PLAIN, type DitherOption } from '@/lib/dither'
import { useToast } from 'primevue/usetoast'
import type { ColorMode } from '@/types/dashboard'
import dashiIcon from '@/assets/icon-64.png'

const route = useRoute()
const router = useRouter()
const toast = useToast()
const id = computed(() => route.params.id as string)

const { draft: localDashboard, isLoading, isError, saveStatus, hasUnsaved, flush } = useDashboardDraft(() => id.value)
const { themes: themeList } = useListThemes()
const { data: widgetThemeList } = useWidgetThemes()
const { auth: dashAuth, isLoadingAuth, setAuth, isSettingAuth, deleteAuth, isDeletingAuth } = useDashboardAuth(() => id.value)

const authUsername = ref('')
const authPassword = ref('')

const protectionFailed = () => {
    toast.add({ severity: 'error', summary: 'Error', detail: 'Failed to update protection', life: 5000 })
}

const submitAuth = async () => {
    try {
        await setAuth({ username: authUsername.value, password: authPassword.value })
        authUsername.value = ''
        authPassword.value = ''
    } catch {
        protectionFailed()
    }
}

const removeAuth = async () => {
    try {
        await deleteAuth()
        authUsername.value = ''
        authPassword.value = ''
    } catch {
        protectionFailed()
    }
}

// A dashboard points at a theme (colours and background) by ID and at a widget
// theme (fonts and icons) by name. Both are edited in their admin sections; an
// empty reference renders the builtin "default".
//
// A dangling reference (the theme was deleted) also renders "default", so it
// is displayed as such. That is decided only once the list has loaded, so a
// valid reference never flashes "Default" first, and the stored value is left
// alone until the user picks another one.
const displayedRef = (stored: string | undefined, known: string[] | undefined): string => {
    const v = stored || 'default'
    return known && !known.includes(v) ? 'default' : v
}

const themeOptions = computed(() =>
    (themeList.value ?? []).map(t => ({ label: t.name, value: t.id })),
)

const themeId = computed({
    get: () => displayedRef(localDashboard.value?.themeId, themeList.value?.map(t => t.id)),
    set: (v: string) => {
        if (localDashboard.value) localDashboard.value.themeId = v
    },
})

const selectedTheme = computed(
    () => (themeList.value ?? []).find(t => t.id === themeId.value) ?? null,
)

// The swatch follows the dashboard's colour mode; "auto" previews light.
const previewMode = computed(() => (localDashboard.value?.colorMode === 'dark' ? 'dark' : 'light'))

const swatch = computed(() => {
    const t = selectedTheme.value
    if (!t) return null
    const palette = t.colors[previewMode.value]
    return {
        background: pageBgValue(t.previewCss, previewMode.value),
        primary: palette.primary,
        secondary: palette.secondary,
    }
})

const widgetThemeOptions = computed(() =>
    (widgetThemeList.value ?? []).map(t => ({ label: t.name, value: t.name })),
)

const widgetTheme = computed({
    get: () => displayedRef(localDashboard.value?.widgetTheme, widgetThemeList.value?.map(t => t.name)),
    set: (v: string) => {
        if (localDashboard.value) localDashboard.value.widgetTheme = v
    },
})

// Plain dithering is the absent field, so picking it removes the key rather
// than storing a value the server would have to know about.
const dither = computed<DitherOption['value']>({
    get: () => localDashboard.value?.dither || DITHER_PLAIN,
    set: (v) => {
        if (!localDashboard.value) return
        if (v === DITHER_PLAIN) delete localDashboard.value.dither
        else localDashboard.value.dither = v
    },
})

// Common widths to pick from; the Select is editable, so any CSS width can
// still be typed in.
const MAX_WIDTH_PRESETS = ['100%', '1920px', '1600px', '1440px', '1280px', '1024px', '800px']

const ditherHint = computed(() => DITHER_OPTIONS.find(o => o.value === dither.value)?.hint ?? '')

const saveIndicator = computed(() => {
    if (saveStatus.value === 'error') return { icon: 'ti ti-alert-circle', label: 'Save failed' }
    if (saveStatus.value === 'saving') return { icon: 'ti ti-loader-2', label: 'Saving…' }
    if (hasUnsaved.value) return { icon: 'ti ti-pencil', label: 'Unsaved changes' }
    if (saveStatus.value === 'saved') return { icon: 'ti ti-check', label: 'Saved' }
    return null
})

const goToAdmin = async () => {
    await flush()
    router.push({ name: 'admin-dashboards' })
}

// vue-router records the previous in-app location in history.state.back; it
// is null when this page was opened directly, in which case "back" means the
// editor rather than leaving the app.
const goBack = async () => {
    await flush()
    if (router.options.history.state.back) {
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
        <button type="button" class="app-topbar-title" @click="goToAdmin">Dashi</button>
    </header>
    <div class="settings-view">
        <div v-if="isLoading" class="p-4">Loading...</div>
        <div v-else-if="isError" class="p-4">Failed to load dashboard.</div>
        <template v-else-if="localDashboard">
            <div class="flex flex-wrap align-items-center gap-2 mb-4">
                <Button
                    icon="ti ti-arrow-left"
                    severity="secondary"
                    text
                    rounded
                    aria-label="Back"
                    @click="goBack"
                />
                <span class="settings-title text-xl font-bold text-color">{{ localDashboard.name }} — Settings</span>
                <span
                    v-if="saveIndicator"
                    class="save-status"
                    :class="{ 'save-status-error': saveStatus === 'error' }"
                    aria-live="polite"
                >
                    <i :class="saveIndicator.icon" />
                    {{ saveIndicator.label }}
                </span>
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
                    <div v-if="localDashboard.type === 'image'" class="flex flex-column gap-1">
                        <label class="font-semibold text-sm">Dither</label>
                        <Select
                            :modelValue="dither"
                            @update:modelValue="(v: DitherOption['value'] | undefined) => { if (v !== undefined) dither = v }"
                            :options="DITHER_OPTIONS"
                            optionLabel="label"
                            optionValue="value"
                            class="w-full"
                            data-testid="dither-select"
                        />
                        <small class="settings-hint" data-testid="dither-hint">{{ ditherHint }}</small>
                        <small class="settings-hint">
                            Applies to the e-ink formats (png-bw, png-spectra6, bw, spectra6).
                            Compare the options with E-ink preview in the editor.
                        </small>
                    </div>
                </div>

                <!-- Appearance (theme, widget theme, colour mode, container) -->
                <div v-if="activeSection === 'appearance'" class="flex flex-column gap-3">
                    <label class="settings-subsection">Theme</label>
                    <div class="flex flex-column gap-1">
                        <label class="font-semibold text-sm">Theme</label>
                        <Select
                            :modelValue="themeId"
                            @update:modelValue="(v: string | undefined) => { if (v !== undefined) themeId = v }"
                            :options="themeOptions"
                            optionLabel="label"
                            optionValue="value"
                            class="w-full"
                        />
                        <small class="settings-hint">
                            Colours and background. Themes are shared between dashboards
                            and edited in the Themes admin section.
                        </small>
                    </div>
                    <div v-if="selectedTheme && swatch" class="flex align-items-center gap-2">
                        <div class="theme-swatch" :style="{ background: swatch.background }" />
                        <span class="color-chip" :style="{ background: swatch.primary }" title="Primary" />
                        <span class="color-chip" :style="{ background: swatch.secondary }" title="Secondary" />
                        <Button
                            v-if="!selectedTheme.builtin"
                            label="Edit theme"
                            icon="ti ti-external-link"
                            text
                            @click="router.push({ name: 'admin-theme-edit', params: { id: selectedTheme.id } })"
                        />
                    </div>
                    <div class="flex flex-column gap-1">
                        <label class="font-semibold text-sm">Widget theme</label>
                        <Select
                            :modelValue="widgetTheme"
                            @update:modelValue="(v: string | undefined) => { if (v !== undefined) widgetTheme = v }"
                            :options="widgetThemeOptions"
                            optionLabel="label"
                            optionValue="value"
                            class="w-full"
                        />
                        <small class="settings-hint">Fonts and icons.</small>
                    </div>
                    <div class="flex flex-column gap-1">
                        <label class="font-semibold text-sm">Color Mode</label>
                        <Select
                            :modelValue="localDashboard.colorMode || 'auto'"
                            @update:modelValue="(v: ColorMode | undefined) => { if (localDashboard && v !== undefined) localDashboard.colorMode = v }"
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
                    <label class="settings-subsection">Container</label>
                    <div class="flex flex-column gap-1">
                        <label class="font-semibold text-sm">Max Width</label>
                        <Select
                            v-model="localDashboard.container.maxWidth"
                            editable
                            :options="MAX_WIDTH_PRESETS"
                            placeholder="e.g. 1200px, 80%, 100%"
                            class="w-full"
                            data-testid="max-width-select"
                        />
                    </div>
                    <div class="flex flex-column gap-1">
                        <span class="font-semibold text-sm">Alignment</span>
                        <AlignmentPicker
                            v-model:vertical="localDashboard.container.verticalAlign"
                            v-model:horizontal="localDashboard.container.horizontalAlign"
                        />
                        <small class="settings-hint">Where the container sits on the screen.</small>
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
                            <div class="flex flex-wrap gap-2">
                                <Button
                                    label="Update"
                                    icon="ti ti-check"
                                    severity="secondary"
                                    size="small"
                                    :loading="isSettingAuth"
                                    :disabled="!authUsername || !authPassword"
                                    @click="submitAuth"
                                />
                                <Button
                                    label="Remove Protection"
                                    icon="ti ti-lock-open"
                                    severity="danger"
                                    size="small"
                                    outlined
                                    :loading="isDeletingAuth"
                                    @click="removeAuth"
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
                                @click="submitAuth"
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

<style scoped lang="scss">
@use '@/assets/breakpoints' as bp;

.settings-view {
    max-width: 1200px;
    margin: 0 auto;
    padding: 1.5rem 1rem;

    @include bp.below-md {
        padding: 1rem 0.75rem;
    }
}

.settings-title {
    flex: 1 1 auto;
    min-width: 0;
    overflow-wrap: break-word;
}

.settings-layout {
    display: flex;
    gap: 2rem;

    /* Phones: the section list becomes a tab strip above the panel. */
    @include bp.below-md {
        flex-direction: column;
        gap: 1rem;
    }
}

.settings-nav {
    position: sticky;
    top: 1rem;
    align-self: flex-start;
    min-width: 180px;
    display: flex;
    flex-direction: column;
    gap: 0.25rem;

    @include bp.below-md {
        position: static;
        align-self: stretch;
        min-width: 0;
        flex-direction: row;
        overflow-x: auto;
        padding-bottom: 0.5rem;
        border-bottom: 1px solid var(--p-surface-200);
    }
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

.settings-hint {
    color: var(--p-text-muted-color);
}

.theme-swatch {
    width: 64px;
    height: 40px;
    border-radius: 4px;
    border: 1px solid var(--p-surface-border);
    background-size: cover;
}

.color-chip {
    width: 20px;
    height: 20px;
    border-radius: 50%;
    border: 1px solid var(--p-surface-border);
}

.save-status {
    display: inline-flex;
    align-items: center;
    gap: 0.25rem;
    font-size: 0.8rem;
    color: var(--p-text-muted-color);
    white-space: nowrap;
}

.save-status-error {
    color: var(--p-red-500);
}
</style>
