<script setup lang="ts">
// The admin-curated pool of images that any background can reference with a
// `shared:` reference. Backgrounds themselves live in AdminBackgrounds.vue;
// this page is only the file library behind them.
import DataTableView from '@/components/admin/DataTableView.vue'
</script>

<template>
    <DataTableView kind="backgrounds" />
</template>
