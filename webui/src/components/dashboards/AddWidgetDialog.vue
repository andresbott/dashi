<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import Dialog from 'primevue/dialog'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import AddWidgetPreview from '@/components/dashboards/AddWidgetPreview.vue'
import { getWidgetTypeOptions } from '@/lib/widgetRegistry'

// Shared widget-type picker used by rows and by container widgets (stack).
// The host decides where the picked widget goes; `exclude` hides types a host
// cannot contain (a stack cannot nest another stack). Clicking a widget
// selects and previews it; Insert (or a double-click) hands the type to the
// host, which adds it with its default config, not the preview's sample one.
const props = withDefaults(defineProps<{
    header?: string
    exclude?: string[]
}>(), {
    header: 'Add Widget',
    exclude: () => [],
})

const visible = defineModel<boolean>('visible', { default: false })

const emit = defineEmits<{
    select: [type: string]
}>()

const filter = ref('')
const selectedType = ref<string | null>(null)

// Every opening starts from a clean picker.
watch(visible, (open) => {
    if (!open) return
    filter.value = ''
    selectedType.value = null
})

const options = computed(() => getWidgetTypeOptions().filter(o => !props.exclude.includes(o.value)))

const filteredOptions = computed(() => {
    const query = filter.value.trim().toLowerCase()
    if (!query) return options.value
    return options.value.filter(o =>
        [o.label, o.description, o.value].some(text => text.toLowerCase().includes(query)),
    )
})

const insert = (type: string) => {
    emit('select', type)
    visible.value = false
}

const insertSelected = () => {
    if (selectedType.value) insert(selectedType.value)
}
</script>

<template>
    <Dialog
        v-model:visible="visible"
        :header="header"
        modal
        :closable="true"
        :draggable="false"
        style="width: 48rem"
    >
        <div class="add-widget-body">
            <div class="add-widget-list-pane">
                <InputText
                    v-model="filter"
                    type="text"
                    placeholder="Filter widgets..."
                    aria-label="Filter widgets"
                    fluid
                    autofocus
                />
                <div class="add-widget-list flex flex-column gap-1">
                    <button
                        v-for="opt in filteredOptions"
                        :key="opt.value"
                        type="button"
                        class="widget-type-option"
                        :class="{ 'widget-type-option--selected': opt.value === selectedType }"
                        :data-type="opt.value"
                        :aria-pressed="opt.value === selectedType"
                        @click="selectedType = opt.value"
                        @dblclick="insert(opt.value)"
                    >
                        <i :class="'ti ' + opt.icon" class="widget-type-icon" aria-hidden="true" />
                        <span class="widget-type-text">
                            <span class="font-semibold">{{ opt.label }}</span>
                            <span class="text-xs widget-type-description">{{ opt.description }}</span>
                        </span>
                    </button>
                    <p v-if="filteredOptions.length === 0" class="add-widget-empty text-sm">
                        No widgets match "{{ filter.trim() }}".
                    </p>
                </div>
            </div>

            <div class="add-widget-preview-pane" aria-live="polite">
                <AddWidgetPreview v-if="selectedType" :key="selectedType" :type="selectedType" />
                <p v-else class="add-widget-preview-hint text-sm">
                    Select a widget to see a preview.
                </p>
            </div>
        </div>

        <template #footer>
            <div class="flex align-items-center justify-content-end gap-3 w-full">
                <Button label="Insert" icon="ti ti-plus" :disabled="!selectedType" @click="insertSelected" />
                <Button label="Cancel" icon="ti ti-x" severity="secondary" @click="visible = false" />
            </div>
        </template>
    </Dialog>
</template>

<style scoped lang="scss">
@use '@/assets/breakpoints' as bp;

.add-widget-body {
    display: grid;
    grid-template-columns: minmax(0, 1fr) minmax(0, 1.4fr);
    gap: 1rem;

    @include bp.below-md {
        grid-template-columns: minmax(0, 1fr);
    }
}

.add-widget-list-pane {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    min-width: 0;
}

.add-widget-list {
    max-height: 26rem;
    overflow-y: auto;

    @include bp.below-md {
        max-height: 14rem;
    }
}

.add-widget-preview-pane {
    min-width: 0;
    padding: 0.75rem;
    border-radius: 8px;
    background: var(--p-surface-50);
}

.add-widget-preview-hint,
.add-widget-empty {
    margin: 0;
    color: var(--p-text-muted-color);
}

.add-widget-preview-hint {
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 8rem;
    text-align: center;
}

.widget-type-option {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.75rem;
    border: 0;
    border-radius: 6px;
    background: none;
    color: var(--p-text-color);
    font: inherit;
    text-align: left;
    cursor: pointer;
}

.widget-type-option:hover,
.widget-type-option:focus-visible {
    background: var(--p-surface-100);
    outline: none;
}

/* PrimeVue's selected-option tokens, as in its Listbox. */
.widget-type-option--selected,
.widget-type-option--selected:hover,
.widget-type-option--selected:focus-visible {
    background: var(--p-highlight-background);
    color: var(--p-highlight-color);
}

.widget-type-icon {
    font-size: 1.5rem;
    width: 2rem;
    text-align: center;
    flex-shrink: 0;
}

.widget-type-text {
    display: flex;
    flex-direction: column;
}

.widget-type-description {
    color: var(--p-text-muted-color);
}
</style>
