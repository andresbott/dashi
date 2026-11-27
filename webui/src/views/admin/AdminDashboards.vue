<script setup>
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import Card from 'primevue/card'
import Button from 'primevue/button'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import CreateDashboardDialog from '@/components/dashboards/CreateDashboardDialog.vue'
import { useListDashboards } from '@/composables/useDashboards'
import { downloadDashboard } from '@/lib/api/dashboard'
import { dashboardViewUrl } from '@/lib/serverInfo'
import { useToast } from 'primevue/usetoast'

const router = useRouter()
const toast = useToast()
const {
    dashboards: dashboardsData,
    isLoading,
    createDashboard,
    deleteDashboard,
    duplicateDashboard,
    uploadZip,
    isUploadingZip,
    setDefaultDashboard,
} = useListDashboards()

const dashboards = computed(() => dashboardsData.value ?? [])

const deleteDialogVisible = ref(false)
const dashboardToDelete = ref(null)

const showDeleteDialog = (d) => {
    dashboardToDelete.value = d
    deleteDialogVisible.value = true
}

const confirmDelete = async () => {
    if (!dashboardToDelete.value) return
    try {
        await deleteDashboard(dashboardToDelete.value.id)
        deleteDialogVisible.value = false
        dashboardToDelete.value = null
    } catch {
        toast.add({ severity: 'error', summary: 'Error', detail: 'Failed to delete dashboard', life: 5000 })
    }
}

const createDialogVisible = ref(false)
const isCreating = ref(false)

const handleCreate = async ({ name, type, container }) => {
    isCreating.value = true
    try {
        const created = await createDashboard({ name, icon: '', type, container, pages: [{ name: '', rows: [] }] })
        createDialogVisible.value = false
        router.push({ name: 'dashboard-edit', params: { id: created.id } })
    } catch {
        toast.add({ severity: 'error', summary: 'Error', detail: 'Failed to create dashboard', life: 5000 })
    } finally {
        isCreating.value = false
    }
}

const uploadInput = ref(null)

const handleUploadZip = async (event) => {
    const file = event.target.files?.[0]
    if (!file) return
    try {
        const data = await file.arrayBuffer()
        const created = await uploadZip(data)
        toast.add({ severity: 'success', summary: 'Imported', detail: `Dashboard "${created.name}" imported`, life: 3000 })
    } catch {
        toast.add({ severity: 'error', summary: 'Error', detail: 'Failed to import dashboard', life: 5000 })
    } finally {
        if (uploadInput.value) uploadInput.value.value = ''
    }
}

// Picking the default dashboard behaves like a radio group: the backend clears
// the flag on the previous default, so there is nothing to un-set here.
// Clicking the row that is already default is not short-circuited on purpose —
// data written before the backend enforced exclusivity (or by hand) can have
// several defaults at once, and re-asserting is the only way to repair it.
const pendingDefaultId = ref(null)

const handleSetDefault = async (dashboard) => {
    pendingDefaultId.value = dashboard.id
    try {
        await setDefaultDashboard(dashboard.id)
    } catch {
        toast.add({ severity: 'error', summary: 'Error', detail: 'Failed to set default dashboard', life: 5000 })
    } finally {
        pendingDefaultId.value = null
    }
}

// Visual cue for how a dashboard renders: interactive ones are served as live
// HTML, image ones as a PNG for e-ink displays. Anything unset falls back to
// interactive, which is the backend default.
const DASHBOARD_TYPES = {
    interactive: { icon: 'ti-device-desktop', label: 'Interactive dashboard' },
    image: { icon: 'ti-photo', label: 'Image dashboard' },
}

const dashboardType = (d) => DASHBOARD_TYPES[d.type] ?? DASHBOARD_TYPES.interactive

// E-ink preview, the same page the editor opens, for image dashboards only
// (an interactive one answers a display request with HTML). It opens on the
// first page; `from` lets its Back button return here without history.
const openEinkPreview = (d) => {
    router.push({ name: 'dashboard-eink-preview', params: { id: d.id }, query: { from: 'list' } })
}

// Stays on the list: the copy shows up as "<name> (copy)" and can be renamed
// from its settings.
const pendingDuplicateId = ref(null)

const handleDuplicate = async (dashboard) => {
    pendingDuplicateId.value = dashboard.id
    try {
        const created = await duplicateDashboard(dashboard.id)
        toast.add({ severity: 'success', summary: 'Duplicated', detail: `Created "${created.name}"`, life: 3000 })
    } catch {
        toast.add({ severity: 'error', summary: 'Error', detail: 'Failed to duplicate dashboard', life: 5000 })
    } finally {
        pendingDuplicateId.value = null
    }
}

const handleDownload = async (id) => {
    try {
        await downloadDashboard(id)
    } catch {
        toast.add({ severity: 'error', summary: 'Error', detail: 'Failed to download dashboard', life: 5000 })
    }
}
</script>

