<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import Tabs from 'primevue/tabs'
import TabList from 'primevue/tablist'
import Tab from 'primevue/tab'
import TabPanels from 'primevue/tabpanels'
import TabPanel from 'primevue/tabpanel'
import Button from 'primevue/button'
import SelectButton from 'primevue/selectbutton'
import { getWidgetEntry } from '@/lib/widgetRegistry'
import { allDocMocks, getWidgetDocs, type WidgetParamDoc } from '@/lib/widgetDocs'
import { provideDocQueryClient } from '@/lib/docQueryClient'
import WidgetPreview from '@/components/docs/WidgetPreview.vue'
import DocText from '@/components/docs/DocText.vue'

// One widget's documentation: a short summary, then an Overview tab (the
// widget rendered from sample data next to its real settings form, wired
// live) and a Details tab (parameter reference). Data comes from the mock
// QueryClient, which also serves the settings form (e.g. sysinfo's disks).
provideDocQueryClient(allDocMocks())

const route = useRoute()
const type = computed(() => String(route.params.type ?? ''))
const entry = computed(() => getWidgetEntry(type.value))
const docs = computed(() => getWidgetDocs(type.value))

const tab = ref('overview')
const config = ref<Record<string, unknown>>({})
// Bumped on reset so the form re-reads its initial values.
const formKey = ref(0)

const reset = () => {
    config.value = structuredClone(docs.value?.preview.config ?? {})
    formKey.value++
}

watch(type, () => {
    tab.value = 'overview'
    reset()
}, { immediate: true })

// Forms emit either a complete object or just the keys they own (same merge
// as the widget settings dialog).
const onUpdateConfig = (update: Record<string, unknown>) => {
    config.value = { ...config.value, ...update }
}

const presets = computed(() => docs.value?.preview.presets ?? [])
const showForm = computed(() => !!entry.value?.configComponent && !docs.value?.preview.hideForm)
const activePreset = computed({
    get: () => presets.value.find(p => JSON.stringify(p.config) === JSON.stringify(config.value))?.label ?? null,
    set: (label: string | null) => {
        const preset = presets.value.find(p => p.label === label)
        if (preset) config.value = structuredClone(preset.config)
    },
})

// A wide settings form (laid out in columns) goes under the preview.
const stacked = computed(() => entry.value?.settingsSize === 'wide')

const paramGroups = computed(() => {
    const groups: { name: string; params: WidgetParamDoc[] }[] = []
    for (const p of docs.value?.params ?? []) {
        const name = p.group ?? ''
        const last = groups[groups.length - 1]
        if (last && last.name === name) last.params.push(p)
        else groups.push({ name, params: [p] })
    }
    return groups
})
</script>

