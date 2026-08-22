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
    } catch (err) {
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
    } catch (err) {
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
    } catch (err) {
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

const handleDownload = async (id) => {
    try {
        await downloadDashboard(id)
    } catch (err) {
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
                    <Column header="Actions" style="width: 220px">
                        <template #body="{ data }">
                            <div class="flex gap-1 justify-content-end">
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
                                <Button
                                    as="a"
                                    :href="dashboardViewUrl(data.id)"
                                    target="_blank"
                                    rel="noopener"
                                    icon="ti ti-eye"
                                    text
                                    rounded
                                    class="p-1 action-link"
                                />
                                <Button
                                    icon="ti ti-download"
                                    text
                                    rounded
                                    class="p-1"
                                    @click="handleDownload(data.id)"
                                />
                                <Button
                                    as="a"
                                    :href="router.resolve({ name: 'dashboard-edit', params: { id: data.id } }).href"
                                    icon="ti ti-pencil"
                                    text
                                    rounded
                                    class="p-1 action-link"
                                />
                                <Button
                                    as="a"
                                    :href="router.resolve({ name: 'dashboard-settings', params: { id: data.id } }).href"
                                    icon="ti ti-settings"
                                    text
                                    rounded
                                    class="p-1 action-link"
                                />
                                <Button
                                    icon="ti ti-trash"
                                    text
                                    rounded
                                    severity="danger"
                                    class="p-1"
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
.admin-section {
    display: flex;
    flex-direction: column;
    gap: 1rem;
}

.admin-section-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 1rem;
}

.admin-section-title {
    font-size: 1.5rem;
    font-weight: 700;
    color: var(--p-text-color);
    margin: 0;
}

.info-message {
    padding: 1rem;
    text-align: center;
    color: var(--p-text-muted-color);
}

.action-link {
    text-decoration: none;
}

.dashboard-type-icon {
    font-size: 1.1rem;
    color: var(--p-text-muted-color);
}
</style>
