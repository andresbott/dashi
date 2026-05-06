import { useQuery } from '@tanstack/vue-query'
import { listDataImages } from './api'

export function useDataImages() {
    return useQuery({
        queryKey: ['data-images'],
        queryFn: listDataImages,
    })
}
