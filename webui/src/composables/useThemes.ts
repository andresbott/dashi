import { computed } from 'vue'
import { useQuery, useMutation, useQueryClient } from '@tanstack/vue-query'
import {
    listThemes,
    getTheme,
    createTheme,
    updateTheme,
    deleteTheme,
    copyTheme,
    listThemeAssets,
    uploadThemeAsset,
} from '@/lib/api/themes'
import type { Theme, CreateThemeDTO } from '@/types/theme'
import { invalidateAndRefetch } from '@/composables/queryUtils'

const THEMES_QUERY_KEY = ['themes']

export function useListThemes() {
    const queryClient = useQueryClient()
    const doInvalidate = () => invalidateAndRefetch(queryClient, THEMES_QUERY_KEY)

    const query = useQuery({
        queryKey: THEMES_QUERY_KEY,
        queryFn: listThemes,
    })

    const createMutation = useMutation({
        mutationFn: (payload: CreateThemeDTO) => createTheme(payload),
        onSuccess: doInvalidate,
    })

    const deleteMutation = useMutation({
        mutationFn: (id: string) => deleteTheme(id),
        onSuccess: doInvalidate,
    })

    const copyMutation = useMutation({
        mutationFn: (id: string) => copyTheme(id),
        onSuccess: doInvalidate,
    })

    return {
        themes: query.data,
        isLoading: query.isLoading,
        isError: query.isError,
        error: query.error,

        createTheme: createMutation.mutateAsync,
        isCreating: createMutation.isPending,

        deleteTheme: deleteMutation.mutateAsync,
        isDeleting: deleteMutation.isPending,

        copyTheme: copyMutation.mutateAsync,
        isCopying: copyMutation.isPending,
    }
}

export function useGetTheme(id: () => string) {
    const idRef = computed(id)
    return useQuery({
        queryKey: ['theme', idRef],
        queryFn: () => getTheme(idRef.value),
        enabled: computed(() => !!idRef.value),
    })
}

export function useUpdateTheme() {
    const queryClient = useQueryClient()
    const mutation = useMutation({
        mutationFn: ({ id, payload }: { id: string; payload: Theme }) =>
            updateTheme(id, payload),
        onSuccess: (data, variables) => {
            invalidateAndRefetch(queryClient, THEMES_QUERY_KEY)
            queryClient.setQueryData(['theme', variables.id], data)
        },
    })
    return {
        updateTheme: mutation.mutateAsync,
        isUpdating: mutation.isPending,
    }
}

export function useThemeAssets(id: () => string) {
    const idRef = computed(id)
    const queryClient = useQueryClient()
    const key = computed(() => ['theme-assets', idRef.value])
    const doInvalidate = () => invalidateAndRefetch(queryClient, ['theme-assets', idRef.value])

    const query = useQuery({
        queryKey: key,
        queryFn: () => listThemeAssets(idRef.value),
        enabled: computed(() => !!idRef.value),
    })

    const uploadMutation = useMutation({
        mutationFn: ({ name, bytes }: { name: string; bytes: ArrayBuffer }) =>
            uploadThemeAsset(idRef.value, name, bytes),
        onSuccess: doInvalidate,
    })

    // No delete here: the editor's Remove only clears a ref and pruning on save
    // deletes the file (deleteThemeAsset stays in the API client).
    return {
        assets: query.data,
        isLoadingAssets: query.isLoading,
        uploadAsset: uploadMutation.mutateAsync,
        isUploadingAsset: uploadMutation.isPending,
    }
}
