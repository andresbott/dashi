<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import DashboardRow from '@/components/dashboards/DashboardRow.vue'

import { useGetDashboard, useUpdateDashboard } from '@/composables/useDashboards'
import { useToast } from 'primevue/usetoast'
import type { Dashboard, Row } from '@/types/dashboard'
import { onBeforeUnmount } from 'vue'
import {
    createDashboard as apiCreateDashboard,
    updateDashboard as apiUpdateDashboard,
    deleteDashboard as apiDeleteDashboard
} from '@/lib/api/dashboard'
import { v4 as uuidv4 } from 'uuid'
import Dialog from 'primevue/dialog'
import Select from 'primevue/select'
import Checkbox from 'primevue/checkbox'

const route = useRoute()
const router = useRouter()
const toast = useToast()
const id = computed(() => route.params.id as string)

const { data: serverDashboard, isLoading, isError } = useGetDashboard(() => id.value)
const { updateDashboard, isUpdating } = useUpdateDashboard()

const localDashboard = ref<Dashboard | null>(null)
const containerDialogVisible = ref(false)

watch(serverDashboard, (val) => {
    if (val && !localDashboard.value) {
        localDashboard.value = JSON.parse(JSON.stringify(val))
    }
}, { immediate: true })

watch(() => localDashboard.value?.type, (newType) => {
    if (!localDashboard.value) return
    if (newType === 'image' && !localDashboard.value.imageConfig) {
        localDashboard.value.imageConfig = { width: 1024, height: 0 }
    }
})

const rows = computed(() => localDashboard.value?.rows ?? [])

const addRow = () => {
    if (!localDashboard.value) return
    localDashboard.value.rows.push({
        id: uuidv4(),
        height: 'auto',
        width: '100%',
        widgets: []
    })
}

const updateRow = (index: number, row: Row) => {
    if (!localDashboard.value) return
    localDashboard.value.rows[index] = row
}

const deleteRow = (index: number) => {
    if (!localDashboard.value) return
    localDashboard.value.rows.splice(index, 1)
}

const moveRowUp = (index: number) => {
    if (!localDashboard.value || index <= 0) return
    const rows = localDashboard.value.rows
    ;[rows[index - 1], rows[index]] = [rows[index], rows[index - 1]]
}

const moveRowDown = (index: number) => {
    if (!localDashboard.value) return
    const rows = localDashboard.value.rows
    if (index >= rows.length - 1) return
    ;[rows[index], rows[index + 1]] = [rows[index + 1], rows[index]]
}

const save = async () => {
    if (!localDashboard.value) return
    try {
        await updateDashboard({ id: id.value, payload: localDashboard.value })
        toast.add({ severity: 'success', summary: 'Saved', detail: 'Dashboard saved successfully', life: 3000 })
    } catch (err) {
        toast.add({ severity: 'error', summary: 'Error', detail: 'Failed to save dashboard', life: 5000 })
    }
}

const cancel = () => {
    router.push({ name: 'dashboards' })
}

const previewId = ref<string | null>(null)

const isPreviewing = ref(false)

const preview = async () => {
    if (!localDashboard.value) return
    isPreviewing.value = true
    try {
        const prevId = id.value + '-prev'
        const previewPayload = {
            id: prevId,
            name: localDashboard.value.name + ' - preview',
            icon: localDashboard.value.icon,
            type: localDashboard.value.type,
            container: JSON.parse(JSON.stringify(localDashboard.value.container)),
            imageConfig: localDashboard.value.imageConfig
                ? JSON.parse(JSON.stringify(localDashboard.value.imageConfig))
                : undefined,
            rows: JSON.parse(JSON.stringify(localDashboard.value.rows))
        }
        if (previewId.value) {
            await apiUpdateDashboard(prevId, { ...previewPayload } as Dashboard)
        } else {
            await apiCreateDashboard(previewPayload)
            previewId.value = prevId
        }
        const resolved = router.resolve({ name: 'dashboard-view', params: { id: previewId.value! } })
        window.open(resolved.href, '_blank')
    } catch (err) {
        console.error('Preview failed:', err)
        toast.add({ severity: 'error', summary: 'Error', detail: 'Failed to create preview', life: 5000 })
    } finally {
        isPreviewing.value = false
    }
}

const cleanupPreview = async () => {
    if (previewId.value) {
        try {
            await apiDeleteDashboard(previewId.value)
        } catch {
            // ignore cleanup errors
        }
        previewId.value = null
    }
}

onBeforeUnmount(cleanupPreview)
</script>

