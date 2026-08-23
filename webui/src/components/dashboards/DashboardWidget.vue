<script setup lang="ts">
import { ref, computed } from 'vue'
import InputText from 'primevue/inputtext'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import { getWidgetEntry } from '@/lib/widgetRegistry'
import type { Widget } from '@/types/dashboard'

const props = defineProps<{
    widget: Widget
}>()

const emit = defineEmits<{
    update: [widget: Widget]
    delete: []
}>()

const settingsVisible = ref(false)
const editTitle = ref('')
const editConfig = ref<Record<string, unknown> | null>(null)

const entry = computed(() => getWidgetEntry(props.widget.type))

const openSettings = () => {
    editTitle.value = props.widget.title
    editConfig.value = props.widget.config ? { ...props.widget.config } : {}
    settingsVisible.value = true
}

const saveSettings = () => {
    const updated = { ...props.widget, title: editTitle.value }
    if (editConfig.value) {
        updated.config = editConfig.value
    }
    emit('update', updated)
    settingsVisible.value = false
}

const onDelete = () => {
    settingsVisible.value = false
    emit('delete')
}

const onUpdateConfig = (config: Record<string, unknown>) => {
    editConfig.value = config
}
</script>

<template>
    <div class="dashboard-widget">
        <div class="widget-header">
            <slot name="handle" />
            <Button
                icon="ti ti-pencil"
                text
                rounded
                class="p-1"
                @click="openSettings"
                v-tooltip.top="'Widget settings'"
            />
        </div>
        <div class="widget-body">
            <component
                v-if="entry"
                :is="entry.component"
                v-bind="entry.noWidgetProp ? {} : { widget }"
                @update:widget="emit('update', $event)"
            />
        </div>
        <span class="widget-width-label">{{ widget.width }}/12</span>
        <slot name="resize" />
    </div>

    <Dialog
        v-model:visible="settingsVisible"
        header="Widget Settings"
        modal
        :closable="true"
        :draggable="false"
        style="width: 28rem"
    >
        <div class="flex flex-column gap-3">
            <div v-if="!entry?.configComponent" class="flex flex-column gap-1">
                <label class="text-sm font-semibold">Title</label>
                <InputText v-model="editTitle" placeholder="Widget title" @keydown.enter="saveSettings" />
            </div>
            <component
                v-if="entry?.configComponent"
                :is="entry.configComponent"
                :config="editConfig"
                @update:config="onUpdateConfig"
            />
        </div>
        <div class="flex align-items-center gap-3 mt-4">
            <Button label="Delete" icon="ti ti-trash" severity="danger" text @click="onDelete" v-tooltip.top="'Delete widget'" />
            <span class="flex-grow-1"></span>
            <Button label="Save" icon="ti ti-check" @click="saveSettings" />
            <Button label="Cancel" icon="ti ti-x" severity="secondary" @click="settingsVisible = false" />
        </div>
    </Dialog>
</template>

<style scoped>
.dashboard-widget {
    position: relative;
    background: var(--p-surface-0);
    border: 1px solid var(--p-surface-300);
    border-radius: 8px;
}

/* Grip (drag handle) sits at the left, the edit button at the right, so both
   stay reachable even when the card is a single column wide. Delete lives in the
   settings dialog rather than the header to keep only these two targets here. */
.widget-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.25rem;
    min-height: 2rem;
    padding: 0.125rem 0.375rem;
    border-bottom: 1px solid var(--p-surface-200);
}

.widget-body {
    padding: 0.5rem 0.75rem 1.5rem;
}

.widget-width-label {
    position: absolute;
    bottom: 0.375rem;
    left: 0.625rem;
    font-size: 0.75rem;
    color: var(--p-text-muted-color);
    pointer-events: none;
}
</style>
