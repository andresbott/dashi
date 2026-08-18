<script setup lang="ts">
import { computed, ref, inject, watch } from 'vue'
import draggable from 'vuedraggable'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import { getWidgetEntry, getWidgetTypeOptions } from '@/lib/widgetRegistry'
import WidgetPlaceholder from '@/components/dashboards/WidgetPlaceholder.vue'
import type { Widget } from '@/types/dashboard'
import { EDITING_MODE } from '@/lib/injectionKeys'
import { v4 as uuidv4 } from 'uuid'

const props = defineProps<{
    widget: Widget
}>()

const emit = defineEmits<{
    'update:widget': [widget: Widget]
}>()

const isEditing = inject(EDITING_MODE, false)

// Local ref so vuedraggable can update immediately (computed v-model
// round-trips through the emit chain too slowly for cross-group drags).
const localChildren = ref<Widget[]>([])

// Sync props -> local when parent data changes
watch(
    () => {
        const config = props.widget.config as Record<string, unknown> | undefined
        return (config?.widgets as Widget[]) ?? []
    },
    (val) => { localChildren.value = [...val] },
    { immediate: true, deep: true },
)

// Emit changes to parent (called by vuedraggable @change and manual mutations)
const emitChildren = () => {
    const filtered = localChildren.value.filter((w) => w.type !== 'stack')
    emit('update:widget', {
        ...props.widget,
        config: { ...props.widget.config, widgets: filtered },
    })
}

// Writable computed for manual mutations (add/delete/settings) that
// need both immediate local update and parent emit.
const children = computed({
    get: () => localChildren.value,
    set: (val: Widget[]) => {
        localChildren.value = val
        emitChildren()
    },
})

// --- Add widget dialog ---
const addWidgetDialogVisible = ref(false)
const widgetTypeOptions = computed(() =>
    getWidgetTypeOptions().filter((o) => o.value !== 'stack'),
)

const addWidget = (type: string) => {
    const option = widgetTypeOptions.value.find((o) => o.value === type)
    const newChild: Widget = {
        id: uuidv4(),
        type,
        title: option?.label ?? 'New Widget',
        width: 6,
    }
    children.value = [...children.value, newChild]
    addWidgetDialogVisible.value = false
}

const deleteChild = (index: number) => {
    children.value = children.value.filter((_, i) => i !== index)
}

// --- Child settings dialog ---
const childSettingsVisible = ref(false)
const editingChildIndex = ref<number | null>(null)
const editChildTitle = ref('')
const editChildConfig = ref<Record<string, unknown> | null>(null)

const openChildSettings = (index: number) => {
    const child = children.value[index]
    editingChildIndex.value = index
    editChildTitle.value = child.title
    editChildConfig.value = child.config ? { ...child.config } : {}
    childSettingsVisible.value = true
}

const saveChildSettings = () => {
    if (editingChildIndex.value === null) return
    const updated = [...children.value]
    updated[editingChildIndex.value] = {
        ...updated[editingChildIndex.value],
        title: editChildTitle.value,
        config: editChildConfig.value ?? undefined,
    }
    children.value = updated
    childSettingsVisible.value = false
    editingChildIndex.value = null
}

const onUpdateChildConfig = (config: Record<string, unknown>) => {
    editChildConfig.value = config
}

const editingChildEntry = computed(() => {
    if (editingChildIndex.value === null) return undefined
    return getWidgetEntry(children.value[editingChildIndex.value].type)
})
</script>

