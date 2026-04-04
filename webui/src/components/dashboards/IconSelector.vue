<script setup lang="ts">
import { ref, computed } from 'vue'
import Button from 'primevue/button'
import Popover from 'primevue/popover'
import InputText from 'primevue/inputtext'

const props = defineProps<{
    modelValue: string
}>()

const emit = defineEmits<{
    'update:modelValue': [value: string]
}>()

const popover = ref()
const search = ref('')

const icons = [
    'ti-layout-dashboard', 'ti-home', 'ti-chart-bar', 'ti-chart-line',
    'ti-chart-pie', 'ti-server', 'ti-database', 'ti-cloud',
    'ti-shield', 'ti-lock', 'ti-users', 'ti-settings',
    'ti-bell', 'ti-mail', 'ti-calendar', 'ti-clock',
    'ti-bookmark', 'ti-star', 'ti-heart', 'ti-flag',
    'ti-folder', 'ti-file', 'ti-code', 'ti-terminal',
    'ti-bug', 'ti-rocket', 'ti-bolt', 'ti-flame',
    'ti-world', 'ti-map', 'ti-compass', 'ti-sun',
    'ti-moon', 'ti-eye', 'ti-camera', 'ti-music',
    'ti-microphone', 'ti-video', 'ti-phone', 'ti-message',
    'ti-brand-github', 'ti-brand-docker', 'ti-cpu', 'ti-device-desktop',
    'ti-wifi', 'ti-battery', 'ti-plug', 'ti-palette'
]

const filteredIcons = computed(() => {
    if (!search.value) return icons
    return icons.filter(i => i.includes(search.value.toLowerCase()))
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
            <div style="width: 280px">
                <InputText
                    v-model="search"
                    placeholder="Search icons..."
                    size="small"
                    class="w-full mb-2"
                />
                <div class="flex flex-wrap gap-1" style="max-height: 200px; overflow-y: auto">
                    <Button
                        v-for="icon in filteredIcons"
                        :key="icon"
                        :icon="'ti ' + icon"
                        text
                        rounded
                        class="p-1"
                        :severity="icon === modelValue ? 'primary' : 'secondary'"
                        @click="selectIcon(icon)"
                    />
                </div>
            </div>
        </Popover>
    </div>
</template>
