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
import dashiIcon from '@/assets/icon-64.png'
import { useSettings } from '@/composables/useSettings'
import { useToast } from 'primevue/usetoast'

const router = useRouter()
const toast = useToast()
const {
    dashboards: dashboardsData,
    isLoading,
    createDashboard,
    deleteDashboard,
    deletePreviews,
    isDeletingPreviews,
} = useListDashboards()

const { data: settings } = useSettings()
const readOnly = computed(() => settings.value?.readOnly ?? false)

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

const handleDeletePreviews = async () => {
    try {
        const result = await deletePreviews()
        toast.add({ severity: 'success', summary: 'Done', detail: `Deleted ${result.deleted} preview(s)`, life: 3000 })
    } catch (err) {
        toast.add({ severity: 'error', summary: 'Error', detail: 'Failed to delete previews', life: 5000 })
    }
}

const handleDownload = async (id) => {
    try {
        await downloadDashboard(id)
    } catch (err) {
        toast.add({ severity: 'error', summary: 'Error', detail: 'Failed to download dashboard', life: 5000 })
    }
}
</script>

<template>
    <header class="app-topbar">
        <img :src="dashiIcon" alt="Dashi" class="app-topbar-icon" />
        <span class="app-topbar-title" @click="router.push('/dashboards')">Dashi</span>
    </header>
    <div class="dashboard-list-view">
        <div class="flex align-items-center justify-content-between mb-4">
            <h1 class="text-2xl font-bold text-color">Dashboards</h1>
            <div class="flex gap-2">
                <Button
                    label="Documentation"
                    icon="ti ti-book"
                    severity="secondary"
                    @click="router.push({ name: 'doc-dashboards' })"
                />
                <Button
                    v-if="!readOnly"
                    label="Delete Previews"
                    icon="ti ti-trash"
                    severity="secondary"
                    :loading="isDeletingPreviews"
                    @click="handleDeletePreviews"
                />
                <span v-if="readOnly" class="read-only-badge">Read-only mode</span>
                <Button
                    v-if="!readOnly"
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
                            <span>{{ data.name }}</span>
                            <i v-if="data.default" class="ti ti-home default-icon" title="Default dashboard" />
                        </template>
                    </Column>
                    <Column header="Actions" style="width: 180px">
                        <template #body="{ data }">
                            <div class="flex gap-1 justify-content-end">
                                <Button
                                    as="a"
                                    :href="router.resolve({ name: 'dashboard-view', params: { id: data.id } }).href"
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
                                    v-if="!readOnly"
                                    as="a"
                                    :href="router.resolve({ name: 'dashboard-edit', params: { id: data.id } }).href"
                                    icon="ti ti-pencil"
                                    text
                                    rounded
                                    class="p-1 action-link"
                                />
                                <Button
                                    v-if="!readOnly"
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
.dashboard-list-view {
    max-width: 1600px;
    margin: 0 auto;
    padding: 1.5rem 1rem;
}

.info-message {
    padding: 1rem;
    text-align: center;
    color: var(--p-text-muted-color);
}

.action-link {
    text-decoration: none;
}

.default-icon {
    margin-left: 0.5rem;
    color: var(--p-primary-color);
    font-size: 1rem;
}

.read-only-badge {
    color: var(--p-text-muted-color);
    font-size: 0.875rem;
    font-style: italic;
}
</style>
