import axios from 'axios'
import { withBase } from '@/lib/base'

export const apiClient = axios.create({
    baseURL: withBase('/api/v0'),
    headers: {
        'Content-Type': 'application/json'
    }
})
