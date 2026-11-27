<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import Card from 'primevue/card'
import Checkbox from 'primevue/checkbox'
import InputText from 'primevue/inputtext'
import Panel from 'primevue/panel'
import Tabs from 'primevue/tabs'
import TabList from 'primevue/tablist'
import Tab from 'primevue/tab'
import TabPanels from 'primevue/tabpanels'
import TabPanel from 'primevue/tabpanel'
import { useToast } from 'primevue/usetoast'

import ThemeBackgroundEditor from '@/components/admin/ThemeBackgroundEditor.vue'
import ThemeColorsEditor from '@/components/admin/ThemeColorsEditor.vue'
import ThemePreview from '@/components/admin/ThemePreview.vue'
import ThemeSectionHeader from '@/components/admin/ThemeSectionHeader.vue'
import { useGetTheme, useUpdateTheme } from '@/composables/useThemes'
import { useAutosave } from '@/composables/useAutosave'
import { isValidModeBackground, isValidRoleColor, PALETTE_ROLES } from '@/lib/themeCss'
import { plainCopy } from '@/lib/plainCopy'
import { seedDark } from '@/lib/themeEditor'
import type { Theme, PaletteRole } from '@/types/theme'

const route = useRoute()
const router = useRouter()
const toast = useToast()
const id = computed(() => route.params.id as string)

const { data: serverTheme, isLoading, isError } = useGetTheme(() => id.value)

// Builtins are served read-only (the server answers 403 to writes); the list
// hides their Edit button, but the URL can still be typed.
const isBuiltin = computed(() => serverTheme.value?.builtin === true)
const { updateTheme } = useUpdateTheme()

// Editing happens on a local clone; autosave persists it after each change.
const local = ref<Theme | null>(null)

watch(serverTheme, (val) => {
    if (val && !local.value) {
        const copy = plainCopy(val)
        // An empty background may arrive as null from hand-edited data.
        copy.background ??= { light: {}, dark: {} }
        copy.background.light ??= {}
        copy.background.dark ??= {}
        local.value = copy
    }
}, { immediate: true })

// While an image uploads the whole editor is read-only and autosave is off:
// useAutosave drops changes made while it is disabled, so an edit typed
// during the upload would otherwise be lost (see ThemeBackgroundEditor).
const uploading = ref(false)
const editorDisabled = computed(() => isBuiltin.value || uploading.value)

// ---- colour modes ----

type Mode = 'light' | 'dark'
const MODES: { value: Mode; label: string }[] = [
    { value: 'light', label: 'Light' },
    { value: 'dark', label: 'Dark' },
]

// The tab being edited. The editor always opens on Light.
const mode = ref<Mode>('light')

// Ticking dark mode the first time fills dark from light (seedDark); dark data
// already there is kept. Unticking only hides it: the server keeps dormant
// dark data, so re-ticking restores it.
const darkMode = computed({
    get: () => local.value?.darkMode === true,
    set: (on: boolean) => {
        const t = local.value
        if (!t || editorDisabled.value) return
        if (on) {
            local.value = { ...seedDark(t), darkMode: true }
        } else {
            t.darkMode = false
            mode.value = 'light'
        }
    },
})

const setMode = (v: string | number) => {
    if (v === 'light' || v === 'dark') mode.value = v
}

// ---- validity ----
// Mirrors the server's Validate for the fields the editor lets you break, so
// a half-typed value is highlighted instead of autosaved into a 400.

const roles = PALETTE_ROLES.map(r => r.role)

const nameInvalid = computed(() => !local.value?.name.trim())

// Per mode and per section, so a collapsed Colours or Background Panel can
// still point to the invalid field it hides.
type SectionValidity = { colors: boolean; background: boolean }

