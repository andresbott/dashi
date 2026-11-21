<script setup lang="ts">
import { ref, computed, watch, toRaw } from 'vue'
import InputText from 'primevue/inputtext'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import { getWidgetEntry } from '@/lib/widgetRegistry'
import type { Widget } from '@/types/dashboard'

// Shared widget settings dialog: edits a staged copy of the widget and hands
// the result back on Save, so Cancel never leaks a partial edit. Used by row
// cards and by the stack editor for its children.
const props = withDefaults(defineProps<{
    widget: Widget
    // Show a Delete action (confirmed) in the dialog footer.
    deletable?: boolean
}>(), {
    deletable: true,
})

const visible = defineModel<boolean>('visible', { default: false })

const emit = defineEmits<{
    save: [widget: Widget]
    delete: []
}>()

const editTitle = ref('')
const editConfig = ref<Record<string, unknown>>({})
const confirmDeleteVisible = ref(false)

const entry = computed(() => getWidgetEntry(props.widget.type))

// A widget can ask for a wide dialog (registration `settingsSize`); the global
// .p-dialog max-width still caps it to the screen on phones.
const dialogWidth = computed(() => (entry.value?.settingsSize === 'wide' ? '60rem' : '28rem'))

// Stage a fresh copy every time the dialog opens; a structured clone keeps
// nested config (arrays, objects) detached from the live widget until Save.
// toRaw first: the config is a reactive proxy, which structuredClone rejects.
watch(visible, (open) => {
    if (!open) return
    editTitle.value = props.widget.title
    editConfig.value = props.widget.config ? structuredClone(toRaw(props.widget.config)) : {}
})

// Config forms emit either a complete object or just the keys they own, so
// always merge: a form can never drop keys it does not know about.
const onUpdateConfig = (config: Record<string, unknown>) => {
    editConfig.value = { ...editConfig.value, ...config }
}

const save = () => {
    emit('save', { ...props.widget, title: editTitle.value, config: editConfig.value })
    visible.value = false
}

const confirmDelete = () => {
    confirmDeleteVisible.value = false
    visible.value = false
    emit('delete')
}
</script>

<template>
    <Dialog
        v-model:visible="visible"
        header="Widget Settings"
        modal
        :closable="true"
        :draggable="false"
        :style="{ width: dialogWidth }"
    >
        <div class="flex flex-column gap-3" data-testid="widget-settings-body">
            <div v-if="!entry?.configComponent" class="flex flex-column gap-1">
                <label class="text-sm font-semibold" for="widget-title">Title</label>
                <InputText id="widget-title" v-model="editTitle" placeholder="Widget title" @keydown.enter="save" />
            </div>
            <component
                v-if="entry?.configComponent"
                :is="entry.configComponent"
                :config="editConfig"
                @update:config="onUpdateConfig"
            />
        </div>
        <!-- The footer sits outside PrimeVue's scrolling content area, so the
             actions stay visible when a long form scrolls. -->
        <template #footer>
            <div class="flex align-items-center gap-3 w-full">
                <Button
                    v-if="deletable"
                    label="Delete"
                    icon="ti ti-trash"
                    severity="danger"
                    text
                    @click="confirmDeleteVisible = true"
                />
                <span class="flex-grow-1"></span>
                <Button label="Save" icon="ti ti-check" @click="save" />
                <Button label="Cancel" icon="ti ti-x" severity="secondary" @click="visible = false" />
            </div>
        </template>
    </Dialog>

    <ConfirmDialog
        v-model:visible="confirmDeleteVisible"
        title="Delete widget"
        message="Delete this widget? This cannot be undone."
        :name="widget.title"
        @confirm="confirmDelete"
    />
</template>
