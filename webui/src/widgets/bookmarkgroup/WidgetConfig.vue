<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import InputText from 'primevue/inputtext'
import Checkbox from 'primevue/checkbox'
import Button from 'primevue/button'
import SelectButton from 'primevue/selectbutton'
import BookmarkIconField from '../bookmark/BookmarkIconField.vue'
import { isSafeBookmarkUrl, BOOKMARK_URL_ERROR } from '../bookmark/url'
import type { BookmarkGroupDirection, BookmarkGroupItem, BookmarkGroupWidgetConfig } from './types'

const props = defineProps<{
    config: BookmarkGroupWidgetConfig | null
}>()

const emit = defineEmits<{
    'update:config': [config: BookmarkGroupWidgetConfig]
}>()

// Each row carries a local key so Vue keeps the right DOM (and the open
// row) through reorders and removals. Keys never leave the form.
type Row = BookmarkGroupItem & { key: number }
let nextKey = 0

const rows = ref<Row[]>([])
const textBelow = ref(false)
const newTab = ref(false)
const direction = ref<BookmarkGroupDirection>('horizontal')
const openKey = ref<number | null>(null)

const directionOptions = [
    { label: 'Horizontal', value: 'horizontal' },
    { label: 'Vertical', value: 'vertical' },
]

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
    }
}

const current = (): BookmarkGroupWidgetConfig => ({
    bookmarks: rows.value.map(({ url, icon, title, subtitle }) => ({ url, icon, title, subtitle })),
    textBelow: textBelow.value,
    newTab: newTab.value,
    direction: direction.value,
})

function load(cfg: BookmarkGroupWidgetConfig | null) {
    const n = normalize(cfg)
    rows.value = n.bookmarks.map(b => ({ ...b, key: nextKey++ }))
    textBelow.value = n.textBelow ?? false
    newTab.value = n.newTab ?? false
    direction.value = n.direction ?? 'horizontal'
}

load(props.config)

watch(() => props.config, (val) => {
    if (JSON.stringify(normalize(val)) !== JSON.stringify(current())) {
        load(val)
        openKey.value = null
    }
})

// Same rule the server enforces on save (see ../bookmark/url.ts). While any
// bookmark fails it, nothing is emitted, so the dashboard is never
// autosaved with it.
const hasUnsafeUrl = computed(() => rows.value.some(r => !isSafeBookmarkUrl(r.url)))

const emitUpdate = () => {
    if (hasUnsafeUrl.value) return
    emit('update:config', current())
}

const toggle = (key: number) => {
    openKey.value = openKey.value === key ? null : key
}

const add = () => {
    const row: Row = { url: '', icon: 'ti-bookmark', title: '', subtitle: '', key: nextKey++ }
    rows.value.push(row)
    openKey.value = row.key
    emitUpdate()
}

const remove = (index: number) => {
    rows.value.splice(index, 1)
    emitUpdate()
}

const move = (index: number, by: -1 | 1) => {
    const target = index + by
    if (target < 0 || target >= rows.value.length) return
    const [row] = rows.value.splice(index, 1)
    rows.value.splice(target, 0, row)
    emitUpdate()
}

