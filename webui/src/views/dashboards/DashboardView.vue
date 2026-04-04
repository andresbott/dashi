<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import WidgetPlaceholder from '@/components/dashboards/WidgetPlaceholder.vue'
import WeatherWidget from '@/components/dashboards/WeatherWidget.vue'
import BookmarkWidget from '@/components/dashboards/BookmarkWidget.vue'
import ClockWidget from '@/components/dashboards/ClockWidget.vue'
import { useGetDashboard } from '@/composables/useDashboards'

const route = useRoute()
const router = useRouter()
const id = computed(() => route.params.id as string)
const { data: dashboard, isLoading, isError } = useGetDashboard(() => id.value)
</script>

<template>
    <div v-if="isLoading" class="p-4">Loading...</div>
    <div v-else-if="isError" class="not-found">
        <i class="ti ti-error-404" />
        <h1>Page not found</h1>
        <p>The dashboard you're looking for doesn't exist.</p>
        <Button label="Go to Dashboards" icon="ti ti-arrow-left" @click="router.push({ name: 'dashboards' })" />
    </div>
    <div v-else-if="dashboard" class="dashboard-view" :class="{ 'show-boxes': dashboard.container.showBoxes }">
        <div
            class="dashboard-container"
            :style="{
                maxWidth: dashboard.container.maxWidth,
                marginLeft: dashboard.container.horizontalAlign === 'right' ? 'auto' : dashboard.container.horizontalAlign === 'center' ? 'auto' : undefined,
                marginRight: dashboard.container.horizontalAlign === 'left' ? 'auto' : dashboard.container.horizontalAlign === 'center' ? 'auto' : undefined,
            }"
        >
            <div
                class="dashboard-rows"
                :style="{
                    display: 'flex',
                    flexDirection: 'column',
                    minHeight: '80vh',
                    justifyContent: dashboard.container.verticalAlign === 'center' ? 'center' : dashboard.container.verticalAlign === 'bottom' ? 'flex-end' : 'flex-start',
                }"
            >
                <div
                    v-for="row in dashboard.rows"
                    :key="row.id"
                    class="dashboard-row"
                    :style="{ height: row.height, maxWidth: row.width, margin: '0 auto', width: '100%' }"
                >
                    <h3 v-if="row.title" class="row-title">{{ row.title }}</h3>
                    <div class="grid">
                        <div
                            v-for="widget in row.widgets"
                            :key="widget.id"
                            :class="'col-' + widget.width"
                        >
                            <WeatherWidget v-if="widget.type === 'weather'" :widget="widget" />
                            <BookmarkWidget v-else-if="widget.type === 'bookmark'" :widget="widget" />
                            <ClockWidget v-else-if="widget.type === 'clock'" :widget="widget" />
                            <WidgetPlaceholder v-else :title="widget.title" />
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
.dashboard-row {
    padding: 0.5rem;
}

.row-title {
    margin: 0 0 0.5rem 0;
    font-size: 1rem;
    font-weight: 600;
}

.not-found {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    min-height: 60vh;
    gap: 0.75rem;
    color: var(--p-text-muted-color);
}

.not-found .ti-error-404 {
    font-size: 4rem;
}

.not-found h1 {
    margin: 0;
    color: var(--p-text-color);
}

.show-boxes .dashboard-container {
    border: 2px dashed var(--p-primary-color);
}

.show-boxes .dashboard-row {
    border: 2px dashed var(--p-orange-500, #f59e0b);
}

.show-boxes .grid > div {
    border: 1px dotted var(--p-text-muted-color);
}
</style>
