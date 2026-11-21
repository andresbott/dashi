<script setup lang="ts">
import { ref, computed } from 'vue'
import Card from 'primevue/card'
import Button from 'primevue/button'
import Tag from 'primevue/tag'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import { useAdminWidgetThemes } from '@/composables/useWidgetThemes'
import { widgetThemeDownloadUrl } from '@/lib/api/widgetThemes'
import { useToast } from 'primevue/usetoast'

// Widget themes supply fonts and icons. Colours and backgrounds belong to
// themes (AdminThemes.vue).
const toast = useToast()
const {
    widgetThemes: themesData,
    isLoading,
    uploadWidgetTheme,
    isUploading,
    deleteWidgetTheme,
} = useAdminWidgetThemes()
const themes = computed(() => themesData.value ?? [])

// Upload
const uploadInput = ref<HTMLInputElement | null>(null)

const handleFilePicked = async (event: Event) => {
    const input = event.target as HTMLInputElement
    const file = input.files?.[0]
    if (!file) return
    try {
        const bytes = await file.arrayBuffer()
        const info = await uploadWidgetTheme(bytes)
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
        await deleteWidgetTheme(name)
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
            <h2 class="admin-section-title">Widget themes</h2>
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
                <div class="info-message">No widget themes yet. Upload a widget theme zip to get started.</div>
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
                    <Column field="description" header="Description" headerClass="hide-below-md" bodyClass="hide-below-md" />
                    <Column header="Actions" style="width: 120px">
                        <template #body="{ data }">
                            <div class="flex gap-1 justify-content-end">
                                <Button
                                    as="a"
                                    :href="widgetThemeDownloadUrl(data.name)"
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
        title="Delete widget theme"
        message="Are you sure you want to delete this widget theme? This cannot be undone."
        @confirm="confirmDelete"
    />
</template>

<style scoped>
/* .admin-section* and .info-message are shared: see assets/style.scss. */
.action-link {
    text-decoration: none;
}
</style>
