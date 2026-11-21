<script setup lang="ts">
import { computed } from 'vue'
import Dialog from 'primevue/dialog'
import { getWidgetTypeOptions } from '@/lib/widgetRegistry'

// Shared widget-type picker used by rows and by container widgets (stack).
// The host decides where the picked widget goes; `exclude` hides types a host
// cannot contain (a stack cannot nest another stack).
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

const options = computed(() => getWidgetTypeOptions().filter(o => !props.exclude.includes(o.value)))

const pick = (type: string) => {
    emit('select', type)
    visible.value = false
}
</script>

<template>
    <Dialog
        v-model:visible="visible"
        :header="header"
        modal
        :closable="true"
        :draggable="false"
        style="width: 24rem"
    >
        <div class="flex flex-column gap-1">
            <button
                v-for="opt in options"
                :key="opt.value"
                type="button"
                class="widget-type-option"
                @click="pick(opt.value)"
            >
                <i :class="'ti ' + opt.icon" class="widget-type-icon" aria-hidden="true" />
                <span class="widget-type-text">
                    <span class="font-semibold">{{ opt.label }}</span>
                    <span class="text-xs widget-type-description">{{ opt.description }}</span>
                </span>
            </button>
        </div>
    </Dialog>
</template>

<style scoped>
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
