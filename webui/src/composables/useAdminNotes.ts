import { useQuery, useMutation, useQueryClient } from '@tanstack/vue-query'
import { listMarkdownFiles, deleteMarkdown } from '@/widgets/markdown/api'
import { invalidateAndRefetch } from '@/composables/queryUtils'

const NOTES_QUERY_KEY = ['admin-notes']

export function useAdminNotes() {
    const queryClient = useQueryClient()
    const doInvalidate = () => invalidateAndRefetch(queryClient, NOTES_QUERY_KEY)

    const query = useQuery({
        queryKey: NOTES_QUERY_KEY,
        queryFn: listMarkdownFiles,
    })

    const deleteMutation = useMutation({
        mutationFn: (name: string) => deleteMarkdown(name),
        onSuccess: doInvalidate,
    })

    return {
        notes: query.data,
        isLoading: query.isLoading,
        isError: query.isError,
        refetch: doInvalidate,

        deleteNote: deleteMutation.mutateAsync,
        isDeleting: deleteMutation.isPending,
    }
}
