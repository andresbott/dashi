<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import Card from 'primevue/card'
import Button from 'primevue/button'
import Tag from 'primevue/tag'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import { useListThemes } from '@/composables/useThemes'
import { pageBgValue } from '@/lib/themeCss'
import { plainCopy } from '@/lib/plainCopy'
import type { ThemeMeta } from '@/types/theme'
import { useToast } from 'primevue/usetoast'

const router = useRouter()
const toast = useToast()
const {
    themes: data,
    isLoading,
    createTheme,
    isCreating,
    deleteTheme,
    copyTheme,
    isCopying,
} = useListThemes()

const themes = computed(() => data.value ?? [])

// The builtins are always listed; the default one's palettes seed a new theme.
const builtin = computed(() => themes.value.find(t => t.id === 'default') ?? null)

const openEditor = (id: string) => {
    router.push({ name: 'admin-theme-edit', params: { id } })
}

const handleCreate = async () => {
    if (!builtin.value) return
    try {
        const created = await createTheme({
            name: 'New theme',
            darkMode: true,
            background: {},
            // A plain copy: query data is a reactive proxy.
            colors: plainCopy(builtin.value.colors),
        })
        openEditor(created.id)
    } catch {
        toast.add({ severity: 'error', summary: 'Error', detail: 'Failed to create theme', life: 5000 })
    }
}

// Copy works on every row, the builtin included: it is how the builtin gets
// customised. The copy opens straight in the editor.
const handleCopy = async (row: ThemeMeta) => {
    try {
        const copy = await copyTheme(row.id)
        openEditor(copy.id)
    } catch {
        toast.add({ severity: 'error', summary: 'Error', detail: 'Failed to copy theme', life: 5000 })
    }
}

const deleteDialogVisible = ref(false)
const toDelete = ref<ThemeMeta | null>(null)

const showDeleteDialog = (row: ThemeMeta) => {
    toDelete.value = row
    deleteDialogVisible.value = true
}

// Deleting a referenced theme is allowed: those dashboards fall back to the
// default theme. The count is here so the choice is an informed one.
const deleteMessage = computed(() => {
    const n = toDelete.value?.usedBy ?? 0
    if (n === 0) return 'Are you sure you want to delete this theme?'
    return `${n} dashboard${n === 1 ? '' : 's'} use this theme and will fall back to the default theme. Delete anyway?`
})

const confirmDelete = async () => {
    if (!toDelete.value) return
    try {
        await deleteTheme(toDelete.value.id)
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
            <h2 class="admin-section-title">Themes</h2>
            <Button
                label="New theme"
                icon="ti ti-plus"
                :loading="isCreating"
                :disabled="!builtin"
                @click="handleCreate"
            />
        </div>

        <Card v-if="!themes.length && !isLoading">
            <template #content>
                <div class="info-message">No themes yet. Create one to get started.</div>
            </template>
        </Card>

        <Card v-else>
            <template #content>
                <DataTable
                    :value="themes"
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
                    <Column header="Name">
                        <template #body="{ data: row }">
                            <span>{{ row.name }}</span>
                            <Tag
                                v-if="row.builtin"
                                value="Built-in"
                                severity="info"
                                class="ml-2"
                            />
                        </template>
                    </Column>
                    <Column header="Colours" style="width: 120px" headerClass="hide-below-md" bodyClass="hide-below-md">
                        <template #body="{ data: row }">
                            <div class="flex gap-1">
                                <span class="color-chip" :style="{ background: row.colors.light.primary }" title="Primary" />
                                <span class="color-chip" :style="{ background: row.colors.light.secondary }" title="Secondary" />
                                <span class="color-chip" :style="{ background: row.colors.light.surface }" title="Card surface" />
                            </div>
                        </template>
                    </Column>
                    <Column header="Used by" style="width: 160px" headerClass="hide-below-md" bodyClass="hide-below-md">
                        <template #body="{ data: row }">
                            <span :class="{ 'text-muted': row.usedBy === 0 }">
                                {{ row.usedBy }} dashboard{{ row.usedBy === 1 ? '' : 's' }}
                            </span>
                        </template>
                    </Column>
                    <Column header="Actions" style="width: 140px">
                        <template #body="{ data: row }">
                            <div class="flex gap-1 justify-content-end">
                                <Button
                                    v-if="!row.builtin"
                                    icon="ti ti-pencil"
                                    aria-label="Edit"
                                    text
                                    rounded
                                    class="p-1"
                                    @click="openEditor(row.id)"
                                />
                                <Button
                                    icon="ti ti-copy"
                                    aria-label="Copy"
                                    text
                                    rounded
                                    class="p-1"
                                    :disabled="isCopying"
                                    @click="handleCopy(row)"
                                />
                                <Button
                                    v-if="!row.builtin"
                                    icon="ti ti-trash"
                                    aria-label="Delete"
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
        title="Delete theme"
        :message="deleteMessage"
        @confirm="confirmDelete"
    />
</template>

<style scoped>
/* .admin-section* and .info-message are shared: see assets/style.scss. */
.swatch {
    width: 64px;
    height: 40px;
    border-radius: 4px;
    border: 1px solid var(--p-content-border-color);
    background-size: cover;
}

.color-chip {
    width: 18px;
    height: 18px;
    border-radius: 50%;
    border: 1px solid var(--p-content-border-color);
}

.text-muted {
    color: var(--p-text-muted-color);
}
</style>
