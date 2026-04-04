import axios from 'axios'

export const apiClient = axios.create({
    baseURL: '/api/v0',
    headers: {
        'Content-Type': 'application/json'
    }
})
