<script setup lang="ts">
import { ref, computed } from 'vue'
import Card from 'primevue/card'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import InputText from 'primevue/inputtext'
import Textarea from 'primevue/textarea'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import { useAdminNotes } from '@/composables/useAdminNotes'
import { getMarkdownRaw, saveMarkdown } from '@/widgets/markdown/api'
import { useToast } from 'primevue/usetoast'

const toast = useToast()
const { notes: notesData, isLoading, deleteNote, refetch } = useAdminNotes()
const notes = computed(() => notesData.value ?? [])

const deleteDialogVisible = ref(false)
const noteToDelete = ref<string | null>(null)

const showDeleteDialog = (name: string) => {
    noteToDelete.value = name
    deleteDialogVisible.value = true
}

const confirmDelete = async () => {
    if (!noteToDelete.value) return
    try {
        await deleteNote(noteToDelete.value)
        deleteDialogVisible.value = false
        noteToDelete.value = null
    } catch {
        toast.add({ severity: 'error', summary: 'Error', detail: 'Failed to delete note', life: 5000 })
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
    if (!name.toLowerCase().endsWith('.md')) name += '.md'
    if (notes.value.includes(name)) return { ok: false, error: `"${name}" already exists` }
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
        await saveMarkdown(result.name, '')
        await refetch()
        createDialogVisible.value = false
        openEdit(result.name)
    } catch {
        createError.value = 'Failed to create file'
    } finally {
        creating.value = false
    }
}

// Edit dialog
const editDialogVisible = ref(false)
const editFilename = ref<string | null>(null)
const editContent = ref('')
const editLoading = ref(false)
const editLoadError = ref(false)
const saving = ref(false)
const saveError = ref(false)

const openEdit = async (name: string) => {
    editFilename.value = name
    editContent.value = ''
    editLoadError.value = false
    saveError.value = false
    editDialogVisible.value = true
    editLoading.value = true
    try {
        editContent.value = await getMarkdownRaw(name)
    } catch {
        editLoadError.value = true
    } finally {
        editLoading.value = false
    }
}

const save = async () => {
    if (!editFilename.value || editLoadError.value) return
    saving.value = true
    saveError.value = false
    try {
        await saveMarkdown(editFilename.value, editContent.value)
        editDialogVisible.value = false
    } catch {
        saveError.value = true
    } finally {
        saving.value = false
    }
}
</script>

<template>
    <div class="admin-section">
        <div class="admin-section-header">
            <h2 class="admin-section-title">Notes</h2>
            <Button label="New" icon="ti ti-plus" @click="openCreate" />
        </div>

        <Card v-if="!notes.length && !isLoading">
            <template #content>
                <div class="info-message">No notes yet. Create your first note to get started.</div>
            </template>
        </Card>

        <Card v-else>
            <template #content>
                <DataTable
                    :value="notes.map((n) => ({ name: n }))"
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
                                    class="p-1"
                                    @click="openEdit(data.name)"
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

    <Dialog
        v-model:visible="createDialogVisible"
        modal
        :draggable="false"
        header="New Note"
        style="min-width: 400px"
    >
        <div class="flex flex-column gap-2">
            <label class="text-sm font-semibold">Filename</label>
            <InputText
                v-model="newFilename"
                placeholder="notes.md"
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
        :header="editFilename ?? 'Note'"
        style="min-width: 600px"
    >
        <div v-if="editLoading" class="text-sm" style="color: var(--p-text-muted-color)">Loading...</div>
        <div v-else-if="editLoadError" class="text-sm text-red-500">Failed to load file.</div>
        <Textarea
            v-else
            v-model="editContent"
            :autoResize="true"
            rows="15"
            class="w-full md-editor"
        />
        <div class="flex align-items-center justify-content-end gap-2 mt-3">
            <small v-if="saveError" class="text-red-500">Failed to save</small>
            <Button label="Close" icon="ti ti-x" severity="secondary" @click="editDialogVisible = false" />
            <Button
                label="Save"
                icon="ti ti-device-floppy"
                :loading="saving"
                :disabled="editLoadError || editLoading"
                @click="save"
            />
        </div>
    </Dialog>

    <ConfirmDialog
        v-model:visible="deleteDialogVisible"
        :name="noteToDelete ?? ''"
        title="Delete Note"
        message="Are you sure you want to delete this note?"
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

.md-editor {
    font-family: monospace;
    font-size: 0.85rem;
    line-height: 1.5;
}
</style>
