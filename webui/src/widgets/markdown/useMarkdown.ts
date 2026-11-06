import { useQuery } from '@tanstack/vue-query'
import { getMarkdownHtml } from '@/lib/api/notes'
import type { Ref } from 'vue'
import { computed } from 'vue'

export function useMarkdown(filename: Ref<string>) {
    const enabled = computed(() => !!filename.value)

    return useQuery({
        queryKey: ['data-notes-html', filename],
        queryFn: () => getMarkdownHtml(filename.value),
        enabled,
    })
}
