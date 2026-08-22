<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import Card from 'primevue/card'
import Button from 'primevue/button'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import { useListBackgrounds } from '@/composables/useBackgrounds'
import { pageBgValue } from '@/lib/backgroundCss'
import type { BackgroundMeta } from '@/types/background'
import { useToast } from 'primevue/usetoast'

const router = useRouter()
const toast = useToast()
const {
    backgrounds: data,
    isLoading,
    createBackground,
    isCreating,
    deleteBackground,
} = useListBackgrounds()

const backgrounds = computed(() => data.value ?? [])

const openEditor = (id: string) => {
    router.push({ name: 'admin-background-edit', params: { id } })
}

const handleCreate = async () => {
    try {
        const created = await createBackground({ name: 'New background' })
        openEditor(created.id)
    } catch {
        toast.add({ severity: 'error', summary: 'Error', detail: 'Failed to create background', life: 5000 })
    }
}

const deleteDialogVisible = ref(false)
const toDelete = ref<BackgroundMeta | null>(null)

const showDeleteDialog = (row: BackgroundMeta) => {
    toDelete.value = row
    deleteDialogVisible.value = true
}

// Deleting a referenced background is allowed: those dashboards fall back to
// the theme background. The count is here so the choice is an informed one.
const deleteMessage = computed(() => {
    const n = toDelete.value?.usedBy ?? 0
    if (n === 0) return 'Are you sure you want to delete this background?'
    return `${n} dashboard${n === 1 ? '' : 's'} use this background and will fall back to the theme background. Delete anyway?`
})

const confirmDelete = async () => {
    if (!toDelete.value) return
    try {
        await deleteBackground(toDelete.value.id)
        deleteDialogVisible.value = false
        toDelete.value = null
    } catch {
        toast.add({ severity: 'error', summary: 'Error', detail: 'Failed to delete', life: 5000 })
    }
}
</script>

<template>
    <div class="admin-section">
        <div class="admin-section-header">
            <h2 class="admin-section-title">Backgrounds</h2>
            <Button
                label="New background"
                icon="ti ti-plus"
                :loading="isCreating"
                @click="handleCreate"
            />
        </div>

        <Card v-if="!backgrounds.length && !isLoading">
            <template #content>
                <div class="info-message">No backgrounds yet. Create one to get started.</div>
            </template>
        </Card>

        <Card v-else>
            <template #content>
                <DataTable
                    :value="backgrounds"
                    :loading="isLoading"
                    data-key="id"
                    class="p-datatable-sm"
                    stripedRows
                >
                    <Column header="" style="width: 80px">
                        <template #body="{ data: row }">
                            <div class="swatch" :style="{ background: pageBgValue(row.previewCss) }" />
                        </template>
                    </Column>
                    <Column field="name" header="Name" />
                    <Column header="Used by" style="width: 160px">
                        <template #body="{ data: row }">
                            <span :class="{ 'text-muted': row.usedBy === 0 }">
                                {{ row.usedBy }} dashboard{{ row.usedBy === 1 ? '' : 's' }}
                            </span>
                        </template>
                    </Column>
                    <Column header="Actions" style="width: 120px">
                        <template #body="{ data: row }">
                            <div class="flex gap-1 justify-content-end">
                                <Button
                                    icon="ti ti-pencil"
                                    text
                                    rounded
                                    class="p-1"
                                    @click="openEditor(row.id)"
                                />
                                <Button
                                    icon="ti ti-trash"
                                    text
                                    rounded
                                    severity="danger"
                                    class="p-1"
                                    @click="showDeleteDialog(row)"
                                />
                            </div>
                        </template>
                    </Column>
                </DataTable>
            </template>
        </Card>
    </div>

    <ConfirmDialog
        v-model:visible="deleteDialogVisible"
        :name="toDelete?.name ?? ''"
        title="Delete background"
        :message="deleteMessage"
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

.swatch {
    width: 64px;
    height: 40px;
    border-radius: 4px;
    border: 1px solid var(--p-surface-border);
    background-size: cover;
}

.text-muted {
    color: var(--p-text-muted-color);
}
</style>
