import { createRouter, createWebHistory } from 'vue-router'
import { appBase } from '@/lib/base'

const router = createRouter({
    history: createWebHistory(appBase()),
    routes: [
        {
            // The Vue app is the admin UI; landing goes straight to /admin.
            // The public dashboard view is served by the backend's public port.
            path: '/',
            redirect: '/admin',
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
                { path: 'backgrounds/:id', name: 'admin-background-edit', component: () => import('@/views/admin/AdminBackgroundEdit.vue') },
                { path: 'background-images', name: 'admin-background-images', component: () => import('@/views/admin/AdminBackgroundImages.vue') },
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