// Go: the dark palette is required in full only with dark mode on; with it
// off, the roles that are set are still checked. The light palette and both
// backgrounds are always checked.
const colorsValid = (t: Theme, m: Mode) => {
    const roleOk = (role: PaletteRole) =>
        (m === 'dark' && !t.darkMode && !t.colors.dark?.[role]) || isValidRoleColor(role, t.colors[m]?.[role])
    return roles.every(roleOk)
}

const sectionValid = computed<Record<Mode, SectionValidity>>(() => {
    const t = local.value
    const of = (m: Mode): SectionValidity =>
        t ? { colors: colorsValid(t, m), background: isValidModeBackground(t.background[m]) }
            : { colors: false, background: false }
    return { light: of('light'), dark: of('dark') }
})

const modeValid = computed<Record<Mode, boolean>>(() => {
    const ok = (m: Mode) => sectionValid.value[m].colors && sectionValid.value[m].background
    return { light: ok('light'), dark: ok('dark') }
})

const isValid = computed(() => !!local.value && !nameInvalid.value && modeValid.value.light && modeValid.value.dark)

// Dormant dark data that the server would reject blocks every save, yet with
// dark mode off nothing on screen is highlighted: say where to look.
const hiddenDarkInvalid = computed(() => !!local.value && !local.value.darkMode && !modeValid.value.dark)

// ---- autosave ----

