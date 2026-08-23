<script setup lang="ts">
import { ref, computed, watch, provide } from 'vue'
import { DASHBOARD_THEME, DASHBOARD_ID, ACTIVE_PAGE, TOTAL_PAGES, EDITING_MODE } from '@/lib/injectionKeys'
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import DashboardRow from '@/components/dashboards/DashboardRow.vue'

import { useGetDashboard, useUpdateDashboard } from '@/composables/useDashboards'
import { useAutosave } from '@/composables/useAutosave'
import { useToast } from 'primevue/usetoast'
import type { Dashboard, Row } from '@/types/dashboard'
import { insertByColumn } from '@/lib/rowLayout'
import { v4 as uuidv4 } from 'uuid'
import Dialog from 'primevue/dialog'
import dashiIcon from '@/assets/icon-64.png'

const route = useRoute()
const router = useRouter()
const toast = useToast()
const id = computed(() => route.params.id as string)

const { data: serverDashboard, isLoading, isError } = useGetDashboard(() => id.value)
const { updateDashboard } = useUpdateDashboard()

const localDashboard = ref<Dashboard | null>(null)
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
    }
}, { immediate: true })

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

// Cross-row widget move: DashboardRow emits this on drop when the pointer was
// over a different row than the widget's own. Remove it from the source row and
// insert it into the target row at the drop column (placeRow resolves overlap).
const moveWidgetAcross = (fromIndex: number, payload: { widgetId: string; toRowId: string; column: number }) => {
    if (!activePage.value) return
    const rows = activePage.value.rows
    const fromRow = rows[fromIndex]
    if (!fromRow) return
    const widget = fromRow.widgets.find(w => w.id === payload.widgetId)
    const toIndex = rows.findIndex(r => r.id === payload.toRowId)
    if (!widget || toIndex < 0 || toIndex === fromIndex) return
    rows[fromIndex] = { ...fromRow, widgets: fromRow.widgets.filter(w => w.id !== payload.widgetId) }
    const toRow = rows[toIndex]
    rows[toIndex] = { ...toRow, widgets: insertByColumn(toRow.widgets, { ...widget, column: payload.column }) }
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

const { status: saveStatus, flush } = useAutosave<Dashboard>({
    source: () => localDashboard.value,
    save: (d) => updateDashboard({ id: id.value, payload: d }),
})

watch(saveStatus, (s) => {
    if (s === 'error') {
        toast.add({ severity: 'error', summary: 'Error', detail: 'Failed to save dashboard', life: 5000 })
    }
})

const goToAdmin = async () => {
    await flush()
    router.push({ name: 'admin-dashboards' })
}

const goToSettings = async () => {
    await flush()
    router.push({ name: 'dashboard-settings', params: { id: id.value } })
}

</script>

<template>
    <header class="app-topbar">
        <img :src="dashiIcon" alt="Dashi" class="app-topbar-icon" />
        <span class="app-topbar-title" @click="router.push('/admin')">Dashi</span>
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
                @click="goToAdmin"
            />
            <span class="text-xl font-bold text-color flex-grow-1">{{ localDashboard.name }}</span>
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
            <div class="flex flex-column gap-3" style="min-width: 350px">
                <div class="flex flex-column gap-1">
                    <label class="font-semibold text-sm">Page Name</label>
                    <InputText v-model="renamePageName" placeholder="Page name" />
                </div>
                <div class="flex flex-column gap-1">
                    <label class="font-semibold text-sm">Refresh Interval (seconds)</label>
                    <InputText
                        :modelValue="String(localDashboard!.pages[renamePageIndex]?.refreshInterval ?? 0)"
                        @update:modelValue="(v: string | undefined) => {
                            if (!localDashboard || v === undefined) return
                            localDashboard.pages[renamePageIndex].refreshInterval = parseInt(v) || 0
                        }"
                        placeholder="0 (no header)"
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
</style>
