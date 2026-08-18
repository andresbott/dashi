import { useQuery } from '@tanstack/vue-query'
import { getSysinfo } from './api'

const ONE_MINUTE = 60 * 1000

export function useSysinfo() {
    return useQuery({
        queryKey: ['sysinfo'],
        queryFn: () => getSysinfo(),
        refetchInterval: ONE_MINUTE,
    })
}
