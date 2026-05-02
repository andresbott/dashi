import { useQuery } from '@tanstack/vue-query'
import { getMarkdownHtml } from './api'
import type { Ref } from 'vue'
import { computed } from 'vue'

export function useMarkdown(dashboardId: Ref<string>, filename: Ref<string>) {
    const enabled = computed(() => !!dashboardId.value && !!filename.value)

    return useQuery({
        queryKey: ['markdown', dashboardId, filename],
        queryFn: () => getMarkdownHtml(dashboardId.value, filename.value),
        enabled,
    })
}
