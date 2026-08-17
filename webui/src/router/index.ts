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
            path: '/admin',
            component: () => import('@/views/admin/AdminLayout.vue'),
            children: [
                { path: '', redirect: { name: 'admin-dashboards' } },
                { path: 'dashboards', name: 'admin-dashboards', component: () => import('@/views/admin/AdminDashboards.vue') },
                { path: 'notes', name: 'admin-notes', component: () => import('@/views/admin/AdminNotes.vue') },
                { path: 'images', name: 'admin-images', component: () => import('@/views/admin/AdminImages.vue') },
                { path: 'backgrounds', name: 'admin-backgrounds', component: () => import('@/views/admin/AdminBackgrounds.vue') },
                { path: 'themes', name: 'admin-themes', component: () => import('@/views/admin/AdminThemes.vue') },
                { path: 'docs/dashboards', name: 'doc-dashboards', component: () => import('@/views/docs/DocDashboards.vue') },
                { path: 'docs/widgets', name: 'doc-widgets', component: () => import('@/views/docs/DocWidgets.vue') },
                { path: 'docs/theming', name: 'doc-theming', component: () => import('@/views/docs/DocTheming.vue') },
            ],
        },
        {
            path: '/dashboards/:id/edit',
            name: 'dashboard-edit',
            component: () => import('@/views/dashboards/DashboardEditView.vue')
        },
        {
            path: '/dashboards/:id/settings',
            name: 'dashboard-settings',
            component: () => import('@/views/dashboards/DashboardSettingsView.vue')
        },
        {
            path: '/:id',
            name: 'dashboard-view',
            component: () => import('@/views/dashboards/DashboardView.vue')
        }
    ]
})

export default router
