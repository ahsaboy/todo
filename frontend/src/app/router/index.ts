import { createRouter, createWebHashHistory } from 'vue-router'
import AppLayout from '@/widgets/AppLayout.vue'

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('@/pages/LoginPage.vue'),
      meta: { requiresGuest: true, shell: 'auth' },
    },
    {
      path: '/register',
      name: 'register',
      component: () => import('@/pages/RegisterPage.vue'),
      meta: { requiresGuest: true, shell: 'auth' },
    },
    {
      path: '/',
      component: AppLayout,
      meta: { requiresAuth: true, shell: 'app' },
      children: [
        {
          path: '',
          redirect: '/tasks',
        },
        {
          path: 'tasks',
          name: 'tasks',
          component: () => import('@/pages/TasksPage.vue'),
          meta: { motion: 'page' },
        },
        {
          path: 'tasks/today',
          name: 'tasks-today',
          component: () => import('@/pages/TodayTasksPage.vue'),
          meta: { motion: 'page' },
        },
        {
          path: 'tasks/upcoming',
          name: 'tasks-upcoming',
          component: () => import('@/pages/UpcomingTasksPage.vue'),
          meta: { motion: 'page' },
        },
        {
          path: 'tasks/board',
          name: 'tasks-board',
          component: () => import('@/pages/BoardPage.vue'),
          meta: { motion: 'board' },
        },
        {
          path: 'reminder-configs',
          name: 'reminder-configs',
          component: () => import('@/pages/ReminderConfigsPage.vue'),
          meta: { motion: 'page' },
        },
        {
          path: 'reminder-logs',
          name: 'reminder-logs',
          component: () => import('@/pages/ReminderLogsPage.vue'),
          meta: { motion: 'page' },
        },
        {
          path: 'api-keys',
          name: 'api-keys',
          component: () => import('@/pages/ApiKeysPage.vue'),
          meta: { motion: 'page' },
        },
        {
          path: 'profile',
          name: 'profile',
          component: () => import('@/pages/ProfilePage.vue'),
          meta: { motion: 'page' },
        },
      ],
    },
    {
      path: '/admin/login',
      name: 'admin-login',
      component: () => import('@/pages/admin/AdminLoginPage.vue'),
      meta: { requiresAdminGuest: true },
    },
    {
      path: '/admin',
      component: () => import('@/widgets/AdminLayout.vue'),
      meta: { requiresAdmin: true },
      children: [
        { path: '', redirect: '/admin/dashboard' },
        { path: 'dashboard', name: 'admin-dashboard', component: () => import('@/pages/admin/AdminDashboardPage.vue') },
        { path: 'users', name: 'admin-users', component: () => import('@/pages/admin/AdminUsersPage.vue') },
        { path: 'tasks', name: 'admin-tasks', component: () => import('@/pages/admin/AdminTasksPage.vue') },
        { path: 'reminder-configs', name: 'admin-reminder-configs', component: () => import('@/pages/admin/AdminReminderConfigsPage.vue') },
        { path: 'reminder-logs', name: 'admin-reminder-logs', component: () => import('@/pages/admin/AdminReminderLogsPage.vue') },
        { path: 'system-logs', name: 'admin-system-logs', component: () => import('@/pages/admin/AdminSystemLogsPage.vue') },
        { path: 'config', name: 'admin-config', component: () => import('@/pages/admin/AdminConfigPage.vue') },
      ],
    },
  ],
})

export default router
