<script setup lang="ts">
import { computed } from 'vue'
import { docSegments } from '@/lib/widgetDocs'

// Renders a widget-docs string: `backtick` spans become <code>, the rest is
// plain text (no v-html).
const props = defineProps<{ text: string }>()

const segments = computed(() => docSegments(props.text))
</script>

<template>
    <template v-for="(s, i) in segments" :key="i">
        <code v-if="s.code">{{ s.text }}</code>
        <template v-else>{{ s.text }}</template>
    </template>
</template>
