import { useQuery } from '@tanstack/vue-query'
import { getThemes, getFontIcon } from '@/lib/api/themes'
import type { Ref } from 'vue'

export function useThemes() {
    return useQuery({
        queryKey: ['themes'],
        queryFn: getThemes,
        staleTime: 5 * 60 * 1000, // themes rarely change
    })
}

export function useFontIconClass(themeName: Ref<string>, iconName: Ref<string>, enabled: Ref<boolean>) {
    return useQuery({
        queryKey: ['theme-icon', themeName, iconName],
        queryFn: () => getFontIcon(themeName.value || 'default', iconName.value),
        enabled: () => enabled.value && !!iconName.value,
        staleTime: 60 * 60 * 1000, // icon mappings are stable
    })
}
