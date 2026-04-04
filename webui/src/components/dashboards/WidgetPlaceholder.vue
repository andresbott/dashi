<script setup lang="ts">
defineProps<{
    title: string
}>()
</script>

<template>
    <div class="widget-placeholder">
    </div>
</template>

<style scoped>
.widget-placeholder {
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 80px;
}

.widget-placeholder-title {
    color: var(--p-text-muted-color);
    font-size: 0.875rem;
}
</style>
