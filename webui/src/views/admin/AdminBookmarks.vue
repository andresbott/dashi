<script setup lang="ts">
// Bookmark files: YAML lists in {DataDir}/content/bookmarks that a bookmark
// group widget can load instead of keeping its bookmarks in the dashboard.
import { ref, computed } from 'vue'
import Card from 'primevue/card'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import InputText from 'primevue/inputtext'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import { useBookmarkFiles } from '@/composables/useBookmarkFiles'
import { getBookmarkFile, type BookmarkFileItem } from '@/lib/api/bookmarkFiles'
import BookmarkListEditor from '@/widgets/bookmarkgroup/BookmarkListEditor.vue'
import { isSafeBookmarkUrl } from '@/widgets/bookmark/url'
import { apiErrorMessage } from '@/lib/api/client'
import { useToast } from 'primevue/usetoast'

const toast = useToast()
const { files, isLoading, saveFile, deleteFile } = useBookmarkFiles()

const deleteDialogVisible = ref(false)
const fileToDelete = ref<string | null>(null)

const showDeleteDialog = (name: string) => {
    fileToDelete.value = name
    deleteDialogVisible.value = true
}

const confirmDelete = async () => {
    if (!fileToDelete.value) return
    try {
        await deleteFile(fileToDelete.value)
        deleteDialogVisible.value = false
        fileToDelete.value = null
    } catch {
        toast.add({ severity: 'error', summary: 'Error', detail: 'Failed to delete bookmarks file', life: 5000 })
    }
}

// Create dialog
const createDialogVisible = ref(false)
const newFilename = ref('')
const createError = ref('')
const creating = ref(false)

const validateNewFilename = (raw: string): { ok: true; name: string } | { ok: false; error: string } => {
    let name = raw.trim()
    if (!name) return { ok: false, error: 'Filename is required' }
    if (name.length > 100) return { ok: false, error: 'Filename is too long' }
    if (name.includes('/') || name.includes('..')) return { ok: false, error: 'Invalid characters in filename' }
    const lower = name.toLowerCase()
    if (!lower.endsWith('.yaml') && !lower.endsWith('.yml')) name += '.yaml'
    if (files.value.includes(name)) return { ok: false, error: `"${name}" already exists` }
    return { ok: true, name }
}

const openCreate = () => {
    newFilename.value = ''
    createError.value = ''
    createDialogVisible.value = true
}

const confirmCreate = async () => {
    const result = validateNewFilename(newFilename.value)
    if (!result.ok) {
        createError.value = result.error
        return
    }
    creating.value = true
    createError.value = ''
    try {
        await saveFile({ name: result.name, bookmarks: [] })
        createDialogVisible.value = false
        openEdit(result.name)
    } catch (err) {
        createError.value = apiErrorMessage(err) ?? 'Failed to create file'
    } finally {
        creating.value = false
    }
}

// Edit dialog
const editDialogVisible = ref(false)
const editFilename = ref<string | null>(null)
const editBookmarks = ref<BookmarkFileItem[]>([])
const editLoading = ref(false)
const editLoadError = ref('')
const saving = ref(false)
const saveError = ref('')

const openEdit = async (name: string) => {
    editFilename.value = name
    editBookmarks.value = []
    editLoadError.value = ''
    saveError.value = ''
    editDialogVisible.value = true
    editLoading.value = true
    try {
        editBookmarks.value = await getBookmarkFile(name)
    } catch (err) {
        editLoadError.value = apiErrorMessage(err) ?? 'Failed to load file.'
    } finally {
        editLoading.value = false
    }
}

// Same URL rule as bookmarks kept in a dashboard (see widgets/bookmark/url.ts):
// a file is not saved while any bookmark fails it.
const hasUnsafeUrl = computed(() => editBookmarks.value.some(b => !isSafeBookmarkUrl(b.url)))

const save = async () => {
    if (!editFilename.value || editLoadError.value || hasUnsafeUrl.value) return
    saving.value = true
    saveError.value = ''
    try {
        await saveFile({ name: editFilename.value, bookmarks: editBookmarks.value })
        editDialogVisible.value = false
    } catch (err) {
        saveError.value = apiErrorMessage(err) ?? 'Failed to save'
    } finally {
        saving.value = false
    }
}
</script>

