<script setup lang="ts">
import { ref, computed } from 'vue'
import Card from 'primevue/card'
import Button from 'primevue/button'
import Tag from 'primevue/tag'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import { useAdminThemes } from '@/composables/useThemes'
import { themeDownloadUrl } from '@/lib/api/themes'
import { useToast } from 'primevue/usetoast'

const toast = useToast()
const {
    themes: themesData,
    isLoading,
    uploadTheme,
    isUploading,
    deleteTheme,
} = useAdminThemes()
const themes = computed(() => themesData.value ?? [])

const typeSeverity = (type: string): 'info' | 'success' | 'warning' | 'secondary' => {
    switch (type) {
        case 'theme': return 'info'
        case 'icon':  return 'success'
        case 'style': return 'warning'
        default:      return 'secondary'
    }
}

// Upload
const uploadInput = ref<HTMLInputElement | null>(null)

const handleFilePicked = async (event: Event) => {
    const input = event.target as HTMLInputElement
    const file = input.files?.[0]
    if (!file) return
    try {
        const bytes = await file.arrayBuffer()
        const info = await uploadTheme(bytes)
        toast.add({ severity: 'success', summary: 'Uploaded', detail: info.name, life: 3000 })
    } catch (err: unknown) {
        // axios error shape: err.response?.data?.error
        const axErr = err as { response?: { data?: { error?: string } } }
        const message = axErr?.response?.data?.error ?? 'Failed to upload'
        toast.add({ severity: 'error', summary: 'Upload failed', detail: message, life: 5000 })
    } finally {
        if (uploadInput.value) uploadInput.value.value = ''
    }
}

// Delete
const deleteDialogVisible = ref(false)
const themeToDelete = ref<string | null>(null)

const showDeleteDialog = (name: string) => {
    themeToDelete.value = name
    deleteDialogVisible.value = true
}

const confirmDelete = async () => {
    if (!themeToDelete.value) return
    const name = themeToDelete.value
    try {
        await deleteTheme(name)
        deleteDialogVisible.value = false
        themeToDelete.value = null
        toast.add({ severity: 'success', summary: 'Deleted', detail: name, life: 3000 })
    } catch {
        toast.add({ severity: 'error', summary: 'Error', detail: 'Failed to delete', life: 5000 })
    }
}
</script>

<template>
    <div class="admin-section">
        <div class="admin-section-header">
            <h2 class="admin-section-title">Themes</h2>
            <Button
                label="Upload"
                icon="ti ti-upload"
                :loading="isUploading"
                @click="uploadInput?.click()"
            />
            <input
                ref="uploadInput"
                type="file"
                accept=".zip,application/zip"
                style="display: none"
                @change="handleFilePicked"
            />
        </div>

        <Card v-if="!themes.length && !isLoading">
            <template #content>
                <div class="info-message">No themes yet. Upload a theme zip to get started.</div>
            </template>
        </Card>

        <Card v-else>
            <template #content>
                <DataTable
                    :value="themes"
                    :loading="isLoading"
                    data-key="name"
                    class="p-datatable-sm"
                    stripedRows
                >
                    <Column field="name" header="Name">
                        <template #body="{ data }">
                            <span>{{ data.name }}</span>
                            <Tag
                                v-if="data.builtin"
                                value="Builtin"
                                severity="info"
                                class="ml-2"
                            />
                        </template>
                    </Column>
                    <Column header="Type" style="width: 120px">
                        <template #body="{ data }">
                            <Tag :value="data.type" :severity="typeSeverity(data.type)" />
                        </template>
                    </Column>
                    <Column field="description" header="Description" />
                    <Column header="Actions" style="width: 120px">
                        <template #body="{ data }">
                            <div class="flex gap-1 justify-content-end">
                                <Button
                                    as="a"
                                    :href="themeDownloadUrl(data.name)"
                                    :download="data.name + '.zip'"
                                    icon="ti ti-download"
                                    text
                                    rounded
                                    class="p-1 action-link"
                                />
                                <Button
                                    v-if="!data.builtin"
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
        v-model:visible="deleteDialogVisible"
        :name="themeToDelete ?? ''"
        title="Delete Theme"
        message="Are you sure you want to delete this theme? This cannot be undone."
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
</style>