<template>
    <section v-if="entry && docs" class="doc-widget" :data-type="type">
        <router-link :to="{ name: 'doc-widgets' }" class="doc-widget-back">
            <i class="ti ti-arrow-left" aria-hidden="true" /> All widgets
        </router-link>
        <h2 class="doc-widget-title">
            <i :class="`ti ${entry.icon}`" aria-hidden="true" />
            {{ entry.label }}
        </h2>
        <p class="doc-widget-summary"><DocText :text="docs.summary" /></p>

        <Tabs v-model:value="tab" class="doc-widget-tabs">
            <TabList>
                <Tab value="overview">Overview</Tab>
                <Tab value="details">Details</Tab>
            </TabList>
            <TabPanels>
                <TabPanel value="overview">
                    <div class="doc-overview" :class="{ 'doc-overview--stacked': stacked }">
                        <div class="doc-overview-preview">
                            <div class="doc-pane-header">
                                <h3 class="doc-pane-title">Preview</h3>
                            </div>
                            <WidgetPreview :key="type" :type="type" :config="config" />
                        </div>
                        <div class="doc-overview-settings">
                            <div class="doc-pane-header">
                                <h3 class="doc-pane-title">Settings</h3>
                                <Button
                                    v-if="docs.preview.config"
                                    label="Reset"
                                    icon="ti ti-restore"
                                    text
                                    size="small"
                                    @click="reset"
                                />
                            </div>
                            <p v-if="docs.preview.settingsNote" class="doc-hint">{{ docs.preview.settingsNote }}</p>
                            <SelectButton
                                v-if="presets.length"
                                v-model="activePreset"
                                :options="presets.map(p => p.label)"
                                :allow-empty="false"
                                aria-label="Sample"
                            />
                            <component
                                :is="entry.configComponent"
                                v-else-if="showForm"
                                :key="`${type}-${formKey}`"
                                :config="config"
                                @update:config="onUpdateConfig"
                            />
                            <p v-else-if="!docs.preview.settingsNote" class="doc-hint">This widget has no settings.</p>
                        </div>
                    </div>
                </TabPanel>
                <TabPanel value="details">
                    <h3>Parameters</h3>
                    <p v-if="!docs.params.length">{{ docs.paramsNote ?? 'This widget has no configuration properties.' }}</p>
                    <template v-for="group in paramGroups" :key="group.name">
                        <h4 v-if="group.name">{{ group.name }}</h4>
                        <table class="doc-params">
                            <colgroup>
                                <col class="doc-params-name" />
                                <col class="doc-params-type" />
                                <col class="doc-params-default" />
                                <col />
                            </colgroup>
                            <thead>
                                <tr><th>Parameter</th><th>Type</th><th>Default</th><th>Description</th></tr>
                            </thead>
                            <tbody>
                                <tr v-for="p in group.params" :key="p.name" :data-param="p.name">
                                    <td><code>{{ p.name }}</code></td>
                                    <td>{{ p.type }}</td>
                                    <td><code v-if="p.default">{{ p.default }}</code><template v-else>&mdash;</template></td>
                                    <td><DocText :text="p.description" /></td>
                                </tr>
                            </tbody>
                        </table>
                    </template>

                    <template v-for="section in docs.sections ?? []" :key="section.title">
                        <h3>{{ section.title }}</h3>
                        <p v-for="(t, i) in section.text ?? []" :key="i"><DocText :text="t" /></p>
                        <ul v-if="section.list">
                            <li v-for="(item, i) in section.list" :key="i"><DocText :text="item" /></li>
                        </ul>
                        <table v-if="section.table">
                            <thead>
                                <tr><th v-for="h in section.table.head" :key="h">{{ h }}</th></tr>
                            </thead>
                            <tbody>
                                <tr v-for="(row, r) in section.table.rows" :key="r">
                                    <td v-for="(cell, c) in row" :key="c"><DocText :text="cell" /></td>
                                </tr>
                            </tbody>
                        </table>
                    </template>

                    <p v-for="(note, i) in docs.notes ?? []" :key="i" class="doc-hint"><DocText :text="note" /></p>
                </TabPanel>
            </TabPanels>
        </Tabs>
    </section>

    <section v-else class="doc-widget">
        <router-link :to="{ name: 'doc-widgets' }" class="doc-widget-back">
            <i class="ti ti-arrow-left" aria-hidden="true" /> All widgets
        </router-link>
        <h2>Unknown widget</h2>
        <p>There is no widget called <code>{{ type }}</code>.</p>
    </section>
</template>

<style scoped lang="scss">
@use '@/assets/breakpoints' as bp;

.doc-widget-back {
    display: inline-flex;
    align-items: center;
    gap: 0.3rem;
    margin-bottom: 0.75rem;
    font-size: 0.85rem;
    color: var(--p-primary-color);
    text-decoration: none;

    &:hover {
        text-decoration: underline;
    }
}

.doc-widget-title {
    display: flex;
    align-items: center;
    gap: 0.5rem;

    i {
        color: var(--p-primary-color);
    }
}

.doc-content .doc-widget-summary {
    font-size: 0.95rem;
    margin-bottom: 1rem;
}

/* The tab panels sit on the page, not in a card. */
.doc-widget-tabs :deep(.p-tabpanels) {
    padding: 1rem 0 0;
    background: transparent;
}

/* Same column widths in every group's table. */
.doc-params {
    table-layout: fixed;
}

.doc-params td code {
    white-space: normal;
    overflow-wrap: anywhere;
}

.doc-params-name {
    width: 26%;
}

.doc-params-type {
    width: 13%;
}

.doc-params-default {
    width: 14%;
}

.doc-overview {
    display: grid;
    grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
    gap: 2rem;
    align-items: start;

    @include bp.below-lg {
        grid-template-columns: minmax(0, 1fr);
        gap: 1.5rem;
    }
}

.doc-overview--stacked {
    grid-template-columns: minmax(0, 1fr);

    .doc-overview-preview {
        max-width: 28rem;
    }
}

.doc-overview-preview,
.doc-overview-settings {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    min-width: 0;
}

/* Side by side, the preview stays in view while a long form scrolls. */
.doc-overview:not(.doc-overview--stacked) .doc-overview-preview {
    position: sticky;
    top: 1rem;

    @include bp.below-lg {
        position: static;
    }
}

.doc-pane-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.5rem;
    min-height: 2.25rem;
}

.doc-content .doc-pane-title {
    margin: 0;
    font-size: 0.75rem;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--p-text-muted-color);
}
</style>
