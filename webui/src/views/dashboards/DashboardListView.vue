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
        const created = await createDashboard({ name, icon: '', type, container, rows: [] })
        createDialogVisible.value = false
        router.push({ name: 'dashboard-edit', params: { id: created.id } })
    } catch (err) {
        toast.add({ severity: 'error', summary: 'Error', detail: 'Failed to create dashboard', life: 5000 })
    } finally {
        isCreating.value = false
    }
}

const editDashboard = (d) => {
    router.push({ name: 'dashboard-edit', params: { id: d.id } })
}

const handleDeletePreviews = async () => {
    try {
        const result = await deletePreviews()
        toast.add({ severity: 'success', summary: 'Done', detail: `Deleted ${result.deleted} preview(s)`, life: 3000 })
    } catch (err) {
        toast.add({ severity: 'error', summary: 'Error', detail: 'Failed to delete previews', life: 5000 })
    }
}
</script>

<template>
    <header class="app-topbar">
        <span class="app-topbar-title" @click="router.push('/')">Dashi</span>
    </header>
    <div class="dashboard-list-view">
        <div class="flex align-items-center justify-content-between mb-4">
            <h1 class="text-2xl font-bold text-color">Dashboards</h1>
            <div class="flex gap-2">
                <Button
                    label="Delete Previews"
                    icon="ti ti-trash"
                    severity="secondary"
                    :loading="isDeletingPreviews"
                    @click="handleDeletePreviews"
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
                    selectionMode="single"
                    @row-click="(e) => router.push({ name: 'dashboard-view', params: { id: e.data.id } })"
                >
                    <Column field="name" header="Name" />
                    <Column header="Actions" style="width: 100px">
                        <template #body="{ data }">
                            <div class="flex gap-1 justify-content-end">
                                <Button
                                    icon="ti ti-pencil"
                                    text
                                    rounded
                                    class="p-1"
                                    @click="editDashboard(data)"
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
</style>
