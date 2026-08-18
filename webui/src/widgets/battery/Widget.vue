<script setup lang="ts">
import type { Widget } from '@/types/dashboard'

defineProps<{
    widget: Widget
}>()
</script>

<template>
    <div class="battery-widget">
        <div class="battery-value">{{ widget.title }}</div>
        <div class="battery-hint">Value set via ?battery=0-100 query parameter</div>
    </div>
</template>

<style scoped>
.battery-widget {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 0.75rem 0.5rem;
    min-height: 60px;
}

.battery-value {
    font-size: 1.5rem;
    font-weight: 700;
}

.battery-hint {
    font-size: 0.75rem;
    color: var(--p-text-muted-color);
}
</style>
