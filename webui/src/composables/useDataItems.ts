import { useQuery, useMutation, useQueryClient } from '@tanstack/vue-query'
import { listData, uploadData, deleteData, type DataKind } from '@/lib/api/data'
import { invalidateAndRefetch } from '@/composables/queryUtils'

export function useDataItems(kind: DataKind) {
    const queryClient = useQueryClient()
    const queryKey = ['data', kind]
    const doInvalidate = () => invalidateAndRefetch(queryClient, queryKey)

    const query = useQuery({
        queryKey,
        queryFn: () => listData(kind),
    })

    const uploadMutation = useMutation({
        mutationFn: ({ name, bytes }: { name: string; bytes: ArrayBuffer }) =>
            uploadData(kind, name, bytes),
        onSuccess: doInvalidate,
    })

    const deleteMutation = useMutation({
        mutationFn: (name: string) => deleteData(kind, name),
        onSuccess: doInvalidate,
    })

    return {
        items: query.data,
        isLoading: query.isLoading,
        isError: query.isError,

        uploadItem: uploadMutation.mutateAsync,
        isUploading: uploadMutation.isPending,

        deleteItem: deleteMutation.mutateAsync,
        isDeleting: deleteMutation.isPending,
    }
}
