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
  ],
})

export default router
