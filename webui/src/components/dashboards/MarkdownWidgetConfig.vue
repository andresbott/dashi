<script setup lang="ts">
import { ref, watch, inject, computed } from 'vue'
import { DASHBOARD_ID } from '@/lib/injectionKeys'
import { getMarkdownRaw, saveMarkdown } from '@/lib/api/markdown'
import { useMarkdownFiles } from '@/composables/useMarkdownFiles'
import InputText from 'primevue/inputtext'
import Textarea from 'primevue/textarea'
import Button from 'primevue/button'
import Select from 'primevue/select'

const props = defineProps<{
    config: { filename?: string } | null
}>()

const emit = defineEmits<{
    'update:config': [config: { filename: string }]
}>()

const dashboardId = inject(DASHBOARD_ID, ref(''))
const { files, isLoading: filesLoading, invalidate: refreshFiles } = useMarkdownFiles(dashboardId)

const configFilename = computed(() => props.config?.filename ?? '')

type Mode = 'select' | 'create' | 'edit'
const mode = ref<Mode>('select')

const selected = ref<string | null>(null)
const missing = computed(() => {
    if (filesLoading.value) return false
    const c = configFilename.value
    if (!c) return false
    return !files.value.includes(c)
})

const editContent = ref('')
const editLoading = ref(false)
const loadError = ref(false)
const saving = ref(false)
const saveStatus = ref<'idle' | 'saved' | 'error'>('idle')

const newFilename = ref('')
const createError = ref('')
const creating = ref(false)

const loadContent = async (filename: string) => {
    if (!dashboardId.value || !filename) return
    editLoading.value = true
    loadError.value = false
    saveStatus.value = 'idle'
    try {
        editContent.value = await getMarkdownRaw(dashboardId.value, filename)
    } catch {
        editContent.value = ''
        loadError.value = true
    } finally {
        editLoading.value = false
    }
}

const syncFromProps = () => {
    const c = configFilename.value
    if (!c) {
        selected.value = null
        mode.value = 'select'
        return
    }
    if (files.value.includes(c)) {
        selected.value = c
        mode.value = 'edit'
        loadContent(c)
    } else {
        selected.value = null
        mode.value = 'select'
    }
}

watch([files, () => props.config, () => dashboardId.value], syncFromProps, { immediate: true })

const onSelect = (value: string | null) => {
    if (!value) return
    selected.value = value
    mode.value = 'edit'
    emit('update:config', { filename: value })
}

const startCreate = () => {
    newFilename.value = ''
    createError.value = ''
    mode.value = 'create'
}

const cancelCreate = () => {
    newFilename.value = ''
    createError.value = ''
    mode.value = selected.value ? 'edit' : 'select'
}

const validateNewFilename = (raw: string): { ok: true; name: string } | { ok: false; error: string } => {
    let name = raw.trim()
    if (!name) return { ok: false, error: 'Filename is required' }
    if (name.length > 100) return { ok: false, error: 'Filename is too long' }
    if (name.includes('/') || name.includes('..')) return { ok: false, error: 'Invalid characters in filename' }
    if (!name.toLowerCase().endsWith('.md')) name += '.md'
    if (files.value.includes(name)) return { ok: false, error: `"${name}" already exists` }
    return { ok: true, name }
}

const confirmCreate = async () => {
    const result = validateNewFilename(newFilename.value)
    if (!result.ok) {
        createError.value = result.error
        return
    }
    if (!dashboardId.value) return
    creating.value = true
    createError.value = ''
    try {
        await saveMarkdown(dashboardId.value, result.name, '')
        await refreshFiles()
        selected.value = result.name
        editContent.value = ''
        mode.value = 'edit'
        emit('update:config', { filename: result.name })
    } catch {
        createError.value = 'Failed to create file'
    } finally {
        creating.value = false
    }
}

const save = async () => {
    if (!selected.value || !dashboardId.value || loadError.value) return
    saving.value = true
    saveStatus.value = 'idle'
    try {
        await saveMarkdown(dashboardId.value, selected.value, editContent.value)
        saveStatus.value = 'saved'
    } catch {
        saveStatus.value = 'error'
    } finally {
        saving.value = false
    }
}

const CREATE_SENTINEL = '__create__'
const options = computed(() => {
    const opts = files.value.map((f) => ({ label: f, value: f }))
    opts.push({ label: '+ New file…', value: CREATE_SENTINEL })
    return opts
})

const onDropdownChange = (value: string | null) => {
    if (value === CREATE_SENTINEL) {
        startCreate()
        const current = configFilename.value
        selected.value = current && files.value.includes(current) ? current : null
        return
    }
    onSelect(value)
}
</script>

<template>
    <div class="flex flex-column gap-3">
        <div class="flex flex-column gap-1">
            <label class="text-sm font-semibold">File</label>
            <Select
                :modelValue="selected"
                :options="options"
                optionLabel="label"
                optionValue="value"
                placeholder="Select a markdown file"
                :loading="filesLoading"
                class="w-full"
                @update:modelValue="onDropdownChange"
            />
            <small v-if="missing" class="text-red-500">
                "{{ configFilename }}" is missing — pick another or create a new one.
            </small>
            <small v-else class="text-color-secondary">
                Files from the dashboard's <code>md/</code> folder
            </small>
        </div>

        <div v-if="mode === 'create'" class="flex flex-column gap-1">
            <label class="text-sm font-semibold">New filename</label>
            <InputText
                v-model="newFilename"
                placeholder="notes.md"
                @keyup.enter="confirmCreate"
                autofocus
            />
            <small v-if="createError" class="text-red-500">{{ createError }}</small>
            <div class="flex align-items-center gap-2">
                <Button
                    label="Create"
                    icon="ti ti-plus"
                    size="small"
                    :loading="creating"
                    @click="confirmCreate"
                />
                <Button
                    label="Cancel"
                    severity="secondary"
                    size="small"
                    :disabled="creating"
                    @click="cancelCreate"
                />
            </div>
        </div>

        <div v-if="mode === 'edit'" class="flex flex-column gap-1">
            <label class="text-sm font-semibold">Content</label>
            <div v-if="editLoading" class="text-sm" style="color: var(--p-text-muted-color)">
                <i class="ti ti-loader-2 md-config-spinner" /> Loading...
            </div>
            <template v-else>
                <div v-if="loadError" class="text-sm text-red-500">
                    Failed to load file. Try selecting it again.
                </div>
                <Textarea
                    v-else
                    v-model="editContent"
                    :autoResize="true"
                    rows="10"
                    class="w-full md-editor"
                />
                <div v-if="!loadError" class="flex align-items-center gap-2">
                    <Button
                        label="Save"
                        icon="ti ti-device-floppy"
                        size="small"
                        :loading="saving"
                        :disabled="loadError"
                        @click="save"
                    />
                    <small v-if="saveStatus === 'saved'" class="text-green-500">Saved</small>
                    <small v-if="saveStatus === 'error'" class="text-red-500">Failed to save</small>
                </div>
            </template>
        </div>
    </div>
</template>

<style scoped>
.md-editor {
    font-family: monospace;
    font-size: 0.85rem;
    line-height: 1.5;
}

@keyframes spin {
    to { transform: rotate(360deg); }
}

.md-config-spinner {
    animation: spin 1s linear infinite;
}
</style>
