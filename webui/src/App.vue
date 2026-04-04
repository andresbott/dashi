<script setup>
import Toast from 'primevue/toast'
</script>

<template>
    <Toast />
    <div class="app-layout">
        <main>
            <router-view />
        </main>
    </div>
</template>

<style lang="css">
.app-layout {
    display: flex;
    flex-direction: column;
    min-height: 100vh;
}

.app-topbar {
    display: flex;
    align-items: center;
    padding: 0 1rem;
    height: 3rem;
    background: var(--p-primary-color);
    color: var(--p-primary-contrast-color);
}

.app-topbar-title {
    font-size: 1.25rem;
    font-weight: 700;
    cursor: pointer;
}

.app-layout main {
    flex: 1;
}
</style>
