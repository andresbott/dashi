import { computed } from 'vue'
import { useQuery, useMutation, useQueryClient } from '@tanstack/vue-query'
import {
    listBackgrounds,
    getBackground,
    createBackground,
    updateBackground,
    deleteBackground,
    listBackgroundAssets,
    uploadBackgroundAsset,
    deleteBackgroundAsset,
} from '@/lib/api/backgrounds'
import type { Background, CreateBackgroundDTO } from '@/types/background'
import { invalidateAndRefetch } from '@/composables/queryUtils'

const BACKGROUNDS_QUERY_KEY = ['backgrounds']

export function useListBackgrounds() {
    const queryClient = useQueryClient()
    const doInvalidate = () => invalidateAndRefetch(queryClient, BACKGROUNDS_QUERY_KEY)

    const query = useQuery({
        queryKey: BACKGROUNDS_QUERY_KEY,
        queryFn: listBackgrounds,
    })

    const createMutation = useMutation({
        mutationFn: (payload: CreateBackgroundDTO) => createBackground(payload),
        onSuccess: doInvalidate,
    })

    const deleteMutation = useMutation({
        mutationFn: (id: string) => deleteBackground(id),
        onSuccess: doInvalidate,
    })

    return {
        backgrounds: query.data,
        isLoading: query.isLoading,
        isError: query.isError,
        error: query.error,

        createBackground: createMutation.mutateAsync,
        isCreating: createMutation.isPending,

        deleteBackground: deleteMutation.mutateAsync,
        isDeleting: deleteMutation.isPending,
    }
}

export function useGetBackground(id: () => string) {
    const idRef = computed(id)
    return useQuery({
        queryKey: ['background', idRef],
        queryFn: () => getBackground(idRef.value),
        enabled: computed(() => !!idRef.value),
    })
}

export function useUpdateBackground() {
    const queryClient = useQueryClient()
    const mutation = useMutation({
        mutationFn: ({ id, payload }: { id: string; payload: Background }) =>
            updateBackground(id, payload),
        onSuccess: (data, variables) => {
            invalidateAndRefetch(queryClient, BACKGROUNDS_QUERY_KEY)
            queryClient.setQueryData(['background', variables.id], data)
        },
    })
    return {
        updateBackground: mutation.mutateAsync,
        isUpdating: mutation.isPending,
    }
}

export function useBackgroundAssets(id: () => string) {
    const idRef = computed(id)
    const queryClient = useQueryClient()
    const key = computed(() => ['background-assets', idRef.value])
    const doInvalidate = () => invalidateAndRefetch(queryClient, ['background-assets', idRef.value])

    const query = useQuery({
        queryKey: key,
        queryFn: () => listBackgroundAssets(idRef.value),
        enabled: computed(() => !!idRef.value),
    })

    const uploadMutation = useMutation({
        mutationFn: ({ name, bytes }: { name: string; bytes: ArrayBuffer }) =>
            uploadBackgroundAsset(idRef.value, name, bytes),
        onSuccess: doInvalidate,
    })

    const deleteMutation = useMutation({
        mutationFn: (name: string) => deleteBackgroundAsset(idRef.value, name),
        onSuccess: doInvalidate,
    })

    return {
        assets: query.data,
        isLoadingAssets: query.isLoading,
        uploadAsset: uploadMutation.mutateAsync,
        isUploadingAsset: uploadMutation.isPending,
        deleteAsset: deleteMutation.mutateAsync,
    }
}