<template>
    <div class="admin-section">
        <div class="admin-section-header">
            <h2 class="admin-section-title">Dashboards</h2>
            <div class="flex gap-2">
                <Button
                    label="Import"
                    icon="ti ti-upload"
                    severity="secondary"
                    :loading="isUploadingZip"
                    @click="uploadInput?.click()"
                />
                <input
                    ref="uploadInput"
                    type="file"
                    accept=".zip"
                    style="display: none"
                    @change="handleUploadZip"
                />
                <Button
                    label="New Dashboard"
                    icon="ti ti-plus"
                    @click="createDialogVisible = true"
                />
            </div>
        </div>

        <Card v-if="!dashboards.length && !isLoading">
            <template #content>
                <div class="info-message">
                    No dashboards yet. Create your first dashboard to get started.
                </div>
            </template>
        </Card>

        <Card v-else>
            <template #content>
                <DataTable
                    :value="dashboards"
                    :loading="isLoading"
                    data-key="id"
                    class="p-datatable-sm"
                    stripedRows
                >
                    <Column field="name" header="Name">
                        <template #body="{ data }">
                            <span class="flex align-items-center gap-2">
                                <i
                                    v-tooltip.top="dashboardType(data).label"
                                    class="ti dashboard-type-icon"
                                    :class="dashboardType(data).icon"
                                    :aria-label="dashboardType(data).label"
                                    role="img"
                                />
                                <span>{{ data.name }}</span>
                            </span>
                        </template>
                    </Column>
                    <Column header="Actions" style="width: 360px">
                        <template #body="{ data }">
                            <!--
                                Grouped left to right: default flag | open / preview / edit |
                                settings / duplicate / export | delete (destructive, kept last).
                            -->
                            <div class="flex flex-wrap md:flex-nowrap gap-1 justify-content-end align-items-center">
                                <Button
                                    v-tooltip.top="data.default ? 'Default dashboard' : 'Make default dashboard'"
                                    :aria-label="data.default ? 'Default dashboard' : 'Make default dashboard'"
                                    icon="ti ti-home"
                                    :text="!data.default"
                                    :severity="data.default ? 'primary' : 'secondary'"
                                    :loading="pendingDefaultId === data.id"
                                    :aria-pressed="!!data.default"
                                    rounded
                                    class="p-1"
                                    @click="handleSetDefault(data)"
                                />

                                <span class="action-divider" aria-hidden="true" />

                                <Button
                                    v-tooltip.top="'Open dashboard in a new tab'"
                                    aria-label="Open dashboard in a new tab"
                                    as="a"
                                    :href="dashboardViewUrl(data.id)"
                                    target="_blank"
                                    rel="noopener"
                                    icon="ti ti-eye"
                                    text
                                    rounded
                                    class="p-1 action-link"
                                    data-testid="view-button"
                                />
                                <Button
                                    v-if="data.type === 'image'"
                                    v-tooltip.top="'E-ink preview (PNG)'"
                                    aria-label="E-ink preview (PNG)"
                                    icon="ti ti-device-tablet"
                                    text
                                    rounded
                                    class="p-1"
                                    data-testid="eink-preview-button"
                                    @click="openEinkPreview(data)"
                                />
                                <!-- Invisible stand-in so the columns line up across image and interactive rows. -->
                                <Button
                                    v-else
                                    icon="ti ti-device-tablet"
                                    text
                                    rounded
                                    class="p-1 action-placeholder"
                                    tabindex="-1"
                                    aria-hidden="true"
                                    disabled
                                />
                                <Button
                                    v-tooltip.top="'Edit layout and widgets'"
                                    aria-label="Edit layout and widgets"
                                    as="a"
                                    :href="router.resolve({ name: 'dashboard-edit', params: { id: data.id } }).href"
                                    icon="ti ti-pencil"
                                    text
                                    rounded
                                    class="p-1 action-link"
                                    data-testid="edit-button"
                                />

                                <span class="action-divider" aria-hidden="true" />

                                <Button
                                    v-tooltip.top="'Dashboard settings'"
                                    aria-label="Dashboard settings"
                                    as="a"
                                    :href="router.resolve({ name: 'dashboard-settings', params: { id: data.id } }).href"
                                    icon="ti ti-settings"
                                    text
                                    rounded
                                    class="p-1 action-link"
                                    data-testid="settings-button"
                                />
                                <Button
                                    v-tooltip.top="'Duplicate dashboard'"
                                    aria-label="Duplicate dashboard"
                                    icon="ti ti-copy"
                                    text
                                    rounded
                                    class="p-1"
                                    :loading="pendingDuplicateId === data.id"
                                    data-testid="duplicate-button"
                                    @click="handleDuplicate(data)"
                                />
                                <Button
                                    v-tooltip.top="'Export as ZIP'"
                                    aria-label="Export as ZIP"
                                    icon="ti ti-download"
                                    text
                                    rounded
                                    class="p-1"
                                    data-testid="export-button"
                                    @click="handleDownload(data.id)"
                                />

                                <span class="action-divider" aria-hidden="true" />

                                <Button
                                    v-tooltip.top="'Delete dashboard'"
                                    aria-label="Delete dashboard"
                                    icon="ti ti-trash"
                                    text
                                    rounded
                                    severity="danger"
                                    class="p-1"
                                    data-testid="delete-button"
                                    @click="showDeleteDialog(data)"
                                />
                            </div>
                        </template>
                    </Column>
                </DataTable>
            </template>
        </Card>
    </div>

    <CreateDashboardDialog
        v-model:visible="createDialogVisible"
        v-model:saving="isCreating"
        @confirm="handleCreate"
    />

    <ConfirmDialog
        v-model:visible="deleteDialogVisible"
        :name="dashboardToDelete?.name"
        title="Delete Dashboard"
        message="Are you sure you want to delete this dashboard?"
        @confirm="confirmDelete"
    />
</template>

<style scoped>
/* .admin-section* and .info-message are shared: see assets/style.scss. */
.action-link {
    text-decoration: none;
}

.action-divider {
    width: 1px;
    height: 1.25rem;
    margin: 0 0.25rem;
    background: var(--p-content-border-color);
}

.action-placeholder {
    visibility: hidden;
}

.dashboard-type-icon {
    font-size: 1.1rem;
    color: var(--p-text-muted-color);
}
</style>
