<script setup lang="ts">
// The admin-curated pool of images that any theme background can reference
// with a `shared:` reference. Themes themselves live in AdminThemes.vue; this
// page is only the file library behind them.
import DataTableView from '@/components/admin/DataTableView.vue'
</script>

<template>
    <DataTableView kind="backgrounds" />
</template>
