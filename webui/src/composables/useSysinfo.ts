import { useQuery } from '@tanstack/vue-query'
import { getSysinfo } from '@/lib/api/sysinfo'

const ONE_MINUTE = 60 * 1000

export function useSysinfo() {
    return useQuery({
        queryKey: ['sysinfo'],
        queryFn: () => getSysinfo(),
        refetchInterval: ONE_MINUTE,
    })
}
