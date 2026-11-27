<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import Checkbox from 'primevue/checkbox'
import SelectButton from 'primevue/selectbutton'
import Slider from 'primevue/slider'
import Select from 'primevue/select'
import { useBookmarkFiles } from '@/composables/useBookmarkFiles'
import BookmarkListEditor from './BookmarkListEditor.vue'
import { isSafeBookmarkUrl } from '../bookmark/url'
import { MAX_GROUP_COLUMNS } from './types'
import type { BookmarkGroupDirection, BookmarkGroupItem, BookmarkGroupSource, BookmarkGroupWidgetConfig } from './types'

const props = defineProps<{
    config: BookmarkGroupWidgetConfig | null
}>()

const emit = defineEmits<{
    'update:config': [config: BookmarkGroupWidgetConfig]
}>()

const bookmarks = ref<BookmarkGroupItem[]>([])
const textBelow = ref(false)
const newTab = ref(false)
const direction = ref<BookmarkGroupDirection>('horizontal')
const columns = ref(0)
const source = ref<BookmarkGroupSource>('dashboard')
const file = ref('')

const sourceOptions = [
    { label: 'In dashboard', value: 'dashboard' },
    { label: 'From file', value: 'file' },
]

// The content files to pick from. A stored name that no longer exists is
// still offered, so the form shows what the widget points at.
const { files } = useBookmarkFiles()
const fileOptions = computed(() =>
    file.value && !files.value.includes(file.value) ? [file.value, ...files.value] : files.value,
)

const directionOptions = [
    { label: 'Horizontal', value: 'horizontal' },
    { label: 'Vertical', value: 'vertical' },
]

function normalizeColumns(n: unknown): number {
    return typeof n === 'number' && Number.isInteger(n) && n > 0 && n <= MAX_GROUP_COLUMNS ? n : 0
}

// The config shape this form emits, from whatever is stored; also used to
// tell an echo of our own update apart from a config handed in from outside.
function normalize(cfg: Partial<BookmarkGroupWidgetConfig> | null): BookmarkGroupWidgetConfig {
    return {
        bookmarks: (cfg?.bookmarks ?? []).map(b => ({
            url: b.url ?? '',
            icon: b.icon ?? '',
            title: b.title ?? '',
            subtitle: b.subtitle ?? '',
        })),
        textBelow: cfg?.textBelow ?? false,
        newTab: cfg?.newTab ?? false,
        direction: cfg?.direction === 'vertical' ? 'vertical' : 'horizontal',
        columns: normalizeColumns(cfg?.columns),
        source: cfg?.source === 'file' ? 'file' : 'dashboard',
        file: cfg?.file ?? '',
    }
}

const current = (): BookmarkGroupWidgetConfig => ({
    bookmarks: bookmarks.value.map(({ url, icon, title, subtitle }) => ({ url, icon, title, subtitle })),
    textBelow: textBelow.value,
    newTab: newTab.value,
    direction: direction.value,
    columns: columns.value,
    source: source.value,
    file: file.value,
})

function load(cfg: BookmarkGroupWidgetConfig | null) {
    const n = normalize(cfg)
    bookmarks.value = n.bookmarks
    textBelow.value = n.textBelow ?? false
    newTab.value = n.newTab ?? false
    direction.value = n.direction ?? 'horizontal'
    columns.value = n.columns ?? 0
    source.value = n.source ?? 'dashboard'
    file.value = n.file ?? ''
}

load(props.config)

watch(() => props.config, (val) => {
    if (JSON.stringify(normalize(val)) !== JSON.stringify(current())) {
        load(val)
    }
})

// Same rule the server enforces on save (see ../bookmark/url.ts). While any
// bookmark fails it, nothing is emitted, so the dashboard is never
// autosaved with it.
const hasUnsafeUrl = computed(() => bookmarks.value.some(b => !isSafeBookmarkUrl(b.url)))

const emitUpdate = () => {
    if (hasUnsafeUrl.value) return
    emit('update:config', current())
}

</script>

<template>
    <div class="flex flex-column gap-3">
        <div class="flex flex-column gap-1">
            <label class="text-sm font-semibold">Direction</label>
            <SelectButton
                v-model="direction"
                :options="directionOptions"
                optionLabel="label"
                optionValue="value"
                :allowEmpty="false"
                class="bookmarkgroup-direction"
                @update:modelValue="emitUpdate"
            />
        </div>
        <div v-if="direction === 'horizontal'" class="flex flex-column gap-1">
            <!-- 0 is "Auto": as many per row as the width allows. -->
            <label class="text-sm font-semibold">Bookmarks per row: <span class="bookmarkgroup-columns-value">{{ columns || 'Auto' }}</span></label>
            <Slider
                v-model="columns"
                :min="0"
                :max="MAX_GROUP_COLUMNS"
                :step="1"
                class="bookmarkgroup-columns w-full"
                @slideend="emitUpdate"
            />
        </div>
        <div class="flex align-items-center gap-2">
            <Checkbox v-model="textBelow" :binary="true" inputId="bookmarkgroupTextBelow" @update:modelValue="emitUpdate" />
            <label for="bookmarkgroupTextBelow" class="text-sm">Text below icon</label>
        </div>
        <div class="flex align-items-center gap-2">
            <Checkbox v-model="newTab" :binary="true" inputId="bookmarkgroupNewTab" @update:modelValue="emitUpdate" />
            <label for="bookmarkgroupNewTab" class="text-sm">Open in a new tab</label>
        </div>

        <div class="flex flex-column gap-1">
            <label class="text-sm font-semibold">Bookmarks source</label>
            <SelectButton
                v-model="source"
                :options="sourceOptions"
                optionLabel="label"
                optionValue="value"
                :allowEmpty="false"
                class="bookmarkgroup-source"
                @update:modelValue="emitUpdate"
            />
        </div>

        <div v-if="source === 'file'" class="flex flex-column gap-1">
            <label for="bookmarkgroupFile" class="text-sm font-semibold">Bookmarks file</label>
            <Select
                v-model="file"
                :options="fileOptions"
                inputId="bookmarkgroupFile"
                placeholder="Choose a file"
                class="bookmarkgroup-file w-full"
                @update:modelValue="emitUpdate"
            />
            <small class="bookmarkgroup-file-hint" style="color: var(--p-text-muted-color)">
                YAML files in content/bookmarks, managed under Admin → Content → Bookmarks. Changes made outside the editor show up within 10 minutes.
            </small>
        </div>

        <div v-else class="flex flex-column gap-2">
            <span class="text-sm font-semibold">Bookmarks</span>
            <BookmarkListEditor v-model="bookmarks" @update:modelValue="emitUpdate" />
        </div>
    </div>
</template>
