<script setup lang="ts">
import { ref, computed } from 'vue'
import Button from 'primevue/button'
import Popover from 'primevue/popover'
import InputText from 'primevue/inputtext'
import { tablerIcons } from '@/data/tablerIcons'

const props = defineProps<{
    modelValue: string
}>()

const emit = defineEmits<{
    'update:modelValue': [value: string]
}>()

const popover = ref()
const search = ref('')

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

const selectIcon = (icon: string) => {
    emit('update:modelValue', icon)
    popover.value.hide()
}
</script>

<template>
    <div>
        <Button
            :icon="'ti ' + modelValue"
            text
            rounded
            class="p-2"
            @click="toggle"
            v-tooltip.top="'Select icon'"
        />
        <Popover ref="popover">
            <div style="width: 300px">
                <InputText
                    v-model="search"
                    placeholder="Search 5000+ icons..."
                    size="small"
                    class="w-full mb-2"
                />
                <div class="flex flex-wrap gap-1" style="max-height: 250px; overflow-y: auto">
                    <Button
                        v-for="icon in filteredIcons"
                        :key="icon"
                        :icon="'ti ' + icon"
                        text
                        rounded
                        class="p-1"
                        :severity="icon === modelValue ? 'primary' : 'secondary'"
                        @click="selectIcon(icon)"
                        v-tooltip.top="icon.replace('ti-', '')"
                    />
                </div>
                <div v-if="totalMatches > MAX_DISPLAY" class="text-xs text-color-secondary mt-1">
                    Showing {{ MAX_DISPLAY }} of {{ totalMatches }} matches. Refine your search.
                </div>
            </div>
        </Popover>
    </div>
</template>
