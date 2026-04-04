import { useQuery, useMutation, useQueryClient } from '@tanstack/vue-query'
import {
    getDashboards,
    getDashboard,
    createDashboard,
    updateDashboard,
    deleteDashboard,
    deletePreviews
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
        isDeletingPreviews: deletePreviewsMutation.isPending
    }
}

export function useGetDashboard(id: () => string) {
    return useQuery({
        queryKey: ['dashboard', id],
        queryFn: () => getDashboard(id())
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
