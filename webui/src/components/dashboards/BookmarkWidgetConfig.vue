<script setup lang="ts">
import { ref, watch } from 'vue'
import InputText from 'primevue/inputtext'
import IconSelector from '@/components/dashboards/IconSelector.vue'
import type { BookmarkWidgetConfig } from '@/types/bookmark'

const props = defineProps<{
    config: BookmarkWidgetConfig | null
}>()

const emit = defineEmits<{
    'update:config': [config: BookmarkWidgetConfig]
}>()

const editUrl = ref(props.config?.url ?? '')
const editIcon = ref(props.config?.icon ?? 'ti-bookmark')
const editTitle = ref(props.config?.title ?? '')
const editSubtitle = ref(props.config?.subtitle ?? '')

watch(() => props.config, (val) => {
    if (val) {
        editUrl.value = val.url
        editIcon.value = val.icon
        editTitle.value = val.title
        editSubtitle.value = val.subtitle
    }
})

const emitUpdate = () => {
    emit('update:config', {
        url: editUrl.value,
        icon: editIcon.value,
        title: editTitle.value,
        subtitle: editSubtitle.value,
    })
}
</script>

<template>
    <div class="flex flex-column gap-3">
        <div class="flex flex-column gap-1">
            <label class="text-sm font-semibold">Icon</label>
            <IconSelector v-model="editIcon" @update:modelValue="emitUpdate" />
        </div>
        <div class="flex flex-column gap-1">
            <label class="text-sm font-semibold">Title</label>
            <InputText v-model="editTitle" placeholder="Bookmark title" @input="emitUpdate" />
        </div>
        <div class="flex flex-column gap-1">
            <label class="text-sm font-semibold">Subtitle</label>
            <InputText v-model="editSubtitle" placeholder="Short description" @input="emitUpdate" />
        </div>
        <div class="flex flex-column gap-1">
            <label class="text-sm font-semibold">URL</label>
            <InputText v-model="editUrl" placeholder="https://example.com" @input="emitUpdate" />
        </div>
    </div>
</template>
