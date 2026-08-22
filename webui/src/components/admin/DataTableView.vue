<script setup lang="ts">
import { ref, computed } from 'vue'
import Card from 'primevue/card'
import Button from 'primevue/button'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import { useDataItems } from '@/composables/useDataItems'
import { dataUrl, type DataKind } from '@/lib/api/data'
import { useToast } from 'primevue/usetoast'

const props = defineProps<{
    kind: DataKind
}>()

const toast = useToast()
const { items: itemsData, isLoading, uploadItem, isUploading, deleteItem } = useDataItems(props.kind)
const items = computed(() => itemsData.value ?? [])

// Both kinds this component serves are image libraries: 'images' feeds the
// image widget, 'backgrounds' is the shared pool that backgrounds reference
// with `shared:`. Background *configurations* live in AdminBackgrounds.vue.
const title = computed(() => (props.kind === 'images' ? 'Images' : 'Shared background images'))

// Upload
const uploadInput = ref<HTMLInputElement | null>(null)
const overwriteDialogVisible = ref(false)
const pendingUpload = ref<{ name: string; bytes: ArrayBuffer } | null>(null)

const handleFilePicked = async (event: Event) => {
    const input = event.target as HTMLInputElement
    const file = input.files?.[0]
    if (!file) return
    try {
        const bytes = await file.arrayBuffer()
        const name = file.name
        if (items.value.some((i) => i.name === name)) {
            pendingUpload.value = { name, bytes }
            overwriteDialogVisible.value = true
        } else {
            await uploadItem({ name, bytes })
            toast.add({ severity: 'success', summary: 'Uploaded', detail: name, life: 3000 })
        }
    } catch {
        toast.add({ severity: 'error', summary: 'Error', detail: 'Failed to upload', life: 5000 })
    } finally {
        if (uploadInput.value) uploadInput.value.value = ''
    }
}

const confirmOverwrite = async () => {
    const p = pendingUpload.value
    overwriteDialogVisible.value = false
    pendingUpload.value = null
    if (!p) return
    try {
        await uploadItem({ name: p.name, bytes: p.bytes })
        toast.add({ severity: 'success', summary: 'Uploaded', detail: p.name, life: 3000 })
    } catch {
        toast.add({ severity: 'error', summary: 'Error', detail: 'Failed to upload', life: 5000 })
    }
}

// Delete
const deleteDialogVisible = ref(false)
const itemToDelete = ref<string | null>(null)

const showDeleteDialog = (name: string) => {
    itemToDelete.value = name
    deleteDialogVisible.value = true
}

const confirmDelete = async () => {
    if (!itemToDelete.value) return
    try {
        await deleteItem(itemToDelete.value)
        deleteDialogVisible.value = false
        itemToDelete.value = null
    } catch {
        toast.add({ severity: 'error', summary: 'Error', detail: 'Failed to delete', life: 5000 })
    }
}

const downloadUrl = (name: string) => dataUrl(props.kind, name)

const formatSize = (bytes: number): string => {
    if (bytes < 1024) return `${bytes} B`
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
    return `${(bytes / 1024 / 1024).toFixed(1)} MB`
}

const formatTime = (t: string): string => {
    try {
        return new Date(t).toLocaleString()
    } catch {
        return t
    }
}
</script>

<template>
    <div class="admin-section">
        <div class="admin-section-header">
            <h2 class="admin-section-title">{{ title }}</h2>
            <Button
                label="Upload"
                icon="ti ti-upload"
                :loading="isUploading"
                @click="uploadInput?.click()"
            />
            <input
                ref="uploadInput"
                type="file"
                accept="image/*"
                style="display: none"
                @change="handleFilePicked"
            />
        </div>

        <Card v-if="!items.length && !isLoading">
            <template #content>
                <div class="info-message">No {{ title.toLowerCase() }} yet. Upload one to get started.</div>
            </template>
        </Card>

        <Card v-else>
            <template #content>
                <DataTable
                    :value="items"
                    :loading="isLoading"
                    data-key="name"
                    class="p-datatable-sm"
                    stripedRows
                >
                    <Column header="" style="width: 64px">
                        <template #body="{ data }">
                            <img :src="downloadUrl(data.name)" :alt="data.name" class="thumb" />
                        </template>
                    </Column>
                    <Column field="name" header="Name" />
                    <Column header="Size" style="width: 100px">
                        <template #body="{ data }">{{ formatSize(data.size) }}</template>
                    </Column>
                    <Column header="Modified" style="width: 200px">
                        <template #body="{ data }">{{ formatTime(data.modTime) }}</template>
                    </Column>
                    <Column header="Actions" style="width: 120px">
                        <template #body="{ data }">
                            <div class="flex gap-1 justify-content-end">
                                <Button
                                    as="a"
                                    :href="downloadUrl(data.name)"
                                    :download="data.name"
                                    icon="ti ti-download"
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
                                    @click="showDeleteDialog(data.name)"
                                />
                            </div>
                        </template>
                    </Column>
                </DataTable>
            </template>
        </Card>
    </div>

    <ConfirmDialog
        v-model:visible="overwriteDialogVisible"
        :name="pendingUpload?.name ?? ''"
        title="Overwrite?"
        message="A file with this name already exists. Overwrite?"
        @confirm="confirmOverwrite"
    />

    <ConfirmDialog
        v-model:visible="deleteDialogVisible"
        :name="itemToDelete ?? ''"
        title="Delete image"
        message="Are you sure you want to delete this file?"
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

.thumb {
    width: 48px;
    height: 48px;
    object-fit: cover;
    border-radius: 4px;
    border: 1px solid var(--p-surface-border);
    background: var(--p-surface-100);
}

.action-link {
    text-decoration: none;
}
</style>
