<script setup lang="ts">
import { computed } from 'vue'
import { useFontIconClass } from '@/composables/useThemes'
import { getIconUrl } from '@/lib/api/themes'
import type { ThemeInfo } from '@/types/theme'

const props = defineProps<{
    iconName: string
    themeName?: string
    themes?: ThemeInfo[]
}>()

const resolvedThemeName = computed(() => props.themeName || 'default')

const themeType = computed(() => {
    if (!props.themes) return 'font'
    return props.themes.find(t => t.name === resolvedThemeName.value)?.type ?? 'font'
})

const isFontTheme = computed(() => themeType.value === 'font')
const iconNameRef = computed(() => props.iconName)
const themeNameRef = computed(() => resolvedThemeName.value)

const { data: fontData } = useFontIconClass(themeNameRef, iconNameRef, isFontTheme)

const fontClass = computed(() => fontData.value?.class ?? '')
const imageUrl = computed(() => getIconUrl(resolvedThemeName.value, props.iconName))
</script>

<template>
    <i v-if="themeType === 'font'" :class="fontClass" />
    <img
        v-else
        :src="imageUrl"
        :alt="iconName"
        class="weather-theme-icon"
    />
</template>

<style scoped>
.weather-theme-icon {
    width: 1em;
    height: 1em;
    vertical-align: middle;
}
</style>
