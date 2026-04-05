import { createRouter, createWebHistory } from 'vue-router'
import { apiClient } from '@/lib/api/client'

const router = createRouter({
    history: createWebHistory('/'),
    routes: [
        {
            path: '/',
            redirect: '/dashboards'
        },
        {
            path: '/dashboards',
            name: 'dashboards',
            component: () => import('@/views/dashboards/DashboardListView.vue')
        },
        {
            path: '/dashboards/:id/edit',
            name: 'dashboard-edit',
            component: () => import('@/views/dashboards/DashboardEditView.vue')
        },
        {
            path: '/:id',
            name: 'dashboard-view',
            component: () => import('@/views/dashboards/DashboardView.vue')
        }
    ]
})

router.beforeEach(async (to) => {
    if (to.name === 'dashboard-edit') {
        try {
            const { data } = await apiClient.get<{ readOnly: boolean }>('/settings')
            if (data.readOnly) {
                return { name: 'dashboards' }
            }
        } catch {
            // allow navigation if settings fetch fails
        }
    }
})

export default router
