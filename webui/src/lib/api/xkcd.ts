import { apiClient } from '@/lib/api/client'
import type { XkcdComic } from '@/types/xkcd'

const XKCD_PATH = '/widgets/xkcd'

export const getXkcd = async (mode: string = 'latest'): Promise<XkcdComic> => {
    const { data } = await apiClient.get<XkcdComic>(XKCD_PATH, {
        params: { mode }
    })
    return data
}