const onIconChange = (row: Row, icon: string) => {
    row.icon = icon
    emitUpdate()
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
        <div class="flex align-items-center gap-2">
            <Checkbox v-model="textBelow" :binary="true" inputId="bookmarkgroupTextBelow" @update:modelValue="emitUpdate" />
            <label for="bookmarkgroupTextBelow" class="text-sm">Text below icon</label>
        </div>
        <div class="flex align-items-center gap-2">
            <Checkbox v-model="newTab" :binary="true" inputId="bookmarkgroupNewTab" @update:modelValue="emitUpdate" />
            <label for="bookmarkgroupNewTab" class="text-sm">Open in a new tab</label>
        </div>

        <div class="flex flex-column gap-2">
            <span class="text-sm font-semibold">Bookmarks</span>
            <div
                v-for="(row, i) in rows"
                :key="row.key"
                class="bookmarkgroup-item"
                :class="{ 'bookmarkgroup-item--invalid': !isSafeBookmarkUrl(row.url), 'bookmarkgroup-item--open': openKey === row.key }"
            >
                <div class="bookmarkgroup-item-header">
                    <button
                        type="button"
                        class="bookmarkgroup-toggle"
                        :aria-expanded="openKey === row.key"
                        @click="toggle(row.key)"
                    >
                        <i class="ti" :class="openKey === row.key ? 'ti-chevron-down' : 'ti-chevron-right'" />
                        <span class="bookmarkgroup-item-title" :class="{ 'bookmarkgroup-item-title--empty': !row.title }">{{ row.title || 'Untitled bookmark' }}</span>
                        <i v-if="!isSafeBookmarkUrl(row.url)" class="ti ti-alert-circle text-red-500" aria-label="Invalid URL" />
                    </button>
                    <Button
                        class="bookmarkgroup-up" icon="ti ti-arrow-up" text rounded size="small"
                        aria-label="Move up" :disabled="i === 0" @click="move(i, -1)"
                    />
                    <Button
                        class="bookmarkgroup-down" icon="ti ti-arrow-down" text rounded size="small"
                        aria-label="Move down" :disabled="i === rows.length - 1" @click="move(i, 1)"
                    />
                    <Button
                        class="bookmarkgroup-remove" icon="ti ti-trash" text rounded size="small" severity="danger"
                        aria-label="Remove bookmark" @click="remove(i)"
                    />
                </div>
                <div v-if="openKey === row.key" class="bookmarkgroup-item-body flex flex-column gap-3">
                    <BookmarkIconField :modelValue="row.icon" @update:modelValue="(v: string) => onIconChange(row, v)" />
                    <div class="flex flex-column gap-1">
                        <label class="text-sm font-semibold">Title</label>
                        <InputText v-model="row.title" placeholder="Bookmark title" @input="emitUpdate" />
                    </div>
                    <div class="flex flex-column gap-1">
                        <label class="text-sm font-semibold">Subtitle</label>
                        <InputText v-model="row.subtitle" placeholder="Short description" @input="emitUpdate" />
                    </div>
                    <div class="flex flex-column gap-1">
                        <label class="text-sm font-semibold">URL</label>
                        <InputText v-model="row.url" placeholder="https://example.com" @input="emitUpdate" :invalid="!isSafeBookmarkUrl(row.url)" />
                        <small v-if="!isSafeBookmarkUrl(row.url)" class="text-red-500 bookmark-url-error">{{ BOOKMARK_URL_ERROR }}</small>
                    </div>
                </div>
            </div>
            <Button
                class="bookmarkgroup-add align-self-start" icon="ti ti-plus" label="Add bookmark"
                severity="secondary" size="small" @click="add"
            />
        </div>
    </div>
</template>

<style scoped>
.bookmarkgroup-item {
    border: 1px solid var(--p-content-border-color);
    border-radius: 6px;
}

.bookmarkgroup-item--invalid {
    border-color: var(--p-red-500);
}

.bookmarkgroup-item-header {
    display: flex;
    align-items: center;
    gap: 0.25rem;
    padding: 0.25rem;
}

.bookmarkgroup-toggle {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    flex: 1;
    min-width: 0;
    padding: 0.25rem 0.5rem;
    border: none;
    background: none;
    color: inherit;
    font: inherit;
    text-align: left;
    cursor: pointer;
}

.bookmarkgroup-item-title {
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
}

.bookmarkgroup-item-title--empty {
    color: var(--p-text-muted-color);
    font-style: italic;
}

.bookmarkgroup-item-body {
    padding: 0.5rem 0.75rem 0.75rem;
    border-top: 1px solid var(--p-content-border-color);
}
</style>
