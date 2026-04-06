import { useQuery } from '@tanstack/vue-query'
import { getXkcd } from '@/lib/api/xkcd'
import type { Ref } from 'vue'

export function useXkcd(mode: Ref<string>) {
    return useQuery({
        queryKey: ['xkcd', mode],
        queryFn: () => getXkcd(mode.value),
        refetchInterval: 60 * 60 * 1000,
    })
}
