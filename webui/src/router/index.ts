import { createRouter, createWebHistory } from 'vue-router'
import { queryClient } from '@/lib/queryClient'
import { getSettings } from '@/lib/api/settings'

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
            path: '/docs',
            component: () => import('@/views/DocumentationView.vue'),
            children: [
                { path: '', redirect: { name: 'doc-dashboards' } },
                { path: 'dashboards', name: 'doc-dashboards', component: () => import('@/views/docs/DocDashboards.vue') },
                { path: 'widgets', name: 'doc-widgets', component: () => import('@/views/docs/DocWidgets.vue') },
                { path: 'theming', name: 'doc-theming', component: () => import('@/views/docs/DocTheming.vue') },
            ],
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
            const settings = await queryClient.fetchQuery({
                queryKey: ['settings'],
                queryFn: getSettings,
                staleTime: 5 * 60 * 1000,
            })
            if (settings.readOnly) {
                return { name: 'dashboards' }
            }
        } catch {
            // allow navigation if settings fetch fails
        }
    }
})

export default router
