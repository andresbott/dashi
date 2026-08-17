import { listData, type DataItem } from '@/lib/api/data'

export type { DataItem }

export const listDataImages = (): Promise<DataItem[]> => listData('images')
