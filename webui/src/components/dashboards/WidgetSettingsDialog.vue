<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import InputText from 'primevue/inputtext'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import { getWidgetEntry } from '@/lib/widgetRegistry'
import { plainCopy } from '@/lib/plainCopy'
import type { Widget } from '@/types/dashboard'

// Shared widget settings dialog: edits a staged copy of the widget and hands
// the result back on Save, so Cancel never leaks a partial edit. Used by row
// cards and by the stack editor for its children. Deleting is not a setting:
// it lives in the card's (or stack child's) header bar.
const props = defineProps<{
    widget: Widget
}>()

const visible = defineModel<boolean>('visible', { default: false })

const emit = defineEmits<{
    save: [widget: Widget]
}>()

const editTitle = ref('')
const editConfig = ref<Record<string, unknown>>({})

const entry = computed(() => getWidgetEntry(props.widget.type))

// A widget can ask for a wide dialog (registration `settingsSize`); the global
// .p-dialog max-width still caps it to the screen on phones.
const dialogWidth = computed(() => (entry.value?.settingsSize === 'wide' ? '60rem' : '28rem'))

// Stage a fresh deep copy every time the dialog opens, so nested config
// (arrays, objects, a stack's children) stays detached from the live widget
// until Save.
watch(visible, (open) => {
    if (!open) return
    editTitle.value = props.widget.title
    editConfig.value = props.widget.config ? plainCopy(props.widget.config) : {}
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
            <div class="flex align-items-center justify-content-end gap-3 w-full">
                <Button label="Save" icon="ti ti-check" @click="save" />
                <Button label="Cancel" icon="ti ti-x" severity="secondary" @click="visible = false" />
            </div>
        </template>
    </Dialog>
</template>
