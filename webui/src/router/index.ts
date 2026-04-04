import { createRouter, createWebHistory } from 'vue-router'

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

export default router
