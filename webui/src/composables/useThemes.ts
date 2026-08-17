import { useQuery, useMutation, useQueryClient } from '@tanstack/vue-query'
import { getThemes, getFontIcon, uploadTheme, deleteTheme } from '@/lib/api/themes'
import { invalidateAndRefetch } from '@/composables/queryUtils'
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

// useAdminThemes powers the admin CRUD page. It reuses the same query
// key as useThemes so uploads/deletes invalidate the cached list
// everywhere.
export function useAdminThemes() {
    const queryClient = useQueryClient()
    const queryKey = ['themes']
    const doInvalidate = () => invalidateAndRefetch(queryClient, queryKey)

    const query = useQuery({
        queryKey,
        queryFn: getThemes,
    })

    const uploadMutation = useMutation({
        mutationFn: (zipBytes: ArrayBuffer) => uploadTheme(zipBytes),
        onSuccess: doInvalidate,
    })

    const deleteMutation = useMutation({
        mutationFn: (name: string) => deleteTheme(name),
        onSuccess: doInvalidate,
    })

    return {
        themes: query.data,
        isLoading: query.isLoading,
        isError: query.isError,

        uploadTheme: uploadMutation.mutateAsync,
        isUploading: uploadMutation.isPending,

        deleteTheme: deleteMutation.mutateAsync,
        isDeleting: deleteMutation.isPending,
    }
}
