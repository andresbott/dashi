<script setup lang="ts">
import { computed, inject, ref } from 'vue'
import { DASHBOARD_ID } from '@/lib/injectionKeys'
import type { Widget } from '@/types/dashboard'
import type { ImageWidgetConfig } from '@/types/image'

const props = defineProps<{
    widget: Widget
}>()

const dashboardId = inject(DASHBOARD_ID, ref(''))

const config = computed<ImageWidgetConfig>(() => {
    const c = props.widget.config as unknown as ImageWidgetConfig | undefined
    return {
        image: c?.image ?? '',
        fit: c?.fit ?? 'cover',
    }
})

const imageUrl = computed(() => {
    if (!config.value.image || !dashboardId.value) return ''
    const assetDashId = dashboardId.value.endsWith('-prev')
        ? dashboardId.value.slice(0, -5)
        : dashboardId.value
    return `/api/v0/dashboards/${assetDashId}/assets/${encodeURIComponent(config.value.image)}`
})
</script>

<template>
    <div class="image-widget">
        <img
            v-if="imageUrl"
            :src="imageUrl"
            :style="{ objectFit: config.fit }"
        />
        <div v-else class="image-empty">
            <i class="ti ti-photo" />
        </div>
    </div>
</template>

<style scoped>
.image-widget {
    width: 100%;
    height: 100%;
    overflow: hidden;
}

.image-widget img {
    width: 100%;
    height: 100%;
    display: block;
}

.image-empty {
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--p-text-muted-color);
    font-size: 2rem;
}
</style>