<template>
    <header class="app-topbar">
        <span class="app-topbar-title" @click="router.push('/')">Dashi</span>
    </header>
    <div class="dashboard-edit-view">
        <div v-if="isLoading" class="p-4">Loading...</div>
        <div v-else-if="isError" class="p-4">Failed to load dashboard.</div>
        <template v-else-if="localDashboard">
        <div class="flex align-items-center gap-2 mb-3">
            <span class="text-xl font-bold text-color flex-grow-1">{{ localDashboard.name }}</span>
            <Button
                icon="ti ti-settings"
                label="Settings"
                severity="secondary"
                @click="containerDialogVisible = true"
            />
            <Button
                icon="ti ti-eye"
                label="Preview"
                severity="secondary"
                :loading="isPreviewing"
                @click="preview"
            />
            <Button
                label="Save"
                icon="ti ti-check"
                :loading="isUpdating"
                @click="save"
            />
            <Button
                label="Cancel"
                icon="ti ti-x"
                severity="secondary"
                @click="cancel"
            />
        </div>

        <DashboardRow
            v-for="(row, index) in rows"
            :key="row.id"
            :row="row"
            :is-first="index === 0"
            :is-last="index === rows.length - 1"
            @update="updateRow(index, $event)"
            @delete="deleteRow(index)"
            @move-up="moveRowUp(index)"
            @move-down="moveRowDown(index)"
        />

        <div class="mt-2">
            <Button
                label="Add Row"
                icon="ti ti-plus"
                severity="secondary"
                @click="addRow"
            />
        </div>

        <Dialog
            v-model:visible="containerDialogVisible"
            modal
            :closable="true"
            :draggable="false"
            header="Dashboard Settings"
        >
            <div class="flex flex-column gap-3" style="min-width: 350px">
                <div class="flex flex-column gap-1">
                    <label class="font-semibold text-sm">Name</label>
                    <InputText v-model="localDashboard.name" placeholder="Dashboard name" />
                </div>
                <div class="flex flex-column gap-1">
                    <label class="font-semibold text-sm">Type</label>
                    <Select
                        v-model="localDashboard.type"
                        :options="[
                            { label: 'Interactive', value: 'interactive' },
                            { label: 'Static', value: 'static' },
                            { label: 'Image', value: 'image' },
                        ]"
                        optionLabel="label"
                        optionValue="value"
                        class="w-full"
                    />
                </div>
                <div class="flex flex-column gap-1">
                    <label class="font-semibold text-sm">Max Width</label>
                    <InputText v-model="localDashboard.container.maxWidth" placeholder="e.g. 1200px, 80%, 100%" />
                </div>
                <div class="flex flex-column gap-1">
                    <label class="font-semibold text-sm">Vertical Align</label>
                    <Select
                        v-model="localDashboard.container.verticalAlign"
                        :options="[
                            { label: 'Top', value: 'top' },
                            { label: 'Center', value: 'center' },
                            { label: 'Bottom', value: 'bottom' },
                        ]"
                        optionLabel="label"
                        optionValue="value"
                        class="w-full"
                    />
                </div>
                <div class="flex flex-column gap-1">
                    <label class="font-semibold text-sm">Horizontal Align</label>
                    <Select
                        v-model="localDashboard.container.horizontalAlign"
                        :options="[
                            { label: 'Left', value: 'left' },
                            { label: 'Center', value: 'center' },
                            { label: 'Right', value: 'right' },
                        ]"
                        optionLabel="label"
                        optionValue="value"
                        class="w-full"
                    />
                </div>
                <div class="flex align-items-center gap-2">
                    <Checkbox v-model="localDashboard.container.showBoxes" :binary="true" inputId="showBoxes" />
                    <label for="showBoxes" class="font-semibold text-sm">Show boxes</label>
                </div>
                <template v-if="localDashboard.type === 'image'">
                    <div class="flex flex-column gap-1">
                        <label class="font-semibold text-sm">Image Width (px)</label>
                        <InputText
                            :modelValue="String(localDashboard.imageConfig?.width ?? 1024)"
                            @update:modelValue="(v: string) => {
                                if (!localDashboard.imageConfig) localDashboard.imageConfig = { width: 1024, height: 0 }
                                localDashboard.imageConfig.width = parseInt(v) || 1024
                            }"
                            placeholder="1024"
                        />
                    </div>
                    <div class="flex flex-column gap-1">
                        <label class="font-semibold text-sm">Image Height (px, 0 = auto)</label>
                        <InputText
                            :modelValue="String(localDashboard.imageConfig?.height ?? 0)"
                            @update:modelValue="(v: string) => {
                                if (!localDashboard.imageConfig) localDashboard.imageConfig = { width: 1024, height: 0 }
                                localDashboard.imageConfig.height = parseInt(v) || 0
                            }"
                            placeholder="0 (auto)"
                        />
                    </div>
                </template>
            </div>
            <div class="flex justify-content-end mt-4">
                <Button label="Done" icon="ti ti-check" @click="containerDialogVisible = false" />
            </div>
        </Dialog>
        </template>

    </div>
</template>

<style scoped>
.dashboard-edit-view {
    max-width: 1600px;
    margin: 0 auto;
    padding: 1.5rem 1rem;
}
</style>
