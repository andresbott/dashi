import { computed } from 'vue'
import { useQuery, useMutation, useQueryClient } from '@tanstack/vue-query'
import {
    getDashboards,
    getDashboard,
    createDashboard,
    updateDashboard,
    deleteDashboard,
    getBackgrounds,
    getDashboardAssets,
    uploadDashboardZip,
    getDashboardAuth,
    setDashboardAuth,
    deleteDashboardAuth,
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
        onSuccess: (data, variables) => {
            invalidateAndRefetch(queryClient, DASHBOARDS_QUERY_KEY)
            queryClient.setQueryData(['dashboard', variables.id], data)
        }
    })

    return {
        updateDashboard: mutation.mutateAsync,
        isUpdating: mutation.isPending
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

export function useDashboardAssets(dashboardId: () => string) {
    const idRef = computed(dashboardId)
    return useQuery({
        queryKey: ['dashboard-assets', idRef],
        queryFn: () => getDashboardAssets(idRef.value),
        enabled: computed(() => !!idRef.value),
    })
}

export function useDashboardAuth(dashboardId: () => string) {
    const idRef = computed(dashboardId)
    const queryClient = useQueryClient()

    const query = useQuery({
        queryKey: ['dashboard-auth', idRef],
        queryFn: () => getDashboardAuth(idRef.value),
        enabled: computed(() => !!idRef.value),
    })

    const setMutation = useMutation({
        mutationFn: ({ username, password }: { username: string; password: string }) =>
            setDashboardAuth(idRef.value, username, password),
        onSuccess: () => invalidateAndRefetch(queryClient, ['dashboard-auth', idRef.value]),
    })

    const deleteMutation = useMutation({
        mutationFn: () => deleteDashboardAuth(idRef.value),
        onSuccess: () => invalidateAndRefetch(queryClient, ['dashboard-auth', idRef.value]),
    })

    return {
        auth: query.data,
        isLoadingAuth: query.isLoading,
        setAuth: setMutation.mutateAsync,
        isSettingAuth: setMutation.isPending,
        deleteAuth: deleteMutation.mutateAsync,
        isDeletingAuth: deleteMutation.isPending,
    }
}
