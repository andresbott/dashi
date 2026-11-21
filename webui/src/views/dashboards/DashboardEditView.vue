<script setup lang="ts">
import { ref, computed, provide } from 'vue'
import { DASHBOARD_WIDGET_THEME, DASHBOARD_ID, ACTIVE_PAGE, TOTAL_PAGES, EDITING_MODE } from '@/lib/injectionKeys'
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import InputNumber from 'primevue/inputnumber'
import DashboardRow from '@/components/dashboards/DashboardRow.vue'

import { useDashboardDraft } from '@/composables/useDashboardDraft'
import type { Row } from '@/types/dashboard'
import { insertByColumn, clampRowToLine, fitsInRow } from '@/lib/rowLayout'
import { v4 as uuidv4 } from 'uuid'
import Dialog from 'primevue/dialog'
import dashiIcon from '@/assets/icon-64.png'

const route = useRoute()
const router = useRouter()
const id = computed(() => route.params.id as string)

const { draft: localDashboard, isLoading, isError, saveStatus, hasUnsaved, flush } = useDashboardDraft(() => id.value)

const activePageIndex = ref(0)

// Consumed only by widget config forms; the preview itself is theme-agnostic.
const dashboardWidgetTheme = computed(() => localDashboard.value?.widgetTheme || 'default')
provide(DASHBOARD_WIDGET_THEME, dashboardWidgetTheme)
provide(DASHBOARD_ID, id)
const editTotalPages = computed(() => localDashboard.value ? localDashboard.value.pages.length : 1)
provide(ACTIVE_PAGE, activePageIndex)
provide(TOTAL_PAGES, editTotalPages)
provide(EDITING_MODE, true)
const renamePageDialogVisible = ref(false)
const renamePageIndex = ref(0)
const renamePageName = ref('')
const renamePageInterval = ref<number | null>(0)

// Only image dashboards render to PNG: the server answers an interactive
// dashboard's display request with browser HTML, so there is nothing to show.
const isImageDashboard = computed(() => localDashboard.value?.type === 'image')

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

// The dialog edits staged copies; the page is only touched on Confirm, so
// Cancel (or closing the dialog) leaves it exactly as it was.
const openRenamePage = (index: number) => {
    const page = localDashboard.value?.pages[index]
    renamePageIndex.value = index
    renamePageName.value = page?.name ?? ''
    renamePageInterval.value = page?.refreshInterval ?? 0
    renamePageDialogVisible.value = true
}

const confirmRenamePage = () => {
    const page = localDashboard.value?.pages[renamePageIndex.value]
    if (!page) return
    page.name = renamePageName.value
    page.refreshInterval = Math.max(0, Math.floor(renamePageInterval.value || 0))
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

// Cross-row widget move: DashboardRow emits this on drop when the pointer was
// over a different row than the widget's own. Remove it from the source row and
// insert it into the target row at the drop column (placeRow resolves overlap,
// clampRowToLine keeps the target row on one line). The row already refuses
// the drop visually when the widget does not fit; this re-checks it.
const moveWidgetAcross = (fromIndex: number, payload: { widgetId: string; toRowId: string; column: number }) => {
    if (!activePage.value) return
    const rows = activePage.value.rows
    const fromRow = rows[fromIndex]
    if (!fromRow) return
    const widget = fromRow.widgets.find(w => w.id === payload.widgetId)
    const toIndex = rows.findIndex(r => r.id === payload.toRowId)
    if (!widget || toIndex < 0 || toIndex === fromIndex) return
    const toRow = rows[toIndex]
    if (!fitsInRow(toRow.widgets, widget.width)) return
    rows[fromIndex] = { ...fromRow, widgets: fromRow.widgets.filter(w => w.id !== payload.widgetId) }
    rows[toIndex] = {
        ...toRow,
        widgets: clampRowToLine(insertByColumn(toRow.widgets, { ...widget, column: payload.column })),
    }
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

const goToSettings = async () => {
    await flush()
    router.push({ name: 'dashboard-settings', params: { id: id.value } })
}

// Pending edits are saved first, so the server renders what is on screen;
// the preview opens on the page being edited.
const goToEinkPreview = async () => {
    await flush()
    router.push({
        name: 'dashboard-eink-preview',
        params: { id: id.value },
        query: { page: String(activePageIndex.value), from: 'edit' },
    })
}

</script>

<template>
    <header class="app-topbar">
        <img :src="dashiIcon" alt="Dashi" class="app-topbar-icon" />
        <button type="button" class="app-topbar-title" @click="goToAdmin">Dashi</button>
    </header>
    <div class="dashboard-edit-view">
        <div v-if="isLoading" class="p-4">Loading...</div>
        <div v-else-if="isError" class="p-4">Failed to load dashboard.</div>
        <template v-else-if="localDashboard">
        <div class="edit-toolbar mb-3">
            <Button
                icon="ti ti-arrow-left"
                severity="secondary"
                text
                rounded
                aria-label="Back to dashboards"
                @click="goToAdmin"
            />
            <span class="edit-toolbar-title text-xl font-bold text-color">{{ localDashboard.name }}</span>
            <span
                v-if="saveIndicator"
                class="save-status"
                :class="{ 'save-status-error': saveStatus === 'error' }"
                aria-live="polite"
            >
                <i :class="saveIndicator.icon" />
                {{ saveIndicator.label }}
            </span>
            <Button
                v-if="isImageDashboard"
                icon="ti ti-device-tablet"
                label="E-ink preview"
                severity="secondary"
                data-testid="eink-preview-button"
                @click="goToEinkPreview"
            />
            <Button
                icon="ti ti-settings"
                label="Settings"
                severity="secondary"
                @click="goToSettings"
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
            @move-widget="moveWidgetAcross(index, $event)"
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
            v-model:visible="renamePageDialogVisible"
            modal
            :closable="true"
            :draggable="false"
            header="Rename Page"
        >
            <div class="flex flex-column gap-3 dialog-form">
                <div class="flex flex-column gap-1">
                    <label class="font-semibold text-sm">Page Name</label>
                    <InputText v-model="renamePageName" placeholder="Page name" />
                </div>
                <div class="flex flex-column gap-1">
                    <label class="font-semibold text-sm">Refresh Interval (seconds)</label>
                    <InputNumber
                        v-model="renamePageInterval"
                        :min="0"
                        :useGrouping="false"
                        placeholder="0 = no auto refresh"
                    />
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

<style scoped lang="scss">
@use '@/assets/breakpoints' as bp;

.dashboard-edit-view {
    max-width: 1600px;
    margin: 0 auto;
    padding: 1.5rem 1rem;

    @include bp.below-md {
        padding: 1rem 0.5rem;
    }
}

/* Wraps instead of overflowing: on a phone the buttons drop below the title. */
.edit-toolbar {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0.5rem;
}

.edit-toolbar-title {
    flex: 1 1 auto;
    min-width: 0;
    overflow-wrap: break-word;
}

/* Tabs stay on one line and scroll sideways when there are more than fit. */
.page-tabs {
    display: flex;
    align-items: center;
    border-bottom: 1px solid var(--p-surface-border);
    margin-bottom: 1.5rem;
    gap: 0.5rem;
    overflow-x: auto;
}

.page-tab {
    display: flex;
    flex-shrink: 0;
    align-items: center;
    gap: 0.5rem;
    padding: 0.75rem 1rem;
    cursor: pointer;
    border-bottom: 2px solid transparent;
    transition: all 0.2s;
    white-space: nowrap;

    @include bp.below-md {
        padding: 0.5rem 0.75rem;
    }
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
    flex-shrink: 0;
    margin-left: auto;
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