<template>
    <!-- View mode -->
    <div v-if="!isEditing" class="widget-stack">
        <div v-for="child in children" :key="child.id" class="stack-child">
            <component
                v-if="getWidgetEntry(child.type)"
                :is="getWidgetEntry(child.type)!.component"
                v-bind="
                    getWidgetEntry(child.type)!.noWidgetProp
                        ? {}
                        : { widget: { ...child, width: props.widget.width } }
                "
            />
            <WidgetPlaceholder v-else :title="child.title" />
        </div>
    </div>

    <!-- Edit mode -->
    <div v-else class="widget-stack-editor">
        <draggable
            v-model="localChildren"
            group="widgets"
            item-key="id"
            handle=".stack-drag-handle"
            class="stack-children"
            @change="emitChildren"
        >
            <template #item="{ element, index }">
                <div class="stack-child-editor">
                    <div class="stack-child-controls flex align-items-center gap-1">
                        <div class="stack-drag-handle" style="cursor: grab">
                            <i class="ti ti-grip-vertical" style="color: var(--p-text-muted-color)" />
                        </div>
                        <span class="text-sm flex-grow-1">{{ element.title }} <span style="color: var(--p-text-muted-color)">({{ element.type }})</span></span>
                        <Button
                            icon="ti ti-pencil"
                            text
                            rounded
                            class="p-1"
                            @click="openChildSettings(index)"
                        />
                        <Button
                            icon="ti ti-trash"
                            text
                            rounded
                            severity="danger"
                            class="p-1"
                            @click="deleteChild(index)"
                        />
                    </div>
                    <div class="stack-child-preview">
                        <component
                            v-if="getWidgetEntry(element.type)"
                            :is="getWidgetEntry(element.type)!.component"
                            v-bind="
                                getWidgetEntry(element.type)!.noWidgetProp
                                    ? {}
                                    : { widget: { ...element, width: props.widget.width } }
                            "
                        />
                        <WidgetPlaceholder v-else :title="element.title" />
                    </div>
                </div>
            </template>
        </draggable>

        <Button
            icon="ti ti-plus"
            label="Add"
            text
            size="small"
            class="mt-1"
            @click="addWidgetDialogVisible = true"
        />
    </div>

    <!-- Add widget dialog -->
    <Dialog
        v-model:visible="addWidgetDialogVisible"
        header="Add Widget to Stack"
        modal
        :closable="true"
        :draggable="false"
        style="width: 24rem"
    >
        <div class="flex flex-column gap-1">
            <div
                v-for="opt in widgetTypeOptions"
                :key="opt.value"
                class="widget-type-option"
                @click="addWidget(opt.value)"
            >
                <i :class="'ti ' + opt.icon" class="widget-type-icon" />
                <div>
                    <div class="font-semibold">{{ opt.label }}</div>
                    <div class="text-xs" style="color: var(--p-text-muted-color)">{{ opt.description }}</div>
                </div>
            </div>
        </div>
    </Dialog>

    <!-- Child settings dialog -->
    <Dialog
        v-model:visible="childSettingsVisible"
        header="Widget Settings"
        modal
        :closable="true"
        :draggable="false"
        style="width: 28rem"
    >
        <div class="flex flex-column gap-3">
            <div v-if="!editingChildEntry?.configComponent" class="flex flex-column gap-1">
                <label class="text-sm font-semibold">Title</label>
                <InputText v-model="editChildTitle" placeholder="Widget title" @keydown.enter="saveChildSettings" />
            </div>
            <component
                v-if="editingChildEntry?.configComponent"
                :is="editingChildEntry.configComponent"
                :config="editChildConfig"
                @update:config="onUpdateChildConfig"
            />
        </div>
        <div class="flex justify-content-end gap-3 mt-4">
            <Button label="Save" icon="ti ti-check" @click="saveChildSettings" />
            <Button label="Cancel" icon="ti ti-x" severity="secondary" @click="childSettingsVisible = false" />
        </div>
    </Dialog>
</template>

<style scoped>
.widget-stack-editor {
    min-height: 3rem;
    padding-top: 2.5rem;
}

.stack-children {
    min-height: 2rem;
}

.stack-child-editor {
    background: var(--p-surface-50);
    border: 1px solid var(--p-surface-200);
    border-radius: 6px;
    padding: 0.25rem 0.5rem;
    margin-bottom: 0.25rem;
}

.stack-child-preview {
    pointer-events: none;
    opacity: 0.8;
}

.widget-type-option {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.75rem;
    border-radius: 6px;
    cursor: pointer;
}

.widget-type-option:hover {
    background: var(--p-surface-100);
}

.widget-type-icon {
    font-size: 1.5rem;
    width: 2rem;
    text-align: center;
    flex-shrink: 0;
}
</style>
