<script setup lang="ts">
import { ref, computed } from 'vue'
import InputText from 'primevue/inputtext'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import WidgetPlaceholder from '@/components/dashboards/WidgetPlaceholder.vue'
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

const onUpdateConfig = (config: Record<string, unknown>) => {
    editConfig.value = config
}
</script>

<template>
    <div class="dashboard-widget">
        <div class="widget-controls flex align-items-center gap-1">
            <Button
                icon="ti ti-pencil"
                text
                rounded
                class="p-1"
                @click="openSettings"
                v-tooltip.top="'Widget settings'"
            />
            <Button
                icon="ti ti-trash"
                text
                rounded
                severity="danger"
                class="p-1"
                @click="emit('delete')"
            />
        </div>
        <component
            v-if="entry"
            :is="entry.component"
            v-bind="entry.noWidgetProp ? {} : { widget }"
            @update:widget="emit('update', $event)"
        />
        <WidgetPlaceholder v-else :title="widget.title" />
        <span class="widget-width-label">{{ widget.width }}/12</span>
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
        <div class="flex justify-content-end gap-3 mt-4">
            <Button label="Save" icon="ti ti-check" @click="saveSettings" />
            <Button label="Cancel" icon="ti ti-x" severity="secondary" @click="settingsVisible = false" />
        </div>
    </Dialog>
</template>

<style scoped>
.dashboard-widget {
    background: var(--p-surface-0);
    border: 1px solid var(--p-surface-300);
    border-radius: 8px;
    padding: 0.5rem;
}

.widget-controls {
    position: absolute;
    top: 0.25rem;
    right: 0.25rem;
    z-index: 5;
}

.dashboard-widget {
    position: relative;
}

.widget-width-label {
    position: absolute;
    bottom: 0.25rem;
    right: 0.5rem;
    font-size: 0.75rem;
    color: var(--p-text-muted-color);
}
</style>