<template>
    <div class="admin-section">
        <div class="admin-section-header">
            <h2 class="admin-section-title">Bookmarks</h2>
            <Button label="New" icon="ti ti-plus" @click="openCreate" />
        </div>

        <Card v-if="!files.length && !isLoading">
            <template #content>
                <div class="info-message">
                    No bookmark files yet. A bookmark group widget can load its bookmarks from one of these files
                    instead of keeping them in the dashboard.
                </div>
            </template>
        </Card>

        <Card v-else>
            <template #content>
                <DataTable
                    :value="files.map((n) => ({ name: n }))"
                    :loading="isLoading"
                    data-key="name"
                    class="p-datatable-sm"
                    stripedRows
                >
                    <Column field="name" header="Name" />
                    <Column header="Actions" style="width: 140px">
                        <template #body="{ data }">
                            <div class="flex gap-1 justify-content-end">
                                <Button
                                    icon="ti ti-pencil"
                                    text
                                    rounded
                                    class="p-1 bookmarkfile-edit"
                                    aria-label="Edit"
                                    @click="openEdit(data.name)"
                                />
                                <Button
                                    icon="ti ti-trash"
                                    text
                                    rounded
                                    severity="danger"
                                    class="p-1 bookmarkfile-delete"
                                    aria-label="Delete"
                                    @click="showDeleteDialog(data.name)"
                                />
                            </div>
                        </template>
                    </Column>
                </DataTable>
            </template>
        </Card>
    </div>

    <Dialog
        v-model:visible="createDialogVisible"
        modal
        :draggable="false"
        header="New Bookmarks File"
        style="min-width: min(400px, calc(100vw - 1rem))"
    >
        <div class="flex flex-column gap-2">
            <label class="text-sm font-semibold">Filename</label>
            <InputText
                v-model="newFilename"
                placeholder="links.yaml"
                autofocus
                @keyup.enter="confirmCreate"
            />
            <small v-if="createError" class="text-red-500">{{ createError }}</small>
        </div>
        <div class="flex justify-content-end gap-2 mt-4">
            <Button label="Cancel" icon="ti ti-x" severity="secondary" @click="createDialogVisible = false" />
            <Button label="Create" icon="ti ti-plus" :loading="creating" @click="confirmCreate" />
        </div>
    </Dialog>

    <Dialog
        v-model:visible="editDialogVisible"
        modal
        :draggable="false"
        :header="editFilename ?? 'Bookmarks'"
        style="width: min(640px, calc(100vw - 1rem))"
    >
        <div v-if="editLoading" class="text-sm" style="color: var(--p-text-muted-color)">Loading...</div>
        <div v-else-if="editLoadError" class="text-sm text-red-500 bookmarkfile-load-error">
            {{ editLoadError }} Fix the file on disk, or delete it and create it again.
        </div>
        <template v-else>
            <small class="block mb-3" style="color: var(--p-text-muted-color)">
                Bookmark groups set to "From file" show these bookmarks, with the layout set in the widget.
            </small>
            <div v-if="!editBookmarks.length" class="text-sm mb-2 bookmarkfile-empty" style="color: var(--p-text-muted-color)">
                No bookmarks yet.
            </div>
            <BookmarkListEditor v-model="editBookmarks" />
        </template>
        <div class="flex align-items-center justify-content-end gap-2 mt-3">
            <small v-if="saveError" class="text-red-500 bookmarkfile-save-error">{{ saveError }}</small>
            <Button label="Close" icon="ti ti-x" severity="secondary" @click="editDialogVisible = false" />
            <Button
                label="Save"
                icon="ti ti-device-floppy"
                :loading="saving"
                :disabled="!!editLoadError || editLoading || hasUnsafeUrl"
                class="bookmarkfile-save"
                @click="save"
            />
        </div>
    </Dialog>

    <ConfirmDialog
        v-model:visible="deleteDialogVisible"
        :name="fileToDelete ?? ''"
        title="Delete Bookmarks File"
        message="Are you sure you want to delete this file? Widgets that load it will show an error."
        @confirm="confirmDelete"
    />
</template>
