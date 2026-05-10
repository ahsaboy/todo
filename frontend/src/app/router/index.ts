import { createRouter, createWebHashHistory } from 'vue-router'
import AppLayout from '@/widgets/AppLayout.vue'

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('@/pages/LoginPage.vue'),
      meta: { requiresGuest: true },
    },
    {
      path: '/register',
      name: 'register',
      component: () => import('@/pages/RegisterPage.vue'),
      meta: { requiresGuest: true },
    },
    {
      path: '/',
      component: AppLayout,
      meta: { requiresAuth: true },
      children: [
        {
          path: '',
          redirect: '/tasks',
        },
        {
          path: 'tasks',
          name: 'tasks',
          component: () => import('@/pages/TasksPage.vue'),
        },
        {
          path: 'tasks/today',
          name: 'tasks-today',
          component: () => import('@/pages/TodayTasksPage.vue'),
        },
        {
          path: 'tasks/upcoming',
          name: 'tasks-upcoming',
          component: () => import('@/pages/UpcomingTasksPage.vue'),
        },
        {
          path: 'tasks/board',
          name: 'tasks-board',
          component: () => import('@/pages/BoardPage.vue'),
        },
        {
          path: 'reminder-configs',
          name: 'reminder-configs',
          component: () => import('@/pages/ReminderConfigsPage.vue'),
        },
        {
          path: 'reminder-logs',
          name: 'reminder-logs',
          component: () => import('@/pages/ReminderLogsPage.vue'),
        },
        {
          path: 'api-keys',
          name: 'api-keys',
          component: () => import('@/pages/ApiKeysPage.vue'),
        },
        {
          path: 'profile',
          name: 'profile',
          component: () => import('@/pages/ProfilePage.vue'),
        },
      ],
    },
  ],
})

export default router
