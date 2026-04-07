import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
    history: createWebHistory('/'),
    routes: [
        {
            path: '/',
            name: 'default-dashboard',
            component: () => import('@/views/dashboards/DashboardView.vue'),
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

export default router
