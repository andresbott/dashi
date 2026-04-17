import { useQuery, useQueryClient } from '@tanstack/vue-query'
import { listMarkdownFiles } from '@/lib/api/markdown'
import type { Ref } from 'vue'
import { computed } from 'vue'

export function useMarkdownFiles(dashboardId: Ref<string>) {
    const queryClient = useQueryClient()
    const enabled = computed(() => !!dashboardId.value)

    const query = useQuery({
        queryKey: ['markdownFiles', dashboardId],
        queryFn: () => listMarkdownFiles(dashboardId.value),
        enabled,
    })

    const files = computed(() => query.data.value ?? [])

    const invalidate = () =>
        queryClient.invalidateQueries({ queryKey: ['markdownFiles', dashboardId.value] })

    return {
        files,
        isLoading: query.isLoading,
        isError: query.isError,
        invalidate,
    }
}