// Invalid states are neither scheduled nor sent: a save timer started by an
// earlier valid edit is skipped too. The edit that makes the theme valid
// again schedules a save of the whole theme.
const { status: saveStatus, error: saveError, flush } = useAutosave<Theme>({
    source: () => local.value,
    save: (theme) => (isValid.value ? updateTheme({ id: id.value, payload: theme }) : Promise.resolve()),
    enabled: () => !isBuiltin.value && isValid.value && !uploading.value,
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
                    <div class="edit-fields">
                        <div class="flex flex-column gap-1">
                            <label class="font-semibold text-sm" for="theme-name">Name</label>
                            <InputText
                                id="theme-name"
                                v-model="local.name"
                                :invalid="nameInvalid"
                                :disabled="editorDisabled"
                                data-testid="theme-name"
                                class="w-full"
                            />
                        </div>

                        <div class="flex align-items-center gap-2">
                            <Checkbox
                                v-model="darkMode"
                                binary
                                inputId="theme-dark-mode"
                                :disabled="editorDisabled"
                                data-testid="dark-mode-toggle"
                            />
                            <label for="theme-dark-mode" class="text-sm">Enable dark mode</label>
                        </div>

                        <!-- Save state, above the sections so it stays in view. -->
                        <div class="flex flex-column gap-1">
                            <small class="edit-hint" data-testid="save-status">
                                <template v-if="uploading">Uploading image…</template>
                                <template v-else-if="isValid">Changes are saved automatically.</template>
                                <template v-else>Fix the highlighted fields to save your changes.</template>
                            </small>
                            <small v-if="hiddenDarkInvalid" class="edit-warning">
                                Dark mode has invalid values; enable it to fix them.
                            </small>
                        </div>

                        <!-- One set of sections per colour mode: Light and Dark tabs
                             with dark mode on, the light set alone with it off.
                             Not lazy, so switching tabs keeps each mode's state,
                             including which of its Panels are collapsed (each
                             Panel keeps that itself; open at first, never saved). -->
                        <Tabs v-if="darkMode" :value="mode" @update:value="setMode">
                            <TabList>
                                <Tab v-for="m in MODES" :key="m.value" :value="m.value">
                                    {{ m.label }}
                                    <!-- Points to the tab to fix while another one is open. -->
                                    <i
                                        v-if="!modeValid[m.value]"
                                        class="ti ti-alert-circle tab-invalid"
                                        role="img"
                                        :aria-label="`${m.label} has invalid values`"
                                        :data-testid="`tab-invalid-${m.value}`"
                                    />
                                </Tab>
                            </TabList>
                            <TabPanels class="mode-panels">
                                <TabPanel v-for="m in MODES" :key="m.value" :value="m.value" class="mode-sections">
                                    <Panel header="Colours" toggleable :data-testid="`section-colors-${m.value}`">
                                        <template #header="{ id, class: titleClass }">
                                            <ThemeSectionHeader
                                                title="Colours"
                                                :invalid="!sectionValid[m.value].colors"
                                                :markerTestid="`section-invalid-colors-${m.value}`"
                                                :titleId="id"
                                                :titleClass="titleClass"
                                            />
                                        </template>
                                        <ThemeColorsEditor
                                            v-model:palette="local.colors[m.value]"
                                            :mode="m.value"
                                            :disabled="editorDisabled"
                                        />
                                    </Panel>
                                    <Panel header="Background" toggleable :data-testid="`section-background-${m.value}`">
                                        <template #header="{ id, class: titleClass }">
                                            <ThemeSectionHeader
                                                title="Background"
                                                :invalid="!sectionValid[m.value].background"
                                                :markerTestid="`section-invalid-background-${m.value}`"
                                                :titleId="id"
                                                :titleClass="titleClass"
                                            />
                                        </template>
                                        <ThemeBackgroundEditor
                                            v-model:background="local.background[m.value]"
                                            :themeId="local.id"
                                            :mode="m.value"
                                            :flush="flush"
                                            :disabled="editorDisabled"
                                            @uploading="(v: boolean) => (uploading = v)"
                                        />
                                    </Panel>
                                    <Panel header="Preview" toggleable :data-testid="`section-preview-${m.value}`">
                                        <ThemePreview :theme="local" :mode="m.value" />
                                    </Panel>
                                </TabPanel>
                            </TabPanels>
                        </Tabs>
                        <div v-else class="mode-sections">
                            <Panel header="Colours" toggleable data-testid="section-colors-light">
                                <template #header="{ id, class: titleClass }">
                                    <ThemeSectionHeader
                                        title="Colours"
                                        :invalid="!sectionValid.light.colors"
                                        markerTestid="section-invalid-colors-light"
                                        :titleId="id"
                                        :titleClass="titleClass"
                                    />
                                </template>
                                <ThemeColorsEditor
                                    v-model:palette="local.colors.light"
                                    mode="light"
                                    :disabled="editorDisabled"
                                />
                            </Panel>
                            <Panel header="Background" toggleable data-testid="section-background-light">
                                <template #header="{ id, class: titleClass }">
                                    <ThemeSectionHeader
                                        title="Background"
                                        :invalid="!sectionValid.light.background"
                                        markerTestid="section-invalid-background-light"
                                        :titleId="id"
                                        :titleClass="titleClass"
                                    />
                                </template>
                                <ThemeBackgroundEditor
                                    v-model:background="local.background.light"
                                    :themeId="local.id"
                                    mode="light"
                                    :flush="flush"
                                    :disabled="editorDisabled"
                                    @uploading="(v: boolean) => (uploading = v)"
                                />
                            </Panel>
                            <Panel header="Preview" toggleable data-testid="section-preview-light">
                                <ThemePreview :theme="local" mode="light" />
                            </Panel>
                        </div>
                    </div>
                </template>
            </Card>
        </template>
    </div>
</template>

<style scoped lang="scss">
/* .admin-section* and .info-message are shared: see assets/style.scss. The
   form spans the whole card: the preview is a section of each mode. */
.edit-fields {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    min-width: 0;
}

.mode-sections {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
}

/* The tab panels sit inside the card: no extra inset beyond the tab strip. */
.mode-panels {
    padding: 0.75rem 0 0;
}

.edit-hint {
    color: var(--p-text-muted-color);
}

.edit-warning {
    color: var(--p-red-500);
}

/* Marks an invalid tab; ThemeSectionHeader gives Panel headers the same look. */
.tab-invalid {
    margin-left: 0.35rem;
    color: var(--p-red-500);
}
</style>
