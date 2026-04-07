import { computed } from 'vue'
import { useQuery, useMutation, useQueryClient } from '@tanstack/vue-query'
import {
    getDashboards,
    getDashboard,
    createDashboard,
    updateDashboard,
    deleteDashboard,
    deletePreviews,
    getBackgrounds,
    getDashboardAssets,
    uploadDashboardAsset,
    uploadDashboardZip,
} from '@/lib/api/dashboard'
import type { CreateDashboardDTO, Dashboard } from '@/types/dashboard'
import { invalidateAndRefetch } from '@/composables/queryUtils'

const DASHBOARDS_QUERY_KEY = ['dashboards']

export function useListDashboards() {
    const queryClient = useQueryClient()
    const doInvalidate = () => invalidateAndRefetch(queryClient, DASHBOARDS_QUERY_KEY)

    const query = useQuery({
        queryKey: DASHBOARDS_QUERY_KEY,
        queryFn: getDashboards
    })

    const createMutation = useMutation({
        mutationFn: (payload: CreateDashboardDTO) => createDashboard(payload),
        onSuccess: doInvalidate
    })

    const deleteMutation = useMutation({
        mutationFn: (id: string) => deleteDashboard(id),
        onSuccess: doInvalidate
    })

    const deletePreviewsMutation = useMutation({
        mutationFn: () => deletePreviews(),
        onSuccess: doInvalidate
    })

    const uploadZipMutation = useMutation({
        mutationFn: (data: ArrayBuffer) => uploadDashboardZip(data),
        onSuccess: doInvalidate
    })

    return {
        dashboards: query.data,
        isLoading: query.isLoading,
        isError: query.isError,
        error: query.error,

        createDashboard: createMutation.mutateAsync,
        isCreating: createMutation.isPending,

        deleteDashboard: deleteMutation.mutateAsync,
        isDeleting: deleteMutation.isPending,

        deletePreviews: deletePreviewsMutation.mutateAsync,
        isDeletingPreviews: deletePreviewsMutation.isPending,

        uploadZip: uploadZipMutation.mutateAsync,
        isUploadingZip: uploadZipMutation.isPending
    }
}

export function useGetDashboard(id: () => string) {
    const idRef = computed(id)
    return useQuery({
        queryKey: ['dashboard', idRef],
        queryFn: () => getDashboard(idRef.value)
    })
}

export function useUpdateDashboard() {
    const queryClient = useQueryClient()

    const mutation = useMutation({
        mutationFn: ({ id, payload }: { id: string; payload: Dashboard }) =>
            updateDashboard(id, payload),
        onSuccess: (_data, variables) => {
            invalidateAndRefetch(queryClient, DASHBOARDS_QUERY_KEY)
            invalidateAndRefetch(queryClient, ['dashboard', variables.id])
        }
    })

    return {
        updateDashboard: mutation.mutateAsync,
        isUpdating: mutation.isPending
    }
}

export function usePreviewDashboard() {
    const queryClient = useQueryClient()
    const doInvalidate = () => invalidateAndRefetch(queryClient, DASHBOARDS_QUERY_KEY)

    const createMutation = useMutation({
        mutationFn: (payload: CreateDashboardDTO) => createDashboard(payload),
        onSuccess: doInvalidate,
    })

    const updateMutation = useMutation({
        mutationFn: ({ id, payload }: { id: string; payload: Dashboard }) =>
            updateDashboard(id, payload),
        onSuccess: doInvalidate,
    })

    const deleteMutation = useMutation({
        mutationFn: (id: string) => deleteDashboard(id),
        onSuccess: doInvalidate,
    })

    return {
        createPreview: createMutation.mutateAsync,
        updatePreview: updateMutation.mutateAsync,
        deletePreview: deleteMutation.mutateAsync,
    }
}

export function useBackgrounds(dashboardId: () => string) {
    const idRef = computed(dashboardId)
    return useQuery({
        queryKey: ['backgrounds', idRef],
        queryFn: () => getBackgrounds(idRef.value),
        enabled: computed(() => !!idRef.value),
    })
}

export function useUploadDashboardAsset() {
    const queryClient = useQueryClient()

    const mutation = useMutation({
        mutationFn: ({ dashboardId, filename, data }: { dashboardId: string; filename: string; data: ArrayBuffer }) =>
            uploadDashboardAsset(dashboardId, filename, data),
        onSuccess: (_data, variables) => {
            invalidateAndRefetch(queryClient, ['dashboard-assets', variables.dashboardId])
            invalidateAndRefetch(queryClient, ['backgrounds', variables.dashboardId])
        }
    })

    return {
        uploadAsset: mutation.mutateAsync,
        isUploading: mutation.isPending
    }
}

export function useDashboardAssets(dashboardId: () => string) {
    const idRef = computed(dashboardId)
    return useQuery({
        queryKey: ['dashboard-assets', idRef],
        queryFn: () => getDashboardAssets(idRef.value),
        enabled: computed(() => !!idRef.value),
    })
}
