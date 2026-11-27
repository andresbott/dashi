<script setup lang="ts">
import { ref, computed, nextTick, useId, type ComponentPublicInstance } from 'vue'
import Button from 'primevue/button'
import Popover from 'primevue/popover'
import InputText from 'primevue/inputtext'
import { tablerIcons } from '@/data/tablerIcons'

const props = defineProps<{
    modelValue: string
    // id of the visible field label; with the current value it forms the
    // trigger's accessible name ("Icon bookmark").
    labelledBy: string
}>()

const emit = defineEmits<{
    'update:modelValue': [value: string]
}>()

const popover = ref()
const searchInput = ref<ComponentPublicInstance | null>(null)
const open = ref(false)
const search = ref('')
const valueId = useId()

// The stored value is the CSS class (ti-bookmark); people know it as "bookmark".
const displayName = (icon: string) => icon.replace(/^ti-/, '')
const iconName = computed(() => displayName(props.modelValue))

const MAX_DISPLAY = 150

const filteredIcons = computed(() => {
    if (!search.value) return tablerIcons.slice(0, MAX_DISPLAY)
    const q = search.value.toLowerCase()
    const matched = tablerIcons.filter(i => i.includes(q))
    return matched.slice(0, MAX_DISPLAY)
})

const totalMatches = computed(() => {
    if (!search.value) return tablerIcons.length
    return tablerIcons.filter(i => i.includes(search.value.toLowerCase())).length
})

const toggle = (event: Event) => {
    popover.value.toggle(event)
}

// The search field is rendered with the panel, so focus it once that is in.
const onShow = async () => {
    open.value = true
    await nextTick()
    const input = searchInput.value?.$el as HTMLElement | undefined
    input?.focus()
}

const selectIcon = (icon: string) => {
    emit('update:modelValue', icon)
    popover.value.hide()
}
</script>

<template>
    <div>
        <button
            type="button"
            class="icon-selector-trigger"
            aria-haspopup="dialog"
            :aria-expanded="open"
            :aria-labelledby="`${labelledBy} ${valueId}`"
            data-testid="icon-selector-trigger"
            @click="toggle"
        >
            <i v-if="modelValue" :class="['ti', modelValue]" />
            <span :id="valueId" class="icon-selector-label" :class="{ 'icon-selector-placeholder': !modelValue }">
                {{ iconName || 'Select icon' }}
            </span>
            <i class="ti ti-chevron-down icon-selector-chevron" />
        </button>
        <Popover ref="popover" @show="onShow" @hide="open = false">
            <div class="icon-selector-panel">
                <InputText
                    ref="searchInput"
                    v-model="search"
                    placeholder="Search 5000+ icons..."
                    size="small"
                    class="w-full mb-2"
                />
                <div class="flex flex-wrap gap-1 icon-selector-grid">
                    <Button
                        v-for="icon in filteredIcons"
                        :key="icon"
                        :icon="'ti ' + icon"
                        text
                        rounded
                        class="p-1"
                        :severity="icon === modelValue ? 'primary' : 'secondary'"
                        data-testid="icon-option"
                        @click="selectIcon(icon)"
                        v-tooltip.top="displayName(icon)"
                    />
                </div>
                <div v-if="totalMatches > MAX_DISPLAY" class="text-xs text-color-secondary mt-1">
                    Showing {{ MAX_DISPLAY }} of {{ totalMatches }} matches. Refine your search.
                </div>
            </div>
        </Popover>
    </div>
</template>

<style scoped>
/* Looks like an InputText: it uses the form-field tokens InputText and Select
   are themed from, so it follows the theme and colour mode. */
.icon-selector-trigger {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    width: 100%;
    padding: var(--p-form-field-padding-y) var(--p-form-field-padding-x);
    font: inherit;
    text-align: left;
    color: var(--p-form-field-color);
    background: var(--p-form-field-background);
    border: 1px solid var(--p-form-field-border-color);
    border-radius: var(--p-form-field-border-radius);
    box-shadow: var(--p-form-field-shadow);
    cursor: pointer;
    transition: border-color var(--p-form-field-transition-duration);
}

.icon-selector-trigger:hover {
    border-color: var(--p-form-field-hover-border-color);
}

.icon-selector-trigger:focus-visible {
    border-color: var(--p-form-field-focus-border-color);
    outline: var(--p-form-field-focus-ring-width) var(--p-form-field-focus-ring-style) var(--p-form-field-focus-ring-color);
    outline-offset: var(--p-form-field-focus-ring-offset);
}

.icon-selector-label {
    flex: 1 1 auto;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.icon-selector-placeholder {
    color: var(--p-form-field-placeholder-color);
}

.icon-selector-chevron {
    color: var(--p-form-field-icon-color);
}

.icon-selector-panel {
    width: 300px;
}

.icon-selector-grid {
    max-height: 250px;
    overflow-y: auto;
}
</style>
