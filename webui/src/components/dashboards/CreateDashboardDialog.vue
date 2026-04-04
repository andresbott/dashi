<script setup lang="ts">
import { ref, watch } from 'vue'
import Dialog from 'primevue/dialog'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import type { DashboardType, Container } from '@/types/dashboard'

const visible = defineModel<boolean>('visible', { default: false })
const emit = defineEmits<{
    confirm: [payload: { name: string; type: DashboardType; container: Container }]
}>()

const name = ref('New Dashboard')
const dashboardType = ref<DashboardType>('interactive')
const isSaving = defineModel<boolean>('saving', { default: false })

const typeOptions = [
    { label: 'Interactive', value: 'interactive' },
    { label: 'Static', value: 'static' },
    { label: 'Image', value: 'image' },
]

watch(visible, (val) => {
    if (val) {
        name.value = 'New Dashboard'
        dashboardType.value = 'interactive'
    }
})

const handleSave = () => {
    emit('confirm', {
        name: name.value,
        type: dashboardType.value,
        container: {
            maxWidth: '100%',
            verticalAlign: 'top',
            horizontalAlign: 'center',
        },
    })
}
</script>

<template>
    <Dialog
        v-model:visible="visible"
        modal
        :closable="true"
        :draggable="false"
        header="New Dashboard"
    >
        <div class="flex flex-column gap-3" style="min-width: 350px">
            <div class="flex align-items-center gap-2">
                <InputText
                    v-model="name"
                    placeholder="Dashboard name"
                    class="flex-grow-1"
                    autofocus
                    @keydown.enter="handleSave"
                />
            </div>
            <Select
                v-model="dashboardType"
                :options="typeOptions"
                optionLabel="label"
                optionValue="value"
                placeholder="Dashboard type"
                class="w-full"
            />
        </div>
        <div class="flex justify-content-end gap-3 mt-4">
            <Button
                type="button"
                label="Save"
                icon="ti ti-check"
                :loading="isSaving"
                @click="handleSave"
            />
            <Button
                type="button"
                label="Cancel"
                icon="ti ti-x"
                severity="secondary"
                @click="visible = false"
            />
        </div>
    </Dialog>
</template>
