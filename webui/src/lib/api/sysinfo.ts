import { apiClient } from '@/lib/api/client'
import type { SystemInfo } from '@/types/sysinfo'

const SYSINFO_PATH = '/widgets/sysinfo'

export const getSysinfo = async (): Promise<SystemInfo> => {
    const { data } = await apiClient.get<SystemInfo>(SYSINFO_PATH)
    return data
}
