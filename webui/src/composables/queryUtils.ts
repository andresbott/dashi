import type { QueryClient } from '@tanstack/vue-query'

export function invalidateAndRefetch(queryClient: QueryClient, queryKey: unknown[]): void {
    queryClient.invalidateQueries({ queryKey })
}
