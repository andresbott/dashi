<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import InputText from 'primevue/inputtext'
import Checkbox from 'primevue/checkbox'
import BookmarkIconField from './BookmarkIconField.vue'
import type { BookmarkWidgetConfig } from './types'
import { isSafeBookmarkUrl, BOOKMARK_URL_ERROR } from './url'

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
const editTextBelow = ref(props.config?.textBelow ?? false)
const editNewTab = ref(props.config?.newTab ?? false)

watch(() => props.config, (val) => {
    if (val) {
        editUrl.value = val.url
        editIcon.value = val.icon
        editTitle.value = val.title
        editSubtitle.value = val.subtitle
        editTextBelow.value = val.textBelow ?? false
        editNewTab.value = val.newTab ?? false
    }
})

// Same rule the server enforces on save (see ./url.ts). While the URL fails
// it, no change is emitted, so the dashboard is never autosaved with it.
const urlError = computed(() => (isSafeBookmarkUrl(editUrl.value) ? '' : BOOKMARK_URL_ERROR))

const emitUpdate = () => {
    if (urlError.value) return
    emit('update:config', {
        url: editUrl.value,
        icon: editIcon.value,
        title: editTitle.value,
        subtitle: editSubtitle.value,
        textBelow: editTextBelow.value,
        newTab: editNewTab.value,
    })
}

const onIconChange = (icon: string) => {
    editIcon.value = icon
    emitUpdate()
}
</script>

<template>
    <div class="flex flex-column gap-3">
        <BookmarkIconField :modelValue="editIcon" @update:modelValue="onIconChange" />
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
            <InputText v-model="editUrl" placeholder="https://example.com" @input="emitUpdate" :invalid="!!urlError" />
            <small v-if="urlError" class="text-red-500 bookmark-url-error">{{ urlError }}</small>
        </div>
        <div class="flex align-items-center gap-2">
            <Checkbox v-model="editTextBelow" :binary="true" inputId="bookmarkTextBelow" @update:modelValue="emitUpdate" />
            <label for="bookmarkTextBelow" class="text-sm">Text below icon</label>
        </div>
        <div class="flex align-items-center gap-2">
            <Checkbox v-model="editNewTab" :binary="true" inputId="bookmarkNewTab" @update:modelValue="emitUpdate" />
            <label for="bookmarkNewTab" class="text-sm">Open in a new tab</label>
        </div>
    </div>
</template>
