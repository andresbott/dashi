import { useQuery, useQueryClient } from '@tanstack/vue-query'
import { listMarkdownFiles } from './api'
import { computed } from 'vue'

export function useMarkdownFiles() {
    const queryClient = useQueryClient()

    const query = useQuery({
        queryKey: ['data-notes-list'],
        queryFn: () => listMarkdownFiles(),
    })

    const files = computed(() => query.data.value ?? [])

    const invalidate = () =>
        queryClient.invalidateQueries({ queryKey: ['data-notes-list'] })

    return {
        files,
        isLoading: query.isLoading,
        isError: query.isError,
        invalidate,
    }
}
