<script setup lang="ts">
// Placeholder so the admin-background-edit route resolves. The real editor
// (name, colour/gradient base, image with fit/position/repeat, dark variants
// and a live preview) lands in the next task.
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import Card from 'primevue/card'

const route = useRoute()
const router = useRouter()
const id = computed(() => route.params.id as string)
</script>

<template>
    <div class="admin-section">
        <div class="admin-section-header">
            <h2 class="admin-section-title">Edit background</h2>
            <Button
                label="Back"
                icon="ti ti-arrow-left"
                severity="secondary"
                @click="router.push({ name: 'admin-backgrounds' })"
            />
        </div>
        <Card>
            <template #content>
                <div class="info-message">Editor for background {{ id }} is not built yet.</div>
            </template>
        </Card>
    </div>
</template>

<style scoped>
.admin-section {
    display: flex;
    flex-direction: column;
    gap: 1rem;
}

.admin-section-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 1rem;
}

.admin-section-title {
    font-size: 1.5rem;
    font-weight: 700;
    color: var(--p-text-color);
    margin: 0;
}

.info-message {
    padding: 1rem;
    text-align: center;
    color: var(--p-text-muted-color);
}
</style>
