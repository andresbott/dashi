import { useQuery, useMutation, useQueryClient } from '@tanstack/vue-query'
import { getWidgetThemes, getFontIcon, uploadWidgetTheme, deleteWidgetTheme } from '@/lib/api/widgetThemes'
import { invalidateAndRefetch } from '@/composables/queryUtils'
import type { Ref } from 'vue'

const WIDGET_THEMES_QUERY_KEY = ['widget-themes']

export function useWidgetThemes() {
    return useQuery({
        queryKey: WIDGET_THEMES_QUERY_KEY,
        queryFn: getWidgetThemes,
        staleTime: 5 * 60 * 1000, // widget themes rarely change
    })
}

export function useFontIconClass(themeName: Ref<string>, iconName: Ref<string>, enabled: Ref<boolean>) {
    return useQuery({
        queryKey: ['widget-theme-icon', themeName, iconName],
        queryFn: () => getFontIcon(themeName.value || 'default', iconName.value),
        enabled: () => enabled.value && !!iconName.value,
        staleTime: 60 * 60 * 1000, // icon mappings are stable
    })
}

// useAdminWidgetThemes powers the admin page. It shares useWidgetThemes'
// query key so uploads and deletes invalidate the cached list everywhere.
export function useAdminWidgetThemes() {
    const queryClient = useQueryClient()
    const doInvalidate = () => invalidateAndRefetch(queryClient, WIDGET_THEMES_QUERY_KEY)

    const query = useQuery({
        queryKey: WIDGET_THEMES_QUERY_KEY,
        queryFn: getWidgetThemes,
    })

    const uploadMutation = useMutation({
        mutationFn: (zipBytes: ArrayBuffer) => uploadWidgetTheme(zipBytes),
        onSuccess: doInvalidate,
    })

    const deleteMutation = useMutation({
        mutationFn: (name: string) => deleteWidgetTheme(name),
        onSuccess: doInvalidate,
    })

    return {
        widgetThemes: query.data,
        isLoading: query.isLoading,
        isError: query.isError,

        uploadWidgetTheme: uploadMutation.mutateAsync,
        isUploading: uploadMutation.isPending,

        deleteWidgetTheme: deleteMutation.mutateAsync,
        isDeleting: deleteMutation.isPending,
    }
}
